/**
 * markdown.js — Renderer Markdown + balises <think> pour MJAYNRI
 *
 * Construit le DOM avec createElement/appendChild uniquement.
 * Aucun innerHTML sur du contenu non maîtrisé → immunisé XSS.
 *
 * Éléments pris en charge :
 *   Blocs  : <think>, ```code```, # h1-h3, listes -, *, 1., blockquotes >, paragraphes
 *   Inline : **gras**, *italique*, _italique_, `code inline`
 *
 * Usage :
 *   bubble.replaceChildren(renderMarkdown(fullText));
 */

'use strict';

// ── Point d'entrée public ─────────────────────────────────────────────────────

/**
 * Convertit du texte Markdown + balises <think> en un DocumentFragment DOM.
 * @param {string} text - Texte brut reçu du modèle IA
 * @returns {DocumentFragment}
 */
function renderMarkdown(text) {
  const frag = document.createDocumentFragment();

  for (const seg of splitThinkSegments(text)) {
    if (seg.isThink) {
      frag.appendChild(buildThinkBlock(seg.text));
    } else if (seg.text.length > 0) {
      appendBlockElements(frag, seg.text);
    }
  }

  return frag;
}

// ── Gestion des blocs <think> ─────────────────────────────────────────────────

/**
 * Découpe le texte en segments think / non-think.
 * Une balise <think> sans </think> (streaming en cours) est traitée comme
 * un bloc think s'étendant jusqu'à la fin du texte.
 * @param {string} text
 * @returns {{text: string, isThink: boolean}[]}
 */
function splitThinkSegments(text) {
  const segments = [];
  const thinkPattern = /<think>([\s\S]*?)(?:<\/think>|$)/gi;
  let lastIndex = 0;
  let found;

  // Parcourir toutes les occurrences de <think>…</think>
  while ((found = thinkPattern.test(text) && text.match(thinkPattern)) !== null) {
    break; // On utilise matchAll à la place
  }

  // Utilisation de matchAll (ES2020, disponible dans tous les navigateurs modernes)
  for (const m of text.matchAll(/<think>([\s\S]*?)(?:<\/think>|$)/gi)) {
    if (m.index > lastIndex) {
      segments.push({ text: text.slice(lastIndex, m.index), isThink: false });
    }
    segments.push({ text: m[1], isThink: true });
    lastIndex = m.index + m[0].length;
  }

  if (lastIndex < text.length) {
    segments.push({ text: text.slice(lastIndex), isThink: false });
  }

  return segments.length ? segments : [{ text, isThink: false }];
}

/**
 * Construit le bloc visuel pour le contenu <think>.
 * Affiché en italique grisé avec un label discret.
 * @param {string} content
 * @returns {HTMLElement}
 */
function buildThinkBlock(content) {
  const wrapper = document.createElement('div');
  wrapper.className = 'think-block';

  const label = document.createElement('span');
  label.className = 'think-block__label';
  label.textContent = '💭 Réflexion interne';
  wrapper.appendChild(label);

  const body = document.createElement('div');
  body.className = 'think-block__body';
  const trimmed = content.trim();
  if (trimmed) appendBlockElements(body, trimmed);
  wrapper.appendChild(body);

  return wrapper;
}

// ── Rendu bloc (paragraphes, listes, code…) ───────────────────────────────────

/**
 * Analyse le texte ligne par ligne et produit des éléments bloc HTML.
 * @param {Node} container - Nœud DOM cible (DocumentFragment ou Element)
 * @param {string} text
 */
