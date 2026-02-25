package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds the LLM client configuration.
type Config struct {
	BaseURL   string
	Model     string
	APIKey    string
	Timeout   time.Duration
	MaxTokens int
}

// DefaultConfig returns a default configuration for LM Studio.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "http://localhost:7090/v1",
		Model:     "openai/gpt-oss-20b",
		APIKey:    "not-needed",
		Timeout:   120 * time.Second,
		MaxTokens: 4096,
	}
}

// Client is an OpenAI-compatible LLM API client.
type Client struct {
	cfg  Config
	http *http.Client
}

// NewClient creates a new LLM client.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// ChatCompletion sends a non-streaming chat completion request and returns the full response text.
func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(5 * time.Second)
		}
		result, err := c.doChatCompletion(ctx, messages)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("chat completion failed after retries: %w", lastErr)
}

func (c *Client) doChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := ChatCompletionRequest{
		Model:     c.cfg.Model,
		Messages:  messages,
		Stream:    false,
		MaxTokens: c.cfg.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return chatResp.Choices[0].Message.Content, nil
}

// ChatCompletionStream sends a streaming chat completion request and calls onChunk for each text delta.
// Returns the full accumulated text.
func (c *Client) ChatCompletionStream(ctx context.Context, messages []ChatMessage, onChunk func(delta string)) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(5 * time.Second)
		}
		result, err := c.doChatCompletionStream(ctx, messages, onChunk)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("streaming chat completion failed after retries: %w", lastErr)
}

func (c *Client) doChatCompletionStream(ctx context.Context, messages []ChatMessage, onChunk func(delta string)) (string, error) {
	reqBody := ChatCompletionRequest{
		Model:     c.cfg.Model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: c.cfg.MaxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullContent.WriteString(delta)
				if onChunk != nil {
					onChunk(delta)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent.String(), fmt.Errorf("reading stream: %w", err)
	}

	return fullContent.String(), nil
}
