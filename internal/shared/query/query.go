package query

type Result[T any] struct {
	Data    []T
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}
