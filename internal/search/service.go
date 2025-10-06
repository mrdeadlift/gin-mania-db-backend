package search

import (
	"context"
	"errors"
)

var (
	// ErrRepositoryNotConfigured indicates that the service was constructed without a backing repository.
	ErrRepositoryNotConfigured = errors.New("search repository not configured")
	// ErrInvalidPagination is returned when the requested pagination parameters are negative.
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// Service provides search capabilities backed by a repository implementation.
type Service struct {
	repo Repository
}

// NewService constructs a new Service using the provided repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Search retrieves gins that satisfy the provided filter parameters.
func (s *Service) Search(ctx context.Context, filter SearchFilter) ([]Gin, error) {
	if s.repo == nil {
		return nil, ErrRepositoryNotConfigured
	}

	if filter.Limit < 0 || filter.Offset < 0 {
		return nil, ErrInvalidPagination
	}

	return s.repo.Search(ctx, filter)
}

// SearchByQuery is a convenience wrapper for simple query-driven searches without pagination.
func (s *Service) SearchByQuery(ctx context.Context, query string) ([]Gin, error) {
	return s.Search(ctx, SearchFilter{Query: query})
}
