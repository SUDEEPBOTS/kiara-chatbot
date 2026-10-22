package memory

import (
	"sync"
	"time"

	"kiara-chatbot/internal/ai"
)

type Session struct {
	ChatID     int64
	Model      string
	History    []ai.Message
	LastActive time.Time
}

type Store struct {
	mu         sync.RWMutex
	sessions   map[int64]*Session
	maxHistory int
	ttl        time.Duration
}

func NewStore(maxHistory int, ttl time.Duration) *Store {
	s := &Store{
		sessions:   make(map[int64]*Session),
		maxHistory: maxHistory,
		ttl:        ttl,
	}

	// Periodic cleanup of stale sessions
	go s.cleanupLoop()

	return s
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, sess := range s.sessions {
			if now.Sub(sess.LastActive) > s.ttl {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) GetOrCreate(chatID int64, defaultModel string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[chatID]
	if !exists {
		sess = &Session{
			ChatID:     chatID,
			Model:      defaultModel,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[chatID] = sess
	}
	sess.LastActive = time.Now()
	return sess
}

func (s *Store) GetModel(chatID int64, defaultModel string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, exists := s.sessions[chatID]; exists && sess.Model != "" {
		return sess.Model
	}
	return defaultModel
}

func (s *Store) SetModel(chatID int64, model string, defaultModel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[chatID]
	if !exists {
		sess = &Session{
			ChatID:     chatID,
			Model:      model,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[chatID] = sess
		return
	}
	sess.Model = model
	sess.LastActive = time.Now()
}

func (s *Store) AddTurn(chatID int64, userMsg, botMsg string, defaultModel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[chatID]
	if !exists {
		sess = &Session{
			ChatID:     chatID,
			Model:      defaultModel,
			History:    make([]ai.Message, 0),
			LastActive: time.Now(),
		}
		s.sessions[chatID] = sess
	}

	sess.History = append(sess.History,
		ai.Message{Role: "user", Content: userMsg},
		ai.Message{Role: "assistant", Content: botMsg},
	)

	// Prune older history to stay within maxHistory turns
	if len(sess.History) > s.maxHistory*2 {
		overflow := len(sess.History) - (s.maxHistory * 2)
		sess.History = sess.History[overflow:]
	}

	sess.LastActive = time.Now()
}

func (s *Store) GetMessagesWithSystem(chatID int64, systemPrompt, defaultModel string) []ai.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := make([]ai.Message, 0)
	if systemPrompt != "" {
		msgs = append(msgs, ai.Message{Role: "system", Content: systemPrompt})
	}

	if sess, exists := s.sessions[chatID]; exists {
		msgs = append(msgs, sess.History...)
	}

	return msgs
}

func (s *Store) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, exists := s.sessions[chatID]; exists {
		sess.History = make([]ai.Message, 0)
		sess.LastActive = time.Now()
	}
}

func (s *Store) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
