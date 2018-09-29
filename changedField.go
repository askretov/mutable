package mutable

import (
	"encoding/json"

	"github.com/go-ext/logger"
)

// ChangedFields is a map of ChangedField objects with a field name as a key
type ChangedFields map[string]*ChangedField

// ChangedField contains struct's fields changes data
type ChangedField struct {
	Name         string        `json:"-"`                       // Field name
	OldValue     interface{}   `json:"old_value"`               // Old value
	NewValue     interface{}   `json:"new_value"`               // New value
	NestedFields ChangedFields `json:"nested_fields,omitempty"` // Nested fields changes data (if a field is a struct and has "mutable:deep" tag value)
}

// Contains reports whether a field with fieldName exists within c
func (c ChangedFields) Contains(fieldName string) bool {
	_, exists := c[fieldName]
	return exists
}

// Keys returns an array of changed field names
func (c ChangedFields) Keys() []string {
	var result = make([]string, 0, len(c))
	for _, field := range c {
		result = append(result, field.Name)
	}
	return result
}

// GetField returns ChangedField object by field name
func (c ChangedFields) GetField(fieldName string) *ChangedField {
	return c[fieldName]
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

// String implements Stringer interface for ChangedFields
func (c ChangedFields) String() string {
	return string(c.JSON(true))
}
