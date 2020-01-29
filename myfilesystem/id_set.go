package myfilesystem

// wrapper over a map, that works as a set

type IDSet struct {
	List map[ID]struct{}
}

func (s *IDSet) Has(v ID) bool {
	_, ok := s.List[v]
	return ok
}

func (s *IDSet) Add(v ID) {
	s.List[v] = struct{}{}
}

func (s *IDSet) Remove(v ID) {
	delete(s.List, v)
}

func (s *IDSet) Clear() {
	s.List = make(map[ID]struct{})
}

func NewIdSet() *IDSet {
	s := &IDSet{}
	s.List = make(map[ID]struct{})
	return s
}
