# Roadmap MJAYNRI

## V0 — Fondation IA (branche actuelle)

| Feature | Statut | Branche |
|---|---|---|
| Détection Ollama au démarrage | ✅ | feature/v0-llm |
| Détection LM Studio au démarrage | ✅ | feature/v0-llm |
| Badge de connexion (vert/orange/rouge) | ✅ | feature/v0-homepage |
| Page d'accueil avec chat | ✅ | feature/v0-homepage |
| Streaming SSE de la réponse IA | ✅ | feature/v0-homepage |
| Rafraîchissement provider à chaud | ✅ | feature/v0-homepage |
| Tests unitaires (config, llm, handlers) | ✅ | feature/v0-tests |
| CI GitHub Actions | ✅ | feature/v0-ci |
| Documentation dev + user | ✅ | feature/v0-docs |

## V1 — Multi-système & Gestion de partie

| Feature | Statut | Priorité |
|---|---|---|
| Abstraction GameSystem (interface Go) | 🔲 | Haute |
| Adapter D&D 5e | 🔲 | Haute |
| Adapter Rolemaster | 🔲 | Haute |
| gm-adapter.md par système (instructions MJ IA) | 🔲 | Haute |
| Création de personnage (PJ/PNJ) | 🔲 | Moyenne |
| Gestion de partie (démarrer/reprendre) | 🔲 | Moyenne |
| Journal de session | 🔲 | Moyenne |
| Historique consultable | 🔲 | Moyenne |
| Sélection provider/modèle depuis l'IHM | 🔲 | Basse |

## V2 — Contenu & Features avancées

| Feature | Statut | Priorité |
|---|---|---|
| Bestiaire par système | 🔲 | Moyenne |
| Équipement & inventaire | 🔲 | Moyenne |
| Magie (sorts, PP, réalms) | 🔲 | Moyenne |
| Génération d'images (fal.ai) | 🔲 | Basse |
| Support Anthropic Claude (cloud fallback) | 🔲 | Basse |
| Mode multi-joueurs (sessions partagées) | 🔲 | Basse |

## Principes de priorisation

1. Les features bloquantes pour le gameplay passent avant le polish
2. Chaque feature = une branche + PR + tests
3. Rien ne merge sur main sans validation utilisateur explicite
