// Package handlers contient les handlers HTTP de MJAYNRI.
// Chaque handler reçoit le Manager LLM en dépendance et n'a pas d'état global.
package handlers

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
	"github.com/AnthonyDpz/MJAYNRI/web"
)

// Handler regroupe toutes les dépendances partagées par les handlers HTTP.
type Handler struct {
	manager   *llm.Manager
	templates *template.Template
	// resolver permet de relancer la détection des providers (bouton ↻ dans l'IHM).
	// Peut être nil — dans ce cas Refresh retourne le statut actuel sans re-scan.
	resolver *llm.Resolver
}

// New crée un Handler avec le Manager LLM et parse les templates HTML.
// Panic si les templates sont invalides (détecté au démarrage, pas à la requête).
func New(manager *llm.Manager) *Handler {
	// Parse tous les templates depuis le système de fichiers embarqué
	tmpl := template.Must(
		template.New("").ParseFS(web.TemplateFS, "templates/*.html"),
	)
	return &Handler{
		manager:   manager,
		templates: tmpl,
	}
}

// WithResolver injecte un Resolver pour permettre le rafraîchissement des providers.
func (h *Handler) WithResolver(r *llm.Resolver) *Handler {
	h.resolver = r
	return h
}

// StaticFiles retourne le http.FileSystem pour les assets statiques embarqués.
// fs.Sub retire le préfixe "static/" du embed.FS afin que les URLs
// /static/css/main.css soient correctement résolues vers css/main.css.
func (h *Handler) StaticFiles() http.FileSystem {
	sub, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		panic("web: impossible d'accéder au sous-dossier static: " + err.Error())
	}
	return http.FS(sub)
}
