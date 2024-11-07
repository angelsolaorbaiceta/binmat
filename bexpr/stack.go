package bexpr

type stack[T any] struct {
	items []T
}

func (s *stack[T]) size() int {
	return len(s.items)
}

func (s *stack[T]) push(item T) {
	s.items = append(s.items, item)
}

func (s *stack[T]) pop() (T, bool) {
	if s.size() == 0 {
		var zero T
		return zero, false
	}

	item := s.items[len(s.items)-1]
	s.items = s.items[0 : len(s.items)-1]

	return item, true
}

func (s *stack[T]) peek(count int) ([]T, bool) {
	if s.size() < count {
		return nil, false
	}

	var (
		start = s.size() - count
		end   = s.size()
	)

	return s.items[start:end], true
}
