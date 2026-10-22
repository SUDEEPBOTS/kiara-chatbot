package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"kiara-chatbot/internal/ai"
	"kiara-chatbot/internal/config"
	"kiara-chatbot/internal/media"
	"kiara-chatbot/internal/memory"
	"kiara-chatbot/internal/ratelimit"
	"kiara-chatbot/internal/stickers"
)

type BotHandler struct {
	client       *telegram.Client
	ai           *ai.Client
	memory       *memory.Store
	cfg          *config.Config
	botUser      *telegram.UserObj
	systemPrompt string
	limiter      *ratelimit.LeakyBucket
	stickers     *stickers.Pool
	mu           sync.RWMutex
}

func NewBotHandler(
	client *telegram.Client,
	aiClient *ai.Client,
	memStore *memory.Store,
	cfg *config.Config,
	botUser *telegram.UserObj,
	limiter *ratelimit.LeakyBucket,
	stickersPool *stickers.Pool,
) *BotHandler {
	bh := &BotHandler{
		client:   client,
		ai:       aiClient,
		memory:   memStore,
		cfg:      cfg,
		botUser:  botUser,
		limiter:  limiter,
		stickers: stickersPool,
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

// FormatChatContext builds live runtime metadata about the chat & user
func FormatChatContext(m *telegram.NewMessage, isGroup bool) string {
	var sb strings.Builder
	sb.WriteString("\n\n[Runtime Chat & User Context:\n")
	if isGroup {
		sb.WriteString("- Chat Type: Group Chat\n")
		if m.Chat != nil {
			sb.WriteString(fmt.Sprintf("- Group Name: %s\n", m.Chat.Title))
			sb.WriteString(fmt.Sprintf("- Group Chat ID: %d\n", m.Chat.ID))
		} else if m.Channel != nil {
			sb.WriteString(fmt.Sprintf("- Group Name: %s\n", m.Channel.Title))
			sb.WriteString(fmt.Sprintf("- Group Chat ID: %d\n", m.Channel.ID))
			if m.Channel.Username != "" {
				sb.WriteString(fmt.Sprintf("- Group Username: @%s\n", m.Channel.Username))
			}
		} else {
			sb.WriteString(fmt.Sprintf("- Group Chat ID: %d\n", m.ChatID()))
		}
	} else {
		sb.WriteString("- Chat Type: Direct Message (DM/Private)\n")
		sb.WriteString(fmt.Sprintf("- Chat ID: %d\n", m.ChatID()))
	}

	if m.Sender != nil {
		name := strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName)
		if name != "" {
			sb.WriteString(fmt.Sprintf("- User Real Name: %s\n", name))
		}
		if m.Sender.Username != "" {
			sb.WriteString(fmt.Sprintf("- User Username: @%s\n", m.Sender.Username))
		}
		sb.WriteString(fmt.Sprintf("- User ID: %d\n", m.Sender.ID))
	} else {
		sb.WriteString(fmt.Sprintf("- User ID: %d\n", m.SenderID()))
	}

	sb.WriteString("Instruction: If the user asks about their own name, who they are, their user ID, or about the group name/details, use this actual data naturally in your response while maintaining your sassy Gen Z persona (do NOT sound like a robot or dump JSON).]\n")
	return sb.String()
}

// ShouldRespond determines if the message is directed at the bot.
// Returns: (shouldRespond, cleanPromptText, hasVisualMedia)
func (b *BotHandler) ShouldRespond(m *telegram.NewMessage) (bool, string, bool) {
	text := strings.TrimSpace(m.Text())
	hasMedia := (m.Photo() != nil || m.Sticker() != nil)

	// Ignore commands (they are processed by separate command handlers)
	if strings.HasPrefix(text, "/") {
		return false, "", false
	}

	// 1. Direct Message (DM) -> Always respond
	if m.IsPrivate() {
		return true, text, hasMedia
	}

	// 2. Group Chat (GC)
	botUsername := ""
	if b.botUser != nil {
		botUsername = strings.ToLower(b.botUser.Username)
	}

	lowerText := strings.ToLower(text)
	botMention := "@" + botUsername

	// A. Tagged via @BotUsername
	if botUsername != "" && strings.Contains(lowerText, botMention) {
		cleanText := strings.TrimSpace(strings.ReplaceAll(lowerText, botMention, ""))
		return true, cleanText, hasMedia
	}

	// B. Quote reply to bot's message
	if m.IsReply() && b.botUser != nil {
		if m.ReplySenderID() == b.botUser.ID {
			return true, text, hasMedia
		}
	}

	// C. Mentioned by name ("kiara" or "kia")
	namePattern := regexp.MustCompile(`(?i)\b(kiara|kia)\b`)
	if namePattern.MatchString(text) {
		cleanText := strings.TrimSpace(namePattern.ReplaceAllString(text, ""))
		if cleanText == "" {
			cleanText = text
		}
		return true, cleanText, hasMedia
	}

	return false, "", false
}

// HandleIncomingMessage is the master message listener
func (b *BotHandler) HandleIncomingMessage(m *telegram.NewMessage) error {
	// 1. Auto-collect incoming stickers into sticker pool
	if sticker := m.Sticker(); sticker != nil {
		b.stickers.Add(sticker, m.SenderID())
	}

	// 2. Anti-spam Leaky Bucket check for stickers in DM
	if m.IsPrivate() && m.Sticker() != nil {
		allowed, justBlocked, remSec := b.limiter.CheckSticker(m.SenderID())
		if !allowed {
			if justBlocked {
				_, _ = m.Reply(fmt.Sprintf("bro stop spamming stickers 🙄 baltti bhar gayi teri! blocked for %ds, shanti rakh.", remSec))
			}
			return nil
		}
	}

	should, prompt, hasMedia := b.ShouldRespond(m)
	if !should {
		return nil
	}

	chatID := m.ChatID()
	senderID := m.SenderID()
	isGroup := !m.IsPrivate()
	sessKey := memory.SessionKey(chatID, senderID, isGroup)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] in HandleIncomingMessage: %v", r)
			}
		}()

		// Start Telegram typing action
		_, _ = b.client.SendAction(chatID, "typing")

		// Periodic typing pulse while awaiting AI response
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
		defer close(stopTyping)

		// 3. Process Visual Media if present (Photo or Sticker frame)
		var procMedia *media.ProcessedMedia
		if hasMedia {
			var err error
			procMedia, err = media.ProcessMessageMedia(m)
			if err != nil {
				log.Printf("[WARN] Media processing failed for chat %d: %v", chatID, err)
			}
		}

		// 4. Retrieve active model & settings for this specific user session
		currentModel := b.memory.GetModel(sessKey, b.cfg.DefaultModel)

		// Base system prompt with live runtime chat and user info
		systemWithContext := b.GetSystemPrompt() + FormatChatContext(m, isGroup)

		// Build conversation messages
		messages := b.memory.GetMessagesWithSystem(sessKey, systemWithContext, b.cfg.DefaultModel)

		var turnUserMsg ai.Message
		var memoryUserSummary string

		if procMedia != nil && procMedia.Base64DataURL != "" {
			queryText := prompt
			if queryText == "" {
				if procMedia.MediaType == "photo" {
					queryText = "Here is a photo I sent you. Look at it and give your thoughts or reaction."
				} else {
					emojiHint := ""
					if procMedia.Emoji != "" {
						emojiHint = fmt.Sprintf(" (Sticker emoji: %s)", procMedia.Emoji)
					}
					queryText = fmt.Sprintf("Here is a sticker%s I sent. React to it authentically. If you feel like sending a sticker back from your collection, include [SEND_STICKER] in your reply.", emojiHint)
				}
			}

			turnUserMsg = ai.Message{
				Role: "user",
				Content: []ai.ContentPart{
					{Type: "text", Text: queryText},
					{Type: "image_url", ImageURL: &ai.ImageURLField{URL: procMedia.Base64DataURL}},
				},
			}
			memoryUserSummary = fmt.Sprintf("[%s: %s]", procMedia.MediaType, queryText)
		} else {
			turnUserMsg = ai.Message{Role: "user", Content: prompt}
			memoryUserSummary = prompt
		}

		messages = append(messages, turnUserMsg)

		// 5. Generate AI reply
		reply, err := b.ai.GenerateReply(currentModel, messages)
		if err != nil {
			log.Printf("[ERROR] AI generation failed for session %s: %v", sessKey, err)
			_, _ = m.Reply("ugh my brain lagged for a sec... say that again? 😭")
			return
		}

		cleanReply := strings.TrimSpace(reply)
		if cleanReply == "" {
			cleanReply = "..."
		}

		// 6. Sticker Reply Logic
		shouldSendSticker := strings.Contains(cleanReply, "[SEND_STICKER]")
		cleanReply = strings.TrimSpace(strings.ReplaceAll(cleanReply, "[SEND_STICKER]", ""))

		lowerPrompt := strings.ToLower(prompt)
		if strings.Contains(lowerPrompt, "sticker bhej") || strings.Contains(lowerPrompt, "send sticker") || strings.Contains(lowerPrompt, "send a sticker") {
			shouldSendSticker = true
		}

		// 25% chance to send a sticker back if user sent a sticker
		if !shouldSendSticker && procMedia != nil && strings.HasPrefix(procMedia.MediaType, "sticker") {
			if rand.Float64() < 0.25 {
				shouldSendSticker = true
			}
		}

		if shouldSendSticker {
			if stkMedia, _ := b.stickers.GetRandom(); stkMedia != nil {
				_, _ = m.ReplyMedia(stkMedia)
			}
		}

		// Save turn into session memory
		b.memory.AddTurn(sessKey, memoryUserSummary, cleanReply, b.cfg.DefaultModel)

		// Send text reply if non-empty
		if cleanReply != "" {
			_, err = m.Reply(cleanReply)
			if err != nil {
				log.Printf("[ERROR] Failed to send reply to %s: %v", sessKey, err)
			}
		}
	}()

	return nil
}
