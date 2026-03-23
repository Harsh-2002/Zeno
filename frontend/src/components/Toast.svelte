<script>
  let { message = '', visible = false } = $props();
</script>

<div class="toast" class:visible>
  {message}
</div>

<style>
  .toast {
    position: fixed;
    bottom: 32px; left: 50%; transform: translateX(-50%);
    background: var(--bg-secondary);
    color: var(--text-primary);
    padding: 8px 16px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    font-size: 12px;
    font-family: var(--font-mono);
    z-index: 2000;
    opacity: 0;
    transition: opacity 0.2s;
    pointer-events: none;
    box-shadow: 0 4px 16px var(--shadow);
  }
  .toast.visible { opacity: 1; }
</style>
