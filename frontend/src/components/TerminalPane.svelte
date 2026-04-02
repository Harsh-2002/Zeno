<script>
  import { onMount } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { SearchAddon } from '@xterm/addon-search';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import { WebglAddon } from '@xterm/addon-webgl';
  import { Unicode11Addon } from '@xterm/addon-unicode11';
  import { registerPane, getPane, setFocusedPane, paneState } from '../lib/stores/panes.svelte.js';
  import { themeState } from '../lib/stores/theme.svelte.js';
  import { setTabTitle, markUnread } from '../lib/stores/tabs.svelte.js';
  import { THEMES } from '../lib/constants/themes.js';
  import { createWebSocket, sendData, sendBinary, sendResize, sendSessionConnect, MSG_DATA, MSG_SESSION } from '../lib/utils/ws.js';
  import SearchBar from './SearchBar.svelte';
  import DropOverlay from './DropOverlay.svelte';

  let { paneId, tabId, canClose = false, onClosePane } = $props();
  let terminalEl;
  let searchVisible = $state(false);
  let dropVisible = $state(false);
  let dragCounter = 0;

  const isFocused = $derived(paneState.focusedPaneId === paneId);

  export function toggleSearch() { searchVisible = !searchVisible; }

  onMount(() => {
    let resources = getPane(paneId);
    if (resources) {
      // Re-attach existing terminal to new DOM element.
      // xterm.js doesn't support re-opening, so we must dispose the old
      // rendering layer and create a fresh terminal with the same WebSocket.
      const oldTerm = resources.term;
      const ws = resources.ws;

      // Create fresh terminal instance
      const term = new Terminal({
        cursorBlink: themeState.cursorBlink,
        cursorStyle: themeState.cursorStyle,
        fontSize: themeState.fontSize,
        fontFamily: `"${themeState.fontFamily}", Menlo, Monaco, "Cascadia Code", "Courier New", monospace`,
        fontWeight: '400', fontWeightBold: '700',
        lineHeight: themeState.lineHeight,
        letterSpacing: 0,
        scrollback: themeState.scrollback,
        allowProposedApi: true,
        theme: THEMES[themeState.themeId]
      });

      const fitAddon = new FitAddon();
      const searchAddon = new SearchAddon();
      term.loadAddon(fitAddon);
      term.loadAddon(searchAddon);
      term.loadAddon(new WebLinksAddon({ handler: (e, uri) => window.open(uri, '_blank', 'noopener') }));
      const u11 = new Unicode11Addon();
      term.loadAddon(u11);
      term.unicode.activeVersion = '11';
      term.open(terminalEl);
      try { const wgl = new WebglAddon(); wgl.onContextLost(() => wgl.dispose()); term.loadAddon(wgl); } catch (e) {}

      // Wire the new terminal to the existing WebSocket
      term.onData((data) => sendData(ws, data));
      term.onBinary((data) => sendBinary(ws, data));
      term.onTitleChange((title) => { if (title) setTabTitle(tabId, title); });

      // Redirect incoming WS data to the new terminal
      ws.onmessage = (event) => {
        const d = new Uint8Array(event.data);
        if (d.length > 0 && d[0] === MSG_DATA) {
          term.write(d.slice(1));
          markUnread(tabId);
        }
      };

      // Dispose old terminal rendering (not the WebSocket)
      oldTerm.dispose();

      // Update resources
      resources.term = term;
      resources.fitAddon = fitAddon;
      resources.searchAddon = searchAddon;
      resources.el = terminalEl;
      resources.toggleSearch = () => { searchVisible = !searchVisible; };

      requestAnimationFrame(() => {
        fitAddon.fit();
        sendResize(ws, term.cols, term.rows);
      });
      return;
    }

    const term = new Terminal({
      cursorBlink: themeState.cursorBlink,
      cursorStyle: themeState.cursorStyle,
      fontSize: themeState.fontSize,
      fontFamily: `"${themeState.fontFamily}", Menlo, Monaco, "Cascadia Code", "Courier New", monospace`,
      fontWeight: '400', fontWeightBold: '700',
      lineHeight: themeState.lineHeight,
      letterSpacing: 0,
      scrollback: themeState.scrollback,
      allowProposedApi: true,
      theme: THEMES[themeState.themeId]
    });

    const fitAddon = new FitAddon();
    const searchAddon = new SearchAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(searchAddon);
    term.loadAddon(new WebLinksAddon({ handler: (e, uri) => window.open(uri, '_blank', 'noopener') }));
    const u11 = new Unicode11Addon();
    term.loadAddon(u11);
    term.unicode.activeVersion = '11';
    term.open(terminalEl);

    try {
      const wgl = new WebglAddon();
      wgl.onContextLost(() => wgl.dispose());
      term.loadAddon(wgl);
    } catch (e) {}

    let sessionId = '';
    let reconnectAttempts = 0;
    let currentWs = null;

    // Register input handlers ONCE — they use currentWs which updates on reconnect
    term.onData((data) => sendData(currentWs, data));
    term.onBinary((data) => sendBinary(currentWs, data));
    term.onTitleChange((title) => { if (title) setTabTitle(tabId, title); });

    function connectWs() {
      const ws = createWebSocket();
      currentWs = ws;
      let isReconnect = reconnectAttempts > 0;

      ws.onmessage = (event) => {
        const data = new Uint8Array(event.data);
        if (data.length === 0) return;
        if (data[0] === MSG_SESSION) {
          try {
            const msg = JSON.parse(new TextDecoder().decode(data.slice(1)));
            if (msg.sessionID) {
              sessionId = msg.sessionID;
              const res = getPane(paneId);
              if (res) res.sessionId = sessionId;
            }
          } catch (e) {}
          return;
        }
        if (data[0] === MSG_DATA) {
          term.write(data.slice(1));
          markUnread(tabId);
        }
      };

      ws.onopen = () => {
        reconnectAttempts = 0;
        // On reconnect, reset terminal before ring buffer replay
        if (isReconnect) {
          term.reset();
        }
        sendSessionConnect(ws, sessionId);
        requestAnimationFrame(() => {
          fitAddon.fit();
          sendResize(ws, term.cols, term.rows);
          term.focus();
        });
      };

      ws.onclose = () => {
        const r = getPane(paneId);
        if (r && !r.closed && reconnectAttempts < 5) {
          reconnectAttempts++;
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts - 1), 8000);
          term.write(`\r\n\x1b[33m[Reconnecting in ${delay/1000}s...]\x1b[0m\r\n`);
          setTimeout(() => {
            if (r && !r.closed) connectWs();
          }, delay);
        } else if (r && !r.closed) {
          term.write('\r\n\x1b[1;31m[Session ended]\x1b[0m\r\n');
        }
      };
      ws.onerror = () => {};

      return ws;
    }

    const ws = connectWs();

    registerPane(paneId, { term, fitAddon, searchAddon, ws: currentWs, el: terminalEl, closed: false, sessionId: '', toggleSearch: () => { searchVisible = !searchVisible; }, getWs: () => currentWs });

    // Custom scrollbar tracking
    setupScrollbar(term);
  });

  let scrollThumbTop = $state(0);
  let scrollThumbHeight = $state(0);
  let scrollVisible = $state(false);
  let scrollDragging = $state(false);
  let scrollHideTimer = null;

  let isScrollable = $state(false);
  let paneHovered = $state(false);

  function setupScrollbar(term) {
    function update() {
      const vp = term.element?.querySelector('.xterm-viewport');
      if (!vp) return;
      const scrollHeight = vp.scrollHeight;
      const clientHeight = vp.clientHeight;
      isScrollable = scrollHeight > clientHeight;
      if (!isScrollable) { scrollVisible = false; return; }
      const ratio = clientHeight / scrollHeight;
      scrollThumbHeight = Math.max(20, ratio * clientHeight);
      scrollThumbTop = (vp.scrollTop / (scrollHeight - clientHeight)) * (clientHeight - scrollThumbHeight);
      scrollVisible = true;
      clearTimeout(scrollHideTimer);
      scrollHideTimer = setTimeout(() => { if (!scrollDragging) scrollVisible = false; }, 1500);
    }
    term.onScroll(update);
    term.onWriteParsed(update);
    // Watch viewport resize to recalculate
    requestAnimationFrame(() => {
      const vp = term.element?.querySelector('.xterm-viewport');
      if (vp) new ResizeObserver(update).observe(vp);
    });
  }

  function handleScrollThumbDown(e) {
    e.preventDefault();
    scrollDragging = true;
    const r = getPane(paneId);
    if (!r) return;
    const vp = r.term.element?.querySelector('.xterm-viewport');
    if (!vp) return;
    const startY = e.clientY;
    const startScrollTop = vp.scrollTop;
    const trackHeight = vp.clientHeight - scrollThumbHeight;
    const scrollRange = vp.scrollHeight - vp.clientHeight;

    function onMove(e2) {
      const delta = e2.clientY - startY;
      vp.scrollTop = startScrollTop + (delta / trackHeight) * scrollRange;
    }
    function onUp() {
      scrollDragging = false;
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    }
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }

  $effect(() => { const r = getPane(paneId); if (r) r.term.options.theme = THEMES[themeState.themeId]; });
  $effect(() => {
    const r = getPane(paneId);
    if (r) {
      r.term.options.fontSize = themeState.fontSize;
      // Refit after font size change so terminal recalculates cols/rows
      requestAnimationFrame(() => {
        r.fitAddon.fit();
        sendResize(r.getWs ? r.getWs() : r.ws, r.term.cols, r.term.rows);
      });
    }
  });
  $effect(() => { const r = getPane(paneId); if (r) r.term.options.cursorStyle = themeState.cursorStyle; });
  $effect(() => { const r = getPane(paneId); if (r) r.term.options.cursorBlink = themeState.cursorBlink; });
  $effect(() => {
    const r = getPane(paneId);
    if (r) {
      r.term.options.fontFamily = `"${themeState.fontFamily}", Menlo, Monaco, "Cascadia Code", "Courier New", monospace`;
      requestAnimationFrame(() => { r.fitAddon.fit(); sendResize(r.getWs ? r.getWs() : r.ws, r.term.cols, r.term.rows); });
    }
  });
  $effect(() => {
    const r = getPane(paneId);
    if (r) {
      r.term.options.lineHeight = themeState.lineHeight;
      requestAnimationFrame(() => {
        r.fitAddon.fit();
        sendResize(r.getWs ? r.getWs() : r.ws, r.term.cols, r.term.rows);
      });
    }
  });

  function handleMouseDown() { setFocusedPane(paneId); }
  function handleDragEnter(e) { if (e.dataTransfer?.types?.includes('Files')) { e.preventDefault(); dragCounter++; dropVisible = true; } }
  function handleDragLeave() { dragCounter--; if (dragCounter <= 0) { dragCounter = 0; dropVisible = false; } }
  function handleDragOver(e) { if (e.dataTransfer?.types?.includes('Files')) e.preventDefault(); }
  function handleDrop(e) {
    e.preventDefault(); dragCounter = 0; dropVisible = false;
    const files = e.dataTransfer?.files;
    if (!files || files.length === 0) return;
    const r = getPane(paneId);
    const sid = r?.sessionId || '';
    for (const file of files) {
      const form = new FormData();
      form.append('file', file);
      fetch(`/api/upload?session=${sid}`, { method: 'POST', body: form })
        .then(res => res.ok ? res.json() : Promise.reject(res.statusText))
        .then(data => r?.term?.write(`\r\n\x1b[32m[Uploaded: ${data.name} (${(data.size/1024).toFixed(1)}KB)]\x1b[0m\r\n`))
        .catch(err => r?.term?.write(`\r\n\x1b[31m[Upload failed: ${err}]\x1b[0m\r\n`));
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="pane" data-pane-id={paneId}
  onmousedown={handleMouseDown}
  onmouseenter={() => { paneHovered = true; }}
  onmouseleave={() => { paneHovered = false; }}
  ondragenter={handleDragEnter} ondragleave={handleDragLeave}
  ondragover={handleDragOver} ondrop={handleDrop}>

  <div class="pane-frame" class:focused={isFocused}>
    {#if canClose}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div class="pane-close" onclick={(e) => { e.stopPropagation(); onClosePane?.(paneId); }}>&times;</div>
    {/if}
    <div class="pane-terminal" bind:this={terminalEl}></div>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="custom-scrollbar" class:active={(scrollVisible || paneHovered) && isScrollable || scrollDragging}>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="custom-scrollbar-thumb" class:dragging={scrollDragging}
        style="top:{scrollThumbTop}px; height:{scrollThumbHeight}px"
        onmousedown={handleScrollThumbDown}></div>
    </div>
  </div>

  {#if searchVisible}
    <SearchBar {paneId} onClose={() => { searchVisible = false; getPane(paneId)?.term.focus(); }} />
  {/if}
  <DropOverlay visible={dropVisible} />
</div>

<style>
  .pane {
    position: relative;
    overflow: hidden;
    flex: 1;
    display: flex;
    padding: 3px;
  }

  .pane-frame {
    flex: 1;
    border: 1.5px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
    transition: border-color 0.2s ease;
    background: var(--bg-primary);
  }

  .pane-frame.focused {
    border-color: var(--border-focus);
  }

  .pane-close {
    position: absolute; top: 8px; right: 8px; z-index: 15;
    width: 22px; height: 22px;
    display: flex; align-items: center; justify-content: center;
    background: var(--bg-secondary); border: 1px solid var(--border);
    border-radius: 6px; color: var(--text-muted);
    font-size: 14px; cursor: pointer;
    opacity: 0; transition: opacity 0.15s, background 0.15s;
  }
  .pane:hover .pane-close { opacity: 0.7; }
  .pane:hover .pane-close:hover { opacity: 1; background: var(--danger); color: #fff; border-color: var(--danger); }

  .pane-terminal {
    height: 100%;
    padding: 4px;
  }

  .pane-terminal :global(.xterm) {
    height: 100%;
  }
</style>
