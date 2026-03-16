package llm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AnthonyDpz/MJAYNRI/internal/config"
)

// Resolver sonde les serveurs LLM locaux au démarrage et retourne les providers détectés.
//
// Ordre de détection :
//  1. Ollama     (localhost:11434) — priorité locale
//  2. LM Studio  (localhost:1234)  — priorité locale
//
// Les deux peuvent être actifs simultanément ; le Manager choisira l'actif.
type Resolver struct {
	cfg config.LLMConfig
}

// NewResolver crée un Resolver avec la configuration LLM fournie.
func NewResolver(cfg config.LLMConfig) *Resolver {
	return &Resolver{cfg: cfg}
}

// DetectedProvider regroupe un provider opérationnel et les modèles qu'il expose.
type DetectedProvider struct {
	Provider Provider
	Models   []string
}

// Detect sonde tous les providers connus et retourne ceux qui répondent.
// Cette fonction est rapide (timeout court par provider) et ne bloque pas au démarrage.
func (r *Resolver) Detect(ctx context.Context) ([]DetectedProvider, error) {
	var detected []DetectedProvider

	// Sonde Ollama
	if p, err := r.probeOllama(ctx); err == nil {
		detected = append(detected, p)
		log.Printf("[LLM] Ollama détecté : %d modèle(s) disponible(s)", len(p.Models))
	} else {
		log.Printf("[LLM] Ollama non disponible : %v", err)
	}

	// Sonde LM Studio
	if p, err := r.probeLMStudio(ctx); err == nil {
		detected = append(detected, p)
		log.Printf("[LLM] LM Studio détecté : %d modèle(s) disponible(s)", len(p.Models))
	} else {
		log.Printf("[LLM] LM Studio non disponible : %v", err)
	}

	if len(detected) == 0 {
		log.Println("[LLM] Aucun provider local trouvé — IHM démarrée en mode déconnecté")
	}

	return detected, nil
}

// probeOllama tente de se connecter au serveur Ollama et récupère les modèles.
func (r *Resolver) probeOllama(ctx context.Context) (DetectedProvider, error) {
	probeCtx, cancel := context.WithTimeout(ctx, r.cfg.ProbeTimeout)
	defer cancel()

	// Timeout court pour le client de sondage
	provider := NewOllamaProvider(r.cfg.OllamaURL, "", r.cfg.ProbeTimeout)

	if !provider.Ping(probeCtx) {
		return DetectedProvider{}, fmt.Errorf("ollama: ping échoué sur %s", r.cfg.OllamaURL)
	}

	models, err := provider.ListModels(probeCtx)
	if err != nil {
		return DetectedProvider{}, fmt.Errorf("ollama: list models: %w", err)
	}
	if len(models) == 0 {
		return DetectedProvider{}, fmt.Errorf("ollama: aucun modèle disponible")
	}

	// Sélectionner le modèle par défaut : celui configuré, ou le premier disponible
	selectedModel := r.selectModel(models, r.cfg.DefaultModel)
	finalProvider := NewOllamaProvider(r.cfg.OllamaURL, selectedModel, 30*time.Second)

	return DetectedProvider{Provider: finalProvider, Models: models}, nil
}

// probeLMStudio tente de se connecter à LM Studio et récupère les modèles.
func (r *Resolver) probeLMStudio(ctx context.Context) (DetectedProvider, error) {
	probeCtx, cancel := context.WithTimeout(ctx, r.cfg.ProbeTimeout)
	defer cancel()

	provider := NewLMStudioProvider(r.cfg.LMStudioURL, "", r.cfg.ProbeTimeout)

	if !provider.Ping(probeCtx) {
		return DetectedProvider{}, fmt.Errorf("lmstudio: ping échoué sur %s", r.cfg.LMStudioURL)
	}

	models, err := provider.ListModels(probeCtx)
	if err != nil {
		return DetectedProvider{}, fmt.Errorf("lmstudio: list models: %w", err)
	}
	if len(models) == 0 {
		return DetectedProvider{}, fmt.Errorf("lmstudio: aucun modèle disponible")
	}

	selectedModel := r.selectModel(models, r.cfg.DefaultModel)
	finalProvider := NewLMStudioProvider(r.cfg.LMStudioURL, selectedModel, 30*time.Second)

	return DetectedProvider{Provider: finalProvider, Models: models}, nil
}

// selectModel retourne preferred si présent dans models, sinon le premier modèle.
func (r *Resolver) selectModel(models []string, preferred string) string {
	if preferred != "" {
		for _, m := range models {
			if m == preferred {
				return m
			}
		}
	}
	return models[0]
}
