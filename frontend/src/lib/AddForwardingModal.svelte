<script>
  import { onMount } from 'svelte';
  import { addForwardingRule, pauseRefresh, resumeRefresh } from './stores.js';
  import { isValidIPv4 } from './utils.js';

  let { onclose } = $props();

  onMount(() => {
    pauseRefresh();
    return () => resumeRefresh();
  });

  let srcPort = $state('');
  let dstIP = $state('');
  let dstPort = $state('');
  let protocol = $state('both');
  let comment = $state('');
  let limitMbps = $state('0');
  let submitting = $state(false);
  let errors = $state({});

  function validate() {
    const newErrors = {};

    const srcPortNum = parseInt(srcPort, 10);
    if (isNaN(srcPortNum) || srcPortNum < 1 || srcPortNum > 65535) {
      newErrors.srcPort = 'Source port must be between 1 and 65535';
    }

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
      await addForwardingRule(
        parseInt(srcPort, 10),
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

<div
  class="fixed inset-0 bg-black/70 flex items-center justify-center z-[1000]"
  onclick={onclose}
  role="presentation"
>
  <div
    class="card p-6 w-full max-w-[420px] max-h-[90vh] overflow-y-auto bg-surface-100"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
  >
    <div class="flex justify-between items-center mb-5">
      <h3 class="text-lg font-semibold">Add Forwarding Rule</h3>
      <button
        class="bg-transparent border-none text-2xl cursor-pointer text-surface-600 hover:text-surface-900 p-0 leading-none"
        onclick={onclose}
      >
        &times;
      </button>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="mb-4">
        <label for="srcPort" class="label mb-2">
          <span>Source Port</span>
        </label>
        <input
          type="number"
          id="srcPort"
          class="input"
          class:input-error={errors.srcPort}
          bind:value={srcPort}
          placeholder="e.g. 12103"
          min="1"
          max="65535"
        />
        {#if errors.srcPort}
          <span class="text-error-500 text-xs mt-1 block">{errors.srcPort}</span>
        {/if}
        <span class="text-surface-600 text-xs mt-1 block">The local port to forward from</span>
      </div>

      <div class="mb-4">
        <label for="dstIP" class="label mb-2">
          <span>Destination IP</span>
        </label>
        <input
          type="text"
          id="dstIP"
          class="input"
          class:input-error={errors.dstIP}
          bind:value={dstIP}
          placeholder="e.g. 192.168.1.100"
        />
        {#if errors.dstIP}
          <span class="text-error-500 text-xs mt-1 block">{errors.dstIP}</span>
        {/if}
        <span class="text-surface-600 text-xs mt-1 block">The target IP address</span>
      </div>

      <div class="mb-4">
        <label for="dstPort" class="label mb-2">
          <span>Destination Port</span>
        </label>
        <input
          type="number"
          id="dstPort"
          class="input"
          class:input-error={errors.dstPort}
          bind:value={dstPort}
          placeholder="e.g. 22"
          min="1"
          max="65535"
        />
        {#if errors.dstPort}
          <span class="text-error-500 text-xs mt-1 block">{errors.dstPort}</span>
        {/if}
        <span class="text-surface-600 text-xs mt-1 block">The target port</span>
      </div>

      <div class="mb-4">
        <label for="protocol" class="label mb-2">
          <span>Protocol</span>
        </label>
        <select id="protocol" class="select" bind:value={protocol}>
          <option value="both">TCP + UDP</option>
          <option value="tcp">TCP only</option>
          <option value="udp">UDP only</option>
        </select>
      </div>

      <div class="mb-4">
        <label for="comment" class="label mb-2">
          <span>Comment (optional)</span>
        </label>
        <input
          type="text"
          id="comment"
          class="input"
          bind:value={comment}
          placeholder="e.g. SSH tunnel"
          maxlength="100"
        />
      </div>

      <div class="mb-6">
        <label for="limitMbps" class="label mb-2">
          <span>Bandwidth Limit (Mbps)</span>
        </label>
        <input
          type="number"
          id="limitMbps"
          class="input"
          class:input-error={errors.limitMbps}
          bind:value={limitMbps}
          placeholder="0"
          min="0"
        />
        {#if errors.limitMbps}
          <span class="text-error-500 text-xs mt-1 block">{errors.limitMbps}</span>
        {/if}
        <span class="text-surface-600 text-xs mt-1 block">0 = no limit, or set max Mbps (e.g. 10, 100)</span>
      </div>

      <div class="flex justify-end gap-3">
        <button type="button" class="btn variant-soft" onclick={onclose}>
          Cancel
        </button>
        <button type="submit" class="btn variant-filled-primary" disabled={submitting}>
          {submitting ? 'Adding...' : 'Add Rule'}
        </button>
      </div>
    </form>
  </div>
</div>
