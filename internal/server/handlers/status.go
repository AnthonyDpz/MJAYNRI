package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// StatusResponse est la réponse JSON de GET /api/status.
// Utilisée par le badge de connexion pour les mises à jour en temps réel.
type StatusResponse struct {
	Status    string   `json:"status"`     // "connected" | "detecting" | "disconnected"
	Color     string   `json:"color"`      // "green" | "orange" | "red"
	Provider  string   `json:"provider"`   // Nom du provider, ex : "Ollama"
	Model     string   `json:"model"`      // Modèle actif, ex : "llama3.2:3b"
	Models    []string `json:"models"`     // Tous les modèles disponibles
	BaseURL   string   `json:"base_url"`   // URL du serveur
}

// Status retourne l'état courant de la connexion IA en JSON.
// Route : GET /api/status
func (h *Handler) Status(c *gin.Context) {
	info := h.manager.StatusInfo()
	c.JSON(http.StatusOK, infoToStatusResponse(info))
}

// Models retourne la liste complète des providers et modèles disponibles.
// Route : GET /api/models
func (h *Handler) Models(c *gin.Context) {
	type modelEntry struct {
		Provider string   `json:"provider"`
		BaseURL  string   `json:"base_url"`
		Models   []string `json:"models"`
		Active   bool     `json:"active"`
	}

	available := h.manager.Available()
	active := h.manager.Active()
	entries := make([]modelEntry, 0, len(available))

	for _, dp := range available {
		isActive := active != nil && dp.Provider.Name() == active.Name()
		entries = append(entries, modelEntry{
			Provider: dp.Provider.Name(),
			BaseURL:  dp.Provider.BaseURL(),
			Models:   dp.Models,
			Active:   isActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"providers": entries})
}

// Refresh force un re-scan des providers LLM locaux.
// Route : POST /api/refresh
// Utilisé par le bouton "Rafraîchir" de l'IHM quand l'utilisateur
// démarre Ollama ou LM Studio après le lancement de l'application.
func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if h.resolver != nil {
		h.manager.Refresh(ctx, h.resolver)
	}
	info := h.manager.StatusInfo()
	c.JSON(http.StatusOK, infoToStatusResponse(info))
}

// infoToStatusResponse convertit un ProviderInfo en StatusResponse JSON.
func infoToStatusResponse(info llm.ProviderInfo) StatusResponse {
	resp := StatusResponse{
		Status:  string(info.Status),
		Models:  info.Models,
		BaseURL: info.BaseURL,
	}

	switch info.Status {
	case llm.StatusConnected:
		resp.Color = "green"
		resp.Provider = info.Name
		resp.Model = info.ActiveModel // Modèle réellement actif, pas forcément le premier de la liste
	case llm.StatusDetecting:
		resp.Color = "orange"
		resp.Provider = "Détection…"
	default:
		resp.Color = "red"
		resp.Provider = "Déconnecté"
	}

	return resp
}
