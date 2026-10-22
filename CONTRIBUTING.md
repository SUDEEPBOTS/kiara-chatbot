# Contributing to Kiara Chatbot

Thank you for your interest in contributing to Kiara Chatbot! We welcome contributions from the community.

---

## 🛠️ Development Setup

1. **Prerequisites:**
   - Go 1.23 or higher
   - FFmpeg (for media processing)
   - Docker & Docker Compose (optional)

2. **Clone and run:**
   ```bash
   git clone https://github.com/your-username/kiara-chatbot.git
   cd kiara-chatbot
   cp .env.example .env
   # Add your Telegram BOT_TOKEN in .env
   go build -o kiara cmd/bot/main.go
   ./kiara
   ```

---

## 📋 Git Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) standard:
- `feat: add new feature`
- `fix: resolve bug in moderation filter`
- `docs: update documentation in README`
- `refactor: optimize token bucket algorithm`
- `test: add unit tests for normalization`

---

## 🧪 Testing

Always verify before opening a pull request:
```bash
go vet ./...
go test -v ./...
```
