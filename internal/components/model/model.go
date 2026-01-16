package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"audit-workflow/internal/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type ChatModel interface {
	Generate(ctx context.Context, msgs []*schema.Message, opts ...einomodel.Option) (*schema.Message, error)
}

func NewChatModel(ctx context.Context, cfg *config.RootConfig) (ChatModel, error) {
	timeout := time.Duration(cfg.AI.TimeoutS * float64(time.Second))
	baseURL := strings.TrimRight(cfg.AI.BaseURL, "/")

	switch strings.ToLower(cfg.AI.Provider) {
	case "", "doubao-ai", "ark":
		modelConfig := &ark.ChatModelConfig{
			APIKey:  cfg.AI.APIKey,
			Model:   cfg.AI.Model,
			BaseURL: baseURL,
			Timeout: &timeout,
			Region:  "cn-beijing",
		}
		return ark.NewChatModel(ctx, modelConfig)
	case "openai", "openai_compat", "openai-compatible", "deepseek", "chaitin":
		return newOpenAICompatChatModel(openAICompatConfig{
			BaseURL: baseURL,
			APIKey:  cfg.AI.APIKey,
			Model:   cfg.AI.Model,
			Timeout: timeout,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", cfg.AI.Provider)
	}
}

type openAICompatConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type openAICompatChatModel struct {
	cfg openAICompatConfig
	hc  *http.Client
}

func newOpenAICompatChatModel(cfg openAICompatConfig) *openAICompatChatModel {
	to := cfg.Timeout
	if to <= 0 {
		to = 60 * time.Second
	}
	return &openAICompatChatModel{cfg: cfg, hc: &http.Client{Timeout: to}}
}

type openAICompatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatRequest struct {
	Model    string                `json:"model"`
	Messages []openAICompatMessage `json:"messages"`
	Stream   bool                  `json:"stream,omitempty"`
}

type openAICompatResponse struct {
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

func (m *openAICompatChatModel) Generate(ctx context.Context, msgs []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	base := strings.TrimRight(m.cfg.BaseURL, "/")
	if base == "" {
		return nil, fmt.Errorf("empty base_url")
	}

	oaiMsgs := make([]openAICompatMessage, 0, len(msgs))
	for _, sm := range msgs {
		if sm == nil {
			continue
		}
		oaiMsgs = append(oaiMsgs, openAICompatMessage{Role: "user", Content: sm.Content})
	}

	reqBody, _ := json.Marshal(openAICompatRequest{Model: m.cfg.Model, Messages: oaiMsgs})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAICompatChatCompletionsURL(base), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(m.cfg.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	}

	resp, err := m.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai_compat http %d: %s", resp.StatusCode, string(b))
	}

	var out openAICompatResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	if out.Error != nil && strings.TrimSpace(out.Error.Message) != "" {
		return nil, fmt.Errorf("openai_compat error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("openai_compat empty choices")
	}

	return schema.AssistantMessage(out.Choices[0].Message.Content, nil), nil
}

func openAICompatChatCompletionsURL(base string) string {
	b := strings.TrimRight(base, "/")
	lb := strings.ToLower(b)
	if strings.HasSuffix(lb, "/v1") {
		return b + "/chat/completions"
	}
	return b + "/v1/chat/completions"
}
