<script>
  import Icon from '@iconify/svelte';
  import { addTagTo, removeTagFrom, toggleFavoriteOn } from '../utils/dataSourceUtils.js';
  import TagEditor from './TagEditor.svelte';
  let {
    open = $bindable(false),
    existingNames = [],
    existingTags = [],
    placeholders = {
      confName: 'Conference name, e.g. IPAC25',
      baseUrl: 'https://indico.jacow.org',
      eventId: '123',
      timeout: '60s',
    },
    saving = false,
    apiTokens = [],
    onCreate = (_payload) => {},
    onCancel = () => {},
  } = $props();

  // component state for the new Indico entry
  let newIndico = $state({
    name: '',
    baseUrl: 'https://',
    eventId: 0,
    // store the selected token name (reference), not the raw token value
    apiTokenName: '',
    timeout: '60s',
    // New fields
    favorite: false,
    description: '',
    tags: [],
  });
  // validation errors for the form fields
  let newIndicoErrors = $state({ name: '', baseUrl: '', eventId: '', timeout: '' });

  // initialize when opened
  $effect(() => {
    if (open) initialize();
  });

  function initialize() {
    newIndico.name = '';
    newIndico.baseUrl = placeholders.baseUrl || 'https://';
    newIndico.eventId = parseInt(String(placeholders.eventId || '0'), 10) || 0;
    // If there are apiTokens available, default to the first name; otherwise, use placeholder
    newIndico.apiTokenName = apiTokens && apiTokens.length > 0 ? apiTokens[0].name || '' : '';
    newIndico.timeout = placeholders.timeout || '60s';
    newIndicoErrors = { name: '', baseUrl: '', eventId: '', timeout: '' };
    // initialize new fields
    newIndico.favorite = false;
    newIndico.description = '';
    newIndico.tags = [];
  }

  function validateNewIndico() {
    newIndicoErrors = { name: '', baseUrl: '', eventId: '', timeout: '' };
    let isValid = true;
    // name
    const trimmedName = String(newIndico.name || '').trim();
    if (trimmedName === '') {
      newIndicoErrors.name = 'Name is required';
      isValid = false;
    } else if (
      Array.isArray(existingNames) &&
      existingNames.some((n) => String(n || '').trim() === trimmedName)
    ) {
      // Name conflicts with an existing entry — report an error (no suggestion helper)
      newIndicoErrors.name = 'Name already exists';
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
    if (!/^[0-9]+(ms|s|m|h)$/.test(t)) {
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
      // include new fields
      favorite: newIndico.favorite,
      description: newIndico.description,
      tags: newIndico.tags || [],
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
      <div class="flex items-center justify-between mb-3">
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Add Indico Data Source
        </h4>
        <div>
          <button
            type="button"
            class="p-1 rounded focus:outline-none focus:ring-2 focus:ring-indigo-300"
            onclick={() => toggleFavoriteOn(newIndico)}
            aria-pressed={newIndico.favorite}
            title={newIndico.favorite ? 'Unmark favorite' : 'Mark favorite'}
          >
            <Icon icon={newIndico.favorite ? 'mdi:star' : 'mdi:star-outline'} class="w-8 h-8 text-yellow-500" aria-hidden="true" />
            <span class="sr-only">Toggle favorite</span>
          </button>
        </div>
      </div>
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
            placeholder={placeholders.confName}
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
            placeholder={placeholders.baseUrl}
            list="baseUrl-suggestions"
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
            placeholder={placeholders.timeout}
          />
          {#if newIndicoErrors.timeout}
            <p class="text-xs text-red-500 mt-1">{newIndicoErrors.timeout}</p>
          {/if}
        </div>

        <!-- New fields: Description, Tags (Favorite moved to header) -->
        <div class="pt-2 border-t border-gray-100 dark:border-gray-800">
          <div class="grid grid-cols-1 gap-2">
            <div>
              <label for="new-indico-description" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
              <input
                id="new-indico-description"
                type="text"
                bind:value={newIndico.description}
                placeholder="Optional note about this data source"
                class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            <div>
              <label for="new-indico-tags" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tags</label>
              <TagEditor
                tags={newIndico.tags}
                onAdd={(t) => addTagTo(newIndico, t)}
                onRemove={(idx) => removeTagFrom(newIndico, idx)}
                suggestions={existingTags ? existingTags.filter((t) => !(newIndico.tags || []).includes(t)) : []}
                placeholder="Add tag and press Enter"
              />
            </div>
          </div>
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
