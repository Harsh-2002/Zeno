<script>
  let { visible = false, onClose, sessionId = '' } = $props();

  let currentPath = $state('.');
  let fullPath = $state('');
  let entries = $state([]);
  let loading = $state(false);
  let dropActive = $state(false);
  let dragCounter = 0;
  let menuEntry = $state(null);
  let menuPos = $state({ x: 0, y: 0 });
  let renamingEntry = $state(null);
  let renameValue = $state('');
  let renameInputEl = $state(null);

  $effect(() => {
    if (visible && sessionId) loadDir('.');
  });

  $effect(() => {
    if (renamingEntry && renameInputEl) { renameInputEl.focus(); renameInputEl.select(); }
  });

  async function loadDir(path) {
    loading = true;
    currentPath = path;
    try {
      const res = await fetch(`/api/files?session=${sessionId}&path=${encodeURIComponent(path)}`);
      if (!res.ok) throw new Error(res.statusText);
      const data = await res.json();
      fullPath = data.path;
      entries = data.entries || [];
    } catch (e) {
      entries = [];
      fullPath = 'Error loading directory';
    }
    loading = false;
  }

  function navigateTo(name) {
    const newPath = currentPath === '.' ? name : `${currentPath}/${name}`;
    loadDir(newPath);
  }

  function navigateUp() {
    if (currentPath === '.') return;
    const parts = currentPath.split('/');
    parts.pop();
    loadDir(parts.length === 0 ? '.' : parts.join('/'));
  }

  function downloadFile(name) {
    const filePath = currentPath === '.' ? name : `${currentPath}/${name}`;
    const url = `/api/download?session=${sessionId}&path=${encodeURIComponent(filePath)}`;
    const a = document.createElement('a');
    a.href = url; a.download = name;
    a.click();
  }

  function formatSize(bytes) {
    if (bytes === 0) return '';
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
  }

  function shortenPath(p) {
    if (!p) return '';
    const home = p.match(/^\/Users\/[^/]+/)?.[0] || p.match(/^\/home\/[^/]+/)?.[0];
    if (home) return '~' + p.slice(home.length);
    return p;
  }

  function openMenu(e, entry) {
    e.stopPropagation();
    const rect = e.currentTarget.getBoundingClientRect();
    menuEntry = entry;
    menuPos = { x: rect.right - 140, y: rect.bottom + 2 };
  }

  function closeMenu() { menuEntry = null; }

  function startRename(entry) {
    closeMenu();
    renamingEntry = entry;
    renameValue = entry.name;
  }

  async function doRename() {
    if (!renamingEntry || !renameValue.trim() || renameValue === renamingEntry.name) {
      renamingEntry = null;
      return;
    }
    const filePath = currentPath === '.' ? renamingEntry.name : `${currentPath}/${renamingEntry.name}`;
    try {
      await fetch(`/api/rename?session=${sessionId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: filePath, newName: renameValue.trim() })
      });
    } catch (e) {}
    renamingEntry = null;
    loadDir(currentPath);
  }

  async function doDelete(entry) {
    closeMenu();
    if (!confirm(`Delete "${entry.name}"?`)) return;
    const filePath = currentPath === '.' ? entry.name : `${currentPath}/${entry.name}`;
    try {
      await fetch(`/api/delete?session=${sessionId}&path=${encodeURIComponent(filePath)}`, { method: 'DELETE' });
    } catch (e) {}
    loadDir(currentPath);
  }

  function handleDragEnter(e) { e.preventDefault(); dragCounter++; dropActive = true; }
  function handleDragLeave() { dragCounter--; if (dragCounter <= 0) { dragCounter = 0; dropActive = false; } }
  function handleDragOver(e) { e.preventDefault(); }
  function handleDrop(e) {
    e.preventDefault(); dragCounter = 0; dropActive = false;
    const files = e.dataTransfer?.files;
    if (!files || files.length === 0) return;
    for (const file of files) {
      const form = new FormData();
      form.append('file', file);
      const uploadPath = currentPath === '.' ? '' : currentPath;
      fetch(`/api/upload?session=${sessionId}&path=${encodeURIComponent(uploadPath)}`, { method: 'POST', body: form })
        .then(res => res.ok ? res.json() : Promise.reject(res.statusText))
        .then(() => loadDir(currentPath))
        .catch(() => {});
    }
  }
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="overlay" onclick={onClose}></div>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="file-panel" oncontextmenu={(e) => e.stopPropagation()}
    ondragenter={handleDragEnter} ondragleave={handleDragLeave}
    ondragover={handleDragOver} ondrop={handleDrop}>
    <div class="header">
      <span class="title">Files</span>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <span class="close" onclick={onClose}>&times;</span>
    </div>

    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="breadcrumb" onclick={navigateUp}>
      {#if currentPath !== '.'}
        <span class="back"><svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10 3L5 8l5 5"/></svg></span>
      {/if}
      <svg class="breadcrumb-icon" width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3.5h4.5l1.5 1.5H14v8H2z"/></svg>
      <span class="path">{shortenPath(fullPath)}</span>
    </div>

    <div class="file-list">
      {#if loading}
        <div class="empty">Loading...</div>
      {:else if entries.length === 0}
        <div class="empty">Empty directory</div>
      {:else}
        {#each entries as entry}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div class="file-row" onclick={() => { if (renamingEntry !== entry) { entry.isDir ? navigateTo(entry.name) : downloadFile(entry.name); } }}>
            <span class="file-icon">
              {#if entry.isDir}
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3.5h4.5l1.5 1.5H14v8H2z"/></svg>
              {:else}
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"><path d="M4 1.5h5l4 4v9H4z"/><path d="M9 1.5v4h4"/></svg>
              {/if}
            </span>
            {#if renamingEntry === entry}
              <input class="rename-input" bind:this={renameInputEl} bind:value={renameValue}
                onclick={(e) => e.stopPropagation()}
                onblur={doRename}
                onkeydown={(e) => { if (e.key === 'Enter') e.target.blur(); if (e.key === 'Escape') { renamingEntry = null; } e.stopPropagation(); }} />
            {:else}
              <span class="file-name">{entry.name}</span>
              {#if !entry.isDir}
                <span class="file-size">{formatSize(entry.size)}</span>
              {/if}
            {/if}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <span class="file-menu-btn" onclick={(e) => openMenu(e, entry)}>
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><circle cx="8" cy="3" r="1.2"/><circle cx="8" cy="8" r="1.2"/><circle cx="8" cy="13" r="1.2"/></svg>
            </span>
          </div>
        {/each}

        {#if menuEntry}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div class="file-menu-overlay" onclick={closeMenu}></div>
          <div class="file-menu" style="left:{menuPos.x}px; top:{menuPos.y}px">
            {#if !menuEntry.isDir}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div class="file-menu-item" onclick={() => { downloadFile(menuEntry.name); closeMenu(); }}>Download</div>
            {/if}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="file-menu-item" onclick={() => startRename(menuEntry)}>Rename</div>
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="file-menu-item danger" onclick={() => doDelete(menuEntry)}>Delete</div>
          </div>
        {/if}
      {/if}
    </div>

    <div class="upload-zone" class:active={dropActive}>
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"><path d="M8 10V3"/><path d="M5 5l3-3 3 3"/><path d="M2 11v2h12v-2"/></svg>
      <span>{dropActive ? 'Drop to upload' : 'Drag files here'}</span>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed; inset: 0; z-index: 900;
    background: rgba(0,0,0,0.3);
  }
  .file-panel {
    position: fixed; top: 0; left: 0; bottom: 0;
    width: 280px; max-width: 80vw;
    z-index: 950;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex; flex-direction: column;
    box-shadow: 2px 0 8px rgba(0,0,0,0.15);
    animation: slide-left 0.15s ease;
  }
  @keyframes slide-left {
    from { transform: translateX(-100%); }
    to { transform: translateX(0); }
  }
  .header {
    display: flex; align-items: center; justify-content: space-between;
    padding: 14px 16px; border-bottom: 1px solid var(--border);
  }
  .title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
  .close {
    font-size: 18px; color: var(--text-muted); cursor: pointer;
    width: 26px; height: 26px; display: flex; align-items: center;
    justify-content: center; border-radius: var(--radius-xs);
    transition: background 0.1s;
  }
  .close:hover { background: var(--bg-hover); color: var(--text-primary); }

  .breadcrumb {
    padding: 8px 16px; display: flex; align-items: center; gap: 6px;
    color: var(--text-muted); font-size: 11px; font-family: var(--font-mono);
    cursor: pointer; border-bottom: 1px solid var(--border);
    transition: background 0.1s;
    overflow: hidden; white-space: nowrap; text-overflow: ellipsis;
    min-height: 32px;
  }
  .breadcrumb:hover { background: var(--bg-hover); }
  .breadcrumb-icon { flex-shrink: 0; }
  .back { display: flex; flex-shrink: 0; }
  .path { overflow: hidden; text-overflow: ellipsis; }

  .file-list {
    flex: 1; overflow-y: auto; padding: 4px 0;
  }
  .file-list::-webkit-scrollbar { width: 4px; }
  .file-list::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 2px; }

  .file-row {
    display: flex; align-items: center; gap: 8px;
    padding: 6px 16px; cursor: pointer;
    transition: background 0.08s;
    font-size: 13px; color: var(--text-primary);
  }
  .file-row:hover { background: var(--bg-hover); }
  .file-icon { font-size: 14px; flex-shrink: 0; }
  .file-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .file-size { color: var(--text-dim); font-size: 11px; font-family: var(--font-mono); flex-shrink: 0; }

  .file-menu-btn {
    display: flex; align-items: center; justify-content: center;
    width: 22px; height: 22px; border-radius: var(--radius-xs);
    color: var(--text-dim); cursor: pointer; flex-shrink: 0;
    opacity: 0; transition: opacity 0.1s, background 0.1s;
  }
  .file-row:hover .file-menu-btn { opacity: 0.6; }
  .file-menu-btn:hover { opacity: 1 !important; background: var(--bg-hover); }

  .rename-input {
    flex: 1; min-width: 0;
    background: var(--bg-hover); color: var(--text-primary);
    border: 1px solid var(--border-focus); border-radius: 3px;
    padding: 1px 6px; font-size: 13px; font-family: inherit; outline: none;
  }

  .file-menu-overlay { position: fixed; inset: 0; z-index: 960; }
  .file-menu {
    position: fixed; z-index: 970;
    background: var(--surface-overlay); border: 1px solid var(--border);
    border-radius: var(--radius-sm); padding: 3px 0; min-width: 120px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  }
  .file-menu-item {
    padding: 5px 12px; font-size: 12px; color: var(--text-primary);
    cursor: pointer; transition: background 0.08s;
  }
  .file-menu-item:hover { background: var(--bg-hover); }
  .file-menu-item.danger { color: var(--danger); }
  .file-menu-item.danger:hover { background: var(--danger); color: #fff; }

  .empty {
    padding: 24px 16px; text-align: center;
    color: var(--text-dim); font-size: 13px;
  }

  .upload-zone {
    padding: 12px 16px;
    display: flex; align-items: center; justify-content: center; gap: 6px;
    color: var(--text-dim); font-size: 11px;
    border-top: 1px solid var(--border);
    transition: all 0.15s;
  }
  .upload-zone.active {
    background: var(--accent-subtle);
    color: var(--accent);
    border-top-color: var(--accent);
  }

  @media (max-width: 480px) { .file-panel { width: 100%; } }
</style>
