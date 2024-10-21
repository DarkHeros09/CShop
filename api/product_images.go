package api

import (
	"errors"
	"fmt"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/imagekit-developer/imagekit-go/api/media"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type createProductImagesParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductImagesJsonRequest struct {
	ProductImage1 string `json:"product_image_1" validate:"required,url"`
	ProductImage2 string `json:"product_image_2" validate:"required,url"`
	ProductImage3 string `json:"product_image_3" validate:"required,url"`
}

func (server *Server) createProductImages(ctx *fiber.Ctx) error {
	params := &createProductImagesParamsRequest{}
	req := &createProductImagesJsonRequest{}

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

	arg := db.AdminCreateProductImagesParams{
		AdminID:       authPayload.AdminID,
		ProductImage1: req.ProductImage1,
		ProductImage2: req.ProductImage2,
		ProductImage3: req.ProductImage3,
	}

	productImages, err := server.store.AdminCreateProductImages(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productImages)
	return nil
}

// Define a struct for the image URLs
type imageResponse struct {
	URL string `json:"url"`
}

type listproductImagesParamsResquest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type listproductImagesQueryRequest struct {
	Path string `query:"path" validate:"omitempty,required,alphaunicode"`
	Tag  string `query:"tag" validate:"omitempty,required,alphaunicode"`
}

func (server *Server) listproductImages(ctx *fiber.Ctx) error {
	params := &listproductImagesParamsResquest{}
	query := &listproductImagesQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	resp, err := server.ik.ListAndSearch(ctx.Context(), media.FilesParam{
		Path: query.Path,
		Tags: query.Tag})

	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	imagesURLs := make([]imageResponse, len(resp.Data))

	for i := 0; i < len(resp.Data); i++ {
		imagesURLs[i] = imageResponse{
			URL: resp.Data[i].Url,
		}
	}

	ctx.Status(fiber.StatusOK).JSON(imagesURLs)
	return nil
}

//////////////* Pagination List API //////////////

type listProductImagesV2QueryRequest struct {
	Limit int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductImagesV2(ctx *fiber.Ctx) error {
	query := &listProductImagesV2QueryRequest{}
	// var maxPage int64

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productImages, err := server.store.ListProductImagesV2(ctx.Context(), query.Limit)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productImages) == 0 {
		ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusOK).JSON([]db.ListProductImagesV2Row{})
		return nil
	}

	ctx.Set("Next-Available", fmt.Sprint(productImages[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productImages)
	return nil

}

type listProductImagesNextPageQueryRequest struct {
	ProductCursor int64 `query:"product_cursor" validate:"required,min=1"`
	Limit         int32 `query:"limit" validate:"required,min=5,max=10"`
}

func (server *Server) listProductImagesNextPage(ctx *fiber.Ctx) error {
	query := &listProductImagesNextPageQueryRequest{}
	// var maxPage int64

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductImagesNextPageParams{
		Limit: query.Limit,
		ID:    query.ProductCursor,
	}

	productImages, err := server.store.ListProductImagesNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	if len(productImages) == 0 {
		// ctx.Set("Next-Available", fmt.Sprint(false))
		ctx.Status(fiber.StatusNotFound).JSON(errorResponse(pgx.ErrNoRows))
		// ctx.Status(fiber.StatusNotFound).JSON([]db.ListProductImagesNextPageRow{})
		return nil
	}

	// ctx.Set("Max-Page", fmt.Sprint(maxPage))

	ctx.Set("Next-Available", fmt.Sprint(productImages[0].NextAvailable))
	ctx.Status(fiber.StatusOK).JSON(productImages)
	return nil

}

//////////////* Update API //////////////

type updateProductImagesParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateProductImagesJsonRequest struct {
	ProductImage1 *string `json:"product_image_1" validate:"omitempty,required,url"`
	ProductImage2 *string `json:"product_image_2" validate:"omitempty,required,url"`
	ProductImage3 *string `json:"product_image_3" validate:"omitempty,required,url"`
}

func (server *Server) updateProductImages(ctx *fiber.Ctx) error {
	params := &updateProductImagesParamsRequest{}
	req := &updateProductImagesJsonRequest{}

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

	arg := db.AdminUpdateProductImageParams{
		AdminID:       authPayload.AdminID,
		ID:            params.ID,
		ProductImage1: null.StringFromPtr(req.ProductImage1),
		ProductImage2: null.StringFromPtr(req.ProductImage2),
		ProductImage3: null.StringFromPtr(req.ProductImage3),
	}

	productImage, err := server.store.AdminUpdateProductImage(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productImage)
	return nil
}
