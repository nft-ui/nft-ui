import { writable, derived, get } from 'svelte/store';
import { fetchQuotas, addPort, deletePort } from './api.js';

// Core state
export const quotas = writable([]);
export const allowedPorts = writable([]);
export const loading = writable(false);
export const error = writable(null);
export const selectedIds = writable(new Set());
export const readOnly = writable(false);
export const refreshInterval = writable(5);

// Derived stores
export const sortedQuotas = derived(quotas, ($quotas) =>
  [...$quotas].sort((a, b) => {
    // Sort by status (exceeded first), then by usage %
    if (a.status === 'exceeded' && b.status !== 'exceeded') return -1;
    if (b.status === 'exceeded' && a.status !== 'exceeded') return 1;
    return b.usage_percent - a.usage_percent;
  })
);

export const hasSelection = derived(selectedIds, ($ids) => $ids.size > 0);

export const selectedCount = derived(selectedIds, ($ids) => $ids.size);

// Actions
export async function loadQuotas() {
  loading.set(true);
  error.set(null);

  try {
    const data = await fetchQuotas();
    quotas.set(data.quotas || []);
    allowedPorts.set(data.allowed_ports || []);
    readOnly.set(data.read_only);
    refreshInterval.set(data.refresh_interval);
  } catch (e) {
    error.set(e.message);
  } finally {
    loading.set(false);
  }
}

export function toggleSelection(id) {
  selectedIds.update((ids) => {
    const newIds = new Set(ids);
    if (newIds.has(id)) {
      newIds.delete(id);
    } else {
      newIds.add(id);
    }
    return newIds;
  });
}

export function clearSelection() {
  selectedIds.set(new Set());
}

export function selectAll() {
  const allIds = new Set(get(quotas).map((q) => q.id));
  selectedIds.set(allIds);
}

// Notifications
export const notifications = writable([]);
let notificationId = 0;

export function addNotification(message, type = 'info', duration = 3000) {
  const id = ++notificationId;
  notifications.update((n) => [...n, { id, message, type }]);

  if (duration > 0) {
    setTimeout(() => {
      removeNotification(id);
    }, duration);
  }

  return id;
}

export function removeNotification(id) {
  notifications.update((n) => n.filter((item) => item.id !== id));
}

export const success = (msg) => addNotification(msg, 'success');
export const errorNotify = (msg) => addNotification(msg, 'error', 5000);
export const warning = (msg) => addNotification(msg, 'warning');

// Port management actions
export async function addAllowedPort(port) {
  try {
    await addPort(port);
    success('Port added successfully');
    await loadQuotas();
  } catch (e) {
    errorNotify(`Failed to add port: ${e.message}`);
    throw e;
  }
}

export async function removeAllowedPort(handle) {
  try {
    await deletePort(handle);
    success('Port deleted successfully');
    await loadQuotas();
  } catch (e) {
    errorNotify(`Failed to delete port: ${e.message}`);
    throw e;
  }
}
