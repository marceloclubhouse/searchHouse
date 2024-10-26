package spider

type StringSet struct {
	m map[string]bool
}

func (s *StringSet) Add(str string) {
	// Initialize set
	if s.m == nil {
		s.m = map[string]bool{}
	}
	s.m[str] = true
}

func (s *StringSet) Remove(str string) {
	delete(s.m, str)
}

func (s *StringSet) Contains(str string) bool {
	_, contains := s.m[str]
	return contains
}

func (s *StringSet) Merge(right StringSet) {
	for key, _ := range right.m {
		s.Add(key)
	}
}
