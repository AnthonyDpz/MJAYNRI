package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Nettoyer toute variable d'environnement préexistante
	vars := []string{
		"MJAYNRI_PORT", "MJAYNRI_OLLAMA_URL", "MJAYNRI_LMSTUDIO_URL",
		"MJAYNRI_PROBE_TIMEOUT", "MJAYNRI_DEFAULT_MODEL",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}

	cfg := config.Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("Port: got %q, want %q", cfg.Server.Port, "8080")
	}
	if cfg.LLM.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL: got %q, want %q", cfg.LLM.OllamaURL, "http://localhost:11434")
	}
	if cfg.LLM.LMStudioURL != "http://localhost:1234" {
		t.Errorf("LMStudioURL: got %q, want %q", cfg.LLM.LMStudioURL, "http://localhost:1234")
	}
	if cfg.LLM.ProbeTimeout != 2*time.Second {
		t.Errorf("ProbeTimeout: got %v, want 2s", cfg.LLM.ProbeTimeout)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("MJAYNRI_PORT", "9090")
	os.Setenv("MJAYNRI_OLLAMA_URL", "http://192.168.1.10:11434")
	os.Setenv("MJAYNRI_PROBE_TIMEOUT", "5")
	defer func() {
		os.Unsetenv("MJAYNRI_PORT")
		os.Unsetenv("MJAYNRI_OLLAMA_URL")
		os.Unsetenv("MJAYNRI_PROBE_TIMEOUT")
	}()

	cfg := config.Load()

	if cfg.Server.Port != "9090" {
		t.Errorf("Port: got %q, want %q", cfg.Server.Port, "9090")
	}
	if cfg.LLM.OllamaURL != "http://192.168.1.10:11434" {
		t.Errorf("OllamaURL: got %q, want %q", cfg.LLM.OllamaURL, "http://192.168.1.10:11434")
	}
	if cfg.LLM.ProbeTimeout != 5*time.Second {
		t.Errorf("ProbeTimeout: got %v, want 5s", cfg.LLM.ProbeTimeout)
	}
}

func TestLoad_InvalidDuration_FallsBack(t *testing.T) {
	os.Setenv("MJAYNRI_PROBE_TIMEOUT", "not-a-number")
	defer os.Unsetenv("MJAYNRI_PROBE_TIMEOUT")

	cfg := config.Load()

	// Doit retomber sur la valeur par défaut
	if cfg.LLM.ProbeTimeout != 2*time.Second {
		t.Errorf("ProbeTimeout fallback: got %v, want 2s", cfg.LLM.ProbeTimeout)
	}
}
