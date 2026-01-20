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

<div class="query-page">
  <div class="query-container">
    <h1>Port Usage Query</h1>
    <p class="subtitle">Enter your token to check current bandwidth usage</p>

    <form onsubmit={handleSubmit}>
      <div class="input-group">
        <input
          type="text"
          bind:value={token}
          placeholder="Enter 8-character token"
          maxlength="8"
          class="token-input"
          autocomplete="off"
          spellcheck="false"
        />
        <button type="submit" class="btn-primary" disabled={loading}>
          {loading ? 'Checking...' : 'Check Usage'}
        </button>
      </div>
    </form>

    {#if error}
      <div class="error-message">{error}</div>
    {/if}

    {#if result}
      <div class="result-card">
        <div class="result-header">
          <span class="port-label">Port {result.port}</span>
          <span class="status-badge" style="background-color: {statusColor}">
            {result.status}
          </span>
        </div>

        <div class="usage-display">
          <div class="ring-container">
            <svg class="progress-ring" viewBox="0 0 120 120">
              <circle class="ring-bg" cx="60" cy="60" r="52" />
              <circle
                class="ring-progress"
                cx="60" cy="60" r="52"
                stroke={progressColor}
                stroke-dasharray="{ringPercent * 3.27}, 327"
              />
            </svg>
            <div class="ring-text">
              <span class="percent">{formatPercent(result.usage_percent)}</span>
            </div>
          </div>

          <div class="usage-details">
            <div class="usage-row">
              <span class="label">Used:</span>
              <span class="value">{formatBytes(result.used_bytes)}</span>
            </div>
            <div class="usage-row">
              <span class="label">Quota:</span>
              <span class="value">{formatBytes(result.quota_bytes)}</span>
            </div>
            {#if result.comment}
              <div class="usage-row">
                <span class="label">Comment:</span>
                <span class="value">{result.comment}</span>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .query-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background-color: var(--color-bg);
  }

  .query-container {
    max-width: 480px;
    width: 100%;
    text-align: center;
  }

  h1 {
    font-size: 28px;
    font-weight: 600;
    margin-bottom: 8px;
    color: var(--color-text);
  }

  .subtitle {
    color: var(--color-text-muted);
    margin-bottom: 32px;
  }

  .input-group {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
  }

  .token-input {
    flex: 1;
    padding: 12px 16px;
    font-size: 18px;
    font-family: monospace;
    letter-spacing: 2px;
    text-transform: uppercase;
    border: 2px solid var(--color-border);
    border-radius: 8px;
    background-color: var(--color-surface);
    color: var(--color-text);
    text-align: center;
  }

  .token-input:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .token-input::placeholder {
    text-transform: none;
    letter-spacing: normal;
    font-size: 14px;
  }

  .btn-primary {
    padding: 12px 24px;
    font-size: 16px;
    font-weight: 500;
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: background-color 0.2s;
    white-space: nowrap;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: var(--color-primary-hover);
  }

  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .error-message {
    background-color: rgba(248, 113, 113, 0.2);
    border: 1px solid var(--color-danger);
    border-radius: 8px;
    padding: 12px 16px;
    margin-bottom: 24px;
    color: var(--color-danger);
  }

  .result-card {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 12px;
    padding: 24px;
    text-align: left;
  }

  .result-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .port-label {
    font-size: 24px;
    font-weight: 600;
  }

  .status-badge {
    padding: 6px 16px;
    border-radius: 20px;
    font-size: 14px;
    font-weight: 500;
    color: white;
    text-transform: capitalize;
  }

  .usage-display {
    display: flex;
    gap: 24px;
    align-items: center;
  }

  .ring-container {
    position: relative;
    width: 120px;
    height: 120px;
    flex-shrink: 0;
  }

  .progress-ring {
    transform: rotate(-90deg);
  }

  .ring-bg {
    fill: none;
    stroke: var(--color-bg);
    stroke-width: 8;
  }

  .ring-progress {
    fill: none;
    stroke-width: 8;
    stroke-linecap: round;
    transition: stroke-dasharray 0.5s ease;
  }

  .ring-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
  }

  .ring-text .percent {
    font-size: 20px;
    font-weight: 600;
  }

  .usage-details {
    flex: 1;
  }

  .usage-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid var(--color-border);
  }

  .usage-row:last-child {
    border-bottom: none;
  }

  .usage-row .label {
    color: var(--color-text-muted);
  }

  .usage-row .value {
    font-weight: 500;
  }

  @media (max-width: 480px) {
    .input-group {
      flex-direction: column;
    }

    .usage-display {
      flex-direction: column;
    }

    .ring-container {
      margin: 0 auto;
    }
  }
</style>
