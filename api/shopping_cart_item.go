package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// //////////////* Create API //////////////

// type createShoppingCartItemParamsRequest struct {
// 	UserID         int64 `params:"id" validate:"required,min=1"`
// 	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
// }

// type data struct {
// 	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
// 	QTY           int32 `json:"qty" validate:"required,min=1"`
// }

// type createShoppingCartItemsRequest struct {
// 	ShopCartItem []data `json:"data" validate:"required,dive,required"`
// }

// func (server *Server) createShoppingCartItem(ctx *fiber.Ctx) error {
// 	params := &createShoppingCartItemParamsRequest{}
// 	req := &createShoppingCartItemsRequest{}
// 	var arg []db.CreateShoppingCartItemParams
// 	var shoppingCartItems []db.ShoppingCartItem
// 	var err1 error

// 	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
// 		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
// 		return nil
// 	}

// 	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
// 	if authPayload.UserID != params.UserID {
// 		err := errors.New("account deosn't belong to the authenticated user")
// 		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
// 		return nil
// 	}

// 	for i := 0; i < len(req.ShopCartItem); i++ {
// 		arg = append(arg, db.CreateShoppingCartItemParams{
// 			ShoppingCartID: params.ShoppingCartID,
// 			ProductItemID:  req.ShopCartItem[i].ProductItemID,
// 			Qty:            req.ShopCartItem[i].QTY,
// 		})
// 	}

// 	result := server.store.CreateShoppingCartItem(ctx.Context(), arg)

// 	result.Query(func(i int, sci []db.ShoppingCartItem, err error) {
// 		err1 = err
// 		shoppingCartItems = append(shoppingCartItems, sci...)
// 	})

// 	if err1 != nil {
// 		if pqErr, ok := err1.(*pgconn.PgError); ok {
// 			switch pqErr.Message {
// 			case util.ForeignKeyViolation, util.UniqueViolation:
// 				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err1))
// 				return nil
// 			}
// 		}
// 		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err1))
// 		return nil
// 	}
// 	ctx.Status(fiber.StatusOK).JSON(shoppingCartItems)
// 	return nil
// }

//////////////* Create API //////////////

type createShoppingCartItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}
type createShoppingCartItemRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
	SizeID        int64 `json:"size_id" validate:"required,min=1"`
	QTY           int32 `json:"qty" validate:"required,min=1"`
}

func (server *Server) createShoppingCartItem(ctx *fiber.Ctx) error {
	params := &createShoppingCartItemParamsRequest{}
	req := &createShoppingCartItemRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateShoppingCartItemParams{
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  req.ProductItemID,
		Qty:            req.QTY,
		SizeID:         req.SizeID,
	}

	shoppingCartItem, err := server.store.CreateShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCartItem)
	return nil
}

// ////////////* Get API //////////////
type getShoppingCartItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}

func (server *Server) getShoppingCartItem(ctx *fiber.Ctx) error {
	params := &getShoppingCartItemParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetShoppingCartItemByUserIDCartIDParams{
		UserID: params.UserID,
		ID:     params.ShoppingCartID,
	}

	shoppingCartItem, err := server.store.GetShoppingCartItemByUserIDCartID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if shoppingCartItem == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCartItem)
	return nil
}

//////////////* List API //////////////

type listShoppingCartItemsParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShoppingCartItemsResponse struct {
	// ShopCartItems []db.ListShoppingCartItemsByUserIDRow `json:"shop_cart_items"`
	ID             null.Int  `json:"id"`
	ShoppingCartID null.Int  `json:"shopping_cart_id"`
	CreatedAt      null.Time `json:"created_at"`
	UpdatedAt      null.Time `json:"updated_at"`
	// ProductItems  []db.ListProductItemsByIDsRow         `json:"product_items"`
	ProductItemID             null.Int    `json:"product_item_id"`
	Name                      null.String `json:"name"`
	SizeID                    null.Int    `json:"size_id"`
	SizeValue                 null.String `json:"size_value"`
	SizeQty                   null.Int32  `json:"size_qty"`
	Color                     null.String `json:"color"`
	Qty                       null.Int    `json:"qty"`
	ProductID                 int64       `json:"product_id"`
	ProductImage              string      `json:"product_image"`
	Price                     string      `json:"price"`
	CategoryPromoID           null.Int    `json:"category_promo_id"`
	CategoryPromoName         null.String `json:"category_promo_name"`
	CategoryPromoDescription  null.String `json:"category_promo_description"`
	CategoryPromoDiscountRate null.Int    `json:"category_promo_discount_rate"`
	CategoryPromoStartDate    null.Time   `json:"category_promo_start_date"`
	CategoryPromoEndDate      null.Time   `json:"category_promo_end_date"`
	BrandPromoID              null.Int    `json:"brand_promo_id"`
	BrandPromoName            null.String `json:"brand_promo_name"`
	BrandPromoDescription     null.String `json:"brand_promo_description"`
	BrandPromoDiscountRate    null.Int    `json:"brand_promo_discount_rate"`
	BrandPromoStartDate       null.Time   `json:"brand_promo_start_date"`
	BrandPromoEndDate         null.Time   `json:"brand_promo_end_date"`
	ProductPromoID            null.Int    `json:"product_promo_id"`
	ProductPromoName          null.String `json:"product_promo_name"`
	ProductPromoDescription   null.String `json:"product_promo_description"`
	ProductPromoDiscountRate  null.Int    `json:"product_promo_discount_rate"`
	ProductPromoStartDate     null.Time   `json:"product_promo_start_date"`
	ProductPromoEndDate       null.Time   `json:"product_promo_end_date"`
	Active                    bool        `json:"active"`
	ProductPromoActive        bool        `json:"product_promo_active"`
	CategoryPromoActive       bool        `json:"category_promo_active"`
	BrandPromoActive          bool        `json:"brand_promo_active"`
}

