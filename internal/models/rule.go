package models

import "time"

type RuleType string

const (
	RuleTypeKeyword RuleType = "keyword"
	RuleTypeTLD     RuleType = "tld"
	RuleTypeNumber  RuleType = "number"
)

type Rule struct {
	ID        uint     `gorm:"primaryKey"`
	Type      RuleType `gorm:"type:text;not null;index"`
	Pattern   string   `gorm:"type:text;not null"`
	Score     int      `gorm:"not null"`
	Enabled   bool     `gorm:"not null;default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
