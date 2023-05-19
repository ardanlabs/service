package summary

import "github.com/google/uuid"

// Summary represents information about an individual user and their products.
type Summary struct {
	UserID     uuid.UUID
	UserName   string
	TotalCount int
	TotalCost  float64
}
