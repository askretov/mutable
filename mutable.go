package mutable

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-ext/logger"
	"github.com/pquerna/ffjson/ffjson"
)

// LevelSeparator is a separator of path levels through nested structs (eg. FieldA/NestedStructFieldZ)
var LevelSeparator = "/"

// Mutable provides object changes tracking features and the way to set values to the struct dynamically
// by a destination field name (including nested structs)
type Mutable struct {
	originalState interface{}   // Original state of an object
	target        interface{}   // Pointer to a target object
	MutableStatus Status        `json:"-"` // Mutable status of an object
	ChangedFields ChangedFields `json:"-"` // Changed fields data
}

const (
	flagIgnore         = "ignore"
	flagDeepAnalyze    = "deep"
	mutTypeName        = "mutable.Mutable"
	mutFieldName       = "Mutable"
	mutStatusFieldName = "MutableStatus"
	mutTagName         = "mutable"
)

// ResetMutableState resets current mutable state and updates original state with given self value
// It also resets all nested mutable objects
func (m *Mutable) ResetMutableState(self interface{}) error {
	if reflect.ValueOf(self).Kind() != reflect.Ptr {
		return errNotPointer
	}
	// Set a target
	if m.target == nil {
		m.target = self
	}
	// Update mutable status
	m.MutableStatus = NotChanged
	// Reset original state
	m.originalState = nil
	m.originalState = reflect.ValueOf(self).Elem().Interface()
	// Reset changed fields arrays
	m.ChangedFields = ChangedFields{}
	// Reset all nested mutable objects
	reflectValue := reflect.ValueOf(m.target).Elem()
	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		if field.Kind() == reflect.Ptr {
			field = reflectValue.Field(i).Elem()
		}
		if !field.IsValid() || field.Kind() != reflect.Struct {
			continue
		}
		if isMutable(field) && field.CanInterface() {
			if err := field.Addr().Interface().(Mutabler).ResetMutableState(field.Addr().Interface()); err != nil {
				logger.Error(err)
				return errNestedResetError
			}
		}
	}
	return nil
}

// SetValue sets a value for given field by its name.
// JSON tag value will be used to find an appropriate field as a default source for a name, otherwise
// a real (as it stated in struct) field name will be used.
// Package var LevelSeparator value used as a separator for nested structs (eg. car/engine/price)
func (m *Mutable) SetValue(fieldName string, value interface{}) error {
	// Try to set a value
	return trySetValueToObject(reflect.ValueOf(m.target).Elem(), "", fieldName, value)
}

// AnalyzeChanges analyzes changes of a target object and returns changed fields data
func (m *Mutable) AnalyzeChanges() ChangedFields {
	return tryAnalyzeChanges(reflect.ValueOf(m.target).Elem(), reflect.ValueOf(m.originalState))
}

// trySetValueToObject tries to set a value to a destination field of given object
func trySetValueToObject(object reflect.Value, levelPrefix, dstFieldName string, value interface{}) error {
	// Iterate over struct fields
	for z := 0; z < object.NumField(); z++ {
		field := object.Field(z)
		if field.Kind() == reflect.Ptr {
			// Get value the pointer points to
			field = object.Field(z).Elem()
		}
		// Get current field's json name
		fieldName, tagExists := object.Type().Field(z).Tag.Lookup("json")
		if !tagExists {
			// Get a field name from a struct metadata
			fieldName = object.Type().Field(z).Name
		}
		if len(levelPrefix) > 0 {
			// Prepend a level prefix
			fieldName = levelPrefix + LevelSeparator + fieldName
		}
		if fieldName == dstFieldName {
			if err := trySetValueToField(field, value); err != nil {
				logger.Warningf("Error: %s, Field: %s", err, fieldName)
				return errCannotSetValue(fieldName, value)
			}
			return nil
		} else if field.Kind() == reflect.Struct && strings.HasPrefix(dstFieldName, fieldName+LevelSeparator) {
			// Go down recursively
			return trySetValueToObject(field, fieldName, dstFieldName, value)
		}
	}
	return errCannotFind(dstFieldName)
}

// trySetValueToField sets the value to the given field
func trySetValueToField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return errNotSettable
	}
	if !field.CanInterface() {
		return errNotInterfaceable
	}
	var fieldType = reflect.TypeOf(field.Interface())
	if fieldType == reflect.TypeOf(value) {
		// Set a value
		field.Set(reflect.ValueOf(value))
	} else {
		var srcValue []byte
		switch value.(type) {
		case string:
			srcValue = []byte(value.(string))
		case []byte:
			// Try to parse a []byte value into destination type
			srcValue = value.([]byte)
		default:
			// Unsupported type
			return errUnsupportedType
		}
		// Try to parse a string value into destination type
		parsedValue, err := parseValue(srcValue, fieldType)
		if err != nil {
			logger.Error(err)
			return errCannotParse
		}
		// Set a parsed value
		field.Set(reflect.ValueOf(parsedValue))
	}
	return nil
}

// parseValue returns a value parsed into destination type dstType.
// Value should be a valid JSON value
func parseValue(value []byte, dstType reflect.Type) (interface{}, error) {
	if json.Valid(value) {
		dstValue := reflect.New(dstType)
		if err := ffjson.Unmarshal(value, dstValue.Interface()); err != nil {
			return nil, err
		}
		return dstValue.Elem().Interface(), nil
	}
	return nil, errNotJSON
}

