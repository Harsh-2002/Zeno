<script>
  import Tab from './Tab.svelte';
  import { tabState, createTab, switchToTab, closeTab, reorderTabs, setTabTitle } from '../lib/stores/tabs.svelte.js';

  let { onOpenSettings } = $props();
  let draggingTabId = $state(null);

  function handleDrop(e, tabId) {
    e.preventDefault();
    if (!draggingTabId || draggingTabId === tabId) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const side = e.clientX < rect.left + rect.width / 2 ? 'left' : 'right';
    reorderTabs(draggingTabId, tabId, side);
    draggingTabId = null;
  }
</script>

<div id="tab-bar">
  <div id="tabs">
    {#each tabState.tabs as tab (tab.id)}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        ondragstart={(e) => { draggingTabId = tab.id; e.dataTransfer.effectAllowed = 'move'; e.dataTransfer.setData('text/plain', String(tab.id)); }}
        ondragover={(e) => { if (draggingTabId && draggingTabId !== tab.id) { e.preventDefault(); e.dataTransfer.dropEffect = 'move'; } }}
        ondrop={(e) => handleDrop(e, tab.id)}
        ondragend={() => { draggingTabId = null; }}
      >
        <Tab
          {tab}
          isActive={tab.id === tabState.activeTabId}
          onSwitch={switchToTab}
          onClose={closeTab}
          onRename={(title) => setTabTitle(tab.id, title)}
        />
      </div>
    {/each}
  </div>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="tab-bar-btn" onclick={createTab} title="New Tab">+</div>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="tab-bar-btn" onclick={onOpenSettings} title="Settings">&#x2699;</div>
</div>

<style>
  #tab-bar {
    display: flex; align-items: center;
    background: var(--bg-secondary);
    height: 36px; padding: 0 6px;
    user-select: none; -webkit-user-select: none;
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0; position: relative; z-index: 50;
  }
  #tabs { display: flex; flex: 1; overflow-x: auto; min-width: 0; scroll-behavior: smooth; }
  #tabs::-webkit-scrollbar { height: 0; }
  .tab-bar-btn {
    display: flex; align-items: center; justify-content: center;
    width: 28px; height: 28px; margin-left: 2px;
    border-radius: var(--radius-sm); cursor: pointer;
    color: var(--text-muted); font-size: 16px; flex-shrink: 0;
    transition: background var(--transition), color var(--transition);
  }
  .tab-bar-btn:hover { background: var(--bg-hover); color: #fff; }
  @media (max-width: 768px) { #tab-bar { height: 32px; } .tab-bar-btn { width: 26px; height: 26px; font-size: 14px; } }
</style>
