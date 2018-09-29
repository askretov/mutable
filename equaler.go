package mutable

// Equaler is the interface that wraps custom Equal function allowing to check differences
// between two objects of the same type
type Equaler interface {
	Equal(interface{}) bool
}
