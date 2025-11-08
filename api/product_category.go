package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// ////////////* Create API //////////////
type createProductCategoryParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductCategoryJsonRequest struct {
	ParentCategoryID *int64 `json:"parent_category_id" validate:"omitempty,required,min=1"`
	CategoryName     string `json:"category_name" validate:"required,alphanumunicode_space"`
	CategoryImage    string `json:"category_image" validate:"required,http_url"`
}

func (server *Server) createProductCategory(ctx *fiber.Ctx) error {
	params := &createProductCategoryParamsRequest{}
	req := &createProductCategoryJsonRequest{}

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

	arg := db.CreateProductCategoryParams{
		ParentCategoryID: null.IntFromPtr(req.ParentCategoryID),
		CategoryName:     req.CategoryName,
		CategoryImage:    req.CategoryImage,
	}

	productCategory, err := server.store.CreateProductCategory(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productCategory)
	return nil
}

//////////////* Get API //////////////

type getProductCategoryParamsRequest struct {
	CategoryID int64 `params:"categoryId" validate:"required,min=1"`
}

func (server *Server) getProductCategory(ctx *fiber.Ctx) error {
	params := &getProductCategoryParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
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

	if productCategory == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productCategory)
	return nil
}

//////////////* List API //////////////

// type listProductCategoriesQueryRequest struct {
// 	PageID   int32 `query:"page_id" validate:"required,min=1"`
// 	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
// }

func (server *Server) listProductCategories(ctx *fiber.Ctx) error {
	// query := &listProductCategoriesQueryRequest{}

	// if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
	// 	ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	// 	return nil
	// }

	// arg := db.ListProductCategoriesParams{
	// 	Limit:  query.PageSize,
	// 	Offset: (query.PageID - 1) * query.PageSize,
	// }
	productCategories, err := server.store.ListProductCategories(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productCategories == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productCategories)
	return nil

}

//////////////* Update API //////////////

type updateProductCategoryParamsRequest struct {
	AdminID    int64 `params:"adminId" validate:"required,min=1"`
	CategoryID int64 `params:"categoryId" validate:"required,min=1"`
}

type updateProductCategoryJsonRequest struct {
	//? should be revised
	CategoryName     string `json:"category_name" validate:"omitempty,required"`
	ParentCategoryID *int64 `json:"parent_category_id" validate:"omitempty,required,min=1"`
}

func (server *Server) updateProductCategory(ctx *fiber.Ctx) error {
	params := &updateProductCategoryParamsRequest{}
	req := &updateProductCategoryJsonRequest{}

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

	arg := db.UpdateProductCategoryParams{
		ID:               params.CategoryID,
		CategoryName:     req.CategoryName,
		ParentCategoryID: null.IntFromPtr(req.ParentCategoryID),
	}

	productCategory, err := server.store.UpdateProductCategory(ctx.Context(), arg)
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
	params := &deleteProductCategoryParamsRequest{}
	req := &deleteProductCategoryJsonRequest{}

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

	arg := db.DeleteProductCategoryParams{
		ID:               params.CategoryID,
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
	}

	err := server.store.DeleteProductCategory(ctx.Context(), arg)
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
