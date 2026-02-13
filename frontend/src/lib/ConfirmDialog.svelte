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
  class="modal-backdrop"
  onclick={handleCancel}
  role="presentation"
>
  <div
    class="modal min-w-[350px] max-w-[90%]"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
  >
    <h2 class="text-xl font-semibold mb-5" style="color: var(--text);">{title}</h2>
    <p class="mb-5 leading-relaxed" style="color: var(--text-muted);">{message}</p>
    <div class="flex justify-end gap-3">
      <button class="btn btn-secondary" onclick={handleCancel}>
        {cancelText}
      </button>
      <button
        class="btn {danger ? 'btn-danger' : 'btn-primary'}"
        onclick={handleConfirm}
      >
        {confirmText}
      </button>
    </div>
  </div>
</div>
