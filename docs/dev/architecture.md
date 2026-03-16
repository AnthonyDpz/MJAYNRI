# Architecture MJAYNRI

## Vue d'ensemble

MJAYNRI est structuré en couches indépendantes. Le cœur métier (`internal/`) ne connaît pas le framework HTTP, et le serveur HTTP ne connaît pas les détails des providers IA. Cette séparation facilite les tests et l'évolution.

```
┌─────────────────────────────────────────────────────┐
│  Interfaces utilisateur                             │
│  sw-web (Gin/HTMX)   ·   sw-dm (CLI)               │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP
┌──────────────────────▼──────────────────────────────┐
│  internal/server                                    │
│  routes.go  ·  handlers/  ·  server.go              │
└──────────────────────┬──────────────────────────────┘
                       │ Appels directs Go
┌──────────────────────▼──────────────────────────────┐
│  internal/llm  (Abstraction providers)              │
│  Provider interface  ·  Manager  ·  Resolver        │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │ Ollama     │  │  LM Studio   │  │  + futur    │ │
│  │ adapter    │  │  adapter     │  │  adapter    │ │
│  └────────────┘  └──────────────┘  └─────────────┘ │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP local
┌──────────────────────▼──────────────────────────────┐
│  Services externes                                  │
│  localhost:11434 (Ollama)  ·  localhost:1234 (LMS)  │
└─────────────────────────────────────────────────────┘
```

## Détection des providers au démarrage

```
main.go
  → Resolver.Detect(ctx)
      → probeOllama()   GET :11434/api/tags    → [modèles]
      → probeLMStudio() GET :1234/v1/models    → [modèles]
  → Manager.New(detected)
      → active = premier provider disponible
  → server.New(cfg, manager)
  → server.Start(ctx)
```

## Flux d'un message chat

```
Navigateur         Serveur             Provider IA
   │                  │                    │
   │── POST /api/chat ──>│                 │
   │                  │── provider.Chat() ─>│
   │                  │<── chan Chunk ──────│
   │<── SSE chunk ────│                    │
   │<── SSE chunk ────│                    │
   │<── SSE done ─────│                    │
```

## Package `internal/llm`

| Fichier | Rôle |
|---|---|
| `provider.go` | Interface `Provider` + types `Status`, `ProviderInfo` |
| `message.go` | Types `Message`, `Chunk`, constantes `Role` |
| `resolver.go` | Sondage au démarrage, sélection du modèle par défaut |
| `manager.go` | Provider actif, changement à chaud, refresh |
| `ollama.go` | Adapter Ollama (API native NDJSON) |
| `lmstudio.go` | Adapter LM Studio (API OpenAI-compatible SSE) |

## Assets web (embed.FS)

Les templates HTML et les assets statiques (CSS, JS) sont embarqués dans le binaire Go via `//go:embed`. Cela permet un **déploiement monobinaire** : copier `bin/mjaynri` suffit.

```go
//go:embed templates
var TemplateFS embed.FS

//go:embed static
var StaticFS embed.FS
```

## Décisions d'architecture notables

| Décision | Justification |
|---|---|
| Pas de base de données pour V0 | Simplicité, état en mémoire suffisant |
| Pas de framework frontend (React/Vue) | Pas de build step, embarquable facilement |
| SSE plutôt que WebSocket pour le chat | Plus simple, unidirectionnel suffit |
| `embed.FS` pour les assets | Déploiement monobinaire |
| `context.Context` partout | Annulation propre, timeouts sans goroutines orphelines |
