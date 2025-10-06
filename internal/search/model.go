package search

import (
	"time"

	"github.com/lib/pq"
)

// Gin represents a gin entry persisted in the database.
type Gin struct {
	ID          uint           `json:"-" gorm:"column:id;primaryKey"`
	Name        string         `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Country     string         `json:"country" gorm:"column:country;type:varchar(255);not null"`
	Botanicals  pq.StringArray `json:"botanicals" gorm:"column:botanicals;type:text[]"`
	Description string         `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time      `json:"-" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"-" gorm:"column:updated_at"`
}

// TableName specifies the PostgreSQL table name for gins.
func (Gin) TableName() string {
	return "gin"
}
