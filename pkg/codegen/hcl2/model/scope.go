package model

type scope map[string]Node

func (s scope) bindReference(name string) (Node, bool) {
	def, ok := s[name]
	return def, ok
}

func (s scope) define(name string, node Node) bool {
	if _, exists := s[name]; exists {
		return false
	}
	s[name] = node
	return true
}

type scopes struct {
	stack []scope
}

func (s *scopes) push() scope {
	next := scope{}
	s.stack = append(s.stack, next)
	return next
}

func (s *scopes) pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *scopes) bindReference(name string) (Node, bool) {
	for _, s := range s.stack {
		def, ok := s.bindReference(name)
		if ok {
			return def, true
		}
	}
	return nil, false
}
