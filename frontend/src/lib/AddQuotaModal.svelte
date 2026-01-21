<script>
  import { onMount } from 'svelte';
  import { addQuota } from './api.js';
  import { loadQuotas, success, errorNotify, pauseRefresh, resumeRefresh } from './stores.js';
  import { parseBytes } from './utils.js';

  let { onclose } = $props();

  onMount(() => {
    pauseRefresh();
    return () => resumeRefresh();
  });

  let port = $state('');
  let quotaValue = $state('');
  let quotaUnit = $state('GB');
  let comment = $state('');
  let submitting = $state(false);
  let errors = $state({});

  function validate() {
    errors = {};

    const portNum = parseInt(port, 10);
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      errors.port = 'Port must be between 1 and 65535';
    }

    const quota = parseFloat(quotaValue);
    if (isNaN(quota) || quota <= 0) {
      errors.quota = 'Quota must be a positive number';
    }

    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validate()) return;

    submitting = true;
    try {
      const bytes = parseBytes(parseFloat(quotaValue), quotaUnit);
      await addQuota(parseInt(port, 10), bytes, comment);
      success('Quota rule added successfully');
      await loadQuotas();
      onclose?.();
    } catch (e) {
      errorNotify(`Failed to add quota: ${e.message}`);
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
    <h2>Add Quota Rule</h2>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="form-group">
        <label for="port">Port</label>
        <input
          id="port"
          type="number"
          bind:value={port}
          placeholder="8080"
          min="1"
          max="65535"
          class:error={errors.port}
        />
        {#if errors.port}
          <span class="error-text">{errors.port}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="quota">Quota Limit</label>
        <div class="form-row">
          <input
            id="quota"
            type="number"
            bind:value={quotaValue}
            placeholder="100"
            min="1"
            step="any"
            class:error={errors.quota}
          />
          <select bind:value={quotaUnit}>
            <option value="MB">MB</option>
            <option value="GB">GB</option>
            <option value="TB">TB</option>
          </select>
        </div>
        {#if errors.quota}
          <span class="error-text">{errors.quota}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="comment">Comment (optional)</label>
        <input
          id="comment"
          type="text"
          bind:value={comment}
          placeholder="block port after limit"
        />
      </div>

      <div class="modal-actions">
        <button type="button" class="btn-secondary" onclick={handleCancel}>
          Cancel
        </button>
        <button type="submit" class="btn-primary" disabled={submitting}>
          {submitting ? 'Adding...' : 'Add Rule'}
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
</style>
