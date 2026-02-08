<script>
  import { selectedIds, toggleSelection, readOnly, loadQuotas, success, errorNotify, allowedPorts } from './stores.js';
  import { resetQuota, deleteQuota } from './api.js';
  import { formatBytes, formatPercent, getProgressColor, getStatusColor } from './utils.js';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import EditQuotaModal from './EditQuotaModal.svelte';

  let { quota } = $props();

  let expanded = $state(false);
  let showResetConfirm = $state(false);
  let showDeleteConfirm = $state(false);
  let showEditModal = $state(false);
  let processing = $state(false);
  let copiedToken = $state(false);
  let copiedUrl = $state(false);

  let isSelected = $derived($selectedIds.has(quota.id));
  let progressColor = $derived(getProgressColor(quota.usage_percent));
  let statusColor = $derived(getStatusColor(quota.status));
  let hasInbound = $derived($allowedPorts.some(p => p.port === quota.port));
  let ringPercent = $derived(Math.min(quota.usage_percent, 100));
  let queryUrl = $derived(quota.token ? `${window.location.origin}/query?token=${quota.token}` : '');

  function handleCheckbox(e) {
    e.stopPropagation();
    toggleSelection(quota.id);
  }

  function toggleExpand() {
    expanded = !expanded;
  }

  async function handleReset() {
    processing = true;
    try {
      await resetQuota(quota.id);
      success('Quota reset successfully');
      await loadQuotas();
    } catch (e) {
      errorNotify(`Failed to reset quota: ${e.message}`);
    } finally {
      processing = false;
      showResetConfirm = false;
    }
  }

  async function handleDelete() {
    processing = true;
    try {
      await deleteQuota(quota.id);
      success('Quota deleted successfully');
      await loadQuotas();
    } catch (e) {
      errorNotify(`Failed to delete quota: ${e.message}`);
    } finally {
      processing = false;
      showDeleteConfirm = false;
    }
  }

  function copyToken() {
    if (quota.token) {
      navigator.clipboard.writeText(quota.token);
      copiedToken = true;
      setTimeout(() => copiedToken = false, 2000);
    }
  }

  function copyQueryUrl() {
    if (queryUrl) {
      navigator.clipboard.writeText(queryUrl);
      copiedUrl = true;
      setTimeout(() => copiedUrl = false, 2000);
    }
  }
</script>

<div
  class="border-b border-surface-300 last:border-b-0 transition-colors hover:bg-surface-50"
  class:selected={isSelected}
