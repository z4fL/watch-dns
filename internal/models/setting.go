package models

import "time"

type Setting struct {
	ID uint `gorm:"primaryKey"`

	Key string `gorm:"uniqueIndex"`

	Value string

	CreatedAt time.Time
	UpdatedAt time.Time
}
