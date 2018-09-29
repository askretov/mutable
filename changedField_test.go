package mutable

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"encoding/json"
)

func TestChangedFields_Contains(t *testing.T) {
	cf := ChangedFields{
		"FieldA": &ChangedField{
			Name: "FieldA",
		},
		"FieldB": &ChangedField{
			Name: "FieldB",
		},
	}
	assert.True(t, cf.Contains("FieldA"))
	assert.False(t, cf.Contains("FieldC"))
}

func TestChangedFields_Keys(t *testing.T) {
	cf := ChangedFields{
		"FieldA": &ChangedField{
			Name: "FieldA",
		},
		"FieldB": &ChangedField{
			Name: "FieldB",
		},
	}
	assert.Subset(t, []string{"FieldA", "FieldB"}, cf.Keys())
}

func TestChangedFields_GetField(t *testing.T) {
	cf := ChangedFields{
		"FieldA": &ChangedField{
			Name: "FieldA",
			OldValue: "one",
			NewValue: "two",
		},
		"FieldB": &ChangedField{
			Name: "FieldB",
		},
	}
	assert.Equal(t, &ChangedField{
		Name: "FieldA",
		OldValue: "one",
		NewValue: "two",
	}, cf.GetField("FieldA"))
	assert.Equal(t, &ChangedField{
		Name: "FieldB",
	}, cf.GetField("FieldB"))
}

func TestChangedFields_JSON(t *testing.T) {
	cf := ChangedFields{
		"FieldA": &ChangedField{
			Name: "FieldA",
			OldValue: "one",
			NewValue: "two",
		},
		"FieldB": &ChangedField{
			Name: "FieldB",
		},
		"FieldC": &ChangedField{
			Name: "FieldC",
			NestedFields: ChangedFields{
				"FieldA": &ChangedField{
					Name: "FieldA",
				},
				"FieldB": &ChangedField{
					Name: "FieldB",
				},
			},
		},
	}
	assert.True(t, json.Valid(cf.JSON(true)))
	assert.True(t, json.Valid(cf.JSON(false)))
}