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
	AiAPIURL         string
	AiAPIKey         string
	DefaultModel     string
	SystemPromptPath string
	BotName          string
	MaxHistory       int
	OwnerID          int64
	LogLevel         string
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func Load() *Config {
	// Load .env if present
	_ = godotenv.Load()

	appIDInt, _ := strconv.Atoi(getEnv("API_ID", "38674666"))
	ownerIDInt, _ := strconv.ParseInt(getEnv("OWNER_ID", "0"), 10, 64)
	maxHistoryInt, _ := strconv.Atoi(getEnv("MAX_HISTORY", "20"))

	cfg := &Config{
		AppID:            int32(appIDInt),
		AppHash:          getEnv("API_HASH", "b4f0fbf8fb560c4bc9e7b9f3698e474c"),
		BotToken:         getEnv("BOT_TOKEN", ""),
		AiAPIURL:         getEnv("AI_API_URL", "http://localhost:8081/v1/chat/completions"),
		AiAPIKey:         getEnv("AI_API_KEY", "sudeepbots"),
		DefaultModel:     getEnv("DEFAULT_MODEL", "argon-4@think=low"),
		SystemPromptPath: getEnv("SYSTEM_PROMPT_PATH", "system_prompt.txt"),
		BotName:          strings.ToLower(getEnv("BOT_NAME", "kiara")),
		MaxHistory:       maxHistoryInt,
		OwnerID:          ownerIDInt,
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
	}

	if cfg.BotToken == "" {
		log.Println("[WARN] BOT_TOKEN is empty! Please set it in .env")
	}

	return cfg
}
