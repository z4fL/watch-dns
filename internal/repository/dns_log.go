package repository

import (
	"context"

	"github.com/z4fL/watch-dns/internal/models"
	"gorm.io/gorm"
)

type DNSLogRepository interface {
	Create(ctx context.Context, log *models.DNSLog) error
}

type dnsLogRepository struct {
	db *gorm.DB
}

func NewDNSLogRepository(db *gorm.DB) DNSLogRepository {
	return &dnsLogRepository{
		db: db,
	}
}

func (r *dnsLogRepository) Create(
	ctx context.Context,
	log *models.DNSLog,
) error {
	return r.db.WithContext(ctx).Create(log).Error
}
