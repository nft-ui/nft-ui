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
  let progressColor = $derived(result ? getProgressColor(result.usage_percent) : 'var(--primary)');
  let statusColor = $derived(result ? getStatusColor(result.status) : 'var(--primary)');
</script>

<div class="min-h-screen flex items-center justify-center p-5" style="background-color: var(--bg);">
  <div class="max-w-[480px] w-full text-center">
    <h1 class="text-[28px] font-semibold mb-2" style="color: var(--text);">Port Usage Query</h1>
    <p class="mb-8" style="color: var(--text-muted);">Enter your token to check current bandwidth usage</p>

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
        <button type="submit" class="btn btn-primary px-6 whitespace-nowrap" disabled={loading}>
          {loading ? 'Checking...' : 'Check Usage'}
        </button>
      </div>
    </form>

    {#if error}
      <div class="alert alert-error mb-6">{error}</div>
    {/if}

    {#if result}
      <div class="card p-6 text-left">
        <div class="flex justify-between items-center mb-6">
          <span class="text-2xl font-semibold" style="color: var(--text);">Port {result.port}</span>
          <span
            class="badge px-4 py-1.5 text-sm font-medium capitalize"
            style="background-color: {statusColor}; color: white; border-color: {statusColor};"
          >
            {result.status}
          </span>
        </div>

        <div class="flex flex-col gap-6">
          <div class="w-full">
            <div class="flex justify-between items-center mb-2">
              <span class="text-sm" style="color: var(--text-muted);">Usage</span>
              <span class="text-lg font-semibold" style="color: var(--text);">{formatPercent(result.usage_percent)}</span>
            </div>
            <div class="w-full h-3 rounded-full overflow-hidden" style="background-color: var(--border);">
              <div
                class="h-full rounded-full transition-all duration-500"
                style="width: {ringPercent}%; background-color: {progressColor};"
              ></div>
            </div>
          </div>

          <div class="flex-1 w-full">
            <div class="flex justify-between py-2" style="border-bottom: 1px solid var(--border);">
              <span style="color: var(--text-muted);">Used:</span>
              <span class="font-medium" style="color: var(--text);">{formatBytes(result.used_bytes)}</span>
            </div>
            <div class="flex justify-between py-2" style="border-bottom: 1px solid var(--border);">
              <span style="color: var(--text-muted);">Quota:</span>
              <span class="font-medium" style="color: var(--text);">{formatBytes(result.quota_bytes)}</span>
            </div>
            {#if result.comment}
              <div class="flex justify-between py-2">
                <span style="color: var(--text-muted);">Comment:</span>
                <span class="font-medium" style="color: var(--text);">{result.comment}</span>
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
