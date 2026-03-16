package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
)

// ChatRequest est le corps JSON attendu pour POST /api/chat.
type ChatRequest struct {
	// Messages est l'historique complet de la conversation (contexte + dernier message).
	Messages []llm.Message `json:"messages" binding:"required,min=1"`
}

// Chat reçoit une conversation et stream la réponse de l'IA via SSE.
//
// Route : POST /api/chat
// Content-Type réponse : text/event-stream
//
// Format SSE émis :
//
//	event: chunk
//	data: <fragment de texte>
//
//	event: done
//	data: {"finish":"stop"}
//
//	event: error
//	data: <message d'erreur>
func (h *Handler) Chat(c *gin.Context) {
	// Vérifier qu'un provider est disponible
	provider := h.manager.Active()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Aucun provider IA disponible. Démarrez Ollama ou LM Studio.",
		})
		return
	}

	// Parser le corps de la requête
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corps JSON invalide : " + err.Error()})
		return
	}

	// Configurer les en-têtes SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Désactive le buffering nginx si présent

	// Démarrer le streaming depuis le provider
	stream, err := provider.Chat(c.Request.Context(), req.Messages)
	if err != nil {
		sendSSEEvent(c.Writer, "error", err.Error())
		return
	}

	// Relayer les chunks au client
	c.Stream(func(w io.Writer) bool {
		chunk, ok := <-stream
		if !ok {
			// Canal fermé normalement
			return false
		}
		if chunk.Err != nil {
			sendSSEEvent(w, "error", chunk.Err.Error())
			return false
		}
		if chunk.Done {
			sendSSEEvent(w, "done", `{"finish":"stop"}`)
			return false
		}
		sendSSEEvent(w, "chunk", chunk.Content)
		return true
	})
}

// sendSSEEvent écrit un événement SSE formaté dans w.
// Format : "event: <name>\ndata: <data>\n\n"
func sendSSEEvent(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
