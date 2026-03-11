<script>
  import {
    AddAPIToken,
    DeleteAPIToken,
    RevealAPIToken,
    CheckAllAPITokenSecrets,
  } from '../../wailsjs/go/main/App';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import RevealDialog from './RevealDialog.svelte';
  import { Toast } from 'flowbite-svelte';
  import { tick } from 'svelte';
  import Icon from '@iconify/svelte';

  let {
    apiTokens = [],
    disabled = false,
    onAdd = (_token) => {},
    onEdit = (_payload) => {},
    onDelete = (_index) => {},
  } = $props();

  let showModal = $state(false);
  let showHelp = $state(false);
  let editingIndex = $state(-1);
  let token = $state({ name: '', baseUrl: '', username: '', token: '' });

  // keyring status map: token name → true (found) | false (missing) | undefined (unchecked)
  /** @type {Record<string, boolean>} */
  let keyringStatus = $state({});
  let checkingKeyring = $state(false);

  async function checkAllTokens() {
    if (checkingKeyring) return;
    checkingKeyring = true;
    try {
      const result = await CheckAllAPITokenSecrets();
      keyringStatus = result || {};
    } catch (e) {
      showToastMsg('Failed to check keyring: ' + (e && e.message ? e.message : e), 'error');
    } finally {
      checkingKeyring = false;
    }
  }

  // toast state (local, transient feedback in place of alert())
  let showToast = $state(false);
  let toastMessage = $state('');
  let toastType = $state('success'); // 'success' | 'error' | 'info'
  let toastTimeoutId = null;

  async function showToastMsg(msg, type = 'success', duration = 3500) {
    if (toastTimeoutId) {
      clearTimeout(toastTimeoutId);
      toastTimeoutId = null;
    }
    toastMessage = msg || '';
    toastType = type || 'success';

    // Restart animation by toggling
    showToast = false;
    await tick();
    showToast = true;

    toastTimeoutId = setTimeout(() => {
      showToast = false;
      toastTimeoutId = null;
    }, duration);
  }

  // reveal state for the Reveal modal
  let reveal = $state({ show: false, name: '', token: '', loading: false, error: '' });

  // Confirm dialog state for delete/reveal operations
  let showDeleteTokenConfirm = $state(false);
  let deleteTokenIndex = $state(-1);
  let deleteTokenName = $state('');

  let showRevealTokenConfirm = $state(false);
  let revealTokenIndex = $state(-1);
  let revealTokenName = $state('');

  function openAdd() {
    editingIndex = -1;
    token = { name: '', baseUrl: '', username: '', token: '' };
    showModal = true;
  }

  function openEdit(i) {
    editingIndex = i;
    // copy metadata but never expose the stored secret; leave token input blank so user can enter a new secret if desired
    token = Object.assign({}, apiTokens[i]);
    // normalize baseUrl: support both camelCase and snake_case from persisted config
    token.baseUrl = apiTokens[i].baseUrl || apiTokens[i].base_url || '';
    token.token = '';
    showModal = true;
  }

  function onCancel() {
    showModal = false;
  }

  async function onSave() {
    // onSave may be called from a click handler; no anchor tracking is performed

    // basic validation
    if (!token.name || token.name.trim() === '') return;

    const entry = {
      name: token.name.trim(),
      // keep both forms so YAML persistence (snake_case) and JSON/Wails (camelCase) are both satisfied
      base_url: token.baseUrl || '',
      baseUrl: token.baseUrl || '',
      username: token.username || '',
      token: '',
    };

    // If adding
    if (editingIndex < 0) {
      // raw token must be provided when adding
      if (!token.token || token.token.trim() === '') {
        showToastMsg('Please provide the token value to store in the system keyring.', 'error');
        return;
      }
      try {
        await AddAPIToken(entry, token.token);
      } catch (e) {
        showToastMsg('Failed to store token: ' + (e && e.message ? e.message : e), 'error');
        return;
      }
      // set metadata token field to indicate managed status
      entry.token = '';
      onAdd(entry);
    } else {
      // editing existing entry
      const payload = { index: editingIndex, entry: Object.assign({}, apiTokens[editingIndex]) };
      // update metadata fields
      payload.entry.name = entry.name;
      payload.entry.base_url = entry.base_url;
      payload.entry.baseUrl = entry.baseUrl;
      payload.entry.username = entry.username;

      // if user provided a new raw token, store it (overwrite existing secret)
      if (token.token && token.token.trim() !== '') {
        try {
          await AddAPIToken(payload.entry, token.token);
        } catch (e) {
          showToastMsg('Failed to update token: ' + (e && e.message ? e.message : e), 'error');
          return;
        }
        payload.entry.token = '';
      }

      onEdit(payload);
    }

    showModal = false;
  }

  async function onDeleteClick(i) {
    // open confirm dialog and handle deletion on confirm
    deleteTokenIndex = i;
    deleteTokenName = apiTokens[i].name;
    showDeleteTokenConfirm = true;
  }

  async function onRevealClick(i) {
    // Ask user to confirm reveal action via ConfirmDialog, actual reveal proceeds on confirm
    revealTokenIndex = i;
    revealTokenName = apiTokens[i].name;
    showRevealTokenConfirm = true;
  }

  function closeReveal() {
    reveal.show = false;
    reveal.name = '';
    reveal.token = '';
    reveal.loading = false;
    reveal.error = '';
  }

  // Handlers for ConfirmDialog events
  async function handleConfirmDelete() {
    const i = deleteTokenIndex;
    showDeleteTokenConfirm = false;
    if (i < 0 || i >= apiTokens.length) return;
    const name = apiTokens[i].name;
    try {
      await DeleteAPIToken(name);
    } catch (e) {
      showToastMsg('Failed to delete token: ' + (e && e.message ? e.message : e), 'error');
      return;
    }
    // remove keyring status entry
    const newStatus = { ...keyringStatus };
    delete newStatus[name];
    keyringStatus = newStatus;
    onDelete(i);
    deleteTokenIndex = -1;
    deleteTokenName = '';
  }

  function handleCancelDelete() {
    showDeleteTokenConfirm = false;
    deleteTokenIndex = -1;
    deleteTokenName = '';
  }

  async function handleConfirmReveal() {
    const i = revealTokenIndex;
    showRevealTokenConfirm = false;
    if (i < 0 || i >= apiTokens.length) return;
    const name = apiTokens[i].name;
    reveal.show = true;
    reveal.name = name;
    reveal.loading = true;
    reveal.error = '';
    reveal.token = '';
    try {
      const tok = await RevealAPIToken(name);
      reveal.token = tok || '';
    } catch (e) {
      reveal.error = e && e.message ? e.message : String(e);
    } finally {
      reveal.loading = false;
    }
    revealTokenIndex = -1;
    revealTokenName = '';
  }

  function handleCancelReveal() {
    showRevealTokenConfirm = false;
    revealTokenIndex = -1;
    revealTokenName = '';
  }
