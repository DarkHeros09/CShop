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

//////////////* Create API //////////////

type createUserReviewParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createUserReviewRequest struct {
	OrderedProductID int64 `json:"ordered_product_id" validate:"required,min=1"`
	RatingValue      int32 `json:"rating_value" validate:"required,min=0,max=5"`
}

func (server *Server) createUserReview(ctx *fiber.Ctx) error {
	params := &createUserReviewParamsRequest{}
	req := &createUserReviewRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateUserReviewParams{
		UserID:           authPayload.UserID,
		OrderedProductID: req.OrderedProductID,
		RatingValue:      req.RatingValue,
	}

	userReview, err := server.store.CreateUserReview(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(userReview)
	return nil
}

//////////////* Get API //////////////

type getUserReviewParamsRequest struct {
	UserID   int64 `params:"id" validate:"required,min=1"`
	ReviewID int64 `params:"reviewId" validate:"required,min=1"`
}

func (server *Server) getUserReview(ctx *fiber.Ctx) error {
	params := &getUserReviewParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetUserReviewParams{
		ID:     params.ReviewID,
		UserID: authPayload.UserID,
	}
	userReview, err := server.store.GetUserReview(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(userReview)
	return nil
}

//////////////* List API //////////////

type listUserReviewParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listUserReviewsRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listUserReviews(ctx *fiber.Ctx) error {
	params := &listUserReviewParamsRequest{}
	query := &listUserReviewsRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListUserReviewsParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	userReviews, err := server.store.ListUserReviews(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(userReviews)
	return nil
}

// ////////////* UPDATE API //////////////
type updateUserReviewParamsRequest struct {
	UserID   int64 `params:"id" validate:"required,min=1"`
	ReviewID int64 `params:"reviewId" validate:"required,min=1"`
}

type updateUserReviewJsonRequest struct {
	OrderedProductID *int64 `json:"ordered_product_id" validate:"omitempty,required,min=1"`
	RatingValue      *int64 `json:"rating_value" validate:"omitempty,required,min=0,max=5"`
}

func (server *Server) updateUserReview(ctx *fiber.Ctx) error {
	params := &updateUserReviewParamsRequest{}
	req := &updateUserReviewJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg1 := db.UpdateUserReviewParams{
		UserID:           authPayload.UserID,
		OrderedProductID: null.IntFromPtr(req.OrderedProductID),
		RatingValue:      null.IntFromPtr(req.RatingValue),
		ID:               params.ReviewID,
	}

	userReview, err := server.store.UpdateUserReview(ctx.Context(), arg1)
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

	ctx.Status(fiber.StatusOK).JSON(userReview)
	return nil
}

// ////////////* Delete API //////////////
type deleteUserReviewParamsRequest struct {
	UserID   int64 `params:"id" validate:"required,min=1"`
	ReviewID int64 `params:"reviewId" validate:"required,min=1"`
}

func (server *Server) deleteUserReview(ctx *fiber.Ctx) error {
	params := &deleteUserReviewParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteUserReviewParams{
		ID:     params.ReviewID,
		UserID: params.UserID,
	}

	_, err := server.store.DeleteUserReview(ctx.Context(), arg)
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
