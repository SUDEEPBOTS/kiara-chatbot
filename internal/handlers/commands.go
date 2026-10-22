package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"kiara-chatbot/internal/memory"
)

type ModelItem struct {
	ID   string
	Name string
	Desc string
}

var ModelCatalog = []ModelItem{
	{ID: "argon-4", Name: "Gemini 4 Argon", Desc: "Flagship Frontier Reasoning"},
	{ID: "gemini-pro", Name: "Gemini 3.1 Pro", Desc: "Advanced Pro & Logic"},
	{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Desc: "Fast & Balanced"},
	{ID: "gemini-flash-lite", Name: "Gemini Flash-Lite", Desc: "Lightweight & Fastest"},
	{ID: "deepseek-chat", Name: "DeepSeek V3", Desc: "Smart Dialogue & Natural"},
	{ID: "deepseek-reasoner", Name: "DeepSeek R1", Desc: "Deep Reasoning & Logic"},
	{ID: "gpt-4o", Name: "OpenAI GPT-4o", Desc: "Frontier Multimodal"},
	{ID: "chatgpt-4o", Name: "ChatGPT 4o", Desc: "ChatGPT Web Tier"},
	{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet", Desc: "Top Coding & Nuance"},
	{ID: "grok-3", Name: "xAI Grok 3", Desc: "xAI Frontier Model"},
	{ID: "gemini-3.7-flash", Name: "Gemini 3.7 Flash", Desc: "Latest Flash Model"},
	{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Desc: "Fast Conversational"},
}

const PageSize = 4

// RenderModelsPage generates the text and inline keyboard for a specific page and session
func (b *BotHandler) RenderModelsPage(sessKey string, page int) (string, *telegram.ReplyInlineMarkup) {
	totalModels := len(ModelCatalog)
	totalPages := (totalModels + PageSize - 1) / PageSize

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	currentModel := b.memory.GetModel(sessKey, b.cfg.DefaultModel)
	currentThink := b.memory.GetThink(sessKey, "low")

	start := page * PageSize
	end := start + PageSize
	if end > totalModels {
		end = totalModels
	}

	var sb strings.Builder
	sb.WriteString("<b>✨ Select AI Model for this Chat ✨</b>\n\n")
	sb.WriteString(fmt.Sprintf("• <b>Active Model:</b> <code>%s</code>\n", currentModel))
	sb.WriteString(fmt.Sprintf("• <b>Thinking Depth:</b> <code>%s</code>\n", strings.ToUpper(currentThink)))
	sb.WriteString(fmt.Sprintf("• <b>Page:</b> <code>%d/%d</code>\n\n", page+1, totalPages))
	sb.WriteString("<i>Tap a button below to switch model instantly:</i>")

	kb := telegram.NewKeyboard()

	// 1. Model buttons (2 per row for clean mobile layout)
	pageModels := ModelCatalog[start:end]
	for i := 0; i < len(pageModels); i += 2 {
		var row []telegram.KeyboardButton

		m1 := pageModels[i]
		m1Label := m1.Name
		if strings.HasPrefix(currentModel, m1.ID) {
			m1Label = "✅ " + m1.Name
		}
		row = append(row, telegram.Button.Data(m1Label, fmt.Sprintf("setm:%s:%d", m1.ID, page)))

		if i+1 < len(pageModels) {
			m2 := pageModels[i+1]
			m2Label := m2.Name
			if strings.HasPrefix(currentModel, m2.ID) {
				m2Label = "✅ " + m2.Name
			}
			row = append(row, telegram.Button.Data(m2Label, fmt.Sprintf("setm:%s:%d", m2.ID, page)))
		}

		kb.AddRow(row...)
	}

	// 2. Navigation row (Prev, Page counter, Next)
	var navRow []telegram.KeyboardButton
	if page > 0 {
		navRow = append(navRow, telegram.Button.Data("◀️ Prev", fmt.Sprintf("page:%d", page-1)))
	} else {
		navRow = append(navRow, telegram.Button.Data("•", "noop"))
	}

	navRow = append(navRow, telegram.Button.Data(fmt.Sprintf("%d / %d", page+1, totalPages), "noop"))

	if page < totalPages-1 {
		navRow = append(navRow, telegram.Button.Data("Next ▶️", fmt.Sprintf("page:%d", page+1)))
	} else {
		navRow = append(navRow, telegram.Button.Data("•", "noop"))
	}
	kb.AddRow(navRow...)

	// 3. Settings & Controls row (Thinking toggle & Close)
	thinkLabel := fmt.Sprintf("⚡ Think: %s", strings.ToUpper(currentThink))
	kb.AddRow(
		telegram.Button.Data(thinkLabel, fmt.Sprintf("think:%d", page)),
		telegram.Button.Data("❌ Close", "close"),
	)

	return sb.String(), kb.Build()
}

func (b *BotHandler) HandleStart(m *telegram.NewMessage) error {
	reply := "heyy, what's up? i'm kiara ✨\n\nyou can text me directly in dm, or tag me/reply to me in groups.\n\ncommands:\n• /models - interactive model selector\n• /model <name> - switch model by name\n• /imagine <prompt> - generate ai images\n• /clear - reset chat memory\n• /status - view active settings"
	_, err := m.Reply(reply)
	return err
}

func (b *BotHandler) HandleHelp(m *telegram.NewMessage) error {
	reply := `<b>✨ Kiara Chatbot Commands ✨</b>

• <b>/models</b> — Interactive model menu with inline buttons & pagination
• <b>/model &lt;name&gt;</b> — Switch model directly (e.g. <code>/model argon-4</code>)
• <b>/imagine &lt;prompt&gt;</b> — Generate images via AI
• <b>/clear</b> — Reset memory for this chat
• <b>/status</b> — Check active model & chat state
• <b>/help</b> — Show this help message

<b>Group Chat Triggers:</b>
1. Mention: <code>@` + b.botUser.Username + `</code>
2. Reply: Reply to any message sent by me
3. Name: Include "kiara" or "kia" in your message`

	_, err := m.Reply(reply)
	return err
}

// HandleModels opens the paginated interactive inline keyboard
func (b *BotHandler) HandleModels(m *telegram.NewMessage) error {
	sessKey := memory.SessionKey(m.ChatID(), m.SenderID(), !m.IsPrivate())
	text, markup := b.RenderModelsPage(sessKey, 0)
	_, err := m.Reply(text, &telegram.SendOptions{ReplyMarkup: markup})
	return err
}

// HandleModel allows direct switching via command: /model argon-4
func (b *BotHandler) HandleModel(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		return b.HandleModels(m)
	}

	sessKey := memory.SessionKey(m.ChatID(), m.SenderID(), !m.IsPrivate())
	targetModel := strings.ToLower(args)
	think := b.memory.GetThink(sessKey, "low")

	// If it's an argon/gemini model and doesn't specify thinking, append current think level
	if (strings.Contains(targetModel, "argon") || strings.Contains(targetModel, "gemini")) && !strings.Contains(targetModel, "@think=") {
		targetModel = fmt.Sprintf("%s@think=%s", targetModel, think)
	}

	b.memory.SetModel(sessKey, targetModel, b.cfg.DefaultModel)
	_, err := m.Reply(fmt.Sprintf("✔ Switched model to <code>%s</code> for your session!", targetModel))
	return err
}

