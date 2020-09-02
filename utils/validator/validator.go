package validator

import (
	"github.com/asaskevich/govalidator"
)

func ValidateStruct(st interface{}) (bool, error) {
	return govalidator.ValidateStruct(st)
}