function appendBlockElements(container, text) {
  const lines = text.split('\n');
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trimStart();

    // ── Bloc de code ``` ────────────────────────────────────────────────────
    if (trimmed.startsWith('```')) {
      const lang = trimmed.slice(3).trim();
      const codeLines = [];
      i++;
      while (i < lines.length && !lines[i].trimStart().startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // sauter la ligne de fermeture

      const pre = document.createElement('pre');
      pre.className = 'md-code-block';
      const code = document.createElement('code');
      if (lang) code.dataset.lang = lang;
      code.textContent = codeLines.join('\n');
      pre.appendChild(code);
      container.appendChild(pre);
      continue;
    }

    // ── Titres # ## ### ─────────────────────────────────────────────────────
    const headingM = trimmed.match(/^(#{1,3})\s+(.+)/);
    if (headingM) {
      const level = headingM[1].length;
      const h = document.createElement(`h${level}`);
      h.className = `md-h${level}`;
      appendInlineNodes(h, headingM[2]);
      container.appendChild(h);
      i++;
      continue;
    }

    // ── Liste non ordonnée ─────────────────────────────────────────────────
    if (/^\s*[-*]\s+/.test(line)) {
      const ul = document.createElement('ul');
      ul.className = 'md-list';
      while (i < lines.length && /^\s*[-*]\s+/.test(lines[i])) {
        const li = document.createElement('li');
        appendInlineNodes(li, lines[i].replace(/^\s*[-*]\s+/, ''));
        ul.appendChild(li);
        i++;
      }
      container.appendChild(ul);
      continue;
    }

    // ── Liste ordonnée ─────────────────────────────────────────────────────
    if (/^\s*\d+\.\s+/.test(line)) {
      const ol = document.createElement('ol');
      ol.className = 'md-list';
      while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
        const li = document.createElement('li');
        appendInlineNodes(li, lines[i].replace(/^\s*\d+\.\s+/, ''));
        ol.appendChild(li);
        i++;
      }
      container.appendChild(ol);
      continue;
    }

    // ── Blockquote > ───────────────────────────────────────────────────────
    if (/^>\s?/.test(line)) {
      const bq = document.createElement('blockquote');
      bq.className = 'md-blockquote';
      const bqLines = [];
      while (i < lines.length && /^>\s?/.test(lines[i])) {
        bqLines.push(lines[i].replace(/^>\s?/, ''));
        i++;
      }
      appendBlockElements(bq, bqLines.join('\n'));
      container.appendChild(bq);
      continue;
    }

    // ── Ligne vide ─────────────────────────────────────────────────────────
    if (line.trim() === '') {
      i++;
      continue;
    }

    // ── Paragraphe (lignes consécutives non-bloc) ───────────────────────────
    const paraLines = [];
    while (
      i < lines.length &&
      lines[i].trim() !== '' &&
      !lines[i].trimStart().startsWith('```') &&
      !/^#{1,3}\s/.test(lines[i].trimStart()) &&
      !/^\s*[-*]\s+/.test(lines[i]) &&
      !/^\s*\d+\.\s+/.test(lines[i]) &&
      !/^>\s?/.test(lines[i])
    ) {
      paraLines.push(lines[i]);
      i++;
    }

    if (paraLines.length > 0) {
      const p = document.createElement('p');
      p.className = 'md-para';
      appendInlineNodes(p, paraLines.join('\n'));
      container.appendChild(p);
    }
  }
}

// ── Rendu inline (gras, italique, code) ──────────────────────────────────────

/**
 * Analyse le texte inline et ajoute les nœuds formatés à container.
 * Recherche le premier motif présent (position la plus faible) à chaque passe.
 * @param {Node} container
 * @param {string} text
 */
function appendInlineNodes(container, text) {
  // Motifs inline — traités du plus prioritaire au moins
  const patterns = [
    { pattern: /\*\*(.+?)\*\*/,  tag: 'strong', cls: null             },
    { pattern: /\*(.+?)\*/,      tag: 'em',     cls: null             },
    { pattern: /_(.+?)_/,        tag: 'em',     cls: null             },
    { pattern: /`([^`\n]+)`/,    tag: 'code',   cls: 'md-inline-code' },
  ];

  let remaining = text;

  while (remaining.length > 0) {
    let bestMatch = null;
    let bestIndex = Infinity;
    let bestPat   = null;

    // Trouver la première occurrence parmi tous les motifs
    for (const pat of patterns) {
      const m = remaining.match(pat.pattern);
      if (m && m.index < bestIndex) {
        bestMatch = m;
        bestIndex = m.index;
        bestPat   = pat;
      }
    }

    if (!bestMatch) {
      container.appendChild(document.createTextNode(remaining));
      break;
    }

    // Texte brut avant le motif
    if (bestIndex > 0) {
      container.appendChild(document.createTextNode(remaining.slice(0, bestIndex)));
    }

    // Élément formaté
    const el = document.createElement(bestPat.tag);
    if (bestPat.cls) el.className = bestPat.cls;
    el.textContent = bestMatch[1];
    container.appendChild(el);

    remaining = remaining.slice(bestIndex + bestMatch[0].length);
  }
}
