package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
	"github.com/AnthonyDpz/MJAYNRI/internal/server/handlers"
)

// newManagerWith crée un Manager avec un provider mock pour les tests de switch.
func newManagerWith(providers []llm.DetectedProvider) *llm.Manager {
	return llm.NewManager(providers)
}

// TestSwitch_Success vérifie qu'on peut changer de modèle sur un provider disponible.
func TestSwitch_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	p := &mockProvider{name: "Ollama", model: "llama3:8b", url: "http://localhost:11434", pingOK: true}
	detected := []llm.DetectedProvider{
		{Provider: p, Models: []string{"llama3:8b", "mistral:7b"}},
	}
	mgr := llm.NewManager(detected)
	h := handlers.New(mgr)

	engine := gin.New()
	engine.POST("/api/switch", h.Switch)

	body, _ := json.Marshal(map[string]string{
		"provider": "Ollama",
		"model":    "mistral:7b",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/switch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200, obtenu %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse JSON invalide : %v", err)
	}

	// Le modèle actif dans la réponse doit être celui demandé
	if got := resp["model"]; got != "mistral:7b" {
		t.Errorf("model attendu %q, obtenu %q", "mistral:7b", got)
	}
	if got := resp["color"]; got != "green" {
		t.Errorf("color attendue %q, obtenue %q", "green", got)
	}
}

// TestSwitch_UnknownProvider vérifie le 400 si le provider n'existe pas.
func TestSwitch_UnknownProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	p := &mockProvider{name: "Ollama", model: "llama3:8b", url: "http://localhost:11434", pingOK: true}
	mgr := llm.NewManager([]llm.DetectedProvider{{Provider: p, Models: []string{"llama3:8b"}}})
	h := handlers.New(mgr)

	engine := gin.New()
	engine.POST("/api/switch", h.Switch)

	body, _ := json.Marshal(map[string]string{
		"provider": "ProviderInexistant",
		"model":    "llama3:8b",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/switch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400, obtenu %d", w.Code)
	}
}

// TestSwitch_UnknownModel vérifie le 400 si le modèle n'est pas dans la liste.
func TestSwitch_UnknownModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	p := &mockProvider{name: "Ollama", model: "llama3:8b", url: "http://localhost:11434", pingOK: true}
	mgr := llm.NewManager([]llm.DetectedProvider{{Provider: p, Models: []string{"llama3:8b"}}})
	h := handlers.New(mgr)

	engine := gin.New()
	engine.POST("/api/switch", h.Switch)

	body, _ := json.Marshal(map[string]string{
		"provider": "Ollama",
		"model":    "modele-inexistant",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/switch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400, obtenu %d", w.Code)
	}
}

// TestSwitch_InvalidBody vérifie le 400 si le JSON est malformé.
func TestSwitch_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mgr := llm.NewManager(nil)
	h := handlers.New(mgr)

	engine := gin.New()
	engine.POST("/api/switch", h.Switch)

	req := httptest.NewRequest(http.MethodPost, "/api/switch", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400, obtenu %d", w.Code)
	}
}
