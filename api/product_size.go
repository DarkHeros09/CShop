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

// ////////////* Create API //////////////
type createProductSizeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductSizeJsonRequest struct {
	SizeValue     string `json:"size_value" validate:"required"`
	ProductItemId int64  `json:"product_item_id" validate:"required,min=1"`
	Qty           int64  `json:"qty" validate:"required,min=1"`
}

func (server *Server) createProductSize(ctx *fiber.Ctx) error {
	params := &createProductSizeParamsRequest{}
	req := &createProductSizeJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminCreateProductSizeParams{
		AdminID:       authPayload.AdminID,
		SizeValue:     req.SizeValue,
		ProductItemID: req.ProductItemId,
		Qty:           int32(req.Qty),
	}

	productSize, err := server.store.AdminCreateProductSize(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productSize)
	return nil
}

// ////////////* List API //////////////

type listProductSizesParamsRequest struct {
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) listProductSizes(ctx *fiber.Ctx) error {
	params := &getProductItemsParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productSizes, err := server.store.ListProductSizesByProductItemID(ctx.Context(), params.ProductItemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productSizes)
	return nil
}

//////////////* Update API //////////////

type updateProductSizeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateProductSizeJsonRequest struct {
	Size          *string `json:"size" validate:"omitempty,required,alphaunicode"`
	Qty           *int64  `json:"qty" validate:"omitempty,required,min=1"`
	ProductItemID int64   `json:"product_item_id" validate:"required,min=1"`
}

func (server *Server) updateProductSize(ctx *fiber.Ctx) error {
	params := &updateProductSizeParamsRequest{}
	req := &updateProductSizeJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminUpdateProductSizeParams{
		AdminID:       authPayload.AdminID,
		ID:            params.ID,
		SizeValue:     null.StringFromPtr(req.Size),
		Qty:           null.IntFromPtr(req.Qty),
		ProductItemID: req.ProductItemID,
	}

	size, err := server.store.AdminUpdateProductSize(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(size)
	return nil
}
