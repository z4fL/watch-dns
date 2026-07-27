package models

import "time"

type DNSLog struct {
	ID uint `gorm:"primaryKey"`

	Timestamp time.Time `gorm:"index;not null"`

	Domain string `gorm:"index;not null"`
	Root   string `gorm:"index"`

	Encrypted bool

	Protocol string

	ClientIP string
	Client   string

	DeviceID    string `gorm:"index"`
	DeviceName  string
	DeviceModel string

	Status string `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
