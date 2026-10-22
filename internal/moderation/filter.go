package moderation

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type Filter struct {
	mu         sync.RWMutex
	words      map[string]struct{}
	stems      []string
	sourcePath string
	sourceURL  string
}

func NewFilter(sourcePath string, sourceURL string) *Filter {
	f := &Filter{
		words:      make(map[string]struct{}),
		stems:      make([]string, 0),
		sourcePath: sourcePath,
		sourceURL:  sourceURL,
	}
	f.Load()
	return f
}

// Load reads blacklist from local file and optional remote URL
func (f *Filter) Load() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 1. Load from remote URL if configured
	if f.sourceURL != "" {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(f.sourceURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			f.parseFromReader(resp.Body)
			log.Printf("[INFO] Loaded abuse blocklist from remote URL: %s", f.sourceURL)
		} else if err != nil {
			log.Printf("[WARN] Failed to fetch remote abuse list from %s: %v", f.sourceURL, err)
		}
	}

	// 2. Load from local file if exists
	if f.sourcePath != "" {
		file, err := os.Open(f.sourcePath)
		if err == nil {
			defer file.Close()
			f.parseFromReader(file)
			log.Printf("[INFO] Loaded abuse blocklist from local file: %s (total: %d words)", f.sourcePath, len(f.words))
		} else {
			log.Printf("[INFO] Local abuse file %s not found. Creating empty template.", f.sourcePath)
			_ = os.WriteFile(f.sourcePath, []byte("# Add one abusive word or phrase per line\n"), 0644)
		}
	}
}

func (f *Filter) parseFromReader(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		normalized := f.NormalizeText(line)
		if normalized != "" {
			f.words[normalized] = struct{}{}
			if len(normalized) >= 4 {
				f.stems = append(f.stems, normalized)
			}
		}
	}
}

// AddWord manually registers a word
func (f *Filter) AddWord(word string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	norm := f.NormalizeText(word)
	if norm != "" {
		f.words[norm] = struct{}{}
		if len(norm) >= 4 {
			f.stems = append(f.stems, norm)
		}
	}
}

// NormalizeText handles Unicode font conversion, leetspeak, punctuation removal, and repeated chars
func (f *Filter) NormalizeText(input string) string {
	// 1. Unicode font mapping (Math Bold, Script, Fraktur, Fullwidth to ASCII a-z)
	var sb strings.Builder
	for _, r := range input {
		ascii := unicodeToASCII(r)
		sb.WriteRune(ascii)
	}
	text := strings.ToLower(sb.String())

	// 2. Leetspeak mapping
	leetMap := map[rune]rune{
		'0': 'o',
		'1': 'i',
		'!': 'i',
		'|': 'i',
		'3': 'e',
		'4': 'a',
		'@': 'a',
		'$': 's',
		'5': 's',
		'7': 't',
		'8': 'b',
	}
	var sb2 strings.Builder
	for _, r := range text {
		if replacement, ok := leetMap[r]; ok {
			sb2.WriteRune(replacement)
		} else {
			sb2.WriteRune(r)
		}
	}
	text = sb2.String()

	// 3. Strip all non-alphabetic separators (dots, dashes, stars, underscores, spaces)
	// Example: "f.u.c.k" -> "fuck", "b_h_e_n" -> "bhen"
	reg := regexp.MustCompile(`[^a-zA-Z\p{Devanagari}\p{Gurmukhi}]`)
	clean := reg.ReplaceAllString(text, "")

	// 4. Collapse repeating characters (e.g., "fuuuuck" -> "fuck", "saaaale" -> "sale")
	return collapseRepeatedRunes(clean)
}

func unicodeToASCII(r rune) rune {
	// Mathematical Monospace, Bold, Sans-Serif, Script, Fraktur, etc.
	// Ranges like U+1D400..U+1D7FF and U+FF01..U+FF5E (Fullwidth)
	if r >= 0xFF21 && r <= 0xFF3A { // Fullwidth Latin Capital
		return r - 0xFF21 + 'a'
	}
	if r >= 0xFF41 && r <= 0xFF5A { // Fullwidth Latin Small
		return r - 0xFF41 + 'a'
	}
	// Common Math Alphanumeric Symbols offset mapping
	if r >= 0x1D400 && r <= 0x1D7CB {
		// Calculate letter offset within alphabets
		val := uint32(r)
		for base := uint32(0x1D400); base <= 0x1D700; base += 52 {
			if val >= base && val < base+26 {
				return rune('a' + (val - base))
			}
			if val >= base+26 && val < base+52 {
				return rune('a' + (val - (base + 26)))
			}
		}
	}
	return r
}

func collapseRepeatedRunes(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	var result []rune
	for i, r := range runes {
		if i == 0 || r != runes[i-1] {
			result = append(result, r)
		}
	}
	return string(result)
}

// IsAbusive checks if text contains blacklisted profanity, dots evasion, or typos
func (f *Filter) IsAbusive(text string) (bool, string) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.words) == 0 {
		return false, ""
	}

	norm := f.NormalizeText(text)
	if norm == "" {
		return false, ""
	}

	// 1. Exact match on normalized string
	if _, ok := f.words[norm]; ok {
		return true, norm
	}

	// 2. Substring / Stem check
	for _, stem := range f.stems {
		if strings.Contains(norm, stem) {
			return true, stem
		}
	}

	// 3. Word token check with fuzzy Levenshtein distance
	tokens := strings.Fields(text)
	for _, tok := range tokens {
		normTok := f.NormalizeText(tok)
		if normTok == "" {
			continue
		}
		if _, ok := f.words[normTok]; ok {
			return true, normTok
		}

		// Fuzzy check for minor typos if token length >= 4
		if len(normTok) >= 4 {
			for _, stem := range f.stems {
				if abs(len(normTok)-len(stem)) <= 1 {
					dist := levenshteinDistance(normTok, stem)
					if dist <= 1 {
						return true, stem
					}
				}
			}
		}
	}

	return false, ""
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func levenshteinDistance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)
		}
	}
	return d[len(s)][len(t)]
}

// ModerateMessage inspects incoming group message and deletes + warns if abusive
func (f *Filter) ModerateMessage(m *telegram.NewMessage) bool {
	// Moderation is strictly for group chats
	if m.IsPrivate() {
		return false
	}

	rawText := m.Text()
	if rawText == "" {
		return false
	}

	isBad, matched := f.IsAbusive(rawText)
	if !isBad {
		return false
	}

	log.Printf("[MODERATION] Abusive message detected from user %d in chat %d (matched: %s). Deleting...",
		m.SenderID(), m.ChatID(), matched)

	// 1. Delete the offensive message
	_, _ = m.Delete()

	// 2. Mention user and send blockquote warning
	senderName := "user"
	if m.Sender != nil {
		if m.Sender.Username != "" {
			senderName = "@" + m.Sender.Username
		} else if m.Sender.FirstName != "" {
			senderName = m.Sender.FirstName
		}
	}

	warning := fmt.Sprintf("<blockquote>⚠️ %s, abusive and vulgar language is not allowed here. Keep the chat respectful!</blockquote>", senderName)

	warnMsg, err := m.Respond(warning)
	if err == nil && warnMsg != nil {
		// 3. Auto-delete warning message after 8 seconds
		go func() {
			time.Sleep(8 * time.Second)
			_, _ = warnMsg.Delete()
		}()
	}

	return true
}
