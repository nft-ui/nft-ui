<script>
  import { modifyQuota } from './api.js';
  import { loadQuotas, success, errorNotify } from './stores.js';
  import { formatBytes, parseBytes } from './utils.js';

  let { quota, onclose } = $props();

  // Initialize with current value
  let quotaValue = $state('');
  let quotaUnit = $state('GB');
  let submitting = $state(false);
  let error = $state('');

  // Set initial values based on quota
  $effect(() => {
    const bytes = quota.quota_bytes;
    if (bytes >= 1000 * 1000 * 1000 * 1000) {
      quotaValue = (bytes / (1000 * 1000 * 1000 * 1000)).toString();
      quotaUnit = 'TB';
    } else if (bytes >= 1000 * 1000 * 1000) {
      quotaValue = (bytes / (1000 * 1000 * 1000)).toString();
      quotaUnit = 'GB';
    } else {
      quotaValue = (bytes / (1000 * 1000)).toString();
      quotaUnit = 'MB';
    }
  });

  function validate() {
    const value = parseFloat(quotaValue);
    if (isNaN(value) || value <= 0) {
      error = 'Quota must be a positive number';
      return false;
    }
    error = '';
    return true;
  }

  async function handleSubmit() {
    if (!validate()) return;

    submitting = true;
    try {
      const bytes = parseBytes(parseFloat(quotaValue), quotaUnit);
      await modifyQuota(quota.id, bytes);
      success('Quota modified successfully');
      await loadQuotas();
      onclose?.();
    } catch (e) {
      errorNotify(`Failed to modify quota: ${e.message}`);
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
    <h2>Edit Quota Rule</h2>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="form-group">
        <label for="port-display">Port</label>
        <input id="port-display" type="text" value={quota.port} disabled />
      </div>

      <div class="form-group">
        <label for="usage-display">Current Usage</label>
        <input id="usage-display" type="text" value={formatBytes(quota.used_bytes)} disabled />
      </div>

      <div class="form-group">
        <label for="quota">New Quota Limit</label>
        <div class="form-row">
          <input
            id="quota"
            type="number"
            bind:value={quotaValue}
            placeholder="100"
            min="1"
            step="any"
            class:error={error}
          />
          <select bind:value={quotaUnit}>
            <option value="MB">MB</option>
            <option value="GB">GB</option>
            <option value="TB">TB</option>
          </select>
        </div>
        {#if error}
          <span class="error-text">{error}</span>
        {/if}
        <span class="hint">Note: Modifying the quota will reset the used traffic to 0.</span>
      </div>

      <div class="modal-actions">
        <button type="button" class="btn-secondary" onclick={handleCancel}>
          Cancel
        </button>
        <button type="submit" class="btn-primary" disabled={submitting}>
          {submitting ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </form>
  </div>
</div>

<style>
  input:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

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
    color: var(--color-warning);
    font-size: 12px;
    margin-top: 8px;
  }
</style>
