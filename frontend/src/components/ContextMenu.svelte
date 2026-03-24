<script>
  import { shortcutLabel } from '../lib/utils/shortcuts.js';

  let { visible = false, x = 0, y = 0, onAction, onClose } = $props();
  const items = [
    { label: 'Copy', shortcut: shortcutLabel('C'), action: 'copy' },
    { label: 'Paste', shortcut: shortcutLabel('V'), action: 'paste' },
    { label: 'Select All', action: 'selectAll' },
    { sep: true },
    { label: 'Clear Terminal', action: 'clear' },
    { label: 'Save Output', action: 'saveOutput' },
    { label: 'Search', shortcut: shortcutLabel('F'), action: 'search' },
    { sep: true },
    { label: 'Split Vertical', shortcut: shortcutLabel('D'), action: 'splitV' },
    { label: 'Split Horizontal', shortcut: shortcutLabel('\u21E7D'), action: 'splitH' },
    { label: 'Close Pane', shortcut: shortcutLabel('W'), action: 'closePane' },
    { sep: true },
    { label: 'Settings...', shortcut: shortcutLabel(','), action: 'settings' },
  ];

  function clampX() { return Math.min(x, window.innerWidth - 240); }
  function clampY() { return Math.min(y, window.innerHeight - 360); }
  function handleAction(action, data) { onAction(action, data); onClose(); }
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="context-menu" style="left:{clampX()}px; top:{clampY()}px" onclick={(e) => e.stopPropagation()}>
    {#each items as item}
      {#if item.sep}
        <div class="sep"></div>
      {:else}
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div class="item" onclick={() => handleAction(item.action)}>
          {item.label}
          {#if item.shortcut}<span class="shortcut">{item.shortcut}</span>{/if}
        </div>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .context-menu {
    position: fixed; z-index: 1000;
    background: var(--surface-overlay); border: 1px solid var(--border);
    border-radius: var(--radius); padding: 4px 0; min-width: 200px;
    box-shadow: 0 8px 24px var(--shadow); font-family: var(--font-ui); font-size: 13px;
    backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  }
  .item {
    padding: 6px 12px; color: var(--text-primary); cursor: pointer;
    display: flex; justify-content: space-between; align-items: center;
    transition: background 0.08s; border-radius: var(--radius-xs); margin: 0 4px;
  }
  .item:hover { background: var(--accent); color: #fff; }
  .item:hover .shortcut { color: rgba(255,255,255,0.6); }
  .sep { height: 1px; background: var(--border); margin: 4px 8px; }
  .shortcut { color: var(--text-dim); font-size: 11px; margin-left: 20px; }
  @media (max-width: 768px) { .context-menu { min-width: 180px; } .item { font-size: 12px; padding: 8px 12px; } }
  @media (hover: none) and (pointer: coarse) { .item { min-height: 40px; padding: 10px 14px; } }
</style>
