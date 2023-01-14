package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// ////////////* Create API //////////////
type createProductItemsParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductItemsJsonRequest struct {
	ProductID    int64  `json:"product_id" validate:"required,min=1"`
	ProductSKU   int64  `json:"product_sku" validate:"required"`
	QtyInStock   int32  `json:"qty_in_stock" validate:"required"`
	ProductImage string `json:"product_image" validate:"required,url"`
	Price        string `json:"price" validate:"required"`
	Active       bool   `json:"active" validate:"boolean"`
}

func (server *Server) createProductItem(ctx *fiber.Ctx) error {
	var params createProductItemsParamsRequest
	var req createProductItemsJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateProductItemParams{
		ProductID:    req.ProductID,
		ProductSku:   req.ProductSKU,
		QtyInStock:   req.QtyInStock,
		ProductImage: req.ProductImage,
		Price:        req.Price,
		Active:       req.Active,
	}

	productItem, err := server.store.CreateProductItem(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productItem)
	return nil
}

//////////////* Get API //////////////

type getProductItemsParamsRequest struct {
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) getProductItem(ctx *fiber.Ctx) error {
	var params getProductItemsParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productItem, err := server.store.GetProductItem(ctx.Context(), params.ProductItemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(productItem)
	return nil
}

//////////////* List API //////////////

type listProductItemsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItems(ctx *fiber.Ctx) error {
	var query listProductItemsQueryRequest

	if err := ctx.QueryParser(&query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	productItems, err := server.store.ListProductItems(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Update API //////////////

type updateProductItemParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateProductItemJsonRequest struct {
	ProductID    int64  `json:"product_id" validate:"required,min=1"`
	ProductSKU   int64  `json:"product_sku" validate:"omitempty,required"`
	QtyInStock   int64  `json:"qty_in_stock" validate:"omitempty,required"`
	ProductImage string `json:"product_image" validate:"omitempty,required,url"`
	Price        string `json:"price" validate:"omitempty,required"`
	Active       bool   `json:"active" validate:"boolean"`
}

func (server *Server) updateProductItem(ctx *fiber.Ctx) error {
	var params updateProductItemParamsRequest
	var req updateProductItemJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateProductItemParams{
		ID:           params.ProductItemID,
		ProductID:    req.ProductID,
		ProductSku:   null.IntFromPtr(&req.ProductSKU),
		QtyInStock:   null.IntFromPtr(&req.QtyInStock),
		ProductImage: null.StringFromPtr(&req.ProductImage),
		Price:        null.StringFromPtr(&req.Price),
		Active:       null.BoolFromPtr(&req.Active),
	}

	productItem, err := server.store.UpdateProductItem(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productItem)
	return nil
}

//////////////* Delete API //////////////

type deleteProductItemParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) deleteProductItem(ctx *fiber.Ctx) error {
	var params deleteProductItemParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err := server.store.DeleteProductItem(ctx.Context(), params.ProductItemID)
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
