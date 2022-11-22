package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
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
	go server.gracefullShutDown(server.router)
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/users/reset_password", server.resetPassword)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	userRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, false))
	adminRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, true))

	userRoutes.GET("/users/:id", server.getUser)        //* Finished With tests (token and changed response... No Etag)
	adminRoutes.GET("/users", server.listUsers)         //! Admin Only # Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/:id", server.updateUser)     //* Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/users/:id", server.deleteUser) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/addresses", server.createUserAddress)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/addresses/:id", server.getUserAddress)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/addresses", server.listUserAddresses)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/addresses/:id", server.updateUserAddress)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/addresses/:id", server.deleteUserAddress) //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/reviews", server.createUserReview)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/reviews/:id", server.getUserReview)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/reviews", server.listUserReviews)         //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/reviews/:id", server.updateUserReview)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/reviews/:id", server.deleteUserReview) //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/cart", server.createShoppingCartItem)                          //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/cart/:shopping-cart-id", server.getShoppingCartItem)            //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/cart", server.listShoppingCartItems)                            //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/cart/:shopping-cart-id", server.updateShoppingCartItem)         //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/cart/:shopping-cart-item-id", server.deleteShoppingCartItem) //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/cart/delete-all", server.deleteShoppingCartItemAllByUser)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/cart/purchase", server.finishPurchase)                          //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/wish-list", server.createWishListItem)                       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/wish-list/:id", server.getWishListItem)                       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/wish-list", server.listWishListItems)                         //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/wish-list/:id", server.updateWishListItem)                    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/wish-list/:wish-list-item-id", server.deleteWishListItem)  //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/wish-list/delete-all", server.deleteWishListItemAllByUser) //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/payment-method", server.createPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/payment-method/:id", server.getPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/payment-method", server.listPaymentMethodes)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/payment-method/:id", server.updatePaymentMethod)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/payment-method/:id", server.deletePaymentMethod) //* Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/products", server.createProduct)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/products/:id", server.getProduct)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/products", server.listProducts)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/products/:id", server.updateProduct)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/products/:id", server.deleteProduct) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/promotions", server.createPromotion)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/promotions/:id", server.getPromotion)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/promotions", server.listPromotions)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/promotions/:id", server.updatePromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/promotions/:id", server.deletePromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/product-categories", server.createProductCategory)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/product-categories/:id", server.getProductCategory)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/product-categories", server.listProductCategories)             //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/product-categories/:id", server.updateProductCategory)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/product-categories/:id", server.deleteProductCategory) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/product-promotions", server.createProductPromotion)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/product-promotions/:id", server.getProductPromotion)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/product-promotions", server.listProductPromotions)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/product-promotions/:id", server.updateProductPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/product-promotions/:id", server.deleteProductPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/category-promotions", server.createCategoryPromotion)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/category-promotions/:id", server.getCategoryPromotion)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/category-promotions", server.listCategoryPromotions)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/category-promotions/:id", server.updateCategoryPromotion)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/category-promotions/:id", server.deleteCategoryPromotion) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/variations", server.createVariation)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/variations/:id", server.getVariation)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/variations", server.listVariations)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/variations/:id", server.updateVariation)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/variations/:id", server.deleteVariation) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/variation-options", server.createVariationOption)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/variation-options/:id", server.getVariationOption)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/variation-options", server.listVariationOptions)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/variation-options/:id", server.updateVariationOption)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/variation-options/:id", server.deleteVariationOption) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/product-items", server.createProductItem)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/product-items/:id", server.getProductItem)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/product-items", server.listProductItems)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/product-items/:id", server.updateProductItem)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/product-items/:id", server.deleteProductItem) //! Admin Only # Finished With tests (token and changed response... No Etag)

	adminRoutes.POST("/product-configurations", server.createProductConfiguration)       //! Admin Only # Finished With tests (token and changed response... No Etag)
	router.GET("/product-configurations/:id", server.getProductConfiguration)            //? no auth required # Finished With tests (token and changed response... No Etag)
	router.GET("/product-configurations", server.listProductConfigurations)              //? no auth required # Finished With tests (token and changed response.)
	adminRoutes.PUT("/product-configurations/:id", server.updateProductConfiguration)    //! Admin Only # Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/product-configurations/:id", server.deleteProductConfiguration) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRoutes.GET("/users/shop-order/:id", server.getShopOrderItem) //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/shop-order", server.listShopOrderItems)   //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/order-status", server.createOrderStatus)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/order-status/:id", server.getOrderStatus)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/order-status", server.listOrderStatuses)         //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/order-status/:id", server.updateOrderStatus)     //* Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/users/order-status/:id", server.deleteOrderStatus) //! Admin Only # Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/shipping-method", server.createShippingMethod)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/shipping-method/:id", server.getShippingMethod)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/shipping-method", server.listShippingMethodes)         //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/shipping-method/:id", server.updateShippingMethod)     //* Finished With tests (token and changed response... No Etag)
	adminRoutes.DELETE("/users/shipping-method/:id", server.deleteShippingMethod) //! Admin Only # Finished With tests (token and changed response... No Etag)

	server.router = router

}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) gracefullShutDown(router *gin.Engine) {
	srv := &http.Server{
		Addr:    server.config.ServerAddress,
		Handler: router,
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
