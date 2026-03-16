package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/config"
	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// newTestOllamaServer crée un serveur de test simulant l'API Ollama.
func newTestOllamaServer(models []string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		type modelEntry struct {
			Name string `json:"name"`
		}
		type response struct {
			Models []modelEntry `json:"models"`
		}
		entries := make([]modelEntry, len(models))
		for i, m := range models {
			entries[i] = modelEntry{Name: m}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{Models: entries})
	})
	return httptest.NewServer(mux)
}

// newTestLMStudioServer crée un serveur de test simulant l'API LM Studio (OpenAI-compat).
func newTestLMStudioServer(models []string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		type modelData struct {
			ID string `json:"id"`
		}
		type response struct {
			Data []modelData `json:"data"`
		}
		data := make([]modelData, len(models))
		for i, m := range models {
			data[i] = modelData{ID: m}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{Data: data})
	})
	return httptest.NewServer(mux)
}

func TestResolver_DetectsOllama(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3.2:3b", "mistral:7b"})
	defer srv.Close()

	resolver := llm.NewResolver(config.LLMConfig{
		OllamaURL:    srv.URL,
		LMStudioURL:  "http://localhost:1", // port inaccessible
		ProbeTimeout: time.Second,
	})

	providers, err := resolver.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: unexpected error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("Detect: got %d providers, want 1", len(providers))
	}
	if providers[0].Provider.Name() != "Ollama" {
		t.Errorf("Provider name: got %q, want %q", providers[0].Provider.Name(), "Ollama")
	}
	if len(providers[0].Models) != 2 {
		t.Errorf("Models count: got %d, want 2", len(providers[0].Models))
	}
}

func TestResolver_DetectsLMStudio(t *testing.T) {
	srv := newTestLMStudioServer([]string{"meta-llama-3.2-3b"})
	defer srv.Close()

	resolver := llm.NewResolver(config.LLMConfig{
		OllamaURL:    "http://localhost:1", // port inaccessible
		LMStudioURL:  srv.URL,
		ProbeTimeout: time.Second,
	})

	providers, err := resolver.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: unexpected error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("Detect: got %d providers, want 1", len(providers))
	}
	if providers[0].Provider.Name() != "LM Studio" {
		t.Errorf("Provider name: got %q, want %q", providers[0].Provider.Name(), "LM Studio")
	}
}

func TestResolver_NoneAvailable(t *testing.T) {
	resolver := llm.NewResolver(config.LLMConfig{
		OllamaURL:    "http://localhost:1",
		LMStudioURL:  "http://localhost:2",
		ProbeTimeout: 100 * time.Millisecond,
	})

	providers, err := resolver.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: unexpected error: %v", err)
	}
	if len(providers) != 0 {
		t.Errorf("Detect: got %d providers, want 0", len(providers))
	}
}

func TestResolver_DetectsBoth(t *testing.T) {
	ollamaSrv := newTestOllamaServer([]string{"llama3.2:3b"})
	defer ollamaSrv.Close()
	lmsSrv := newTestLMStudioServer([]string{"my-model"})
	defer lmsSrv.Close()

	resolver := llm.NewResolver(config.LLMConfig{
		OllamaURL:    ollamaSrv.URL,
		LMStudioURL:  lmsSrv.URL,
		ProbeTimeout: time.Second,
	})

	providers, _ := resolver.Detect(context.Background())
	if len(providers) != 2 {
		t.Errorf("Detect: got %d providers, want 2", len(providers))
	}
}

func TestManager_ActiveIsFirstDetected(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3.2:3b"})
	defer srv.Close()

	resolver := llm.NewResolver(config.LLMConfig{
		OllamaURL:    srv.URL,
		LMStudioURL:  "http://localhost:1",
		ProbeTimeout: time.Second,
	})

	detected, _ := resolver.Detect(context.Background())
	manager := llm.NewManager(detected)

	if manager.Active() == nil {
		t.Fatal("Active() should not be nil when a provider is detected")
	}
	if manager.Status() != llm.StatusConnected {
		t.Errorf("Status: got %q, want %q", manager.Status(), llm.StatusConnected)
	}
}

func TestManager_NoProviders_Disconnected(t *testing.T) {
	manager := llm.NewManager(nil)

	if manager.Active() != nil {
		t.Error("Active() should be nil when no providers")
	}
	if manager.Status() != llm.StatusDisconnected {
		t.Errorf("Status: got %q, want %q", manager.Status(), llm.StatusDisconnected)
	}
}
