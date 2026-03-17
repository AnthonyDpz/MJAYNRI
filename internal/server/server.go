// Package server configure et démarre le serveur HTTP de MJAYNRI.
// Il utilise Gin comme framework HTTP et expose les routes définies dans routes.go.
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AnthonyDpz/MJAYNRI/internal/config"
	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
	"github.com/AnthonyDpz/MJAYNRI/internal/server/handlers"
)

// Server encapsule le serveur HTTP et ses dépendances.
type Server struct {
	http    *http.Server
	cfg     config.ServerConfig
	manager *llm.Manager
}

// New crée un Server Gin configuré avec toutes les routes V0.
func New(cfg config.ServerConfig, manager *llm.Manager) *Server {
	// Mode release en production (pas de logs debug Gin)
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Middlewares globaux
	engine.Use(gin.Recovery())              // Récupère les panics
	engine.Use(loggerMiddleware())          // Log structuré des requêtes
	engine.Use(securityHeadersMiddleware()) // En-têtes de sécurité HTTP

	// Injection des handlers avec leurs dépendances
	h := handlers.New(manager)
	registerRoutes(engine, h)

	return &Server{
		cfg:     cfg,
		manager: manager,
		http: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      engine,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// Start démarre le serveur HTTP et bloque jusqu'à ce que ctx soit annulé.
// Implémente un arrêt gracieux : attend que les connexions actives se terminent.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		log.Printf("[SERVER] Démarrage sur http://localhost:%s", s.cfg.Port)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("serveur HTTP : %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("[SERVER] Arrêt gracieux en cours…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	}
}

// loggerMiddleware retourne un middleware Gin qui logue chaque requête.
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[HTTP] %s %s %d %v",
			c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), time.Since(start))
	}
}

// securityHeadersMiddleware ajoute des en-têtes de sécurité HTTP standard.
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
