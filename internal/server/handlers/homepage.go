package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// HomepageData contient les données transmises au template de la page d'accueil.
type HomepageData struct {
	AppName string     // Nom de l'application (affiché dans le <title> et le header)
	Status  StatusData // Données du badge de connexion
}

// StatusData est la représentation IHM du statut de connexion IA.
type StatusData struct {
	Color string // Classe CSS : "status-green", "status-orange", "status-red"
	Label string // Texte affiché, ex : "Ollama — llama3.2:3b"
	Raw   string // Valeur brute du statut pour les attributs data- HTML
}

// Homepage sert la page d'accueil avec le statut IA courant.
// Route : GET /
func (h *Handler) Homepage(c *gin.Context) {
	info := h.manager.StatusInfo()
	data := HomepageData{
		AppName: "MJAYNRI",
		Status:  providerInfoToStatusData(info),
	}

	if err := h.templates.ExecuteTemplate(c.Writer, "homepage.html", data); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}

// providerInfoToStatusData convertit les infos du provider en données d'affichage.
func providerInfoToStatusData(info llm.ProviderInfo) StatusData {
	switch info.Status {
	case llm.StatusConnected:
		model := "aucun modèle"
		if len(info.Models) > 0 {
			model = info.Models[0]
		}
		return StatusData{
			Color: "status-green",
			Label: info.Name + " — " + model,
			Raw:   string(llm.StatusConnected),
		}
	case llm.StatusDetecting:
		return StatusData{
			Color: "status-orange",
			Label: "Détection en cours…",
			Raw:   string(llm.StatusDetecting),
		}
	default:
		return StatusData{
			Color: "status-red",
			Label: "Aucun provider disponible",
			Raw:   string(llm.StatusDisconnected),
		}
	}
}
