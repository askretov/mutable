package mutable

import (
	"fmt"
	"errors"
)

var (
	errCannotSetValue = func(field string, value interface{}) error {
		return fmt.Errorf("cannot set value (%s) for the field (%s)", value, field)
	}
	errCannotFind = func(field string) error {
		return fmt.Errorf("cannot find suitable field (%s)", field)
	}
	errNotPointer = errors.New("given value is not a Pointer type")
)
