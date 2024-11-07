package bexpr

type queue[T any] struct {
	items []T
}

func (q *queue[T]) push(item T) {
	q.items = append(q.items, item)
}

func (q *queue[T]) pop() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item, true
}
