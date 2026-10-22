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

func (p *Pool) Add(doc *telegram.DocumentObj, userID int64) bool {
	if doc == nil {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.stickers {
		if s.ID == doc.ID {
			return false
		}
	}

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

	if len(p.stickers) >= 200 {
		p.stickers = p.stickers[1:]
	}
	p.stickers = append(p.stickers, entry)
	p.save()
	return true
}

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

func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.stickers)
}
