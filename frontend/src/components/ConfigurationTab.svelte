<script>
  import { onMount } from 'svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';
  import { TrashBinOutline } from 'flowbite-svelte-icons';

  let configData = null;
  let loading = true;
  let applying = false;
  let applyError = '';
  let applySuccess = '';
  // Toast state
  let showToast = false;
  let toastMessage = '';
  let toastType = 'success'; // 'success' | 'error' | 'info'
  let toastTimeoutId = null;

  function showToastMsg(msg, type = 'success', duration = 3500) {
    // clear previous timeout
    if (toastTimeoutId) {
      clearTimeout(toastTimeoutId);
      toastTimeoutId = null;
    }
    toastMessage = msg || '';
    toastType = type || 'success';
    showToast = true;
    // auto-hide
    toastTimeoutId = setTimeout(() => {
      showToast = false;
      toastTimeoutId = null;
    }, duration);
  }

  // expandedSources keyed by data-source index to avoid problems when renaming
  let expandedSources = {};
  // track active selection by index so we can rename sources safely
  let currentActiveIndex = 0;
  let selectedActiveIndex = 0;
  let showConfigFile = false;
  // name validation errors keyed by data-source index
  let nameErrors = {};

  let indicoDataSourcePlaceholders = {
    baseUrl: 'https://indico.example.org',
    eventId: '123',
    apiToken: 'indp_...',
    timeout: '60s'
  }

  let testDataSourcePlaceholders = {
    dataDir: './testdata',
    eventInfo: 'info.json',
    abstracts: 'abstracts.json',
    contribs: 'contribs.json'
  }

  // -- New state for Add Indico dialog --
  let showAddIndico = false;
  let newIndico = {
    name: '',
    baseUrl: 'https://',
    eventId: 0,
    apiToken: '',
    timeout: '60s'
  };

  // Errors for the Add Indico dialog fields
  let newIndicoErrors = {};

  // Helper to validate the newIndico fields client-side
  function validateNewIndico() {
    newIndicoErrors = {};
    // name
    if (!newIndico.name || String(newIndico.name).trim() === '') {
      newIndicoErrors.name = 'Name is required';
    }
    // baseUrl: must be a valid http/https URL
    try {
      const url = new URL(String(newIndico.baseUrl || '').trim());
      if (url.protocol !== 'http:' && url.protocol !== 'https:') {
        newIndicoErrors.baseUrl = 'Base URL must start with http:// or https://';
      }
    } catch (e) {
      newIndicoErrors.baseUrl = 'Base URL is not a valid URL';
    }
    // eventId: must be non-empty numeric
    if (newIndico.eventId === null || newIndico.eventId === undefined) {
      newIndicoErrors.eventId = 'Event ID is required';
    } else if (isNaN(Number(newIndico.eventId)) || Number(newIndico.eventId) <= 0) {
      newIndicoErrors.eventId = 'Event ID must be a positive number';
    } else if (!Number.isInteger(Number(newIndico.eventId))) {
      newIndicoErrors.eventId = 'Event ID must be an integer';
    }
    // timeout: basic pattern like 500ms, 15s, 1m, 2h
    const t = String(newIndico.timeout || '').trim();
    if (!/^\d+(ms|s|m|h)$/.test(t)) {
      newIndicoErrors.timeout = 'Timeout must be a duration like 500ms, 15s, 1m, or 2h';
    }
    // apiToken: optional but if present should be non-empty after trim
    if (newIndico.apiToken !== null && newIndico.apiToken !== undefined) {
      if (String(newIndico.apiToken).trim() === '') {
        // allow empty token (some servers might not need it) - do not error
      }
    }

    return Object.keys(newIndicoErrors).length === 0;
  }

  function openAddIndicoDialog() {
    // initialize template values and open dialog
    newIndico.name = getUniqueName('Conference Name');
    // provide sensible defaults so validation passes immediately and Save is clickable
    newIndico.baseUrl = indicoDataSourcePlaceholders.baseUrl;
    newIndico.eventId = parseInt(String(indicoDataSourcePlaceholders.eventId), 10);
    newIndico.apiToken = indicoDataSourcePlaceholders.apiToken;
    newIndico.timeout = indicoDataSourcePlaceholders.timeout;
    newIndicoErrors = {};
    showAddIndico = true;
  }

  function cancelAddIndico() {showAddIndico = false;
    newIndicoErrors = {};
  }

  async function saveNewIndico() {
    // client-side validation
    if (!validateNewIndico()) {
      return;
    }

    // basic validation: name non-empty (already validated above)
    const name = (newIndico.name || '').trim();
    // ensure configData.dataSources exists
    if (!configData) configData = {};
    if (!Array.isArray(configData.dataSources)) configData.dataSources = [];

    // If name collides, make it unique by appending suffix
    const existingNames = new Set(configData.dataSources.map(ds => (ds && ds.name) ? String(ds.name) : ''));
    let finalName = name;
    if (existingNames.has(finalName)) {
      let i = 2;
      while (existingNames.has(`${finalName} (${i})`)) i++;
      finalName = `${finalName} (${i})`;
    }

    // build the new data source object
    const newSource = {
      name: finalName,
      type: 'indico',
      indico: {
        baseUrl: (newIndico.baseUrl || '').trim(),
        // ensure eventId is sent as an integer
        eventId: Number.isInteger(Number(newIndico.eventId)) ? parseInt(String(newIndico.eventId), 10) : Number(newIndico.eventId),
        apiToken: newIndico.apiToken || '',
        timeout: newIndico.timeout || '60s'
      }
    };

    // push to config and update UI state
    const newIndex = configData.dataSources.length;
    configData.dataSources.push(newSource);
    expandedSources[newIndex] = true; // expand newly added source
    // re-run name validation and set selection to the new source
    validateNames();
    selectedActiveIndex = newIndex;

    // Persist immediately by calling apply(); await result and close modal on success
    const ok = await apply();
    if (ok) {
      // If apply succeeded, clear modal and errors
      showAddIndico = false;
      newIndicoErrors = {};
    } else {
      // apply() set applyError; keep modal open so user can fix issues
    }
  }

  // Helper to create a unique default name
  function getUniqueName(base = 'Conference Name') {
    if (!configData || !Array.isArray(configData.dataSources) || configData.dataSources.length === 0) return base;
    const existing = new Set(configData.dataSources.map(ds => (ds && ds.name) ? String(ds.name) : ''));
    if (!existing.has(base)) return base;
    let i = 2;
    while (existing.has(`${base} (${i})`)) i++;
    return `${base} (${i})`;
  }

  // Validate data source names: non-empty and unique
  function validateNames() {
    nameErrors = {};
    if (!configData || !Array.isArray(configData.dataSources)) return;
    const counts = {};
    // First pass: trim names and count occurrences
    configData.dataSources.forEach((ds, i) => {
      // Normalize name by trimming whitespace and write back so UI shows normalized value
      const rawName = (ds && ds.name) ? String(ds.name) : '';
      const name = rawName.trim();
      if (ds) ds.name = name;
      // store trimmed name back into configData

      if (!name) {
        nameErrors[i] = 'Name cannot be empty';
      } else {
        counts[name] = (counts[name] || 0) + 1;
      }
    });
    // find duplicates
    configData.dataSources.forEach((ds, i) => {
      const name = (ds && ds.name) ? String(ds.name).trim() : '';
      if (name && counts[name] > 1) {
        nameErrors[i] = 'Name must be unique';
      }
    });
  }

  // Re-validate whenever configData changes
  $: if (configData) validateNames();

  // Derived flag used to disable Apply when there are name validation errors
  $: hasNameErrors = Object.values(nameErrors).some(Boolean);

  // Reactive validity for newIndico so the Save button can be disabled
  $: newIndicoValid = validateNewIndico();

  async function loadConfig() {
    try {
      configData = await GetStructuredConfigUI();
      // Initialize cache config with defaults if not present
      if (!configData.cache) {
        configData.cache = {
          ttl: '24h',
          maxSize: '100MB',
          cacheDir: ''
        };
      }
      // Initialize all sources as collapsed (by index)
      (configData.dataSources || []).forEach((ds, i) => {
        expandedSources[i] = false;
      });

      // ensure names are validated for the loaded config
      validateNames();

      // find active index from name provided by backend; default to 0
      selectedActiveIndex = (configData.dataSources || []).findIndex(ds => ds.name === configData.activeDataSourceName);
      if (selectedActiveIndex < 0) selectedActiveIndex = 0;
      currentActiveIndex = selectedActiveIndex;
      loading = false;
    } catch (e) {
      loading = false;
      applyError = `Failed to load config: ${e}`;
    }
  }

  onMount(loadConfig);

  function toggleSource(index) {
    expandedSources[index] = !expandedSources[index];
  }

  async function apply() {
    applyError = '';
    applySuccess = '';
    applying = true;
    try {
      // validate before applying
      validateNames();
      if (Object.values(nameErrors).some(Boolean)) {
        applyError = 'Please fix data source name errors before applying.';
        return false;
      }
      // Validate and coerce indico eventId fields to integers so backend can unmarshal
      if (configData && Array.isArray(configData.dataSources)) {
        for (let i = 0; i < configData.dataSources.length; i++) {
          const ds = configData.dataSources[i];
          if (ds && ds.type === 'indico' && ds.indico) {
            const ev = ds.indico.eventId;
            if (ev === '' || ev === null || ev === undefined || isNaN(Number(ev)) || Number(ev) <= 0 || !Number.isInteger(Number(ev))) {
              applyError = `Event ID for data source "${ds.name || ('#' + i)}" must be a positive integer`;
              return false;
            }
            // coerce to integer
            ds.indico.eventId = parseInt(String(ds.indico.eventId), 10);
          }
        }
      }
      // Ensure backend activeDataSourceName is set from the currently selected index
      if (configData && configData.dataSources && configData.dataSources[selectedActiveIndex]) {
        configData.activeDataSourceName = configData.dataSources[selectedActiveIndex].name;
      }
      await ApplyStructuredConfigUI(configData);
      currentActiveIndex = selectedActiveIndex;
      applySuccess = 'Configuration applied successfully';
      // show a transient toast for success
      showToastMsg(applySuccess, 'success');
      return true;
    } catch (e) {
      applyError = `Failed to apply configuration: ${e}`;
      // show error toast too
      showToastMsg(applyError, 'error');
      return false;
    } finally {
      applying = false;
    }
  }

  // -- Delete data source state and handlers --
  let showDeleteConfirm = false;
  let deleteIndex = null;
  let deleteName = '';

  function openDeleteConfirm(i) {
    deleteIndex = i;
    deleteName = (configData?.dataSources?.[i]?.name) || '';
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    showDeleteConfirm = false;
    deleteIndex = null;
    deleteName = '';
  }

  async function confirmDelete() {
    if (deleteIndex === null || !configData || !Array.isArray(configData.dataSources)) return;
    // remove the data source
    configData.dataSources.splice(deleteIndex, 1);
    // rebuild expandedSources to avoid stale keys
    const newExpanded = {};
    (configData.dataSources || []).forEach((_, idx) => {
      newExpanded[idx] = !!expandedSources[idx];
    });
    expandedSources = newExpanded;

    // adjust selectedActiveIndex and currentActiveIndex
    if (selectedActiveIndex === deleteIndex) {
      selectedActiveIndex = 0;
    } else if (selectedActiveIndex > deleteIndex) {
      selectedActiveIndex = Math.max(0, selectedActiveIndex - 1);
    }
    if (currentActiveIndex === deleteIndex) {
      currentActiveIndex = 0;
    } else if (currentActiveIndex > deleteIndex) {
      currentActiveIndex = Math.max(0, currentActiveIndex - 1);
    }

    // re-run name validation
    validateNames();

    // persist changes
    const ok = await apply();
    if (ok) {
      showToastMsg(`Deleted data source "${deleteName || ''}"`, 'success');
      cancelDelete();
    } else {
      // apply() will have set applyError and shown toast
    }
  }
