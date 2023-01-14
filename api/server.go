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
	//* Users
	app.Post("/api/v1/users", server.createUser)
	app.Post("/api/v1/users/login", server.loginUser)
	app.Post("/api/v1/users/reset-password", server.resetPassword)

	//* Tokens
	app.Post("/api/v1/tokens/renew-access", server.renewAccessToken)

	//*Products
	app.Get("/api/v1/products/:productId", server.getProduct) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/products", server.listProducts)          //? no auth required # Finished With tests (token and changed response.)

	//*Promotions
	app.Get("/api/v1/promotions/:promotionId", server.getPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/promotions", server.listPromotions)            //? no auth required # Finished With tests (token and changed response.)

	//* Product-Categories
	app.Get("/api/v1/categories/:categoryId", server.getProductCategory) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/categories", server.listProductCategories)          //? no auth required # Finished With tests (token and changed response.)

	//* Products-Promotions
	app.Get("/api/v1/product-promotions/:promotionId/products/:productId", server.getProductPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-promotions/products", server.listProductPromotions)                       //? no auth required # Finished With tests (token and changed response.)

	//* Category-Promotions
	app.Get("/api/v1/category-promotions/:promotionId/categories/:categoryId", server.getCategoryPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/category-promotions/categories", server.listCategoryPromotions)                        //? no auth required # Finished With tests (token and changed response.)

	//* Variations
	app.Get("/api/v1/variations/:variationId", server.getVariation) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/variations", server.listVariations)            //? no auth required # Finished With tests (token and changed response.)

	//* Variation-Options
	app.Get("/api/v1/variation-options/:id", server.getVariationOption) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/variation-options", server.listVariationOptions)   //? no auth required # Finished With tests (token and changed response.)

	//* Product-Items
	app.Get("/api/v1/product-items/:itemId", server.getProductItem) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-items", server.listProductItems)       //? no auth required # Finished With tests (token and changed response.)

	//* Product-Configuration
	app.Get("/api/v1/product-configurations/:itemId/variation-options/:variationId", server.getProductConfiguration) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-configurations/:itemId", server.listProductConfigurations)                              //? no auth required # Finished With tests (token and changed response.)

	userRouter := app.Group("/api/v1").Use(authMiddleware(server.tokenMaker, false))
	adminRouter := app.Group("/api/admin/:adminId/v1").Use(authMiddleware(server.tokenMaker, true))

	userRouter.Get("/users/:id", server.getUser)        //* Finished With tests (token and changed response... No Etag)
	adminRouter.Get("/users", server.listUsers)         //! Admin Only # Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id", server.updateUser)     //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/users/:id", server.deleteUser) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/addresses", server.createUserAddress)              //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/addresses/:addressId", server.getUserAddress)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/addresses", server.listUserAddresses)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/addresses/:addressId", server.updateUserAddress)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/addresses/:addressId", server.deleteUserAddress) //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/reviews", server.createUserReview)             //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/reviews/:reviewId", server.getUserReview)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/reviews", server.listUserReviews)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/reviews/:reviewId", server.updateUserReview)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/reviews/:reviewId", server.deleteUserReview) //* Finished With tests (token and changed response... No Etag)

	//? /items is shoppingCartItems ID in the Table
	userRouter.Post("/users/:id/carts/:cartId/items", server.createShoppingCartItem)           //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/carts/:cartId/items", server.getShoppingCartItem)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/carts/items", server.listShoppingCartItems)                     //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/carts/:cartId/items/:itemId", server.updateShoppingCartItem)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/carts/:cartId/items/:itemId", server.deleteShoppingCartItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/carts/:cartId", server.deleteShoppingCartItemAllByUser)      //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/carts/:cartId/purchase", server.finishPurchase)                 //* Finished With tests (token and changed response... No Etag)

	//? /items is WishListItems ID in the Table
	userRouter.Post("/users/:id/wish-lists/:wishId", server.createWishListItem)                 //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/wish-lists/:wishId/items/:itemId", server.getWishListItem)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/wish-lists/:wishId", server.listWishListItems)                   //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/wish-lists/:wishId/items/:itemId", server.updateWishListItem)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/wish-lists/:wishId/items/:itemId", server.deleteWishListItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/wish-lists/:wishId", server.deleteWishListItemAll)            //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/payment-methods", server.createPaymentMethod)              //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/payment-methods/:paymentId", server.getPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/payment-methods", server.listPaymentMethodes)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/payment-methods/:paymentId", server.updatePaymentMethod)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/payment-methods/:paymentId", server.deletePaymentMethod) //* Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/products", server.createProduct)              //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/products/:productId", server.updateProduct)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/products/:productId", server.deleteProduct) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/promotions", server.createPromotion)                //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/promotions/:promotionId", server.updatePromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/promotions/:promotionId", server.deletePromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/categories", server.createProductCategory)               //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/categories/:categoryId", server.updateProductCategory)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/categories/:categoryId", server.deleteProductCategory) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-promotions/:promotionId/products/:productId", server.createProductPromotion)   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-promotions/:promotionId/products/:productId", server.updateProductPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-promotions/:promotionId/products/:productId", server.deleteProductPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/category-promotions/:promotionId/categories/:categoryId", server.createCategoryPromotion)   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/category-promotions/:promotionId/categories/:categoryId", server.updateCategoryPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/category-promotions/:promotionId/categories/:categoryId", server.deleteCategoryPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/variations", server.createVariation)                //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/variations/:variationId", server.updateVariation)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/variations/:variationId", server.deleteVariation) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/variation-options", server.createVariationOption)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/variation-options/:id", server.updateVariationOption)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/variation-options/:id", server.deleteVariationOption) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-items", server.createProductItem)           //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-items/:itemId", server.updateProductItem)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-items/:itemId", server.deleteProductItem) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-configurations/:itemId", server.createProductConfiguration)                                  //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-configurations/:itemId", server.updateProductConfiguration)                                   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-configurations/:itemId/variation-options/:variationId", server.deleteProductConfiguration) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Get("/users/:id/shop-orders/:orderId", server.getShopOrderItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shop-orders", server.listShopOrderItems)        //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/order-status", server.createOrderStatus)          //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/order-status/:statusId", server.getOrderStatus)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/order-status", server.listOrderStatuses)           //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/order-status/:statusId", server.updateOrderStatus) //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/order-status/:statusId", server.deleteOrderStatus)       //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/shipping-method", server.createShippingMethod)          //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shipping-method/:methodId", server.getShippingMethod)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shipping-method", server.listShippingMethodes)           //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/shipping-method/:methodId", server.updateShippingMethod) //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/shipping-method/:methodId", server.deleteShippingMethod)       //! Admin Only # Finished With tests (token and changed response... No Etag)

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
