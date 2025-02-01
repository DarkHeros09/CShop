package api

import (
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Custom validation function
func IsAlphanumUnicodeWithSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, c := range value {
		// Allow letters, numbers, and spaces
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

type Input struct {
	params any
	req    any
	query  any
}

func (server *Server) parseAndValidate(ctx *fiber.Ctx, input Input) error {

	switch {
	case input.params != nil && input.req == nil && input.query == nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := server.validate.Struct(input.params); err != nil {
			return err
		}
		return nil

	case input.params == nil && input.req != nil && input.query == nil:
		if err := ctx.BodyParser(input.req); err != nil {
			return err
		}
		if err := server.validate.Struct(input.req); err != nil {
			return err
		}
		return nil

	case input.params == nil && input.req == nil && input.query != nil:
		if err := ctx.QueryParser(input.query); err != nil {
			return err
		}
		if err := server.validate.Struct(input.query); err != nil {
			return err
		}
		return nil

	case input.params != nil && input.req != nil && input.query == nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := server.validate.Struct(input.params); err != nil {
			return err
		}

		if err := ctx.BodyParser(input.req); err != nil {
			return err
		}
		if err := server.validate.Struct(input.req); err != nil {
			return err
		}
		return nil

	case input.params != nil && input.req == nil && input.query != nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := server.validate.Struct(input.params); err != nil {
			return err
		}

		if err := ctx.QueryParser(input.query); err != nil {
			return err
		}
		if err := server.validate.Struct(input.query); err != nil {
			return err
		}
		return nil
	}
	return nil
}
