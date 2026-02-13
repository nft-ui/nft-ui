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

<section class="card p-5 mb-6">
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold m-0" style="color: var(--text);">Raw Ruleset Data</h2>
    <button class="btn btn-sm btn-secondary" onclick={toggleExpanded}>
      {expanded ? 'Hide' : 'Show'}
    </button>
  </div>

  {#if expanded}
    <div class="mt-4">
      {#if loading}
        <div class="text-center py-5" style="color: var(--text-muted);">Loading raw ruleset...</div>
      {:else if error}
        <div class="alert alert-error flex justify-between items-center">
          <span>Error: {error}</span>
          <button class="btn btn-sm btn-danger" onclick={loadRawData}>Retry</button>
        </div>
      {:else if rawData}
        <pre class="code-block m-0 whitespace-pre-wrap break-words">{rawData}</pre>
      {:else}
        <div class="text-center py-5" style="color: var(--text-muted);">No data loaded</div>
      {/if}
    </div>
  {/if}
</section>
