package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(baseURL, model string) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	if model == "" {
		model = "llama3.2"
	}

	return &OllamaProvider{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}, nil
}

type ollamaGenerateRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	System  string        `json:"system,omitempty"`
	Stream  bool          `json:"stream"`
	Options ollamaOptions `json:"options,omitempty"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  ollamaOptions   `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

type ollamaOptions struct {
	Temperature float32 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type ollamaGenerateResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithOptions(ctx, prompt, GenerateOptions{})
}

func (p *OllamaProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	// Convert single prompt to message history format
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return p.GenerateWithHistory(ctx, messages, opts)
}

func (p *OllamaProvider) GenerateWithHistory(ctx context.Context, messages []Message, opts GenerateOptions) (string, error) {
	model := p.model
	if opts.Model != "" {
		model = opts.Model
	}

	// Convert our Message format to Ollama's format
	ollamaMessages := make([]ollamaMessage, 0, len(messages)+1)

	// Add system message if we have system prompt
	systemPrompt := opts.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = PrependAgentsContext("")
	} else {
		systemPrompt = PrependAgentsContext(systemPrompt)
	}

	if systemPrompt != "" {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range messages {
		ollamaMessages = append(ollamaMessages, ollamaMessage(msg))
	}

	reqBody := ollamaChatRequest{
		Model:    model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	if opts.Temperature > 0 {
		reqBody.Options.Temperature = opts.Temperature
	}

	if opts.MaxTokens > 0 {
		reqBody.Options.NumPredict = opts.MaxTokens
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use /api/chat endpoint for conversation history
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp ollamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

func (p *OllamaProvider) Stream(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse)

	go func() {
		defer close(ch)

		reqBody := ollamaGenerateRequest{
			Model:  p.model,
			Prompt: prompt,
			Stream: true,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			ch <- StreamResponse{
				Error: fmt.Errorf("failed to marshal request: %w", err),
				Done:  true,
			}
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
		if err != nil {
			ch <- StreamResponse{
				Error: fmt.Errorf("failed to create request: %w", err),
				Done:  true,
			}
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			ch <- StreamResponse{
				Error: fmt.Errorf("failed to send request: %w", err),
				Done:  true,
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- StreamResponse{
				Error: fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body)),
				Done:  true,
			}
			return
		}

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp ollamaGenerateResponse
			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					break
				}
				ch <- StreamResponse{
					Error: fmt.Errorf("failed to decode stream response: %w", err),
					Done:  true,
				}
				return
			}

			if streamResp.Response != "" {
				ch <- StreamResponse{
					Content: streamResp.Response,
					Done:    false,
				}
			}

			if streamResp.Done {
				ch <- StreamResponse{
					Done: true,
				}
				break
			}
		}
	}()

	return ch, nil
}

type ollamaListResponse struct {
	Models []Model `json:"models"`
}

func (p *OllamaProvider) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var listResp ollamaListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return listResp.Models, nil
}

func (p *OllamaProvider) GetCurrentModel() string {
	return p.model
}

func (p *OllamaProvider) SetModel(model string) {
	p.model = model
}
