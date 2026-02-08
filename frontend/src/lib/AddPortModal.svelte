<script>
  import { onMount } from 'svelte';
  import { addAllowedPort, pauseRefresh, resumeRefresh } from './stores.js';

  let { onclose } = $props();

  onMount(() => {
    pauseRefresh();
    return () => resumeRefresh();
  });

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

<div
  class="fixed inset-0 bg-black/70 flex items-center justify-center z-[1000]"
  onclick={handleCancel}
  role="presentation"
>
  <div
    class="card p-6 min-w-[400px] max-w-[90%] bg-surface-100"
    onclick={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
  >
    <h2 class="text-xl font-semibold mb-5">Add Allowed Port</h2>

    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      <div class="mb-6">
        <label for="port" class="label mb-2">
          <span>Port Number</span>
        </label>
        <input
          id="port"
          type="number"
          class="input"
          class:input-error={error}
          bind:value={port}
          placeholder="8080"
          min="1"
          max="65535"
          autofocus
        />
        {#if error}
          <span class="text-error-500 text-xs mt-1 block">{error}</span>
        {/if}
        <span class="text-surface-600 text-xs mt-2 block">
          This will add a TCP inbound rule: tcp dport &lt;port&gt; accept
        </span>
      </div>

      <div class="flex justify-end gap-3">
        <button type="button" class="btn variant-soft" onclick={handleCancel}>
          Cancel
        </button>
        <button type="submit" class="btn variant-filled-primary" disabled={submitting}>
          {submitting ? 'Adding...' : 'Add Port'}
        </button>
      </div>
    </form>
  </div>
</div>
