import { tabState } from './tabs.svelte.js';
import { paneResources } from './panes.svelte.js';
import { themeState } from './theme.svelte.js';
import { forEachPane } from '../utils/paneTree.js';

let saveTimer = null;

function serializeNode(node) {
  if (node.type === 'pane') {
    const res = paneResources.get(node.id);
    return { type: 'pane', sessionId: res?.sessionId || '' };
  }
  return {
    type: 'split',
    direction: node.direction,
    ratio: node.ratio,
    first: serializeNode(node.first),
    second: serializeNode(node.second),
  };
}

export function saveWorkspace() {
  if (!themeState.persistSessions) return;
  clearTimeout(saveTimer);
  saveTimer = setTimeout(async () => {
    const activeIdx = tabState.tabs.findIndex(t => t.id === tabState.activeTabId);
    const state = {
      version: 1,
      activeTabIndex: Math.max(0, activeIdx),
      tabs: tabState.tabs.map(t => ({
        title: t.title,
        rootNode: serializeNode(t.rootNode),
      })),
    };
    try {
      await fetch('/api/workspace', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(state),
      });
    } catch (e) {}
  }, 500);
}

export async function loadWorkspace() {
  try {
    const res = await fetch('/api/workspace');
    if (!res.ok) return null;
    return await res.json();
  } catch (e) {
    return null;
  }
}
