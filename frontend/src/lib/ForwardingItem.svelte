<script>
  import {
    readOnly,
    removeForwardingRule,
    enableForwardingRule,
    disableForwardingRule,
  } from './stores.js';
  import { formatProtocol } from './utils.js';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import EditForwardingModal from './EditForwardingModal.svelte';

  let { rule } = $props();

  let expanded = $state(false);
  let showEditModal = $state(false);
  let showDeleteConfirm = $state(false);
  let processing = $state(false);

  async function handleToggleEnabled() {
    processing = true;
    try {
      if (rule.enabled) {
        await disableForwardingRule(rule.id);
      } else {
        await enableForwardingRule(rule.id);
      }
    } finally {
      processing = false;
    }
  }

  async function handleDelete() {
    processing = true;
    try {
      await removeForwardingRule(rule.id);
    } finally {
      processing = false;
      showDeleteConfirm = false;
    }
  }
</script>

<div class="forwarding-item" class:expanded class:disabled={!rule.enabled} class:unmanaged={!rule.managed}>
  <button class="item-row" onclick={() => expanded = !expanded} type="button">
    <div class="col-status">
      <span
        class="status-indicator"
        class:enabled={rule.enabled}
        class:unmanaged={!rule.managed}
        title={!rule.managed ? 'Unmanaged (external)' : rule.enabled ? 'Enabled' : 'Disabled'}
      ></span>
    </div>
    <div class="col-source">
      <span class="port">{rule.src_port}</span>
      {#if !rule.managed}
        <span class="unmanaged-badge" title="Not managed by nft-ui">ext</span>
      {/if}
    </div>
    <div class="col-dest">
      <span class="dest-ip">{rule.dst_ip}</span>
      <span class="dest-sep">:</span>
      <span class="dest-port">{rule.dst_port}</span>
    </div>
    <div class="col-protocol">
      <span class="protocol-badge">{formatProtocol(rule.protocol)}</span>
    </div>
    <div class="col-actions">
      <span class="expand-icon">{expanded ? '−' : '+'}</span>
    </div>
  </button>

  {#if expanded}
    <div class="item-details">
      {#if rule.comment}
        <div class="detail-row">
          <span class="label">Comment:</span>
          <span class="value">{rule.comment}</span>
        </div>
      {/if}
      <div class="detail-row">
        <span class="label">ID:</span>
        <span class="value code">{rule.id}</span>
      </div>
      <div class="detail-row">
        <span class="label">Status:</span>
        <span class="value">{rule.enabled ? 'Enabled' : 'Disabled'}</span>
      </div>
      <div class="detail-row">
        <span class="label">Managed:</span>
        <span class="value">{rule.managed ? 'Yes (nft-ui)' : 'No (external)'}</span>
      </div>
      {#if rule.limit_mbps > 0}
        <div class="detail-row">
          <span class="label">Bandwidth Limit:</span>
          <span class="value">{rule.limit_mbps} Mbps</span>
        </div>
      {/if}

      {#if !$readOnly && rule.managed}
        <div class="actions">
          <button
            class="btn-secondary"
            onclick={handleToggleEnabled}
            disabled={processing}
          >
            {rule.enabled ? 'Disable' : 'Enable'}
          </button>
          <button
            class="btn-secondary"
            onclick={() => showEditModal = true}
            disabled={processing}
          >
            Edit
          </button>
          <button
            class="btn-danger"
            onclick={() => showDeleteConfirm = true}
            disabled={processing}
          >
            Delete
          </button>
        </div>
      {:else if !$readOnly && !rule.managed}
        {#if rule.enabled}
          <div class="unmanaged-note">
            This rule was created externally and cannot be modified through nft-ui.
          </div>
        {:else}
          <div class="actions">
            <button
              class="btn-secondary"
              onclick={handleToggleEnabled}
              disabled={processing}
            >
              Enable
            </button>
            <button
              class="btn-danger"
              onclick={() => showDeleteConfirm = true}
              disabled={processing}
            >
              Delete
            </button>
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>

{#if showDeleteConfirm}
  <ConfirmDialog
    title="Delete Forwarding Rule"
    message={`Are you sure you want to delete the forwarding rule for port ${rule.src_port}?`}
    confirmText="Delete"
    danger={true}
    onconfirm={handleDelete}
    oncancel={() => showDeleteConfirm = false}
  />
{/if}

{#if showEditModal}
  <EditForwardingModal {rule} onclose={() => showEditModal = false} />
{/if}

<style>
  .forwarding-item {
    border-bottom: 1px solid var(--color-border);
    transition: background-color 0.2s;
  }

  .forwarding-item:last-child {
    border-bottom: none;
  }

  .forwarding-item:hover {
    background-color: var(--color-surface-hover);
  }

  .forwarding-item.disabled {
    opacity: 0.6;
  }

  .item-row {
    display: grid;
    grid-template-columns: 80px 120px 1fr 100px 120px;
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

  .col-status {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .status-indicator {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background-color: var(--color-text-muted);
    transition: background-color 0.2s;
  }

  .status-indicator.enabled {
    background-color: var(--color-success);
  }

  .status-indicator.unmanaged {
    background-color: var(--color-warning);
  }

  .unmanaged-badge {
    font-size: 10px;
    background-color: var(--color-warning);
    color: #000;
    padding: 1px 4px;
    border-radius: 3px;
    margin-left: 6px;
    font-weight: 500;
    text-transform: uppercase;
  }

  .forwarding-item.unmanaged {
    opacity: 0.85;
  }

  .unmanaged-note {
    font-size: 13px;
    color: var(--color-text-muted);
    background-color: var(--color-bg);
    padding: 10px 12px;
    border-radius: 6px;
    margin-top: 12px;
  }

  .col-source .port {
    font-weight: 600;
    font-size: 16px;
    font-family: monospace;
  }

  .col-dest {
    display: flex;
    align-items: center;
    gap: 2px;
    font-family: monospace;
  }

  .dest-ip {
    color: var(--color-primary);
  }

  .dest-sep {
    color: var(--color-text-muted);
  }

  .dest-port {
    font-weight: 600;
  }

  .col-protocol {
    display: flex;
    align-items: center;
  }

  .protocol-badge {
    background-color: var(--color-bg);
    border: 1px solid var(--color-border);
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
  }

  .col-actions {
    text-align: center;
  }

  .expand-icon {
    font-size: 20px;
    color: var(--color-text-muted);
  }

  .item-details {
    padding: 0 16px 16px 96px;
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

  .actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }

  @media (max-width: 768px) {
    .item-row {
      grid-template-columns: 40px 1fr 60px;
      gap: 8px;
    }

    .col-protocol,
    .col-dest {
      display: none;
    }

    .item-details {
      padding-left: 16px;
    }
  }
</style>
