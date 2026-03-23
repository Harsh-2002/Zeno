const isMac = typeof navigator !== 'undefined' && /Mac/.test(navigator.platform);

export function isModKey(e) {
  return isMac ? e.metaKey : e.ctrlKey;
}

export function modLabel() {
  return isMac ? '\u2318' : 'Ctrl+';
}

export function shortcutLabel(key) {
  return `${modLabel()}${key}`;
}
