// Package ai provides a unified client for calling OpenAI, Anthropic, Gemini, and Kiro (Amazon Bedrock) chat completion APIs.
package ai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Provider identifies which AI backend to use.
type Provider string

const (
	OpenAI    Provider = "openai"
	Anthropic Provider = "anthropic"
	Gemini    Provider = "gemini"
	Kiro      Provider = "kiro" // Amazon Bedrock
)

// DefaultModel returns the default model for a provider. The
// IQTOOLKIT_AI_MODEL environment variable overrides all provider defaults.
func DefaultModel(p Provider) string {
	if m := os.Getenv("IQTOOLKIT_AI_MODEL"); m != "" {
		return m
	}
	switch p {
	case OpenAI:
		return "gpt-4o"
	case Anthropic:
		return "claude-sonnet-4-5"
	case Gemini:
		return "gemini-2.5-pro"
	case Kiro:
		return "anthropic.claude-sonnet-4-5-v1:0"
	default:
		return ""
	}
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request holds parameters for a completion call.
type Request struct {
	Model     string
	Messages  []Message
	System    string // system prompt (used directly by Anthropic, prepended as message for OpenAI)
	MaxTokens int    // maximum output tokens (default: 4096)
}

func (r Request) maxTokens() int {
	return cmp.Or(r.MaxTokens, 4096)
}

// HTTPError is returned when a provider responds with a non-200 status.
type HTTPError struct {
	Provider   Provider
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("ai: %s returned %d: %s", e.Provider, e.StatusCode, e.Body)
}

// Response holds the result of a completion call.
type Response struct {
	Content string
}

// Client calls an AI provider's chat completion API.
type Client struct {
	Provider   Provider
	APIKey     string
	Region     string        // AWS region for Kiro/Bedrock
	MaxRetries int           // retry attempts on transient errors (default: 3)
	Timeout    time.Duration // per-request timeout (default: 60s)
	HTTPClient *http.Client
}

// NewClient creates a Client for the given provider and API key.
func NewClient(provider Provider, apiKey string) *Client {
	return &Client{Provider: provider, APIKey: apiKey, HTTPClient: http.DefaultClient}
}

// Complete sends a chat completion request with retry on transient failures.
func (c *Client) Complete(ctx context.Context, req Request) (*Response, error) {
	var resp *Response
	var err error
	for attempt := range c.maxRetries() {
		reqCtx, cancel := context.WithTimeoutCause(ctx, c.timeout(), errors.New("ai: request timed out"))
		resp, err = c.dispatch(reqCtx, req)
		cancel()
		if err == nil {
			return resp, nil
		}
		if !isRetryable(err) {
			return nil, err
		}
		delay := time.Duration(1<<attempt) * time.Second // 1s, 2s, 4s
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, err
}

func (c *Client) maxRetries() int {
	if c.MaxRetries > 0 {
		return c.MaxRetries
	}
	return 3
}

func (c *Client) timeout() time.Duration {
	return cmp.Or(c.Timeout, 60*time.Second)
}

func isRetryable(err error) bool {
	// HTTP-level errors: retry on rate limits (429) and server errors (5xx).
	if he, ok := errors.AsType[*HTTPError](err); ok {
		return he.StatusCode == http.StatusTooManyRequests || he.StatusCode >= 500
	}
	// Network-level errors: timeouts and dropped connections.
	if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

func (c *Client) dispatch(ctx context.Context, req Request) (*Response, error) {
	switch c.Provider {
	case OpenAI:
		return c.completeOpenAI(ctx, req)
	case Anthropic:
		return c.completeAnthropic(ctx, req)
	case Gemini:
		return c.completeGemini(ctx, req)
	case Kiro:
		return c.completeKiro(ctx, req)
	default:
		return nil, fmt.Errorf("ai: unsupported provider %q", c.Provider)
	}
}

// --- OpenAI ---

func (c *Client) completeOpenAI(ctx context.Context, req Request) (*Response, error) {
	msgs := req.Messages
	if req.System != "" {
		msgs = append([]Message{{Role: "system", Content: req.System}}, msgs...)
	}
	body, _ := json.Marshal(map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": req.maxTokens(),
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{Provider: OpenAI, StatusCode: resp.StatusCode, Body: string(b)}
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("ai: openai returned no choices")
	}
	return &Response{Content: result.Choices[0].Message.Content}, nil
}

// --- Anthropic ---

func (c *Client) completeAnthropic(ctx context.Context, req Request) (*Response, error) {
	payload := map[string]any{
		"model":      req.Model,
		"max_tokens": req.maxTokens(),
		"messages":   req.Messages,
	}
	if req.System != "" {
		payload["system"] = req.System
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{Provider: Anthropic, StatusCode: resp.StatusCode, Body: string(b)}
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("ai: anthropic returned no content")
	}
	return &Response{Content: result.Content[0].Text}, nil
}

// --- Gemini ---

func (c *Client) completeGemini(ctx context.Context, req Request) (*Response, error) {
	var contents []map[string]any
	for _, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		})
	}
	payload := map[string]any{
		"contents":         contents,
		"generationConfig": map[string]any{"maxOutputTokens": req.maxTokens()},
	}
	if req.System != "" {
		payload["systemInstruction"] = map[string]any{
			"parts": []map[string]string{{"text": req.System}},
		}
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", req.Model, c.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{Provider: Gemini, StatusCode: resp.StatusCode, Body: string(b)}
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("ai: gemini returned no content")
	}
	return &Response{Content: result.Candidates[0].Content.Parts[0].Text}, nil
}

// --- Kiro (Amazon Bedrock) ---

func (c *Client) completeKiro(ctx context.Context, req Request) (*Response, error) {
	region := c.Region
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("ai: failed to load AWS config: %w", err)
	}
	client := bedrockruntime.NewFromConfig(cfg)

	var msgs []types.Message
	for _, m := range req.Messages {
		role := types.ConversationRoleUser
		if m.Role == "assistant" {
			role = types.ConversationRoleAssistant
		}
		msgs = append(msgs, types.Message{
			Role:    role,
			Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: m.Content}},
		})
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(req.Model),
		Messages:        msgs,
		InferenceConfig: &types.InferenceConfiguration{MaxTokens: new(int32(req.maxTokens()))},
	}
	if req.System != "" {
		input.System = []types.SystemContentBlock{&types.SystemContentBlockMemberText{Value: req.System}}
	}

	output, err := client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("ai: bedrock converse failed: %w", err)
	}
	if output.Output == nil {
		return nil, fmt.Errorf("ai: bedrock returned no output")
	}
	msg, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok || len(msg.Value.Content) == 0 {
		return nil, fmt.Errorf("ai: bedrock returned no content")
	}
	text, ok := msg.Value.Content[0].(*types.ContentBlockMemberText)
	if !ok {
		return nil, fmt.Errorf("ai: bedrock returned non-text content")
	}
	return &Response{Content: text.Value}, nil
}
