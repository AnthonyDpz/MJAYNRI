# Guide d'utilisation — MJAYNRI

## Page d'accueil

### Badge de connexion (en haut à droite)

Le badge indique l'état de connexion à votre IA :

| Couleur | Signification |
|---|---|
| 🟢 Vert | Provider connecté, modèle prêt |
| 🟠 Orange | Détection en cours… |
| 🔴 Rouge | Aucun provider disponible |

**Exemples d'affichage :**
- `● Ollama — llama3.2:3b` → Ollama connecté avec le modèle llama3.2
- `● LM Studio — meta-llama-3.2-3b` → LM Studio avec un modèle Meta
- `● Aucun provider disponible` → Démarrer Ollama ou LM Studio

Le bouton **↻** à droite du badge force un nouveau scan des providers locaux.
Utile si vous démarrez Ollama après avoir lancé MJAYNRI.

### Chat

1. **Écrivez votre message** dans le champ de saisie en bas
2. **Envoyez** avec `Entrée` (ou le bouton ➤)
3. Pour un **saut de ligne**, utilisez `Maj + Entrée`
4. La réponse de l'IA **s'affiche progressivement** (streaming)

### Conseils d'utilisation

- Donnez du **contexte** dans vos messages ("Je joue un elfe ranger dans un monde médiéval fantasy")
- Vous pouvez **poser des questions de règles**, demander des descriptions de scènes, générer des PNJ, etc.
- L'historique de la conversation est maintenu pendant la session (rechargement de page = reset)

## Résolution de problèmes

### Le badge est rouge
- Vérifier qu'Ollama ou LM Studio est démarré
- Cliquer sur **↻** pour relancer la détection
- Vérifier avec `ollama list` ou depuis l'interface LM Studio que des modèles sont disponibles

### La réponse s'arrête au milieu
- Le modèle a peut-être manqué de mémoire — essayer un modèle plus petit
- Vérifier les logs d'Ollama : `ollama logs`

### Port déjà utilisé
```bash
MJAYNRI_PORT=9000 ./bin/mjaynri
```
