package service

import (
	"context"
	"fmt"

	"github.com/z4fL/watch-dns/internal/models"
	"github.com/z4fL/watch-dns/internal/nextdns"
	"github.com/z4fL/watch-dns/internal/repository"
)

type DNSLogService struct {
	repo repository.DNSLogRepository
}

func NewDNSLogService(repo repository.DNSLogRepository) *DNSLogService {
	return &DNSLogService{
		repo: repo,
	}
}

func (s *DNSLogService) Store(
	ctx context.Context,
	log nextdns.Log,
) error {
	dnsLog := toDNSLog(log)

	if err := s.repo.Create(ctx, &dnsLog); err != nil {
		return fmt.Errorf("store DNS log: %w", err)
	}

	return nil
}

func toDNSLog(log nextdns.Log) models.DNSLog {
	reasons := make([]models.DNSLogReason, 0, len(log.Reasons))
	for _, reason := range log.Reasons {
		reasons = append(reasons, models.DNSLogReason{
			ReasonID:   reason.ID,
			ReasonName: reason.Name,
		})
	}

	dnsLog := models.DNSLog{
		Timestamp: log.Timestamp,

		Domain: log.Domain,
		Root:   log.Root,

		Encrypted: log.Encrypted,

		Protocol: log.Protocol,

		ClientIP: log.ClientIP,
		Client:   log.Client,

		Status: log.Status,

		Reasons: reasons,
	}

	if log.Device != nil {
		dnsLog.DeviceID = log.Device.ID
		dnsLog.DeviceName = log.Device.Name
		dnsLog.DeviceModel = log.Device.Model
	}

	return dnsLog
}
