package dialect_test

import (
	"bytes"
	"testing"

	"github.com/ardanlabs/service/business/sdk/sqldb/dialect"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name    string
		dialect dialect.Dialect
		want    string
	}{
		{
			name:    "postgres",
			dialect: dialect.Postgres{},
			want:    " OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY",
		},
		{
			name:    "sqlite",
			dialect: dialect.SQLite{},
			want:    " LIMIT :rows_per_page OFFSET :offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.dialect.Paginate(&buf)

			if got := buf.String(); got != tt.want {
				t.Errorf("Paginate: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		dialect dialect.Dialect
		want    string
	}{
		{dialect.Postgres{}, "postgres"},
		{dialect.SQLite{}, "sqlite"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dialect.Name(); got != tt.want {
				t.Errorf("Name: got %q, want %q", got, tt.want)
			}
		})
	}
}
