/**
 * chat.js — Logique frontend du chat MJAYNRI
 *
 * Responsabilités :
 *  1. Envoi des messages via POST /api/chat (fetch + ReadableStream)
 *  2. Affichage des réponses en streaming (SSE côté serveur)
 *  3. Mise à jour du badge de connexion via GET /api/status
 *  4. Auto-resize du textarea, raccourcis clavier
 *
 * Pas de framework — vanilla JS ES2020+.
 * Aucune dépendance externe.
 */

'use strict';

// ── Éléments DOM ──────────────────────────────────────────────────────────────
const chatMessages = document.getElementById('chat-messages');
const chatForm     = document.getElementById('chat-form');
const textarea     = document.getElementById('chat-textarea');
const sendBtn      = document.getElementById('send-btn');
const typingEl     = document.getElementById('chat-typing');
const statusDot    = document.getElementById('status-dot');
const statusLabel  = document.getElementById('status-label');
const modelSelect  = document.getElementById('model-select');
const refreshBtn   = document.getElementById('refresh-btn');
const stopBtn      = document.getElementById('stop-btn');

/** Historique de la conversation envoyé à l'API à chaque tour. */
const conversation = [];

/** Indique si un streaming est en cours (bloque l'envoi concurrent). */
let isStreaming = false;

/** AbortController du fetch SSE courant — null quand aucun stream n'est actif. */
let abortController = null;

// ── Initialisation ────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  setupTextarea();
  setupForm();
  setupStopButton();
  setupRefresh();
  setupModelSelect();
  pollStatus();        // Premier check du statut au chargement
  loadModelSelector(); // Charger la liste des modèles et afficher le select si nécessaire
});

// ── Textarea auto-resize ──────────────────────────────────────────────────────

function setupTextarea() {
  textarea.addEventListener('input', autoResize);
  textarea.addEventListener('keydown', (e) => {
    // Entrée seule → envoyer ; Maj+Entrée → saut de ligne
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (!isStreaming) submitMessage();
    }
  });
}

/** Ajuste la hauteur du textarea à son contenu (max 160px défini en CSS). */
function autoResize() {
  textarea.style.height = 'auto';
  textarea.style.height = Math.min(textarea.scrollHeight, 160) + 'px';
}

// ── Envoi du formulaire ───────────────────────────────────────────────────────

function setupForm() {
  chatForm.addEventListener('submit', (e) => {
    e.preventDefault();
    if (!isStreaming) submitMessage();
  });
}

async function submitMessage() {
  const content = textarea.value.trim();
  if (!content) return;

  // Effacer et réinitialiser la saisie
  textarea.value = '';
  autoResize();
  isStreaming = true;
  setInputDisabled(true); // Masque Envoyer, affiche Stop

  // Supprimer l'écran de bienvenue au premier message
  const welcome = document.querySelector('.chat-welcome');
  if (welcome) welcome.remove();

  // Ajouter le message utilisateur à l'historique
  conversation.push({ role: 'user', content });
  appendMessage('user', content);

  // Préparer la bulle de réponse IA (sera remplie en streaming)
  const aiMessageEl = appendMessage('ai', '');
  aiMessageEl.querySelector('.message__bubble').classList.add('streaming');
  showTyping(true);

  try {
    await streamResponse(aiMessageEl);
  } catch (err) {
    showError(aiMessageEl, err.message);
  } finally {
    isStreaming = false;
    aiMessageEl.querySelector('.message__bubble').classList.remove('streaming');
    showTyping(false);
    setInputDisabled(false); // Masque Stop, affiche Envoyer
    textarea.focus();
  }
}

// ── Streaming de la réponse ───────────────────────────────────────────────────

/**
 * Envoie la conversation à POST /api/chat et affiche la réponse en streaming.
 * Utilise fetch + ReadableStream pour lire le flux SSE sans EventSource
 * (EventSource ne supporte pas la méthode POST).
 *
 * @param {HTMLElement} aiEl - Élément du message IA à remplir
 */
