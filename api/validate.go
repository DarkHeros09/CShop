package api

import (
	"regexp"
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

// phoneRegex is a pre-compiled regular expression for phone number validation.
var phoneRegex = regexp.MustCompile(`^09[1-5]\d{7}$`)

// validatePhoneNumber is a custom validation function for the specific phone number format.
func validatePhoneNumber(fl validator.FieldLevel) bool {
	telephone := fl.Field().String()
	return phoneRegex.MatchString(telephone)
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
