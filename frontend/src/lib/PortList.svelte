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

<section class="mt-8">
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold m-0" style="color: var(--text);">Allowed Inbound Ports</h2>
    {#if !$readOnly}
      <button class="btn btn-sm btn-primary" onclick={handleAddClick}>
        + Add Port
      </button>
    {/if}
  </div>

  {#if $allowedPorts.length === 0}
    <p class="text-center py-5" style="color: var(--text-muted);">No allowed port rules found</p>
  {:else}
    <div class="flex flex-wrap gap-2">
      {#each sortedPorts as port}
        <div
          class="badge flex items-center gap-2 relative transition-all"
          class:badge-primary={port.managed}
          style="padding: 0.5rem 0.75rem;"
        >
          <span class="font-mono font-semibold" style="color: {port.managed ? 'var(--primary)' : 'var(--text)'};">{port.port}</span>
          {#if port.comment}
            <span class="text-xs" style="color: var(--text-muted);">{port.comment}</span>
          {/if}
          {#if port.managed && !$readOnly}
            <button
              class="bg-transparent border-none text-lg p-0 px-1 cursor-pointer leading-none ml-1 transition-opacity"
              style="color: var(--danger); opacity: 0.6;"
              onmouseover={(e) => e.currentTarget.style.opacity = '1'}
              onmouseout={(e) => e.currentTarget.style.opacity = '0.6'}
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
