<script>
  import { onMount } from 'svelte';

  let { message = '', type = 'info', onclose } = $props();

  let visible = $state(false);

  onMount(() => {
    // Trigger animation
    requestAnimationFrame(() => {
      visible = true;
    });
  });

  function handleClose() {
    visible = false;
    setTimeout(() => {
      onclose?.();
    }, 200);
  }

  const colorMap = {
    info: 'var(--primary)',
    success: 'var(--success)',
    error: 'var(--danger)',
    warning: 'var(--warning)',
  };

  let borderColor = $derived(colorMap[type] || colorMap.info);
</script>

<div
  class="flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg min-w-[250px] max-w-[400px] transition-all duration-200"
  class:opacity-0={!visible}
  class:translate-x-full={!visible}
  class:opacity-100={visible}
  class:translate-x-0={visible}
  style="background-color: var(--surface); border: 1px solid var(--border); border-left: 4px solid {borderColor}; backdrop-filter: blur(8px);"
>
  <span class="flex-1 text-sm" style="color: var(--text);">{message}</span>
  <button
    class="bg-transparent border-none text-xl p-0 cursor-pointer leading-none transition-opacity"
    style="color: var(--text-muted); opacity: 0.7;"
    onmouseover={(e) => e.currentTarget.style.opacity = '1'}
    onmouseout={(e) => e.currentTarget.style.opacity = '0.7'}
    onclick={handleClose}
  >
    &times;
  </button>
</div>
