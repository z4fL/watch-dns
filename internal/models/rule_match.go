package models

import "time"

type RuleMatch struct {
	ID uint `gorm:"primaryKey"`

	DNSLogID uint `gorm:"index"`

	Rule  string
	Score int

	CreatedAt time.Time
	UpdatedAt time.Time
}
