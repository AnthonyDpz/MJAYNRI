# MJAYNRI — Maître de Jeu IA

> Plateforme JDR multi-système alimentée par IA locale (Ollama, LM Studio) ou cloud.

[![CI](https://github.com/AnthonyDpz/MJAYNRI/actions/workflows/ci.yml/badge.svg)](https://github.com/AnthonyDpz/MJAYNRI/actions/workflows/ci.yml)

## Démarrage rapide

### Prérequis
- Go 1.22+
- [Ollama](https://ollama.ai) **ou** [LM Studio](https://lmstudio.ai) démarré avec au moins un modèle chargé

### Lancer l'application

```bash
git clone https://github.com/AnthonyDpz/MJAYNRI.git
cd MJAYNRI
make run
# Ouvrir http://localhost:8080
```

L'application **détecte automatiquement** Ollama (`:11434`) et LM Studio (`:1234`) au démarrage. Si aucun provider n'est trouvé, l'interface démarre en mode déconnecté et un bouton de rafraîchissement permet de retenter la connexion.

## V0 — Fonctionnalités

| Feature | État |
|---|---|
| Détection Ollama / LM Studio | ✅ |
| Badge de connexion (vert/orange/rouge) | ✅ |
| Chat streaming avec l'IA | ✅ |
| Rafraîchissement de connexion à chaud | ✅ |

## Roadmap

Voir [docs/dev/roadmap.md](docs/dev/roadmap.md) pour le plan complet.

## Documentation

| Document | Audience |
|---|---|
| [Architecture](docs/dev/architecture.md) | Développeurs |
| [Démarrage développeur](docs/dev/getting-started.md) | Développeurs |
| [Contribution & Branches](docs/dev/contributing.md) | Développeurs |
| [Stratégie de tests](docs/dev/testing.md) | Développeurs |
| [Installation](docs/user/installation.md) | Utilisateurs |
| [Guide d'utilisation](docs/user/guide.md) | Utilisateurs |

## Structure du projet

```
MJAYNRI/
├── cmd/mjaynri/          # Point d'entrée
├── internal/
│   ├── config/           # Configuration
│   ├── llm/              # Abstraction providers IA
│   └── server/           # Serveur HTTP + handlers
├── web/
│   ├── templates/        # Templates HTML
│   └── static/           # CSS, JS (embarqués dans le binaire)
├── docs/
│   ├── dev/              # Documentation développeur
│   └── user/             # Documentation utilisateur
└── .claude/              # Configuration Claude Code
```

## Licence

MIT
