<script>
  import { addAllowedPort } from './stores.js';

  let { onclose } = $props();

  let port = $state('');
  let submitting = $state(false);
  let error = $state('');

  function validate() {
    const portNum = parseInt(port, 10);
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      error = 'Port must be between 1 and 65535';
      return false;
    }
    error = '';
    return true;
  }

  async function handleSubmit() {
    if (!validate()) return;

    submitting = true;
    try {
      await addAllowedPort(parseInt(port, 10));
      onclose?.();
    } catch (e) {
      // Error already shown by store
    } finally {
      submitting = false;
    }
  }

  function handleCancel() {
    onclose?.();
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      handleCancel();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-backdrop" onclick={handleCancel} role="presentation">
  <div class="modal" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
    <h2>Add Allowed Port</h2>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="form-group">
        <label for="port">Port Number</label>
        <input
          id="port"
          type="number"
          bind:value={port}
          placeholder="8080"
          min="1"
          max="65535"
          class:error={error}
          autofocus
        />
        {#if error}
          <span class="error-text">{error}</span>
        {/if}
        <span class="hint">This will add a TCP inbound rule: tcp dport &lt;port&gt; accept</span>
      </div>

      <div class="modal-actions">
        <button type="button" class="btn-secondary" onclick={handleCancel}>
          Cancel
        </button>
        <button type="submit" class="btn-primary" disabled={submitting}>
          {submitting ? 'Adding...' : 'Add Port'}
        </button>
      </div>
    </form>
  </div>
</div>

<style>
  input.error {
    border-color: var(--color-danger);
  }

  .error-text {
    display: block;
    color: var(--color-danger);
    font-size: 12px;
    margin-top: 4px;
  }

  .hint {
    display: block;
    color: var(--color-text-muted);
    font-size: 12px;
    margin-top: 8px;
  }
</style>
