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

<div class="quota-item" class:expanded class:selected={isSelected}>
  <button class="quota-row" onclick={toggleExpand} type="button">
    <div class="col-checkbox">
      <input
        type="checkbox"
        checked={isSelected}
        onclick={handleCheckbox}
      />
    </div>
    <div class="col-port">
      <span class="inbound-indicator" class:has-inbound={hasInbound} title={hasInbound ? 'Inbound allowed' : 'No inbound rule'}></span>
      <span class="port">{quota.port}</span>
    </div>
    <div class="col-usage">
      <span class="used">{formatBytes(quota.used_bytes)}</span>
      <span class="separator">/</span>
      <span class="quota">{formatBytes(quota.quota_bytes)}</span>
    </div>
    <div class="col-progress">
      <div class="ring-wrapper">
        <svg class="progress-ring" viewBox="0 0 36 36">
          <circle class="ring-bg" cx="18" cy="18" r="15.5" />
          <circle
            class="ring-progress"
            cx="18" cy="18" r="15.5"
            stroke={progressColor}
            stroke-dasharray="{ringPercent}, 100"
          />
        </svg>
        <span class="ring-text">{formatPercent(quota.usage_percent)}</span>
      </div>
    </div>
    <div class="col-status">
      <span class="status-dot" style="background-color: {statusColor};"></span>
      <span class="status-text">{quota.status}</span>
    </div>
    <div class="col-actions">
      <span class="expand-icon">{expanded ? '−' : '+'}</span>
    </div>
  </button>

  {#if expanded}
    <div class="quota-details">
      {#if quota.comment}
        <div class="detail-row">
          <span class="label">Comment:</span>
          <span class="value">{quota.comment}</span>
        </div>
      {/if}
      <div class="detail-row">
        <span class="label">ID:</span>
        <span class="value code">{quota.id}</span>
      </div>
      {#if quota.token}
        <div class="detail-row">
          <span class="label">Query Token:</span>
          <span class="value code token-value">
            {quota.token}
            <button class="copy-btn" onclick={copyToken} title="Copy token">
              {copiedToken ? 'Copied!' : 'Copy'}
            </button>
          </span>
        </div>
        <div class="detail-row">
          <span class="label">Query URL:</span>
          <span class="value code url-value">
            <a href={queryUrl} target="_blank">/query?token={quota.token}</a>
            <button class="copy-btn" onclick={copyQueryUrl} title="Copy URL">
              {copiedUrl ? 'Copied!' : 'Copy'}
            </button>
          </span>
        </div>
      {/if}

      {#if !$readOnly}
        <div class="actions">
          <button
            class="btn-secondary"
            onclick={() => (showResetConfirm = true)}
            disabled={processing}
          >
            Reset
          </button>
          <button
            class="btn-secondary"
            onclick={() => (showEditModal = true)}
            disabled={processing}
          >
            Edit
          </button>
          <button
            class="btn-danger"
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
  .quota-item {
    border-bottom: 1px solid var(--color-border);
    transition: background-color 0.2s;
  }

  .quota-item:last-child {
    border-bottom: none;
  }

  .quota-item:hover {
    background-color: var(--color-surface-hover);
  }

  .quota-item.selected {
    background-color: rgba(74, 158, 255, 0.1);
  }

  .quota-row {
    display: grid;
    grid-template-columns: 40px 100px 180px 120px 100px 50px;
    padding: 12px 16px;
    align-items: center;
    cursor: pointer;
    width: 100%;
    background: transparent;
    border: none;
    color: inherit;
    font: inherit;
    text-align: left;
  }

  .col-checkbox input {
    cursor: pointer;
    width: 18px;
    height: 18px;
  }

  .col-port {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .inbound-indicator {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: var(--color-warning);
    flex-shrink: 0;
  }

  .inbound-indicator.has-inbound {
    background-color: var(--color-success);
  }

  .port {
    font-weight: 600;
    font-size: 16px;
  }

  .col-usage {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 14px;
  }

  .col-usage .used {
    font-weight: 600;
  }

  .col-usage .separator {
    color: var(--color-text-muted);
  }

  .col-usage .quota {
    color: var(--color-text-muted);
  }

  .col-progress {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .ring-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .progress-ring {
    width: 44px;
    height: 44px;
    transform: rotate(-90deg);
  }

  .ring-bg {
    fill: none;
    stroke: var(--color-bg);
    stroke-width: 3;
  }

  .ring-progress {
    fill: none;
    stroke-width: 3;
    stroke-linecap: round;
    transition: stroke-dasharray 0.3s ease;
  }

  .ring-text {
    font-size: 12px;
    font-weight: 600;
    min-width: 40px;
  }

  .col-status {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .status-text {
    font-size: 14px;
    text-transform: capitalize;
  }

  .expand-icon {
    font-size: 20px;
    color: var(--color-text-muted);
    text-align: center;
    width: 100%;
    display: block;
  }

  .quota-details {
    padding: 0 16px 16px 56px;
    animation: slideDown 0.2s ease;
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

  .detail-row {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
    font-size: 14px;
  }

  .label {
    color: var(--color-text-muted);
  }

  .code {
    font-family: monospace;
    font-size: 12px;
    background-color: var(--color-bg);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .token-value,
  .url-value {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .url-value a {
    color: var(--color-primary);
    text-decoration: none;
  }

  .url-value a:hover {
    text-decoration: underline;
  }

  .copy-btn {
    padding: 2px 8px;
    font-size: 11px;
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    cursor: pointer;
    color: var(--color-text-muted);
    transition: all 0.2s;
  }

  .copy-btn:hover {
    background-color: var(--color-surface-hover);
    color: var(--color-text);
  }

  .actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }

  @media (max-width: 768px) {
    .quota-row {
      grid-template-columns: 40px 80px 1fr 60px;
      gap: 8px;
    }

    .col-usage,
    .col-status {
      display: none;
    }

    .quota-details {
      padding-left: 16px;
    }
  }
</style>
