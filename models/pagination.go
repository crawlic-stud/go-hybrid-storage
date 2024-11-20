package models

type PaginatedItems[T any] struct {
	Items      []T   `json:"items"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"pageSize"`
	IsNextPage bool  `json:"isNextPage"`
}
