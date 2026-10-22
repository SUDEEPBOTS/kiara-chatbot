# Kiara Chatbot (Telegram AI Bot in Go)

A high-performance, mature Telegram AI chatbot built with **Golang** and **gogram** (`github.com/amarnathcjd/gogram`).

Powered by Google's flagship **Gemini 4 Argon** (`argon-4`) model via `unified-web2api`, running the realistic 500+ line **Gen Z "Kiara"** persona (sassy, human, low emojis, playful nakhre, smart common sense, and multimodal perception).

---

## 🌟 Features

### 1. 🧠 Multimodal Vision (Photos & Stickers)
* **Photo Understanding:** Send any photo (with or without caption); Kiara perceives the image via multimodal vision (`image_url` base64 upload to Gemini) and reacts naturally.
* **Static Stickers (WebP):** Automatically converted to universal JPEG via FFmpeg and fed into Vision.
* **Video Stickers (WebM):** Video stickers are dynamically sampled via FFmpeg (`-ss 00:00:00.5 -vframes 1`) to extract a crisp middle frame for AI vision perception.
* **Animated Stickers (TGS / Lottie):** Thumbnail extracted and passed to Vision.

### 2. 🪣 Leaky Bucket Sticker Anti-Spam ("Baltti" Rate Limiter)
* Built-in token/leaky bucket rate limiter in direct messages (DMs).
* Burst capacity of 3 stickers, leaking at 1 sticker per 3 seconds.
* If a user spams stickers rapidly, the bucket overflows:
  * Kiara temporarily blocks them for 60 seconds with a sassy warning (`"bro stop spamming stickers 🙄 baltti bhar gayi teri! blocked for 60s, shanti rakh."`).
  * Incoming spam is dropped until cooldown expires.

### 3. 🎨 Dynamic Sticker Pool & AI Sticker Replies
* Kiara automatically collects cool stickers sent by users into `data/stickers_pool.json`.
* When replying or asked (`"ek sticker bhej"`, or if she decides via `[SEND_STICKER]`), Kiara sends a real Telegram sticker from her collection!

### 4. 👤 Live Chat & User Metadata Context
* Kiara dynamically receives real-time Telegram metadata:
  * **User Info:** Real Name, Username, User ID.
  * **Group Info:** Group Name, Group Username, Chat ID.
* If user asks `"mera naam kya hai?"` or `"who am i?"`, Kiara answers casually with their real name (e.g. `"sudeep, obv. it's written right there 🙄"`).
* If asked about group details (`"is group ka naam kya hai?"`), Kiara knows without being robotic.

### 5. 👥 Independent Per-User Sessions
* Every user has an isolated session:
  * **DM:** `dm_<userID>`
  * **Group Chat:** `gc_<chatID>_u_<userID>`
* User A and User B in the same group maintain separate conversation histories and active model/thinking settings.

### 6. 📱 Interactive Inline Buttons with Pagination
* **`/models`** command displays 12+ AI models (Gemini 4 Argon, DeepSeek V3/R1, GPT-4o, Claude 3.5 Sonnet, etc.).
* **`[◀️ Prev]`** and **`[Next ▶️]`** pagination buttons.
* Active model indicator **`[✅ ActiveModel]`**.
* **`[⚡ Think: LOW / MED / HIGH / OFF]`** dynamic toggle button right on the inline keyboard.

### 7. 🎨 AI Image Generation
* `/imagine <prompt>` or `/draw <prompt>` to generate images via `unified-web2api`.

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
| `/clear` | Clear chat history & reset memory for your session |
| `/status` | View active model, thinking depth, active sessions & sticker pool count |
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
├── system_prompt.txt     # 500+ line Gen Z Kiara persona definition
├── cmd/
│   └── bot/
│       └── main.go       # Entry point & event registration
└── internal/
    ├── ai/
    │   └── client.go     # Multimodal AI client (chat, vision, image generation)
    ├── config/
    │   └── config.go     # Environment loader
    ├── handlers/
    │   ├── bot.go        # DM/GC message triggers, typing pulse & vision handler
    │   └── commands.go   # Inline pagination, model buttons & commands
    ├── media/
    │   └── processor.go  # FFmpeg sticker frame extraction & photo vision processor
    ├── memory/
    │   └── session.go    # Thread-safe per-user isolated memory store
    ├── ratelimit/
    │   └── bucket.go     # Leaky Bucket ("Baltti") anti-spam limiter
    └── stickers/
        └── pool.go       # Telegram sticker collection & response pool
```
