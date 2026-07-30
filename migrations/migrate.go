package migrations

import (
	"github.com/z4fL/watch-dns/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DNSLog{},
		&models.BlockedDomain{},
		&models.Rule{},
		&models.RuleMatch{},
		&models.Setting{},
		&models.TelegramReport{},
		&models.AllowList{},
	)
}
