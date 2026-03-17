package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LMStudioProvider implémente l'interface Provider pour LM Studio.
// LM Studio expose une API compatible OpenAI sur le port 1234 par défaut.
// Doc : https://lmstudio.ai/docs/local-server
type LMStudioProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewLMStudioProvider crée un provider LM Studio sur baseURL avec le modèle model.
func NewLMStudioProvider(baseURL, model string, timeout time.Duration) *LMStudioProvider {
	return &LMStudioProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (l *LMStudioProvider) Name() string      { return "LM Studio" }
func (l *LMStudioProvider) ModelName() string { return l.model }
func (l *LMStudioProvider) BaseURL() string   { return l.baseURL }

// Ping vérifie la disponibilité de LM Studio en appelant GET /v1/models.
func (l *LMStudioProvider) Ping(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+"/v1/models", nil)
	if err != nil {
		return false
	}
	resp, err := l.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// openAIModelsResponse est la réponse de GET /v1/models (format OpenAI).
type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ListModels retourne tous les modèles disponibles sur LM Studio.
func (l *LMStudioProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: build request: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: GET /v1/models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lmstudio: GET /v1/models: status %d", resp.StatusCode)
	}

	var models openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("lmstudio: decode models: %w", err)
	}

	ids := make([]string, 0, len(models.Data))
	for _, m := range models.Data {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

// openAIChatRequest est le corps de POST /v1/chat/completions (format OpenAI).
type openAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatChunk est un fragment SSE de la réponse streaming OpenAI.
// Format : data: {"choices":[{"delta":{"content":"..."},"finish_reason":null}]}
type openAIChatChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// Chat envoie les messages à LM Studio et retourne un canal de chunks (SSE OpenAI).
func (l *LMStudioProvider) Chat(ctx context.Context, messages []Message) (<-chan Chunk, error) {
	openAIMsgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		openAIMsgs[i] = openAIMessage{Role: string(m.Role), Content: m.Content}
	}

	body, err := json.Marshal(openAIChatRequest{
		Model:    l.model,
		Messages: openAIMsgs,
		Stream:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("lmstudio: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		l.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("lmstudio: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: POST /v1/chat/completions: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("lmstudio: chat: status %d", resp.StatusCode)
	}

	ch := make(chan Chunk, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Les lignes SSE commencent par "data: "
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")

			// "[DONE]" signale la fin du stream OpenAI
			if payload == "[DONE]" {
				ch <- Chunk{Done: true}
				return
			}

			var chunk openAIChatChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				ch <- Chunk{Err: fmt.Errorf("lmstudio: decode chunk: %w", err)}
				return
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]
			done := choice.FinishReason != nil && *choice.FinishReason == "stop"
			ch <- Chunk{Content: choice.Delta.Content, Done: done}

			if done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Chunk{Err: fmt.Errorf("lmstudio: stream read: %w", err)}
		}
	}()

	return ch, nil
}
