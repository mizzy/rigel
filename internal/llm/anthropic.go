package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
	client *anthropic.Client
	model  string
}

func NewAnthropicProvider(apiKey string, model string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	return &AnthropicProvider{
		client: client,
		model:  model,
	}, nil
}

func (p *AnthropicProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithOptions(ctx, prompt, GenerateOptions{})
}

func (p *AnthropicProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	model := p.model
	if opts.Model != "" {
		model = opts.Model
	}

	maxTokens := 4096
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(model),
		Messages:  anthropic.F(messages),
		MaxTokens: anthropic.F(int64(maxTokens)),
	}

	if opts.SystemPrompt != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(opts.SystemPrompt),
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
			Model: anthropic.F(p.model),
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
