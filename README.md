# Kiara Chatbot

<p align="center">
  <img src=".github/assets/banner.png" alt="Kiara Chatbot Banner" width="100%"/>
</p>

<p align="center">
  <strong>Enterprise-Grade, High-Performance Gen Z Telegram AI Chatbot & Auto-Moderation Suite in Golang.</strong>
</p>

<p align="center">
  <a href="https://github.com/SUDEEPBOTS/kiara-chatbot/actions/workflows/ci.yml"><img src="https://github.com/SUDEEPBOTS/kiara-chatbot/actions/workflows/ci.yml/badge.svg?branch=master" alt="CI Status"/></a>
  <a href="https://goreportcard.com/report/github.com/SUDEEPBOTS/kiara-chatbot"><img src="https://goreportcard.com/badge/github.com/SUDEEPBOTS/kiara-chatbot" alt="Go Report Card"/></a>
  <a href="https://pkg.go.dev/github.com/amarnathcjd/gogram"><img src="https://img.shields.io/badge/Powered%20By-gogram-00ADD8.svg" alt="gogram"/></a>
  <a href="https://hub.docker.com/"><img src="https://img.shields.io/badge/Docker-Ready-2496ED.svg?logo=docker&logoColor=white" alt="Docker Ready"/></a>
  <a href="https://t.me/Z0iiw"><img src="https://img.shields.io/badge/Telegram-@Z0iiw-2CA5E0.svg?logo=telegram&logoColor=white" alt="Telegram"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"/></a>
</p>

---

## 📑 Table of Contents

