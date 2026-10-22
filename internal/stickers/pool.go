package stickers

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type SavedSticker struct {
	ID            int64  `json:"id"`
	AccessHash    int64  `json:"access_hash"`
	FileReference []byte `json:"file_reference"`
	MimeType      string `json:"mime_type"`
	Emoji         string `json:"emoji,omitempty"`
	AddedBy       int64  `json:"added_by"`
	AddedAt       int64  `json:"added_at"`
}

type Pool struct {
	mu       sync.RWMutex
	filePath string
	stickers []SavedSticker
	rng      *rand.Rand
}

func NewPool(filePath string) *Pool {
	p := &Pool{
		filePath: filePath,
		stickers: make([]SavedSticker, 0),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	p.load()
	return p
}

func (p *Pool) load() {
	if p.filePath == "" {
		return
	}
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return
	}
	var loaded []SavedSticker
	if err := json.Unmarshal(data, &loaded); err == nil {
		p.stickers = loaded
	}
}

func (p *Pool) save() {
	if p.filePath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p.filePath), 0755)
	data, err := json.MarshalIndent(p.stickers, "", "  ")
	if err == nil {
		_ = os.WriteFile(p.filePath, data, 0644)
	}
}

// Add saves a new sticker to the collection if not duplicate
func (p *Pool) Add(doc *telegram.DocumentObj, userID int64) bool {
	if doc == nil {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already in pool
	for _, s := range p.stickers {
		if s.ID == doc.ID {
			return false
		}
	}

	// Extract emoji from sticker attributes if available
	emoji := ""
	for _, attr := range doc.Attributes {
		if st, ok := attr.(*telegram.DocumentAttributeSticker); ok {
			emoji = st.Alt
			break
		}
	}

	entry := SavedSticker{
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
		MimeType:      doc.MimeType,
		Emoji:         emoji,
		AddedBy:       userID,
		AddedAt:       time.Now().Unix(),
	}

	// Cap at 200 stickers
	if len(p.stickers) >= 200 {
		p.stickers = p.stickers[1:]
	}
	p.stickers = append(p.stickers, entry)
	p.save()
	return true
}

// GetRandom returns a random sticker as an InputMediaDocument ready to be sent
func (p *Pool) GetRandom() (*telegram.InputMediaDocument, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.stickers) == 0 {
		return nil, ""
	}

	idx := p.rng.Intn(len(p.stickers))
	s := p.stickers[idx]

	media := &telegram.InputMediaDocument{
		ID: &telegram.InputDocumentObj{
			ID:            s.ID,
			AccessHash:    s.AccessHash,
			FileReference: s.FileReference,
		},
	}
	return media, s.Emoji
}

// Count returns total stickers collected
func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.stickers)
}
