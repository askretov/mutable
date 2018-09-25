package mutable

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-ext/helpers"
	"github.com/go-ext/logger"
)

type Mutable struct {
	// Original state of an object
	originalState interface{} `json:"-"`
	// Pointer to a target object
	target interface{} `json:"-"`
	// Mutable status of an object
	MutableStatus Status `json:"-"`
	// Changed fields data
	ChangedFields ChangedFields `json:"-"`
}

const (
	flagIgnore      = "ignore"
	flagDeepAnalyze = "deep"
	typeName        = "mutable.Mutable"
	mutTagName      = "mutable"
	levelSeparator  = "_"
)

// ResetMutableState resets current mutable status of an object and updates previous state with given currentState
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

	return nil
}

// SetValue sets a value for given field by its name
// JSON tag value will be used to find an appropriate field by its name
func (m *Mutable) SetValue(fieldName, value string) (error, bool) {
	// Try to find an appropriate field and set a new value
	if err, success := trySetValueForObject(reflect.ValueOf(m.target).Elem(), "", fieldName, value); err != nil {
		logger.Warning(err.Error())
		return errCannotSetValue(fieldName, value), false
	} else if !success {
		// If we have neither errors nor success state, it means that we didn't find a field
		return errCannotFind(fieldName), false
	}
	return nil, true
}

func trySetValueForObject(valueVar reflect.Value, currentContainerName, fieldJsonName, value string) (error, bool) {
	// Check whether current object is mutable
	var isMutableFlag = isMutable(valueVar)
	// Iterate over struct fields
	for z := 0; z < valueVar.NumField(); z++ {
		// Get current field json name
		currentFieldJsonName, _ := valueVar.Type().Field(z).Tag.Lookup("json")
		currentFieldJsonName = currentContainerName + currentFieldJsonName
		// Check whether a field has a container tag
		containerTag, _ := valueVar.Type().Field(z).Tag.Lookup("container")
		isContainer, _ := strconv.ParseBool(containerTag)

		if isContainer {
			// Go next level
			// Check whether nested struct is mutable
			var nestedStructIsMutable = isMutable(valueVar.Field(z))
			// Try to set a value
			if err, success := trySetValueForObject(valueVar.Field(z), currentFieldJsonName+"_", fieldJsonName, value); err != nil {
				logger.Warning(err)
				return errCannotSetValue(fieldJsonName, value), false
			} else if success {
				// Update mutable status
				if !nestedStructIsMutable && isMutableFlag {
					setMutableStatus(valueVar, Changed)
				}
				// Return success flag
				return nil, true
			}
		} else {
			// Check current field
			if currentFieldJsonName == fieldJsonName {
				// Try to set new value
				if err := setValueForField(valueVar.Field(z), value); err != nil {
					logger.Warning(err)
					return errCannotSetValue(fieldJsonName, value), false
				}
				// Update mutable status
				if isMutableFlag {
					setMutableStatus(valueVar, Changed)
				}
				// Return success flag
				return nil, true
			}
		}
	}

	return nil, false
}

func setValueForField(dstField reflect.Value, value string) error {
	switch fieldType := dstField.Interface().(type) {
	case string:
		dstField.SetString(value)
	case float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		dstField.SetFloat(val)
	case int, int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		dstField.SetInt(val)
	case uint, uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		dstField.SetUint(val)
	case bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		dstField.SetBool(val)
	case time.Time:
		// Try to parse time
		var layout = "2006-01-02"
		if len(value) > 10 {
			layout = "2006-01-02 15:04:05"
		}
		val, err := time.Parse(layout, value)
		if err != nil {
			// Try to parse and Unix timestamp
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			val = time.Unix(intVal, 0)
		}
		dstField.Set(reflect.ValueOf(val))
	case []int64:
		val := helper.SplitStringToInt64(value, ",", true)
		dstField.Set(reflect.ValueOf(val))
	case []string:
		val := helper.SplitString(value, ",", true)
		dstField.Set(reflect.ValueOf(val))
	default:
		logger.Error("not implemented field type: %T", fieldType)
	}

	return nil
}

// isMutable reports whether a value is a mutable object
func isMutable(value reflect.Value) bool {
	// Try to get a Mutable field by name
	mf := value.FieldByName("Mutable")
	if mf.IsValid() {
		switch mf.Interface().(type) {
		case Mutable:
			return true
		}
	}
	return false
}

// setMutableStatus sets a status for a given value
func setMutableStatus(value reflect.Value, status Status) {
	mf := value.FieldByName("Mutable")
	if mf.IsValid() {
		mf.FieldByName("MutableStatus").Set(reflect.ValueOf(status))
	}
}

