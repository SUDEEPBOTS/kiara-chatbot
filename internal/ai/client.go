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

package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ContentPart struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	ImageURL *ImageURLField `json:"image_url,omitempty"`
}

type ImageURLField struct {
	URL string `json:"url"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Client struct {
	apiURL     string
	apiKey     string
	groqKey    string
	geminiKey  string
	provider   string
	httpClient *http.Client
}

func NewClient(apiURL, apiKey, groqKey, geminiKey, provider string) *Client {
	return &Client{
		apiURL:    apiURL,
		apiKey:    apiKey,
		groqKey:   groqKey,
		geminiKey: geminiKey,
		provider:  provider,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) resolveEndpoint(model string) (string, string) {
	if c.groqKey != "" && (c.provider == "groq" || strings.HasPrefix(model, "llama-") || strings.HasPrefix(model, "mixtral-") || strings.HasPrefix(model, "gemma")) {
		return "https://api.groq.com/openai/v1/chat/completions", c.groqKey
	}

	if c.geminiKey != "" && (c.provider == "gemini" || strings.HasPrefix(model, "official-gemini-")) {
		return "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", c.geminiKey
	}

	return c.apiURL, c.apiKey
}

func (c *Client) GenerateReply(model string, messages []Message) (string, error) {
	targetURL, targetKey := c.resolveEndpoint(model)
	cleanModel := strings.TrimPrefix(model, "official-gemini-")

	reqBody := ChatRequest{
		Model:     cleanModel,
		Messages:  messages,
		MaxTokens: 300,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if targetKey != "" {
		req.Header.Set("Authorization", "Bearer "+targetKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

type ImageGenRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

type ImageGenResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) GenerateImage(prompt string) (string, error) {
	imgURL := c.apiURL
	if len(imgURL) > 17 && imgURL[len(imgURL)-17:] == "/chat/completions" {
		imgURL = imgURL[:len(imgURL)-17] + "/images/generations"
	}

	reqBody := ImageGenRequest{
		Prompt: prompt,
		N:      1,
		Size:   "1024x1024",
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode image request: %w", err)
	}

	req, err := http.NewRequest("POST", imgURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create image request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("image api request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var imgResp ImageGenResponse
	if err := json.Unmarshal(bodyBytes, &imgResp); err != nil {
		return "", fmt.Errorf("failed to parse image response: %w", err)
	}

	if imgResp.Error != nil && imgResp.Error.Message != "" {
		return "", fmt.Errorf("image api error: %s", imgResp.Error.Message)
	}

	if len(imgResp.Data) == 0 || imgResp.Data[0].URL == "" {
		return "", fmt.Errorf("no image url returned")
	}

	return imgResp.Data[0].URL, nil
}