// HandleCallback processes button clicks from inline keyboards
func (b *BotHandler) HandleCallback(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	sessKey := memory.SessionKey(cb.GetChatID(), cb.GetSenderID(), !cb.IsPrivate())

	if data == "noop" {
		_, err := cb.Answer("")
		return err
	}

	if data == "close" {
		_, _ = cb.Answer("Closed!")
		_, err := cb.Delete()
		return err
	}

	// 1. Page navigation: "page:<num>"
	if strings.HasPrefix(data, "page:") {
		pageNum, err := strconv.Atoi(strings.TrimPrefix(data, "page:"))
		if err == nil {
			text, markup := b.RenderModelsPage(sessKey, pageNum)
			_, _ = cb.Edit(text, &telegram.SendOptions{ReplyMarkup: markup})
		}
		_, err = cb.Answer("")
		return err
	}

	// 2. Model selection: "setm:<model_id>:<page>"
	if strings.HasPrefix(data, "setm:") {
		parts := strings.Split(data, ":")
		if len(parts) >= 3 {
			modelID := parts[1]
			pageNum, _ := strconv.Atoi(parts[2])

			think := b.memory.GetThink(sessKey, "low")
			fullModel := modelID
			if (strings.Contains(modelID, "argon") || strings.Contains(modelID, "gemini")) && !strings.Contains(modelID, "@think=") {
				fullModel = fmt.Sprintf("%s@think=%s", modelID, think)
			}

			b.memory.SetModel(sessKey, fullModel, b.cfg.DefaultModel)
			_, _ = cb.Answer(fmt.Sprintf("✔ Model switched to: %s", modelID))

			text, markup := b.RenderModelsPage(sessKey, pageNum)
			_, err := cb.Edit(text, &telegram.SendOptions{ReplyMarkup: markup})
			return err
		}
	}

	// 3. Thinking depth cycle: "think:<page>"
	if strings.HasPrefix(data, "think:") {
		pageNum, _ := strconv.Atoi(strings.TrimPrefix(data, "think:"))
		currentThink := b.memory.GetThink(sessKey, "low")

		// Cycle: low -> medium -> high -> off -> low
		nextThink := "medium"
		switch currentThink {
		case "low":
			nextThink = "medium"
		case "medium":
			nextThink = "high"
		case "high":
			nextThink = "off"
		case "off":
			nextThink = "low"
		default:
			nextThink = "low"
		}

		b.memory.SetThink(sessKey, nextThink)

		// Also update current model string if it's a gemini/argon model
		curModel := b.memory.GetModel(sessKey, b.cfg.DefaultModel)
		if idx := strings.Index(curModel, "@think="); idx != -1 {
			curModel = curModel[:idx] + "@think=" + nextThink
			b.memory.SetModel(sessKey, curModel, b.cfg.DefaultModel)
		}

		_, _ = cb.Answer(fmt.Sprintf("⚡ Thinking depth: %s", strings.ToUpper(nextThink)))

		text, markup := b.RenderModelsPage(sessKey, pageNum)
		_, err := cb.Edit(text, &telegram.SendOptions{ReplyMarkup: markup})
		return err
	}

	return nil
}

