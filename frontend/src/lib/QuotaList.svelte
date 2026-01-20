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

<div class="quota-list">
  <!-- Toolbar -->
  <div class="toolbar">
    <div class="toolbar-left">
      {#if $hasSelection}
        <span class="selection-info">{$selectedCount} selected</span>
        <button class="btn-secondary" onclick={clearSelection}>
          Clear Selection
        </button>
        {#if !$readOnly}
          <button
            class="btn-danger"
            onclick={() => (showBatchResetConfirm = true)}
            disabled={batchResetting}
          >
            Reset Selected
          </button>
        {/if}
      {:else}
        <button class="btn-secondary" onclick={selectAll}>
          Select All
        </button>
      {/if}
    </div>
    <div class="toolbar-right">
      {#if !$readOnly}
        <button class="btn-primary" onclick={() => (showAddModal = true)}>
          + Add Rule
        </button>
      {/if}
    </div>
  </div>

  <!-- Table header -->
  <div class="table-header">
    <div class="col-checkbox"></div>
    <div class="col-port">Port</div>
    <div class="col-usage">Usage</div>
    <div class="col-progress">Progress</div>
    <div class="col-status">Status</div>
    <div class="col-actions"></div>
  </div>

  <!-- Quota items -->
  {#if $loading && $sortedQuotas.length === 0}
    <div class="loading">Loading...</div>
  {:else if $sortedQuotas.length === 0}
    <div class="empty">No quota rules found</div>
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

<style>
  .quota-list {
    background-color: var(--color-surface);
    border-radius: 12px;
    overflow: hidden;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--color-border);
  }

  .toolbar-left,
  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .selection-info {
    color: var(--color-text-muted);
    font-size: 14px;
  }

  .table-header {
    display: grid;
    grid-template-columns: 40px 100px 180px 120px 100px 50px;
    padding: 12px 16px;
    background-color: var(--color-bg);
    font-size: 12px;
    font-weight: 600;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .loading,
  .empty {
    padding: 40px;
    text-align: center;
    color: var(--color-text-muted);
  }

  @media (max-width: 768px) {
    .table-header {
      display: none;
    }
  }
</style>
