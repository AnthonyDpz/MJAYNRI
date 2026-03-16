# Installation de MJAYNRI

## Prérequis

### 1. Un provider IA (au choix)

**Option A — Ollama** (recommandé, gratuit)
```bash
# macOS / Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Télécharger un modèle (ex : llama3.2 ~2GB)
ollama pull llama3.2:3b

# Vérifier qu'Ollama tourne
ollama list
```

**Option B — LM Studio**
1. Télécharger sur [lmstudio.ai](https://lmstudio.ai)
2. Installer un modèle depuis l'interface
3. Activer le serveur local : `Local Server → Start Server`

### 2. Go 1.22+ (pour build depuis les sources)

```bash
# macOS
brew install go

# Linux (Ubuntu/Debian)
sudo apt install golang-go

# Vérifier
go version
```

## Installation depuis les sources

```bash
git clone https://github.com/AnthonyDpz/MJAYNRI.git
cd MJAYNRI
make build
./bin/mjaynri
```

L'application est accessible sur **http://localhost:8080**.

## Configuration (optionnel)

Toutes les options ont des valeurs par défaut. Modifier uniquement si nécessaire :

```bash
# Port HTTP (défaut : 8080)
export MJAYNRI_PORT=9000

# URL Ollama si sur une machine distante
export MJAYNRI_OLLAMA_URL=http://192.168.1.10:11434

# Modèle par défaut
export MJAYNRI_DEFAULT_MODEL=llama3.2:3b

./bin/mjaynri
```

## Mise à jour

```bash
git pull
make build
./bin/mjaynri
```
