package llm_test

import (
	"testing"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// testTimeout est le délai utilisé pour les providers dans les tests.
const testTimeout = 5 * time.Second

// --- helpers ----------------------------------------------------------------

// newTestResolver retourne un Resolver pré-configuré avec deux faux serveurs.
// Utilisé pour tester Manager sans vraie connexion réseau.
func makeManager(providers []llm.DetectedProvider) *llm.Manager {
	return llm.NewManager(providers)
}

// --- tests ------------------------------------------------------------------

func TestManager_Switch_OK(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3:8b", "mistral:7b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3:8b", testTimeout)
	mgr := makeManager([]llm.DetectedProvider{
		{Provider: p, Models: []string{"llama3:8b", "mistral:7b"}},
	})

	if err := mgr.Switch("Ollama", "mistral:7b"); err != nil {
		t.Fatalf("Switch inattendu : %v", err)
	}

	info := mgr.StatusInfo()
	if info.ActiveModel != "mistral:7b" {
		t.Errorf("ActiveModel attendu %q, obtenu %q", "mistral:7b", info.ActiveModel)
	}
	if info.Status != llm.StatusConnected {
		t.Errorf("statut attendu connected, obtenu %q", info.Status)
	}
}

func TestManager_Switch_UnknownProvider(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3:8b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3:8b", testTimeout)
	mgr := makeManager([]llm.DetectedProvider{
		{Provider: p, Models: []string{"llama3:8b"}},
	})

	err := mgr.Switch("ProviderInconnu", "llama3:8b")
	if err == nil {
		t.Fatal("attendait une erreur pour provider inconnu")
	}
}

func TestManager_Switch_UnknownModel(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3:8b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3:8b", testTimeout)
	mgr := makeManager([]llm.DetectedProvider{
		{Provider: p, Models: []string{"llama3:8b"}},
	})

	err := mgr.Switch("Ollama", "modele-inexistant")
	if err == nil {
		t.Fatal("attendait une erreur pour modèle inconnu")
	}
}

func TestManager_Switch_PreservesOtherProvider(t *testing.T) {
	// Deux providers : on switch sur LM Studio, Ollama doit rester dans la liste.
	ollamaSrv := newTestOllamaServer([]string{"llama3:8b"})
	defer ollamaSrv.Close()
	lmSrv := newTestLMStudioServer([]string{"mistral-7b"})
	defer lmSrv.Close()

	pOllama := llm.NewOllamaProvider(ollamaSrv.URL, "llama3:8b", testTimeout)
	pLM := llm.NewLMStudioProvider(lmSrv.URL, "mistral-7b", testTimeout)

	mgr := makeManager([]llm.DetectedProvider{
		{Provider: pOllama, Models: []string{"llama3:8b"}},
		{Provider: pLM, Models: []string{"mistral-7b"}},
	})

	// Actif initial = Ollama
	if mgr.Active().Name() != "Ollama" {
		t.Fatalf("provider initial attendu Ollama, obtenu %s", mgr.Active().Name())
	}

	// Switch vers LM Studio
	if err := mgr.Switch("LM Studio", "mistral-7b"); err != nil {
		t.Fatalf("Switch : %v", err)
	}

	if mgr.Active().Name() != "LM Studio" {
		t.Errorf("provider attendu LM Studio, obtenu %s", mgr.Active().Name())
	}

	// Les deux providers sont toujours disponibles
	if len(mgr.Available()) != 2 {
		t.Errorf("attendu 2 providers disponibles, obtenu %d", len(mgr.Available()))
	}
}

func TestManager_StatusInfo_ReflectsSwitch(t *testing.T) {
	srv := newTestOllamaServer([]string{"llama3:8b", "gemma2:27b"})
	defer srv.Close()

	p := llm.NewOllamaProvider(srv.URL, "llama3:8b", testTimeout)
	mgr := makeManager([]llm.DetectedProvider{
		{Provider: p, Models: []string{"llama3:8b", "gemma2:27b"}},
	})

	_ = mgr.Switch("Ollama", "gemma2:27b")

	info := mgr.StatusInfo()
	if info.ActiveModel != "gemma2:27b" {
		t.Errorf("StatusInfo.ActiveModel attendu %q, obtenu %q", "gemma2:27b", info.ActiveModel)
	}
	// La liste complète des modèles doit toujours être présente
	if len(info.Models) != 2 {
		t.Errorf("attendu 2 modèles dans info.Models, obtenu %d", len(info.Models))
	}
}