async function streamResponse(aiEl) {
  const bubble = aiEl.querySelector('.message__bubble');
  let fullContent = '';

  // Un AbortController par requête — permet d'annuler via le bouton Stop ou Échap.
  abortController = new AbortController();

  try {
    const response = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ messages: conversation }),
      signal: abortController.signal,
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({ error: 'Erreur inconnue' }));
      throw new Error(data.error || `HTTP ${response.status}`);
    }

    // Lire le flux SSE ligne par ligne
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    // Suivi du type d'événement SSE courant (persiste entre les lectures)
    let pendingEvent = null;

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop(); // Garder la ligne incomplète en buffer

      for (const line of lines) {
        // Ligne "event: <type>" → mémoriser le type, attendre la ligne data
        if (line.startsWith('event: ')) {
          pendingEvent = line.slice(7).trim();
          continue;
        }

        // Ligne "data: <payload>" → traiter selon le type d'événement en attente
        if (line.startsWith('data: ')) {
          const data = line.slice(6);

          if (pendingEvent === 'done') {
            conversation.push({ role: 'assistant', content: fullContent });
            return;
          }

          if (pendingEvent === 'error') {
            throw new Error(data || 'Erreur de streaming IA');
          }

          if (pendingEvent === 'chunk') {
            fullContent += data;
            bubble.replaceChildren(renderMarkdown(fullContent));
            scrollToBottom();
          }

          pendingEvent = null;
        }
      }
    }

    // Fin propre du canal sans event: done (stream tronqué côté serveur)
    if (fullContent) {
      conversation.push({ role: 'assistant', content: fullContent });
    }
  } catch (err) {
    if (err.name === 'AbortError') {
      // Arrêt volontaire (bouton Stop ou Échap) — pas d'erreur affichée.
      // La réponse partielle reste visible mais n'est pas sauvegardée dans
      // l'historique pour éviter que le modèle reboucle sur un contexte tronqué.
      return;
    }
    throw err; // Propager les vraies erreurs vers submitMessage
  } finally {
    abortController = null;
  }
}

// ── Helpers DOM ───────────────────────────────────────────────────────────────

/**
 * Crée et ajoute une bulle de message dans la zone de chat.
 * @param {'user'|'ai'|'error'} role
 * @param {string} content - Texte initial (peut être vide pour le streaming)
 * @returns {HTMLElement} L'élément message créé
 */
function appendMessage(role, content) {
  const msg = document.createElement('div');
  msg.className = `message message--${role}`;

  const avatar = document.createElement('div');
  avatar.className = 'message__avatar';
  avatar.textContent = role === 'user' ? 'V' : '⚔';
  avatar.setAttribute('aria-hidden', 'true');

  const bubble = document.createElement('div');
  bubble.className = 'message__bubble';
  bubble.textContent = content;

  msg.appendChild(avatar);
  msg.appendChild(bubble);
  chatMessages.appendChild(msg);
  scrollToBottom();

  return msg;
}

/** Affiche un message d'erreur dans la bulle IA. */
function showError(aiEl, message) {
  aiEl.className = 'message message--error';
  aiEl.querySelector('.message__bubble').textContent = '⚠ Erreur : ' + message;
}

/** Active/désactive la zone de saisie et bascule entre Send ↔ Stop. */
function setInputDisabled(disabled) {
  textarea.disabled = disabled;
  sendBtn.disabled = disabled;
  sendBtn.hidden   = disabled;  // Masqué pendant le streaming
  stopBtn.hidden   = !disabled; // Visible pendant le streaming
}

/** Affiche ou masque l'indicateur "l'IA écrit…". */
function showTyping(visible) {
  typingEl.hidden = !visible;
  if (visible) scrollToBottom();
}

