package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"flowtask-server/internal/config"
)

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	maxRetries int
	httpClient *http.Client
}

func NewClient(cfg config.AIConfig) *Client {
	return &Client{
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		maxRetries: cfg.MaxRetries,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type StreamDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *Client) Chat(messages []ChatMessage) (string, error) {
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	var lastErr error
	retryDelays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[AI_CLIENT] Retry attempt %d/%d after %v delay", attempt, c.maxRetries, retryDelays[attempt-1])
			time.Sleep(retryDelays[attempt-1])
		}

		log.Printf("[AI_CLIENT] Sending request to AI service (attempt %d/%d)", attempt+1, c.maxRetries+1)
		startTime := time.Now()

		result, err := c.doRequest(reqBody)
		elapsed := time.Since(startTime)

		if err == nil {
			log.Printf("[AI_CLIENT] Request successful in %v", elapsed)
			return result, nil
		}

		log.Printf("[AI_CLIENT] Request failed in %v: %v", elapsed, err)
		lastErr = err
	}

	log.Printf("[AI_CLIENT] All %d attempts failed, last error: %v", c.maxRetries+1, lastErr)
	return "", fmt.Errorf("AI request failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) ChatStream(messages []ChatMessage, onDelta func(content string) error) error {
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

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

		var delta StreamDelta
		if err := json.Unmarshal([]byte(data), &delta); err != nil {
			continue
		}

		if len(delta.Choices) > 0 && delta.Choices[0].Delta.Content != "" {
			if err := onDelta(delta.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func (c *Client) doRequest(reqBody ChatRequest) (string, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateTasks generates tasks using AI based on a prompt
func (c *Client) GenerateTasks(prompt string) ([]GeneratedTask, error) {
	messages := []ChatMessage{
		{Role: "system", Content: "你是一个学习计划生成助手。请根据用户的学习目标生成详细的任务列表。只返回 JSON 数组，不要添加任何解释。"},
		{Role: "user", Content: prompt},
	}

	content, err := c.Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("AI chat failed: %w", err)
	}

	// Parse the response
	parser := NewParser()
	var tasks []GeneratedTask
	if err := parser.ParseAIResponse(content, &tasks); err != nil {
		return nil, fmt.Errorf("parse AI response: %w", err)
	}

	return tasks, nil
}
