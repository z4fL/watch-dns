package config

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	NextDNS   NextDNSConfig
	Telegram  TelegramConfig
	Scheduler SchedulerConfig
	Log       LogConfig
	Rule      RuleConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	Path string
}

type NextDNSConfig struct {
	ProfileID string
	APIKey    string
	LogStream bool
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

type SchedulerConfig struct {
	ReportTime string
	Timezone   string
}

type RuleConfig struct {
	RiskThreshold int
}

type LogConfig struct {
	Level string
}
