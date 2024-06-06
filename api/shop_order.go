package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"

	"firebase.google.com/go/messaging"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// //////////////* UPDATE API ///////////////
type updateShopOrderParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	ShopOrderID int64 `params:"shopOrderId" validate:"required,min=1"`
}

type updateShopOrderJsonRequest struct {
	UserID            int64  `json:"user_id" validate:"omitempty,required,min=1"`
	ShippingAddressID int64  `json:"shipping_address_id" validate:"omitempty,required,min=1"`
	ShippingMethodID  int64  `json:"shipping_method_id" validate:"omitempty,required,min=1"`
	OrderStatusID     int64  `json:"order_status_id" validate:"omitempty,required,min=1"`
	TrackNumber       string `json:"track_number" validate:"omitempty,required"`
	OrderTotal        string `json:"order_total" validate:"omitempty,required"`
	DeviceID          string `json:"device_id" validate:"omitempty,required"`
}

func (server *Server) updateShopOrder(ctx *fiber.Ctx) error {
	params := &updateShopOrderParamsRequest{}
	req := &updateShopOrderJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateShopOrderParams{
		ID:                params.ShopOrderID,
		TrackNumber:       null.StringFromPtr(&req.TrackNumber),
		UserID:            null.IntFromPtr(&req.UserID),
		ShippingAddressID: null.IntFromPtr(&req.ShippingAddressID),
		OrderTotal:        null.StringFromPtr(&req.OrderTotal),
		ShippingMethodID:  null.IntFromPtr(&req.ShippingMethodID),
		OrderStatusID:     null.IntFromPtr(&req.OrderStatusID),
	}

	shopOrder, err := server.store.UpdateShopOrder(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shopOrder)

	//? new code fcm
	contextBG := context.Background()

	notiArg := db.GetNotificationParams{
		UserID:   shopOrder.UserID,
		DeviceID: null.StringFrom(req.DeviceID),
	}

	notification, err := server.store.GetNotification(contextBG, notiArg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	fcmClient, err := server.fb.Messaging(contextBG)
	if err != nil {
		log.Printf("error getting Messaging client: %v\n", err)
		return nil
	}

	// This registration token comes from the client FCM SDKs.
	// registrationToken := "YOUR_REGISTRATION_TOKEN"
	registrationToken := notification.FcmToken.String
	// fmt.Println("Registration token is:", registrationToken)
	// See documentation on defining a message payload.
	message := &messaging.Message{
		// Data: map[string]string{
		// 	"score": "850",
		// 	"time":  "2:45",
		// },
		Notification: &messaging.Notification{
			Title: "$جاهز",
			Body:  "الشحنة جاهزة للتسليم 🎉🎉🎉",
		},
		// Android: &messaging.AndroidConfig{
		// 	Data: map[string]string{
		// 		"score": "850",
		// 		"time":  "2:45",
		// 	},
		// 	// TTL: &oneHour,
		// 	Notification: &messaging.AndroidNotification{
		// 		Icon:  "stock_ticker_update",
		// 		Color: "#f45342",
		// 	}},
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	isVerified, err := fcmClient.Send(contextBG, message)
	if err != nil {
		log.Println(err)
	} else {
		// Response is a message ID string.
		fmt.Println("Successfully verified message:", isVerified)
		response, err := fcmClient.Send(contextBG, message)
		if err != nil {
			log.Println(err)
		}
		// Response is a message ID string.
		fmt.Println("Successfully sent message:", response)
	}
	//? end fcm

	return nil
}

//////////////* List API //////////////

type listShopOrdersParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listShopOrdersQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listShopOrders(ctx *fiber.Ctx) error {
	params := &listShopOrdersParamsRequest{}
	query := &listShopOrdersQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	shopOrders, err := server.store.ListShopOrdersByUserID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}

//////////////* Pagination List API //////////////

type listShopOrdersV2ParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShopOrdersV2QueryRequest struct {
	Limit       int32       `query:"limit" validate:"required,min=5,max=10"`
	OrderStatus null.String `query:"order_status" validate:"omitempty,required"`
}

func (server *Server) listShopOrdersV2(ctx *fiber.Ctx) error {
	params := &listShopOrdersV2ParamsRequest{}
	query := &listShopOrdersV2QueryRequest{}
	var maxPage int64

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDV2Params{
		UserID:      authPayload.UserID,
		Limit:       query.Limit,
		OrderStatus: query.OrderStatus,
	}
	shopOrders, err := server.store.ListShopOrdersByUserIDV2(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if len(shopOrders) != 0 {
		maxPage = int64(math.Ceil(float64(shopOrders[0].TotalCount) / float64(query.Limit)))
	} else {
		maxPage = 0
	}

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}

type listShopOrdersNextPageParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShopOrdersNextPageQueryRequest struct {
	Cursor      int64       `query:"cursor" validate:"required,min=1"`
	Limit       int32       `query:"limit" validate:"required,min=5,max=10"`
	OrderStatus null.String `query:"order_status" validate:"omitempty,required"`
}

func (server *Server) listShopOrdersNextPage(ctx *fiber.Ctx) error {
	params := &listShopOrdersNextPageParamsRequest{}
	query := &listShopOrdersNextPageQueryRequest{}
	var maxPage int64

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDNextPageParams{
		UserID:      authPayload.UserID,
		ShopOrderID: query.Cursor,
		Limit:       query.Limit,
		OrderStatus: query.OrderStatus,
	}
	shopOrders, err := server.store.ListShopOrdersByUserIDNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if len(shopOrders) != 0 {
		maxPage = int64(math.Ceil(float64(shopOrders[0].TotalCount) / float64(query.Limit)))
	} else {
		maxPage = 0
	}

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}