// HandleImagine generates an image from prompt using unified-web2api
func (b *BotHandler) HandleImagine(m *telegram.NewMessage) error {
	prompt := strings.TrimSpace(m.Args())
	if prompt == "" {
		_, err := m.Reply("give me a prompt to draw! e.g.:\n<code>/imagine a cute cat drinking iced coffee in mumbai</code>")
		return err
	}

	_, _ = b.client.SendAction(m.ChatID(), "upload_photo")
	statusMsg, _ := m.Reply("painting your idea... give me a sec 🎨")

	go func() {
		imgURL, err := b.ai.GenerateImage(prompt)
		if err != nil {
			if statusMsg != nil {
				_, _ = statusMsg.Edit(fmt.Sprintf("ugh couldn't draw that: %v 😭", err))
			}
			return
		}

		caption := fmt.Sprintf("✨ <b>Prompt:</b> <i>%s</i>\n🎨 <i>Generated by Kiara</i>", prompt)
		_, err = m.ReplyMedia(imgURL, &telegram.MediaOptions{Caption: caption})
		if err == nil && statusMsg != nil {
			_, _ = statusMsg.Delete()
		}
	}()

	return nil
}

func (b *BotHandler) HandleClear(m *telegram.NewMessage) error {
	sessKey := memory.SessionKey(m.ChatID(), m.SenderID(), !m.IsPrivate())
	b.memory.Clear(sessKey)
	_, err := m.Reply("chat memory cleared! clean slate. what's up? ✨")
	return err
}

