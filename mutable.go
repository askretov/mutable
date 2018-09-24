package mutable

import (
	"fmt"
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
	// Mutable status of an object
	MutableStatus Status `json:"-"`
	// Changed fields data
	ChangedFields ChangedFields `json:"-"`
}

// ResetMutableState resets current mutable status of an object and updates previous state with given currentState
func (m *Mutable) ResetMutableState(currentState interface{}) {
	// Update mutable status
	m.MutableStatus = NotChanged
	// Reset previous state object
	m.originalState = nil
	m.originalState = currentState
	// Reset changed fields arrays
	m.ChangedFields = ChangedFields{}
}

func TrySetValue(fieldJsonName, value string, object reflect.Value) (error, bool) {
	// Try to find an appropriate field and set a new value
	if err, success := trySetValueForObject(object, "", fieldJsonName, value); err != nil {
		logger.Warning(err.Error())
		return errCannotSetValue(fieldJsonName, value), false
	} else if !success {
		// If we have neither errors nor success state, it means that we didn't find a field
		return errCannotFind(fieldJsonName), false
	}

	return nil, true
}

func trySetValueForObject(valueVar reflect.Value, currentContainerName, fieldJsonName, value string) (error, bool) {
	// Check whether current object is mutable
	var isMutable = checkIfMutable(valueVar)
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
			var nestedStructIsMutable = checkIfMutable(valueVar.Field(z))
			// Try to set a value
			if err, success := trySetValueForObject(valueVar.Field(z), currentFieldJsonName+"_", fieldJsonName, value); err != nil {
				logger.Warning(err)
				return errCannotSetValue(fieldJsonName, value), false
			} else if success {
				// Update mutable status
				if !nestedStructIsMutable && isMutable {
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
				if isMutable {
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

func checkIfMutable(valueVar reflect.Value) bool {
	// Iterate over struct fields
	for z := 0; z < valueVar.NumField(); z++ {
		switch valueVar.Field(z).Interface().(type) {
		case Mutable:
			return true
		}
	}

	return false
}

func setMutableStatus(valueVar reflect.Value, status Status) {
	// Iterate over struct fields
	for z := 0; z < valueVar.NumField(); z++ {
		switch valueVar.Field(z).Interface().(type) {
		case Mutable:
			valueVar.Field(z).FieldByName("Status").Set(reflect.ValueOf(status))
			return
		}
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

func CheckChanges(currentValueVar, previousValueVar reflect.Value, currentContainerName string) (result bool, changedFields ChangedFields) {
	changedFields = ChangedFields{}
	// Check whether current object is mutable
	var isMutable = checkIfMutable(currentValueVar)
	// Iterate over struct fields
	for z := 0; z < currentValueVar.NumField(); z++ {
		var _, dontCheck = currentValueVar.Type().Field(z).Tag.Lookup("dontCheck")
		if strings.Contains(currentValueVar.Type().Field(z).Type.String(), "Mutable") || dontCheck {
			// Pass Mutable itself and fields marked as "don't check"
			continue
		}
		// Get current field name
		currentFieldName := currentContainerName + currentValueVar.Type().Field(z).Name

		// Check whether a field has a container tag
		containerTag, _ := currentValueVar.Type().Field(z).Tag.Lookup("container")
		isContainer, _ := strconv.ParseBool(containerTag)

		if isContainer {
			// Go next level
			// Check whether nested struct is mutable
			var nestedStructIsMutable = checkIfMutable(currentValueVar.Field(z))
			// Try to check a value
			if changed, fields := CheckChanges(currentValueVar.Field(z), previousValueVar.Field(z), currentFieldName+"_"); changed {
				// Update mutable status
				if !nestedStructIsMutable && isMutable {
					setMutableStatus(currentValueVar, Changed)
					appendChangedField(currentValueVar, fields)
				}
				// Set changed flag
				result = true
			}
		} else if currentValueVar.Field(z).CanInterface() {
			// Check current field
			var equals bool
			if _, ok := currentValueVar.Field(z).Interface().(Equaler); ok {
				// Compare with type's Equal method
				equals = currentValueVar.Field(z).Interface().(Equaler).Equal(previousValueVar.Field(z).Interface())
			} else {
				// Compare with reflect's DeepEqual
				equals = reflect.DeepEqual(currentValueVar.Field(z).Interface(), previousValueVar.Field(z).Interface())
			}
			if !equals {
				// Prepare ChangedField
				changedField := ChangedField{
					Name:     currentFieldName,
					OldValue: previousValueVar.Field(z).Interface(),
					NewValue: currentValueVar.Field(z).Interface(),
				}

				// Update mutable status
				if isMutable {
					setMutableStatus(currentValueVar, Changed)
					appendChangedField(currentValueVar, ChangedFields{changedField.Name: changedField})
				}
				// Set changed flag
				result = true
				// Append changed fields
				changedFields[changedField.Name] = changedField
			}
		}
	}

	return result, changedFields
}

func CheckDifferences(firstValueVar, secondValueVar reflect.Value, changedFields []string, currentContainerName string) (bool, []string) {
	var result bool
	// Iterate over struct fields
	for z := 0; z < firstValueVar.NumField(); z++ {
		var _, dontCheck = firstValueVar.Type().Field(z).Tag.Lookup("dontCheck")
		if strings.Contains(firstValueVar.Type().Field(z).Type.String(), "Mutable") || dontCheck {
			// Pass Mutable itself and fields marked as "don't check"
			continue
		}
		// Get current field name
		currentFieldName := currentContainerName + firstValueVar.Type().Field(z).Name

		// Check whether a field has a container tag
		containerTag, _ := firstValueVar.Type().Field(z).Tag.Lookup("container")
		isContainer, _ := strconv.ParseBool(containerTag)

		if isContainer {
			// Go next level
			// Try to check a value
			if changed, _ := CheckChanges(firstValueVar.Field(z), secondValueVar.Field(z), currentFieldName+"_"); changed {
				// Set changed flag
				result = true
			}
		} else if firstValueVar.Field(z).CanInterface() {
			// Check current field
			equals := reflect.DeepEqual(firstValueVar.Field(z).Interface(), secondValueVar.Field(z).Interface())
			if !equals {
				// Set changed flag
				result = true
				// Append changed fields names
				changedFieldData := fmt.Sprintf("Field: %s, Object 1: %s, Object 2: %s", currentFieldName, firstValueVar.Field(z).Interface(), secondValueVar.Field(z).Interface())
				changedFields = append(changedFields, changedFieldData)
			}
		}
	}

	return result, changedFields
}
