package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
)

const (
	authorizationHeaderKey       = "authorization"
	authorizationTypeBearer      = "bearer"
	authorizationUserPayloadKey  = "authorization_user_payload"
	authorizationAdminPayloadKey = "authorization_admin_payload"
)

func authMiddleware(tokenMaker token.Maker, admin bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorizationHeader := ctx.Get(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
			return nil
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
			return nil
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
			return nil
		}

		accessToken := fields[1]
		var adminPayload *token.AdminPayload
		var userPayload *token.UserPayload
		var err error
		if admin {
			adminPayload, err = tokenMaker.VerifyTokenForAdmin(accessToken)
			if err != nil {
				if err.Error() == "token has expired" {
					err = fmt.Errorf("access token has expired")
				}
				ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
				return nil
			}

			ctx.Locals(authorizationAdminPayloadKey, adminPayload)
			ctx.Next()
		}

		userPayload, err = tokenMaker.VerifyTokenForUser(accessToken)
		if err != nil {
			if err.Error() == "token has expired" {
				err = fmt.Errorf("access token has expired")
			}
			ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
			return nil
		}

		ctx.Locals(authorizationUserPayloadKey, userPayload)
		ctx.Next()

		return nil
	}
}
