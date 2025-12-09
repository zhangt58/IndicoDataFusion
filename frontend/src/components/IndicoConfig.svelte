<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let open = false; // bindable
  export let initialName = '';
  export let existingNames = [];
  export let placeholders = {
    baseUrl: 'https://indico.example.org',
    eventId: '123',
    apiToken: 'indp_...',
    timeout: '60s'
  };
  export let saving = false;

  let newIndico = {
    name: '',
    baseUrl: 'https://',
    eventId: 0,
    apiToken: '',
    timeout: '60s'
  };
  let newIndicoErrors = {};

  // initialize when opened
  $: if (open) initialize();

  function initialize() {
    newIndico.name = initialName || getUniqueName('Conference Name');
    newIndico.baseUrl = placeholders.baseUrl || 'https://';
    newIndico.eventId = parseInt(String(placeholders.eventId || '0'), 10) || 0;
    newIndico.apiToken = placeholders.apiToken || '';
    newIndico.timeout = placeholders.timeout || '60s';
    newIndicoErrors = {};
  }

  function getUniqueName(base = 'Conference Name') {
    if (!Array.isArray(existingNames) || existingNames.length === 0) return base;
    const existing = new Set(existingNames.map(n => String(n || '')));
    if (!existing.has(base)) return base;
    let i = 2;
    while (existing.has(`${base} (${i})`)) i++;
    return `${base} (${i})`;
  }

  function validateNewIndico() {
    newIndicoErrors = {};
    // name
    if (!newIndico.name || String(newIndico.name).trim() === '') {
      newIndicoErrors.name = 'Name is required';
    }
    // baseUrl
    try {
      const url = new URL(String(newIndico.baseUrl || '').trim());
      if (url.protocol !== 'http:' && url.protocol !== 'https:') {
        newIndicoErrors.baseUrl = 'Base URL must start with http:// or https://';
      }
    } catch (e) {
      newIndicoErrors.baseUrl = 'Base URL is not a valid URL';
    }
    // eventId
    if (newIndico.eventId === null || newIndico.eventId === undefined) {
      newIndicoErrors.eventId = 'Event ID is required';
    } else if (isNaN(Number(newIndico.eventId)) || Number(newIndico.eventId) <= 0) {
      newIndicoErrors.eventId = 'Event ID must be a positive number';
    } else if (!Number.isInteger(Number(newIndico.eventId))) {
      newIndicoErrors.eventId = 'Event ID must be an integer';
    }
    // timeout
    const t = String(newIndico.timeout || '').trim();
    if (!/^\d+(ms|s|m|h)$/.test(t)) {
      newIndicoErrors.timeout = 'Timeout must be a duration like 500ms, 15s, 1m, or 2h';
    }

    return Object.keys(newIndicoErrors).length === 0;
  }

  function onCancel() {
    dispatch('cancel');
  }

  function onSave() {
    if (!validateNewIndico()) return;
    // emit the raw newIndico payload; parent will coerce/validate further
    dispatch('create', newIndico);
  }
</script>

<!-- Modal UI (visible when `open` is true) -->
{#if open}
  <div class="fixed inset-0 z-40 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/40" role="button" tabindex="0" aria-label="Close dialog" on:click={onCancel} on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar') { e.preventDefault(); onCancel(); } }}></div>
    <div role="dialog" aria-modal="true" tabindex="0" on:keydown|stopPropagation={(e) => { if (e.key === 'Escape') onCancel(); }} class="relative z-50 w-full max-w-lg mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
      <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Add Indico Data Source</h4>
      <div class="space-y-3">
        <div>
          <label for="new-indico-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
          <input id="new-indico-name" type="text" bind:value={newIndico.name} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm" />
          {#if newIndicoErrors.name}
            <p class="text-red-500 text-xs mt-1">{newIndicoErrors.name}</p>
          {/if}
        </div>
        <div>
          <label for="new-indico-baseUrl" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL</label>
          <input id="new-indico-baseUrl" type="text" bind:value={newIndico.baseUrl} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                 placeholder={placeholders.baseUrl} />
          {#if newIndicoErrors.baseUrl}
            <p class="text-red-500 text-xs mt-1">{newIndicoErrors.baseUrl}</p>
          {/if}
        </div>
        <div>
          <label for="new-indico-eventId" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label>
          <input id="new-indico-eventId" type="number" bind:value={newIndico.eventId} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                 placeholder={placeholders.eventId} />
          {#if newIndicoErrors.eventId}
            <p class="text-red-500 text-xs mt-1">{newIndicoErrors.eventId}</p>
          {/if}
        </div>
        <div>
          <label for="new-indico-apiToken" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label>
          <input id="new-indico-apiToken" type="text" bind:value={newIndico.apiToken} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono"
                 placeholder={placeholders.apiToken} />
        </div>
        <div>
          <label for="new-indico-timeout" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
          <input id="new-indico-timeout" type="text" bind:value={newIndico.timeout} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                 placeholder={placeholders.timeout} />
          {#if newIndicoErrors.timeout}
            <p class="text-red-500 text-xs mt-1">{newIndicoErrors.timeout}</p>
          {/if}
        </div>
      </div>
      <div class="mt-4 flex justify-end gap-2">
        <button type="button" class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm pointer-events-auto" on:click={onCancel}>Cancel</button>
        <button
          type="button"
          class="px-3 py-1 rounded bg-indigo-600 text-white text-sm pointer-events-auto hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-indigo-400 disabled:hover:bg-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-400 disabled:focus:ring-0 transition-colors"
          on:click={onSave}
          disabled={saving}
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  </div>
{/if}
