// Package query provides support for query paging.
package query

import (
	"encoding/json"

	"github.com/ardanlabs/service/business/sdk/page"
)

// Result is the data model used when returning a query result.
type Result[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rowsPerPage"`
}

// NewResult constructs a result value to return query results.
func NewResult[T any](items []T, total int, page page.Page) Result[T] {
	return Result[T]{
		Items:       items,
		Total:       total,
		Page:        page.Number(),
		RowsPerPage: page.RowsPerPage(),
	}
}

// Encode implements the encoder interface.
func (r Result[T]) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
