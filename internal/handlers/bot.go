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
	"kiara-chatbot/internal/moderation"
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
	modFilter    *moderation.Filter
	botMsgCache  map[string]time.Time
	activeUsers  map[string]time.Time
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
	modFilter *moderation.Filter,
) *BotHandler {
	bh := &BotHandler{
		client:      client,
		ai:          aiClient,
		memory:      memStore,
		cfg:         cfg,
		botUser:     botUser,
		limiter:     limiter,
		stickers:    stickersPool,
		modFilter:   modFilter,
		botMsgCache: make(map[string]time.Time),
		activeUsers: make(map[string]time.Time),
	}
	bh.ReloadSystemPrompt()
	return bh
}

func (b *BotHandler) cacheBotMsg(chatID int64, msgID int32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.botMsgCache == nil {
		b.botMsgCache = make(map[string]time.Time)
	}
	key := fmt.Sprintf("%d:%d", chatID, msgID)
	b.botMsgCache[key] = time.Now()

	if len(b.botMsgCache) > 2000 {
		cutoff := time.Now().Add(-2 * time.Hour)
		for k, t := range b.botMsgCache {
			if t.Before(cutoff) {
				delete(b.botMsgCache, k)
			}
		}
	}
}

func (b *BotHandler) isBotMsg(chatID int64, msgID int32) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.botMsgCache == nil {
		return false
	}
	key := fmt.Sprintf("%d:%d", chatID, msgID)
	_, ok := b.botMsgCache[key]
	return ok
}

func (b *BotHandler) recordUserInteraction(chatID, senderID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.activeUsers == nil {
		b.activeUsers = make(map[string]time.Time)
	}
	key := fmt.Sprintf("%d:%d", chatID, senderID)
	b.activeUsers[key] = time.Now()
}

