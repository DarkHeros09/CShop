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
type createProductConfigurationParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

type createProductConfigurationJsonRequest struct {
	VariationOptionID int64 `json:"variation_id" validate:"required,min=1"`
}

func (server *Server) createProductConfiguration(ctx *fiber.Ctx) error {
	params := &createProductConfigurationParamsRequest{}
	req := &createProductConfigurationJsonRequest{}

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

	arg := db.CreateProductConfigurationParams{
		ProductItemID:     params.ProductItemID,
		VariationOptionID: req.VariationOptionID,
	}

	productConfiguration, err := server.store.CreateProductConfiguration(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productConfiguration)
	return nil
}

//////////////* Get API //////////////

type getProductConfigurationParamsRequest struct {
	ProductItemID     int64 `params:"itemId" validate:"required,min=1"`
	VariationOptionID int64 `params:"variationId" validate:"required,min=1"`
}

func (server *Server) getProductConfiguration(ctx *fiber.Ctx) error {
	params := &getProductConfigurationParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetProductConfigurationParams{
		ProductItemID:     params.ProductItemID,
		VariationOptionID: params.VariationOptionID,
	}

	productConfiguration, err := server.store.GetProductConfiguration(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(productConfiguration)
	return nil
}

//////////////* List API //////////////

type listProductConfigurationsParamsRequest struct {
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

type listProductConfigurationsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listProductConfigurations(ctx *fiber.Ctx) error {
	params := &listProductConfigurationsParamsRequest{}
	query := &listProductConfigurationsQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductConfigurationsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	productConfigurations, err := server.store.ListProductConfigurations(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productConfigurations)
	return nil

}

//////////////* Update API //////////////

type updateProductConfigurationParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	ProductItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateProductConfigurationJsonRequest struct {
	VariationOptionID int64 `json:"variation_id" validate:"omitempty,required,min=1"`
}

func (server *Server) updateProductConfiguration(ctx *fiber.Ctx) error {
	params := &updateProductConfigurationParamsRequest{}
	req := &updateProductConfigurationJsonRequest{}

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

	arg := db.UpdateProductConfigurationParams{
		VariationOptionID: null.IntFromPtr(&req.VariationOptionID),
		ProductItemID:     params.ProductItemID,
	}

	productConfiguration, err := server.store.UpdateProductConfiguration(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productConfiguration)
	return nil
}

//////////////* Delete API //////////////

type deleteProductConfigurationParamsRequest struct {
	AdminID           int64 `params:"adminId" validate:"required,min=1"`
	ProductItemID     int64 `params:"itemId" validate:"required,min=1"`
	VariationOptionID int64 `params:"variationId" validate:"omitempty,required,min=1"`
}

func (server *Server) deleteProductConfiguration(ctx *fiber.Ctx) error {
	params := &deleteProductConfigurationParamsRequest{}

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

	arg := db.DeleteProductConfigurationParams{
		ProductItemID:     params.ProductItemID,
		VariationOptionID: params.VariationOptionID,
	}

	err := server.store.DeleteProductConfiguration(ctx.Context(), arg)
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
