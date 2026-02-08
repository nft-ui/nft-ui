<script>
  import {
    sortedQuotas,
    loading,
    selectedIds,
    hasSelection,
    selectedCount,
    readOnly,
    clearSelection,
    selectAll,
    loadQuotas,
    success,
    errorNotify,
  } from './stores.js';
  import { batchResetQuotas } from './api.js';
  import QuotaItem from './QuotaItem.svelte';
  import AddQuotaModal from './AddQuotaModal.svelte';
  import ConfirmDialog from './ConfirmDialog.svelte';

  let showAddModal = $state(false);
  let showBatchResetConfirm = $state(false);
  let batchResetting = $state(false);

  async function handleBatchReset() {
    batchResetting = true;
    try {
      const ids = Array.from($selectedIds);
      await batchResetQuotas(ids);
      success(`Reset ${ids.length} quota(s) successfully`);
      clearSelection();
      await loadQuotas();
    } catch (e) {
      errorNotify(`Failed to reset quotas: ${e.message}`);
    } finally {
      batchResetting = false;
      showBatchResetConfirm = false;
    }
  }
</script>

<div class="card mb-6 overflow-hidden bg-surface-100">
  <!-- Toolbar -->
  <div class="flex justify-between items-center p-4 border-b border-surface-300">
    <div class="flex items-center gap-3">
      {#if $hasSelection}
        <span class="text-surface-600 text-sm">{$selectedCount} selected</span>
        <button class="btn btn-sm variant-soft" onclick={clearSelection}>
          Clear Selection
        </button>
        {#if !$readOnly}
          <button
            class="btn btn-sm variant-filled-error"
            onclick={() => (showBatchResetConfirm = true)}
            disabled={batchResetting}
          >
            Reset Selected
          </button>
        {/if}
      {:else}
        <button class="btn btn-sm variant-soft" onclick={selectAll}>
          Select All
        </button>
      {/if}
    </div>
    <div class="flex items-center gap-3">
      {#if !$readOnly}
        <button class="btn btn-sm variant-filled-primary" onclick={() => (showAddModal = true)}>
          + Add Rule
        </button>
      {/if}
    </div>
  </div>

  <!-- Table header -->
  <div class="hidden md:grid grid-cols-[40px_100px_180px_120px_100px_50px] px-4 py-3 bg-surface-50 text-xs font-semibold text-surface-600 uppercase tracking-wider">
    <div></div>
    <div>Port</div>
    <div>Usage</div>
    <div>Progress</div>
    <div>Status</div>
    <div></div>
  </div>

  <!-- Quota items -->
  {#if $loading && $sortedQuotas.length === 0}
    <div class="py-10 text-center text-surface-600">Loading...</div>
  {:else if $sortedQuotas.length === 0}
    <div class="py-10 text-center text-surface-600">No quota rules found</div>
  {:else}
    {#each $sortedQuotas as quota (quota.id)}
      <QuotaItem {quota} />
    {/each}
  {/if}
</div>

<!-- Add quota modal -->
{#if showAddModal}
  <AddQuotaModal onclose={() => (showAddModal = false)} />
{/if}

<!-- Batch reset confirm dialog -->
{#if showBatchResetConfirm}
  <ConfirmDialog
    title="Reset Quotas"
    message={`Are you sure you want to reset ${$selectedCount} quota(s)? This will set their used traffic to 0.`}
    confirmText="Reset"
    onconfirm={handleBatchReset}
    oncancel={() => (showBatchResetConfirm = false)}
  />
{/if}
