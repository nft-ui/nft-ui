<script>
  let {
    title = 'Confirm',
    message = 'Are you sure?',
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    danger = false,
    onconfirm,
    oncancel,
  } = $props();

  function handleConfirm() {
    onconfirm?.();
  }

  function handleCancel() {
    oncancel?.();
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      handleCancel();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div
  class="fixed inset-0 bg-black/70 flex items-center justify-center z-[1000]"
  onclick={handleCancel}
  role="presentation"
>
  <div
    class="card p-6 min-w-[350px] max-w-[90%] bg-surface-100"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
  >
    <h2 class="text-xl font-semibold mb-5">{title}</h2>
    <p class="text-surface-600 mb-5 leading-relaxed">{message}</p>
    <div class="flex justify-end gap-3">
      <button class="btn variant-soft" onclick={handleCancel}>
        {cancelText}
      </button>
      <button
        class="btn {danger ? 'variant-filled-error' : 'variant-filled-primary'}"
        onclick={handleConfirm}
      >
        {confirmText}
      </button>
    </div>
  </div>
</div>
