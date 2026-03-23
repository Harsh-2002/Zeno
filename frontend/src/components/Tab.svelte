<script>
  import { countPanes, getSplitBadgeType } from '../lib/utils/paneTree.js';

  let { tab, isActive = false, onSwitch, onClose, onRename } = $props();
  let isRenaming = $state(false);
  let renameValue = $state('');
  let inputEl = $state(null);

  function shortenTitle(title) {
    if (!title) return 'Terminal';
    const ci = title.lastIndexOf(':');
    if (ci !== -1) { const p = title.substring(ci + 1).trim(); if (p) { const parts = p.replace(/\/$/, '').split('/'); return parts.length > 2 ? parts.slice(-2).join('/') : p; } }
    if (title.indexOf('/') !== -1) { const parts = title.replace(/\/$/, '').split('/'); if (parts.length > 2) return parts.slice(-2).join('/'); }
    return title.length > 30 ? title.substring(0, 27) + '...' : title;
  }

  function startRename() {
    isRenaming = true;
    renameValue = tab.title;
    requestAnimationFrame(() => { inputEl?.focus(); inputEl?.select(); });
  }

  function finishRename() {
    isRenaming = false;
    const v = renameValue.trim();
    if (v) onRename?.(v);
  }

  function onRenameKeydown(e) {
    if (e.key === 'Enter') { e.target.blur(); }
    if (e.key === 'Escape') { renameValue = tab.title; e.target.blur(); }
    e.stopPropagation();
  }

  const paneCount = $derived(countPanes(tab.rootNode));
  const splitBadge = $derived(getSplitBadgeType(tab.rootNode));
  const displayTitle = $derived(shortenTitle(tab.title));
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="tab" class:active={isActive} onclick={() => onSwitch?.(tab.id)} draggable="true">
  {#if isRenaming}
    <input class="rename-input" bind:this={inputEl} bind:value={renameValue}
      onblur={finishRename} onkeydown={onRenameKeydown} />
  {:else}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <span class="tab-title" ondblclick={startRename}>{displayTitle}</span>
  {/if}

  {#if splitBadge}
    <span class="tab-split-badge">
      <span class="tab-split-icon {splitBadge}"></span>
    </span>
  {/if}

  {#if tab.hasUnread && !isActive}
    <span class="tab-badge"></span>
  {/if}

  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <span class="tab-close" onclick={(e) => { e.stopPropagation(); onClose?.(tab.id); }}>&times;</span>
</div>

<style>
  .tab {
    display: flex; align-items: center;
    padding: 0 10px; height: 30px; margin: 3px 1px;
    background: transparent; color: var(--text-muted);
    border-radius: var(--radius-sm); cursor: pointer;
    font-size: 12px; white-space: nowrap;
    min-width: 50px; max-width: 180px;
    transition: background var(--transition), color var(--transition);
    flex-shrink: 0;
  }
  .tab:hover { background: var(--bg-hover); color: var(--text-primary); }
  .tab.active { background: var(--bg-primary); color: #fff; }
  .tab-title { flex: 1; overflow: hidden; text-overflow: ellipsis; }
  .tab-badge { width: 7px; height: 7px; border-radius: 50%; background: var(--accent); margin-left: 5px; flex-shrink: 0; }
  .tab-split-badge { font-size: 9px; color: var(--text-dim); margin-left: 4px; flex-shrink: 0; display: flex; align-items: center; opacity: 0.7; }
  .tab-split-icon { display: inline-flex; width: 12px; height: 10px; border: 1px solid var(--text-dim); border-radius: 2px; position: relative; overflow: hidden; }
  .tab-split-icon::after { content: ''; position: absolute; background: var(--text-dim); }
  .tab-split-icon.v::after { top: 0; bottom: 0; left: 50%; width: 1px; transform: translateX(-0.5px); }
  .tab-split-icon.h::after { left: 0; right: 0; top: 50%; height: 1px; transform: translateY(-0.5px); }
  .tab-split-icon.multi::after { top: 0; bottom: 0; left: 50%; width: 1px; transform: translateX(-0.5px); }
  .tab-split-icon.multi::before { content: ''; position: absolute; left: 50%; right: 0; top: 50%; height: 1px; background: var(--text-dim); transform: translateY(-0.5px); }
  .tab-close {
    display: flex; align-items: center; justify-content: center;
    width: 18px; height: 18px; margin-left: 4px;
    border-radius: var(--radius-xs); font-size: 13px; color: var(--text-dim);
    opacity: 0; transition: opacity var(--transition), background var(--transition);
    flex-shrink: 0;
  }
  .tab:hover .tab-close, .tab.active .tab-close { opacity: 1; }
  .tab-close:hover { background: rgba(255,255,255,0.1); color: #fff; }
  .rename-input {
    background: var(--bg-primary); color: #fff; border: 1px solid var(--border);
    border-radius: 3px; font-size: 12px; padding: 1px 4px;
    width: 100%; outline: none; font-family: inherit;
  }

  @media (max-width: 768px) { .tab { font-size: 11px; height: 26px; max-width: 120px; } .tab-close { opacity: 1; } }
  @media (max-width: 480px) { .tab { max-width: 90px; padding: 0 6px; } }
</style>
