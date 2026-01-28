<script>
  import { sortedForwardingRules, forwardingLoading, readOnly } from './stores.js';
  import ForwardingItem from './ForwardingItem.svelte';
  import AddForwardingModal from './AddForwardingModal.svelte';

  let showAddModal = $state(false);
</script>

<section class="forwarding-list">
  <div class="section-header">
    <h2>Port Forwarding</h2>
    {#if !$readOnly}
      <button class="btn-primary btn-sm" onclick={() => showAddModal = true}>
        + Add Rule
      </button>
    {/if}
  </div>

  {#if $forwardingLoading}
    <div class="loading-state">Loading forwarding rules...</div>
  {:else if $sortedForwardingRules.length === 0}
    <div class="empty-state">
      <p>No forwarding rules configured</p>
      {#if !$readOnly}
        <p class="hint">Click "Add Rule" to create a new port forwarding rule</p>
      {/if}
    </div>
  {:else}
    <div class="table-container">
      <div class="table-header">
        <div class="col-status">Status</div>
        <div class="col-source">Source Port</div>
        <div class="col-dest">Destination</div>
        <div class="col-protocol">Protocol</div>
        <div class="col-actions">Actions</div>
      </div>
      {#each $sortedForwardingRules as rule (rule.id)}
        <ForwardingItem {rule} />
      {/each}
    </div>
  {/if}
</section>

{#if showAddModal}
  <AddForwardingModal onclose={() => showAddModal = false} />
{/if}

<style>
  .forwarding-list {
    margin-top: 32px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    color: var(--color-text);
  }

  .btn-sm {
    padding: 6px 12px;
    font-size: 13px;
  }

  .loading-state,
  .empty-state {
    text-align: center;
    padding: 32px 20px;
    color: var(--color-text-muted);
  }

  .empty-state .hint {
    font-size: 14px;
    margin-top: 8px;
  }

  .table-container {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 8px;
    overflow: hidden;
  }

  .table-header {
    display: grid;
    grid-template-columns: 80px 120px 1fr 100px 120px;
    padding: 12px 16px;
    background-color: var(--color-bg);
    border-bottom: 1px solid var(--color-border);
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--color-text-muted);
    letter-spacing: 0.05em;
  }

  @media (max-width: 768px) {
    .table-header {
      display: none;
    }
  }
</style>
