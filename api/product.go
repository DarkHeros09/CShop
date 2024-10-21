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
	Name        string `json:"name" validate:"required,alphanumunicode"`
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
	Name        *string `json:"name" validate:"omitempty,required,alphanumunicode"`
	CategoryID  *int64  `json:"category_id" validate:"omitempty,required,min=1"`
	BrandID     *int64  `json:"brand_id" validate:"omitempty,required,min=1"`
	Description *string `json:"description" validate:"omitempty,required"`
	Active      *bool   `json:"active" validate:"omitempty,required,boolean"`
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

	arg := db.AdminUpdateProductParams{
		AdminID:     authPayload.AdminID,
		ID:          params.ProductID,
		Name:        null.StringFromPtr(req.Name),
		CategoryID:  null.IntFromPtr(req.CategoryID),
		BrandID:     null.IntFromPtr(req.BrandID),
		Description: null.StringFromPtr(req.Description),
		Active:      null.BoolFromPtr(req.Active),
	}

	product, err := server.store.AdminUpdateProduct(ctx.Context(), arg)
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

	arg := db.AdminDeleteProductParams{
		AdminID: authPayload.AdminID,
		ID:      params.ProductID,
	}

	err := server.store.AdminDeleteProduct(ctx.Context(), arg)
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

type listProductsV2QueryRequest struct {
	Limit int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductsV2(ctx *fiber.Ctx) error {
	query := &listProductsV2QueryRequest{}
	// var maxPage int64

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	products, err := server.store.ListProductsV2(ctx.Context(), query.Limit)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(products) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductsV2Row{})
		return nil
	}
	// if len(products) != 0 {
	// 	maxPage = int64(math.Ceil(float64(products[0].TotalCount) / float64(query.Limit)))
	// 	// ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	// ctx.Status(fiber.StatusOK).JSON(products)
	// } else {
	// 	maxPage = 0
	// 	// ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	// ctx.Status(fiber.StatusOK).JSON([]db.ListProductsV2Row{})
	// }

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(products[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(products)
	return nil

}

type listProductsNextPageQueryRequest struct {
	ProductCursor int64 `query:"product_cursor" validate:"required,min=1"`
	Limit         int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductsNextPage(ctx *fiber.Ctx) error {
	query := &listProductsNextPageQueryRequest{}
	// var maxPage int64

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductsNextPageParams{
		Limit: query.Limit,
		ID:    query.ProductCursor,
	}

	products, err := server.store.ListProductsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(products) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductsNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(products[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(products)
	return nil

}

//////////////* Paginated Search API //////////////

type searchProductsQueryRequest struct {
	Limit int32  `query:"limit" validate:"required,min=5,max=10"`
	Query string `query:"query" validate:"omitempty,required,alphanumunicode"`
}

func (server *Server) searchProducts(ctx *fiber.Ctx) error {
	query := &searchProductsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.SearchProductsParams{
		Limit: query.Limit,
		Query: query.Query,
	}

	products, err := server.store.SearchProducts(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(products) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.SearchProductItemsRow{})
		return nil
	}
	// //Todo: should be copied to other functions??
	// if len(productItems) > 0 {
	// 	pagesNumber := float64(productItems[0].TotalCount) / float64(query.Limit)
	// 	if len(productItems) == int(productItems[0].TotalCount) {
	// 		ctx.Set("Max-Page", "0")
	// 	} else {
	// 		maxPage := int64(math.Ceil(pagesNumber))
	// 		ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	}
	// } else {
	// 	ctx.Set("Max-Page", "0")
	// }
	//! fix the next line
	ctx.Set("Next-Available", fmt.Sprint(products[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(products)
	return nil

}

type searchProductsNextPageQueryRequest struct {
	ProductCursor int64  `query:"product_cursor" validate:"required,min=1"`
	Limit         int32  `query:"limit" validate:"required,min=5,max=10"`
	Query         string `query:"query" validate:"omitempty,required,alphanumunicode"`
}

func (server *Server) searchProductsNextPage(ctx *fiber.Ctx) error {
	query := &searchProductsNextPageQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.SearchProductsNextPageParams{
		Limit:     query.Limit,
		ProductID: query.ProductCursor,
		Query:     query.Query,
	}

	products, err := server.store.SearchProductsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(products) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductsNextPageRow{})
		return nil
	}
	// if len(products) > 0 {
	// 	pagesNumber := float64(products[0].TotalCount) / float64(query.Limit)
	// 	if len(products) == int(products[0].TotalCount) {
	// 		ctx.Set("Max-Page", "0")
	// 	} else {
	// 		maxPage := int64(math.Ceil(pagesNumber))
	// 		ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	}
	// } else {
	// 	ctx.Set("Max-Page", f"0"))
	// }

	ctx.Set("Next-Available", fmt.Sprint(products[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(products)
	return nil

}
