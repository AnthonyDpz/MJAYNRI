package handlers_test

import (
	"context"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// mockProvider implémente llm.Provider pour les tests de handlers.
// Il répond immédiatement avec des données prédéfinies.
type mockProvider struct {
	name   string
	model  string
	url    string
	chunks []string // Séquence de chunks à retourner par Chat
	pingOK bool
}

func (m *mockProvider) Name() string     { return m.name }
func (m *mockProvider) ModelName() string { return m.model }
func (m *mockProvider) BaseURL() string   { return m.url }

func (m *mockProvider) Ping(_ context.Context) bool {
	return m.pingOK
}

func (m *mockProvider) ListModels(_ context.Context) ([]string, error) {
	return []string{m.model}, nil
}

func (m *mockProvider) Chat(_ context.Context, _ []llm.Message) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, len(m.chunks)+1)
	go func() {
		defer close(ch)
		for _, c := range m.chunks {
			ch <- llm.Chunk{Content: c}
		}
		ch <- llm.Chunk{Done: true}
	}()
	return ch, nil
}
