package spider

type StringSet struct {
	m map[string]bool
}

func (s *StringSet) Add(str string) {
	if s.m == nil {
		s.m = make(map[string]bool)
	}
	s.m[str] = true
}

func (s *StringSet) Remove(str string) {
	delete(s.m, str)
}

func (s *StringSet) Contains(str string) bool {
	return s.m[str]
}

func (s *StringSet) Merge(right StringSet) {
	for key := range right.m {
		s.Add(key)
	}
}
