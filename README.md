# Kiara Chatbot (Telegram AI Bot in Go)

A high-performance, mature Telegram AI chatbot built with **Golang** and **gogram** (`github.com/amarnathcjd/gogram`).

Powered by Google's flagship **Gemini 4 Argon** (`argon-4`) model via `unified-web2api`, running the realistic 466-line **Gen Z "Kiara"** persona (sassy, human, low emojis, playful nakhre, and smart common sense).

---

## 🌟 Features

* **Trigger System (DM & Group Chats):**
  * **Direct Message (DM):** Responds automatically to every message.
  * **Group Chats (GC):**
    * Tagged via `@BotUsername`
    * Replied to bot's messages
    * Triggered by name (`kiara` or `kia`)
* **Interactive Inline Model Selector & Pagination:**
  * Interactive inline keyboard with **`[◀️ Prev]`** and **`[Next ▶️]`** buttons.
  * Active indicator **`[✅ ActiveModel]`** for the currently running AI model.
  * **⚡ Thinking Depth Cycle Button:** Toggle between `LOW`, `MEDIUM`, `HIGH`, and `OFF` directly from Telegram inline buttons!
  * Support for 12+ models (Gemini 4 Argon, Gemini 3.1 Pro, DeepSeek V3, DeepSeek R1, GPT-4o, Claude 3.5 Sonnet, Grok 3, etc.).
* **🎨 AI Image Generation:**
  * `/imagine <prompt>` or `/draw <prompt>` to generate images via `unified-web2api` (`/v1/images/generations`).
  * Shows Telegram `upload_photo` status action and attaches high-res photo.
* **Stateful Chat Memory:**
  * Thread-safe (`sync.RWMutex`) per-chat history.
  * Automatic pruning of older turns to avoid context overflow.
  * Automatic TTL cleanup for inactive chats.
  * Reset memory anytime with `/clear`.
* **Live "Typing..." Action:**
  * Emits Telegram `typing` action pulse while waiting for AI generation.
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
| `/models` | Open interactive **Inline Buttons** menu with pagination & think toggle |
| `/model <name>` | Change model directly via text or open inline menu if empty |
| `/imagine <prompt>` | Generate AI images (`/draw` also works) |
| `/clear` | Clear chat history & reset memory |
| `/status` | View active model, thinking depth & memory stats |
| `/reload` | Reload `system_prompt.txt` from disk (Admin) |

---

## 📁 Repository Structure

```
kiara-chatbot/
├── .env                  # Configuration & Bot credentials
├── .env.example          # Sample environment variables
├── .gitignore            # Git ignore for binary & session tokens
├── README.md             # Documentation
├── go.mod / go.sum       # Go module dependencies (gogram)
├── system_prompt.txt     # 466-line Gen Z Kiara persona definition
├── cmd/
│   └── bot/
│       └── main.go       # Entry point & handler registration
└── internal/
    ├── ai/
    │   └── client.go     # Resilient AI client (chat + image generation)
    ├── config/
    │   └── config.go     # Environment loader
    ├── handlers/
    │   ├── bot.go        # DM/GC message triggers & typing pulse
    │   └── commands.go   # Inline keyboard pagination & bot commands
    └── memory/
        └── session.go    # Thread-safe per-chat session & memory store
```
