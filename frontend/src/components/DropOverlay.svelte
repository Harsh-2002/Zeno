<script>
  let { visible = false } = $props();
</script>

{#if visible}
<div class="drop-overlay">Drop file to upload</div>
{/if}

<style>
  .drop-overlay {
    position: absolute; inset: 4px;
    background: var(--accent-subtle);
    border: 2px dashed var(--accent);
    border-radius: var(--radius);
    z-index: 20;
    pointer-events: none;
    display: flex; align-items: center; justify-content: center;
    color: var(--accent);
    font-size: 14px; font-weight: 600;
    font-family: var(--font-ui);
  }
</style>
