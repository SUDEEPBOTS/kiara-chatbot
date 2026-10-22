package handlers

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

var AvailableModels = map[string]string{
	"argon-4":           "Google Gemini 4 Argon (Flagship Frontier Reasoning)",
	"gemini-pro":        "Google Gemini 3.1 Pro (Deep logic & analysis)",
	"gemini-2.0-flash":  "Google Gemini 2.0 Flash (Fast & reliable)",
	"gemini-flash-lite": "Google Gemini Flash-Lite (Lightweight & fastest)",
	"deepseek-chat":     "DeepSeek V3 (Natural & creative dialogue)",
	"deepseek-reasoner": "DeepSeek R1 (Deep reasoning model)",
	"gpt-4o":            "OpenAI GPT-4o (Frontier multimodal)",
	"gpt-4o-mini":       "OpenAI GPT-4o Mini (Fast conversational)",
	"claude-3-5-sonnet": "Anthropic Claude 3.5 Sonnet",
	"grok-3":            "xAI Grok 3",
}

func (b *BotHandler) HandleStart(m *telegram.NewMessage) error {
	reply := "heyy, what's up? i'm kiara ✨\n\nyou can text me directly in dm, or tag me/reply to me in groups.\n\ncommands:\n• /models - see all models\n• /model <name> - switch ai model\n• /clear - reset conversation\n• /status - view active model"
	_, err := m.Reply(reply)
	return err
}

func (b *BotHandler) HandleHelp(m *telegram.NewMessage) error {
	reply := `<b>✨ Kiara Chatbot Commands ✨</b>

• <b>/models</b> — View all available AI models
• <b>/model &lt;name&gt;</b> — Switch model for this chat (e.g. <code>/model argon-4</code>)
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

func (b *BotHandler) HandleModels(m *telegram.NewMessage) error {
	chatID := m.ChatID()
	currentModel := b.memory.GetModel(chatID, b.cfg.DefaultModel)

	var sb strings.Builder
	sb.WriteString("<b>Available AI Models:</b>\n\n")

	for name, desc := range AvailableModels {
		activeTag := ""
		if strings.HasPrefix(currentModel, name) {
			activeTag = " <b>[ACTIVE]</b>"
		}
		sb.WriteString(fmt.Sprintf("• <code>%s</code>%s\n  <i>%s</i>\n", name, activeTag, desc))
	}

	sb.WriteString("\n<b>To switch:</b> <code>/model &lt;model_name&gt;</code>\nExample: <code>/model argon-4</code>")

	_, err := m.Reply(sb.String())
	return err
}

func (b *BotHandler) HandleModel(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		current := b.memory.GetModel(m.ChatID(), b.cfg.DefaultModel)
		_, err := m.Reply(fmt.Sprintf("Current active model: <code>%s</code>\n\nUse <code>/model &lt;name&gt;</code> to change.\nType <code>/models</code> to see the full list.", current))
		return err
	}

	targetModel := strings.ToLower(args)

	// If it's an argon/gemini model and doesn't specify thinking, add @think=low by default for speed
	if (strings.Contains(targetModel, "argon") || strings.Contains(targetModel, "gemini")) && !strings.Contains(targetModel, "@think=") {
		targetModel = targetModel + "@think=low"
	}

	b.memory.SetModel(m.ChatID(), targetModel, b.cfg.DefaultModel)
	_, err := m.Reply(fmt.Sprintf("✔ Switched model to <code>%s</code> for this chat!", targetModel))
	return err
}

func (b *BotHandler) HandleClear(m *telegram.NewMessage) error {
	b.memory.Clear(m.ChatID())
	_, err := m.Reply("chat memory cleared! clean slate. what's up? ✨")
	return err
}

func (b *BotHandler) HandleStatus(m *telegram.NewMessage) error {
	chatID := m.ChatID()
	currentModel := b.memory.GetModel(chatID, b.cfg.DefaultModel)

	status := fmt.Sprintf(`<b>Chat Status:</b>
• <b>Chat ID:</b> <code>%d</code>
• <b>Active Model:</b> <code>%s</code>
• <b>Total Active Chats:</b> <code>%d</code>
• <b>Persona:</b> Kiara (20, Bandra, Mumbai)`,
		chatID, currentModel, b.memory.ActiveCount())

	_, err := m.Reply(status)
	return err
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

	_, err := m.Reply("✔ System prompt reloaded successfully from disk!")
	return err
}
