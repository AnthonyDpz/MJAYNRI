package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
	"github.com/AnthonyDpz/MJAYNRI/internal/server/handlers"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestEngine crée un moteur Gin de test avec le handler status monté.
func newTestEngine(manager *llm.Manager) *gin.Engine {
	engine := gin.New()
	h := handlers.New(manager)
	engine.GET("/api/status", h.Status)
	engine.GET("/api/models", h.Models)
	return engine
}

func TestStatus_Connected(t *testing.T) {
	// Créer un manager avec un provider mock
	mock := &mockProvider{name: "Ollama", model: "llama3.2:3b", url: "http://localhost:11434"}
	manager := llm.NewManager([]llm.DetectedProvider{
		{Provider: mock, Models: []string{"llama3.2:3b", "mistral:7b"}},
	})

	engine := newTestEngine(manager)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	if resp["status"] != "connected" {
		t.Errorf("status: got %q, want %q", resp["status"], "connected")
	}
	if resp["color"] != "green" {
		t.Errorf("color: got %q, want %q", resp["color"], "green")
	}
	if resp["provider"] != "Ollama" {
		t.Errorf("provider: got %q, want %q", resp["provider"], "Ollama")
	}
	if resp["model"] != "llama3.2:3b" {
		t.Errorf("model: got %q, want %q", resp["model"], "llama3.2:3b")
	}
}

func TestStatus_Disconnected(t *testing.T) {
	manager := llm.NewManager(nil) // Aucun provider

	engine := newTestEngine(manager)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "disconnected" {
		t.Errorf("status: got %q, want %q", resp["status"], "disconnected")
	}
	if resp["color"] != "red" {
		t.Errorf("color: got %q, want %q", resp["color"], "red")
	}
}

func TestModels_ReturnsList(t *testing.T) {
	mock := &mockProvider{name: "Ollama", model: "llama3.2:3b", url: "http://localhost:11434"}
	manager := llm.NewManager([]llm.DetectedProvider{
		{Provider: mock, Models: []string{"llama3.2:3b", "mistral:7b"}},
	})

	engine := newTestEngine(manager)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	providers, ok := resp["providers"].([]interface{})
	if !ok || len(providers) != 1 {
		t.Errorf("providers: got %v, want 1 entry", resp["providers"])
	}
}
