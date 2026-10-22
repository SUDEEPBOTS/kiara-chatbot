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
	log.Println("[INFO] Starting Kiara Chatbot...")

	cfg := config.Load()
	if cfg.BotToken == "" {
		log.Println("[WARN] BOT_TOKEN is empty. Set it in .env before running.")
	}

	aiClient := ai.NewClient(cfg.AiAPIURL, cfg.AiAPIKey)
	memStore := memory.NewStore(cfg.MaxHistory, 2*time.Hour)

	// Leaky Bucket Rate Limiter ("Baltti") to prevent sticker spam in DMs
	// Burst capacity: 3 stickers, drain rate: 0.33 stickers/sec (1 per 3s), temporary block: 60s
	limiter := ratelimit.NewLeakyBucket(3.0, 0.33, 60*time.Second)

	// Dynamic Sticker Pool for saving and sending Telegram stickers
	stickersPool := stickers.NewPool("data/stickers_pool.json")

	// Auto-Moderation Profanity Filter (dots, unicode fonts, leetspeak, fuzzy matching)
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

		// Register command handlers
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

		// Register callback query handler for inline button pagination & settings
		client.AddCallbackHandler("", bh.HandleCallback)

		// Register global message listener for DM and GC mentions/replies
		client.AddMessageHandler("", bh.HandleIncomingMessage)

		log.Println("[INFO] Kiara Chatbot is now active and listening for messages!")
	} else {
		log.Println("[INFO] Bot compiled successfully. Add BOT_TOKEN in .env and start the binary.")
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] Shutting down Kiara Chatbot gracefully...")
}
