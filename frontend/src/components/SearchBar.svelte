<script>
  import { getPane } from '../lib/stores/panes.svelte.js';

  let { paneId, onClose } = $props();
  let query = $state('');
  let options = $state({ regex: false, caseSensitive: false, wholeWord: false });
  let resultText = $state('');
  let inputEl;

  $effect(() => { inputEl?.focus(); inputEl?.select(); });

  function doSearch(dir) {
    const resources = getPane(paneId);
    if (!resources) return;
    if (!query) { resources.searchAddon.clearDecorations(); resultText = ''; return; }
    const opts = {
      regex: options.regex, caseSensitive: options.caseSensitive, wholeWord: options.wholeWord,
      decorations: { matchBackground: '#515c6a', activeMatchBackground: '#515c6a', activeMatchBorder: '#007acc' }
    };
    const found = dir === 'prev' ? resources.searchAddon.findPrevious(query, opts) : resources.searchAddon.findNext(query, opts);
    resultText = found ? '' : 'No results';
  }

  function toggleOption(key) { options[key] = !options[key]; doSearch('next'); }

  function onKeydown(e) {
    if (e.key === 'Enter') { e.preventDefault(); doSearch(e.shiftKey ? 'prev' : 'next'); }
    if (e.key === 'Escape') { e.preventDefault(); onClose(); }
    e.stopPropagation();
  }
</script>

<div class="search-bar">
  <div class="search-options">
    <button class="search-option" class:active={options.regex} onclick={() => toggleOption('regex')} title="Regex">.*</button>
    <button class="search-option" class:active={options.caseSensitive} onclick={() => toggleOption('caseSensitive')} title="Match Case">Aa</button>
    <button class="search-option" class:active={options.wholeWord} onclick={() => toggleOption('wholeWord')} title="Whole Word">W</button>
  </div>
  <input bind:this={inputEl} bind:value={query} oninput={() => doSearch('next')} onkeydown={onKeydown} placeholder="Search..." />
  <span class="search-count">{resultText}</span>
  <button class="search-btn" onclick={() => doSearch('prev')} title="Previous">&#x25B2;</button>
  <button class="search-btn" onclick={() => doSearch('next')} title="Next">&#x25BC;</button>
  <button class="search-btn" onclick={onClose} title="Close">&times;</button>
</div>

<style>
  .search-bar {
    position: absolute; top: 0; right: 12px; z-index: 10;
    display: flex; align-items: center; gap: 4px;
    background: var(--surface-overlay); border: 1px solid var(--border); border-top: none;
    border-radius: 0 0 var(--radius) var(--radius);
    padding: 6px 8px; box-shadow: 0 4px 16px var(--shadow);
  }
  input {
    background: var(--bg-hover); color: var(--text-primary);
    border: 1px solid var(--border); border-radius: var(--radius-xs);
    padding: 5px 8px; font-size: 13px; font-family: var(--font-mono);
    width: 220px; outline: none; transition: border-color var(--transition);
  }
  input:focus { border-color: var(--accent); }
  .search-count { color: var(--text-muted); font-size: 11px; min-width: 50px; text-align: center; white-space: nowrap; }
  .search-options { display: flex; gap: 2px; }
  .search-option {
    display: flex; align-items: center; justify-content: center;
    width: 24px; height: 24px; background: transparent;
    border: 1px solid transparent; border-radius: var(--radius-xs);
    color: var(--text-muted); cursor: pointer; font-size: 11px; font-weight: bold;
    transition: all 0.1s;
  }
  .search-option:hover { background: var(--bg-hover); }
  .search-option.active { color: #fff; border-color: var(--accent); background: var(--accent-subtle); }
  .search-btn {
    display: flex; align-items: center; justify-content: center;
    width: 26px; height: 26px; background: transparent; border: none;
    border-radius: var(--radius-xs); color: var(--text-primary);
    cursor: pointer; font-size: 13px; transition: background 0.1s;
  }
  .search-btn:hover { background: var(--bg-hover); }

  @media (max-width: 768px) { input { width: 160px; font-size: 12px; } .search-bar { right: 4px; } }
  @media (max-width: 480px) { input { width: 130px; } .search-options { display: none; } }
</style>