// isMutable reports whether a value is a mutable object
func isMutable(value reflect.Value) bool {
	// Try to get a Mutable field by name
	mutableField := value.FieldByName(mutFieldName)
	if mutableField.IsValid() {
		switch mutableField.Interface().(type) {
		case Mutable:
			return true
		}
	}
	return false
}

// setMutableStatus sets a status for a given value
func setMutableStatus(value reflect.Value, status Status) {
	mf := value.FieldByName(mutFieldName)
	if mf.IsValid() {
		mf.FieldByName(mutStatusFieldName).Set(reflect.ValueOf(status))
	}
}

// appendChangedFields appends changed fields to a given object
func appendChangedFields(object reflect.Value, changedFields ChangedFields) {
	// Iterate over struct fields
	for z := 0; z < object.NumField(); z++ {
		switch object.Field(z).Interface().(type) {
		case Mutable:
			dst := object.Field(z).Interface().(Mutable).ChangedFields
			// Check dst for nil
			if dst == nil {
				dst = ChangedFields{}
			}
			for _, field := range changedFields {
				dst[field.Name] = field
			}
			return
		}
	}
}

// tryAnalyzeChanges analyzes changes of a target object and returns changed fields data
func tryAnalyzeChanges(currentValue, originalValue reflect.Value) (changedFields ChangedFields) {
	changedFields = ChangedFields{}
	// Iterate over struct fields
	for z := 0; z < currentValue.NumField(); z++ {
		// Get current and original fields
		currentField := currentValue.Field(z)
		originalField := originalValue.Field(z)
		if currentField.Kind() == reflect.Ptr {
			// Get value the pointer points to
			currentField = currentValue.Field(z).Elem()
			originalField = originalValue.Field(z).Elem()
		}
		// Get current field metadata
		currentFieldMeta := currentValue.Type().Field(z)
		tagValue, _ := currentFieldMeta.Tag.Lookup(mutTagName)
		// Check the field for ignored flag
		ignored := strings.Contains(tagValue, flagIgnore)
		// Check ignored fields and Mutable field itself
		if currentFieldMeta.Type.String() == mutTypeName || ignored {
			// Pass through Mutable itself and ignored fields
			continue
		}
		// Check whether a field has deep analyze flag
		isDeepAnalyze := strings.Contains(tagValue, flagDeepAnalyze) && currentField.Kind() == reflect.Struct

		// Analyze the field changes
		switch {
		case !currentField.IsValid() || !originalField.IsValid():
			// Current or original field is not valid
			if changedField := analyzeNotValid(currentFieldMeta.Name, currentField, originalField); changedField != nil {
				changedFields[changedField.Name] = changedField
			}
		case isDeepAnalyze:
			// Deep analyze case
			if nestedChangedFields := analyzeDeep(currentField, originalField); len(nestedChangedFields) > 0 {
				changedFields[currentFieldMeta.Name] = &ChangedField{
					Name:         currentFieldMeta.Name,
					NestedFields: nestedChangedFields,
				}
			}
		default:
			// Regular analyze case (simple value field)
			if changedField := analyzeRegular(currentFieldMeta.Name, currentField, originalField); changedField != nil {
				changedFields[changedField.Name] = changedField
			}
		}

	}
	// Set changed fields data to the current object
	if len(changedFields) > 0 {
		if isMutable(currentValue) {
			appendChangedFields(currentValue, changedFields)
		}
	}
	return changedFields
}

// analyzeNotValid analyzes a case when current or original value is not valid
func analyzeNotValid(fieldName string, current, original reflect.Value) *ChangedField {
	switch {
	case !current.IsValid() && !original.IsValid():
		// Nothing changed, both values are not valid
	case !current.IsValid() && original.IsValid():
		// Current is not valid
		return &ChangedField{
			Name:     fieldName,
			OldValue: original.Interface(),
			NewValue: nil,
		}
	case !original.IsValid() && current.IsValid():
		// Original is not valid
		return &ChangedField{
			Name:     fieldName,
			OldValue: nil,
			NewValue: current.Interface(),
		}
	}
	return nil
}

// analyzeDeep returns changed fields of deep analyze logic.
// Deep analyze logic is the analyze of every field changes of underlying struct (used only for struct values)
func analyzeDeep(current, original reflect.Value) (changedFields ChangedFields) {
	if !current.CanInterface() {
		return changedFields
	}
	// Check whether a nested struct is mutable
	isNestedStructMutable := isMutable(current)
	// Analyze nested struct
	if isNestedStructMutable {
		// Analyze with nested object's own mutable logic
		changedFields = current.Addr().Interface().(Mutabler).AnalyzeChanges()
	} else {
		// Analyze as non-mutable struct
		changedFields = tryAnalyzeChanges(current, original)
	}
	return changedFields
}

// analyzeRegular returns changed fields of regular analyze.
// Regular analyze logic is the direct comparison of current and original values
func analyzeRegular(fieldName string, current, original reflect.Value) *ChangedField {
	if !current.CanInterface() {
		return nil
	}
	var equals bool
	if _, ok := current.Interface().(Equaler); ok {
		// Compare with type's Equal method
		equals = current.Interface().(Equaler).Equal(original.Interface())
	} else {
		// Compare with reflect's DeepEqual
		equals = reflect.DeepEqual(current.Interface(), original.Interface())
	}
	if !equals {
		return &ChangedField{
			Name:     fieldName,
			OldValue: original.Interface(),
			NewValue: current.Interface(),
		}
	}
	return nil
}
