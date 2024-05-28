package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createWishListItemParamsRequest struct {
	UserID     int64 `params:"id" validate:"required,min=1"`
	WishListID int64 `params:"wishId" validate:"required,min=1"`
}

type createWishListItemJsonRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
}

func (server *Server) createWishListItem(ctx *fiber.Ctx) error {
	params := &createWishListItemParamsRequest{}
	req := &createWishListItemJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateWishListItemParams{
		WishListID:    params.WishListID,
		ProductItemID: req.ProductItemID,
	}

	wishListItem, err := server.store.CreateWishListItem(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(wishListItem)
	return nil
}

//////////////* Get API //////////////

type getWishListItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	WishListID     int64 `params:"wishId" validate:"required,min=1"`
	WishListItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) getWishListItem(ctx *fiber.Ctx) error {
	params := &getWishListItemParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetWishListItemByUserIDCartIDParams{
		UserID:     authPayload.UserID,
		ID:         params.WishListItemID,
		WishListID: params.WishListID,
	}

	wishListItem, err := server.store.GetWishListItemByUserIDCartID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(wishListItem)
	return nil
}

//////////////* List API //////////////

type listWishListItemsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
	// WishListID int64 `params:"wishId" validate:"required,min=1"`
}

type listWishListItemsResponse struct {
	ID            null.Int    `json:"id"`
	WishListID    null.Int    `json:"wish_list_id"`
	CreatedAt     null.Time   `json:"created_at"`
	UpdatedAt     null.Time   `json:"updated_at"`
	ProductItemID null.Int    `json:"product_item_id"`
	Name          null.String `json:"name"`
	Size          null.String `json:"size"`
	Color         null.String `json:"color"`
	ProductID     int64       `json:"product_id"`
	ProductImage  string      `json:"product_image"`
	Price         string      `json:"price"`
	Active        bool        `json:"active"`
}

func newlistWishListItemsResponse(wishListItems []db.ListWishListItemsByUserIDRow, productItems []db.ListProductItemsByIDsRow) []listWishListItemsResponse {
	rsp := make([]listWishListItemsResponse, len(productItems))
	for i := 0; i < len(productItems); i++ {
		for j := 0; j < len(wishListItems); j++ {
			if productItems[i].ID == wishListItems[j].ProductItemID.Int64 {
				rsp[i] = listWishListItemsResponse{
					ID:            wishListItems[j].ID,
					WishListID:    wishListItems[j].WishListID,
					CreatedAt:     wishListItems[j].CreatedAt,
					UpdatedAt:     wishListItems[j].UpdatedAt,
					ProductItemID: wishListItems[j].ProductItemID,
					Name:          productItems[i].Name,
					ProductID:     productItems[i].ProductID,
					ProductImage:  productItems[i].ProductImage1.String,
					Size:          productItems[i].SizeValue,
					Color:         productItems[i].ColorValue,
					Price:         productItems[i].Price,
					Active:        productItems[i].Active,
				}
			}
		}
	}

	return rsp
}

func (server *Server) listWishListItems(ctx *fiber.Ctx) error {
	params := &listWishListItemsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	wishListItems, err := server.store.ListWishListItemsByUserID(ctx.Context(), authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	productsItemsIds := make([]int64, len(wishListItems))
	for i := 0; i < len(wishListItems); i++ {
		productsItemsIds[i] = wishListItems[i].ProductItemID.Int64
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

	rsp := newlistWishListItemsResponse(wishListItems, productItems)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// ////////////* UPDATE API //////////////
type updateWishListItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	WishListID     int64 `params:"wishId" validate:"required,min=1"`
	WishListItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateWishListItemJsonRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"omitempty,required"`
}

func (server *Server) updateWishListItem(ctx *fiber.Ctx) error {
	params := &updateWishListItemParamsRequest{}
	req := &updateWishListItemJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateWishListItemParams{
		ID:            params.WishListItemID,
		WishListID:    params.WishListID,
		ProductItemID: null.IntFromPtr(&req.ProductItemID),
	}

	wishList, err := server.store.UpdateWishListItem(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(wishList)
	return nil
}

// ////////////* Delete API //////////////
type deleteWishListItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	WishListID     int64 `params:"wishId" validate:"required,min=1"`
	WishListItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) deleteWishListItem(ctx *fiber.Ctx) error {
	params := &deleteWishListItemParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteWishListItemParams{
		ID:         params.WishListItemID,
		WishListID: params.WishListID,
	}

	err := server.store.DeleteWishListItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
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

type deleteWishListItemAllJsonRequest struct {
	UserID     int64 `params:"id" validate:"required,min=1"`
	WishListID int64 `params:"wishId" validate:"required,min=1"`
}

func (server *Server) deleteWishListItemAll(ctx *fiber.Ctx) error {
	params := &deleteWishListItemAllJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	_, err := server.store.DeleteWishListItemAll(ctx.Context(), params.WishListID)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
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
