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
	app.Get("/api/v1/products/:product_id", server.getProduct) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/products", server.listProducts)           //? no auth required # Finished With tests (token and changed response.)

	//*Promotions
	app.Get("/api/v1/promotions/:promotion_id", server.getPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/promotions", server.listPromotions)             //? no auth required # Finished With tests (token and changed response.)

	//* Product-Categories
	app.Get("/api/v1/categories/:category_id", server.getProductCategory) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/categories", server.listProductCategories)           //? no auth required # Finished With tests (token and changed response.)

	//* Products-Promotions
	app.Get("/api/v1/product-promotions/:promotion_id/products/:product_id", server.getProductPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-promotions/products", server.listProductPromotions)                         //? no auth required # Finished With tests (token and changed response.)

	//* Category-Promotions
	app.Get("/api/v1/category-promotions/:promotion_id/categories/:category_id", server.getCategoryPromotion) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/category-promotions/categories", server.listCategoryPromotions)                          //? no auth required # Finished With tests (token and changed response.)

	//* Variations
	app.Get("/api/v1/variations/:variation_id", server.getVariation) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/variations", server.listVariations)             //? no auth required # Finished With tests (token and changed response.)

	//* Variation-Options
	app.Get("/api/v1/variation-options/:id", server.getVariationOption) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/variation-options", server.listVariationOptions)   //? no auth required # Finished With tests (token and changed response.)

	//* Product-Items
	app.Get("/api/v1/product-items/:item_id", server.getProductItem) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-items", server.listProductItems)        //? no auth required # Finished With tests (token and changed response.)

	//* Product-Configuration
	app.Get("/api/v1/product-configurations/:item_id/variation-options/:variation_id", server.getProductConfiguration) //? no auth required # Finished With tests (token and changed response... No Etag)
	app.Get("/api/v1/product-configurations/:item_id", server.listProductConfigurations)                               //? no auth required # Finished With tests (token and changed response.)

	userRouter := app.Group("/api/v1").Use(authMiddleware(server.tokenMaker, false))
	adminRouter := app.Group("/api/admin/:admin_id/v1").Use(authMiddleware(server.tokenMaker, true))

	userRouter.Get("/users/:id", server.getUser)        //* Finished With tests (token and changed response... No Etag)
	adminRouter.Get("/users", server.listUsers)         //! Admin Only # Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id", server.updateUser)     //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/users/:id", server.deleteUser) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/addresses", server.createUserAddress)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/addresses/:address_id", server.getUserAddress)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/addresses", server.listUserAddresses)                //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/addresses/:address_id", server.updateUserAddress)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/addresses/:address_id", server.deleteUserAddress) //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/reviews", server.createUserReview)              //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/reviews/:review_id", server.getUserReview)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/reviews", server.listUserReviews)                //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/reviews/:review_id", server.updateUserReview)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/reviews/:review_id", server.deleteUserReview) //* Finished With tests (token and changed response... No Etag)

	//? /items is shoppingCartItems ID in the Table
	userRouter.Post("/users/:id/carts/:cart_id/items", server.createShoppingCartItem)            //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/carts/:cart_id/items", server.getShoppingCartItem)                //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/carts/items", server.listShoppingCartItems)                       //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/carts/:cart_id/items/:item_id", server.updateShoppingCartItem)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/carts/:cart_id/items/:item_id", server.deleteShoppingCartItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/carts/:cart_id", server.deleteShoppingCartItemAllByUser)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/carts/:cart_id/purchase", server.finishPurchase)                  //* Finished With tests (token and changed response... No Etag)

	//? /items is WishListItems ID in the Table
	userRouter.Post("/users/:id/wish-lists/:wish_id", server.createWishListItem)                  //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/wish-lists/:wish_id/items/:item_id", server.getWishListItem)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/wish-lists/:wish_id", server.listWishListItems)                    //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/wish-lists/:wish_id/items/:item_id", server.updateWishListItem)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/wish-lists/:wish_id/items/:item_id", server.deleteWishListItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/wish-lists/:wish_id", server.deleteWishListItemAll)             //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/payment-methods", server.createPaymentMethod)               //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/payment-methods/:payment_id", server.getPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/payment-methods", server.listPaymentMethodes)                //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/payment-methods/:payment_id", server.updatePaymentMethod)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Delete("/users/:id/payment-methods/:payment_id", server.deletePaymentMethod) //* Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/products", server.createProduct)               //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/products/:product_id", server.updateProduct)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/products/:product_id", server.deleteProduct) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/promotions", server.createPromotion)                 //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/promotions/:promotion_id", server.updatePromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/promotions/:promotion_id", server.deletePromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/categories", server.createProductCategory)                //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/categories/:category_id", server.updateProductCategory)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/categories/:category_id", server.deleteProductCategory) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-promotions/:promotion_id/products/:product_id", server.createProductPromotion)   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-promotions/:promotion_id/products/:product_id", server.updateProductPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-promotions/:promotion_id/products/:product_id", server.deleteProductPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/category-promotions/:promotion_id/categories/:category_id", server.createCategoryPromotion)   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/category-promotions/:promotion_id/categories/:category_id", server.updateCategoryPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/category-promotions/:promotion_id/categories/:category_id", server.deleteCategoryPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/variations", server.createVariation)                 //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/variations/:variation_id", server.updateVariation)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/variations/:variation_id", server.deleteVariation) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/variation-options", server.createVariationOption)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/variation-options/:id", server.updateVariationOption)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/variation-options/:id", server.deleteVariationOption) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-items", server.createProductItem)            //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-items/:item_id", server.updateProductItem)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-items/:item_id", server.deleteProductItem) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRouter.Post("/product-configurations/:item_id", server.createProductConfiguration)                                   //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Put("/product-configurations/:item_id", server.updateProductConfiguration)                                    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/product-configurations/:item_id/variation-options/:variation_id", server.deleteProductConfiguration) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Get("/users/:id/shop-orders/:order_id", server.getShopOrderItem) //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shop-orders", server.listShopOrderItems)         //* Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/order-status", server.createOrderStatus)           //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/order-status/:status_id", server.getOrderStatus)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/order-status", server.listOrderStatuses)            //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/order-status/:status_id", server.updateOrderStatus) //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/order-status/:status_id", server.deleteOrderStatus)       //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRouter.Post("/users/:id/shipping-method", server.createShippingMethod)           //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shipping-method/:method_id", server.getShippingMethod)    //* Finished With tests (token and changed response... No Etag)
	userRouter.Get("/users/:id/shipping-method", server.listShippingMethodes)            //* Finished With tests (token and changed response... No Etag)
	userRouter.Put("/users/:id/shipping-method/:method_id", server.updateShippingMethod) //* Finished With tests (token and changed response... No Etag)
	adminRouter.Delete("/shipping-method/:method_id", server.deleteShippingMethod)       //! Admin Only # Finished With tests (token and changed response... No Etag)

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
