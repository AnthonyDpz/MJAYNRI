package server

import (
	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/server/handlers"
)

// registerRoutes déclare toutes les routes HTTP de l'application.
// Chaque route est documentée avec son rôle et son handler associé.
//
// Routes V0 :
//
//	GET  /              → Page d'accueil (IHM principale)
//	GET  /api/status    → Statut de connexion IA (JSON) — utilisé par le badge
//	POST /api/chat      → Envoi d'un message et réponse streaming (SSE)
//	GET  /api/models    → Liste des providers + modèles disponibles (JSON)
//	POST /api/refresh   → Force un re-scan des providers locaux
//	POST /api/switch    → Change le provider/modèle actif à chaud
func registerRoutes(engine *gin.Engine, h *handlers.Handler) {
	// Assets statiques (CSS, JS) — servis depuis les fichiers embarqués (embed.FS)
	engine.StaticFS("/static", h.StaticFiles())

	// Page principale
	engine.GET("/", h.Homepage)

	// API
	api := engine.Group("/api")
	{
		api.GET("/status", h.Status)
		api.POST("/chat", h.Chat)
		api.GET("/models", h.Models)
		api.POST("/refresh", h.Refresh)
		api.POST("/switch", h.Switch)
	}
}