func newlistShoppingCartItemsResponse(
	shopCartItems []*db.ListShoppingCartItemsByUserIDRow,
	productItems []*db.ListProductItemsByIDsRow,
	productSizes []*db.ProductSize,
	rsp []*listShoppingCartItemsResponse,
) []*listShoppingCartItemsResponse {

	// Build lookup maps O(n)
	cartMap := make(map[int64]*db.ListShoppingCartItemsByUserIDRow, len(shopCartItems))
	for i := 0; i < len(shopCartItems); i++ {
		sc := shopCartItems[i]
		cartMap[sc.ProductItemID.Int64] = sc
	}

	productMap := make(map[int64]*db.ListProductItemsByIDsRow, len(productItems))
	for i := 0; i < len(productItems); i++ {
		p := productItems[i]
		productMap[p.ID] = p
	}

	sizeMap := make(map[int64]*db.ProductSize, len(productSizes))
	for i := 0; i < len(productSizes); i++ {
		s := productSizes[i]
		sizeMap[s.ProductItemID] = s
	}

	// Fill with classic for-loop
	for i := 0; i < len(productItems); i++ {
		p := productItems[i]
		pid := p.ID

		sc := cartMap[pid]
		if sc == nil {
			continue
		}

		s := sizeMap[pid]
		if s == nil {
			continue
		}

		rsp[i] = &listShoppingCartItemsResponse{
			ID:             sc.ID,
			ShoppingCartID: sc.ShoppingCartID,
			CreatedAt:      sc.CreatedAt,
			UpdatedAt:      sc.UpdatedAt,
			ProductItemID:  sc.ProductItemID,
			Name:           p.Name,
			Qty:            sc.Qty,
			ProductID:      p.ProductID,
			ProductImage:   p.ProductImage1.String,

			SizeID:    null.IntFromPtr(&s.ID),
			SizeValue: null.StringFromPtr(&s.SizeValue),
			SizeQty:   null.Int32FromPtr(&s.Qty),

			Color:  p.ColorValue,
			Price:  p.Price,
			Active: p.Active,

			CategoryPromoID:           p.CategoryPromoID,
			CategoryPromoName:         p.CategoryPromoName,
			CategoryPromoDescription:  p.CategoryPromoDescription,
			CategoryPromoDiscountRate: p.CategoryPromoDiscountRate,
			CategoryPromoActive:       p.CategoryPromoActive,
			CategoryPromoStartDate:    p.CategoryPromoStartDate,
			CategoryPromoEndDate:      p.CategoryPromoEndDate,

			BrandPromoID:           p.BrandPromoID,
			BrandPromoName:         p.BrandPromoName,
			BrandPromoDescription:  p.BrandPromoDescription,
			BrandPromoDiscountRate: p.BrandPromoDiscountRate,
			BrandPromoActive:       p.BrandPromoActive,
			BrandPromoStartDate:    p.BrandPromoStartDate,
			BrandPromoEndDate:      p.BrandPromoEndDate,

			ProductPromoID:           p.ProductPromoID,
			ProductPromoName:         p.ProductPromoName,
			ProductPromoDescription:  p.ProductPromoDescription,
			ProductPromoDiscountRate: p.ProductPromoDiscountRate,
			ProductPromoActive:       p.ProductPromoActive,
			ProductPromoStartDate:    p.ProductPromoStartDate,
			ProductPromoEndDate:      p.ProductPromoEndDate,
		}
	}

	return rsp
}

