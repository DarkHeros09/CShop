package api

import (
	"errors"
	"fmt"
	"math"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

const productTimeLayout = "2006-01-02T15:04:05.999999Z"

// const productTimeLayout = "2006-01-02 15:04:05.999999Z"

// ////////////* Create API //////////////
type createProductItemsParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductItemsJsonRequest struct {
	ProductID int64 `json:"product_id" validate:"required,min=1"`
	// SizeID     int64 `json:"size_id" validate:"required,min=1"`
	ImageID    int64 `json:"image_id" validate:"required,min=1"`
	ColorID    int64 `json:"color_id" validate:"required,min=1"`
	ProductSKU int64 `json:"product_sku" validate:"required"`
	// QtyInStock int32 `json:"qty_in_stock" validate:"required"`
	// ProductImage string `json:"product_image" validate:"required,url"`
	Price  string `json:"price" validate:"required"`
	Active bool   `json:"active" validate:"boolean"`
}

func (server *Server) createProductItem(ctx *fiber.Ctx) error {
	params := &createProductItemsParamsRequest{}
	req := &createProductItemsJsonRequest{}

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

	arg := db.AdminCreateProductItemParams{
		AdminID:    authPayload.AdminID,
		ProductID:  req.ProductID,
		ProductSku: req.ProductSKU,
		// QtyInStock: req.QtyInStock,
		// ProductImage: req.ProductImage,
		// SizeID:  req.SizeID,
		ImageID: req.ImageID,
		ColorID: req.ColorID,
		Price:   req.Price,
		Active:  req.Active,
	}

	productItem, err := server.store.AdminCreateProductItem(ctx.Context(), arg)
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

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
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

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
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
	ProductID  int64  `json:"product_id" validate:"required,min=1"`
	ProductSKU *int64 `json:"product_sku" validate:"omitempty,required"`
	QtyInStock *int64 `json:"qty_in_stock" validate:"omitempty,required"`
	// ProductImage string `json:"product_image" validate:"omitempty,required,url"`
	Price  *string `json:"price" validate:"omitempty,required"`
	Active *bool   `json:"active" validate:"omitempty,boolean"`
}

func (server *Server) updateProductItem(ctx *fiber.Ctx) error {
	params := &updateProductItemParamsRequest{}
	req := &updateProductItemJsonRequest{}

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

	arg := db.AdminUpdateProductItemParams{
		AdminID:    authPayload.AdminID,
		ID:         params.ProductItemID,
		ProductID:  req.ProductID,
		ProductSku: null.IntFromPtr(req.ProductSKU),
		// QtyInStock: null.IntFromPtr(req.QtyInStock),
		// ProductImage: null.StringFromPtr(&req.ProductImage),
		Price:  null.StringFromPtr(req.Price),
		Active: null.BoolFromPtr(req.Active),
	}

	productItem, err := server.store.AdminUpdateProductItem(ctx.Context(), arg)
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

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
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
	Limit            int32     `query:"limit" validate:"required,min=5,max=10"`
	CategoryID       null.Int  `query:"category_id" validate:"omitempty"`
	BrandID          null.Int  `query:"brand_id" validate:"omitempty"`
	SizeID           null.Int  `query:"size_id" validate:"omitempty"`
	ColorID          null.Int  `query:"color_id" validate:"omitempty"`
	IsNew            null.Bool `query:"is_new" validate:"omitempty"`
	IsPromoted       null.Bool `query:"is_promoted" validate:"omitempty"`
	IsFeatured       null.Bool `query:"is_featured" validate:"omitempty"`
	IsQtyLimited     null.Bool `query:"is_qty_limited" validate:"omitempty"`
	OrderByLowPrice  null.Bool `query:"order_by_low_price" validate:"omitempty"`
	OrderByHighPrice null.Bool `query:"order_by_high_price" validate:"omitempty"`
	OrderByNew       null.Bool `query:"order_by_new" validate:"omitempty"`
	OrderByOld       null.Bool `query:"order_by_old" validate:"omitempty"`
}

func (server *Server) listProductItemsV2(ctx *fiber.Ctx) error {
	query := &listProductItemsV2QueryRequest{}
	// var maxPage int64

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsV2Params{
		Limit:      query.Limit,
		CategoryID: query.CategoryID,
		BrandID:    query.BrandID,
		ColorID:    query.ColorID,
		// SizeID:           query.SizeID,
		IsNew:            query.IsNew,
		IsPromoted:       query.IsPromoted,
		IsFeatured:       query.IsFeatured,
		IsQtyLimited:     query.IsQtyLimited,
		OrderByLowPrice:  query.OrderByLowPrice,
		OrderByHighPrice: query.OrderByHighPrice,
		OrderByNew:       query.OrderByNew,
		OrderByOld:       query.OrderByOld,
	}

	productItems, err := server.store.ListProductItemsV2(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsV2Row{})
		return nil
	}
	// if len(productItems) != 0 {
	// 	maxPage = int64(math.Ceil(float64(productItems[0].TotalCount) / float64(query.Limit)))
	// 	// ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	// ctx.Status(fiber.StatusOK).JSON(productItems)
	// } else {
	// 	maxPage = 0
	// 	// ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	// ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsV2Row{})
	// }

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type listProductItemsNextPageQueryRequest struct {
	ProductItemCursor int64       `query:"product_item_cursor" validate:"required,min=1"`
	ProductCursor     int64       `query:"product_cursor" validate:"required,min=1"`
	Limit             int32       `query:"limit" validate:"required,min=5,max=10"`
	CategoryID        null.Int    `query:"category_id" validate:"omitempty"`
	BrandID           null.Int    `query:"brand_id" validate:"omitempty"`
	SizeID            null.Int    `query:"size_id" validate:"omitempty"`
	ColorID           null.Int    `query:"color_id" validate:"omitempty"`
	IsNew             null.Bool   `query:"is_new" validate:"omitempty"`
	IsPromoted        null.Bool   `query:"is_promoted" validate:"omitempty"`
	IsFeatured        null.Bool   `query:"is_featured" validate:"omitempty"`
	IsQtyLimited      null.Bool   `query:"is_qty_limited" validate:"omitempty"`
	PriceCursor       null.String `query:"price_cursor" validate:"omitempty"`
	OrderByLowPrice   null.Bool   `query:"order_by_low_price" validate:"omitempty"`
	OrderByHighPrice  null.Bool   `query:"order_by_high_price" validate:"omitempty"`
	CreatedAtCursor   null.String `query:"created_at_cursor" validate:"omitempty"`
	OrderByNew        null.Bool   `query:"order_by_new" validate:"omitempty"`
	OrderByOld        null.Bool   `query:"order_by_old" validate:"omitempty"`
}

func (server *Server) listProductItemsNextPage(ctx *fiber.Ctx) error {
	query := &listProductItemsNextPageQueryRequest{}
	// var maxPage int64

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	createdAt := util.ParseTimeOrNil(productTimeLayout, query.CreatedAtCursor.String)

	arg := db.ListProductItemsNextPageParams{
		Limit:         query.Limit,
		ProductItemID: query.ProductItemCursor,
		ProductID:     query.ProductCursor,
		CategoryID:    query.CategoryID,
		BrandID:       query.BrandID,
		ColorID:       query.ColorID,
		// SizeID:           query.SizeID,
		IsNew:            query.IsNew,
		IsPromoted:       query.IsPromoted,
		IsFeatured:       query.IsFeatured,
		IsQtyLimited:     query.IsQtyLimited,
		Price:            query.PriceCursor,
		OrderByHighPrice: query.OrderByHighPrice,
		OrderByLowPrice:  query.OrderByLowPrice,
		CreatedAt:        null.TimeFromPtr(createdAt),
		OrderByNew:       query.OrderByNew,
		OrderByOld:       query.OrderByOld,
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
	if len(productItems) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductItemsNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Paginated Search API //////////////

type searchProductItemsQueryRequest struct {
	Limit int32  `query:"limit" validate:"required,min=5,max=10"`
	Query string `query:"query" validate:"omitempty,required,alphanumunicode_space"`
}

func (server *Server) searchProductItems(ctx *fiber.Ctx) error {
	query := &searchProductItemsQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
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
	if len(productItems) == 0 {
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
	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type searchProductItemsNextPageQueryRequest struct {
	ProductItemCursor int64  `query:"product_item_cursor" validate:"required,min=1"`
	ProductCursor     int64  `query:"product_cursor" validate:"required,min=1"`
	Limit             int32  `query:"limit" validate:"required,min=5,max=10"`
	Query             string `query:"query" validate:"omitempty,required,alphanumunicode_space"`
}

func (server *Server) searchProductItemsNextPage(ctx *fiber.Ctx) error {
	query := &searchProductItemsNextPageQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.SearchProductItemsNextPageParams{
		Limit:         query.Limit,
		ProductItemID: query.ProductItemCursor,
		ProductID:     query.ProductCursor,
		Query:         query.Query,
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
	if len(productItems) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductItemsNextPageRow{})
		return nil
	}
	// if len(productItems) > 0 {
	// 	pagesNumber := float64(productItems[0].TotalCount) / float64(query.Limit)
	// 	if len(productItems) == int(productItems[0].TotalCount) {
	// 		ctx.Set("Max-Page", "0")
	// 	} else {
	// 		maxPage := int64(math.Ceil(pagesNumber))
	// 		ctx.Set("Max-Page", fmt.Sprint(maxPage))
	// 	}
	// } else {
	// 	ctx.Set("Max-Page", f"0"))
	// }

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Promotions List API //////////////

type listProductItemsWithPromotionsQueryRequest struct {
	ProductID int64 `query:"product_id" validate:"required,min=1"`
	Limit     int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithPromotions(ctx *fiber.Ctx) error {
	query := &listProductItemsWithPromotionsQueryRequest{}
	// var maxPage int64

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithPromotionsParams{
		ProductID: query.ProductID,
		Limit:     query.Limit,
	}

	productItems, err := server.store.ListProductItemsWithPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsWithPromotionsRow{})
		return nil
	}

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type listProductItemsWithPromotionsNextPageQueryRequest struct {
	ProductItemCursor int64 `query:"product_item_cursor" validate:"required,min=1"`
	ProductCursor     int64 `query:"product_cursor" validate:"required,min=1"`
	Limit             int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithPromotionsNextPage(ctx *fiber.Ctx) error {
	query := &listProductItemsWithPromotionsNextPageQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithPromotionsNextPageParams{
		Limit:         query.Limit,
		ProductItemID: query.ProductItemCursor,
		ProductID:     query.ProductCursor,
	}

	productItems, err := server.store.ListProductItemsWithPromotionsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductItemsNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Products with Brand Promotions List API //////////////

type listProductItemsWithBrandPromotionsQueryRequest struct {
	BrandID int64 `query:"brand_id" validate:"required,min=1"`
	Limit   int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithBrandPromotions(ctx *fiber.Ctx) error {
	query := &listProductItemsWithBrandPromotionsQueryRequest{}
	// var maxPage int64

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithBrandPromotionsParams{
		BrandID: query.BrandID,
		Limit:   query.Limit,
	}

	productItems, err := server.store.ListProductItemsWithBrandPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsWithBrandPromotionsRow{})
		return nil
	}

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type listProductItemsWithBrandPromotionsNextPageQueryRequest struct {
	BrandID           int64 `query:"brand_id" validate:"required,min=1"`
	ProductItemCursor int64 `query:"product_item_cursor" validate:"required,min=1"`
	ProductCursor     int64 `query:"product_cursor" validate:"required,min=1"`
	Limit             int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithBrandPromotionsNextPage(ctx *fiber.Ctx) error {
	query := &listProductItemsWithBrandPromotionsNextPageQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithBrandPromotionsNextPageParams{
		Limit:         query.Limit,
		ProductItemID: query.ProductItemCursor,
		ProductID:     query.ProductCursor,
		BrandID:       query.BrandID,
	}

	productItems, err := server.store.ListProductItemsWithBrandPromotionsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductItemsNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

//////////////* Products with Category Promotions List API //////////////

type listProductItemsWithCategoryPromotionsQueryRequest struct {
	CategoryID int64 `query:"category_id" validate:"required,min=1"`
	Limit      int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithCategoryPromotions(ctx *fiber.Ctx) error {
	query := &listProductItemsWithCategoryPromotionsQueryRequest{}
	// var maxPage int64

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithCategoryPromotionsParams{
		CategoryID: query.CategoryID,
		Limit:      query.Limit,
	}

	productItems, err := server.store.ListProductItemsWithCategoryPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsWithCategoryPromotionsRow{})
		return nil
	}

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

type listProductItemsWithCategoryPromotionsNextPageQueryRequest struct {
	CategoryID        int64 `query:"category_id" validate:"required,min=1"`
	ProductItemCursor int64 `query:"product_item_cursor" validate:"required,min=1"`
	ProductCursor     int64 `query:"product_cursor" validate:"required,min=1"`
	Limit             int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductItemsWithCategoryPromotionsNextPage(ctx *fiber.Ctx) error {
	query := &listProductItemsWithCategoryPromotionsNextPageQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductItemsWithCategoryPromotionsNextPageParams{
		Limit:         query.Limit,
		ProductItemID: query.ProductItemCursor,
		ProductID:     query.ProductCursor,
		CategoryID:    query.CategoryID,
	}

	productItems, err := server.store.ListProductItemsWithCategoryPromotionsNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductItemsNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productItems[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}

////////////// * Products with Best Sells Per Month List API //////////////

type listProductItemsWithBestSalesQueryRequest struct {
	Limit int32 `query:"limit" validate:"required,min=1,max=50"`
}

func (server *Server) listProductItemsWithBestSales(ctx *fiber.Ctx) error {
	query := &listProductItemsWithBestSalesQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productItems, err := server.store.ListProductItemsWithBestSales(ctx.Context(), query.Limit)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productItems) == 0 {
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductItemsWithBestSalesRow{})
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productItems)
	return nil

}
