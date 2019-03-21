package mutable

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	errCannotSetValue = func(field string, value interface{}) error {
		return fmt.Errorf("cannot set value (%v) to the field (%v)", value, field)
	}
	errCannotFind = func(field string) error {
		return fmt.Errorf("cannot find a destination field (%v)", field)
	}
	errNotPointer       = errors.New("given value is not a Pointer type")
	errNestedResetError = errors.New("cannot reset nested mutable object")
	errNotSettable      = errors.New("field is not settable")
	errNotInterfaceable = errors.New("field is not interfaceable")
	errCannotParse      = errors.New("cannot parse value")
	errUnsupportedType  = func(fieldType reflect.Type, value interface{}) error {
		return fmt.Errorf("unsupported value type (%T) for a field (%v)", value, fieldType)
	}
	errNotJSON = errors.New("not a valid JSON value")
)

// IsCannotSetErr reports whether an err is a errCannotSetValue error
func IsCannotSetErr(err error) bool {
	return strings.HasPrefix(err.Error(), "cannot set value")
}

// IsCannotFindErr reports whether an err is a errCannotFind error
func IsCannotFindErr(err error) bool {
	return strings.HasPrefix(err.Error(), "cannot find a destination field")
}
