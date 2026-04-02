<script>
  import { onMount } from 'svelte';
  import TabBar from './components/TabBar.svelte';
  import TerminalArea from './components/TerminalArea.svelte';
  import ContextMenu from './components/ContextMenu.svelte';
  import Settings from './components/Settings.svelte';
  import FileBrowser from './components/FileBrowser.svelte';
  import Toast from './components/Toast.svelte';
  import { tabState, createTab, createTabFromSaved, closeTab, switchToTab, switchRelative, getActiveTab } from './lib/stores/tabs.svelte.js';
  import { themeState, setTheme, setFontSize, applyThemeCSS, loadConfig } from './lib/stores/theme.svelte.js';
  import { saveWorkspace, loadWorkspace } from './lib/stores/workspace.svelte.js';
  import { paneState, getPane, setFocusedPane, paneResources, nextPaneId } from './lib/stores/panes.svelte.js';
  import { forEachPane, findPaneById, findParent } from './lib/utils/paneTree.js';
  import { isModKey } from './lib/utils/shortcuts.js';
  import { sendResize, sendData } from './lib/utils/ws.js';

  let toastMessage = $state('');
  let toastVisible = $state(false);
  let toastTimer = null;
  let contextMenu = $state({ visible: false, x: 0, y: 0 });
  let settingsVisible = $state(false);
  let fileBrowserVisible = $state(false);

  function showToast(msg) {
    toastMessage = msg; toastVisible = true;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { toastVisible = false; }, 1500);
  }

  function fitActiveTab() {
    const tab = getActiveTab();
    if (!tab) return;
    forEachPane(tab.rootNode, (pane) => {
      const r = getPane(pane.id);
      if (r) { r.fitAddon.fit(); sendResize(r.getWs ? r.getWs() : r.ws, r.term.cols, r.term.rows); }
    });
  }

  function splitPane(direction) {
    const tab = getActiveTab();
    if (!tab || !paneState.focusedPaneId) return;
    const target = findPaneById(tab.rootNode, paneState.focusedPaneId);
    if (!target) return;
    const newPaneId = nextPaneId();
    const newPane = { type: 'pane', id: newPaneId };
    // Use the original target reference — NOT a copy.
    // Copying ({ ...target }) causes Svelte to remount the TerminalPane,
    // which calls term.open() again on an already-opened terminal, breaking input.
    const splitNode = { type: 'split', direction, ratio: 0.5, first: target, second: newPane };
    const parentInfo = findParent(tab.rootNode, target);
    if (parentInfo) parentInfo.parent[parentInfo.which] = splitNode;
    else tab.rootNode = splitNode;
    setTimeout(() => { fitActiveTab(); setFocusedPane(newPaneId); saveWorkspace(); }, 50);
  }

  function closePane() {
    const tab = getActiveTab();
    if (!tab || !paneState.focusedPaneId) return;
    if (tab.rootNode.type === 'pane') { closeTab(tab.id); return; }
    const pane = findPaneById(tab.rootNode, paneState.focusedPaneId);
    if (!pane) return;
    const parentInfo = findParent(tab.rootNode, pane);
    if (!parentInfo) return;
    const sibling = parentInfo.which === 'first' ? parentInfo.parent.second : parentInfo.parent.first;
    const r = getPane(pane.id);
    if (r) { r.closed = true; if (r.ws.readyState <= 1) r.ws.close(); r.term.dispose(); paneResources.delete(pane.id); }
    const gp = findParent(tab.rootNode, parentInfo.parent);
    if (gp) gp.parent[gp.which] = sibling;
    else tab.rootNode = sibling;
    let nextFocus = sibling.type === 'pane' ? sibling.id : null;
    if (!nextFocus) forEachPane(sibling, (p) => { if (!nextFocus) nextFocus = p.id; });
    setTimeout(() => { fitActiveTab(); if (nextFocus) setFocusedPane(nextFocus); saveWorkspace(); }, 50);
  }

  function handleContextAction(action, data) {
    const pane = getPane(paneState.focusedPaneId);
    if (action === 'settings') { settingsVisible = true; return; }
    if (action === 'browseFiles') { fileBrowserVisible = true; return; }
    if (action === 'closePane') { closePane(); return; }
    if (!pane) return;
    switch (action) {
      case 'copy': if (pane.term.hasSelection()) navigator.clipboard.writeText(pane.term.getSelection()); break;
      case 'paste': navigator.clipboard.readText().then(t => sendData(pane.getWs ? pane.getWs() : pane.ws, t)); break;
      case 'selectAll': pane.term.selectAll(); break;
      case 'clear': pane.term.clear(); break;
      case 'saveOutput': {
        const buf = pane.term.buffer.active;
        let text = '';
        for (let i = 0; i < buf.length; i++) {
          const line = buf.getLine(i);
          if (line) text += line.translateToString(true) + '\n';
        }
        const blob = new Blob([text], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url; a.download = 'terminal-output.txt';
        a.click(); URL.revokeObjectURL(url);
        showToast('Output saved');
        break;
      }
      case 'search': if (pane.toggleSearch) pane.toggleSearch(); break;
      case 'splitV': splitPane('vertical'); break;
      case 'splitH': splitPane('horizontal'); break;
    }
  }

  function handleContextMenu(e) {
    e.preventDefault();
    const paneEl = e.target.closest('[data-pane-id]');
    if (paneEl) setFocusedPane(parseInt(paneEl.dataset.paneId));
    contextMenu = { visible: true, x: e.clientX, y: e.clientY };
  }

  function handleSettingsOpen() {
    settingsVisible = !settingsVisible;
  }

  function handleFileBrowserOpen() {
    fileBrowserVisible = !fileBrowserVisible;
  }

  function getActiveSessionId() {
    const pane = getPane(paneState.focusedPaneId);
    return pane?.sessionId || '';
  }

  $effect(() => { applyThemeCSS(themeState.themeId); });
  $effect(() => {
    const tab = tabState.tabs.find(t => t.id === tabState.activeTabId);
    if (tab) document.title = `${tab.title} \u2014 Zeno`;
  });
  $effect(() => { if (tabState.activeTabId) setTimeout(fitActiveTab, 20); });

  onMount(() => {
    let resizeTimer;
    const onResize = () => { clearTimeout(resizeTimer); resizeTimer = setTimeout(fitActiveTab, 100); };
    window.addEventListener('resize', onResize);

    const onKeyDown = (e) => {
      const mod = isModKey(e);
      if (mod && e.key === 'f') { e.preventDefault(); const r = getPane(paneState.focusedPaneId); if (r && r.toggleSearch) r.toggleSearch(); return; }
      if (mod && e.key === 'd' && !e.shiftKey) { e.preventDefault(); splitPane('vertical'); return; }
      if (mod && e.shiftKey && e.key === 'D') { e.preventDefault(); splitPane('horizontal'); return; }
      if (mod && e.key === 'w' && !e.shiftKey) { e.preventDefault(); const tab = getActiveTab(); if (tab && tab.rootNode.type !== 'pane') closePane(); else if (tabState.activeTabId) closeTab(tabState.activeTabId); return; }
      if (mod && e.shiftKey && e.key === 'T') { e.preventDefault(); createTab(); return; }
      if (mod && e.shiftKey && e.key === 'W') { e.preventDefault(); if (tabState.activeTabId) closeTab(tabState.activeTabId); return; }
      if (mod && e.shiftKey && (e.key === '[' || e.key === '{')) { e.preventDefault(); switchRelative(-1); return; }
      if (mod && e.shiftKey && (e.key === ']' || e.key === '}')) { e.preventDefault(); switchRelative(1); return; }
      if (mod && e.key >= '1' && e.key <= '9') { const i = parseInt(e.key) - 1; if (i < tabState.tabs.length) { e.preventDefault(); switchToTab(tabState.tabs[i].id); } return; }
      if (mod && (e.key === '=' || e.key === '+')) { e.preventDefault(); setFontSize(themeState.fontSize + 1); showToast(`Font size: ${themeState.fontSize}px`); setTimeout(fitActiveTab, 50); return; }
      if (mod && e.key === '-') { e.preventDefault(); setFontSize(themeState.fontSize - 1); showToast(`Font size: ${themeState.fontSize}px`); setTimeout(fitActiveTab, 50); return; }
      if (mod && e.key === '0') { e.preventDefault(); setFontSize(14); showToast('Font size: 14px'); setTimeout(fitActiveTab, 50); return; }
      if (mod && e.key === ',') { e.preventDefault(); settingsVisible = !settingsVisible; return; }
      if (e.key === 'Escape') { contextMenu.visible = false; settingsVisible = false; fileBrowserVisible = false; }
    };
    window.addEventListener('keydown', onKeyDown);

    const onDocClick = (e) => {
      if (e.button !== 0) return;
      if (contextMenu.visible && !e.target.closest('.context-menu')) contextMenu.visible = false;
    };
    document.addEventListener('mousedown', onDocClick);

    // Reload config when window regains focus (picks up manual YAML edits)
    const onFocus = () => loadConfig();
    window.addEventListener('focus', onFocus);

    // Load config from YAML, then create first tab
    loadConfig().then(async () => {
      if (themeState.persistSessions) {
        const ws = await loadWorkspace();
        if (ws && ws.tabs && ws.tabs.length > 0) {
          ws.tabs.forEach(t => createTabFromSaved(t.title, t.rootNode));
          if (ws.activeTabIndex >= 0 && ws.activeTabIndex < tabState.tabs.length) {
            switchToTab(tabState.tabs[ws.activeTabIndex].id);
          }
          return;
        }
      }
      createTab();
    });

    return () => {
      window.removeEventListener('resize', onResize);
      window.removeEventListener('keydown', onKeyDown);
      document.removeEventListener('mousedown', onDocClick);
      window.removeEventListener('focus', onFocus);
    };
  });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="app" oncontextmenu={handleContextMenu}>
  <TabBar onOpenSettings={handleSettingsOpen} onOpenFiles={handleFileBrowserOpen} />
  <TerminalArea onClosePane={(paneId) => { setFocusedPane(paneId); closePane(); }} />
  <ContextMenu visible={contextMenu.visible} x={contextMenu.x} y={contextMenu.y}
    onAction={handleContextAction} onClose={() => contextMenu.visible = false} />
  <Settings visible={settingsVisible} onClose={() => settingsVisible = false} />
  <FileBrowser visible={fileBrowserVisible} onClose={() => fileBrowserVisible = false} sessionId={getActiveSessionId()} />
  <Toast message={toastMessage} visible={toastVisible} />
</div>

<style>
  .app { display: flex; flex-direction: column; height: 100%; width: 100%; }
</style>
