package api

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/sonic"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *fiber.App
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	go server.gracefulShutdown()
	return server, nil
}

func (server *Server) setupRouter() {
	app := fiber.New(
		fiber.Config{
			JSONEncoder: sonic.Marshal,
			JSONDecoder: sonic.Unmarshal,
		},
	)

	app.Use(etag.New(
		etag.Config{},
	))

	//* Users
	app.Post("/api/v1/users", server.createUser)
	app.Post("/api/v1/users/login", server.loginUser)
	app.Post("/api/v1/users/reset-password", server.resetPassword)

	//* Tokens
	app.Post("/api/v1/auth/access-token", server.renewAccessToken)

	//*Products
	app.Get("/api/v1/products/:productId", server.getProduct) //? no auth required
	app.Get("/api/v1/products", server.listProducts)          //? no auth required

	//*Promotions
	app.Get("/api/v1/promotions/:promotionId", server.getPromotion) //? no auth required
	app.Get("/api/v1/promotions", server.listPromotions)            //? no auth required

	//* Product-Categories
	app.Get("/api/v1/categories/:categoryId", server.getProductCategory) //? no auth required
	app.Get("/api/v1/categories", server.listProductCategories)          //? no auth required

	//* Product-Brands
	app.Get("/api/v1/brands/:brandId", server.getProductBrand) //? no auth required
	app.Get("/api/v1/brands", server.listProductBrands)        //? no auth required

	//* Products-Promotions
	app.Get("/api/v1/product-promotions/:promotionId/products/:productId", server.getProductPromotion) //? no auth required
	app.Get("/api/v1/product-promotions/products", server.listProductPromotions)                       //? no auth required

	//* Category-Promotions
	app.Get("/api/v1/category-promotions/:promotionId/categories/:categoryId", server.getCategoryPromotion) //? no auth required
	app.Get("/api/v1/category-promotions/categories", server.listCategoryPromotions)                        //? no auth required

	//* Brand-Promotions
	app.Get("/api/v1/brand-promotions/:promotionId/brands/:brandId", server.getBrandPromotion) //? no auth required
	app.Get("/api/v1/brand-promotions/brands", server.listBrandPromotions)                     //? no auth required

	//* Variations
	app.Get("/api/v1/variations/:variationId", server.getVariation) //? no auth required
	app.Get("/api/v1/variations", server.listVariations)            //? no auth required

	//* Variation-Options
	app.Get("/api/v1/variation-options/:id", server.getVariationOption) //? no auth required
	app.Get("/api/v1/variation-options", server.listVariationOptions)   //? no auth required

	//* Product-Items
	app.Get("/api/v1/product-items/:itemId", server.getProductItem)                      //? no auth required
	app.Get("/api/v1/product-items", server.listProductItems)                            //? no auth required
	app.Get("/api/v1/product-items-v2", server.listProductItemsV2)                       //? no auth required
	app.Get("/api/v1/product-items-next-page", server.listProductItemsNextPage)          //? no auth required
	app.Get("/api/v1/search-product-items", server.searchProductItems)                   //? no auth required
	app.Get("/api/v1/search-product-items-next-page", server.searchProductItemsNextPage) //? no auth required

	//* Product-Configuration
	app.Get("/api/v1/product-configurations/:itemId/variation-options/:variationId", server.getProductConfiguration) //? no auth required
	app.Get("/api/v1/product-configurations/:itemId", server.listProductConfigurations)                              //? no auth required

	userRouter := app.Group("/usr/v1").Use(authMiddleware(server.tokenMaker, false))
	adminRouter := app.Group("/admin/:adminId/v1").Use(authMiddleware(server.tokenMaker, true))

	userRouter.Get("/users/:id", server.getUser)
	adminRouter.Get("/users", server.listUsers) //! Admin Only
	userRouter.Put("/users/:id", server.updateUser)
	adminRouter.Delete("/users/:id", server.deleteUser) //! Admin Only
	userRouter.Delete("/users/:id/logout", server.logoutUser)

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

	adminRouter.Post("/products", server.createProduct)              //! Admin Only
	adminRouter.Put("/products/:productId", server.updateProduct)    //! Admin Only
	adminRouter.Delete("/products/:productId", server.deleteProduct) //! Admin Only

	adminRouter.Post("/promotions", server.createPromotion)                //! Admin Only
	adminRouter.Put("/promotions/:promotionId", server.updatePromotion)    //! Admin Only
	adminRouter.Delete("/promotions/:promotionId", server.deletePromotion) //! Admin Only

	adminRouter.Post("/categories", server.createProductCategory)               //! Admin Only
	adminRouter.Put("/categories/:categoryId", server.updateProductCategory)    //! Admin Only
	adminRouter.Delete("/categories/:categoryId", server.deleteProductCategory) //! Admin Only

	adminRouter.Post("/brands", server.createProductBrand)            //! Admin Only
	adminRouter.Put("/brands/:brandId", server.updateProductBrand)    //! Admin Only
	adminRouter.Delete("/brands/:brandId", server.deleteProductBrand) //! Admin Only

	adminRouter.Post("/product-promotions/:promotionId/products/:productId", server.createProductPromotion)   //! Admin Only
	adminRouter.Put("/product-promotions/:promotionId/products/:productId", server.updateProductPromotion)    //! Admin Only
	adminRouter.Delete("/product-promotions/:promotionId/products/:productId", server.deleteProductPromotion) //! Admin Only

	adminRouter.Post("/category-promotions/:promotionId/categories/:categoryId", server.createCategoryPromotion)   //! Admin Only
	adminRouter.Put("/category-promotions/:promotionId/categories/:categoryId", server.updateCategoryPromotion)    //! Admin Only
	adminRouter.Delete("/category-promotions/:promotionId/categories/:categoryId", server.deleteCategoryPromotion) //! Admin Only

	adminRouter.Post("/brand-promotions/:promotionId/brands/:brandId", server.createBrandPromotion)   //! Admin Only
	adminRouter.Put("/brand-promotions/:promotionId/brands/:brandId", server.updateBrandPromotion)    //! Admin Only
	adminRouter.Delete("/brand-promotions/:promotionId/brands/:brandId", server.deleteBrandPromotion) //! Admin Only

	adminRouter.Post("/variations", server.createVariation)                //! Admin Only
	adminRouter.Put("/variations/:variationId", server.updateVariation)    //! Admin Only
	adminRouter.Delete("/variations/:variationId", server.deleteVariation) //! Admin Only

	adminRouter.Post("/variation-options", server.createVariationOption)       //! Admin Only
	adminRouter.Put("/variation-options/:id", server.updateVariationOption)    //! Admin Only
	adminRouter.Delete("/variation-options/:id", server.deleteVariationOption) //! Admin Only

	adminRouter.Post("/product-items", server.createProductItem)           //! Admin Only
	adminRouter.Put("/product-items/:itemId", server.updateProductItem)    //! Admin Only
	adminRouter.Delete("/product-items/:itemId", server.deleteProductItem) //! Admin Only

	adminRouter.Post("/product-configurations/:itemId", server.createProductConfiguration)                                  //! Admin Only
	adminRouter.Put("/product-configurations/:itemId", server.updateProductConfiguration)                                   //! Admin Only
	adminRouter.Delete("/product-configurations/:itemId/variation-options/:variationId", server.deleteProductConfiguration) //! Admin Only

	userRouter.Get("/users/:id/shop-order-items/:orderId", server.getShopOrderItems)
	userRouter.Get("/users/:id/shop-order-items", server.listShopOrderItems)

	userRouter.Get("/users/:id/shop-orders", server.listShopOrders)

	userRouter.Post("/users/:id/order-status", server.createOrderStatus)
	userRouter.Get("/users/:id/order-status/:statusId", server.getOrderStatus)
	userRouter.Get("/users/:id/order-status", server.listOrderStatuses)
	userRouter.Put("/users/:id/order-status/:statusId", server.updateOrderStatus)
	adminRouter.Delete("/order-status/:statusId", server.deleteOrderStatus) //! Admin Only

	userRouter.Post("/users/:id/shipping-method", server.createShippingMethod)
	userRouter.Get("/users/:id/shipping-method/:methodId", server.getShippingMethod)
	userRouter.Get("/users/:id/shipping-method", server.listShippingMethods)
	userRouter.Put("/users/:id/shipping-method/:methodId", server.updateShippingMethod)
	adminRouter.Delete("/shipping-method/:methodId", server.deleteShippingMethod) //! Admin Only

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

	_ = <-done
	log.Println("Shutdown server...")
	if err := server.router.Shutdown(); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

}
