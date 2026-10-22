// ==============================================================================
// Copyright (c) 2026 SUDEEPBOTS (Sudeep Sharma). All rights reserved.
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

package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppID            int32
	AppHash          string
	BotToken         string
	AIProvider       string
	AiAPIURL         string
	AiAPIKey         string
	GroqAPIKey       string
	GeminiAPIKey     string
	DefaultModel     string
	SystemPromptPath string
	BotName          string
	MaxHistory       int
	OwnerID          int64
	LogLevel         string
	AbuseFilePath    string
	AbuseRemoteURL   string
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func Load() *Config {
	_ = godotenv.Load()

	appIDInt, _ := strconv.Atoi(getEnv("API_ID", "38674666"))
	ownerIDInt, _ := strconv.ParseInt(getEnv("OWNER_ID", "0"), 10, 64)
	maxHistoryInt, _ := strconv.Atoi(getEnv("MAX_HISTORY", "20"))

	cfg := &Config{
		AppID:            int32(appIDInt),
		AppHash:          getEnv("API_HASH", "b4f0fbf8fb560c4bc9e7b9f3698e474c"),
		BotToken:         getEnv("BOT_TOKEN", ""),
		AIProvider:       strings.ToLower(getEnv("AI_PROVIDER", "web2api")),
		AiAPIURL:         getEnv("AI_API_URL", "http://localhost:8081/v1/chat/completions"),
		AiAPIKey:         getEnv("AI_API_KEY", "sudeepbots"),
		GroqAPIKey:       getEnv("GROQ_API_KEY", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		DefaultModel:     getEnv("DEFAULT_MODEL", "argon-4@think=low"),
		SystemPromptPath: getEnv("SYSTEM_PROMPT_PATH", "system_prompt.txt"),
		BotName:          strings.ToLower(getEnv("BOT_NAME", "kiara")),
		MaxHistory:       maxHistoryInt,
		OwnerID:          ownerIDInt,
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
		AbuseFilePath:    getEnv("ABUSE_FILE_PATH", "abouse.txt"),
		AbuseRemoteURL:   getEnv("ABUSE_REMOTE_URL", ""),
	}

	if cfg.BotToken == "" {
		log.Println("[WARN] BOT_TOKEN is empty. Set it in .env")
	}

	return cfg
}
