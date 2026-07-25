package models

import "time"

type AllowList struct {
	ID uint `gorm:"primaryKey"`

	Domain string `gorm:"uniqueIndex"`

	Reason string

	CreatedAt time.Time
	UpdatedAt time.Time
}
