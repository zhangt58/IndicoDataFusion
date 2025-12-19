<script>
  import { AddAPIToken, DeleteAPIToken, RevealAPIToken } from '../../wailsjs/go/main/App';

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

  // reveal state for the Reveal modal
  let reveal = $state({ show: false, name: '', token: '', loading: false, error: '' });

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
        alert('Please provide the token value to store in the system keyring.');
        return;
      }
      try {
        await AddAPIToken(entry, token.token);
      } catch (e) {
        alert('Failed to store token: ' + (e && e.message ? e.message : e));
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
          alert('Failed to update token: ' + (e && e.message ? e.message : e));
          return;
        }
        payload.entry.token = '';
      }

      onEdit(payload);
    }

    showModal = false;
  }

  async function onDeleteClick(i) {
    if (!confirm(`Delete API token "${apiTokens[i].name}"? This will remove the secret from the system keyring.`)) return;
    const name = apiTokens[i].name;
    try {
      await DeleteAPIToken(name);
    } catch (e) {
      alert('Failed to delete token from keyring: ' + (e && e.message ? e.message : e));
      return;
    }
    onDelete(i);
  }

  async function onRevealClick(i) {
    const name = apiTokens[i].name;
    if (!confirm(`Reveal API token for "${name}"? The value will be shown on screen.`)) return;
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
  }

  function closeReveal() {
    reveal.show = false;
    reveal.name = '';
    reveal.token = '';
    reveal.loading = false;
    reveal.error = '';
  }

  async function copyToClipboard(text) {
    try {
      if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        alert('Token copied to clipboard');
        return;
      }
    } catch (e) {
      // fall through to textarea fallback
    }
    // fallback
    const ta = document.createElement('textarea');
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand('copy');
      alert('Token copied to clipboard');
    } catch (e) {
      alert('Copy failed');
    }
    ta.remove();
  }
</script>

<div
  class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700"
>
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
          <div>
            <div class="font-medium">{t.name} {t.username ? ` — ${t.username}` : ''}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{t.baseUrl || t.base_url}</div>
            {#if t.token}
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{t.token}</div>
            {/if}
          </div>
          <div class="flex items-center gap-2">
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
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
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
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
            />
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
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
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
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono"
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

  {#if reveal.show}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div
        class="absolute inset-0 bg-black/40"
        role="button"
        tabindex="0"
        onclick={closeReveal}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') {
            e.preventDefault();
            closeReveal();
          }
        }}
      ></div>
      <div class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Reveal API Token — {reveal.name ?? ''}</h4>
        {#if reveal.loading}
          <p class="text-sm text-gray-600 dark:text-gray-400">Loading…</p>
        {:else if reveal.error}
          <p class="text-sm text-red-600">Error: {reveal.error}</p>
        {:else}
          <div class="mb-2">
            <div class="block text-xs text-gray-600 dark:text-gray-400 mb-1">Token value</div>
            <div class="font-mono wrap-break-word p-2 rounded bg-gray-100 dark:bg-gray-700 text-sm">{reveal.token}</div>
          </div>
          <div class="flex justify-end gap-2">
            <button class="px-2 py-1 rounded bg-gray-200" onclick={() => copyToClipboard(reveal.token)}>Copy</button>
            <button class="px-2 py-1 rounded bg-gray-200" onclick={closeReveal}>Close</button>
          </div>
        {/if}
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
      <div class="relative z-50 w-full max-w-lg mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
        <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">What the empty token means</h4>
        <p class="text-sm text-gray-700 dark:text-gray-300 mb-2">
          The actual API token value is not stored in the YAML configuration file. Instead, it is stored securely
          in your operating system's secret store (keychain/credential manager/Secret Service). The configuration
          keeps a reference by name which the application uses to look up the secret at runtime.
        </p>
        <p class="text-sm text-gray-700 dark:text-gray-300 mb-2">Notes:</p>
        <ul class="list-disc ml-4 mt-1 text-sm text-gray-700 dark:text-gray-300 mb-2">
          <li>The token is stored locally and encrypted by the OS where possible.</li>
          <li>On Linux this typically uses the Secret Service (gnome-keyring/libsecret) via D-Bus — make sure a session secret service is available if you run the app in a desktop session.</li>
          <li>In headless or CI environments the OS keyring may not be available; you can use the CLI tool <code>manage-secrets</code> to manage tokens or provide tokens via environment variables if you prefer.</li>
          <li>The UI allows you to reveal a token temporarily (for copy). Use this cautiously and avoid pasting tokens into insecure places.</li>
        </ul>
        <div class="flex justify-end">
          <button class="px-3 py-1 rounded bg-indigo-600 text-white" onclick={() => (showHelp = false)}>Close</button>
        </div>
      </div>
    </div>
  {/if}
</div>
