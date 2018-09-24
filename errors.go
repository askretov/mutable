package mutable

import "fmt"

var (
	errCannotSetValue = func(field string, value interface{}) error {
		return fmt.Errorf("cannot set value (%s) for the field (%s)", value, field)
	}
	errCannotFind = func(field string) error {
		return fmt.Errorf("cannot find suitable field (%s)", field)
	}
)
