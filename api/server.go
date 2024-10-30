package api

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	firebase "firebase.google.com/go"
	"github.com/bytedance/sonic"
	db "github.com/cshop/v3/db/sqlc"
	image "github.com/cshop/v3/image"
	"github.com/cshop/v3/mail"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/cshop/v3/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/template/html/v2"
)

type Server struct {
	config          util.Config
	store           db.Store
	fb              *firebase.App
	userTokenMaker  token.Maker
	adminTokenMaker token.Maker
	router          *fiber.App
	taskDistributor worker.TaskDistributor
	ik              image.ImageKitManagement
	sender          mail.EmailSender
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(
	config util.Config,
	store db.Store,
	fb *firebase.App,
	taskDistributor worker.TaskDistributor,
	ik image.ImageKitManagement,
	sender mail.EmailSender,
) (*Server, error) {
	userTokenMaker, err := token.NewPasetoMaker(config.UserTokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	adminTokenMaker, err := token.NewPasetoMaker(config.AdminTokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		fb:              fb,
		userTokenMaker:  userTokenMaker,
		adminTokenMaker: adminTokenMaker,
		taskDistributor: taskDistributor,
		ik:              ik,
		sender:          sender,
	}

	server.setupRouter()
	go server.gracefulShutdown()
	return server, nil
}

func (server *Server) setupRouter() {
	engine := html.New("./web/views", ".html")
	app := fiber.New(
		fiber.Config{
			Views:       engine,
			JSONEncoder: sonic.ConfigFastest.Marshal,
			JSONDecoder: sonic.ConfigFastest.Unmarshal,
		},
	)

	app.Static("/static", "./web/styles")

	// app.Use(app.Use(logger.New()))

	app.Use(etag.New(
		etag.Config{},
	))

	// Get google service account credentials
	// serviceAccount, fileExi := os.LookupEnv("GOOGLE_SERVICE_ACCOUNT")

	// if !fileExi {
	// 	log.Fatalf("Please provide valid firebase auth credential json!")
	// }

	// Initialize the firebase app.
	// opt := option.WithCredentialsFile("serviceAccountKey.json")
	// config := &firebase.Config{ProjectID: "notifications-3eca3"}
	// fb, err := firebase.NewApp(context.Background(), config, opt)
	// if err != nil {
	// 	log.Fatalf("Please provide valid firebase auth credential json!")
	// }

	app.Get("/api/v1/reset_password", server.resetPasswordPage)

	//* Users
	app.Post("/api/v1/users", server.createUser)
	app.Post("/api/v1/users/login", server.loginUser)
	app.Post("/api/v1/users/reset-password", server.resetPassword)
	app.Post("/api/v1/users/new-password", server.newPassword)

	app.Post("/api/v1/users/signup", server.signUp)
	app.Post("/api/v1/users/verify-otp", server.verifyOTP)
	app.Post("/api/v1/users/resend-otp", server.resendOTP)

	//* Admins
	app.Post("/api/v1/admins/login", server.loginAdmin) //! For Admin Only

	//* Tokens
	app.Post("/api/v1/auth/access-token", server.renewAccessToken)
	app.Post("/api/v1/auth/refresh-token", server.renewRefreshToken)

	//* Tokens for Admins
	app.Post("/api/v1/auth/access-token-for-admin", server.renewAccessTokenForAdmin)   //! For Admin Only
	app.Post("/api/v1/auth/refresh-token-for-admin", server.renewRefreshTokenForAdmin) //! For Admin Only

	//*HomePageTextBanner
	app.Get("/api/v1/text-banners/:textBannerId", server.getHomePageTextBanner) //? no auth required
	app.Get("/api/v1/text-banners", server.listHomePageTextBanners)             //? no auth required

	app.Get("/api/v1/app-policy", server.getAppPolicy) //? no auth required

	//*Products
	app.Get("/api/v1/products/:productId", server.getProduct)                      //? no auth required
	app.Get("/api/v1/products", server.listProducts)                               //? no auth required
	app.Get("/api/v1/products-v2", server.listProductsV2)                          //? no auth required                                                       //? no auth required
	app.Get("/api/v1/products-next-page", server.listProductsNextPage)             //? no auth required
	app.Get("/api/v1/search-products", server.searchProducts)                      //? no auth required
	app.Get("/api/v1/search-products-next-page", server.searchProductsNextPage)    //? no auth required
	app.Get("/api/v1/products-best-sellers", server.listProductItemsWithBestSales) //? no auth required

	//*Promotions
	app.Get("/api/v1/promotions/:promotionId", server.getPromotion) //? no auth required
	app.Get("/api/v1/promotions", server.listPromotions)            //? no auth required

	//* Product-Categories
	app.Get("/api/v1/categories/:categoryId", server.getProductCategory) //? no auth required
	app.Get("/api/v1/categories", server.listProductCategories)          //? no auth required

	//* Product-Brands
	app.Get("/api/v1/brands/:brandId", server.getProductBrand) //? no auth required
	app.Get("/api/v1/brands", server.listProductBrands)        //? no auth required

	//* Product-Sizes
	app.Get("/api/v1/sizes", server.listProductSizes) //? no auth required

	//* Product-Colors
	app.Get("/api/v1/colors", server.listProductColors) //? no auth required

	//* Product-Images
	app.Get("/api/v1/images-v2", server.listProductImagesV2)              //? no auth required
	app.Get("/api/v1/images-next-page", server.listProductImagesNextPage) //? no auth required

	//* Products-Promotions
	app.Get("/api/v1/product-promotions/:promotionId/products/:productId", server.getProductPromotion) //? no auth required
	app.Get("/api/v1/product-promotions", server.listProductPromotions)                                //? no auth required
	app.Get("/api/v1/product-promotions-images", server.listProductPromotionsWithImages)               //? no auth required

	//* Category-Promotions
	app.Get("/api/v1/category-promotions/:promotionId/categories/:categoryId", server.getCategoryPromotion) //? no auth required
	app.Get("/api/v1/category-promotions", server.listCategoryPromotions)                                   //? no auth required
	app.Get("/api/v1/category-promotions-images", server.listCategoryPromotionsWithImages)                  //? no auth required

	//* Brand-Promotions
	app.Get("/api/v1/brand-promotions/:promotionId/brands/:brandId", server.getBrandPromotion) //? no auth required
	app.Get("/api/v1/brand-promotions", server.listBrandPromotions)                            //? no auth required
	app.Get("/api/v1/brand-promotions-images", server.listBrandPromotionsWithImages)           //? no auth required

	//* Variations
	app.Get("/api/v1/variations/:variationId", server.getVariation) //? no auth required
	app.Get("/api/v1/variations", server.listVariations)            //? no auth required

	//* Variation-Options
	app.Get("/api/v1/variation-options/:id", server.getVariationOption) //? no auth required
	app.Get("/api/v1/variation-options", server.listVariationOptions)   //? no auth required

	//* Product-Items
	app.Get("/api/v1/product-items/:itemId", server.getProductItem)                                                            //? no auth required
	app.Get("/api/v1/product-items", server.listProductItems)                                                                  //? no auth required
	app.Get("/api/v1/product-items-v2", server.listProductItemsV2)                                                             //? no auth required
	app.Get("/api/v1/product-items-next-page", server.listProductItemsNextPage)                                                //? no auth required
	app.Get("/api/v1/search-product-items", server.searchProductItems)                                                         //? no auth required
	app.Get("/api/v1/search-product-items-next-page", server.searchProductItemsNextPage)                                       //? no auth required
	app.Get("/api/v1/product-items-with-promotions", server.listProductItemsWithPromotions)                                    //? no auth required
	app.Get("/api/v1/product-items-with-promotions-next-page", server.listProductItemsWithPromotionsNextPage)                  //? no auth required
	app.Get("/api/v1/product-items-with-brand-promotions", server.listProductItemsWithBrandPromotions)                         //? no auth required
	app.Get("/api/v1/product-items-with-brand-promotions-next-page", server.listProductItemsWithBrandPromotionsNextPage)       //? no auth required
	app.Get("/api/v1/product-items-with-category-promotions", server.listProductItemsWithCategoryPromotions)                   //? no auth required
	app.Get("/api/v1/product-items-with-category-promotions-next-page", server.listProductItemsWithCategoryPromotionsNextPage) //? no auth required

	//* Product-Configuration
	app.Get("/api/v1/product-configurations/:itemId/variation-options/:variationId", server.getProductConfiguration) //? no auth required
	app.Get("/api/v1/product-configurations/:itemId", server.listProductConfigurations)                              //? no auth required

	userRouter := app.Group("/usr/v1").Use(authMiddleware(server.userTokenMaker, false))
	adminRouter := app.Group("/admin/v1").Use(authMiddleware(server.adminTokenMaker, true)) //! For Admin Only

	userRouter.Get("/users/:id", server.getUser)

	// app.Use(gofiberfirebaseauth.New(
	// 	gofiberfirebaseauth.Config{
	// 		FirebaseApp: fireApp,
	// 	}))

	adminRouter.Post("/admins/:adminId/product-images", server.createProductImages)    //! Admin Only
	adminRouter.Get("/admins/:adminId/product-images/kit", server.listproductImages)   //! Admin Only
	adminRouter.Put("/admins/:adminId/product-images/:id", server.updateProductImages) //! Admin Only

	//* dashboard
	adminRouter.Get("/admins/:adminId/dashboard", server.getDashboardInfo) //! Admin Only

	userRouter.Post("/users/:id/notification", server.createNotification)
	userRouter.Get("/users/:id/notification/:deviceId", server.getNotification)
	userRouter.Put("/users/:id/notification/:deviceId", server.updateNotification)
	userRouter.Delete("/users/:id/notification/:deviceId", server.deleteNotification)

	adminRouter.Post("/admins/:adminId/app-policy", server.createAppPolicy)       //! Admin Only
	adminRouter.Put("/admins/:adminId/app-policy/:id", server.updateAppPolicy)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/app-policy/:id", server.deleteAppPolicy) //! Admin Only

	adminRouter.Get("/admins/:adminId/users", server.listUsers) //! Admin Only
	userRouter.Put("/users/:id", server.updateUser)
	adminRouter.Delete("/admins/:adminId/users/:id", server.deleteUser) //! Admin Only
	userRouter.Delete("/users/:id/logout", server.logoutUser)

	adminRouter.Delete("/admins/:id/logout", server.logoutAdmin) //! Admin Only

	userRouter.Post("/users/:id/addresses", server.createUserAddress)
	userRouter.Get("/users/:id/addresses/:addressId", server.getUserAddress)
	userRouter.Get("/users/:id/addresses", server.listUserAddresses)
	userRouter.Put("/users/:id/addresses/:addressId", server.updateUserAddress)
	userRouter.Delete("/users/:id/addresses/:addressId", server.deleteUserAddress)

	userRouter.Post("/users/:id/reviews", server.createUserReview)
	userRouter.Get("/users/:id/reviews/:reviewId", server.getUserReview)
	userRouter.Get("/users/:id/reviews", server.listUserReviews)
	userRouter.Put("/users/:id/reviews/:reviewId", server.updateUserReview)
	userRouter.Delete("/users/:id/reviews/:reviewId", server.deleteUserReview)

	//? /items is shoppingCartItems ID in the Table
	userRouter.Post("/users/:id/carts/:cartId/items", server.createShoppingCartItem)
	userRouter.Get("/users/:id/carts/:cartId/items", server.getShoppingCartItem)
	userRouter.Get("/users/:id/carts/items", server.listShoppingCartItems)
	userRouter.Put("/users/:id/carts/:cartId/items/:itemId", server.updateShoppingCartItem)
	userRouter.Delete("/users/:id/carts/:cartId/items/:itemId", server.deleteShoppingCartItem)
	userRouter.Delete("/users/:id/carts/:cartId", server.deleteShoppingCartItemAllByUser)
	userRouter.Put("/users/:id/carts/:cartId/purchase", server.finishPurchase)

	//? /items is WishListItems ID in the Table
	userRouter.Post("/users/:id/wish-lists/:wishId/items", server.createWishListItem)
	userRouter.Get("/users/:id/wish-lists/:wishId/items/:itemId", server.getWishListItem)
	userRouter.Get("/users/:id/wish-lists/items", server.listWishListItems)
	userRouter.Put("/users/:id/wish-lists/:wishId/items/:itemId", server.updateWishListItem)
	userRouter.Delete("/users/:id/wish-lists/:wishId/items/:itemId", server.deleteWishListItem)
	userRouter.Delete("/users/:id/wish-lists/:wishId", server.deleteWishListItemAll)

	userRouter.Post("/users/:id/payment-methods", server.createPaymentMethod)
	userRouter.Get("/users/:id/payment-method", server.getPaymentMethod)
	userRouter.Get("/users/:id/payment-methods", server.listPaymentMethods)
	userRouter.Put("/users/:id/payment-methods/:paymentId", server.updatePaymentMethod)
	userRouter.Delete("/users/:id/payment-methods/:paymentId", server.deletePaymentMethod)

	userRouter.Get("/users/:id/payment-types", server.listPaymentTypes)

	adminRouter.Post("/admins/:adminId/text-banners", server.createHomePageTextBanner)                 //! Admin Only
	adminRouter.Put("/admins/:adminId/text-banners/:textBannerId", server.updateHomePageTextBanner)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/text-banners/:textBannerId", server.deleteHomePageTextBanner) //! Admin Only

	adminRouter.Post("/admins/:adminId/products", server.createProduct)              //! Admin Only
	adminRouter.Put("/admins/:adminId/products/:productId", server.updateProduct)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/products/:productId", server.deleteProduct) //! Admin Only

	adminRouter.Post("/admins/:adminId/promotions", server.createPromotion)                //! Admin Only
	adminRouter.Put("/admins/:adminId/promotions/:promotionId", server.updatePromotion)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/promotions/:promotionId", server.deletePromotion) //! Admin Only

	adminRouter.Post("/admins/:adminId/categories", server.createProductCategory)               //! Admin Only
	adminRouter.Put("/admins/:adminId/categories/:categoryId", server.updateProductCategory)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/categories/:categoryId", server.deleteProductCategory) //! Admin Only

	adminRouter.Post("/admins/:adminId/colors", server.createProductColor)    //! Admin Only
	adminRouter.Put("/admins/:adminId/colors/:id", server.updateProductColor) //! Admin Only

	adminRouter.Post("/admins/:adminId/sizes", server.createProductSize)    //! Admin Only
	adminRouter.Put("/admins/:adminId/sizes/:id", server.updateProductSize) //! Admin Only

	adminRouter.Post("/admins/:adminId/brands", server.createProductBrand)            //! Admin Only
	adminRouter.Put("/admins/:adminId/brands/:brandId", server.updateProductBrand)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/brands/:brandId", server.deleteProductBrand) //! Admin Only

	adminRouter.Get("/admins/:adminId/product-promotions", server.listProductPromotionsForAdmins)                             //! Admin Only
	adminRouter.Post("/admins/:adminId/product-promotions", server.createProductPromotion)                                    //! Admin Only
	adminRouter.Put("/admins/:adminId/product-promotions/:promotionId/products/:productId", server.updateProductPromotion)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/product-promotions/:promotionId/products/:productId", server.deleteProductPromotion) //! Admin Only

	adminRouter.Get("/admins/:adminId/category-promotions", server.listCategoryPromotionsForAdmins)                                //! Admin Only
	adminRouter.Post("/admins/:adminId/category-promotions", server.createCategoryPromotion)                                       //! Admin Only
	adminRouter.Put("/admins/:adminId/category-promotions/:promotionId/categories/:categoryId", server.updateCategoryPromotion)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/category-promotions/:promotionId/categories/:categoryId", server.deleteCategoryPromotion) //! Admin Only

	adminRouter.Get("/admins/:adminId/brand-promotions", server.listBrandPromotionsForAdmins)                         //! Admin Only
	adminRouter.Post("/admins/:adminId/brand-promotions", server.createBrandPromotion)                                //! Admin Only
	adminRouter.Put("/admins/:adminId/brand-promotions/:promotionId/brands/:brandId", server.updateBrandPromotion)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/brand-promotions/:promotionId/brands/:brandId", server.deleteBrandPromotion) //! Admin Only

	adminRouter.Post("/admins/:adminId/variations", server.createVariation)                //! Admin Only
	adminRouter.Put("/admins/:adminId/variations/:variationId", server.updateVariation)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/variations/:variationId", server.deleteVariation) //! Admin Only

	adminRouter.Post("/admins/:adminId/variation-options", server.createVariationOption)       //! Admin Only
	adminRouter.Put("/admins/:adminId/variation-options/:id", server.updateVariationOption)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/variation-options/:id", server.deleteVariationOption) //! Admin Only

	adminRouter.Post("/admins/:adminId/product-items", server.createProductItem)           //! Admin Only
	adminRouter.Put("/admins/:adminId/product-items/:itemId", server.updateProductItem)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/product-items/:itemId", server.deleteProductItem) //! Admin Only

	adminRouter.Post("/admins/:adminId/product-configurations/:itemId", server.createProductConfiguration)                                  //! Admin Only
	adminRouter.Put("/admins/:adminId/product-configurations/:itemId", server.updateProductConfiguration)                                   //! Admin Only
	adminRouter.Delete("/admins/:adminId/product-configurations/:itemId/variation-options/:variationId", server.deleteProductConfiguration) //! Admin Only

	//? ShopOrderItems
	userRouter.Get("/users/:id/shop-order-items/:orderId", server.getShopOrderItems)
	userRouter.Get("/users/:id/shop-order-items", server.listShopOrderItems)

	adminRouter.Get("/admins/:adminId/shop-order-items/:orderId", server.getShopOrderItemsForAdmin) //! Admin Only
	adminRouter.Delete("/admins/:adminId/shop-order-items/:id", server.deleteShopOrderItem)         //! Admin Only

	//? ShopOrders
	userRouter.Get("/users/:id/shop-orders", server.listShopOrders)
	userRouter.Get("/users/:id/shop-orders-v2", server.listShopOrdersV2)
	userRouter.Get("/users/:id/shop-orders-next-page", server.listShopOrdersNextPage)

	adminRouter.Get("/admins/:adminId/shop-orders-v2", server.listShopOrdersV2ForAdmin)              //! Admin Only
	adminRouter.Get("/admins/:adminId/shop-orders-next-page", server.listShopOrdersNextPageForAdmin) //! Admin Only
	adminRouter.Put("/admins/:adminId/shop-orders/:shopOrderId", server.updateShopOrder)             //! Admin Only

	userRouter.Post("/users/:id/shipping-method", server.createShippingMethod)
	userRouter.Get("/users/:id/shipping-method/:methodId", server.getShippingMethod)
	userRouter.Get("/users/:id/shipping-method", server.listShippingMethods)
	userRouter.Put("/users/:id/shipping-method/:methodId", server.updateShippingMethod)
	adminRouter.Delete("/admins/:adminId/shipping-method/:methodId", server.deleteShippingMethod) //! Admin Only

	adminRouter.Post("/admins/:adminId/order-status", server.createOrderStatus) //! Admin Only
	userRouter.Get("/users/:id/order-status/:statusId", server.getOrderStatus)
	userRouter.Get("/users/:id/order-status", server.listOrderStatuses)
	adminRouter.Get("/admins/:adminId/order-status", server.listOrderStatusesForAdmin)      //! Admin Only
	adminRouter.Put("/admins/:adminId/order-status/:statusId", server.updateOrderStatus)    //! Admin Only
	adminRouter.Delete("/admins/:adminId/order-status/:statusId", server.deleteOrderStatus) //! Admin Only

	server.router = app

}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Listen(address)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{"error": err.Error()}
}

func (server *Server) gracefulShutdown() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Println("Shutdown server...")
	if err := server.router.Shutdown(); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

}