>
  <button
    class="grid md:grid-cols-[40px_100px_180px_120px_100px_50px] grid-cols-[40px_80px_1fr_60px] gap-2 md:gap-0 p-3 md:px-4 items-center cursor-pointer w-full bg-transparent border-none text-inherit text-left"
    onclick={toggleExpand}
    type="button"
  >
    <div>
      <input
        type="checkbox"
        class="w-[18px] h-[18px] cursor-pointer"
        checked={isSelected}
        onclick={handleCheckbox}
      />
    </div>
    <div class="flex items-center gap-2">
      <span
        class="w-2.5 h-2.5 rounded-full flex-shrink-0"
        class:bg-success-500={hasInbound}
        class:bg-warning-500={!hasInbound}
        title={hasInbound ? 'Inbound allowed' : 'No inbound rule'}
      ></span>
      <span class="font-semibold text-base">{quota.port}</span>
    </div>
    <div class="hidden md:flex items-center gap-1 text-sm">
      <span class="font-semibold">{formatBytes(quota.used_bytes)}</span>
      <span class="text-surface-600">/</span>
      <span class="text-surface-600">{formatBytes(quota.quota_bytes)}</span>
    </div>
    <div class="flex items-center justify-center">
      <div class="relative flex items-center gap-2">
        <svg class="w-11 h-11 -rotate-90" viewBox="0 0 36 36">
          <circle class="fill-none stroke-surface-200 stroke-[3]" cx="18" cy="18" r="15.5" />
          <circle
            class="fill-none stroke-[3] stroke-round transition-all duration-300"
            cx="18"
            cy="18"
            r="15.5"
            style="stroke: {progressColor}; stroke-dasharray: {ringPercent}, 100;"
          />
        </svg>
        <span class="text-xs font-semibold min-w-[40px]">{formatPercent(quota.usage_percent)}</span>
      </div>
    </div>
    <div class="hidden md:flex items-center gap-2">
      <span class="w-2 h-2 rounded-full" style="background-color: {statusColor};"></span>
      <span class="text-sm capitalize">{quota.status}</span>
    </div>
    <div>
      <span class="text-xl text-surface-600 text-center w-full block">{expanded ? '−' : '+'}</span>
    </div>
  </button>

  {#if expanded}
    <div class="px-4 md:pl-14 pb-4 animate-[slideDown_0.2s_ease]">
      {#if quota.comment}
        <div class="flex gap-2 mb-2 text-sm">
          <span class="text-surface-600">Comment:</span>
          <span>{quota.comment}</span>
        </div>
      {/if}
      <div class="flex gap-2 mb-2 text-sm">
        <span class="text-surface-600">ID:</span>
        <span class="font-mono text-xs bg-surface-50 px-1.5 py-0.5 rounded">{quota.id}</span>
      </div>
      {#if quota.token}
        <div class="flex gap-2 mb-2 text-sm">
          <span class="text-surface-600">Query Token:</span>
          <span class="inline-flex items-center gap-2 font-mono text-xs bg-surface-50 px-1.5 py-0.5 rounded">
            {quota.token}
            <button
              class="px-2 py-0.5 text-[11px] bg-surface-100 border border-surface-300 rounded cursor-pointer text-surface-600 hover:bg-surface-200 hover:text-surface-900 transition-all"
              onclick={copyToken}
              title="Copy token"
            >
              {copiedToken ? 'Copied!' : 'Copy'}
            </button>
          </span>
        </div>
        <div class="flex gap-2 mb-2 text-sm">
          <span class="text-surface-600">Query URL:</span>
          <span class="inline-flex items-center gap-2 font-mono text-xs bg-surface-50 px-1.5 py-0.5 rounded">
            <a href={queryUrl} target="_blank" class="text-primary-500 no-underline hover:underline">
              /query?token={quota.token}
            </a>
            <button
              class="px-2 py-0.5 text-[11px] bg-surface-100 border border-surface-300 rounded cursor-pointer text-surface-600 hover:bg-surface-200 hover:text-surface-900 transition-all"
              onclick={copyQueryUrl}
              title="Copy URL"
            >
              {copiedUrl ? 'Copied!' : 'Copy'}
            </button>
          </span>
        </div>
      {/if}

      {#if !$readOnly}
        <div class="flex gap-2 mt-4">
          <button
            class="btn btn-sm variant-soft"
            onclick={() => (showResetConfirm = true)}
            disabled={processing}
          >
            Reset
          </button>
          <button
            class="btn btn-sm variant-soft"
            onclick={() => (showEditModal = true)}
            disabled={processing}
          >
            Edit
          </button>
          <button
            class="btn btn-sm variant-filled-error"
            onclick={() => (showDeleteConfirm = true)}
            disabled={processing}
          >
            Delete
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Reset confirm dialog -->
{#if showResetConfirm}
  <ConfirmDialog
    title="Reset Quota"
    message={`Are you sure you want to reset the quota for port ${quota.port}? This will set the used traffic to 0.`}
    confirmText="Reset"
    onconfirm={handleReset}
    oncancel={() => (showResetConfirm = false)}
  />
{/if}

<!-- Delete confirm dialog -->
{#if showDeleteConfirm}
  <ConfirmDialog
    title="Delete Quota"
    message={`Are you sure you want to delete the quota rule for port ${quota.port}? This action cannot be undone.`}
    confirmText="Delete"
    danger={true}
    onconfirm={handleDelete}
    oncancel={() => (showDeleteConfirm = false)}
  />
{/if}

<!-- Edit modal -->
{#if showEditModal}
  <EditQuotaModal {quota} onclose={() => (showEditModal = false)} />
{/if}

<style>
  .selected {
    background-color: rgba(74, 158, 255, 0.1);
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
