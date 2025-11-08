package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// ////////////* Create API //////////////
type createProductBrandParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductBrandJsonRequest struct {
	BrandName  string `json:"brand_name" validate:"required,alphanumunicode_space"`
	BrandImage string `json:"brand_image" validate:"required,http_url"`
}

func (server *Server) createProductBrand(ctx *fiber.Ctx) error {
	params := &createProductBrandParamsRequest{}
	req := &createProductBrandJsonRequest{}

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

	arg := db.CreateProductBrandParams{
		BrandName:  req.BrandName,
		BrandImage: req.BrandImage,
	}

	productBrand, err := server.store.CreateProductBrand(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productBrand)
	return nil
}

//////////////* Get API //////////////

type getProductBrandParamsRequest struct {
	BrandID int64 `params:"brandId" validate:"required,min=1"`
}

func (server *Server) getProductBrand(ctx *fiber.Ctx) error {
	params := &getProductBrandParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productBrand, err := server.store.GetProductBrand(ctx.Context(), params.BrandID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productBrand == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productBrand)
	return nil
}

//////////////* List API //////////////

// type listProductBrandsQueryRequest struct {
// 	PageID   int32 `query:"page_id" validate:"required,min=1"`
// 	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
// }

func (server *Server) listProductBrands(ctx *fiber.Ctx) error {
	// query := &listProductBrandsQueryRequest{}

	// if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
	// 	ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	// 	return nil
	// }

	// arg := db.ListProductBrandsParams{
	// 	Limit:  query.PageSize,
	// 	Offset: (query.PageID - 1) * query.PageSize,
	// }
	productBrands, err := server.store.ListProductBrands(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productBrands == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productBrands)
	return nil

}

//////////////* Update API //////////////

type updateProductBrandParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	BrandID int64 `params:"brandId" validate:"required,min=1"`
}

type updateProductBrandJsonRequest struct {
	//? should be revised
	BrandName string `json:"brand_name" validate:"omitempty,required"`
}

func (server *Server) updateProductBrand(ctx *fiber.Ctx) error {
	params := &updateProductBrandParamsRequest{}
	req := &updateProductBrandJsonRequest{}

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

	arg := db.UpdateProductBrandParams{
		ID:        params.BrandID,
		BrandName: req.BrandName,
	}

	productBrand, err := server.store.UpdateProductBrand(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productBrand)
	return nil
}

//////////////* Delete API //////////////

type deleteProductBrandParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	BrandID int64 `params:"brandId" validate:"required,min=1"`
}

func (server *Server) deleteProductBrand(ctx *fiber.Ctx) error {
	params := &deleteProductBrandParamsRequest{}

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

	err := server.store.DeleteProductBrand(ctx.Context(), params.BrandID)
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
