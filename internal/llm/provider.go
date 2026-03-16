// Package llm fournit l'abstraction pour les fournisseurs de modèles de langage.
//
// Architecture :
//   - Provider   : interface commune à tous les providers (Ollama, LM Studio, Anthropic…)
//   - Manager    : gère le provider actif et permet de le changer à chaud
//   - Resolver   : sonde les serveurs locaux au démarrage et retourne les providers disponibles
//
// Pour ajouter un nouveau provider, il suffit d'implémenter l'interface Provider
// et de l'enregistrer dans Resolver.Detect().
package llm

import "context"

// Provider représente un fournisseur LLM actif et opérationnel.
// Toute implémentation doit être thread-safe.
type Provider interface {
	// Name retourne le nom lisible du provider (ex : "Ollama", "LM Studio").
	Name() string

	// ModelName retourne le nom du modèle actuellement chargé (ex : "llama3.2:3b").
	ModelName() string

	// BaseURL retourne l'URL de base du serveur (ex : "http://localhost:11434").
	BaseURL() string

	// ListModels retourne la liste des modèles disponibles sur ce provider.
	ListModels(ctx context.Context) ([]string, error)

	// Chat envoie une conversation et retourne un canal de morceaux de texte (streaming).
	// Le canal est fermé à la fin de la réponse ou en cas d'erreur.
	Chat(ctx context.Context, messages []Message) (<-chan Chunk, error)

	// Ping vérifie que le provider est joignable.
	Ping(ctx context.Context) bool
}

// Status représente l'état de connexion d'un provider, utilisé par l'IHM.
type Status string

const (
	// StatusConnected : provider joignable et modèle chargé (icône verte).
	StatusConnected Status = "connected"
	// StatusDetecting : sondage en cours (icône orange).
	StatusDetecting Status = "detecting"
	// StatusDisconnected : aucun provider disponible (icône rouge).
	StatusDisconnected Status = "disconnected"
)

// ProviderInfo contient les métadonnées d'un provider détecté,
// destinées à l'affichage dans l'IHM.
type ProviderInfo struct {
	Name    string   `json:"name"`
	BaseURL string   `json:"base_url"`
	Models  []string `json:"models"`
	Status  Status   `json:"status"`
}
