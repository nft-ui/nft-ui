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
</script>

<div class="toast toast-{type}" class:visible>
  <span class="message">{message}</span>
  <button class="close-btn" onclick={handleClose}>&times;</button>
</div>

<style>
  .toast {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    opacity: 0;
    transform: translateX(100%);
    transition: all 0.2s ease;
    min-width: 250px;
    max-width: 400px;
  }

  .toast.visible {
    opacity: 1;
    transform: translateX(0);
  }

  .toast-info {
    background-color: var(--color-primary);
    color: white;
  }

  .toast-success {
    background-color: var(--color-success);
    color: white;
  }

  .toast-error {
    background-color: var(--color-danger);
    color: white;
  }

  .toast-warning {
    background-color: var(--color-warning);
    color: black;
  }

  .message {
    flex: 1;
    font-size: 14px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: inherit;
    font-size: 20px;
    padding: 0;
    cursor: pointer;
    opacity: 0.7;
    line-height: 1;
  }

  .close-btn:hover {
    opacity: 1;
  }
</style>
