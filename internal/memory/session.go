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

package memory

import (
	"fmt"
	"sync"
	"time"

	"kiara-chatbot/internal/ai"
)

func SessionKey(chatID int64, userID int64, isGroup bool) string {
	if isGroup {
		return fmt.Sprintf("gc_%d_u_%d", chatID, userID)
	}
	return fmt.Sprintf("dm_%d", userID)
}

type Session struct {
	Key        string
	ChatID     int64
	UserID     int64
	Model      string
	ThinkLevel string
	History    []ai.Message
	LastActive time.Time
}

type Store struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	maxHistory int
	ttl        time.Duration
}

func NewStore(maxHistory int, ttl time.Duration) *Store {
	s := &Store{
		sessions:   make(map[string]*Session),
		maxHistory: maxHistory,
		ttl:        ttl,
	}
	go s.cleanupLoop()
	return s
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, sess := range s.sessions {
			if now.Sub(sess.LastActive) > s.ttl {
				delete(s.sessions, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) GetOrCreate(key string, defaultModel string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[key]
	if !exists {
		sess = &Session{
			Key:        key,
			Model:      defaultModel,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[key] = sess
	}
	sess.LastActive = time.Now()
	return sess
}

func (s *Store) GetModel(key string, defaultModel string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, exists := s.sessions[key]; exists && sess.Model != "" {
		return sess.Model
	}
	return defaultModel
}

func (s *Store) GetThink(key string, defaultThink string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, exists := s.sessions[key]; exists && sess.ThinkLevel != "" {
		return sess.ThinkLevel
	}
	return defaultThink
}

func (s *Store) SetThink(key string, thinkLevel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[key]
	if !exists {
		sess = &Session{
			Key:        key,
			ThinkLevel: thinkLevel,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[key] = sess
		return
	}
	sess.ThinkLevel = thinkLevel
	sess.LastActive = time.Now()
}

func (s *Store) SetModel(key string, model string, defaultModel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[key]
	if !exists {
		sess = &Session{
			Key:        key,
			Model:      model,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[key] = sess
		return
	}
	sess.Model = model
	sess.LastActive = time.Now()
}

func (s *Store) AddTurn(key string, userMsg, botMsg string, defaultModel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[key]
	if !exists {
		sess = &Session{
			Key:        key,
			Model:      defaultModel,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[key] = sess
	}

	sess.History = append(sess.History,
		ai.Message{Role: "user", Content: userMsg},
		ai.Message{Role: "assistant", Content: botMsg},
	)

	if len(sess.History) > s.maxHistory*2 {
		overflow := len(sess.History) - (s.maxHistory * 2)
		sess.History = sess.History[overflow:]
	}

	sess.LastActive = time.Now()
}

func (s *Store) GetMessagesWithSystem(key string, systemPrompt, defaultModel string) []ai.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := make([]ai.Message, 0)
	if systemPrompt != "" {
		msgs = append(msgs, ai.Message{Role: "system", Content: systemPrompt})
	}

	if sess, exists := s.sessions[key]; exists {
		msgs = append(msgs, sess.History...)
	}

	return msgs
}

func (s *Store) Clear(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, exists := s.sessions[key]; exists {
		sess.History = make([]ai.Message, 0)
		sess.LastActive = time.Now()
	}
}

func (s *Store) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
