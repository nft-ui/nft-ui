<script>
  import { onMount } from 'svelte';
  import {
    loadQuotas,
    loading,
    error,
    readOnly,
    refreshInterval,
    notifications,
    removeNotification,
  } from './lib/stores.js';
  import QuotaList from './lib/QuotaList.svelte';
  import PortList from './lib/PortList.svelte';
  import Toast from './lib/Toast.svelte';

  let refreshTimer = $state(null);

  onMount(() => {
    loadQuotas();
    startAutoRefresh();
    return () => stopAutoRefresh();
  });

  function startAutoRefresh() {
    stopAutoRefresh();
    const interval = $refreshInterval;
    if (interval > 0) {
      refreshTimer = setInterval(() => {
        loadQuotas();
      }, interval * 1000);
    }
  }

  function stopAutoRefresh() {
    if (refreshTimer) {
      clearInterval(refreshTimer);
      refreshTimer = null;
    }
  }

  function handleRefresh() {
    loadQuotas();
  }
</script>

<div class="app">
  <header>
    <div class="container header-content">
      <h1>nft-ui</h1>
      <div class="header-right">
        {#if $readOnly}
          <span class="badge badge-warning">Read Only</span>
        {/if}
        <button class="btn-secondary" onclick={handleRefresh} disabled={$loading}>
          {$loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>
    </div>
  </header>

  <main class="container">
    {#if $error}
      <div class="error-banner">
        <span>Error: {$error}</span>
        <button onclick={handleRefresh}>Retry</button>
      </div>
    {/if}

    <QuotaList />
    <PortList />
  </main>

  <!-- Toast notifications -->
  <div class="toast-container">
    {#each $notifications as notification (notification.id)}
      <Toast
        message={notification.message}
        type={notification.type}
        onclose={() => removeNotification(notification.id)}
      />
    {/each}
  </div>
</div>

<style>
  .app {
    min-height: 100vh;
  }

  header {
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
    padding: 16px 0;
  }

  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  h1 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .badge {
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
  }

  .badge-warning {
    background-color: var(--color-warning);
    color: #000;
  }

  main {
    padding-top: 24px;
    padding-bottom: 24px;
  }

  .error-banner {
    background-color: rgba(248, 113, 113, 0.2);
    border: 1px solid var(--color-danger);
    border-radius: 8px;
    padding: 12px 16px;
    margin-bottom: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .error-banner button {
    background-color: var(--color-danger);
    color: white;
    padding: 6px 12px;
  }

  .toast-container {
    position: fixed;
    bottom: 20px;
    right: 20px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 2000;
  }
</style>
