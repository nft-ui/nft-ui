<script>
  import { onMount } from 'svelte';
  import { modifyQuota } from './api.js';
  import { loadQuotas, success, errorNotify, pauseRefresh, resumeRefresh } from './stores.js';
  import { formatBytes, parseBytes } from './utils.js';

  let { quota, onclose } = $props();

  onMount(() => {
    pauseRefresh();
    return () => resumeRefresh();
  });

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

<div
  class="modal-backdrop"
  onclick={handleCancel}
  role="presentation"
>
  <div
    class="modal"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
  >
    <h2 class="text-xl font-semibold mb-5" style="color: var(--text);">Edit Quota Rule</h2>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="mb-4">
        <label for="port-display" class="label">
          <span>Port</span>
        </label>
        <input id="port-display" type="text" class="input" value={quota.port} disabled />
      </div>

      <div class="mb-4">
        <label for="usage-display" class="label">
          <span>Current Usage</span>
        </label>
        <input id="usage-display" type="text" class="input" value={formatBytes(quota.used_bytes)} disabled />
      </div>

      <div class="mb-6">
        <label for="quota" class="label">
          <span>New Quota Limit</span>
        </label>
        <div class="flex gap-3">
          <input
            id="quota"
            type="number"
            class="input flex-1"
            class:input-error={error}
            bind:value={quotaValue}
            placeholder="100"
            min="1"
            step="any"
          />
          <select class="select" bind:value={quotaUnit}>
            <option value="MB">MB</option>
            <option value="GB">GB</option>
            <option value="TB">TB</option>
          </select>
        </div>
        {#if error}
          <span class="text-xs mt-1 block" style="color: var(--danger);">{error}</span>
        {/if}
        <span class="text-xs mt-2 block" style="color: var(--warning);">
          Note: Modifying the quota will reset the used traffic to 0.
        </span>
      </div>

      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" onclick={handleCancel}>
          Cancel
        </button>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </form>
  </div>
</div>
