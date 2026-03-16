package llm

import (
	"context"
	"sync"
)

// Manager gère le provider LLM actif et permet de le changer à chaud depuis l'IHM.
// Il est thread-safe : plusieurs goroutines peuvent appeler ses méthodes simultanément.
type Manager struct {
	mu         sync.RWMutex
	active     Provider
	available  []DetectedProvider
	status     Status
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
		Name:    m.active.Name(),
		BaseURL: m.active.BaseURL(),
		Models:  models,
		Status:  m.status,
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
