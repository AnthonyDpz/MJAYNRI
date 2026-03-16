// Commande mjaynri — point d'entrée de l'application MJAYNRI.
//
// Séquence de démarrage :
//  1. Charger la configuration (variables d'environnement)
//  2. Sonder les providers LLM locaux (Ollama, LM Studio)
//  3. Démarrer le serveur HTTP
//  4. Attendre SIGINT/SIGTERM pour un arrêt gracieux
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/AnthonyDpz/MJAYNRI/internal/config"
	"github.com/AnthonyDpz/MJAYNRI/internal/llm"
	"github.com/AnthonyDpz/MJAYNRI/internal/server"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("[MJAYNRI] Démarrage…")

	// ── 1. Configuration ──────────────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Détection des providers LLM ───────────────────────────────────────
	ctx := context.Background()
	resolver := llm.NewResolver(cfg.LLM)

	detected, err := resolver.Detect(ctx)
	if err != nil {
		// Non-fatal : l'application démarre même sans provider (mode déconnecté)
		log.Printf("[LLM] Avertissement : détection partielle : %v", err)
	}

	manager := llm.NewManager(detected)

	switch manager.Status() {
	case llm.StatusConnected:
		info := manager.StatusInfo()
		log.Printf("[LLM] Provider actif : %s — %s", info.Name, info.Models[0])
	default:
		log.Println("[LLM] Aucun provider IA local — l'IHM démarrera en mode déconnecté")
	}

	// ── 3. Serveur HTTP ───────────────────────────────────────────────────────
	srv := server.New(cfg.Server, manager)

	// ── 4. Arrêt gracieux sur SIGINT / SIGTERM ────────────────────────────────
	shutdownCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("[SERVER] Accessible sur http://localhost:%s", cfg.Server.Port)

	if err := srv.Start(shutdownCtx); err != nil {
		log.Fatalf("[SERVER] Erreur fatale : %v", err)
	}

	log.Println("[MJAYNRI] Arrêt propre.")
}
