<script>
  import { onMount } from 'svelte';
  import {
    loadQuotas,
    loadForwardingRules,
    loading,
    error,
    readOnly,
    refreshInterval,
    notifications,
    removeNotification,
    isEditingModal,
  } from './lib/stores.js';
  import QuotaList from './lib/QuotaList.svelte';
  import PortList from './lib/PortList.svelte';
  import ForwardingList from './lib/ForwardingList.svelte';
  import RawRuleset from './lib/RawRuleset.svelte';
  import Toast from './lib/Toast.svelte';
  import PublicQuery from './lib/PublicQuery.svelte';

  // Detect route immediately at script initialization
  function getInitialRoute() {
    if (typeof window !== 'undefined') {
      const path = window.location.pathname;
      if (path === '/query' || path.startsWith('/query')) {
        return 'query';
      }
    }
    return 'admin';
  }

  let refreshTimer = $state(null);
  let currentRoute = $state(getInitialRoute());

  // Theme management
  const THEMES = ['modern', 'wintry', 'rocket', 'seafoam', 'crimson'];
  let currentTheme = $state('modern');

  onMount(() => {
    // Load theme from localStorage
    const savedTheme = localStorage.getItem('theme') || 'modern';
    currentTheme = savedTheme;
    applyTheme(savedTheme);

    // Only load data for admin route
    if (currentRoute === 'admin') {
      loadQuotas();
      loadForwardingRules();
      startAutoRefresh();
    }

    return () => stopAutoRefresh();
  });

  function applyTheme(theme) {
    if (typeof document !== 'undefined') {
      document.body.dataset.theme = theme;
    }
  }

  function handleThemeChange(event) {
    const newTheme = event.target.value;
    currentTheme = newTheme;
    applyTheme(newTheme);
    localStorage.setItem('theme', newTheme);
  }

  function startAutoRefresh() {
    stopAutoRefresh();
    const interval = $refreshInterval;
    if (interval > 0) {
      refreshTimer = setInterval(() => {
        // Skip refresh if user is editing in a modal
        if ($isEditingModal) return;
        loadQuotas();
        loadForwardingRules();
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
    loadForwardingRules();
  }
</script>

{#if currentRoute === 'query'}
  <PublicQuery />
{:else}
  <div class="min-h-screen bg-surface-50">
    <header class="bg-surface-100 border-b border-surface-300">
      <div class="container mx-auto px-5 py-4">
        <div class="flex justify-between items-center">
          <h1 class="text-2xl font-semibold">nft-ui</h1>
          <div class="flex items-center gap-3">
            <select
              class="select text-sm px-3 py-2"
              bind:value={currentTheme}
              onchange={handleThemeChange}
            >
              {#each THEMES as theme}
                <option value={theme}>{theme.charAt(0).toUpperCase() + theme.slice(1)}</option>
              {/each}
            </select>
            {#if $readOnly}
              <span class="badge variant-filled-warning text-xs px-3 py-1">Read Only</span>
            {/if}
            <button
              class="btn variant-soft text-sm"
              onclick={handleRefresh}
              disabled={$loading}
            >
              {$loading ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="container mx-auto px-5 py-6">
      {#if $error}
        <div class="alert variant-filled-error mb-5">
          <div class="alert-message">
            <span>Error: {$error}</span>
          </div>
          <div class="alert-actions">
            <button class="btn btn-sm variant-filled" onclick={handleRefresh}>Retry</button>
          </div>
        </div>
      {/if}

      <QuotaList />
      <PortList />
      <ForwardingList />
      <RawRuleset />
    </main>

    <!-- Toast notifications -->
    <div class="fixed bottom-5 right-5 flex flex-col gap-2 z-[2000]">
      {#each $notifications as notification (notification.id)}
        <Toast
          message={notification.message}
          type={notification.type}
          onclose={() => removeNotification(notification.id)}
        />
      {/each}
    </div>
  </div>
{/if}
