package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/imagekit-developer/imagekit-go/api/media"
	"github.com/jackc/pgconn"
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
	Tag string `query:"tag" validate:"omitempty,required,alpha"`
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

	resp, err := server.ik.ListAndSearch(ctx.Context(), media.FilesParam{Tags: query.Tag})

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
