package mutable

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestA struct {
	Mutable
	FieldA string            `json:"field_a"`
	FieldB float64           `json:"field_b"`
	FieldC byte              `json:"field_c"`
	FieldD TestB             `json:"field_d" mutable:"deep"`
	FieldE *TestB            `json:"field_e"`
	FieldF TestC             `json:"field_f" mutable:"deep"`
	FieldG []*TestC          `json:"field_g"`
	FieldH map[string]*TestC `json:"field_h"`
	FieldI []TestC           `json:"field_i"`
	FieldJ map[string]TestC  `json:"field_j"`
	FieldK []int             `json:"field_k"`
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
		FieldG: []*TestC{
			&TestC{},
			&TestC{},
		},
		FieldH: map[string]*TestC{
			"one": &TestC{},
			"two": &TestC{},
		},
		FieldI: []TestC{
			TestC{},
			TestC{},
		},
		FieldJ: map[string]TestC{},
		FieldK: []int{0, 1},
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

	// Check a target of a nested slice of mutable objects
	for i := range tst.FieldG {
		target, ok := tst.FieldG[i].target.(*TestC)
		assert.True(t, ok, "nested mutable target type")
		assert.Equal(t, tst.FieldG[i], target, "nested mutable target pointer value")
		// Check an original state
		_, ok = tst.FieldG[i].originalState.(TestC)
		assert.True(t, ok, "nested mutable original state object type")
	}

	// Check a target of a nested map of mutable objects
	for key := range tst.FieldH {
		target, ok := tst.FieldH[key].target.(*TestC)
		assert.True(t, ok, "nested mutable target type")
		assert.Equal(t, tst.FieldH[key], target, "nested mutable target pointer value")
		// Check an original state
		_, ok = tst.FieldH[key].originalState.(TestC)
		assert.True(t, ok, "nested mutable original state object type")
	}

	// Check a target of a nested slice of mutable objects
	for i := range tst.FieldI {
		target, ok := tst.FieldI[i].target.(*TestC)
		assert.True(t, ok, "nested mutable target type")
		assert.Equal(t, &tst.FieldI[i], target, "nested mutable target pointer value")
		// Check an original state
		_, ok = tst.FieldI[i].originalState.(TestC)
		assert.True(t, ok, "nested mutable original state object type")
	}
}

func TestMutable_SetValue(t *testing.T) {
	// Create an object of TestA type
	var obj = &struct {
		Mutable
		FieldA string  `json:"field_a"`
		FieldB float64 `json:"field_b"`
		FieldC []int64 `json:"field_c"`
		FieldD TestB   `json:"field_d"`
		FieldE string
	}{
		FieldA: "one",
		FieldB: 2.0,
		FieldD: TestB{FieldA: "green"},
	}
	assert.NoError(t, obj.ResetMutableState(obj), "init")

	// Try to set a value
	err := obj.SetValue("field_a", "two")
	assert.NoError(t, err)
	assert.Equal(t, "two", obj.FieldA)

	// Try to set a string type value to different type field
	err = obj.SetValue("field_b", "2.5")
	assert.NoError(t, err)
	assert.Equal(t, float64(2.5), obj.FieldB)

	// Try to set a slice value
	err = obj.SetValue("field_c", []int64{1, 2, 3})
	assert.NoError(t, err)
	assert.Subset(t, []int64{1, 2, 3}, obj.FieldC)

	// Try to set a slice json value
	err = obj.SetValue("field_c", "[1,2,3]")
	assert.NoError(t, err)
	assert.Subset(t, []int64{1, 2, 3}, obj.FieldC)

	// Try to set a value for a nested struct
	err = obj.SetValue("field_d/field_a", "white")
	assert.NoError(t, err)
	assert.Equal(t, "white", obj.FieldD.FieldA)

	// Try to set a value by field's real name (has no json tag)
	err = obj.SetValue("FieldE", "noJSON")
	assert.NoError(t, err)
	assert.Equal(t, "noJSON", obj.FieldE)

	// Try to set a value for not existing field
	err = obj.SetValue("wrong_field", "two")
	if assert.Error(t, err) {
		assert.True(t, IsCannotFindErr(err))
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
				"FieldA": &ChangedField{
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
				"FieldA": &ChangedField{
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
				"FieldB": &ChangedField{
					Name:     "FieldB",
					OldValue: nil,
					NewValue: TestB{FieldA: "car"},
				},
				"FieldC": &ChangedField{
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
				"FieldA": &ChangedField{
					Name:     "FieldA",
					OldValue: TestC{FieldA: "apple"},
					NewValue: TestC{FieldA: "banana"},
				},
				"FieldB": &ChangedField{
					Name: "FieldB",
					NestedFields: ChangedFields{
						"FieldA": &ChangedField{
							Name:     "FieldA",
							OldValue: "tree",
							NewValue: "stone",
						},
					},
				},
				"FieldD": &ChangedField{
					Name: "FieldD",
					NestedFields: ChangedFields{
						"FieldA": &ChangedField{
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

func TestMutable_ignoredFields(t *testing.T) {
	tst := &struct {
		Mutable
		FieldA string `json:"field_a"`
		FieldB string `json:"field_b" mutable:"ignored"`
	}{
		FieldA: "apple",
		FieldB: "stone",
	}
	tst.ResetMutableState(tst)
	tst.SetValue("field_a", "banana")
	tst.SetValue("field_b", "wood")
	changes := tst.AnalyzeChanges()
	expectedChanges := ChangedFields{
		"FieldA": &ChangedField{
			Name:     "FieldA",
			OldValue: "apple",
			NewValue: "banana",
		},
	}
	assert.Equal(t, expectedChanges.String(), changes.String(), "Expected: %s\nActual: %s\ntestCase: %s", expectedChanges.String(), changes.String())
}

func TestMutable_unexportedFields(t *testing.T) {
	tst := &struct {
		Mutable
		FieldA string `json:"field_a"`
		fieldB string
	}{
		FieldA: "apple",
		fieldB: "stone",
	}
	tst.ResetMutableState(tst)
	assert.NoError(t, tst.SetValue("field_a", "banana"))
	assert.Error(t, tst.SetValue("fieldB", "wood"))
	changes := tst.AnalyzeChanges()
	expectedChanges := ChangedFields{
		"FieldA": &ChangedField{
			Name:     "FieldA",
			OldValue: "apple",
			NewValue: "banana",
		},
	}
	assert.Equal(t, expectedChanges.String(), changes.String(), "Expected: %s\nActual: %s\ntestCase: %s", expectedChanges.String(), changes.String())
}
