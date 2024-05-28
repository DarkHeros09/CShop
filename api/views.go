package api

import (
	// "strconv"

	// db "github.com/cshop/v3/db/sqlc"
	"github.com/gofiber/fiber/v2"
)

func (server *Server) resetPasswordPage(ctx *fiber.Ctx) error {

	emailId := ctx.FormValue("email_id")
	secretCode := ctx.FormValue("secret_code")

	// id, err := strconv.Atoi(emailId)
	// if err != nil {
	// 	return ctx.Render("not_found", fiber.Map{})
	// }

	// arg := db.GetResetPasswordUserIDByIDParams{
	// 	ID:         int64(id),
	// 	SecretCode: secretCode,
	// }

	// _, err = server.store.GetResetPasswordUserIDByID(ctx.Context(), arg)
	// if err != nil {
	// 	return ctx.Render("not_found", fiber.Map{})
	// }

	return ctx.Render("reset_password", fiber.Map{
		"EmailId":    emailId,
		"SecretCode": secretCode,
	})
}
