# Stratégie de tests

## Principes

1. **Zéro régression** — `make test` doit rester vert sur toutes les branches
2. **Tests rapides** — pas de sleep, pas de dépendances réseau réelles
3. **Tests déterministes** — même résultat à chaque exécution
4. **Mocks légers** — `httptest.NewServer` pour simuler Ollama/LMS, pas de bibliothèque de mock

## Commandes

```bash
make test       # Tests + race detector (obligatoire avant PR)
make coverage   # Rapport HTML de couverture
go test -v ./internal/llm/...   # Tests d'un package spécifique
go test -run TestOllama ./...   # Un test précis
```

## Structure des tests par couche

### `internal/config`
- Tests unitaires de `Load()` via `os.Setenv`/`os.Unsetenv`
- Vérifier les valeurs par défaut et la lecture depuis l'environnement

### `internal/llm`
- `resolver_test.go` : utilise `httptest.NewServer` simulant Ollama et LM Studio
- `ollama_test.go` : teste Ping, ListModels, Chat streaming avec un faux serveur Ollama
- `lmstudio_test.go` : idem pour LM Studio (format OpenAI SSE)
- Tester les cas d'erreur : server injoignable, réponse invalide, contexte annulé

### `internal/server/handlers`
- Utiliser `gin.SetMode(gin.TestMode)` dans `TestMain` ou `init()`
- `httptest.NewRecorder()` pour capturer les réponses
- `mockProvider` (dans `mock_provider_test.go`) pour injecter un provider de test

## Exemple de test handler

```go
func TestStatus_Connected(t *testing.T) {
    mock := &mockProvider{name: "Ollama", model: "llama3.2:3b"}
    manager := llm.NewManager([]llm.DetectedProvider{{Provider: mock, Models: []string{"llama3.2:3b"}}})

    engine := gin.New()
    h := handlers.New(manager)
    engine.GET("/api/status", h.Status)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
    engine.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    // ... vérifier le JSON
}
```

## Couverture cible

| Package | Cible |
|---|---|
| `internal/config` | 90%+ |
| `internal/llm` | 85%+ |
| `internal/server/handlers` | 80%+ |

## Tests d'IHM (V1+)

Pour la V1, des tests Playwright seront ajoutés dans `tests/e2e/` :
- Test du badge de connexion (vert/orange/rouge)
- Test du formulaire de chat (envoi, réponse streaming)
- Test du bouton rafraîchir
