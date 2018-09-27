package mutable

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errCannotSetValue = func(field string, value interface{}) error {
		return fmt.Errorf("cannot set value (%s) for the field (%s)", value, field)
	}
	errCannotFind = func(field string) error {
		return fmt.Errorf("cannot find a destination field (%s)", field)
	}
	errNotPointer       = errors.New("given value is not a Pointer type")
	errNestedResetError = errors.New("cannot reset nested mutable object")
)

// IsCannotSetErr reports whether an err is a errCannotSetValue error
func IsCannotSetErr(err error) bool {
	return strings.HasPrefix(err.Error(), "cannot set value")
}

// IsCannotFindErr reports whether an err is a errCannotFind error
func IsCannotFindErr(err error) bool {
	return strings.HasPrefix(err.Error(), "cannot find a destination field")
}