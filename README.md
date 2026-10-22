# Kiara Chatbot (Telegram AI Bot in Go)

A high-performance, mature Telegram AI chatbot built with **Golang** and **gogram** (`github.com/amarnathcjd/gogram`).

Powered by Google's flagship **Gemini 4 Argon** (`argon-4`) model via `unified-web2api`, running the realistic **Gen Z "Kiara"** persona (sassy, human, low emojis, playful nakhre, and smart common sense).

---

## 🌟 Features

* **Trigger System (DM & Group Chats):**
  * **Direct Message (DM):** Responds automatically to every message.
  * **Group Chats (GC):**
    * Tagged via `@BotUsername`
    * Replied to bot's messages
    * Triggered by name (`kiara` or `kia`)
* **Dynamic Model Switching:**
  * Switch models live using `/model <name>` (e.g. `/model argon-4`, `/model gpt-4o`, `/model deepseek-chat`).
  * Check all models with `/models`.
* **Stateful Chat Memory:**
  * Thread-safe (`sync.RWMutex`) per-chat history.
  * Automatic pruning of older turns to avoid context overflow.
  * Automatic TTL cleanup for inactive chats.
  * Reset memory anytime with `/clear`.
* **Live "Typing..." Action:**
  * Emits Telegram `typing` action while waiting for AI generation.
* **Hot-Reload System Prompt:**
  * Edit `system_prompt.txt` on disk and run `/reload` to update personality without restarting the bot.

---

## 🚀 Quick Setup

### 1. Configure `.env`
Edit `/home/ubuntu/kiara-chatbot/.env`:
```env
BOT_TOKEN=your_bot_token_from_botfather
API_ID=38674666
API_HASH=b4f0fbf8fb560c4bc9e7b9f3698e474c
AI_API_URL=http://localhost:8081/v1/chat/completions
AI_API_KEY=sudeepbots
DEFAULT_MODEL=argon-4@think=low
BOT_NAME=kiara
```

### 2. Build & Run
```bash
cd /home/ubuntu/kiara-chatbot
go build -o kiara cmd/bot/main.go
./kiara
```

---

## 🤖 Bot Commands

| Command | Description |
| :--- | :--- |
| `/start` | Welcome message from Kiara |
| `/help` | Show command list & group trigger rules |
| `/models` | List all available AI models |
| `/model <name>` | Change model for current chat (e.g. `/model argon-4`) |
| `/clear` | Clear chat history & reset memory |
| `/status` | View active model & memory stats |
| `/reload` | Reload `system_prompt.txt` from disk (Admin) |