</script>

<style>
  /* Dim placeholders to distinguish them from filled input text. Use slightly different colors in light and dark mode. */
  :global(input::placeholder), :global(textarea::placeholder) {
    color: rgba(107, 114, 128, 0.6); /* gray-500 @ 60% */
  }
  :global(.dark input::placeholder), :global(.dark textarea::placeholder) {
    color: rgba(148, 163, 184, 0.55); /* slate-300-ish for dark mode */
  }
</style>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  <!-- Toast (top-right) -->
  <div class="fixed right-4 top-4 z-50 pointer-events-none">
    <div class="transform transition-all duration-300 ease-out {showToast ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2'}">
      <div class="pointer-events-auto max-w-xs w-full rounded-lg shadow-lg overflow-hidden">
        <div class="flex items-start gap-3 p-3 {toastType === 'success' ? 'bg-green-50 border border-green-200 text-green-800' : toastType === 'error' ? 'bg-red-50 border border-red-200 text-red-800' : 'bg-gray-50 border border-gray-200 text-gray-800'} rounded-lg">
          <div class="shrink-0 mt-0.5">
            {#if toastType === 'success'}
              <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
            {:else if toastType === 'error'}
              <svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            {:else}
              <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01"/></svg>
            {/if}
          </div>
          <div class="flex-1 text-sm leading-tight">
            <div class="font-medium">{toastMessage}</div>
          </div>
          <button class="ml-2 text-xs text-gray-500 hover:text-gray-700" on:click={() => { showToast = false; if (toastTimeoutId) { clearTimeout(toastTimeoutId); toastTimeoutId = null; } }} aria-label="Dismiss toast">×</button>
        </div>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center p-4">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading configuration...</p>
      </div>
    </div>
  {:else}
    <!-- Active Data Source -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Active Data Source</h3>
      <div class="flex items-center gap-2">
        <label for="active-source" class="text-sm font-medium text-gray-700 dark:text-gray-300">Select Data Source:</label>
        <select
          id="active-source"
          bind:value={selectedActiveIndex}
          class="flex-1 rounded border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {#each configData.dataSources as ds, i}
            <option value={i}>{ds.name}</option>
          {/each}
        </select>
      </div>
      <div class="mt-1">
        <span class="text-sm text-green-600 dark:text-green-400">Current Active: <strong>{configData.dataSources?.[currentActiveIndex]?.name}</strong></span>
      </div>
    </div>

    <!-- Data Sources -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Data Sources</h3>
        <!-- Add Indico Source button -->
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="px-2 py-1 rounded-lg bg-green-600 text-white text-sm font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-green-500 dark:hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-400"
            on:click={openAddIndicoDialog}
            disabled={loading}
            title="Add a new Indico data source"
          >
            + Add Indico Source
          </button>
        </div>
      </div>
      <div class="space-y-1">

        {#each configData.dataSources as dataSource, i (i)}
        <div class="bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
          <!-- Header -->
          <div
            role="button"
            tabindex="0"
            on:click={() => toggleSource(i)}
            on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleSource(i); } }}
            class="w-full flex items-center justify-between p-1 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            <span class="flex items-center gap-2">
              <!-- Make the name editable -->
              <input
                id={`ds-name-${i}`}
                type="text"
                bind:value={dataSource.name}
                on:input={validateNames}
                placeholder="Data source name"
                title="Edit data source name"
                aria-label={`Data source name ${i}`}
                class="text-xl md:text-lg font-semibold text-gray-900 dark:text-gray-100 bg-transparent border-b-2 border-transparent focus:border-indigo-500 px-1 py-0.5 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-sm transition-colors placeholder-gray-400 cursor-text"
              />
              {#if nameErrors[i]}
                <span class="ml-2 text-red-500 text-xs font-medium" title={nameErrors[i]}>{nameErrors[i]}</span>
              {/if}
              <!-- pencil icon to indicate editability -->
              <svg aria-hidden="true" class="w-4 h-4 text-gray-400 ml-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4 12.5-12.5z" />
              </svg>
              <span class="px-2 py-0.5 text-xs rounded-full {dataSource.type === 'indico' ? 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200' : 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200'}">
                {dataSource.type === 'indico' ? 'API' : 'Test Data'}
              </span>
              {#if currentActiveIndex === i}
                <span class="px-2 py-0.5 text-xs rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200">Active</span>
              {/if}
            </span>
            <div class="flex items-center gap-2">
              <!-- Delete button: stop propagation so header click doesn't toggle -->
              <button
                type="button"
                class="text-red-500 hover:text-red-700 p-1 rounded focus:outline-none focus:ring-2 focus:ring-red-300"
                on:click|preventDefault|stopPropagation={() => openDeleteConfirm(i)}
                aria-label={`Delete data source ${dataSource.name || i}`}
                title="Delete data source"
              >
                <TrashBinOutline class="w-4 h-4" />
              </button>
              <svg class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedSources[i] ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
              </svg>
            </div>
          </div>

          <!-- Content -->
          {#if expandedSources[i]}
            <div class="px-4 pb-4 pt-2 border-t border-gray-200 dark:border-gray-700 space-y-2">
              {#if dataSource.type === 'indico' && dataSource.indico}
                <!-- Indico API Configuration -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                  <div>
                    <label for={`ds-${i}-baseUrl`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL</label>
                    <input
                      id={`ds-${i}-baseUrl`}
                      type="text"
                      bind:value={dataSource.indico.baseUrl}
                      placeholder={indicoDataSourcePlaceholders.baseUrl}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-eventId`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label>
                    <input
                      id={`ds-${i}-eventId`}
                      type="number"
                      bind:value={dataSource.indico.eventId}
                      placeholder={indicoDataSourcePlaceholders.eventId}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div class="md:col-span-2">
                    <label for={`ds-${i}-apiToken`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label>
                    <input
                      id={`ds-${i}-apiToken`}
                      type="password"
                      bind:value={dataSource.indico.apiToken}
                      placeholder={indicoDataSourcePlaceholders.apiToken}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-timeout`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
                    <input
                      id={`ds-${i}-timeout`}
                      type="text"
                      bind:value={dataSource.indico.timeout}
                      placeholder={indicoDataSourcePlaceholders.timeout}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 15s, 1m, 500ms</p>
                  </div>
                </div>
              {:else if dataSource.type === 'test' && dataSource.test}
                <!-- Test Data Configuration -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                  <div class="md:col-span-2">
                    <label for={`ds-${i}-test-dataDir`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Data Directory</label>
                    <input
                      id={`ds-${i}-test-dataDir`}
                      type="text"
                      bind:value={dataSource.test.dataDir}
                      placeholder={testDataSourcePlaceholders.dataDir}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-eventInfo`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event Info File</label>
                    <input
                      id={`ds-${i}-test-eventInfo`}
                      type="text"
                      bind:value={dataSource.test.eventInfo}
                      placeholder={testDataSourcePlaceholders.eventInfo}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-abstracts`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Abstracts File</label>
                    <input
                      id={`ds-${i}-test-abstracts`}
                      type="text"
                      bind:value={dataSource.test.abstracts}
                      placeholder={testDataSourcePlaceholders.abstracts}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-contribs`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Contributions File</label>
                    <input
                      id={`ds-${i}-test-contribs`}
                      type="text"
                      bind:value={dataSource.test.contribs}
                      placeholder={testDataSourcePlaceholders.contribs}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                </div>
              {/if}
              <!-- Error message for data source name -->
              {#if nameErrors[i]}
                <div class="text-red-500 text-sm mt-1">{nameErrors[i]}</div>
              {/if}
            </div>
          {/if}
        </div>
        {/each}
      </div>
    </div>

    <!-- Add Indico Dialog (Modal) -->
    {#if showAddIndico}
      <div class="fixed inset-0 z-40 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40" role="button" tabindex="0" aria-label="Close dialog" on:click={cancelAddIndico} on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar') { e.preventDefault(); cancelAddIndico(); } }}></div>
        <div role="dialog" aria-modal="true" tabindex="0" on:keydown|stopPropagation={(e) => { if (e.key === 'Escape') cancelAddIndico(); }} class="relative z-50 w-full max-w-lg mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
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
                     placeholder={indicoDataSourcePlaceholders.baseUrl} />
              {#if newIndicoErrors.baseUrl}
                <p class="text-red-500 text-xs mt-1">{newIndicoErrors.baseUrl}</p>
              {/if}
            </div>
            <div>
              <label for="new-indico-eventId" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label>
              <input id="new-indico-eventId" type="number" bind:value={newIndico.eventId} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                     placeholder={indicoDataSourcePlaceholders.eventId} />
              {#if newIndicoErrors.eventId}
                <p class="text-red-500 text-xs mt-1">{newIndicoErrors.eventId}</p>
              {/if}
            </div>
            <div>
              <label for="new-indico-apiToken" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label>
              <input id="new-indico-apiToken" type="text" bind:value={newIndico.apiToken} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono"
                     placeholder={indicoDataSourcePlaceholders.apiToken} />
            </div>
            <div>
              <label for="new-indico-timeout" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
              <input id="new-indico-timeout" type="text" bind:value={newIndico.timeout} class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                     placeholder={indicoDataSourcePlaceholders.timeout} />
              {#if newIndicoErrors.timeout}
                <p class="text-red-500 text-xs mt-1">{newIndicoErrors.timeout}</p>
              {/if}
            </div>
          </div>
          <div class="mt-4 flex justify-end gap-2">
            <button type="button" class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm pointer-events-auto" on:click={cancelAddIndico}>Cancel</button>
            <button
              type="button"
              class="px-3 py-1 rounded bg-indigo-600 text-white text-sm pointer-events-auto hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-indigo-400 disabled:hover:bg-indigo-400 focus:outline-none focus:ring-2 focus:ring-indigo-400 disabled:focus:ring-0 transition-colors"
              on:click={saveNewIndico}
              disabled={applying}
            >
              {applying ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Delete Confirmation Modal -->
    {#if showDeleteConfirm}
      <div class="fixed inset-0 z-40 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40" role="button" tabindex="0" aria-label="Close dialog" on:click={cancelDelete} on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar') { e.preventDefault(); cancelDelete(); } }}></div>
        <div role="dialog" aria-modal="true" tabindex="0" on:keydown|stopPropagation={(e) => { if (e.key === 'Escape') cancelDelete(); }} class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto">
          <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Delete Data Source</h4>
          <p class="text-sm text-gray-700 dark:text-gray-300">Are you sure you want to delete <strong>{deleteName || 'this data source'}</strong>? This action cannot be undone.</p>
          <div class="mt-4 flex justify-end gap-2">
            <button type="button" class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm" on:click={cancelDelete}>Cancel</button>
            <button type="button" class="px-3 py-1 rounded bg-red-600 text-white text-sm hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-red-400"
                    on:click={confirmDelete}
                    disabled={applying}
            >
              {applying ? 'Deleting...' : 'Delete'}
            </button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Cache Configuration -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Cache Configuration</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
        <div>
          <label for="cache-ttl" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            TTL (Time-To-Live)
            <span class="text-xs text-gray-500 dark:text-gray-400 ml-1" title="How long cache entries stay valid before expiring">ⓘ</span>
          </label>
          <input
            id="cache-ttl"
            type="text"
            bind:value={configData.cache.ttl}
            placeholder="24h"
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 24h, 1h30m, 30m</p>
        </div>
        <div>
          <label for="cache-maxsize" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Max Size
            <span class="text-xs text-gray-500 dark:text-gray-400 ml-1" title="Maximum cache size - oldest entries evicted when limit reached">ⓘ</span>
          </label>
          <input
            id="cache-maxsize"
            type="text"
            bind:value={configData.cache.maxSize}
            placeholder="100MB"
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 100MB, 1GB, 500MB</p>
        </div>
        <div>
          <label for="cache-dir" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Cache Directory
            <span class="text-xs text-gray-500 dark:text-gray-400 ml-1" title="Custom cache directory path (leave empty for default)">ⓘ</span>
          </label>
          <input
            id="cache-dir"
            type="text"
            bind:value={configData.cache.cacheDir}
            placeholder="~/.cache/IndicoDataFusion"
            class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Leave empty for default</p>
        </div>
      </div>
    </div>

    <!-- Status Messages -->
    {#if applyError}
      <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-2">
        <p class="text-sm text-red-600 dark:text-red-400">{applyError}</p>
      </div>
    {/if}
    {#if applySuccess}
      <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-2">
        <p class="text-sm text-green-600 dark:text-green-400">{applySuccess}</p>
      </div>
    {/if}

    <!-- Apply Button -->
    <div class="flex justify-end">
      <button
        type="button"
        class="px-2 py-1.5 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:ring-offset-2 transition-colors"
        on:click={apply}
        disabled={applying || hasNameErrors}
      >
        {applying ? 'Applying...' : 'Apply'}
      </button>
    </div>

    <!-- Configuration File Path Info (Collapsible) -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div
        role="button"
        tabindex="0"
        on:click={() => showConfigFile = !showConfigFile}
        on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); showConfigFile = !showConfigFile; } }}
        class="w-full flex items-center justify-between p-3 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      >
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Configuration File</h4>
        <svg class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {showConfigFile ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
        </svg>
      </div>
      {#if showConfigFile}
        <div class="px-4 pb-4 border-t border-gray-200 dark:border-gray-700 space-y-1">
          <div class="flex flex-wrap items-center gap-2 text-sm pt-3">
            <span class="font-medium text-gray-600 dark:text-gray-400">Path:</span>
            <span class="text-gray-800 dark:text-gray-200 font-mono text-xs break-all">{configData.pathInfo?.path || configData.path || 'Not set'}</span>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>
