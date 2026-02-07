<script>
  import { fetchRawRuleset } from './api.js';

  let rawData = $state('');
  let loading = $state(false);
  let error = $state(null);
  let expanded = $state(false);

  async function loadRawData() {
    loading = true;
    error = null;
    try {
      const response = await fetchRawRuleset();
      rawData = response.data;
    } catch (err) {
      error = err.message;
      rawData = '';
    } finally {
      loading = false;
    }
  }

  function toggleExpanded() {
    expanded = !expanded;
    if (expanded && !rawData && !loading) {
      loadRawData();
    }
  }
</script>

<section class="section">
  <div class="section-header">
    <h2>Raw Ruleset Data</h2>
    <button class="btn-secondary" onclick={toggleExpanded}>
      {expanded ? 'Hide' : 'Show'}
    </button>
  </div>

  {#if expanded}
    <div class="content">
      {#if loading}
        <div class="loading">Loading raw ruleset...</div>
      {:else if error}
        <div class="error">
          <span>Error: {error}</span>
          <button class="btn-secondary" onclick={loadRawData}>Retry</button>
        </div>
      {:else if rawData}
        <pre class="raw-data">{rawData}</pre>
      {:else}
        <div class="empty">No data loaded</div>
      {/if}
    </div>
  {/if}
</section>

<style>
  .section {
    background-color: var(--color-surface);
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .content {
    margin-top: 16px;
  }

  .loading,
  .empty {
    text-align: center;
    color: var(--color-text-secondary);
    padding: 20px;
  }

  .error {
    background-color: rgba(248, 113, 113, 0.2);
    border: 1px solid var(--color-danger);
    border-radius: 8px;
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .raw-data {
    background-color: var(--color-bg);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    padding: 16px;
    overflow-x: auto;
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
    line-height: 1.6;
    color: var(--color-text);
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  button {
    padding: 8px 16px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
  }

  .btn-secondary {
    background-color: var(--color-border);
    color: var(--color-text);
  }

  .btn-secondary:hover {
    background-color: #d1d5db;
  }

  .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
