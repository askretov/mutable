package main

import (
	"fmt"

	"github.com/askretov/mutable"
)

// NestedStruct is a simple nested struct type
type NestedStruct struct {
	mutable.Mutable
	FieldY string
	FieldZ []int64
}

// MyStruct is a main struct type
type MyStruct struct {
	mutable.Mutable
	FieldA string
	FieldB int64        `mutable:"ignored"`
	FieldC NestedStruct `mutable:"deep"`
}

func main() {
	var m = &MyStruct{}
	// Mutable state init
	m.ResetMutableState(m)

	// Change values
	m.FieldA = "green"
	m.FieldC.FieldY = "stone"
	// Analyze changes
	fmt.Println(m.AnalyzeChanges())

	// Reset mutable state
	m.ResetMutableState(m)
	// Set values dynamically
	m.SetValue("FieldA", "white")
	m.SetValue("FieldC/FieldZ", "[1,2,3]") // You can set typed value or JSON string as well
	// Analyze changes
	fmt.Println(m.AnalyzeChanges())
}
