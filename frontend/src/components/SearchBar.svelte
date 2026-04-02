<script>
  import { getPane } from '../lib/stores/panes.svelte.js';
  import { UI_COLORS } from '../lib/constants/themes.js';
  import { themeState } from '../lib/stores/theme.svelte.js';

  let { paneId, onClose } = $props();
  let query = $state('');
  let options = $state({ regex: false, caseSensitive: false, wholeWord: false });
  let resultText = $state('');
  let inputEl = $state(null);

  $effect(() => { inputEl?.focus(); inputEl?.select(); });

  function getDecorations() {
    const u = UI_COLORS[themeState.themeId];
    return {
      matchBackground: u.matchBg,
      activeMatchBackground: u.activeMatchBg,
      activeMatchBorder: u.activeMatchBorder,
    };
  }

  function doSearch(dir) {
    const resources = getPane(paneId);
    if (!resources) return;
    if (!query) { resources.searchAddon.clearDecorations(); resultText = ''; return; }
    const opts = {
      regex: options.regex, caseSensitive: options.caseSensitive, wholeWord: options.wholeWord,
      decorations: getDecorations()
    };
    const found = dir === 'prev' ? resources.searchAddon.findPrevious(query, opts) : resources.searchAddon.findNext(query, opts);
    resultText = found ? '' : 'No results';
  }

  function toggleOption(key) { options[key] = !options[key]; doSearch('next'); }

  function onKeydown(e) {
    if (e.key === 'Enter') { e.preventDefault(); doSearch(e.shiftKey ? 'prev' : 'next'); }
    if (e.key === 'Escape') { e.preventDefault(); handleClose(); }
    e.stopPropagation();
  }

  function handleClose() {
    const resources = getPane(paneId);
    if (resources) { resources.searchAddon.clearDecorations(); resources.term.focus(); }
    onClose();
  }
</script>

<div class="search-container">
  <div class="search-row">
    <input bind:this={inputEl} bind:value={query} oninput={() => doSearch('next')} onkeydown={onKeydown} placeholder="Find..." />
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div class="nav-btn" onclick={() => doSearch('prev')} title="Previous">&lsaquo;</div>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div class="nav-btn" onclick={() => doSearch('next')} title="Next">&rsaquo;</div>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div class="close-btn" onclick={handleClose} title="Close">&times;</div>
  </div>
  <div class="search-meta">
    <div class="options">
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <span class="opt" class:active={options.regex} onclick={() => toggleOption('regex')} title="Regex">.*</span>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <span class="opt" class:active={options.caseSensitive} onclick={() => toggleOption('caseSensitive')} title="Match case">Aa</span>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <span class="opt" class:active={options.wholeWord} onclick={() => toggleOption('wholeWord')} title="Whole word">W</span>
    </div>
    {#if resultText}
      <span class="result-text">{resultText}</span>
    {/if}
  </div>
</div>

<style>
  .search-container {
    position: absolute;
    top: 0; right: 12px;
    z-index: 10;
    background: var(--surface-overlay);
    border: 1px solid var(--border);
    border-top: none;
    border-radius: 0 0 var(--radius) var(--radius);
    padding: 8px 10px 6px;
    box-shadow: 0 6px 20px var(--shadow);
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 200px;
    animation: search-in 0.15s ease;
  }

  @keyframes search-in {
    from { transform: translateY(-100%); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }

  .search-row {
    display: flex;
    align-items: center;
    gap: 2px;
  }

  input {
    flex: 1;
    background: var(--bg-hover);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 13px;
    font-family: var(--font-mono);
    outline: none;
    min-width: 0;
    transition: border-color 0.15s;
  }
  input:focus { border-color: var(--border-focus); }
  input::placeholder { color: var(--text-dim); }

  .nav-btn, .close-btn {
    width: 26px; height: 26px;
    display: flex; align-items: center; justify-content: center;
    border-radius: 6px;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 16px;
    flex-shrink: 0;
    transition: background 0.1s, color 0.1s;
  }
  .nav-btn:hover, .close-btn:hover { background: var(--bg-hover); color: var(--text-primary); }

  .search-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-height: 20px;
  }

  .options { display: flex; gap: 2px; }

  .opt {
    display: flex; align-items: center; justify-content: center;
    width: 24px; height: 22px;
    border-radius: 4px;
    border: 1px solid transparent;
    color: var(--text-dim);
    font-size: 11px; font-weight: 600;
    cursor: pointer;
    transition: all 0.1s;
    font-family: var(--font-mono);
  }
  .opt:hover { background: var(--bg-hover); color: var(--text-muted); }
  .opt.active {
    color: var(--text-primary);
    border-color: var(--border-focus);
    background: var(--bg-hover);
  }

  .result-text {
    color: var(--text-dim);
    font-size: 11px;
    white-space: nowrap;
  }

  @media (max-width: 768px) {
    .search-container { right: 4px; min-width: 160px; }
    input { font-size: 12px; padding: 5px 8px; }
  }
  @media (max-width: 480px) { .options { display: none; } }
</style>
