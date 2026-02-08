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

  const variantMap = {
    info: 'variant-filled-primary',
    success: 'variant-filled-success',
    error: 'variant-filled-error',
    warning: 'variant-filled-warning',
  };

  let variant = $derived(variantMap[type] || variantMap.info);
</script>

<div
  class="alert {variant} flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg min-w-[250px] max-w-[400px] transition-all duration-200"
  class:opacity-0={!visible}
  class:translate-x-full={!visible}
  class:opacity-100={visible}
  class:translate-x-0={visible}
>
  <span class="flex-1 text-sm">{message}</span>
  <button
    class="bg-transparent border-none text-current text-xl p-0 cursor-pointer opacity-70 hover:opacity-100 leading-none"
    onclick={handleClose}
  >
    &times;
  </button>
</div>
