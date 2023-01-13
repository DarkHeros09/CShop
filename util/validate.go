package util

import "github.com/go-playground/validator/v10"

func ValidateStruct(s any) error {
	validate := validator.New()
	err := validate.Struct(s)
	return err
}
