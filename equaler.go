package mutable

// Equaler is the interface that wraps Equal function allows to check differences between two objects of the same type
type Equaler interface {
	Equal(interface{}) bool
}
