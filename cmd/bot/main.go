// ==============================================================================
// Copyright (c) 2026 SUDEEPBOTS. All rights reserved.
//
// Project: Kiara Chatbot - High-Performance Telegram AI Assistant
// Repository: https://github.com/SUDEEPBOTS/kiara-chatbot
// License: MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
// ==============================================================================

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"kiara-chatbot/internal/ai"
	"kiara-chatbot/internal/config"
	"kiara-chatbot/internal/handlers"
	"kiara-chatbot/internal/memory"
	"kiara-chatbot/internal/moderation"
	"kiara-chatbot/internal/ratelimit"
	"kiara-chatbot/internal/stickers"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	log.Println("[INFO] Starting Kiara Chatbot service...")

	cfg := config.Load()
	if cfg.BotToken == "" {
		log.Println("[WARN] BOT_TOKEN is empty. Set it in .env before running.")
	}

	aiClient := ai.NewClient(cfg.AiAPIURL, cfg.AiAPIKey, cfg.GroqAPIKey, cfg.GeminiAPIKey, cfg.AIProvider)
	memStore := memory.NewStore(cfg.MaxHistory, 2*time.Hour)
	limiter := ratelimit.NewLeakyBucket(3.0, 0.33, 60*time.Second)
	stickersPool := stickers.NewPool("data/stickers_pool.json")
	modFilter := moderation.NewFilter(cfg.AbuseFilePath, cfg.AbuseRemoteURL)

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     cfg.AppID,
		AppHash:   cfg.AppHash,
		LogLevel:  telegram.LogInfo,
		ParseMode: "HTML",
		Session:   "kiara_bot.session",
	})
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize telegram client: %v", err)
	}

	if cfg.BotToken != "" {
		if err := client.LoginBot(cfg.BotToken); err != nil {
			log.Fatalf("[FATAL] Failed to login bot: %v", err)
		}

		me, err := client.GetMe()
		if err != nil {
			log.Fatalf("[FATAL] Failed to fetch bot identity: %v", err)
		}

		log.Printf("[SUCCESS] Logged in as @%s (ID: %d)", me.Username, me.ID)

		bh := handlers.NewBotHandler(client, aiClient, memStore, cfg, me, limiter, stickersPool, modFilter)

		client.AddCommandHandler("start", bh.HandleStart)
		client.AddCommandHandler("help", bh.HandleHelp)
		client.AddCommandHandler("models", bh.HandleModels)
		client.AddCommandHandler("model", bh.HandleModel)
		client.AddCommandHandler("clear", bh.HandleClear)
		client.AddCommandHandler("status", bh.HandleStatus)
		client.AddCommandHandler("reload", bh.HandleReload)
		client.AddCommandHandler("imagine", bh.HandleImagine)
		client.AddCommandHandler("draw", bh.HandleImagine)
		client.AddCommandHandler("abuse", bh.HandleAbuse)
		client.AddCommandHandler("automod", bh.HandleAbuse)

		client.AddCallbackHandler("", bh.HandleCallback)
		client.AddMessageHandler("", bh.HandleIncomingMessage)

		log.Println("[INFO] Kiara Chatbot is now active and listening for messages!")
	} else {
		log.Println("[INFO] Bot compiled successfully. Add BOT_TOKEN in .env and start the binary.")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] Shutting down Kiara Chatbot gracefully...")
}
