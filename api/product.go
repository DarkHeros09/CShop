package api

import (
	"errors"
	"fmt"
	"math"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createProductParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductJsonRequest struct {
	Name        string `json:"name" validate:"required,alphanum"`
	CategoryID  int64  `json:"category_id" validate:"required,min=1"`
	BrandID     int64  `json:"brand_id" validate:"required,min=1"`
	Description string `json:"description" validate:"required"`
	// ProductImage string `json:"product_image" validate:"required,url"`
	Active bool `json:"active" validate:"boolean"`
}

func (server *Server) createProduct(ctx *fiber.Ctx) error {
	params := &createProductParamsRequest{}
	req := &createProductJsonRequest{}

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

	arg := db.AdminCreateProductParams{
		AdminID:     authPayload.AdminID,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		// ProductImage: req.ProductImage,
		Active: req.Active,
	}

	product, err := server.store.AdminCreateProduct(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(product)
	return nil
}

//////////////* Get API //////////////

type getProductRequest struct {
	ProductID int64 `params:"productId" validate:"required,min=1"`
}

func (server *Server) getProduct(ctx *fiber.Ctx) error {
	params := &getProductRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	product, err := server.store.GetProduct(ctx.Context(), params.ProductID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(product)
	return nil
}

// ////////////* List API //////////////

// type listProductsResponse struct {
// 	MaxPage  int64                `json:"max_page"`
// 	Products []db.ListProductsRow `json:"products"`
// }

// func newListProductsResponse(productsList []db.ListProductsRow, query *listProductsQueryRequest) listProductsResponse {

// 	maxPage := int64(math.Ceil(float64(productsList[0].TotalCount) / float64(query.PageSize)))
// 	return listProductsResponse{
// 		MaxPage:  maxPage,
// 		Products: productsList,
// 	}
// }

type listProductsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listProducts(ctx *fiber.Ctx) error {
	query := &listProductsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	products, err := server.store.ListProducts(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	// rsp := newListProductsResponse(products, query)
	maxPage := int64(math.Ceil(float64(products[0].TotalCount) / float64(query.PageSize)))

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(products)
	return nil
}

//////////////* Update API //////////////

type updateProductParamsRequest struct {
	AdminID   int64 `params:"adminId" validate:"required,min=1"`
	ProductID int64 `params:"productId" validate:"required,min=1"`
}

type updateProductJsonRequest struct {
	Name         string `json:"name" validate:"omitempty,required,alphanum"`
	CategoryID   int64  `json:"category_id" validate:"omitempty,required,min=1"`
	Description  string `json:"description" validate:"omitempty,required"`
	ProductImage string `json:"product_image" validate:"omitempty,required,http_url"`
	Active       bool   `json:"active" validate:"omitempty,required,boolean"`
}

func (server *Server) updateProduct(ctx *fiber.Ctx) error {
	params := &updateProductParamsRequest{}
	req := &updateProductJsonRequest{}

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

	arg := db.UpdateProductParams{
		ID:          params.ProductID,
		Name:        null.StringFromPtr(&req.Name),
		CategoryID:  null.IntFromPtr(&req.CategoryID),
		Description: null.StringFromPtr(&req.Description),
		// ProductImage: null.StringFromPtr(&req.ProductImage),
		Active: null.BoolFromPtr(&req.Active),
	}

	product, err := server.store.UpdateProduct(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(product)
	return nil
}

//////////////* Delete API //////////////

type deleteProductParamsRequest struct {
	AdminID   int64 `params:"adminId" validate:"required,min=1"`
	ProductID int64 `params:"productId" validate:"required,min=1"`
}

func (server *Server) deleteProduct(ctx *fiber.Ctx) error {
	params := &deleteProductParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err := server.store.DeleteProduct(ctx.Context(), params.ProductID)
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
