package mutable

import (
	"encoding/json"

	"github.com/go-ext/logger"
)

type ChangedFields map[string]ChangedField
// TODO: Migrate to pointers

type ChangedField struct {
	Name         string        `json:"-"`
	OldValue     interface{}   `json:"old_value"`
	NewValue     interface{}   `json:"new_value"`
	NestedFields ChangedFields `json:"nested_fields,omitempty"`
}

// Contains reports whether a needle already exists within c
func (c ChangedFields) Contains(needle *ChangedField) bool {

	if _, exists := c[needle.Name]; exists {
		return true
	}

	return false
}

// ContainsByFieldName reports whether a data for given field name exists within c
func (c ChangedFields) ContainsByFieldName(fieldName string) bool {

	if _, exists := c[fieldName]; exists {
		return true
	}

	return false
}

// FieldNames returns an array of changed field names
func (c ChangedFields) FieldNames() []string {

	var result = []string{}

	for _, field := range c {
		result = append(result, field.Name)
	}

	return result
}

// GetFieldByName returns ChangedField object by field name
func (c ChangedFields) GetFieldByName(fieldName string) ChangedField {

	if _, exists := c[fieldName]; exists {
		return c[fieldName]
	}

	return ChangedField{}
}

// JSON serializes c
func (c ChangedFields) JSON(pretty bool) []byte {

	var result []byte
	var err error

	// Serialize an event
	if pretty {
		result, err = json.MarshalIndent(c, "", "\t")
	} else {
		result, err = json.Marshal(c)
	}
	if err != nil {
		logger.Error(err)
	}

	return result
}

// String implements Stringer interface
func (c ChangedFields) String() string {
	return string(c.JSON(true))
}
