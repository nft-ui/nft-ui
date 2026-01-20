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

<div class="modal-backdrop" onclick={handleCancel} role="presentation">
  <div class="modal confirm-dialog" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
    <h2>{title}</h2>
    <p>{message}</p>
    <div class="modal-actions">
      <button class="btn-secondary" onclick={handleCancel}>
        {cancelText}
      </button>
      <button
        class={danger ? 'btn-danger' : 'btn-primary'}
        onclick={handleConfirm}
      >
        {confirmText}
      </button>
    </div>
  </div>
</div>

<style>
  .confirm-dialog {
    min-width: 350px;
  }

  .confirm-dialog p {
    color: var(--color-text-muted);
    margin: 0 0 20px 0;
    line-height: 1.6;
  }
</style>
