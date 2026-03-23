<script>
  import PaneTree from './PaneTree.svelte';
  import { tabState } from '../lib/stores/tabs.svelte.js';

  let { onClosePane } = $props();
</script>

<div id="terminals">
  {#each tabState.tabs as tab (tab.id)}
    <div class="terminal-wrapper" class:active={tab.id === tabState.activeTabId}>
      <PaneTree node={tab.rootNode} tabId={tab.id} hasSplits={tab.rootNode.type === 'split'} {onClosePane} />
    </div>
  {/each}
</div>

<style>
  #terminals { flex: 1; position: relative; overflow: hidden; }
  .terminal-wrapper {
    position: absolute; top: 0; left: 0; width: 100%; height: 100%;
    display: none; padding: 3px;
  }
  .terminal-wrapper.active { display: flex; }
</style>
