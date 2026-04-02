<script>
  import { THEMES } from '../lib/constants/themes.js';
  import {
    themeState, setTheme, setFontSize, setFontFamily, setFontLigatures,
    setCursorStyle, setCursorBlink, setLineHeight, setScrollback, setCopyOnSelect
  } from '../lib/stores/theme.svelte.js';

  let { visible = false, onClose } = $props();

  const themeKeys = Object.keys(THEMES);
  const fontFamilies = ['JetBrains Mono', 'Fira Code', 'Cascadia Code', 'Menlo', 'Monaco', 'Courier New'];
  const cursorStyles = ['block', 'underline', 'bar'];
  const lineHeights = [1.0, 1.1, 1.2, 1.3, 1.4];
  const scrollbackOptions = [10000, 50000, 100000, 500000];
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="overlay" onclick={onClose}></div>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="panel" oncontextmenu={(e) => e.stopPropagation()}>
    <div class="header">
      <span class="title">Settings</span>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <span class="close" onclick={onClose}>&times;</span>
    </div>

    <div class="body">
      <div class="section-label">Appearance</div>

      <label class="row">
        <span class="label">Theme</span>
        <select class="control" value={themeState.themeId} onchange={(e) => setTheme(e.target.value)}>
          {#each themeKeys as key}
            <option value={key}>{THEMES[key].name}</option>
          {/each}
        </select>
      </label>

      <label class="row">
        <span class="label">Font</span>
        <select class="control" value={themeState.fontFamily} onchange={(e) => setFontFamily(e.target.value)}>
          {#each fontFamilies as f}
            <option value={f}>{f}</option>
          {/each}
        </select>
      </label>

      <div class="row">
        <span class="label">Font Size</span>
        <div class="stepper">
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <span class="step-btn" onclick={() => setFontSize(themeState.fontSize - 1)}>-</span>
          <span class="step-val">{themeState.fontSize}px</span>
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <span class="step-btn" onclick={() => setFontSize(themeState.fontSize + 1)}>+</span>
        </div>
      </div>

      <label class="row">
        <span class="label">Cursor Style</span>
        <select class="control" value={themeState.cursorStyle} onchange={(e) => setCursorStyle(e.target.value)}>
          {#each cursorStyles as s}
            <option value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
          {/each}
        </select>
      </label>

      <label class="row">
        <span class="label">Cursor Blink</span>
        <input type="checkbox" class="toggle" checked={themeState.cursorBlink}
          onchange={(e) => setCursorBlink(e.target.checked)} />
      </label>

      <label class="row">
        <span class="label">Ligatures</span>
        <input type="checkbox" class="toggle" checked={themeState.fontLigatures}
          onchange={(e) => setFontLigatures(e.target.checked)} />
      </label>

      <label class="row">
        <span class="label">Line Height</span>
        <select class="control" value={themeState.lineHeight} onchange={(e) => setLineHeight(parseFloat(e.target.value))}>
          {#each lineHeights as lh}
            <option value={lh}>{lh.toFixed(1)}</option>
          {/each}
        </select>
      </label>

      <div class="section-label">Behavior</div>

      <label class="row">
        <span class="label">Scrollback</span>
        <select class="control" value={themeState.scrollback} onchange={(e) => setScrollback(parseInt(e.target.value))}>
          {#each scrollbackOptions as sb}
            <option value={sb}>{(sb / 1000).toFixed(0)}K lines</option>
          {/each}
        </select>
      </label>

      <label class="row">
        <span class="label">Copy on Select</span>
        <input type="checkbox" class="toggle" checked={themeState.copyOnSelect}
          onchange={(e) => setCopyOnSelect(e.target.checked)} />
      </label>

    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed; inset: 0; z-index: 900;
    background: rgba(0,0,0,0.3);
  }
  .panel {
    position: fixed; top: 0; right: 0; bottom: 0;
    width: 320px; max-width: 90vw;
    z-index: 950;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border);
    display: flex; flex-direction: column;
    box-shadow: -2px 0 8px rgba(0,0,0,0.15);
    animation: slide-in 0.15s ease;
  }
  @keyframes slide-in {
    from { transform: translateX(100%); }
    to { transform: translateX(0); }
  }
  .header {
    display: flex; align-items: center; justify-content: space-between;
    padding: 16px 20px; border-bottom: 1px solid var(--border);
  }
  .title { font-size: 15px; font-weight: 600; color: var(--text-primary); }
  .close {
    font-size: 20px; color: var(--text-muted); cursor: pointer;
    width: 28px; height: 28px; display: flex; align-items: center;
    justify-content: center; border-radius: var(--radius-xs);
    transition: background 0.1s;
  }
  .close:hover { background: var(--bg-hover); color: var(--text-primary); }
  .body { flex: 1; overflow-y: auto; padding: 8px 0; }
  .section-label {
    padding: 12px 20px 6px; font-size: 11px; font-weight: 600;
    color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px;
  }
  .row {
    display: flex; align-items: center; justify-content: space-between;
    padding: 10px 20px; min-height: 40px;
  }
  .row.info { opacity: 0.7; }
  .label { font-size: 13px; color: var(--text-primary); }
  .value { font-size: 13px; color: var(--text-muted); }
  .value.mono { font-family: var(--font-mono); font-size: 11px; }
  .control {
    color: var(--text-primary);
    border: 1px solid var(--border); border-radius: var(--radius-xs);
    padding: 7px 32px 7px 10px; font-size: 13px; font-family: var(--font-ui);
    outline: none; cursor: pointer; min-width: 130px;
    appearance: none; -webkit-appearance: none;
    background: var(--bg-hover) url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6' viewBox='0 0 10 6'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%23888' stroke-width='1.5' fill='none' stroke-linecap='round'/%3E%3C/svg%3E") no-repeat right 10px center;
  }
  .control:focus { border-color: var(--border-focus); }
  .toggle {
    width: 40px; height: 22px; appearance: none; -webkit-appearance: none;
    background: var(--border); border-radius: 11px; position: relative;
    cursor: pointer; transition: background 0.2s; border: none; outline: none;
    flex-shrink: 0;
  }
  .toggle::after {
    content: ''; position: absolute; top: 3px; left: 3px;
    width: 16px; height: 16px; border-radius: 50%;
    background: var(--text-muted); transition: transform 0.2s, background 0.2s;
  }
  .toggle:checked { background: #4CAF50; }
  .toggle:checked::after { transform: translateX(18px); background: #fff; }
  .stepper {
    display: flex; align-items: center; gap: 0;
    border: 1px solid var(--border); border-radius: var(--radius-xs);
    overflow: hidden; height: 32px;
  }
  .step-btn {
    width: 32px; height: 100%; display: flex; align-items: center;
    justify-content: center; cursor: pointer; color: var(--text-primary);
    font-size: 14px; transition: background 0.1s;
    background: var(--bg-hover);
  }
  .step-btn:hover { background: var(--border-focus); }
  .step-val {
    min-width: 48px; text-align: center; font-size: 13px;
    color: var(--text-primary); font-family: var(--font-mono);
    background: transparent; padding: 0 4px;
  }

  @media (max-width: 480px) { .panel { width: 100%; } }
</style>
