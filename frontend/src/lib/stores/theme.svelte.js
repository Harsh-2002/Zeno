import { THEMES, UI_COLORS } from '../constants/themes.js';

export const themeState = $state({
  themeId: 'dark',
  fontSize: 14,
  cursorStyle: 'block',
  cursorBlink: true,
  lineHeight: 1.1,
  scrollback: 100000,
  copyOnSelect: false,
  // Server-only (read-only in UI)
  port: 8080,
  shell: '',
});

let saveTimer = null;

export async function loadConfig() {
  try {
    const res = await fetch('/api/config');
    if (!res.ok) return;
    const cfg = await res.json();
    themeState.themeId = cfg.theme || 'dark';
    themeState.fontSize = cfg.fontSize || 14;
    themeState.cursorStyle = cfg.cursorStyle || 'block';
    themeState.cursorBlink = cfg.cursorBlink !== false;
    themeState.lineHeight = cfg.lineHeight || 1.1;
    themeState.scrollback = cfg.scrollback || 100000;
    themeState.copyOnSelect = cfg.copyOnSelect || false;
    themeState.port = cfg.port || 8080;
    themeState.shell = cfg.shell || '';
    applyThemeCSS(themeState.themeId);
  } catch (e) {
    // Fallback to defaults
  }
}

function saveToServer() {
  clearTimeout(saveTimer);
  saveTimer = setTimeout(async () => {
    try {
      await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          theme: themeState.themeId,
          fontSize: themeState.fontSize,
          cursorStyle: themeState.cursorStyle,
          cursorBlink: themeState.cursorBlink,
          lineHeight: themeState.lineHeight,
          scrollback: themeState.scrollback,
          copyOnSelect: themeState.copyOnSelect,
        })
      });
    } catch (e) {}
  }, 300);
}

export function getTheme() { return THEMES[themeState.themeId]; }

export function setTheme(id) {
  if (!THEMES[id]) return;
  themeState.themeId = id;
  applyThemeCSS(id);
  saveToServer();
}

export function setFontSize(size) {
  size = Math.max(8, Math.min(32, size));
  themeState.fontSize = size;
  saveToServer();
}

export function setCursorStyle(style) {
  themeState.cursorStyle = style;
  saveToServer();
}

export function setCursorBlink(blink) {
  themeState.cursorBlink = blink;
  saveToServer();
}

export function setLineHeight(lh) {
  themeState.lineHeight = lh;
  saveToServer();
}

export function setScrollback(sb) {
  themeState.scrollback = sb;
  saveToServer();
}

export function setCopyOnSelect(val) {
  themeState.copyOnSelect = val;
  saveToServer();
}

export function applyThemeCSS(themeId) {
  const t = THEMES[themeId || themeState.themeId];
  const u = UI_COLORS[themeId || themeState.themeId];
  const r = document.documentElement.style;
  r.setProperty('--bg-primary', t.background);
  r.setProperty('--bg-secondary', u.secondary);
  r.setProperty('--bg-tertiary', u.secondary);
  r.setProperty('--bg-hover', u.hover);
  r.setProperty('--border', u.border);
  r.setProperty('--border-subtle', u.borderSubtle);
  r.setProperty('--border-focus', u.focus);
  r.setProperty('--surface-overlay', u.secondary);
  r.setProperty('--match-bg', u.matchBg);
  r.setProperty('--match-active-bg', u.activeMatchBg);
  r.setProperty('--match-active-border', u.activeMatchBorder);
}
