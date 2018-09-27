package mutable

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Check unexported fields in Test structs

type TestA struct {
	Mutable
	FieldA string  `json:"field_a"`
	FieldB float64 `json:"field_b"`
	FieldC byte    `json:"field_c"`
	FieldD TestB   `json:"field_d" mutable:"deep"`
	FieldE *TestB  `json:"field_e"`
	FieldF TestC   `json:"field_f" mutable:"deep"`
}

type TestB struct {
	FieldA string `json:"field_a"`
	FieldB []int  `json:"field_b"`
}

type TestC struct {
	Mutable
	FieldA string `json:"field_a"`
	FieldB []int  `json:"field_b"`
}

func TestMutable_ResetMutableState(t *testing.T) {
	// Create an object of TestA type
	var tst = TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldC: 3,
		FieldD: TestB{},
		FieldE: &TestB{},
		FieldF: TestC{},
	}
	// Check an arg type checking
	assert.NoError(t, tst.ResetMutableState(&tst), "pointer")
	assert.Error(t, tst.ResetMutableState(tst), "non-pointer")

	// Check a target
	_, ok := tst.target.(*TestA)
	assert.True(t, ok, "target type")
	assert.Equal(t, &tst, tst.target.(*TestA), "target pointer value")
	// Check an original state
	_, ok = tst.originalState.(TestA)
	assert.True(t, ok, "original state object type")
	// Check a mutable status
	assert.Equal(t, NotChanged, tst.MutableStatus, "status")
	// Check changed fields
	assert.Equal(t, 0, len(tst.ChangedFields), "changed fields len")

	// Check a target of a nested mutable object
	target, ok := tst.FieldF.target.(*TestC)
	assert.True(t, ok, "nested mutable target type")
	assert.Equal(t, &tst.FieldF, target, "nested mutable target pointer value")
	// Check an original state
	_, ok = tst.FieldF.originalState.(TestC)
	assert.True(t, ok, "nested mutable original state object type")
}

func TestMutable_SetValue(t *testing.T) {
	// Create an object of TestA type
	var obj = &TestA{
		FieldA: "one",
		FieldB: 2.0,
		FieldD: TestB{FieldA: "green"},
	}
	assert.NoError(t, obj.ResetMutableState(obj), "init")
	// Try to set a value
	err := obj.SetValue("field_a", "two")
	assert.NoError(t, err, "setValue error")
	assert.Equal(t, "two", obj.FieldA, "updated value")
	// Try to set a value for a nested struct
	err = obj.SetValue("field_d/field_a", "white")
	assert.NoError(t, err, "nested setValue error")
	assert.Equal(t, "white", obj.FieldD.FieldA, "nested updated value")
	// Try to set a value for not existing field
	err = obj.SetValue("wrong_field", "two")
	if assert.Error(t, err, "setValue error") {
		assert.True(t, IsCannotFindErr(err), "setValue error type")
	}
}

func TestMutable_setMutableStatus(t *testing.T) {
	// Create an object of TestA type
	var obj = &TestA{}
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
	// Prepare test cases
	testCases := []struct {
		name     string
		scenario func() ChangedFields
		expected ChangedFields
	}{
		{
			name: "regular",
			scenario: func() ChangedFields {
				tst := &struct {
					Mutable
					FieldA string
					FieldB int
				}{
					FieldA: "one",
					FieldB: 1,
				}
				tst.ResetMutableState(tst)
				tst.FieldA = "two"
				return tst.AnalyzeChanges()
			},
			expected: ChangedFields{
				"FieldA": ChangedField{
					Name:     "FieldA",
					OldValue: "one",
					NewValue: "two",
				},
			},
		}, {
			name: "regularNonMutStruct",
			scenario: func() ChangedFields {
				tst := &struct {
					Mutable
					FieldA TestB
				}{
					FieldA: TestB{FieldA: "apple"},
				}
				tst.ResetMutableState(tst)
				tst.FieldA.FieldA = "banana"
				return tst.AnalyzeChanges()
			},
			expected: ChangedFields{
				"FieldA": ChangedField{
					Name:     "FieldA",
					OldValue: TestB{FieldA: "apple"},
					NewValue: TestB{FieldA: "banana"},
				},
			},
		}, {
			name: "regularInvalidNonMutStructPtr",
			scenario: func() ChangedFields {
				tst := &struct {
					Mutable
					FieldA *TestB
					FieldB *TestB
					FieldC *TestB
				}{
					FieldA: &TestB{FieldA: "apple"},
					FieldC: &TestB{FieldA: "tree"},
				}
				tst.ResetMutableState(tst)
				// FieldA changes shouldn't be tracked because we use it as pointer
				tst.FieldA.FieldA = "banana"
				tst.FieldB = &TestB{FieldA: "car"}
				tst.FieldC = nil
				return tst.AnalyzeChanges()
			},
			expected: ChangedFields{
				"FieldB": ChangedField{
					Name:     "FieldB",
					OldValue: nil,
					NewValue: TestB{FieldA: "car"},
				},
				"FieldC": ChangedField{
					Name:     "FieldC",
					OldValue: TestB{FieldA: "tree"},
					NewValue: nil,
				},
			},
		}, {
			name: "regularDeepMutStruct",
			scenario: func() ChangedFields {
				tst := &struct {
					Mutable
					FieldA TestC
					FieldB TestC `mutable:"deep"`
					FieldC *TestC
					FieldD *TestC `mutable:"deep"`
				}{
					FieldA: TestC{FieldA: "apple"},
					FieldB: TestC{FieldA: "tree"},
					FieldC: &TestC{FieldA: "red"},
					FieldD: &TestC{FieldA: "dog"},
				}
				tst.ResetMutableState(tst)
				tst.FieldA.FieldA = "banana"
				tst.FieldB.FieldA = "stone"
				// FieldC changes shouldn't be tracked because we use it as pointer and have no deep tag
				tst.FieldC.FieldA = "green"
				tst.FieldD.FieldA = "cat"
				return tst.AnalyzeChanges()
			},
			expected: ChangedFields{
				"FieldA": ChangedField{
					Name:     "FieldA",
					OldValue: TestC{FieldA: "apple"},
					NewValue: TestC{FieldA: "banana"},
				},
				"FieldB": ChangedField{
					Name: "FieldB",
					NestedFields: ChangedFields{
						"FieldA": ChangedField{
							Name:     "FieldA",
							OldValue: "tree",
							NewValue: "stone",
						},
					},
				},
				"FieldD": ChangedField{
					Name: "FieldD",
					NestedFields: ChangedFields{
						"FieldA": ChangedField{
							Name:     "FieldA",
							OldValue: "dog",
							NewValue: "cat",
						},
					},
				},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		result := tc.scenario()
		assert.Equal(t, tc.expected.String(), result.String(), "Expected: %s\nActual: %s\ntestCase: %s", tc.expected, result, tc.name)
	}

}
