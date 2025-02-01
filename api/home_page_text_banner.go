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
type createHomePageTextBannerParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createHomePageTextBannerJsonRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (server *Server) createHomePageTextBanner(ctx *fiber.Ctx) error {
	params := &createHomePageTextBannerParamsRequest{}
	req := &createHomePageTextBannerJsonRequest{}

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

	arg := db.CreateHomePageTextBannerParams{
		Name:        req.Name,
		Description: req.Description,
		AdminID:     authPayload.AdminID,
	}

	textBanner, err := server.store.CreateHomePageTextBanner(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(textBanner)
	return nil
}

//////////////* Get API //////////////

type getHomePageTextBannerParamsRequest struct {
	HomePageTextBannerID int64 `params:"textBannerId" validate:"required,min=1"`
}

func (server *Server) getHomePageTextBanner(ctx *fiber.Ctx) error {
	params := &getHomePageTextBannerParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	textBanner, err := server.store.GetHomePageTextBanner(ctx.Context(), params.HomePageTextBannerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(textBanner)
	return nil
}

//////////////* List API //////////////

// type listHomePageTextBannersQueryRequest struct {
// 	PageID   int32 `query:"page_id" validate:"required,min=1"`
// 	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
// }

func (server *Server) listHomePageTextBanners(ctx *fiber.Ctx) error {
	// query := &listHomePageTextBannersQueryRequest{}

	// if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
	// 	ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	// 	return nil
	// }

	// arg := db.ListHomePageTextBannersParams{
	// 	Limit:  query.PageSize,
	// 	Offset: (query.PageID - 1) * query.PageSize,
	// }
	textBanners, err := server.store.ListHomePageTextBanners(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(textBanners)
	return nil

}

//////////////* Update API //////////////

type updateHomePageTextBannerParamsRequest struct {
	AdminID              int64 `params:"adminId" validate:"required,min=1"`
	HomePageTextBannerID int64 `params:"textBannerId" validate:"required,min=1"`
}

type updateHomePageTextBannerJsonRequest struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
}

func (server *Server) updateHomePageTextBanner(ctx *fiber.Ctx) error {
	params := &updateHomePageTextBannerParamsRequest{}
	req := &updateHomePageTextBannerJsonRequest{}

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

	arg := db.UpdateHomePageTextBannerParams{
		ID:          params.HomePageTextBannerID,
		Name:        null.StringFromPtr(req.Name),
		Description: null.StringFromPtr(req.Description),
		AdminID:     authPayload.AdminID,
	}

	textBanner, err := server.store.UpdateHomePageTextBanner(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(textBanner)
	return nil
}

//////////////* Delete API //////////////

type deleteHomePageTextBannerParamsRequest struct {
	AdminID              int64 `params:"adminId" validate:"required,min=1"`
	HomePageTextBannerID int64 `params:"textBannerId" validate:"required,min=1"`
}

func (server *Server) deleteHomePageTextBanner(ctx *fiber.Ctx) error {
	params := &deleteHomePageTextBannerParamsRequest{}

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

	arg := db.DeleteHomePageTextBannerParams{
		ID:      params.HomePageTextBannerID,
		AdminID: authPayload.AdminID,
	}

	err := server.store.DeleteHomePageTextBanner(ctx.Context(), arg)
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
