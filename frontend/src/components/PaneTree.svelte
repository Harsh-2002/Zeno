<script>
  import PaneTree from './PaneTree.svelte';
  import TerminalPane from './TerminalPane.svelte';
  import SplitDivider from './SplitDivider.svelte';
  import { forEachPane } from '../lib/utils/paneTree.js';
  import { getPane } from '../lib/stores/panes.svelte.js';
  import { sendResize } from '../lib/utils/ws.js';

  let { node, tabId, hasSplits = false, onClosePane } = $props();

  function handleResize(newRatio) {
    if (node.type === 'split') node.ratio = newRatio;
  }

  function handleResizeEnd() {
    forEachPane(node, (pane) => {
      const r = getPane(pane.id);
      if (r) {
        requestAnimationFrame(() => {
          r.fitAddon.fit();
          sendResize(r.getWs ? r.getWs() : r.ws, r.term.cols, r.term.rows);
        });
      }
    });
  }
</script>

{#if node.type === 'pane'}
  <TerminalPane paneId={node.id} {tabId} canClose={hasSplits} {onClosePane} initialSessionId={node.sessionId || ''} />
{:else if node.type === 'split'}
  <div class="pane-split {node.direction}">
    <div class="pane-child" style="flex-basis: {node.ratio * 100}%">
      <PaneTree node={node.first} {tabId} hasSplits={true} {onClosePane} />
    </div>
    <SplitDivider direction={node.direction} onResize={handleResize} onResizeEnd={handleResizeEnd} />
    <div class="pane-child" style="flex-basis: {(1 - node.ratio) * 100}%">
      <PaneTree node={node.second} {tabId} hasSplits={true} {onClosePane} />
    </div>
  </div>
{/if}

<style>
  .pane-split { display: flex; width: 100%; height: 100%; }
  .pane-split.vertical { flex-direction: row; }
  .pane-split.horizontal { flex-direction: column; }
  .pane-child { display: flex; overflow: hidden; min-width: 0; min-height: 0; }
</style>
