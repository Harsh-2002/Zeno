<script>
  let { direction, onResize, onResizeEnd } = $props();

  function handleMouseDown(e) {
    e.preventDefault();
    const parent = e.target.parentElement;
    const rect = parent.getBoundingClientRect();
    const totalSize = direction === 'vertical' ? rect.width : rect.height;

    function onMove(e2) {
      const currentPos = direction === 'vertical' ? e2.clientX : e2.clientY;
      const offset = direction === 'vertical' ? (currentPos - rect.left) : (currentPos - rect.top);
      let newRatio = offset / totalSize;
      newRatio = Math.max(0.1, Math.min(0.9, newRatio));
      onResize(newRatio);
    }

    function onUp() {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      onResizeEnd?.();
    }

    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
    document.body.style.cursor = direction === 'vertical' ? 'col-resize' : 'row-resize';
    document.body.style.userSelect = 'none';
  }

  function handleDblClick(e) {
    e.preventDefault();
    // Reset to equal split
    onResize(0.5);
    onResizeEnd?.();
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="split-divider {direction}"
  onmousedown={handleMouseDown}
  ondblclick={handleDblClick}></div>

<style>
  .split-divider {
    flex-shrink: 0; z-index: 5;
    background: transparent;
    transition: background 0.2s;
  }
  .split-divider.vertical {
    width: 6px; cursor: col-resize;
  }
  .split-divider.horizontal {
    height: 6px; cursor: row-resize;
  }
  .split-divider:hover { background: var(--border-focus); }

  @media (hover: none) and (pointer: coarse) {
    .split-divider.vertical { width: 10px; margin: 0; }
    .split-divider.horizontal { height: 10px; margin: 0; }
  }
</style>
