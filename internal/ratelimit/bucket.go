package ratelimit

import (
	"sync"
	"time"
)

// LeakyBucket implements a token/leaky bucket rate limiter to prevent sticker spam in DMs.
type LeakyBucket struct {
	mu            sync.Mutex
	capacity      float64       // Max stickers allowed in burst (e.g. 3)
	leakRate      float64       // Stickers drained per second (e.g. 0.33 = 1 per 3 seconds)
	water         map[int64]float64
	lastUpdated   map[int64]time.Time
	blockedUntil  map[int64]time.Time
	blockDuration time.Duration // Duration to temporarily block on overflow (e.g. 60s)
}

// NewLeakyBucket creates a new rate limiter instance
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

// CheckSticker evaluates if a sticker sent by userID is permitted.
// Returns:
//   allowed: whether processing should continue
//   justBlocked: true only on the moment of overflow (to send a warning message)
//   remainingBlock: how many seconds remaining if blocked
func (b *LeakyBucket) CheckSticker(userID int64) (allowed bool, justBlocked bool, remainingBlock int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// 1. Check if user is currently blocked
	if until, isBlocked := b.blockedUntil[userID]; isBlocked {
		if now.Before(until) {
			rem := int(until.Sub(now).Seconds())
			if rem <= 0 {
				rem = 1
			}
			return false, false, rem // Still blocked, silently ignore
		}
		// Block expired, reset water
		delete(b.blockedUntil, userID)
		b.water[userID] = 0
		b.lastUpdated[userID] = now
	}

	// 2. Calculate leakage
	last, exists := b.lastUpdated[userID]
	if !exists {
		b.lastUpdated[userID] = now
		b.water[userID] = 1.0
		return true, false, 0
	}

	elapsed := now.Sub(last).Seconds()
	b.lastUpdated[userID] = now

	// Water leaks over time
	currentWater := b.water[userID] - (elapsed * b.leakRate)
	if currentWater < 0 {
		currentWater = 0
	}

	// 3. Add 1 sticker cost to bucket
	if currentWater+1.0 > b.capacity {
		// Overflow! Block user temporarily
		until := now.Add(b.blockDuration)
		b.blockedUntil[userID] = until
		b.water[userID] = b.capacity
		return false, true, int(b.blockDuration.Seconds())
	}

	b.water[userID] = currentWater + 1.0
	return true, false, 0
}
