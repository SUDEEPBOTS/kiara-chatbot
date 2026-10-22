package media

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type ProcessedMedia struct {
	Base64DataURL string
	MediaType     string // "photo", "sticker_static", "sticker_video", "sticker_animated"
	Emoji         string
	CleanPrompt   string
}

// ProcessMessageMedia extracts vision-ready image data and metadata from photos or stickers.
// Returns nil if message contains no supported visual media.
func ProcessMessageMedia(m *telegram.NewMessage) (*ProcessedMedia, error) {
	tmpDir := os.TempDir()
	uniqueID := fmt.Sprintf("%d_%d_%d", m.ChatID(), m.ID, time.Now().UnixNano())

	// 1. Photo handling
	if m.Photo() != nil {
		targetFile := filepath.Join(tmpDir, fmt.Sprintf("kiara_photo_%s.jpg", uniqueID))
		defer os.Remove(targetFile)

		downloadedPath, err := m.Download(&telegram.DownloadOptions{FileName: targetFile})
		if err != nil {
			return nil, fmt.Errorf("failed to download photo: %w", err)
		}
		defer os.Remove(downloadedPath)

		data, err := os.ReadFile(downloadedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read photo bytes: %w", err)
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		return &ProcessedMedia{
			Base64DataURL: "data:image/jpeg;base64," + b64,
			MediaType:     "photo",
			CleanPrompt:   strings.TrimSpace(m.Text()),
		}, nil
	}

	// 2. Sticker handling
	if sticker := m.Sticker(); sticker != nil {
		emoji := ""
		for _, attr := range sticker.Attributes {
			if st, ok := attr.(*telegram.DocumentAttributeSticker); ok {
				emoji = st.Alt
				break
			}
		}

		mime := strings.ToLower(sticker.MimeType)

		// A. Video Sticker (WebM)
		if strings.Contains(mime, "video") || strings.HasSuffix(mime, "webm") {
			rawWebm := filepath.Join(tmpDir, fmt.Sprintf("kiara_vstk_%s.webm", uniqueID))
			frameJpg := filepath.Join(tmpDir, fmt.Sprintf("kiara_frame_%s.jpg", uniqueID))
			defer os.Remove(rawWebm)
			defer os.Remove(frameJpg)

			downloadedPath, err := m.Download(&telegram.DownloadOptions{FileName: rawWebm})
			if err != nil {
				return nil, fmt.Errorf("failed to download video sticker: %w", err)
			}
			defer os.Remove(downloadedPath)

			// Use ffmpeg to grab a frame from the middle (~0.5s)
			cmd := exec.Command("ffmpeg", "-y", "-ss", "00:00:00.5", "-i", downloadedPath, "-vframes", "1", frameJpg)
			if err := cmd.Run(); err != nil {
				// Fallback without -ss in case the video sticker is under 0.5s
				cmd2 := exec.Command("ffmpeg", "-y", "-i", downloadedPath, "-vframes", "1", frameJpg)
				if err2 := cmd2.Run(); err2 != nil {
					return nil, fmt.Errorf("ffmpeg frame extraction failed: %w", err2)
				}
			}

			frameData, err := os.ReadFile(frameJpg)
			if err != nil {
				return nil, fmt.Errorf("failed to read extracted frame: %w", err)
			}

			b64 := base64.StdEncoding.EncodeToString(frameData)
			return &ProcessedMedia{
				Base64DataURL: "data:image/jpeg;base64," + b64,
				MediaType:     "sticker_video",
				Emoji:         emoji,
				CleanPrompt:   strings.TrimSpace(m.Text()),
			}, nil
		}

		// B. Animated Sticker (TGS / Lottie)
		if strings.Contains(mime, "tgsticker") || strings.Contains(mime, "lottie") {
			thumbPath := filepath.Join(tmpDir, fmt.Sprintf("kiara_thumb_%s.jpg", uniqueID))
			defer os.Remove(thumbPath)

			downloadedPath, err := m.Download(&telegram.DownloadOptions{
				ThumbOnly: true,
				FileName:  thumbPath,
			})
			if err == nil && downloadedPath != "" {
				defer os.Remove(downloadedPath)
				data, rErr := os.ReadFile(downloadedPath)
				if rErr == nil && len(data) > 0 {
					b64 := base64.StdEncoding.EncodeToString(data)
					return &ProcessedMedia{
						Base64DataURL: "data:image/jpeg;base64," + b64,
						MediaType:     "sticker_animated",
						Emoji:         emoji,
						CleanPrompt:   strings.TrimSpace(m.Text()),
					}, nil
				}
			}
		}

		// C. Static Sticker (WebP) or general sticker
		rawWebp := filepath.Join(tmpDir, fmt.Sprintf("kiara_stk_%s.webp", uniqueID))
		outJpg := filepath.Join(tmpDir, fmt.Sprintf("kiara_stk_%s.jpg", uniqueID))
		defer os.Remove(rawWebp)
		defer os.Remove(outJpg)

		downloadedPath, err := m.Download(&telegram.DownloadOptions{FileName: rawWebp})
		if err != nil {
			return nil, fmt.Errorf("failed to download sticker: %w", err)
		}
		defer os.Remove(downloadedPath)

		// Convert WebP to JPG for universal Vision model compatibility
		cmd := exec.Command("ffmpeg", "-y", "-i", downloadedPath, outJpg)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("ffmpeg webp conversion failed: %w", err)
		}

		data, err := os.ReadFile(outJpg)
		if err != nil {
			return nil, fmt.Errorf("failed to read converted sticker jpg: %w", err)
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		return &ProcessedMedia{
			Base64DataURL: "data:image/jpeg;base64," + b64,
			MediaType:     "sticker_static",
			Emoji:         emoji,
			CleanPrompt:   strings.TrimSpace(m.Text()),
		}, nil
	}

	return nil, nil
}
