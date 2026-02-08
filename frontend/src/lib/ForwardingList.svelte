<script>
  import { sortedForwardingRules, forwardingLoading, readOnly } from './stores.js';
  import ForwardingItem from './ForwardingItem.svelte';
  import AddForwardingModal from './AddForwardingModal.svelte';

  let showAddModal = $state(false);
</script>

<section class="mt-8">
  <div class="flex justify-between items-center mb-4">
    <h2 class="text-lg font-semibold m-0">Port Forwarding</h2>
    {#if !$readOnly}
      <button class="btn btn-sm variant-filled-primary" onclick={() => showAddModal = true}>
        + Add Rule
      </button>
    {/if}
  </div>

  {#if $forwardingLoading}
    <div class="text-center py-8 text-surface-600">Loading forwarding rules...</div>
  {:else if $sortedForwardingRules.length === 0}
    <div class="text-center py-8 text-surface-600">
      <p>No forwarding rules configured</p>
      {#if !$readOnly}
        <p class="text-sm mt-2">Click "Add Rule" to create a new port forwarding rule</p>
      {/if}
    </div>
  {:else}
    <div class="card overflow-hidden bg-surface-100">
      <div class="hidden md:grid grid-cols-[80px_120px_1fr_100px_120px] px-4 py-3 bg-surface-50 border-b border-surface-300 text-xs font-semibold uppercase text-surface-600 tracking-wider">
        <div>Status</div>
        <div>Source Port</div>
        <div>Destination</div>
        <div>Protocol</div>
        <div>Actions</div>
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
