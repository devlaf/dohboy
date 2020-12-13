package doh

type set struct {
	store map[string]struct{}
}

func NewSet() *set {
	return &set{
		store: make(map[string]struct{}),
	}
}

func (s *set) Add(val string) {
	s.store[val] = struct{}{}
}

func (s *set) Remove(val string) {
	delete(s.store, val)
}

func (s *set) Contains(val string) bool {
	_, c := s.store[val]
	return c
}