func (server *Server) listShoppingCartItems(ctx *fiber.Ctx) error {
	params := &listShoppingCartItemsParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}
	shoppingCartItems, err := server.store.ListShoppingCartItemsByUserID(ctx.Context(), authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if shoppingCartItems == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	productsItemsIds := make([]int64, len(shoppingCartItems))
	for i := 0; i < len(shoppingCartItems); i++ {
		productsItemsIds[i] = shoppingCartItems[i].ProductItemID.Int64
	}

	productItems, err := server.store.ListProductItemsByIDs(ctx.Context(), productsItemsIds)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productItems == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	productsSizesIds := make([]int64, len(shoppingCartItems))
	for i := 0; i < len(shoppingCartItems); i++ {
		productsSizesIds[i] = shoppingCartItems[i].SizeID.Int64
	}

	productsSizes, err := server.store.ListProductSizesByIDs(ctx.Context(), productsSizesIds)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productsSizes == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	// listShoppingCartItemsChan := make(chan []listShoppingCartItemsResponse, len(productItems))

	// go func() {
	// 	rsp := newlistShoppingCartItemsResponse(shoppingCartItems, productItems)
	// 	listShoppingCartItemsChan <- rsp
	// 	}()

	// Pre-allocate final slice - NO append
	rsp := make([]*listShoppingCartItemsResponse, len(productItems))

	// 	rsp := <-listShoppingCartItemsChan
	rsp = newlistShoppingCartItemsResponse(shoppingCartItems, productItems, productsSizes, rsp)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// ////////////* UPDATE API //////////////
type updateShoppingCartItemParamsRequest struct {
	UserID             int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID     int64 `params:"cartId" validate:"required,min=1"`
	ShoppingCartItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateShoppingCartItemJsonRequest struct {
	ProductItemID *int64 `json:"product_item_id" validate:"omitempty,required,min=1"`
	SizeID        *int64 `json:"size_id" validate:"omitempty,required,min=1"`
	QTY           *int64 `json:"qty" validate:"omitempty,required"`
}

func (server *Server) updateShoppingCartItem(ctx *fiber.Ctx) error {
	params := &updateShoppingCartItemParamsRequest{}
	req := &updateShoppingCartItemJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateShoppingCartItemParams{
		ID:             params.ShoppingCartItemID,
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  null.IntFromPtr(req.ProductItemID),
		SizeID:         null.IntFromPtr(req.SizeID),
		Qty:            null.IntFromPtr(req.QTY),
	}

	shoppingCart, err := server.store.UpdateShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCart)
	return nil
}

// ////////////* Delete API //////////////
type deleteShoppingCartItemParamsRequest struct {
	UserID             int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID     int64 `params:"cartId" validate:"required,min=1"`
	ShoppingCartItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) deleteShoppingCartItem(ctx *fiber.Ctx) error {
	params := &deleteShoppingCartItemParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteShoppingCartItemParams{
		UserID:             authPayload.UserID,
		ShoppingCartID:     params.ShoppingCartID,
		ShoppingCartItemID: params.ShoppingCartItemID,
	}

	err := server.store.DeleteShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

// ////////////* Delete All API //////////////

type deleteShoppingCartItemAllParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}

func (server *Server) deleteShoppingCartItemAllByUser(ctx *fiber.Ctx) error {
	params := &deleteShoppingCartItemAllParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteShoppingCartItemAllByUserParams{
		UserID:         params.UserID,
		ShoppingCartID: params.ShoppingCartID,
	}

	_, err := server.store.DeleteShoppingCartItemAllByUser(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

// ////////////* Finish Purshase API //////////////
type finishPurshaseParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}
type finishPurshaseJsonRequest struct {
	UserAddressID    int64  `json:"user_address_id" validate:"required,min=1"`
	PaymentTypeID    int64  `json:"payment_type_id" validate:"required,min=1"`
	ShippingMethodID int64  `json:"shipping_method_id" validate:"required,min=1"`
	OrderStatusID    int64  `json:"order_status_id" validate:"required,min=1"`
	OrderTotal       string `json:"order_total" validate:"required"`
}

func (server *Server) finishPurchase(ctx *fiber.Ctx) error {
	params := &finishPurshaseParamsRequest{}
	req := &finishPurshaseJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.FinishedPurchaseTxParams{
		UserID:           authPayload.UserID,
		AddressID:        req.UserAddressID,
		PaymentTypeID:    req.PaymentTypeID,
		ShoppingCartID:   params.ShoppingCartID,
		ShippingMethodID: req.ShippingMethodID,
		OrderStatusID:    req.OrderStatusID,
		OrderTotal:       req.OrderTotal,
	}

	finishedPurchase, err := server.store.FinishedPurchaseTx(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(finishedPurchase)
	return nil
}
