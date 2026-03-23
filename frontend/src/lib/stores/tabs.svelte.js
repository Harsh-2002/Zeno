import { nextPaneId } from './panes.svelte.js';

// Use an object to hold reassignable state — Svelte 5 module exports
// can't re-export reassigned $state, but object property mutations are fine.
export const tabState = $state({
  tabs: [],
  activeTabId: null,
  counter: 0,
});

export function getActiveTab() {
  return tabState.tabs.find(t => t.id === tabState.activeTabId) || null;
}

export function createTab() {
  const id = ++tabState.counter;
  const paneId = nextPaneId();
  const tab = {
    id,
    title: `Terminal ${id}`,
    rootNode: { type: 'pane', id: paneId },
    hasUnread: false,
  };
  tabState.tabs.push(tab);
  tabState.activeTabId = id;
  return tab;
}

export function switchToTab(id) {
  const tab = tabState.tabs.find(t => t.id === id);
  if (!tab) return;
  tabState.activeTabId = id;
  tab.hasUnread = false;
}

export function closeTab(id) {
  const idx = tabState.tabs.findIndex(t => t.id === id);
  if (idx === -1) return;
  if (tabState.tabs.length === 1) createTab();
  tabState.tabs.splice(idx, 1);
  if (tabState.activeTabId === id && tabState.tabs.length > 0) {
    switchToTab(tabState.tabs[Math.min(idx, tabState.tabs.length - 1)].id);
  }
}

export function switchRelative(offset) {
  if (tabState.tabs.length <= 1) return;
  const ci = tabState.tabs.findIndex(t => t.id === tabState.activeTabId);
  if (ci === -1) return;
  const ni = (ci + offset + tabState.tabs.length) % tabState.tabs.length;
  switchToTab(tabState.tabs[ni].id);
}

export function reorderTabs(fromId, toId, side) {
  const fromIdx = tabState.tabs.findIndex(t => t.id === fromId);
  const toIdx = tabState.tabs.findIndex(t => t.id === toId);
  if (fromIdx === -1 || toIdx === -1 || fromIdx === toIdx) return;
  const moved = tabState.tabs.splice(fromIdx, 1)[0];
  let insertIdx = toIdx > fromIdx ? toIdx - 1 : toIdx;
  if (side === 'right') insertIdx++;
  tabState.tabs.splice(insertIdx, 0, moved);
}

export function setTabTitle(id, title) {
  const tab = tabState.tabs.find(t => t.id === id);
  if (tab) tab.title = title;
}

export function markUnread(tabId) {
  const tab = tabState.tabs.find(t => t.id === tabId);
  if (tab && tabId !== tabState.activeTabId) tab.hasUnread = true;
}
