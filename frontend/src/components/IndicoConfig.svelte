<script>
  let {
    open = $bindable(false),
    initialName = '',
    existingNames = [],
    placeholders = {
      baseUrl: 'https://indico.jacow.org',
      eventId: '123',
      timeout: '60s',
    },
    saving = false,
    apiTokens = [],
    onCreate = (_payload) => {},
    onCancel = () => {},
  } = $props();

  /** @type {{name: string, baseUrl: string, eventId: number, apiTokenName: string, timeout: string}} */
  let newIndico = $state(
    /** @type {any} */ ({
      name: '',
      baseUrl: 'https://',
      eventId: 0,
      // store the selected token name (reference), not the raw token value
      apiTokenName: '',
      timeout: '60s',
    }),
  );
  /** @type {{name?: string, baseUrl?: string, eventId?: string, timeout?: string}} */
  let newIndicoErrors = $state(
    /** @type {any} */ ({
      name: '',
      baseUrl: '',
      eventId: '',
      timeout: '',
    }),
  );

  // initialize when opened
  $effect(() => {
    if (open) initialize();
  });

  function initialize() {
    newIndico.name = initialName || getUniqueName('Conference Name');
    newIndico.baseUrl = placeholders.baseUrl || 'https://';
    newIndico.eventId = parseInt(String(placeholders.eventId || '0'), 10) || 0;
    // If there are apiTokens available, default to the first name; otherwise, use placeholder
    newIndico.apiTokenName =
      apiTokens && apiTokens.length > 0 ? apiTokens[0].name || '' : '';
    newIndico.timeout = placeholders.timeout || '60s';
    newIndicoErrors = { name: '', baseUrl: '', eventId: '', timeout: '' };
  }

  function getUniqueName(base = 'Conference Name') {
    if (!Array.isArray(existingNames) || existingNames.length === 0) return base;
    const existing = new Set(existingNames.map((n) => String(n || '')));
    if (!existing.has(base)) return base;
    let i = 2;
    while (existing.has(`${base} (${i})`)) i++;
    return `${base} (${i})`;
  }

  function validateNewIndico() {
    newIndicoErrors = { name: '', baseUrl: '', eventId: '', timeout: '' };
    let isValid = true;
    // name
    if (!newIndico.name || String(newIndico.name).trim() === '') {
      newIndicoErrors.name = 'Name is required';
      isValid = false;
    }
    // baseUrl
    try {
      const url = new URL(String(newIndico.baseUrl || '').trim());
      if (url.protocol !== 'http:' && url.protocol !== 'https:') {
        newIndicoErrors.baseUrl = 'Base URL must start with http:// or https://';
        isValid = false;
      }
    } catch (e) {
      newIndicoErrors.baseUrl = 'Base URL is not a valid URL';
      isValid = false;
    }
    // eventId
    if (newIndico.eventId === null || newIndico.eventId === undefined) {
      newIndicoErrors.eventId = 'Event ID is required';
      isValid = false;
    } else if (isNaN(Number(newIndico.eventId)) || Number(newIndico.eventId) < 0) {
      newIndicoErrors.eventId = 'Event ID must be a positive number (zero or greater)';
      isValid = false;
    } else if (!Number.isInteger(Number(newIndico.eventId))) {
      newIndicoErrors.eventId = 'Event ID must be an integer';
      isValid = false;
    }
    // timeout
    const t = String(newIndico.timeout || '').trim();
    if (!/^\d+(ms|s|m|h)$/.test(t)) {
      newIndicoErrors.timeout = 'Timeout must be a duration like 500ms, 15s, 1m, or 2h';
      isValid = false;
    }

    return isValid;
  }

  function onCancelClick() {
    onCancel();
  }

  function onSave() {
    if (!validateNewIndico()) return;
    // emit the newIndico payload with apiTokenName (reference) instead of raw token
    const payload = {
      name: newIndico.name,
      baseUrl: newIndico.baseUrl,
      eventId: newIndico.eventId,
      apiTokenName: newIndico.apiTokenName,
      timeout: newIndico.timeout,
    };
    onCreate(payload);
  }
</script>

<!-- Modal UI (visible when `open` is true) -->
{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <div
      class="absolute inset-0 bg-black/40"
      role="button"
      tabindex="0"
      aria-label="Close dialog"
      onclick={onCancelClick}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') {
          e.preventDefault();
          onCancelClick();
        }
      }}
    ></div>
    <div
      class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
    >
      <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
        Add Indico Data Source
      </h4>
      <div class="space-y-2">
        <div>
          <label
            for="new-indico-name"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label
          >
          <input
            id="new-indico-name"
            type="text"
            bind:value={newIndico.name}
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
          />
          {#if newIndicoErrors.name}
            <p class="text-xs text-red-500 mt-1">{newIndicoErrors.name}</p>
          {/if}
        </div>
        <div>
          <label
            for="new-indico-baseUrl"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL</label
          >
          <input
            id="new-indico-baseUrl"
            type="text"
            bind:value={newIndico.baseUrl}
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
            placeholder={placeholders.baseUrl} list="baseUrl-suggestions"
          />
          <datalist id="baseUrl-suggestions">
            <option value="https://indico.jacow.org">https://indico.jacow.org</option>
            <option value="https://indico.global">https://indico.global</option>
          </datalist>
          {#if newIndicoErrors.baseUrl}
            <p class="text-xs text-red-500 mt-1">{newIndicoErrors.baseUrl}</p>
          {/if}
        </div>
        <div>
          <label
            for="new-indico-eventId"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label
          >
          <input
            id="new-indico-eventId"
            type="number"
            bind:value={newIndico.eventId}
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
          />
          {#if newIndicoErrors.eventId}
            <p class="text-xs text-red-500 mt-1">{newIndicoErrors.eventId}</p>
          {/if}
        </div>
        <div>
          <label
            for="new-indico-apiTokenName"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label
          >
          {#if apiTokens && apiTokens.length > 0}
            <select
              id="new-indico-apiTokenName"
              bind:value={newIndico.apiTokenName}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono"
            >
              {#each apiTokens as t}
                <option value={t.name}>{t.name}{t.username ? ` — ${t.username}` : ''}</option>
              {/each}
            </select>
          {:else}
            <input
              id="new-indico-apiTokenName"
              type="text"
              bind:value={newIndico.apiTokenName}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono"
            />
          {/if}
        </div>
        <div>
          <label
            for="new-indico-timeout"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label
          >
          <input
            id="new-indico-timeout"
            type="text"
            bind:value={newIndico.timeout}
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
          />
          {#if newIndicoErrors.timeout}
            <p class="text-xs text-red-500 mt-1">{newIndicoErrors.timeout}</p>
          {/if}
        </div>
      </div>
      <div class="mt-4 flex justify-end gap-2">
        <button
          class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm"
          onclick={onCancelClick}>Cancel</button
        >
        <button
          class="px-3 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50"
          onclick={onSave}
          disabled={saving}
        >
          {saving ? 'Saving...' : 'Add'}
        </button>
      </div>
    </div>
  </div>
{/if}
