package sqlite3

// https://www.davidkaya.com/sets-in-golang/
var exists = struct{}{}

// Set represents an unordered collection of strings.
type Set interface {
	Add(value string)
	Remove(value string)
	Contains(value string) bool
}

type set struct {
	m map[string]struct{}
}

// NewSet creates a collection of strings.
func NewSet(contents ...string) Set {
	s := &set{}
	s.m = make(map[string]struct{})
	for _, i := range contents {
		s.Add(i)
	}
	return s
}

// Add a string to the collection.
func (s *set) Add(value string) {
	s.m[value] = exists
}

// Remove a string from the collection.
func (s *set) Remove(value string) {
	delete(s.m, value)
}

// Contains is a predicate indicating the string's membership in the set.
func (s *set) Contains(value string) bool {
	_, c := s.m[value]
	return c
}