func (b *BotHandler) HandleStatus(m *telegram.NewMessage) error {
	sessKey := memory.SessionKey(m.ChatID(), m.SenderID(), !m.IsPrivate())
	currentModel := b.memory.GetModel(sessKey, b.cfg.DefaultModel)
	currentThink := b.memory.GetThink(sessKey, "low")

	status := fmt.Sprintf(`<b>✨ Kiara Chatbot Status ✨</b>
• <b>Chat ID:</b> <code>%d</code>
• <b>User ID:</b> <code>%d</code>
• <b>Active Model:</b> <code>%s</code>
• <b>Thinking Depth:</b> <code>%s</code>
• <b>Active Sessions:</b> <code>%d</code>
• <b>Collected Stickers:</b> <code>%d</code>
• <b>Persona:</b> Kiara (20, Bandra, Mumbai)`,
		m.ChatID(), m.SenderID(), currentModel, strings.ToUpper(currentThink), b.memory.ActiveCount(), b.stickers.Count())

	_, err := m.Reply(status)
	return err
}

// IsAdmin checks if sender is group administrator or bot owner
func (b *BotHandler) IsAdmin(chatID, userID int64) bool {
	if b.cfg.OwnerID != 0 && userID == b.cfg.OwnerID {
		return true
	}
	p, err := b.client.GetChatMember(chatID, userID)
	if err == nil && p != nil {
		status := strings.ToLower(p.Status)
		if status == "creator" || status == "admin" || status == "administrator" {
			return true
		}
		if p.Rights != nil {
			return true
		}
	}
	return false
}

// HandleAbuse toggles auto-moderation on or off for this group
func (b *BotHandler) HandleAbuse(m *telegram.NewMessage) error {
	if m.IsPrivate() {
		_, err := m.Reply("Auto-moderation is only active in group chats! In DMs you can just text normally ✨")
		return err
	}

	if !b.IsAdmin(m.ChatID(), m.SenderID()) {
		_, err := m.Reply("you need to be an admin in this group to toggle moderation 🙄")
		return err
	}

	arg := strings.ToLower(strings.TrimSpace(m.Args()))
	switch arg {
	case "on", "enable", "start":
		b.modFilter.SetEnabled(m.ChatID(), true)
		_, err := m.Reply("🛡️ <b>Auto-Moderation is now ENABLED for this group!</b>\nAbusive messages will be automatically deleted with self-destructing warnings.")
		return err

	case "off", "disable", "stop":
		b.modFilter.SetEnabled(m.ChatID(), false)
		_, err := m.Reply("⚠️ <b>Auto-Moderation is now DISABLED for this group.</b>")
		return err

	default:
		currentStatus := "ENABLED"
		if !b.modFilter.IsEnabled(m.ChatID()) {
			currentStatus = "DISABLED"
		}
		reply := fmt.Sprintf("🛡️ <b>Auto-Moderation Status:</b> <code>%s</code> (Loaded Words: <code>%d</code>)\n\nUsage:\n• <code>/abuse on</code> — Enable auto-moderation\n• <code>/abuse off</code> — Disable auto-moderation", currentStatus, b.modFilter.Count())
		_, err := m.Reply(reply)
		return err
	}
}

func (b *BotHandler) HandleReload(m *telegram.NewMessage) error {
	if b.cfg.OwnerID != 0 && m.SenderID() != b.cfg.OwnerID {
		_, err := m.Reply("you don't have permission to do this 🙄")
		return err
	}

	if err := b.ReloadSystemPrompt(); err != nil {
		_, err = m.Reply(fmt.Sprintf("Failed to reload system prompt: %v", err))
		return err
	}

	if b.modFilter != nil {
		_ = b.modFilter.Reload()
	}

	wordCount := 0
	if b.modFilter != nil {
		wordCount = b.modFilter.Count()
	}

	_, err := m.Reply(fmt.Sprintf("✔ System prompt & remote blocklist reloaded! (Active Words: <code>%d</code>)", wordCount))
	return err
}
