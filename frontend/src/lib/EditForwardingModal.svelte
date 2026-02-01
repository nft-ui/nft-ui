<script>
  import { onMount } from 'svelte';
  import { editForwardingRule, pauseRefresh, resumeRefresh } from './stores.js';
  import { isValidIPv4 } from './utils.js';

  let { rule, onclose } = $props();

  onMount(() => {
    pauseRefresh();
    return () => resumeRefresh();
  });

  let dstIP = $state(rule.dst_ip);
  let dstPort = $state(rule.dst_port.toString());
  let protocol = $state(rule.protocol);
  let comment = $state(rule.comment || '');
  let limitMbps = $state((rule.limit_mbps || 0).toString());
  let submitting = $state(false);
  let errors = $state({});

  function validate() {
    const newErrors = {};

    if (!isValidIPv4(dstIP)) {
      newErrors.dstIP = 'Please enter a valid IPv4 address';
    }

    const dstPortNum = parseInt(dstPort, 10);
    if (isNaN(dstPortNum) || dstPortNum < 1 || dstPortNum > 65535) {
      newErrors.dstPort = 'Destination port must be between 1 and 65535';
    }

    const limitNum = parseInt(limitMbps, 10);
    if (isNaN(limitNum) || limitNum < 0) {
      newErrors.limitMbps = 'Limit must be 0 or positive (0 = no limit)';
    }

    errors = newErrors;
    return Object.keys(newErrors).length === 0;
  }

  async function handleSubmit() {
    if (!validate()) return;

    submitting = true;
    try {
      await editForwardingRule(
        rule.id,
        dstIP,
        parseInt(dstPort, 10),
        protocol,
        comment,
        parseInt(limitMbps, 10)
      );
      onclose?.();
    } catch (e) {
      // Error notification handled by store
    } finally {
      submitting = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      onclose?.();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-backdrop" onclick={onclose} role="presentation">
  <div class="modal" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
    <div class="modal-header">
      <h3>Edit Forwarding Rule</h3>
      <button class="close-btn" onclick={onclose}>&times;</button>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="form-group">
        <label for="srcPortDisplay">Source Port</label>
        <input type="text" id="srcPortDisplay" value={rule.src_port} disabled />
        <span class="help-text">Source port cannot be changed</span>
      </div>

      <div class="form-group">
        <label for="dstIP">Destination IP</label>
        <input
          type="text"
          id="dstIP"
          bind:value={dstIP}
          placeholder="e.g. 192.168.1.100"
          class:error={errors.dstIP}
        />
        {#if errors.dstIP}
          <span class="error-text">{errors.dstIP}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="dstPort">Destination Port</label>
        <input
          type="number"
          id="dstPort"
          bind:value={dstPort}
          placeholder="e.g. 22"
          min="1"
          max="65535"
          class:error={errors.dstPort}
        />
        {#if errors.dstPort}
          <span class="error-text">{errors.dstPort}</span>
        {/if}
      </div>

      <div class="form-group">
        <label for="protocol">Protocol</label>
        <select id="protocol" bind:value={protocol}>
          <option value="both">TCP + UDP</option>
          <option value="tcp">TCP only</option>
          <option value="udp">UDP only</option>
        </select>
      </div>

      <div class="form-group">
        <label for="comment">Comment (optional)</label>
        <input
          type="text"
          id="comment"
          bind:value={comment}
          placeholder="e.g. SSH tunnel"
          maxlength="100"
        />
      </div>

      <div class="form-group">
        <label for="limitMbps">Bandwidth Limit (Mbps)</label>
        <input
          type="number"
          id="limitMbps"
          bind:value={limitMbps}
          placeholder="0"
          min="0"
          class:error={errors.limitMbps}
        />
        {#if errors.limitMbps}
          <span class="error-text">{errors.limitMbps}</span>
        {/if}
        <span class="help-text">0 = no limit, or set max Mbps (e.g. 10, 100)</span>
      </div>

      <div class="note">
        Note: Editing will briefly disable and re-enable the rule.
      </div>

      <div class="form-actions">
        <button type="button" class="btn-secondary" onclick={onclose}>
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
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background-color: var(--color-surface);
    border-radius: 12px;
    padding: 24px;
    width: 100%;
    max-width: 420px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .close-btn {
    background: transparent;
    border: none;
    font-size: 24px;
    cursor: pointer;
    color: var(--color-text-muted);
    padding: 0;
    line-height: 1;
  }

  .close-btn:hover {
    color: var(--color-text);
  }

  .form-group {
    margin-bottom: 16px;
  }

  label {
    display: block;
    margin-bottom: 6px;
    font-size: 14px;
    font-weight: 500;
  }

  input,
  select {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    background-color: var(--color-bg);
    color: var(--color-text);
    font-size: 14px;
  }

  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  input:focus,
  select:focus {
    outline: none;
    border-color: var(--color-primary);
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

  .help-text {
    display: block;
    color: var(--color-text-muted);
    font-size: 12px;
    margin-top: 4px;
  }

  .note {
    font-size: 13px;
    color: var(--color-text-muted);
    background-color: var(--color-bg);
    padding: 10px 12px;
    border-radius: 6px;
    margin-top: 16px;
  }

  .form-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 24px;
  }
</style>
