package model

// PaginationParams holds common pagination parameters.
type PaginationParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}

// PaginatedResult wraps a list response with pagination metadata.
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// DefaultPagination returns pagination params with sensible defaults.
func DefaultPagination() PaginationParams {
	return PaginationParams{Page: 1, PerPage: 20, Sort: "id", Order: "desc"}
}
