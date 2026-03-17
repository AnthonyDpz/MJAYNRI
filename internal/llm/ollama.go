package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaProvider implémente l'interface Provider pour un serveur Ollama local.
// Ollama expose une API REST native sur le port 11434 par défaut.
// Doc : https://github.com/ollama/ollama/blob/main/docs/api.md
type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaProvider crée un provider Ollama configuré sur baseURL avec le modèle model.
func NewOllamaProvider(baseURL, model string, timeout time.Duration) *OllamaProvider {
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (o *OllamaProvider) Name() string      { return "Ollama" }
func (o *OllamaProvider) ModelName() string { return o.model }
func (o *OllamaProvider) BaseURL() string   { return o.baseURL }

// Ping vérifie la disponibilité d'Ollama en appelant GET /api/tags.
func (o *OllamaProvider) Ping(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ollamaTagsResponse est la réponse de GET /api/tags.
type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels retourne tous les modèles disponibles sur ce serveur Ollama.
func (o *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("ollama: build request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: GET /api/tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: GET /api/tags: status %d", resp.StatusCode)
	}

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("ollama: decode tags: %w", err)
	}

	names := make([]string, 0, len(tags.Models))
	for _, m := range tags.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

// ollamaChatRequest est le corps de POST /api/chat.
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatChunk est un fragment de réponse streaming de l'API Ollama.
type ollamaChatChunk struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Chat envoie les messages à Ollama et retourne un canal de chunks en streaming.
// Le client HTTP utilisé pour le streaming n'a pas de timeout global afin de
// permettre des réponses longues — le contexte ctx assure l'annulation.
func (o *OllamaProvider) Chat(ctx context.Context, messages []Message) (<-chan Chunk, error) {
	// Convertir les messages vers le format Ollama
	ollamaMsgs := make([]ollamaMessage, len(messages))
	for i, m := range messages {
		ollamaMsgs[i] = ollamaMessage{Role: string(m.Role), Content: m.Content}
	}

	body, err := json.Marshal(ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMsgs,
		Stream:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Client sans timeout global pour le streaming
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: POST /api/chat: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama: chat: status %d", resp.StatusCode)
	}

	ch := make(chan Chunk, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var chunk ollamaChatChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				ch <- Chunk{Err: fmt.Errorf("ollama: decode chunk: %w", err)}
				return
			}

			ch <- Chunk{Content: chunk.Message.Content, Done: chunk.Done}
			if chunk.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Chunk{Err: fmt.Errorf("ollama: stream read: %w", err)}
		}
	}()

	return ch, nil
}
