# Guide de contribution

## Stratégie de branches

```
main              ← Production. Merge uniquement depuis dev, après validation utilisateur.
  └── dev         ← Intégration. Toutes les features mergent ici.
        ├── feature/v0-llm-detection
        ├── feature/v0-homepage
        └── fix/badge-refresh
```

### Règles strictes
- ❌ Ne jamais committer directement sur `main` ou `dev`
- ✅ Créer une branche `feature/<description>` ou `fix/<description>` depuis `dev`
- ✅ PR `feature → dev` : CI doit être verte à 100%
- ✅ PR `dev → main` : validation explicite de l'utilisateur requise

## Workflow type

```bash
# 1. Créer la branche feature depuis dev
git checkout dev && git pull
git checkout -b feature/ma-feature

# 2. Développer + tester
make test          # Doit être vert avant tout commit

# 3. Committer (messages en anglais)
git add internal/llm/nouveau.go internal/llm/nouveau_test.go
git commit -m "feat: add nouveau LLM adapter"

# 4. Pousser et ouvrir une PR vers dev
git push -u origin feature/ma-feature
gh pr create --base dev --title "feat: add nouveau LLM adapter"
```

## Format des messages de commit

```
<type>: <description courte en anglais>

[corps optionnel]

[références issues optionnelles]
```

Types valides : `feat` · `fix` · `test` · `docs` · `refactor` · `chore`

Exemples :
```
feat: add LM Studio streaming adapter
fix: resolve ollama timeout on model list
test: add handler tests for /api/status
docs: document LLM provider interface
```

## Checklist avant PR

- [ ] `make test` passe sans erreur
- [ ] `make lint` passe sans erreur
- [ ] Toute nouvelle fonction exportée est documentée (godoc)
- [ ] Tout nouveau fichier Go a son `_test.go`
- [ ] La PR touche une seule fonctionnalité (pas de commits mélangés)

## Protection de branche `main`

Le merge vers `main` nécessite :
1. CI GitHub Actions verte
2. Validation **explicite** de l'utilisateur dans le chat
3. Aucun commit direct (force-push interdit)
