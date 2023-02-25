package api

import (
	"errors"
	"fmt"
	"math"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
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
	params := &createProductItemsParamsRequest{}
	req := &createProductItemsJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
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
	params := &getProductItemsParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
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
	query := &listProductItemsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
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

	maxPage := int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.PageSize)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
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
	params := &updateProductItemParamsRequest{}
	req := &updateProductItemJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
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
	params := &deleteProductItemParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
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

//////////////* Pagination List API //////////////

type listProductItemsV2QueryRequest struct {
	Limit int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsV2(ctx *fiber.Ctx) error {
	query := &listProductItemsV2QueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productItems, err := server.store.ListProductItemsV2(ctx.Context(), query.Limit)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	maxPage := int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.Limit)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type listProductItemsNextPageQueryRequest struct {
	Cursor int64 `query:"cursor" validate:"required,min=1"`
	Limit  int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsNextPage(ctx *fiber.Ctx) error {
	query := &listProductItemsNextPageQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsNextPageParams{
		Limit: query.Limit,
		ID:    query.Cursor,
	}

	productItems, err := server.store.ListProductItemsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	maxPage := int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.Limit)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Paginated Search API //////////////

type searchProductItemsQueryRequest struct {
	Limit int32  `query:"limit" validate:"required,min=5,max=10"`
	Query string `query:"query" validate:"required,alphanum"`
}

func (server *Server) searchProductItems(ctx *fiber.Ctx) error {
	query := &searchProductItemsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.SearchProductItemsParams{
		Limit: query.Limit,
		Query: query.Query,
	}

	productItems, err := server.store.SearchProductItems(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	maxPage := int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.Limit)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type searchProductItemsNextPageQueryRequest struct {
	Cursor int64  `query:"cursor" validate:"required,min=1"`
	Limit  int32  `query:"limit" validate:"required,min=5,max=10"`
	Query  string `query:"query" validate:"required,alphanum"`
}

func (server *Server) searchProductItemsNextPage(ctx *fiber.Ctx) error {
	query := &searchProductItemsNextPageQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.SearchProductItemsNextPageParams{
		Limit: query.Limit,
		ID:    query.Cursor,
		Query: query.Query,
	}

	productItems, err := server.store.SearchProductItemsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	maxPage := int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.Limit)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}