/** Fait défiler la zone des messages vers le bas. */
function scrollToBottom() {
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

// ── Badge de connexion + sélecteur de modèle ─────────────────────────────────

/**
 * Interroge /api/models et peuple le <select> groupé par provider.
 * Si un seul modèle total est disponible, affiche le label statique.
 * Toutes les options sont créées via createElement — aucun innerHTML.
 */
async function loadModelSelector() {
  try {
    const res = await fetch('/api/models');
    if (!res.ok) return;
    const data = await res.json();
    const providers = data.providers || [];

    const totalModels = providers.reduce((sum, p) => sum + (p.models?.length || 0), 0);

    if (totalModels <= 1) {
      showSelectOrLabel(false);
      return;
    }

    // Vider le select proprement (sans innerHTML)
    modelSelect.replaceChildren();

    const currentVal = `${modelSelect.dataset.currentProvider || ''}|${modelSelect.dataset.currentModel || ''}`;

    providers.forEach(p => {
      const group = document.createElement('optgroup');
      group.label = p.provider;

      (p.models || []).forEach(model => {
        const opt = document.createElement('option');
        opt.value = `${p.provider}|${model}`;
        opt.textContent = model;
        if (opt.value === currentVal) opt.selected = true;
        group.appendChild(opt);
      });

      modelSelect.appendChild(group);
    });

    showSelectOrLabel(true);
  } catch (_) {
    // Silencieux — le label statique reste affiché
  }
}

/**
 * Bascule entre le label statique et le <select>.
 * @param {boolean} useSelect
 */
function showSelectOrLabel(useSelect) {
  statusLabel.hidden = useSelect;
  modelSelect.hidden = !useSelect;
}

/**
 * Met à jour le badge de connexion (point coloré + label ou select).
 * @param {{ color: string, provider: string, model: string }} data
 */
function updateBadge(data) {
  // Point coloré
  statusDot.className = 'connection-badge__dot';
  statusDot.classList.add(`status-${data.color}`);

  // Label (visible si déconnecté ou un seul modèle)
  statusLabel.textContent = data.model
    ? `${data.provider} — ${data.model}`
    : data.provider;

  // Mémoriser pour la prochaine reconstruction du select
  modelSelect.dataset.currentModel    = data.model    || '';
  modelSelect.dataset.currentProvider = data.provider || '';

  // Synchroniser l'option sélectionnée si le select est visible
  if (!modelSelect.hidden && data.provider && data.model) {
    const target = `${data.provider}|${data.model}`;
    for (const opt of modelSelect.options) {
      opt.selected = (opt.value === target);
    }
  }
}

/** Interroge /api/status et met à jour le badge. */
async function pollStatus() {
  try {
    const res = await fetch('/api/status');
    if (!res.ok) return;
    const data = await res.json();
    updateBadge(data);
  } catch (_) {
    // Silencieux — l'état déjà affiché reste valide
  }
}

/**
 * Écoute les changements du <select> et appelle POST /api/switch.
 */
function setupModelSelect() {
  modelSelect.addEventListener('change', async () => {
    const [provider, model] = modelSelect.value.split('|');
    if (!provider || !model) return;

    modelSelect.classList.add('switching');
    try {
      const res = await fetch('/api/switch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider, model }),
      });
      if (res.ok) {
        const data = await res.json();
        updateBadge(data);
      }
    } catch (_) {
      // Silencieux
    } finally {
      modelSelect.classList.remove('switching');
    }
  });
}

/**
 * Branche le bouton Stop et le raccourci Échap.
 * Un clic (ou Échap) annule le fetch SSE en cours via AbortController.
 */
function setupStopButton() {
  stopBtn.addEventListener('click', stopStreaming);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && isStreaming) stopStreaming();
  });
}

/** Annule le streaming en cours si un AbortController est actif. */
function stopStreaming() {
  if (abortController) abortController.abort();
}

function setupRefresh() {
  refreshBtn.addEventListener('click', async () => {
    refreshBtn.classList.add('spinning');
    try {
      const res = await fetch('/api/refresh', { method: 'POST' });
      if (res.ok) {
        const data = await res.json();
        updateBadge(data);
        // Recharger la liste des modèles après un refresh (nouveaux providers possibles)
        await loadModelSelector();
      }
    } catch (_) {
      // Silencieux
    } finally {
      refreshBtn.classList.remove('spinning');
    }
  });
}

// Rafraîchissement automatique du statut toutes les 30 secondes
setInterval(pollStatus, 30_000);
