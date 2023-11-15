package validator

import (
	"github.com/go-playground/validator/v10"
)

type validatorLib struct {
	validator *validator.Validate
}

type Interface interface {
	ValidateStruct(interface{}) error
}

func Init() Interface {
	return &validatorLib{validator: validator.New()}
}

func (v *validatorLib) ValidateStruct(data interface{}) error {
	return v.validator.Struct(data)
}