func (b *BotHandler) isUserInActiveWindow(chatID, senderID int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.activeUsers == nil {
		return false
	}
	key := fmt.Sprintf("%d:%d", chatID, senderID)
	lastTime, ok := b.activeUsers[key]
	if !ok {
		return false
	}
	return time.Since(lastTime) <= 60*time.Second
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

func FormatChatContext(m *telegram.NewMessage, isGroup bool) string {
	var sb strings.Builder
	sb.WriteString("\n\n[Runtime Chat & User Context:\n")
	if isGroup {
		sb.WriteString("- Chat Type: Group Chat\n")

		groupName := ""
		if m.Channel != nil && m.Channel.Title != "" {
			groupName = m.Channel.Title
		} else if ch, err := m.GetChannel(); err == nil && ch != nil && ch.Title != "" {
			groupName = ch.Title
		}

		if groupName == "" {
			if m.Chat != nil && m.Chat.ID != m.SenderID() && m.Chat.Title != "" {
				groupName = m.Chat.Title
			} else if chat, err := m.GetChat(); err == nil && chat != nil && chat.ID != m.SenderID() && chat.Title != "" {
				groupName = chat.Title
			}
		}

		if groupName != "" {
			sb.WriteString(fmt.Sprintf("- Group Name: %s\n", groupName))
		}
		sb.WriteString(fmt.Sprintf("- Group Chat ID: %d\n", m.ChatID()))
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

	sb.WriteString("Instruction: Group Name is the title of the Telegram group chat. User Real Name is the name of the user talking to you. Do NEVER confuse the user name with the group name! If asked about group name, only use Group Name. If asked about user name, only use User Real Name.]\n")
	return sb.String()
}

var (
	reTrailingDot = regexp.MustCompile(`\.+(?:\s+|$)`)
	reMultiSpace  = regexp.MustCompile(`[ \t]+`)
)

func SanitizePunctuation(text string) string {
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "!", "")
	text = reTrailingDot.ReplaceAllString(text, " ")
	text = reMultiSpace.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func ExtractImagePrompt(text string) (bool, string) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(.+?)\s*(?:ki|ka)?\s*(?:image|photo|pic)\s+(?:bana\s*de|bana\s*do|banao|bana)$`),
		regexp.MustCompile(`(?i)(?:image|photo|pic|picture)\s+(?:bana\s*de|bana\s*do|banao|bana)\s*[:\-\s]*(.+)`),
		regexp.MustCompile(`(?i)(?:draw|generate\s+(?:an?\s+)?image|create\s+(?:an?\s+)?image|paint)\s*(?:of|a|an|me|for\s*me)?\s*[:\-\s]*(.+)`),
	}

	for _, pat := range patterns {
		matches := pat.FindStringSubmatch(text)
		if len(matches) > 1 {
			res := strings.TrimSpace(matches[1])
			res = strings.TrimPrefix(res, ":")
			res = strings.TrimPrefix(res, "-")
			res = strings.TrimSpace(res)
			if len(res) >= 3 {
				return true, res
			}
		}
	}
	return false, ""
}

func (b *BotHandler) ShouldRespond(m *telegram.NewMessage) (bool, string, bool) {
	text := strings.TrimSpace(m.Text())
	hasMedia := (m.Photo() != nil || m.Sticker() != nil)

	if strings.HasPrefix(text, "/") {
		return false, "", false
	}

	if m.IsPrivate() {
		return true, text, hasMedia
	}

	isGroup := !m.IsPrivate()

	isReplyToBot := false
	isReplyToSomeoneElse := false

	if m.IsReply() && b.botUser != nil {
		if m.ReplySenderID() == b.botUser.ID {
			isReplyToBot = true
		} else if b.isBotMsg(m.ChatID(), m.ReplyToMsgID()) {
			isReplyToBot = true
		} else if replyMsg, err := m.GetReplyMessage(); err == nil && replyMsg != nil {
			if replyMsg.SenderID() == b.botUser.ID {
				b.cacheBotMsg(m.ChatID(), replyMsg.ID)
				isReplyToBot = true
			} else {
				isReplyToSomeoneElse = true
			}
		} else {
			isReplyToSomeoneElse = true
		}
	}

	if isReplyToBot {
		return true, text, hasMedia
	}

	botUsername := ""
	if b.botUser != nil {
		botUsername = strings.ToLower(b.botUser.Username)
	}

	lowerText := strings.ToLower(text)
	botMention := "@" + botUsername

	hasMention := (botUsername != "" && strings.Contains(lowerText, botMention))
	namePattern := regexp.MustCompile(`(?i)\b(kiara|kaira|kira|kia|kyara|kyaara|kiarra|kiyara|kayra|keera)\b`)
	hasName := namePattern.MatchString(text)

	if hasMention {
		cleanText := strings.TrimSpace(strings.ReplaceAll(lowerText, botMention, ""))
		return true, cleanText, hasMedia
	}

	if hasName {
		cleanText := strings.TrimSpace(namePattern.ReplaceAllString(text, ""))
		if cleanText == "" {
			cleanText = text
		}
		return true, cleanText, hasMedia
	}

	if isReplyToSomeoneElse {
		return false, "", false
	}

	if m.Sticker() != nil {
		return false, "", false
	}

	if isGroup && b.isUserInActiveWindow(m.ChatID(), m.SenderID()) {
		return true, text, hasMedia
	}

	return false, "", false
}

func (b *BotHandler) HandleIncomingMessage(m *telegram.NewMessage) error {
	if b.modFilter != nil && b.modFilter.ModerateMessage(m) {
		return nil
	}

	if sticker := m.Sticker(); sticker != nil {
		b.stickers.Add(sticker, m.SenderID())
	}

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

	b.recordUserInteraction(chatID, senderID)
	log.Printf("[MESSAGE] Chat: %d | Sender: %d | IsGroup: %v | Prompt: %q", chatID, senderID, isGroup, prompt)

	if !hasMedia {
		isImageReq, imgPrompt := ExtractImagePrompt(prompt)
		if isImageReq && imgPrompt != "" {
			go func() {
				_, _ = b.client.SendAction(chatID, "upload_photo")
				statusMsg, _ := m.Reply("painting your idea... give me a sec 🎨")

				imgURL, err := b.ai.GenerateImage(imgPrompt)
				if err != nil {
					if statusMsg != nil {
						_, _ = statusMsg.Edit(fmt.Sprintf("ugh couldn't draw that: %v 😭", err))
					}
					return
				}

				caption := fmt.Sprintf("✨ <b>Prompt:</b> <i>%s</i>\n🎨 <i>Generated by Kiara</i>", imgPrompt)
				_, err = m.ReplyMedia(imgURL, &telegram.MediaOptions{Caption: caption})
				if err == nil && statusMsg != nil {
					_, _ = statusMsg.Delete()
				}
			}()
			return nil
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] in HandleIncomingMessage: %v", r)
			}
		}()

		_, _ = b.client.SendAction(chatID, "typing")

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

		var procMedia *media.ProcessedMedia
		if hasMedia {
			var err error
			procMedia, err = media.ProcessMessageMedia(m)
			if err != nil {
				log.Printf("[WARN] Media processing failed for chat %d: %v", chatID, err)
			}
		}

		currentModel := b.memory.GetModel(sessKey, b.cfg.DefaultModel)
		systemWithContext := b.GetSystemPrompt() + FormatChatContext(m, isGroup)
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

		shouldSendSticker := strings.Contains(cleanReply, "[SEND_STICKER]")
		cleanReply = strings.TrimSpace(strings.ReplaceAll(cleanReply, "[SEND_STICKER]", ""))
		cleanReply = SanitizePunctuation(cleanReply)

		lowerPrompt := strings.ToLower(prompt)
		if strings.Contains(lowerPrompt, "sticker bhej") || strings.Contains(lowerPrompt, "send sticker") || strings.Contains(lowerPrompt, "send a sticker") {
			shouldSendSticker = true
		}

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

		b.memory.AddTurn(sessKey, memoryUserSummary, cleanReply, b.cfg.DefaultModel)

		if cleanReply != "" {
			sentMsg, err := m.Reply(cleanReply)
			if err != nil {
				log.Printf("[ERROR] Failed to send reply to %s: %v", sessKey, err)
			} else if sentMsg != nil {
				b.cacheBotMsg(chatID, sentMsg.ID)
				log.Printf("[REPLY] Chat: %d | MsgID: %d | Text: %q", chatID, sentMsg.ID, cleanReply)
			}
		}
	}()

	return nil
}
