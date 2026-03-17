package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SwitchRequest est le corps JSON attendu pour POST /api/switch.
type SwitchRequest struct {
	// Provider est le nom exact du provider cible (ex : "Ollama", "LM Studio").
	Provider string `json:"provider" binding:"required"`
	// Model est le nom exact du modèle à activer (ex : "llama3.2:3b").
	Model string `json:"model" binding:"required"`
}

// Switch change le provider et le modèle actifs à chaud, sans redémarrer le serveur.
//
// Route : POST /api/switch
// Corps  : {"provider": "Ollama", "model": "llama3.2:3b"}
// Réponse: StatusResponse JSON avec le nouvel état
//
// Retourne 400 si le provider ou le modèle est inconnu.
func (h *Handler) Switch(c *gin.Context) {
	var req SwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Corps JSON invalide : " + err.Error(),
		})
		return
	}

	if err := h.manager.Switch(req.Provider, req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retourner le nouvel état pour que le frontend mette à jour le badge immédiatement
	info := h.manager.StatusInfo()
	c.JSON(http.StatusOK, infoToStatusResponse(info))
}
