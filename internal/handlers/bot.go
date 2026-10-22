package handlers

import (
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"kiara-chatbot/internal/ai"
	"kiara-chatbot/internal/config"
	"kiara-chatbot/internal/memory"
)

type BotHandler struct {
	client       *telegram.Client
	ai           *ai.Client
	memory       *memory.Store
	cfg          *config.Config
	botUser      *telegram.UserObj
	systemPrompt string
	mu           sync.RWMutex
}

func NewBotHandler(client *telegram.Client, aiClient *ai.Client, memStore *memory.Store, cfg *config.Config, botUser *telegram.UserObj) *BotHandler {
	bh := &BotHandler{
		client:  client,
		ai:      aiClient,
		memory:  memStore,
		cfg:     cfg,
		botUser: botUser,
	}
	bh.ReloadSystemPrompt()
	return bh
}

func (b *BotHandler) ReloadSystemPrompt() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := os.ReadFile(b.cfg.SystemPromptPath)
	if err != nil {
		log.Printf("[ERROR] Failed to read system prompt from %s: %v", b.cfg.SystemPromptPath, err)
		return err
	}
	b.systemPrompt = string(data)
	log.Printf("[INFO] Loaded system prompt (%d bytes) from %s", len(b.systemPrompt), b.cfg.SystemPromptPath)
	return nil
}

func (b *BotHandler) GetSystemPrompt() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.systemPrompt
}

// ShouldRespond determines if the message is directed at the bot.
// In DM: Always responds.
// In GC: Responds only if:
// 1. Tagged by @username
// 2. Replying to bot's message
// 3. Name mention ("kiara" or "kia")
func (b *BotHandler) ShouldRespond(m *telegram.NewMessage) (bool, string) {
	text := strings.TrimSpace(m.Text())
	if text == "" {
		return false, ""
	}

	// Commands starting with '/' are handled by command handlers
	if strings.HasPrefix(text, "/") {
		return false, ""
	}

	// 1. Private chat (DM) -> Always respond
	if m.IsPrivate() {
		return true, text
	}

	// 2. Group chat
	botUsername := ""
	if b.botUser != nil {
		botUsername = strings.ToLower(b.botUser.Username)
	}

	lowerText := strings.ToLower(text)
	botMention := "@" + botUsername

	// Check if tagged with @username
	if botUsername != "" && strings.Contains(lowerText, botMention) {
		cleanText := strings.TrimSpace(strings.ReplaceAll(lowerText, botMention, ""))
		return true, cleanText
	}

	// Check if message is a reply to the bot
	if m.IsReply() && b.botUser != nil {
		if m.ReplySenderID() == b.botUser.ID {
			return true, text
		}
	}

	// Check if message mentions the bot's name ("kiara" or "kia")
	namePattern := regexp.MustCompile(`(?i)\b(kiara|kia)\b`)
	if namePattern.MatchString(text) {
		cleanText := strings.TrimSpace(namePattern.ReplaceAllString(text, ""))
		if cleanText == "" {
			cleanText = text
		}
		return true, cleanText
	}

	return false, ""
}

// HandleIncomingMessage is the global message listener
func (b *BotHandler) HandleIncomingMessage(m *telegram.NewMessage) error {
	should, prompt := b.ShouldRespond(m)
	if !should || prompt == "" {
		return nil
	}

	chatID := m.ChatID()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] in HandleIncomingMessage: %v", r)
			}
		}()

		// Send typing action to Telegram chat
		_, _ = b.client.SendAction(chatID, "typing")

		// Retrieve active model for this chat
		currentModel := b.memory.GetModel(chatID, b.cfg.DefaultModel)

		// Build conversation messages with system prompt & history
		messages := b.memory.GetMessagesWithSystem(chatID, b.GetSystemPrompt(), b.cfg.DefaultModel)
		messages = append(messages, ai.Message{Role: "user", Content: prompt})

		// Send periodic typing pulse if AI generation takes > 4 seconds
		stopTyping := make(chan struct{})
		go func() {
			ticker := time.NewTicker(4 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_, _ = b.client.SendAction(chatID, "typing")
				case <-stopTyping:
					return
				}
			}
		}()

		reply, err := b.ai.GenerateReply(currentModel, messages)
		close(stopTyping)

		if err != nil {
			log.Printf("[ERROR] AI generation failed for chat %d: %v", chatID, err)
			_, _ = m.Reply("ugh my internet lagged for a sec... say that again? 😭")
			return
		}

		cleanReply := strings.TrimSpace(reply)
		if cleanReply == "" {
			cleanReply = "..."
		}

		// Save turn into chat session memory
		b.memory.AddTurn(chatID, prompt, cleanReply, b.cfg.DefaultModel)

		// Send reply back to Telegram
		_, err = m.Reply(cleanReply)
		if err != nil {
			log.Printf("[ERROR] Failed to send reply to chat %d: %v", chatID, err)
		}
	}()

	return nil
}
