package models

import "time"

type TelegramReport struct {
	ID uint `gorm:"primaryKey"`

	ReportDate time.Time `gorm:"index"`

	Message string

	SentAt time.Time

	CreatedAt time.Time
}
