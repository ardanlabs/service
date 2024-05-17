// Package page provides support for query paging.
package page

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ardanlabs/service/foundation/validate"
)

// Page represents the requested page and rows per page.
type Page struct {
	number int
	rows   int
}

// Parse parses the strings and validates the values are in reason.
func Parse(page string, rowsPerPage string) (Page, error) {
	number := 1
	if page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, validate.NewFieldsError("page", err)
		}
	}

	rows := 10
	if rowsPerPage != "" {
		var err error
		rows, err = strconv.Atoi(rowsPerPage)
		if err != nil {
			return Page{}, validate.NewFieldsError("rows", err)
		}
	}

	if number <= 0 {
		return Page{}, validate.NewFieldsError("page", errors.New("page value too small, must be larger than 0"))
	}

	if rows <= 0 {
		return Page{}, validate.NewFieldsError("rows", errors.New("rows value too small, must be larger than 0"))
	}

	if rows > 100 {
		return Page{}, validate.NewFieldsError("rows", errors.New("rows value too large, must be less than 100"))
	}

	p := Page{
		number: number,
		rows:   rows,
	}

	return p, nil
}

// MustParse creates a paging value for testing.
func MustParse(page string, rowsPerPage string) Page {
	pg, err := Parse(page, rowsPerPage)
	if err != nil {
		panic(err)
	}

	return pg
}

// String implements the stringer interface.
func (p Page) String() string {
	return fmt.Sprintf("page: %d rows: %d", p.number, p.rows)
}

// Number returns the page number.
func (p Page) Number() int {
	return p.number
}

// RowsPerPage returns the rows per page.
func (p Page) RowsPerPage() int {
	return p.rows
}
