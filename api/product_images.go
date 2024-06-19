package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/imagekit-developer/imagekit-go/api/media"
)

func (server *Server) listproductImages(ctx *fiber.Ctx) error {

	resp, err := server.ik.ListAndSearch(ctx.Context(), media.FilesParam{})

	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(resp)
	return nil
}
