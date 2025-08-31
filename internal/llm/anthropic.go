package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
	client *anthropic.Client
	model  Model
	apiKey string
}

func NewAnthropicProvider(apiKey string, model string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	modelStruct := Model{Name: model}

	return &AnthropicProvider{
		client: client,
		model:  modelStruct,
		apiKey: apiKey,
	}, nil
}

func (p *AnthropicProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithOptions(ctx, prompt, GenerateOptions{})
}

func (p *AnthropicProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	// Convert single prompt to message history format
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return p.GenerateWithHistory(ctx, messages, opts)
}

func (p *AnthropicProvider) GenerateWithHistory(ctx context.Context, messages []Message, opts GenerateOptions) (string, error) {
	model := p.model.Name
	if opts.Model != "" {
		model = opts.Model
	}

	maxTokens := 4096
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	// Convert our Message format to Anthropic's MessageParam format
	anthropicMessages := make([]anthropic.MessageParam, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "user" {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == "assistant" {
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(model),
		Messages:  anthropic.F(anthropicMessages),
		MaxTokens: anthropic.F(int64(maxTokens)),
	}

	// Include AGENTS.md content in the system prompt
	systemPrompt := opts.SystemPrompt
	if systemPrompt == "" {
		// Load AGENTS.md even if no system prompt is provided
		systemPrompt = PrependAgentsContext("")
	} else {
		// Prepend AGENTS.md to existing system prompt
		systemPrompt = PrependAgentsContext(systemPrompt)
	}

	if systemPrompt != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		})
	}

	if opts.Temperature > 0 {
		params.Temperature = anthropic.F(float64(opts.Temperature))
	}

	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return message.Content[0].Text, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse)

	go func() {
		defer close(ch)

		stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model: anthropic.F(anthropic.Model(p.model.Name)),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			}),
			MaxTokens: anthropic.F(int64(4096)),
		})

		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case anthropic.MessageStreamEventTypeContentBlockDelta:
				if delta, ok := event.Delta.(anthropic.ContentBlockDeltaEventDelta); ok && delta.Text != "" {
					ch <- StreamResponse{
						Content: delta.Text,
						Done:    false,
					}
				}
			case anthropic.MessageStreamEventTypeMessageStop:
				ch <- StreamResponse{
					Done: true,
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamResponse{
				Error: err,
				Done:  true,
			}
		}
	}()

	return ch, nil
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]Model, error) {
	// Try to fetch models from API
	models, err := p.fetchModelsFromAPI(ctx)
	if err != nil {
		// Fall back to hardcoded list if API call fails
		return []Model{
			{Name: "claude-sonnet-4-20250514", Details: ModelDetails{Family: "claude-4"}},
			{Name: "claude-opus-4-20250131", Details: ModelDetails{Family: "claude-4"}},
			{Name: "claude-3-5-sonnet-20241022", Details: ModelDetails{Family: "claude-3-5"}},
			{Name: "claude-3-5-haiku-20241022", Details: ModelDetails{Family: "claude-3-5"}},
			{Name: "claude-3-opus-20240229", Details: ModelDetails{Family: "claude-3"}},
			{Name: "claude-3-sonnet-20240229", Details: ModelDetails{Family: "claude-3"}},
			{Name: "claude-3-haiku-20240307", Details: ModelDetails{Family: "claude-3"}},
		}, nil
	}
	return models, nil
}

func (p *AnthropicProvider) GetCurrentModel() Model {
	return p.model
}

func (p *AnthropicProvider) GetName() string {
	return "anthropic"
}

func (p *AnthropicProvider) SetModel(model Model) {
	p.model = model
}

// anthropicModelResponse represents the API response structure
type anthropicModelResponse struct {
	Data []struct {
		ID          string    `json:"id"`
		Type        string    `json:"type"`
		DisplayName string    `json:"display_name"`
		CreatedAt   time.Time `json:"created_at"`
	} `json:"data"`
	HasMore bool   `json:"has_more"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
}

func (p *AnthropicProvider) fetchModelsFromAPI(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var modelsResp anthropicModelResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]Model, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		// Extract family from model ID
		family := "claude"
		if len(m.ID) > 6 {
			family = m.ID[:6]
		}

		models = append(models, Model{
			Name: m.ID,
			Details: ModelDetails{
				Family: family,
			},
		})
	}

	// If no models returned from API, return error to trigger fallback
	if len(models) == 0 {
		return nil, fmt.Errorf("no models returned from API")
	}

	return models, nil
}
