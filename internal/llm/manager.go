package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager gère le provider LLM actif et permet de le changer à chaud depuis l'IHM.
// Il est thread-safe : plusieurs goroutines peuvent appeler ses méthodes simultanément.
type Manager struct {
	mu        sync.RWMutex
	active    Provider
	available []DetectedProvider
	status    Status
}

// NewManager crée un Manager à partir des providers détectés par le Resolver.
// Le premier provider détecté devient l'actif par défaut.
func NewManager(detected []DetectedProvider) *Manager {
	m := &Manager{
		available: detected,
		status:    StatusDisconnected,
	}
	if len(detected) > 0 {
		m.active = detected[0].Provider
		m.status = StatusConnected
	}
	return m
}

// Active retourne le provider actuellement sélectionné, ou nil si aucun.
func (m *Manager) Active() Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
}

// Status retourne l'état de connexion global.
func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// StatusInfo retourne les infos d'affichage pour l'IHM (badge de connexion).
func (m *Manager) StatusInfo() ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.active == nil {
		return ProviderInfo{Status: StatusDisconnected}
	}

	var models []string
	for _, dp := range m.available {
		if dp.Provider.Name() == m.active.Name() {
			models = dp.Models
			break
		}
	}

	return ProviderInfo{
		Name:        m.active.Name(),
		BaseURL:     m.active.BaseURL(),
		ActiveModel: m.active.ModelName(),
		Models:      models,
		Status:      m.status,
	}
}

// SetActive change le provider actif par son index dans la liste des disponibles.
func (m *Manager) SetActive(index int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index < 0 || index >= len(m.available) {
		return false
	}
	m.active = m.available[index].Provider
	m.status = StatusConnected
	return true
}

// Available retourne tous les providers détectés (pour un menu de sélection).
func (m *Manager) Available() []DetectedProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.available
}

// Switch change le provider actif ET le modèle à chaud, sans redémarrer le serveur.
//
// Règles :
//   - providerName doit correspondre exactement à Provider.Name() d'un provider disponible.
//   - modelName doit figurer dans la liste Models de ce provider.
//
// Un nouveau Provider est instancié avec le modèle demandé afin que ModelName() reflète
// immédiatement le choix de l'utilisateur sans affecter les autres providers en mémoire.
func (m *Manager) Switch(providerName, modelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, dp := range m.available {
		if dp.Provider.Name() != providerName {
			continue
		}

		// Vérifier que le modèle demandé est bien disponible sur ce provider
		found := false
		for _, mod := range dp.Models {
			if mod == modelName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("modèle %q non disponible sur %s", modelName, providerName)
		}

		// Créer un nouveau provider avec le modèle sélectionné.
		// On récupère l'URL de base depuis le provider existant pour éviter tout couplage
		// avec la configuration — le Manager n'a pas besoin de connaître les URLs directement.
		baseURL := dp.Provider.BaseURL()
		const chatTimeout = 120 * time.Second

		var newProvider Provider
		switch providerName {
		case "Ollama":
			newProvider = NewOllamaProvider(baseURL, modelName, chatTimeout)
		case "LM Studio":
			newProvider = NewLMStudioProvider(baseURL, modelName, chatTimeout)
		default:
			return fmt.Errorf("provider %q : type inconnu, impossible de recréer l'instance", providerName)
		}

		m.active = newProvider
		m.status = StatusConnected
		return nil
	}

	return fmt.Errorf("provider %q non disponible", providerName)
}

// Refresh re-sonde les providers et met à jour l'état. Appelé périodiquement
// ou à la demande de l'utilisateur depuis l'IHM.
func (m *Manager) Refresh(ctx context.Context, resolver *Resolver) {
	m.mu.Lock()
	m.status = StatusDetecting
	m.mu.Unlock()

	detected, _ := resolver.Detect(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.available = detected
	if len(detected) > 0 {
		// Essayer de conserver le provider actif s'il est toujours disponible
		if m.active != nil {
			for _, dp := range detected {
				if dp.Provider.Name() == m.active.Name() {
					m.active = dp.Provider
					m.status = StatusConnected
					return
				}
			}
		}
		// Sinon prendre le premier disponible
		m.active = detected[0].Provider
		m.status = StatusConnected
	} else {
		m.active = nil
		m.status = StatusDisconnected
	}
}
