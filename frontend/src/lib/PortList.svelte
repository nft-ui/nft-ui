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
    <h2 class="text-lg font-semibold m-0">Allowed Inbound Ports</h2>
    {#if !$readOnly}
      <button class="btn btn-sm variant-filled-primary" onclick={handleAddClick}>
        + Add Port
      </button>
    {/if}
  </div>

  {#if $allowedPorts.length === 0}
    <p class="text-surface-600 text-center py-5">No allowed port rules found</p>
  {:else}
    <div class="flex flex-wrap gap-2">
      {#each sortedPorts as port}
        <div
          class="flex items-center gap-2 card p-2 px-3 relative"
          class:border-primary-500={port.managed}
          class:border={port.managed}
        >
          <span class="font-mono font-semibold text-primary-500">{port.port}</span>
          {#if port.comment}
            <span class="text-xs text-surface-600">{port.comment}</span>
          {/if}
          {#if port.managed && !$readOnly}
            <button
              class="bg-transparent border-none text-error-500 text-lg p-0 px-1 cursor-pointer opacity-60 hover:opacity-100 leading-none ml-1"
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
