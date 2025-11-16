package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the behaviour needed from an HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// OpenAIClient handles communication with OpenAI API
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient HTTPClient
}

const defaultOpenAIBaseURL = "https://api.openai.com/v1/chat/completions"

// NewOpenAIClient creates a new OpenAI client with the default HTTP client.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return NewOpenAIClientWithHTTPClient(apiKey, &http.Client{
		Timeout: 30 * time.Second,
	})
}

// NewOpenAIClientWithHTTPClient allows injecting a custom HTTP client (useful for testing).
func NewOpenAIClientWithHTTPClient(apiKey string, client HTTPClient) *OpenAIClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &OpenAIClient{
		apiKey:     apiKey,
		baseURL:    defaultOpenAIBaseURL,
		httpClient: client,
	}
}

// ChatMessage represents a message in the chat API
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the request to OpenAI API
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse represents the response from OpenAI API
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateWorkoutRequest sends a request to OpenAI to generate a workout
func (c *OpenAIClient) GenerateWorkoutRequest(prompt string) (string, error) {
	reqBody := ChatRequest{
		Model: "gpt-4.1-nano", // Using a cost-effective model
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a boxing trainer AI that generates realistic boxing combos. Always respond with valid JSON only, no additional text.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		// If unmarshaling fails, return raw body for non-200 status codes
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for error in response (even for non-200 status codes)
	if chatResp.Error != nil {
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, chatResp.Error.Message)
		}
		return "", fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}
