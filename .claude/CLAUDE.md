# CLAUDE.md — Instructions pour Claude Code sur le projet MJAYNRI

## Contexte du projet

**MJAYNRI** est une plateforme JDR multi-système alimentée par IA locale (Ollama, LM Studio) ou cloud.
Langage principal : **Go 1.22** · Serveur HTTP : **Gin** · Frontend : Vanilla JS + CSS custom.

## Architecture en couches

```
cmd/mjaynri/          → Point d'entrée (main.go)
internal/config/      → Configuration via variables d'environnement
internal/llm/         → Abstraction providers IA (interface + adapters)
internal/server/      → Serveur HTTP Gin + routes
internal/server/handlers/ → Handlers HTTP (1 fichier par responsabilité)
web/templates/        → Templates HTML (Go html/template)
web/static/           → CSS et JS (embarqués dans le binaire via embed.FS)
docs/dev/             → Documentation développeur
docs/user/            → Documentation utilisateur final
```

## Règles impératives

### Git & Branches
- **Ne jamais committer directement sur `main`**
- **Ne jamais committer directement sur `dev`**
- Toute feature → branche `feature/<description-courte>` depuis `dev`
- Toute correction → branche `fix/<description-courte>` depuis `dev`
- PR `feature → dev` : nécessite 100% des tests verts
- PR `dev → main` : nécessite validation explicite de l'utilisateur
- Messages de commit en anglais, format : `type: description courte`
  - Types valides : `feat`, `fix`, `test`, `docs`, `refactor`, `chore`

### Tests — Règle stricte
- **Zéro régression tolérée** : chaque PR doit passer `make test`
- Tout nouveau package Go doit avoir son `_test.go`
- Tout nouveau handler HTTP doit avoir un test `httptest`
- Les tests utilisent des serveurs `httptest.NewServer` pour mocker Ollama/LMStudio
- Jamais de `t.Skip()` sans justification commentée

### Qualité du code
- Commenter toutes les fonctions exportées (godoc)
- Commenter les blocs non-évidents avec des commentaires `//`
- Erreurs wrappées avec `fmt.Errorf("contexte: %w", err)`
- Pas de `panic` hors du `main` et des tests
- Pas de variable globale mutable (sauf registre de providers thread-safe)
- Utiliser `context.Context` pour toutes les opérations réseau

### Frontend
- Pas de framework JS externe (vanilla ES2020+)
- Pas de CDN externe (tout doit fonctionner offline)
- Accessibilité : attributs `aria-`, rôles ARIA, contraste suffisant
- CSS : utiliser les variables CSS définies dans `main.css`

## Commandes utiles

```bash
make build      # Compile le binaire
make test       # Lance tous les tests + race detector
make lint       # golangci-lint
make run        # Lance en développement (port 8080)
make coverage   # Rapport de couverture HTML
```

## Ajouter un nouveau provider LLM

1. Créer `internal/llm/<nom>.go` implémentant l'interface `Provider`
2. Créer `internal/llm/<nom>_test.go` avec un `httptest.Server` mock
3. Enregistrer dans `resolver.go` → méthode `Detect()`
4. Documenter dans `docs/dev/llm-providers.md`

## Ajouter une nouvelle route HTTP

1. Créer le handler dans `internal/server/handlers/<feature>.go`
2. Créer les tests dans `internal/server/handlers/<feature>_test.go`
3. Enregistrer la route dans `internal/server/routes.go`
4. Documenter dans `docs/dev/api.md`

## Variables d'environnement

| Variable               | Défaut                    | Description              |
|------------------------|---------------------------|--------------------------|
| MJAYNRI_PORT           | 8080                      | Port HTTP                |
| MJAYNRI_OLLAMA_URL     | http://localhost:11434    | URL serveur Ollama       |
| MJAYNRI_LMSTUDIO_URL   | http://localhost:1234     | URL serveur LM Studio    |
| MJAYNRI_PROBE_TIMEOUT  | 2 (secondes)              | Timeout sondage provider |
| MJAYNRI_DEFAULT_MODEL  | (premier disponible)      | Modèle à précharger      |

## Plan de développement V0 → V1

### V0 (branche actuelle)
- [x] Détection Ollama + LM Studio au démarrage
- [x] Page d'accueil avec badge de connexion
- [x] Chat streaming avec le provider actif

### V1 (prochaine)
- [ ] Gestion multi-système de jeu (D&D 5e, Rolemaster)
- [ ] Page de gestion des personnages (PJ/PNJ)
- [ ] Démarrage/reprise de parties
- [ ] Journal de session
