package sqlite3

// https://www.davidkaya.com/sets-in-golang/
var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

func NewSet(contents ...string) *set {
	s := &set{}
	s.m = make(map[string]struct{})
	for _, i := range contents {
		s.Add(i)
	}
	return s
}

func (s *set) Add(value string) {
	s.m[value] = exists
}

func (s *set) Remove(value string) {
	delete(s.m, value)
}

func (s *set) Contains(value string) bool {
	_, c := s.m[value]
	return c
}
