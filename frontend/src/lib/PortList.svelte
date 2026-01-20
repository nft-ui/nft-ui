<script>
  import { allowedPorts, readOnly, removeAllowedPort } from './stores.js';
  import AddPortModal from './AddPortModal.svelte';
  import ConfirmDialog from './ConfirmDialog.svelte';

  let showAddModal = $state(false);
  let portToDelete = $state(null);

  let sortedPorts = $derived(
    [...$allowedPorts].sort((a, b) => a.port - b.port)
  );

  function handleAddClick() {
    showAddModal = true;
  }

  function handleDeleteClick(port) {
    portToDelete = port;
  }

  async function confirmDelete() {
    if (portToDelete) {
      try {
        await removeAllowedPort(portToDelete.handle);
      } catch (e) {
        // Error already shown by store
      }
    }
    portToDelete = null;
  }

  function cancelDelete() {
    portToDelete = null;
  }
</script>

<section class="port-list">
  <div class="port-header">
    <h2>Allowed Inbound Ports</h2>
    {#if !$readOnly}
      <button class="btn-primary btn-sm" onclick={handleAddClick}>
        + Add Port
      </button>
    {/if}
  </div>

  {#if $allowedPorts.length === 0}
    <p class="empty-state">No allowed port rules found</p>
  {:else}
    <div class="port-grid">
      {#each sortedPorts as port}
        <div class="port-item" class:managed={port.managed}>
          <span class="port-number">{port.port}</span>
          {#if port.comment}
            <span class="port-comment">{port.comment}</span>
          {/if}
          {#if port.managed && !$readOnly}
            <button
              class="delete-btn"
              onclick={() => handleDeleteClick(port)}
              title="Delete port"
            >
              &times;
            </button>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</section>

{#if showAddModal}
  <AddPortModal onclose={() => showAddModal = false} />
{/if}

{#if portToDelete}
  <ConfirmDialog
    title="Delete Port"
    message={`Are you sure you want to delete port ${portToDelete.port}?`}
    confirmText="Delete"
    danger={true}
    onconfirm={confirmDelete}
    oncancel={cancelDelete}
  />
{/if}

<style>
  .port-list {
    margin-top: 32px;
  }

  .port-header {
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

  .empty-state {
    color: var(--color-text-muted);
    text-align: center;
    padding: 20px;
  }

  .port-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .port-item {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 6px;
    padding: 8px 12px;
    position: relative;
  }

  .port-item.managed {
    border-color: var(--color-primary);
  }

  .port-number {
    font-family: monospace;
    font-weight: 600;
    color: var(--color-primary);
  }

  .port-comment {
    font-size: 12px;
    color: var(--color-text-muted);
  }

  .delete-btn {
    background: transparent;
    border: none;
    color: var(--color-danger);
    font-size: 18px;
    padding: 0 4px;
    cursor: pointer;
    opacity: 0.6;
    line-height: 1;
    margin-left: 4px;
  }

  .delete-btn:hover {
    opacity: 1;
  }
</style>
