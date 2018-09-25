package mutable

import (
	"reflect"
	"testing"

	"github.com/go-ext/logger"
	"github.com/stretchr/testify/assert"
)

type TestA struct {
	Mutable
	FieldA string  `json:"field_a"`
	FieldB float64 `json:"field_b"`
	FieldC byte    `json:"field_c"`
	FieldD TestB   `json:"field_d"`
	FieldE *TestB  `json:"field_e" mutable:"deep"`
	FieldF *TestC  `json:"field_f"`
}

type TestB struct {
	FieldA string
	FieldB []int
}

type TestC struct {
	Mutable
	FieldA string
	FieldB []int
}

func TestMutable_ResetMutableState(t *testing.T) {
	// Create an object of TestA type
	var obj = TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldC: 3,
	}
	// Check an arg type checking
	assert.NoError(t, obj.ResetMutableState(&obj), "pointer")
	assert.Error(t, obj.ResetMutableState(obj), "non-pointer")
	// Change a field and reset
	obj.FieldA = "two"
	obj.ResetMutableState(&obj)

	// Check a target
	_, ok := obj.target.(*TestA)
	assert.True(t, ok, "target type")
	assert.Equal(t, &obj, obj.target.(*TestA), "target pointer value")
	// Check an original state
	_, ok = obj.originalState.(TestA)
	assert.True(t, ok, "original state object type")
	// Check a mutable status
	assert.Equal(t, NotChanged, obj.MutableStatus, "status")
	// Check changed fields
	assert.Equal(t, 0, len(obj.ChangedFields), "changed fields len")
}

func TestMutable_SetValue(t *testing.T) {
	// Create an object of TestA type
	var obj = &TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldC: 3,
	}
	assert.NoError(t, obj.ResetMutableState(obj), "init")
	// Try to set value
	err, ok := obj.SetValue("field_a", "two")
	assert.NoError(t, err, "setValue error")
	assert.True(t, ok, "setValue result")
	assert.Equal(t, "two", obj.FieldA, "updated value")
	// Try to set value
	err, ok = obj.SetValue("wrong_field", "two")
	assert.Error(t, err, "setValue error")
	assert.False(t, ok, "setValue result")

	//TODO: Add nested object's value setting check
}

func TestMutable_setMutableStatus(t *testing.T) {
	// Create an object of TestA type
	var obj = &TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldC: 3,
	}
	assert.NoError(t, obj.ResetMutableState(obj), "init")
	setMutableStatus(reflect.ValueOf(obj).Elem(), Added)
	assert.Equal(t, Added, obj.MutableStatus, "status")
}

func TestMutable_isMutable(t *testing.T) {
	// Create an object of TestA type
	var mut = &TestA{}
	var nonMut = &struct {
		A string
		B string
	}{}
	assert.True(t, isMutable(reflect.ValueOf(mut).Elem()), "mutable")
	assert.False(t, isMutable(reflect.ValueOf(nonMut).Elem()), "non-mutable")
}

func TestMutable_AnalyzeChanges(t *testing.T) {
	// Create an object of TestA type
	var obj = &TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldC: 3,
		FieldD: TestB{
			FieldA: "apple",
			FieldB: []int{1, 2},
		},
		FieldE: &TestB{
			FieldA: "white",
		},
	}
	assert.NoError(t, obj.ResetMutableState(obj), "init")
	// Change the fields
	obj.FieldA = "two"
	obj.FieldC = 0x10
	obj.FieldD.FieldA = "banana"
	obj.FieldE.FieldA = "green"
	// Change the fields with setValue method
	obj.SetValue("field_b", "2.5")

	// Analyze changes
	changes := obj.AnalyzeChanges()
	logger.Debug(changes)


	// Check the changes
	assert.True(t, changes.Contains(&ChangedField{
		Name:     "FieldA",
		OldValue: "one",
		NewValue: "two",
	}), "FieldA")
	assert.True(t, changes.Contains(&ChangedField{
		Name:     "FieldC",
		OldValue: 0x03,
		NewValue: 0x010,
	}), "FieldC")
	assert.True(t, changes.Contains(&ChangedField{
		Name: "FieldD",
		OldValue: TestB{
			FieldA: "apple",
			FieldB: []int{1, 2},
		},
		NewValue: TestB{
			FieldA: "banana",
			FieldB: []int{1, 2},
		},
	}), "FieldD")
	assert.True(t, changes.Contains(&ChangedField{
		Name:     "FieldB",
		OldValue: 2.0,
		NewValue: 2.5,
	}), "FieldB")
}
