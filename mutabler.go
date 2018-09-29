package mutable

type Mutabler interface {
	ResetMutableState(interface{}) error
	SetValue(string, interface{}) error
	AnalyzeChanges() ChangedFields
}