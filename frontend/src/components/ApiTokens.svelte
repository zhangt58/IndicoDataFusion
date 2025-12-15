<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let apiTokens = [];
  export let disabled = false;

  let showModal = false;
  let editingIndex = -1;
  let token = { name: '', baseUrl: '', username: '', token: '' };

  function openAdd() {
    editingIndex = -1;
    token = { name: '', baseUrl: '', username: '', token: '' };
    showModal = true;
  }

  function openEdit(i) {
    editingIndex = i;
    token = Object.assign({}, apiTokens[i]);
    showModal = true;
  }

  function onCancel() {
    showModal = false;
  }

  function onSave() {
    // basic validation
    if (!token.name || token.name.trim() === '') return;
    if (editingIndex >= 0) {
      dispatch('edit', { index: editingIndex, entry: token });
    } else {
      dispatch('add', token);
    }
    showModal = false;
  }

  function onDelete(i) {
    if (!confirm(`Delete API token "${apiTokens[i].name}"?`)) return;
    dispatch('delete', i);
  }
</script>

<div class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
  <div class="flex items-center justify-between">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">API Tokens</h3>
    <div>
      <button class="px-3 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50" on:click={openAdd} disabled={disabled}>Add Token</button>
    </div>
  </div>
  {#if apiTokens && apiTokens.length > 0}
    <ul class="space-y-2 mt-2 text-sm text-gray-800 dark:text-gray-200">
      {#each apiTokens as t, i}
        <li class="flex items-center justify-between p-2 rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-700">
          <div>
            <div class="font-medium">{t.name} {t.username ? ` — ${t.username}` : ''}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{t.baseUrl}</div>
          </div>
          <div class="flex items-center gap-2">
            <button class="text-sm px-2 py-1 rounded bg-gray-200 dark:bg-gray-600" on:click={() => openEdit(i)} disabled={disabled}>Edit</button>
            <button class="text-sm px-2 py-1 rounded bg-red-600 text-white" on:click={() => onDelete(i)} disabled={disabled}>Delete</button>
          </div>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="text-sm text-gray-600 dark:text-gray-400 mt-2">No API tokens configured. Add one to reference by name from data sources.</p>
  {/if}

  {#if showModal}
    <div class="fixed inset-0 z-40 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/40" role="button" tabindex="0" aria-label="Close dialog" on:click={onCancel} on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') { e.preventDefault(); onCancel(); } }}></div>
      <div class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{editingIndex >= 0 ? 'Edit' : 'Add'} API Token</h4>
        <div class="space-y-2">
          <div>
            <label for="api-token-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
            <input id="api-token-name" type="text" bind:value={token.name} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm" />
          </div>
          <div>
            <label for="api-token-baseurl" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL</label>
            <input id="api-token-baseurl" type="text" bind:value={token.baseUrl} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm" />
          </div>
          <div>
            <label for="api-token-username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username (optional)</label>
            <input id="api-token-username" type="text" bind:value={token.username} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm" />
          </div>
          <div>
            <label for="api-token-token" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Token (secret)</label>
            <input id="api-token-token" type="password" bind:value={token.token} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono" />
          </div>
        </div>
        <div class="mt-4 flex justify-end gap-2">
          <button class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm" on:click={onCancel}>Cancel</button>
          <button class="px-3 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700" on:click={onSave}>Save</button>
        </div>
      </div>
    </div>
  {/if}
</div>
