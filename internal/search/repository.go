package search

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// Repository defines access methods to gin data storage.
type Repository interface {
	Search(ctx context.Context, filter SearchFilter) ([]Gin, error)
}

// SearchFilter represents filtering and pagination options supported by the repository.
type SearchFilter struct {
	Query  string
	Limit  int
	Offset int
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository constructs a Repository backed by GORM.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Search(ctx context.Context, filter SearchFilter) ([]Gin, error) {
	var gins []Gin

	tx := r.db.WithContext(ctx).Model(&Gin{})

	trimmed := strings.TrimSpace(filter.Query)
	if trimmed != "" {
		needle := strings.ToLower(trimmed)
		like := "%" + needle + "%"
		tx = tx.Where(
			`(LOWER(name) LIKE @like OR LOWER(country) LIKE @like OR LOWER(description) LIKE @like
            OR EXISTS (SELECT 1 FROM unnest(botanicals) AS botanical WHERE LOWER(botanical) LIKE @like))`,
			map[string]interface{}{"like": like},
		)
	}

	if filter.Limit > 0 {
		tx = tx.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		tx = tx.Offset(filter.Offset)
	}

	if err := tx.Order("name ASC").Find(&gins).Error; err != nil {
		return nil, err
	}

	return gins, nil
}
