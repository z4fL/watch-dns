package service

import (
	"github.com/z4fL/watch-dns/internal/models"
	"github.com/z4fL/watch-dns/internal/nextdns"
)

func toDNSLog(log nextdns.Log) models.DNSLog {
	dnsLog := models.DNSLog{
		Timestamp: log.Timestamp,
		Domain:    log.Domain,
		Root:      log.Root,
		Encrypted: log.Encrypted,
		Protocol:  log.Protocol,
		ClientIP:  log.ClientIP,
		Client:    log.Client,
		Status:    log.Status,
	}

	if log.Device != nil {
		dnsLog.DeviceID = log.Device.ID
		dnsLog.DeviceName = log.Device.Name
		dnsLog.DeviceModel = log.Device.Model
	}

	return dnsLog
}
