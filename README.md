# Watch DNS

Watch DNS or can be named to DNS Guardian is a lightweight self hosted application that monitors DNS activity using the NextDNS Logs Stream API.

It automatically detects suspicious gambling domains through a configurable rule engine and blocks them by adding the domain to the NextDNS Denylist using the official NextDNS API.

This project is intended for personal and family usage with an initial focus on monitoring a single Android device.

## Features

### MVP

* Real time DNS monitoring via NextDNS Logs Stream (SSE)
* Automatic gambling domain detection
* Configurable rule based scoring engine
* Automatic domain blocking through NextDNS API
* SQLite storage for logs and decisions
* Daily Telegram summary report
* Lightweight REST API for monitoring

## Architecture

This project follows a simple layered architecture.

```
handler
    ↓
service
    ↓
repository
    ↓
database
```

Project structure:

```text
cmd/
internal/
├── config/
├── database/
├── handler/
├── model/
├── nextdns/
├── repository/
├── ruleengine/
├── scheduler/
├── service/
└── telegram/

migrations/
pkg/
```

## Tech Stack

* Go 1.25+
* Chi
* GORM
* SQLite
* NextDNS API
* Telegram Bot API
* slog
* godotenv

## Rule Engine

Every DNS query receives a configurable risk score.

Example rules:

| Rule | Score |
|------|------:|
| slot | +30 |
| gacor | +25 |
| casino | +30 |
| togel | +30 |
| bet | +20 |
| .top | +10 |
| .xyz | +10 |
| 777 / 888 / 168 | +10 |

Domains with a total score greater than or equal to **80** will be automatically blocked.

## REST API

| Method | Endpoint | Description |
|---------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/stats` | Statistics |
| GET | `/logs` | DNS logs |
| GET | `/blocked` | Blocked domains |
| POST | `/allowlist` | Add allowlist entry |
| DELETE | `/blocked/{id}` | Remove blocked domain |

## Database

SQLite tables:

* dns_logs
* blocked_domains
* rule_matches
* settings
* telegram_reports
* allowlist

## Getting Started

### Prerequisites

* Go 1.24 or newer
* SQLite
* NextDNS Account
* Telegram Bot Token

### Clone repository

```bash
git clone https://github.com/<username>/dns-guardian.git
cd dns-guardian
```

### Install dependencies

```bash
go mod tidy
```

### Configure environment

```bash
cp .env.example .env
```

Edit the `.env` file with your own configuration.

### Run

```bash
go run ./cmd/server
```

## Roadmap

### MVP

- [ ] Connect to NextDNS Logs Stream
- [ ] Store DNS logs
- [ ] Rule based detection
- [ ] Automatic denylist update
- [ ] Telegram daily report
- [ ] REST API

### Future

- [ ] Multiple devices
- [ ] Multiple NextDNS profiles
- [ ] PostgreSQL support
- [ ] AI assisted scoring
- [ ] Domain reputation lookup
- [ ] Web dashboard
- [ ] Authentication
- [ ] Role management
- [ ] Docker deployment

## License

This project is licensed under the MIT License.