- [Architectural Overview](#-architectural-overview)
- [Key Features](#-key-features)
  - [Multimodal Vision & Media Processing](#1-multimodal-vision--media-processing)
  - [Autonomous Natural Language Image Generation](#2-autonomous-natural-language-image-generation)
  - [Interactive Inline Pagination & Model Selector](#3-interactive-inline-pagination--model-selector)
  - [Evasion-Proof Auto-Moderation Engine](#4-evasion-proof-auto-moderation-engine)
  - [Cloud Blocklist Hosting (GitHub-Safe)](#5-cloud-blocklist-hosting-github-safe)
  - [Leaky Bucket Anti-Spam Rate Limiter](#6-leaky-bucket-anti-spam-rate-limiter)
  - [Dynamic Sticker Pool & Real-Time Context](#7-dynamic-sticker-pool--real-time-context)
- [Bot Commands & Triggers](#-bot-commands--triggers)
- [Deployment Guide](#-deployment-guide)
  - [Automated Shell Installer (Recommended)](#option-1-automated-shell-installer-recommended)
  - [Docker & Docker Compose](#option-2-docker--docker-compose)
  - [Render Cloud Deployment](#option-3-render-cloud-deployment)
  - [Heroku Container Deployment](#option-4-heroku-container-deployment)
  - [Systemd Background Service](#option-5-systemd-background-service)
- [Configuration Reference](#-configuration-reference)
- [Repository Structure](#-repository-structure)
- [Developer & Community](#-developer--community)
- [License](#-license)

---

## 🏛️ Architectural Overview

Kiara Chatbot is engineered from the ground up in **Go (Golang)** using native MTProto networking via `gogram` (`github.com/amarnathcjd/gogram`), offering low-latency message streaming, non-blocking asynchronous event handling, and minimal memory usage (< 25 MB RAM).

```
                      ┌────────────────────────────────────────┐
                      │         Telegram MTProto Gate          │
                      │        (github.com/amarnathcjd)        │
                      └──────────────────┬─────────────────────┘
                                         │
                                         ▼
                      ┌────────────────────────────────────────┐
                      │    Master Handler & Event Dispatcher   │
                      └──────┬───────────┬──────────────┬──────┘
                             │           │              │
         ┌───────────────────┘           │              └────────────────────┐
         ▼                               ▼                                   ▼
┌─────────────────┐             ┌─────────────────┐                 ┌─────────────────┐
│ Auto-Moderation │             │ Media Processor │                 │ Session Memory  │
│ - Unicode Norm  │             │ - FFmpeg Sample │                 │ - Per-User Key  │
│ - Dot Stripping │             │ - WebM Extraction│                │ - Turn Pruning  │
│ - Leetspeak Dec │             │ - Base64 Vision │                 │ - Auto Expire   │
└─────────────────┘             └────────┬────────┘                 └─────────────────┘
                                         │
                                         ▼
                      ┌────────────────────────────────────────┐
                      │       Unified AI Gateway Router        │
                      │  (Web2API / Groq Cloud / Gemini API)   │
                      └────────────────────────────────────────┘
```

---

## 🌟 Key Features

### 1. Multimodal Vision & Media Processing
* **Photos & High-Res Images:** Accepts images in DMs and Group Chats, automatically preparing base64 payload streams for Gemini Vision models.
* **Animated & Video Stickers (WebM):** Employs an integrated **FFmpeg** pipeline to extract an exact middle frame (`-ss 00:00:00.5 -vframes 1`) from Telegram animated video stickers on-the-fly, allowing Kiara to "see" and react to stickers authentically.
* **Static Stickers (WebP):** Converts WebP stickers to universal RGB JPEG formats for model compatibility.

### 2. Autonomous Natural Language Image Generation
Users do not need to memorize commands. Kiara recognizes natural conversational drawing requests in Hindi, Hinglish, and English:
* *"kiara ye image bana de: a futuristic neon cyber cat"*
* *"draw a peaceful sunset over the himalayas"*
* *"ek pyara sa puppy running on green grass ki photo banao"*
* Traditional command `/imagine <prompt>` is also supported.

### 3. Interactive Inline Pagination & Model Selector
* **Paginated Keyboard:** `/models` opens a compact inline keyboard featuring `[◀️ Prev]` and `[Next ▶️]` navigation buttons with zero message spam (smooth in-place edits via `cb.Edit`).
* **Active Model Indicator:** Highlights the current model with a `[✅ ModelName]` badge.
* **⚡ Dynamic Thinking Toggle:** Instantly cycles Gemini 4 reasoning depth (`LOW`, `MEDIUM`, `HIGH`, `OFF`) directly from Telegram inline buttons.
* **Supported Models:** Google Gemini 4 Argon, DeepSeek R1 & V3, Llama 3.3 70B (Groq), OpenAI GPT-4o, Claude 3.5 Sonnet, and Grok 3.

### 4. Evasion-Proof Auto-Moderation Engine
* **Unicode Homoglyphs & Fonts:** Converts stylized mathematical bold, fraktur, script, and fullwidth characters (e.g. `𝓯𝓾𝓬𝓴`, `ｆｕｃｋ`) to standard ASCII.
* **Separator & Dot Stripping:** Eliminates intentional evasion punctuation (`f.u.c.k`, `f_u_c_k`, `f u c k`).
* **Leetspeak Decoding:** Normalizes numerical substitutions (`@` ➔ `a`, `0` ➔ `o`, `1`/`!` ➔ `i`, `$` ➔ `s`, `3` ➔ `e`).
* **Fuzzy Typo Tolerance:** Evaluates Levenshtein edit distance (threshold &le; 1) for minor typos.
* **Instant Action:** Offending messages are immediately deleted (`m.Delete()`). The offender is tagged in an HTML `<blockquote>` warning, which automatically self-destructs after 8 seconds.

### 5. Cloud Blocklist Hosting (GitHub-Safe)
* **Zero Repository Risk:** To prevent GitHub repository bans or content flags, the blocklist is hosted remotely on **[Yukisbox](https://yukisbox.com/static/abuse.txt)** via HTTPS.
* **Dynamic Hot-Reloading:** Admins can trigger `/reload` to pull the latest blocklist updates without restarting the application.

### 6. Leaky Bucket Anti-Spam Rate Limiter
* Prevents sticker and media flooding in Direct Messages (DMs).
* Built on a classic leaky bucket algorithm: capacity of 3 sticker tokens, leaking at 1 token per 3 seconds.
* Overflows result in a temporary 60-second cooldown with a sassy Kiara warning.

### 7. Dynamic Sticker Pool & Real-Time Context
* Automatically harvests stickers sent by group members into an internal pool (`data/stickers_pool.json`).
* Kiara responds with real Telegram stickers when prompted or based on conversational context.
* Provides live chat metadata (sender name, user ID, chat title) to Kiara so she naturally knows who she is talking to.

---

## 🤖 Bot Commands & Triggers

### Available Commands

| Command | Description | Access Level |
| :--- | :--- | :--- |
| **`/start`** | Introduction and overview of Kiara | Public |
| **`/help`** | Displays available commands and group chat rules | Public |
| **`/models`** | Opens the interactive model switcher with inline pagination | Public |
| **`/model <id>`** | Switches AI model directly via text (e.g. `/model argon-4`) | Public |
| **`/imagine <prompt>`** | Generates high-resolution AI images | Public |
| **`/clear`** | Clears session history and resets conversation context | Public |
| **`/abuse <on\|off>`** | Enables or disables auto-moderation for the group | Group Admins |
| **`/status`** | Displays active model, thinking depth, session count & sticker stats | Public |
| **`/reload`** | Hot-reloads `system_prompt.txt` and remote blocklist | Bot Owner |

### Group Chat Triggers
In group chats, Kiara will respond only when addressed:
1. **Mention:** Tagged with `@BotUsername`
2. **Quote Reply:** Replying directly to any message sent by the bot
3. **Name Mention:** Including `"kiara"` or `"kia"` anywhere in the message

---

## 🚀 Deployment Guide

### Option 1: Automated Shell Installer (Recommended)

Run the automated interactive installer script on any Linux VPS (Ubuntu/Debian/CentOS):

```bash
cd /home/ubuntu/kiara-chatbot
chmod +x install.sh
./install.sh
```

The installer verifies system dependencies (Go, FFmpeg, CA certificates), creates your `.env`, builds the stripped binary, and optionally provisions a `systemd` background service.

---

### Option 2: Docker & Docker Compose

```bash
# 1. Clone repository
git clone https://github.com/SUDEEPBOTS/kiara-chatbot.git
cd kiara-chatbot

# 2. Configure environment
cp .env.example .env
nano .env

# 3. Build and launch container
docker compose up -d --build

# 4. Monitor logs
docker compose logs -f
```

---

### Option 3: Render Cloud Deployment

Deploy seamlessly to [Render](https://render.com) using the included [`render.yaml`](render.yaml) Blueprint:

1. Fork or push this repository to GitHub.
2. In the Render Dashboard, select **New +** ➔ **Blueprint**.
3. Connect your repository. Render will automatically parse `render.yaml` and configure the background worker.
4. Set your `BOT_TOKEN`, `API_ID`, and `API_HASH` environment variables in the Render dashboard.

---

### Option 4: Heroku Container Deployment

Deploy using the Heroku Container Registry:

```bash
# 1. Login to Heroku Container Registry
heroku container:login

# 2. Create app with container stack
heroku create kiara-bot --manifest
heroku git:remote -a kiara-bot

# 3. Configure environment variables
heroku config:set BOT_TOKEN="your_token" API_ID="38674666" API_HASH="b4f0fbf8fb560c4bc9e7b9f3698e474c"

# 4. Deploy
git push heroku master
```

---

### Option 5: Systemd Background Service

Create `/etc/systemd/system/kiara.service`:

```ini
[Unit]
Description=Kiara Telegram AI Chatbot Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/kiara-chatbot
ExecStart=/home/ubuntu/kiara-chatbot/kiara
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable kiara
sudo systemctl start kiara
sudo systemctl status kiara
```

---

## ⚙️ Configuration Reference

| Environment Variable | Description | Default Value | Required |
| :--- | :--- | :--- | :---: |
| `BOT_TOKEN` | Telegram Bot Token obtained from [@BotFather](https://t.me/BotFather) | *None* | **Yes** |
| `API_ID` | Telegram API ID from [my.telegram.org](https://my.telegram.org) | `38674666` | **Yes** |
| `API_HASH` | Telegram API Hash from [my.telegram.org](https://my.telegram.org) | `b4f0fbf8fb56...` | **Yes** |
| `AI_PROVIDER` | Backend routing (`web2api`, `groq`, `gemini`) | `web2api` | No |
| `AI_API_URL` | Base chat completions endpoint | `http://localhost:8081/v1/chat/completions` | No |
| `AI_API_KEY` | Authentication bearer token for AI endpoint | `sudeepbots` | No |
| `GROQ_API_KEY` | Optional API key for Groq Cloud inference | *None* | No |
| `GEMINI_API_KEY` | Optional official Google Gemini API key | *None* | No |
| `DEFAULT_MODEL` | Default model string | `argon-4@think=low` | No |
| `BOT_NAME` | Group chat trigger name | `kiara` | No |
| `MAX_HISTORY` | Maximum conversational turns retained per session | `20` | No |
| `OWNER_ID` | Telegram User ID of bot administrator | `0` | No |
| `ABUSE_REMOTE_URL` | Cloud-hosted blocklist URL (Yukisbox) | `https://yukisbox.com/static/abuse.txt` | No |
| `LOG_LEVEL` | Application logging verbosity (`INFO`, `WARN`, `ERROR`) | `INFO` | No |

---

## 📁 Repository Structure

```
kiara-chatbot/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml               # Automated CI test & build pipeline (master branch)
│   │   └── release.yml          # Multi-arch binary compilation & release
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yml       # Issue form for bug tracking
│   │   └── feature_request.yml  # Issue form for feature proposals
│   ├── PULL_REQUEST_TEMPLATE.md # Pull request verification checklist
│   ├── dependabot.yml           # Automated dependency update cadence
│   └── FUNDING.yml              # GitHub Sponsors & donation channels
├── cmd/
│   └── bot/
│       └── main.go              # Production entrypoint with UTC ISO logging
├── internal/
│   ├── ai/
│   │   └── client.go            # Multi-provider AI client (Web2API, Groq, Gemini)
│   ├── config/
│   │   └── config.go            # Environment variable loader
│   ├── handlers/
│   │   ├── bot.go               # Message triggers, natural image extraction & vision
│   │   └── commands.go          # Inline keyboard pagination & admin commands
│   ├── media/
│   │   └── processor.go         # FFmpeg photo & video sticker frame extractor
│   ├── memory/
│   │   └── session.go           # Thread-safe per-user isolated memory store
│   ├── moderation/
│   │   ├── filter.go            # De-obfuscation & auto-moderation engine
│   │   └── filter_test.go       # Moderation unit test suite
│   ├── ratelimit/
│   │   └── bucket.go            # Leaky Bucket ("Baltti") anti-spam limiter
│   └── stickers/
│       └── pool.go              # Telegram sticker collection & response pool
├── Dockerfile                   # Multi-stage minimal Alpine container with FFmpeg
├── docker-compose.yml           # Production Docker Compose orchestration
├── render.yaml                  # Render cloud blueprint specification
├── heroku.yml                   # Heroku container deployment manifest
├── Procfile                     # Heroku process definition
├── app.json                     # Heroku 1-Click deploy manifest
├── install.sh                   # Automated interactive Linux installer
├── system_prompt.txt            # 500+ line Gen Z Kiara persona definition
├── .env.example                 # Configuration environment template
├── .gitignore                   # Ignores sensitive tokens, sessions & wordlists
├── CONTRIBUTING.md              # Contributor conventions & commit standards
├── CODE_OF_CONDUCT.md           # Contributor Covenant v2.1
├── SECURITY.md                  # Security vulnerability reporting guidelines
├── LICENSE                      # MIT Open-Source License
└── README.md                    # Project documentation
```

---

## 👨‍💻 Developer & Community

Architected with ❤️ by **[SUDEEPBOTS](https://github.com/SUDEEPBOTS)**.

* 🐙 **GitHub:** [@SUDEEPBOTS](https://github.com/SUDEEPBOTS)
* ✈️ **Telegram:** [@Z0iiw](https://t.me/Z0iiw)
* 📸 **Instagram:** [@iam_billi00](https://www.instagram.com/iam_billi00)
* 🐦 **Twitter / X:** [@sudeeptelegram](https://x.com/sudeeptelegram)
* 🌐 **Websites:** [yukiapi.site](https://yukiapi.site) & [yukisbox.com](https://yukisbox.com)
* 📧 **Email:** [sudeeptelegram@gmail.com](mailto:sudeeptelegram@gmail.com)

---

## 📄 License

Distributed under the [MIT License](LICENSE). &copy; 2026 **SUDEEPBOTS**. All rights reserved.
