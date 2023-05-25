package stack

// Stack is an implementation of an abstract Stack
type Stack[T any] struct {
	arr []T
}

func New[T any](init ...T) Stack[T] {
	return Stack[T]{arr: init}
}

func (s *Stack[T]) Len() int {
	return len(s.arr)
}

func (s *Stack[T]) Push(elem ...T) {
	s.arr = append(s.arr, elem...)
}

func (s *Stack[T]) PushReverse(elem []T) {
	for i := len(elem) - 1; i >= 0; i-- {
		s.Push(elem[i])
	}
}

func (s *Stack[T]) Pop() T {
	lastIndex := len(s.arr) - 1
	elem := s.arr[lastIndex]
	s.arr = s.arr[:lastIndex]
	return elem
}
