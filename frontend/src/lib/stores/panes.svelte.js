const paneResources = new Map();

export const paneState = $state({
  counter: 0,
  focusedPaneId: null,
});

export function nextPaneId() { return ++paneState.counter; }
export function setFocusedPane(id) { paneState.focusedPaneId = id; }
export function getFocusedPaneId() { return paneState.focusedPaneId; }

export function registerPane(id, resources) { paneResources.set(id, resources); }
export function unregisterPane(id) { paneResources.delete(id); }
export function getPane(id) { return paneResources.get(id); }

export { paneResources };