</script>

<div
  class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700"
>
  <!-- Toast (rendered inline near anchor) -->
  {#if showToast}
    <div class="right-4 bottom-4 w-72 fixed z-50" role="status" aria-live="polite">
      {#if toastType === 'error'}
        <Toast color="red">
          {#snippet icon()}
            <Icon icon="mdi:close-circle" class="h-5 w-5 text-red-600" />
            <span class="sr-only">Error icon</span>
          {/snippet}
          <div class="flex items-center w-full">
            <div class="font-medium text-gray-900 dark:text-gray-100">{toastMessage}</div>
          </div>
        </Toast>
      {:else if toastType === 'info'}
        <Toast color="blue">
          {#snippet icon()}
            <Icon icon="mdi:information" class="h-5 w-5 text-gray-600" />
            <span class="sr-only">Info icon</span>
          {/snippet}
          <div class="flex items-center w-full">
            <div class="font-medium text-gray-900 dark:text-gray-100">{toastMessage}</div>
          </div>
        </Toast>
      {:else}
        <Toast color="green">
          {#snippet icon()}
            <Icon icon="mdi:check" class="h-5 w-5 text-green-600" />
            <span class="sr-only">Success icon</span>
          {/snippet}
          <div class="flex items-center w-full">
            <div class="font-medium text-gray-900 dark:text-gray-100">{toastMessage}</div>
          </div>
        </Toast>
      {/if}
    </div>
  {/if}

  <div class="flex items-center justify-between">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">API Tokens</h3>
    <div class="flex items-center gap-2">
      <button
        class="text-xs px-2 py-1 rounded bg-gray-200 dark:bg-gray-700"
        onclick={() => (showHelp = true)}
        aria-label="What does empty token mean?"
      >
        ⓘ Info
      </button>
      <button
        class="text-xs px-2 py-1 rounded bg-gray-200 dark:bg-gray-700 flex items-center gap-1 disabled:opacity-50"
        onclick={checkAllTokens}
        disabled={checkingKeyring || disabled}
        title="Check all tokens against the system keyring"
        aria-label="Check all tokens in keyring"
      >
        {#if checkingKeyring}
          <Icon icon="mdi:loading" class="h-4 w-4 animate-spin" />
        {:else}
          <Icon icon="mdi:shield-check-outline" class="h-4 w-4" />
        {/if}
        Check
      </button>
      <button
        class="px-3 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50"
        onclick={openAdd}
        {disabled}>Add</button
      >
    </div>
  </div>
  {#if apiTokens && apiTokens.length > 0}
    <ul class="space-y-1 mt-1 text-sm text-gray-800 dark:text-gray-200">
      {#each apiTokens as t, i}
        <li
          class="flex items-center justify-between p-2 rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-700"
        >
          <div class="flex items-center gap-2 min-w-0">
            <!-- Keyring status icon (visible after Check is run) -->
            {#if t.name in keyringStatus}
              {#if keyringStatus[t.name]}
                <span title="Secret found in keyring" class="shrink-0">
                  <Icon icon="mdi:check-circle" class="h-4 w-4 text-green-500" />
                </span>
              {:else}
                <span title="Secret NOT found in keyring" class="shrink-0">
                  <Icon icon="mdi:alert-circle" class="h-4 w-4 text-red-500" />
                </span>
              {/if}
            {:else}
              <span title="Keyring status not checked yet" class="shrink-0">
                <Icon icon="mdi:circle-outline" class="h-4 w-4 text-gray-300 dark:text-gray-600" />
              </span>
            {/if}
            <div class="min-w-0">
              <div class="font-medium truncate">{t.name}{t.username ? ` — ${t.username}` : ''}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                {t.baseUrl || t.base_url || ''}
              </div>
              {#if t.token}
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">{t.token}</div>
              {/if}
            </div>
          </div>
          <div class="flex items-center gap-2 ml-2 shrink-0">
            <button
              class="text-sm px-2 py-1 rounded bg-gray-200 dark:bg-gray-600"
              onclick={() => openEdit(i)}
              {disabled}>Edit</button
            >
            <button
              class="text-sm px-2 py-1 rounded bg-blue-600 text-white"
              onclick={() => onRevealClick(i)}
              {disabled}>Reveal</button
            >
            <button
              class="text-sm px-2 py-1 rounded bg-red-600 text-white"
              onclick={() => onDeleteClick(i)}
              {disabled}>Delete</button
            >
          </div>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="text-sm text-gray-600 dark:text-gray-400 mt-2">
      No API tokens configured. Add one to reference by name from data sources.
    </p>
  {/if}

  {#if showModal}
    <div class="fixed inset-0 z-40 flex items-center justify-center">
      <div
        class="absolute inset-0 bg-black/40"
        role="button"
        tabindex="0"
        aria-label="Close dialog"
        onclick={onCancel}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') {
            e.preventDefault();
            onCancel();
          }
        }}
      ></div>
      <div
        class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
      >
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
          {editingIndex >= 0 ? 'Edit' : 'Add'} API Token
        </h4>
        <div class="space-y-2">
          <div>
            <label
              for="api-token-name"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label
            >
            <input
              id="api-token-name"
              type="text"
              bind:value={token.name}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm subtle-placeholder"
            />
          </div>
          <div>
            <label
              for="api-token-baseurl"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Base URL</label
            >
            <input
              id="api-token-baseurl"
              type="text"
              bind:value={token.baseUrl}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm subtle-placeholder"
              list="baseUrl-suggestions"
            />
            <datalist id="baseUrl-suggestions">
              <option value="https://indico.jacow.org">https://indico.jacow.org</option>
              <option value="https://indico.global">https://indico.global</option>
            </datalist>
          </div>
          <div>
            <label
              for="api-token-username"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Username (optional)</label
            >
            <input
              id="api-token-username"
              type="text"
              bind:value={token.username}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm subtle-placeholder"
            />
          </div>
          <div>
            <label
              for="api-token-token"
              class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
              >Token (secret)</label
            >
            <input
              id="api-token-token"
              type="password"
              bind:value={token.token}
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono subtle-placeholder"
            />
          </div>
        </div>
        <div class="mt-4 flex justify-end gap-2">
          <button class="px-2 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm" onclick={onCancel}
            >Cancel</button
          >
          <button
            class="px-3 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700"
            onclick={onSave}>Save</button
          >
        </div>
      </div>
    </div>
  {/if}

  {#if showHelp}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div
        class="absolute inset-0 bg-black/40"
        role="button"
        tabindex="0"
        onclick={() => (showHelp = false)}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') {
            e.preventDefault();
            showHelp = false;
          }
        }}
      ></div>
      <div
        class="relative z-50 w-full max-w-lg mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
      >
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
          What the empty token means
        </h4>
        <p class="text-sm text-gray-700 dark:text-gray-300 mb-2">
          The actual API token value is not stored in the YAML configuration file. Instead, it is
          stored securely in your operating system's secret store (keychain/credential
          manager/Secret Service). The configuration keeps a reference by name which the application
          uses to look up the secret at runtime.
        </p>
        <p class="text-sm text-gray-700 dark:text-gray-300 mb-2">Notes:</p>
        <ul class="list-disc ml-4 mt-1 text-sm text-gray-700 dark:text-gray-300 mb-2">
          <li>The token is stored locally and encrypted by the OS where possible.</li>
          <li>
            On Linux this typically uses the Secret Service (gnome-keyring/libsecret) via D-Bus —
            make sure a session secret service is available if you run the app in a desktop session.
          </li>
          <li>
            In headless or CI environments the OS keyring may not be available; you can use the CLI
            tool <code>manage-secrets</code> to manage tokens or provide tokens via environment variables
            if you prefer.
          </li>
          <li>
            The UI allows you to reveal a token temporarily. Use this cautiously and avoid pasting
            tokens into insecure places.
          </li>
        </ul>
        <div class="flex justify-end">
          <button
            class="px-3 py-1 rounded bg-indigo-600 text-white"
            onclick={() => (showHelp = false)}>Close</button
          >
        </div>
      </div>
    </div>
  {/if}

  <!-- Confirm dialogs for delete and reveal actions -->
  <ConfirmDialog
    bind:open={showDeleteTokenConfirm}
    title="Delete API Token"
    message={`Delete API token "${deleteTokenName || ''}"? The entry will be removed from the config. If the secret exists in the system keyring it will also be deleted.`}
    confirmLabel="Delete"
    cancelLabel="Cancel"
    danger={true}
    onConfirm={handleConfirmDelete}
    onCancel={handleCancelDelete}
  />

  <ConfirmDialog
    bind:open={showRevealTokenConfirm}
    title="Reveal API Token"
    message={`Reveal API token for "${revealTokenName || ''}"? The value will be shown on screen.`}
    confirmLabel="Reveal"
    cancelLabel="Cancel"
    danger={false}
    onConfirm={handleConfirmReveal}
    onCancel={handleCancelReveal}
  />

  <RevealDialog
    bind:open={reveal.show}
    name={reveal.name}
    token={reveal.token}
    loading={reveal.loading}
    error={reveal.error}
    onClose={closeReveal}
  />
</div>
