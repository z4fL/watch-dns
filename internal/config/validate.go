package config

import (
	"fmt"
	"time"
)

func (c *Config) Validate() error {
	if c.Database.Path == "" {
		return fmt.Errorf("DB_PATH is required")
	}

	if c.NextDNS.ProfileID == "" {
		return fmt.Errorf("NEXTDNS_PROFILE_ID is required")
	}

	if c.NextDNS.APIKey == "" {
		return fmt.Errorf("NEXTDNS_API_KEY is required")
	}

	if c.Telegram.BotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if c.Telegram.ChatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	if c.Rule.RiskThreshold <= 0 {
		return fmt.Errorf("RISK_THRESHOLD must be greater than zero")
	}

	if _, err := time.Parse("15:04", c.Scheduler.ReportTime); err != nil {
		return fmt.Errorf("invalid REPORT_TIME format, expected HH:MM")
	}

	if _, err := time.LoadLocation(c.Scheduler.Timezone); err != nil {
		return fmt.Errorf("invalid TIMEZONE")
	}

	return nil
}
