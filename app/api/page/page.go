// Package page provides support for query paging.
package page

import (
	"strconv"

	"github.com/ardanlabs/service/foundation/validate"
)

// Page represents the requested page and rows per page.
type Page struct {
	Number      int
	RowsPerPage int
}

// Parse parses the request for the page and rows query string. The
// defaults are provided as well.
func Parse(page string, rows string) (Page, error) {
	number := 1
	if page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, validate.NewFieldsError("page", err)
		}
	}

	rowsPerPage := 10
	if rows != "" {
		var err error
		rowsPerPage, err = strconv.Atoi(rows)
		if err != nil {
			return Page{}, validate.NewFieldsError("rows", err)
		}
	}

	p := Page{
		Number:      number,
		RowsPerPage: rowsPerPage,
	}

	return p, nil
}

// Document is the form used for API responses from query API calls.
type Document[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rowsPerPage"`
}

// NewDocument constructs a response value for a web paging trusted.
func NewDocument[T any](items []T, total int, page int, rowsPerPage int) Document[T] {
	return Document[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		RowsPerPage: rowsPerPage,
	}
}
