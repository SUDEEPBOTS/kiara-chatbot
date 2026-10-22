# Kiara Chatbot (Telegram AI Bot in Go)

[![Go CI](https://github.com/your-username/kiara-chatbot/actions/workflows/ci.yml/badge.svg)](https://github.com/your-username/kiara-chatbot/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/kiara-chatbot)](https://goreportcard.com/report/github.com/your-username/kiara-chatbot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Image](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com/)

A high-performance, mature Telegram AI chatbot built with **Golang** and **gogram** (`github.com/amarnathcjd/gogram`).

Powered by Google's flagship **Gemini 4 Argon** (`argon-4`) model via `unified-web2api`, running the realistic 500+ line **Gen Z "Kiara"** persona (sassy, human, low emojis, playful nakhre, smart common sense, and multimodal perception).

---

## 🌟 Key Features

### 1. 🧠 Multimodal Vision (Photos & Stickers)
* **Photo Perception:** Send any photo (with or without caption); Kiara perceives the image via multimodal vision (`image_url` base64 upload to Gemini) and reacts naturally.
* **Static Stickers (WebP):** Automatically converted to universal JPEG via FFmpeg and processed with Vision.
* **Video Stickers (WebM):** Video stickers are dynamically sampled via FFmpeg (`-ss 00:00:00.5 -vframes 1`) to extract a crisp middle frame for AI vision perception.
* **Animated Stickers (TGS / Lottie):** Thumbnail extracted and passed to Vision.

### 2. 🛡️ Auto-Moderation & GitHub-Safe Blocklist
* **Zero GitHub Risk:** Blocklists are hosted remotely on Yukisbox (`ABUSE_REMOTE_URL`) or cloud storage so your repository remains clean and protected from TOS bans.
* **Evasion Detection:** Automatically strips dots (`f.u.c.k`), converts stylized Unicode/math fonts to ASCII, decodes leetspeak (`@` ➔ `a`, `0` ➔ `o`), collapses repeating letters, and catches minor typos via Levenshtein distance.
* **Timed Warning:** Deletes the abusive message instantly, mentions the offender inside a `<blockquote>` warning, and auto-deletes the warning after 8 seconds.
* **Admin Toggle:** Toggle auto-moderation on or off anytime with `/abuse on` and `/abuse off`.

### 3. 🪣 Leaky Bucket Sticker Anti-Spam ("Baltti" Limiter)
* Built-in token/leaky bucket rate limiter in direct messages (DMs).
* Burst capacity of 3 stickers, leaking at 1 sticker per 3 seconds.
* If a user spams stickers rapidly, the bucket overflows:
  * Kiara temporarily blocks them for 60 seconds with a sassy warning (`"bro stop spamming stickers 🙄 baltti bhar gayi teri! blocked for 60s, shanti rakh."`).
  * Incoming spam is dropped until cooldown expires.

### 4. 🎨 Dynamic Sticker Pool & AI Sticker Replies
* Kiara automatically collects cool stickers sent by users into `data/stickers_pool.json`.
* When replying or asked (`"ek sticker bhej"`, or if she decides via `[SEND_STICKER]`), Kiara sends a real Telegram sticker from her collection!

### 5. 👤 Live Chat & User Metadata Context
* Kiara dynamically receives real-time Telegram metadata:
  * **User Info:** Real Name, Username, User ID.
  * **Group Info:** Group Name, Group Username, Chat ID.
* If user asks `"mera naam kya hai?"` or `"who am i?"`, Kiara answers casually with their real name (e.g. `"sudeep, obv. it's written right there 🙄"`).
* If asked about group details (`"is group ka naam kya hai?"`), Kiara knows without being robotic.

### 6. 👥 Independent Per-User Sessions
* Every user has an isolated session:
  * **DM:** `dm_<userID>`
  * **Group Chat:** `gc_<chatID>_u_<userID>`
* User A and User B in the same group maintain separate conversation histories and active model/thinking settings.

### 7. 📱 Interactive Inline Buttons with Pagination
* **`/models`** command displays 12+ AI models (Gemini 4 Argon, DeepSeek V3/R1, GPT-4o, Claude 3.5 Sonnet, etc.).
* **`[◀️ Prev]`** and **`[Next ▶️]`** pagination buttons.
* Active model indicator **`[✅ ActiveModel]`**.
* **`[⚡ Think: LOW / MED / HIGH / OFF]`** dynamic toggle button right on the inline keyboard.

### 8. 🎨 AI Image Generation
* `/imagine <prompt>` or `/draw <prompt>` to generate images via `unified-web2api`.

---

## 🚀 Deployment Options

### Option A: Local / VPS Binary
```bash
cd /home/ubuntu/kiara-chatbot
cp .env.example .env
# Edit .env and insert your BOT_TOKEN
go build -o kiara cmd/bot/main.go
./kiara
```

### Option B: Docker & Docker Compose
```bash
docker compose up -d --build
# View live logs
docker compose logs -f
```

### Option C: Render Cloud Deployment
1. Connect your GitHub repository to [Render](https://render.com).
2. Render will automatically detect [`render.yaml`](file:///home/ubuntu/kiara-chatbot/render.yaml) and configure the background worker.
3. Fill in your environment variables (`BOT_TOKEN`, `API_ID`, `API_HASH`) in the Render dashboard and deploy!

### Option D: Heroku 1-Click Deploy
Click the button below or push with the Heroku CLI:

[![Deploy to Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

```bash
heroku create kiara-bot --manifest
heroku git:remote -a kiara-bot
git push heroku main
```

---

## 🤖 Bot Commands

| Command | Description | Permission |
| :--- | :--- | :--- |
| `/start` | Welcome message from Kiara | Public |
| `/help` | Show command list & group trigger rules | Public |
| `/models` | Open interactive **Inline Buttons** menu with pagination & think toggle | Public |
| `/model <name>` | Change model directly via text or open inline menu if empty | Public |
| `/imagine <prompt>` | Generate AI images (`/draw` also works) | Public |
| `/clear` | Clear chat history & reset memory for your session | Public |
| `/abuse <on\|off>` | Toggle auto-moderation for the group | Group Admins |
| `/status` | View active model, thinking depth, active sessions & sticker pool count | Public |
| `/reload` | Reload `system_prompt.txt` & remote blocklist from URL | Bot Owner |

---

## 📁 Repository Structure

```
kiara-chatbot/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml               # Automated CI test & build pipeline
│   │   └── release.yml          # Multi-arch binary releases
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yml       # Structured bug report form
│   │   └── feature_request.yml  # Feature enhancement request form
│   ├── PULL_REQUEST_TEMPLATE.md # PR guidelines & verification checklist
│   ├── dependabot.yml           # Automated dependency updater
│   └── FUNDING.yml              # GitHub Sponsors & donation channels
├── cmd/
│   └── bot/
│       └── main.go              # Production entrypoint with UTC timestamp logs
├── internal/
│   ├── ai/
│   │   └── client.go            # Multimodal AI client (chat, vision, images)
│   ├── config/
│   │   └── config.go            # Environment variable loader
│   ├── handlers/
│   │   ├── bot.go               # DM/GC message triggers & vision router
│   │   └── commands.go          # Inline pagination & admin commands
│   ├── media/
│   │   └── processor.go         # FFmpeg photo & webm video sticker extractor
│   ├── memory/
│   │   └── session.go           # Thread-safe per-user isolated memory store
│   ├── moderation/
│   │   ├── filter.go            # De-obfuscation & auto-moderation engine
│   │   └── filter_test.go       # Moderation test suite
│   ├── ratelimit/
│   │   └── bucket.go            # Leaky Bucket ("Baltti") anti-spam limiter
│   └── stickers/
│       └── pool.go              # Telegram sticker collection & response pool
├── Dockerfile                   # Multi-stage minimal Alpine container with FFmpeg
├── docker-compose.yml           # Production Docker Compose stack
├── render.yaml                  # Render cloud blueprint specification
├── heroku.yml                   # Heroku container deployment manifest
├── Procfile                     # Heroku process definition
├── app.json                     # Heroku 1-Click deploy manifest
├── system_prompt.txt            # 500+ line Gen Z Kiara persona definition
├── .env.example                 # Sample configuration template
├── .gitignore                   # Security rules (ignores tokens, sessions & wordlists)
├── CONTRIBUTING.md              # Contribution standards & commit guidelines
├── CODE_OF_CONDUCT.md           # Contributor Covenant v2.1
├── SECURITY.md                  # Vulnerability disclosure policy
├── LICENSE                      # MIT Open-Source License
└── README.md                    # Complete project documentation
```