func appendChangedField(valueVar reflect.Value, changedFields ChangedFields) {
	// Iterate over struct fields
	for z := 0; z < valueVar.NumField(); z++ {
		switch valueVar.Field(z).Interface().(type) {
		case Mutable:
			// Get current value
			currentContainer := valueVar.Field(z).Interface().(Mutable).ChangedFields
			// Append fields names to the current container
			for _, field := range changedFields {
				// Append a new value
				if !currentContainer.Contains(&field) {
					// Check currentContainer for nil
					if currentContainer == nil {
						currentContainer = ChangedFields{}
					}
					currentContainer[field.Name] = field
				}
			}
			// Set updated value back to the object
			valueVar.Field(z).FieldByName("ChangedFields").Set(reflect.ValueOf(currentContainer))

			return
		}
	}
}

// parseStructTag parses a struct tag field into map of tags
func parseStructTag(tag string) map[string]string {
	tags := map[string]string{}
	// Split tags string
	fields := strings.Fields(tag)
	for _, field := range fields {
		// Split every tag to key and value
		splitField := strings.Split(field, ":")
		if len(splitField) > 1 {
			// Add a tag data
			tags[splitField[0]] = strings.Replace(splitField[1], "\"", "", -1)
		}
	}
	return tags
}

// AnalyzeChanges analyzes changes of a target object and returns changed fields data
func (m *Mutable) AnalyzeChanges() ChangedFields {
	_, changes := tryAnalyzeChanges(reflect.ValueOf(m.target).Elem(), reflect.ValueOf(m.originalState), "")
	return changes
}

func tryAnalyzeChanges(currentValue, originalValue reflect.Value, currentLevelName string) (result bool, changedFields ChangedFields) {
	// TODO: Remove result, use just changedFields
	changedFields = ChangedFields{}
	// Check whether current object is mutable
	var isCurrentMutable = isMutable(currentValue)
	// Iterate over struct fields
	for z := 0; z < currentValue.NumField(); z++ {
		currentField := currentValue.Field(z)
		originalField := originalValue.Field(z)
		if currentField.Kind() == reflect.Ptr {
			currentField = currentValue.Field(z).Elem()
			originalField = originalValue.Field(z).Elem()
		}

		// Get current field metadata
		currentFieldMeta := currentValue.Type().Field(z)
		logger.Debugf("currentFieldName: %v, [%s]", currentFieldMeta.Name, currentFieldMeta.Type.String()) // TODO: DELETEME
		tagValue, _ := currentFieldMeta.Tag.Lookup(mutTagName)

		// Check field for ignored flag
		ignored := strings.Contains(tagValue, flagIgnore)
		if !currentField.IsValid() || currentFieldMeta.Type.String() == typeName || ignored { // TODO: test this part and move to == instead of contains
			// Pass through for Mutable itself and ignored fields
			continue
		}
		// Get current field name
		currentFieldName := currentLevelName + currentFieldMeta.Name
		// Check whether a field has deep analyze flag
		isDeepAnalyze := strings.Contains(tagValue, flagDeepAnalyze)
		if isDeepAnalyze && currentField.Kind() == reflect.Struct {
			// Check whether a nested struct is mutable
			nestedStructIsMutable := isMutable(currentField)
			// Analyze nested struct
			if changed, fields := tryAnalyzeChanges(currentField, originalField, currentFieldName+levelSeparator); changed {
				// Update mutable status
				if !nestedStructIsMutable && isCurrentMutable {
					setMutableStatus(currentValue, Changed)
					appendChangedField(currentValue, fields)
				}
				// Set true for result
				result = true
			}
		} else if currentField.CanInterface() {
			logger.Debugf("%v vs %v", currentField.Interface(), originalField.Interface())
			// Check current field
			var equals bool
			if _, ok := currentField.Interface().(Equaler); ok {
				// Compare with type's Equal method
				equals = currentField.Interface().(Equaler).Equal(originalField.Interface())
			} else {
				// Compare with reflect's DeepEqual
				equals = reflect.DeepEqual(currentField.Interface(), originalField.Interface())
			}
			if !equals {
				// Prepare ChangedField
				changesData := ChangedField{
					Name:     currentFieldName,
					OldValue: originalField.Interface(),
					NewValue: currentField.Interface(),
				}

				// Update mutable status
				if isCurrentMutable {
					setMutableStatus(currentValue, Changed)
					appendChangedField(currentValue, ChangedFields{changesData.Name: changesData})
				}
				// Set changed flag
				result = true
				// Append changed fields
				changedFields[changesData.Name] = changesData
			}
		}
	}

	return result, changedFields
}
