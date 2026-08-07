package ports

type RandomGenerator[T any] interface {
	IntN(n int) int
	PickOne(items []T) (T, error)
}
