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

<section class="card p-5 mb-6 bg-surface-100">
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold m-0">Raw Ruleset Data</h2>
    <button class="btn btn-sm variant-soft" onclick={toggleExpanded}>
      {expanded ? 'Hide' : 'Show'}
    </button>
  </div>

  {#if expanded}
    <div class="mt-4">
      {#if loading}
        <div class="text-center text-surface-600 py-5">Loading raw ruleset...</div>
      {:else if error}
        <div class="alert variant-filled-error flex justify-between items-center">
          <span>Error: {error}</span>
          <button class="btn btn-sm variant-soft" onclick={loadRawData}>Retry</button>
        </div>
      {:else if rawData}
        <pre class="bg-surface-50 border border-surface-300 rounded-lg p-4 overflow-x-auto font-mono text-sm leading-relaxed m-0 whitespace-pre-wrap break-words">{rawData}</pre>
      {:else}
        <div class="text-center text-surface-600 py-5">No data loaded</div>
      {/if}
    </div>
  {/if}
</section>
