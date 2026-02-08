<script>
  import { onMount } from 'svelte';
  import { formatBytes, formatPercent, getProgressColor, getStatusColor } from './utils.js';

  let token = $state('');
  let result = $state(null);
  let error = $state(null);
  let loading = $state(false);

  onMount(() => {
    // Check URL for token parameter
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get('token');
    if (urlToken) {
      token = urlToken;
      handleQuery();
    }
  });

  async function handleQuery() {
    if (!token || token.length !== 8) {
      error = 'Please enter a valid 8-character token';
      return;
    }

    loading = true;
    error = null;
    result = null;

    try {
      const response = await fetch(`/api/v1/public/query/${encodeURIComponent(token)}`);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Query failed');
      }

      result = data;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function handleSubmit(e) {
    e.preventDefault();
    handleQuery();
  }

  let ringPercent = $derived(result ? Math.min(result.usage_percent, 100) : 0);
  let progressColor = $derived(result ? getProgressColor(result.usage_percent) : 'var(--color-ok)');
  let statusColor = $derived(result ? getStatusColor(result.status) : 'var(--color-ok)');
</script>

<div class="min-h-screen flex items-center justify-center p-5 bg-surface-50">
  <div class="max-w-[480px] w-full text-center">
    <h1 class="text-[28px] font-semibold mb-2">Port Usage Query</h1>
    <p class="text-surface-600 mb-8">Enter your token to check current bandwidth usage</p>

    <form onsubmit={handleSubmit}>
      <div class="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          class="input flex-1 text-lg font-mono tracking-[2px] uppercase text-center"
          bind:value={token}
          placeholder="Enter 8-character token"
          maxlength="8"
          autocomplete="off"
          spellcheck="false"
        />
        <button type="submit" class="btn variant-filled-primary px-6 whitespace-nowrap" disabled={loading}>
          {loading ? 'Checking...' : 'Check Usage'}
        </button>
      </div>
    </form>

    {#if error}
      <div class="alert variant-filled-error mb-6">{error}</div>
    {/if}

    {#if result}
      <div class="card p-6 text-left bg-surface-100">
        <div class="flex justify-between items-center mb-6">
          <span class="text-2xl font-semibold">Port {result.port}</span>
          <span
            class="badge variant-filled px-4 py-1.5 text-sm font-medium capitalize text-white"
            style="background-color: {statusColor}"
          >
            {result.status}
          </span>
        </div>

        <div class="flex flex-col sm:flex-row gap-6 items-center">
          <div class="relative w-[120px] h-[120px] flex-shrink-0">
            <svg class="w-full h-full -rotate-90" viewBox="0 0 120 120">
              <circle class="fill-none stroke-surface-200 stroke-[8]" cx="60" cy="60" r="52" />
              <circle
                class="fill-none stroke-[8] stroke-round transition-all duration-500"
                cx="60"
                cy="60"
                r="52"
                style="stroke: {progressColor}; stroke-dasharray: {ringPercent * 3.27}, 327;"
              />
            </svg>
            <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center">
              <span class="text-xl font-semibold">{formatPercent(result.usage_percent)}</span>
            </div>
          </div>

          <div class="flex-1 w-full">
            <div class="flex justify-between py-2 border-b border-surface-300">
              <span class="text-surface-600">Used:</span>
              <span class="font-medium">{formatBytes(result.used_bytes)}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-surface-300">
              <span class="text-surface-600">Quota:</span>
              <span class="font-medium">{formatBytes(result.quota_bytes)}</span>
            </div>
            {#if result.comment}
              <div class="flex justify-between py-2">
                <span class="text-surface-600">Comment:</span>
                <span class="font-medium">{result.comment}</span>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .input::placeholder {
    text-transform: none;
    letter-spacing: normal;
    font-size: 14px;
  }
</style>
