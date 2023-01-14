package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ValidateStruct(s any) error {
	validate := validator.New()
	err := validate.Struct(s)
	return err
}

type Input struct {
	params any
	req    any
	query  any
}

func parseAndValidate(ctx *fiber.Ctx, input Input) error {

	switch {
	case input.params != nil && input.req == nil && input.query == nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := ValidateStruct(input.params); err != nil {
			return err
		}
		return nil

	case input.params == nil && input.req != nil && input.query == nil:
		if err := ctx.BodyParser(input.req); err != nil {
			return err
		}
		if err := ValidateStruct(input.req); err != nil {
			return err
		}
		return nil

	case input.params == nil && input.req == nil && input.query != nil:
		if err := ctx.QueryParser(input.query); err != nil {
			return err
		}
		if err := ValidateStruct(input.query); err != nil {
			return err
		}
		return nil

	case input.params != nil && input.req != nil && input.query == nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := ValidateStruct(input.params); err != nil {
			return err
		}

		if err := ctx.BodyParser(input.req); err != nil {
			return err
		}
		if err := ValidateStruct(input.req); err != nil {
			return err
		}
		return nil

	case input.params != nil && input.req == nil && input.query != nil:
		if err := ctx.ParamsParser(input.params); err != nil {
			return err
		}
		if err := ValidateStruct(input.params); err != nil {
			return err
		}

		if err := ctx.QueryParser(input.query); err != nil {
			return err
		}
		if err := ValidateStruct(input.query); err != nil {
			return err
		}
		return nil
	}
	return nil
}
