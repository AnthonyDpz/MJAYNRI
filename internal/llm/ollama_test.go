package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

func TestOllamaProvider_Ping_OK(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3.2:3b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3.2:3b", time.Second)
	if !p.Ping(context.Background()) {
		t.Error("Ping should return true for a running Ollama server")
	}
}

func TestOllamaProvider_Ping_Unreachable(t *testing.T) {
	p := llm.NewOllamaProvider("http://localhost:1", "llama3.2:3b", 100*time.Millisecond)
	if p.Ping(context.Background()) {
		t.Error("Ping should return false for an unreachable server")
	}
}

func TestOllamaProvider_ListModels(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3.2:3b", "mistral:7b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3.2:3b", time.Second)
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("ListModels: got %d, want 2", len(models))
	}
	if models[0] != "llama3.2:3b" {
		t.Errorf("ListModels[0]: got %q, want %q", models[0], "llama3.2:3b")
	}
}

func TestOllamaProvider_Chat_Streaming(t *testing.T) {
	// Serveur simulant le streaming NDJSON d'Ollama
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"models": []map[string]string{{"name": "llama3.2:3b"}},
			})
			return
		}
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "application/x-ndjson")
			// Simule 3 chunks + done
			chunks := []map[string]interface{}{
				{"message": map[string]string{"content": "Bonjour"}, "done": false},
				{"message": map[string]string{"content": " !"}, "done": false},
				{"message": map[string]string{"content": ""}, "done": true},
			}
			for _, c := range chunks {
				json.NewEncoder(w).Encode(c)
				w.(http.Flusher).Flush()
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3.2:3b", time.Second)
	ch, err := p.Chat(context.Background(), []llm.Message{
		{Role: llm.RoleUser, Content: "Dis bonjour"},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	var result string
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatalf("Chunk error: %v", chunk.Err)
		}
		result += chunk.Content
	}

	if result != "Bonjour !" {
		t.Errorf("Chat result: got %q, want %q", result, "Bonjour !")
	}
}

func TestOllamaProvider_Metadata(t *testing.T) {
	p := llm.NewOllamaProvider("http://localhost:11434", "llama3.2:3b", time.Second)
	if p.Name() != "Ollama" {
		t.Errorf("Name: got %q, want %q", p.Name(), "Ollama")
	}
	if p.ModelName() != "llama3.2:3b" {
		t.Errorf("ModelName: got %q, want %q", p.ModelName(), "llama3.2:3b")
	}
	if p.BaseURL() != "http://localhost:11434" {
		t.Errorf("BaseURL: got %q, want %q", p.BaseURL(), "http://localhost:11434")
	}
}

func TestOllamaProvider_Chat_ContextCancel(t *testing.T) {
	// Ce test vérifie qu'un contexte déjà annulé empêche l'appel HTTP.
	// On utilise un contexte pré-annulé pour éviter tout blocage réseau.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Annulé immédiatement avant l'appel

	p := llm.NewOllamaProvider("http://localhost:1", "test", 100*time.Millisecond)
	_, err := p.Chat(ctx, []llm.Message{{Role: llm.RoleUser, Content: "test"}})

	// Avec un contexte annulé, Chat doit retourner une erreur immédiatement
	if err == nil {
		t.Error("Chat with cancelled context should return an error")
	}
}
