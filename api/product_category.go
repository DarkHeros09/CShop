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
type createProductCategoryParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductCategoryJsonRequest struct {
	CategoryName     string `json:"category_name" validate:"required,alphanum"`
	ParentCategoryID int64  `json:"parent_category_id" validate:"required,min=1"`
}

func (server *Server) createProductCategory(ctx *fiber.Ctx) error {
	var params createProductCategoryParamsRequest
	var req createProductCategoryJsonRequest

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

	arg := db.CreateProductCategoryParams{
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
		CategoryName:     req.CategoryName,
	}

	productCategory, err := server.store.CreateProductCategory(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productCategory)
	return nil
}

//////////////* Get API //////////////

type getProductCategoryParamsRequest struct {
	CategoryID int64 `params:"categoryId" validate:"required,min=1"`
}

func (server *Server) getProductCategory(ctx *fiber.Ctx) error {
	var params getProductCategoryParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	productCategory, err := server.store.GetProductCategory(ctx.Context(), params.CategoryID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(productCategory)
	return nil
}

//////////////* List API //////////////

type listProductCategoriesQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listProductCategories(ctx *fiber.Ctx) error {
	var query listProductCategoriesQueryRequest

	if err := ctx.QueryParser(&query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductCategoriesParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	productCategorys, err := server.store.ListProductCategories(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productCategorys)
	return nil

}

//////////////* Update API //////////////

type updateProductCategoryParamsRequest struct {
	AdminID    int64 `params:"adminId" validate:"required,min=1"`
	CategoryID int64 `params:"categoryId" validate:"required,min=1"`
}

type updateProductCategoryJsonRequest struct {
	CategoryName     string `json:"category_name" validate:"omitempty,required"`
	ParentCategoryID int64  `json:"parent_category_id" validate:"omitempty,required,min=1"`
}

func (server *Server) updateProductCategory(ctx *fiber.Ctx) error {
	var params updateProductCategoryParamsRequest
	var req updateProductCategoryJsonRequest

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

	arg := db.UpdateProductCategoryParams{
		ID:               params.CategoryID,
		CategoryName:     req.CategoryName,
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
	}

	productCategory, err := server.store.UpdateProductCategory(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productCategory)
	return nil
}

//////////////* Delete API //////////////

type deleteProductCategoryParamsRequest struct {
	AdminID    int64 `params:"adminId" validate:"required,min=1"`
	CategoryID int64 `params:"categoryId" validate:"required,min=1"`
}

type deleteProductCategoryJsonRequest struct {
	ParentCategoryID int64 `json:"parent_category_id" validate:"required,min=1"`
}

func (server *Server) deleteProductCategory(ctx *fiber.Ctx) error {
	var params deleteProductCategoryParamsRequest
	var req deleteProductCategoryJsonRequest

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

	arg := db.DeleteProductCategoryParams{
		ID:               params.CategoryID,
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
	}

	err := server.store.DeleteProductCategory(ctx.Context(), arg)
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
