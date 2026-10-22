// ==============================================================================
// Copyright (c) 2026 SUDEEPBOTS (Sudeep Sharma). All rights reserved.
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

package ratelimit

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	mu            sync.Mutex
	capacity      float64
	leakRate      float64
	water         map[int64]float64
	lastUpdated   map[int64]time.Time
	blockedUntil  map[int64]time.Time
	blockDuration time.Duration
}

func NewLeakyBucket(capacity, leakRate float64, blockDuration time.Duration) *LeakyBucket {
	return &LeakyBucket{
		capacity:      capacity,
		leakRate:      leakRate,
		water:         make(map[int64]float64),
		lastUpdated:   make(map[int64]time.Time),
		blockedUntil:  make(map[int64]time.Time),
		blockDuration: blockDuration,
	}
}

func (b *LeakyBucket) CheckSticker(userID int64) (bool, bool, int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	if until, isBlocked := b.blockedUntil[userID]; isBlocked {
		if now.Before(until) {
			rem := int(until.Sub(now).Seconds())
			if rem <= 0 {
				rem = 1
			}
			return false, false, rem
		}
		delete(b.blockedUntil, userID)
		b.water[userID] = 0
		b.lastUpdated[userID] = now
	}

	last, exists := b.lastUpdated[userID]
	if !exists {
		b.lastUpdated[userID] = now
		b.water[userID] = 1.0
		return true, false, 0
	}

	elapsed := now.Sub(last).Seconds()
	b.lastUpdated[userID] = now

	currentWater := b.water[userID] - (elapsed * b.leakRate)
	if currentWater < 0 {
		currentWater = 0
	}

	if currentWater+1.0 > b.capacity {
		until := now.Add(b.blockDuration)
		b.blockedUntil[userID] = until
		b.water[userID] = b.capacity
		return false, true, int(b.blockDuration.Seconds())
	}

	b.water[userID] = currentWater + 1.0
	return true, false, 0
}
