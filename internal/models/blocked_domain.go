package models

import "time"

type BlockedDomain struct {
	ID uint `gorm:"primaryKey"`

	DNSLogID uint `gorm:"index"`

	Domain string `gorm:"index"`

	RiskScore int

	Reason string

	BlockedAt time.Time

	CreatedAt time.Time
}
