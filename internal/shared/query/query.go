package query

type Result[T any] struct {
	Data    []T
	Limit   int
	Offset  int
	HasMore bool
}
