package mutable

// Mutabler is an interface that wraps Mutable features
type Mutabler interface {
	ResetMutableState(interface{}) error
	SetValue(string, interface{}) error
	AnalyzeChanges() ChangedFields
}