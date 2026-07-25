package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	// Ignore error if .env doesn't exist.
	// Production can rely on system environment variables.

	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "guarddns"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},

		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/guarddns.db"),
		},

		NextDNS: NextDNSConfig{
			ProfileID: getEnv("NEXTDNS_PROFILE_ID", ""),
			APIKey:    getEnv("NEXTDNS_API_KEY", ""),
			LogStream: getEnvAsBool("NEXTDNS_LOG_STREAM", true),
		},

		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
			ChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		},

		Scheduler: SchedulerConfig{
			ReportTime: getEnv("REPORT_TIME", "21:00"),
			Timezone:   getEnv("TIMEZONE", "Asia/Jakarta"),
		},

		Rule: RuleConfig{
			RiskThreshold: getEnvAsInt("RISK_THRESHOLD", 80),
		},

		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return v
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}
