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

	userRoutes.POST("/users/addresses", server.createUserAddress)            //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/addresses/:id", server.getUserAddress)            //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/addresses", server.listUserAddresses)             //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/addresses/:user-id", server.updateUserAddress)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/addresses/:user-id", server.deleteUserAddress) //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/cart", server.createShoppingCartItem)                          //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/cart/:shopping-cart-id", server.getShoppingCartItem)            //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/cart", server.listShoppingCartItems)                            //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/cart/:shopping-cart-id", server.updateShoppingCartItem)         //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/cart/:shopping-cart-item-id", server.deleteShoppingCartItem) //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/cart/delete-all", server.deleteShoppingCartItemAllByUser)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/cart/purchase", server.finishPurchase)                          //* Finished With tests (token and changed response... No Etag)

	userRoutes.POST("/users/payment-method", server.createPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/payment-method/:id", server.getPaymentMethod)       //* Finished With tests (token and changed response... No Etag)
	userRoutes.GET("/users/payment-method", server.listPaymentMethodes)        //* Finished With tests (token and changed response... No Etag)
	userRoutes.PUT("/users/payment-method/:id", server.updatePaymentMethod)    //* Finished With tests (token and changed response... No Etag)
	userRoutes.DELETE("/users/payment-method/:id", server.deletePaymentMethod) //* Finished With tests (token and changed response... No Etag)

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
