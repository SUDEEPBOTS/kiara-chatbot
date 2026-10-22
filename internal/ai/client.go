package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	httpClient *http.Client
}

func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) GenerateReply(model string, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 200,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai api request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response json: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("ai api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned by model")
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
	// Replace /chat/completions with /images/generations
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
