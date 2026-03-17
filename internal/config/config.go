// Package config centralise la configuration de l'application MJAYNRI.
// Les valeurs sont lues depuis les variables d'environnement avec des valeurs
// par défaut raisonnables, sans dépendance externe.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config regroupe toutes les configurations de l'application.
type Config struct {
	Server ServerConfig
	LLM    LLMConfig
}

// ServerConfig contient les paramètres du serveur HTTP.
type ServerConfig struct {
	// Port d'écoute (défaut : 8080)
	Port string
	// ReadTimeout pour les requêtes entrantes
	ReadTimeout time.Duration
	// WriteTimeout pour les réponses (SSE nécessite un timeout long)
	WriteTimeout time.Duration
}

// LLMConfig contient les paramètres de détection des providers IA.
type LLMConfig struct {
	// OllamaURL est l'URL de base du serveur Ollama (défaut : http://localhost:11434)
	OllamaURL string
	// LMStudioURL est l'URL de base de LM Studio (défaut : http://localhost:1234)
	LMStudioURL string
	// ProbeTimeout est le délai maximum pour sonder un provider
	ProbeTimeout time.Duration
	// DefaultModel est le modèle sélectionné au démarrage si plusieurs sont disponibles
	DefaultModel string
}

// Load construit la Config depuis les variables d'environnement.
// Toutes les variables ont des valeurs par défaut — aucune n'est obligatoire pour la V0.
func Load() Config {
	return Config{
		Server: ServerConfig{
			Port:        getEnv("MJAYNRI_PORT", "8080"),
			ReadTimeout: getDuration("MJAYNRI_READ_TIMEOUT", 15*time.Second),
			// WriteTimeout à 0 = aucun timeout d'écriture.
			// Obligatoire pour le streaming SSE : les grands modèles peuvent mettre
			// plusieurs minutes à générer une réponse. La déconnexion du client est
			// gérée par l'annulation du context.Context dans chaque handler.
			WriteTimeout: getDuration("MJAYNRI_WRITE_TIMEOUT", 0),
		},
		LLM: LLMConfig{
			OllamaURL:    getEnv("MJAYNRI_OLLAMA_URL", "http://localhost:11434"),
			LMStudioURL:  getEnv("MJAYNRI_LMSTUDIO_URL", "http://localhost:1234"),
			ProbeTimeout: getDuration("MJAYNRI_PROBE_TIMEOUT", 2*time.Second),
			DefaultModel: getEnv("MJAYNRI_DEFAULT_MODEL", ""),
		},
	}
}

// getEnv retourne la valeur de la variable d'environnement ou le fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getDuration parse une durée depuis une variable d'environnement (en secondes).
// Si la variable est absente ou invalide, retourne le fallback.
func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	secs, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return time.Duration(secs) * time.Second
}
