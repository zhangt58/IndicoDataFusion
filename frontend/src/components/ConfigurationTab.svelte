<script>
  import { onMount, tick } from 'svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';
  import DataSources from './DataSources.svelte';
  import IndicoConfig from './IndicoConfig.svelte';
  import ApiTokens from './ApiTokens.svelte';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import Icon from '@iconify/svelte';

  let configData = $state(null);
  let loading = $state(true);
  let applying = $state(false);
  let applyError = $state('');
  let applySuccess = $state('');
  // Toast state
  let showToast = $state(false);
  let toastMessage = $state('');
  let toastType = $state('success'); // 'success' | 'error' | 'info'
  let toastTimeoutId = null;

  // exposed list of API tokens (from configData.APITokens)
  let apiTokens = $derived(configData && configData.apiTokens ? configData.apiTokens : []);

  async function showToastMsg(msg, type = 'success', duration = 3500) {
    // clear previous timeout
    if (toastTimeoutId) {
      clearTimeout(toastTimeoutId);
      toastTimeoutId = null;
    }

    toastMessage = msg || '';
    toastType = type || 'success';

    // Restart the toast animation reliably by toggling the visibility and yielding a tick.
    // This ensures the inline toast to the left of the Apply button animates on repeated calls.
    showToast = false;
    await tick();
    showToast = true;

    // auto-hide after duration
    toastTimeoutId = setTimeout(() => {
      showToast = false;
      toastTimeoutId = null;
    }, duration);
  }

  // expandedSources keyed by data-source index to avoid problems when renaming
  let expandedSources = $state({});
  // track active selection by index so we can rename sources safely
  let currentActiveIndex = $state(0);
  let selectedActiveIndex = $state(0);
  let showConfigFile = $state(false);
  // top-level advanced panel that groups cache & API tokens
  let showAdvanced = $state(false);
  // Collapsible Cache Configuration inside Advanced (closed by default)
  let showCacheConfig = $state(false);
  // name validation errors keyed by data-source index
  let nameErrors = $state({});

  let indicoDataSourcePlaceholders = {
    confName: 'Conference name, e.g. IPAC25',
    baseUrl: 'https://indico.jacow.org',
    eventId: '123',
    timeout: '60s',
  };

  let testDataSourcePlaceholders = {
    dataDir: './testdata',
    eventInfo: 'info.json',
    abstracts: 'abstracts.json',
    contribs: 'contribs.json',
  };

  // -- Indico dialog integration: we use the extracted component --
  let indicoDialogOpen = $state(false);

  function openAddIndicoDialog() {
    // Do not prefill a suggested name here; leave the dialog name empty so placeholder is visible
    indicoDialogOpen = true;
  }

  // Handler when child component emits 'create' with the raw indico payload
  async function handleCreateIndico(event) {
    const payload = event.detail || event; // allow both direct and event.detail
    const nameRaw = (payload.name || '').trim();
    if (!configData) configData = {};
    if (!Array.isArray(configData.dataSources)) configData.dataSources = [];

    // Use provided name if present; otherwise leave name empty (validation will catch missing names)
    const finalName = nameRaw || '';

    const newSource = {
      name: finalName,
      type: 'indico',
      indico: {
        baseUrl: (payload.baseUrl || '').trim(),
        eventId: Number.isInteger(Number(payload.eventId))
          ? parseInt(String(payload.eventId), 10)
          : Number(payload.eventId),
        // use token name reference (payload may contain apiTokenName)
        apiTokenName: payload.apiTokenName || payload.apiToken || '',
        timeout: payload.timeout || '60s',
      },
    };

    const newIndex = configData.dataSources.length;
    configData.dataSources.push(newSource);
    expandedSources[newIndex] = true;
    validateNames();
    selectedActiveIndex = newIndex;

    const ok = await apply();
    if (ok) {
      indicoDialogOpen = false;
    }
  }

  function cancelCreateIndico() {
    indicoDialogOpen = false;
  }

  // Validate data source names: non-empty and unique
  function validateNames() {
    nameErrors = {};
    if (!configData || !Array.isArray(configData.dataSources)) return;
    const counts = {};
    // First pass: trim names and count occurrences
    configData.dataSources.forEach((ds, i) => {
      // Normalize name by trimming whitespace and write back so UI shows normalized value
      const rawName = ds && ds.name ? String(ds.name) : '';
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
      const name = ds && ds.name ? String(ds.name).trim() : '';
      if (name && counts[name] > 1) {
        nameErrors[i] = 'Name must be unique';
      }
    });
  }

  // New: sort dataSources by name (case-insensitive), but place items with type 'test' at the end.
  // Preserve the active data source selection by name.
  function sortDataSources() {
    if (!configData || !Array.isArray(configData.dataSources)) return;

    // remember active name so we can restore selection after sort
    const activeName =
      configData.activeDataSourceName ||
      (configData.dataSources[selectedActiveIndex] &&
        configData.dataSources[selectedActiveIndex].name) ||
      null;

    configData.dataSources.sort((a, b) => {
      const aIsTest = a && a.type === 'test';
      const bIsTest = b && b.type === 'test';
      if (aIsTest !== bIsTest) {
        // push test types to the end
        return aIsTest ? 1 : -1;
      }
      const na = a && a.name ? String(a.name).toLowerCase() : '';
      const nb = b && b.name ? String(b.name).toLowerCase() : '';
      if (na < nb) return -1;
      if (na > nb) return 1;
      return 0;
    });

    // rebuild expandedSources to avoid stale keys; default to collapsed
    const newExpanded = {};
    (configData.dataSources || []).forEach((_, idx) => {
      newExpanded[idx] = false;
    });
    expandedSources = newExpanded;

    // restore selected/current active indices based on activeName if available
    if (activeName) {
      const newIndex = (configData.dataSources || []).findIndex(
        (ds) => ds && ds.name === activeName,
      );
      if (newIndex >= 0) {
        selectedActiveIndex = newIndex;
        currentActiveIndex = newIndex;
      } else {
        selectedActiveIndex = 0;
        currentActiveIndex = 0;
      }
    } else {
      selectedActiveIndex = 0;
      currentActiveIndex = 0;
    }
  }

  // Re-validate whenever configData changes
  $effect(() => {
    if (configData) validateNames();
  });

  // Derived flag used to disable Apply when there are name validation errors
  let hasNameErrors = $derived(Object.values(nameErrors).some(Boolean));

  async function loadConfig() {
    try {
      configData = await GetStructuredConfigUI();
      // Initialize cache config with defaults if not present
      if (!configData.cache) {
        configData.cache = {
          ttl: '24h',
          maxSize: '100MB',
          cacheDir: '',
        };
      }
      // Trim and validate names first so sorting uses normalized names
      validateNames();

      // sort data sources and initialize expandedSources
      sortDataSources();

      // find active index from name provided by backend; default to 0 (sortDataSources already attempts to restore)
      selectedActiveIndex = (configData.dataSources || []).findIndex(
        (ds) => ds.name === configData.activeDataSourceName,
      );
      if (selectedActiveIndex < 0)
        selectedActiveIndex =
          selectedActiveIndex =
          selectedActiveIndex =
            (selectedActiveIndex = selectedActiveIndex) || selectedActiveIndex; // no-op to keep code location stable
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
            if (
              ev === '' ||
              ev === null ||
              ev === undefined ||
              isNaN(Number(ev)) ||
              Number(ev) < 0 ||
              !Number.isInteger(Number(ev))
            ) {
              applyError = `Event ID for data source "${ds.name || '#' + i}" must be a positive integer (zero or greater).`;
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

      // sort before sending so backend receives a consistent ordering (test types go to the bottom)
      sortDataSources();

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
  let showDeleteConfirm = $state(false);
  let deleteIndex = null;
  let deleteName = $state('');

  function openDeleteConfirm(i) {
    deleteIndex = i;
    deleteName = configData?.dataSources?.[i]?.name || '';
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

    // After removal, sort so test types are pushed to end and names are ordered
    validateNames();
    sortDataSources();

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

  // API tokens handlers for the token manager UI
  function handleAddApiToken(entry) {
    if (!configData) configData = {};
    if (!Array.isArray(configData.apiTokens)) configData.apiTokens = [];
    configData.apiTokens.push(entry);
  }

  function handleEditApiToken(evt) {
    const { index, entry } = evt.detail || evt;
    if (!configData || !Array.isArray(configData.apiTokens)) return;
    if (index >= 0 && index < configData.apiTokens.length) {
      configData.apiTokens[index] = entry;
    }
  }

  function handleDeleteApiToken(evt) {
    const idx = evt.detail || evt;
    if (!configData || !Array.isArray(configData.apiTokens)) return;
    if (idx >= 0 && idx < configData.apiTokens.length) {
      configData.apiTokens.splice(idx, 1);
    }
  }
</script>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  {#if loading}
    <div class="flex items-center justify-center p-4">
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"
        ></div>
        <p class="text-gray-600 dark:text-gray-400">Loading configuration...</p>
      </div>
    </div>
  {:else}
    <!-- Active Data Source -->
    <div
      class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow px-4 py-2 border border-gray-200 dark:border-gray-700"
    >
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">
        Active Data Source
      </h3>
      <div class="flex items-center gap-2">
        <label for="active-source" class="text-sm font-medium text-gray-700 dark:text-gray-300"
          >Select Data Source:</label
        >
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
      <div class="mt-0.5">
        <span class="text-sm text-green-600 dark:text-green-400"
          >Current Active: <strong>{configData.dataSources?.[currentActiveIndex]?.name}</strong
          ></span
        >
      </div>
    </div>

    <!-- Data Sources -->
    <DataSources
      {configData}
      {expandedSources}
      {nameErrors}
      {indicoDataSourcePlaceholders}
      {testDataSourcePlaceholders}
      {loading}
      {applying}
      {validateNames}
      {currentActiveIndex}
      {apiTokens}
      onAddIndico={openAddIndicoDialog}
      onDelete={(index) => openDeleteConfirm(index)}
      onToggle={(index) => toggleSource(index)}
    />

    <ConfirmDialog
      bind:open={showDeleteConfirm}
      title="Delete Data Source"
      message={`Are you sure you want to delete "${deleteName || 'this data source'}"? This action cannot be undone.`}
      confirmLabel={applying ? 'Deleting...' : 'Delete'}
      cancelLabel="Cancel"
      danger={true}
      onConfirm={confirmDelete}
      onCancel={cancelDelete}
    />

    <!-- Advanced (collapsible): groups Cache Configuration + API Tokens -->
    <div
      class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden"
    >
      <div
        role="button"
        tabindex="0"
        onclick={() => (showAdvanced = !showAdvanced)}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            showAdvanced = !showAdvanced;
          }
        }}
        class="w-full flex items-center justify-between px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      >
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Advanced</h3>
        <div class="flex items-center gap-1">
          <span class="text-sm text-gray-600 dark:text-gray-400"
            >{(configData?.apiTokens || []).length} tokens</span
          >
          <Icon icon="mdi:chevron-down" class={`w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform ${showAdvanced ? 'rotate-180' : ''}`} />
        </div>
      </div>
      {#if showAdvanced}
        <div class="p-2 space-y-2">
          <!-- Cache Configuration (moved back from CacheTab) -->
          <div
            class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden"
          >
            <div
              role="button"
              tabindex="0"
              onclick={() => (showCacheConfig = !showCacheConfig)}
              onkeydown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  showCacheConfig = !showCacheConfig;
                }
              }}
              class="w-full flex items-center justify-between p-3 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            >
              <h4 class="text-md font-semibold text-gray-900 dark:text-gray-100">
                Cache Configuration
              </h4>
              <Icon icon="mdi:chevron-down" class={`w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform ${showCacheConfig ? 'rotate-180' : ''}`} />
            </div>

            {#if showCacheConfig}
              <div class="p-4">
                <div class="pt-2">
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
                    <div>
                      <label
                        for="cache-ttl"
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >TTL (Time-To-Live)
                        <span
                          class="text-xs text-gray-500 dark:text-gray-400 ml-1"
                          title="How long cache entries stay valid before expiring">ⓘ</span
                        >
                      </label>
                      <input
                        id="cache-ttl"
                        type="text"
                        bind:value={configData.cache.ttl}
                        placeholder="24h"
                        class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      />
                      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        e.g., 24h, 1h30m, 30m
                      </p>
                    </div>
                    <div>
                      <label
                        for="cache-maxsize"
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >Max Size
                        <span
                          class="text-xs text-gray-500 dark:text-gray-400 ml-1"
                          title="Maximum cache size - oldest entries evicted when limit reached"
                          >ⓘ</span
                        >
                      </label>
                      <input
                        id="cache-maxsize"
                        type="text"
                        bind:value={configData.cache.maxSize}
                        placeholder="100MB"
                        class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      />
                      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        e.g., 100MB, 1GB, 500MB
                      </p>
                    </div>
                    <div>
                      <label
                        for="cache-dir"
                        class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >Cache Directory
                        <span
                          class="text-xs text-gray-500 dark:text-gray-400 ml-1"
                          title="Custom cache directory path (leave empty for default)">ⓘ</span
                        >
                      </label>
                      <input
                        id="cache-dir"
                        type="text"
                        bind:value={configData.cache.cacheDir}
                        placeholder="~/.cache/IndicoDataFusion"
                        class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      />
                      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        Leave empty for default
                      </p>
                    </div>
                  </div>
                </div>
                <!-- Note: Use the top-level Apply button to persist these changes -->
              </div>
            {/if}
          </div>

          <!-- API Tokens Manager -->
          <ApiTokens
            {apiTokens}
            onAdd={(entry) => handleAddApiToken(entry)}
            onEdit={(payload) => handleEditApiToken(payload)}
            onDelete={(index) => handleDeleteApiToken(index)}
          />
        </div>
      {/if}
    </div>
    <!-- CLOSE: Advanced container -->
    <div class="flex items-center justify-end gap-2">
      <!-- Toast (inline, left of Apply) -->
      {#if showToast}
        <div
          class="transform transition-all duration-300 ease-out flex items-start gap-3 rounded-lg shadow-lg overflow-hidden px-3 py-2"
          role="status"
          aria-live={toastType === 'error' ? 'assertive' : 'polite'}
        >
          <div class="mt-0.5">
            {#if toastType === 'success'}
              <Icon icon="mdi:check" class="w-5 h-5 text-green-600" />
            {:else if toastType === 'error'}
              <Icon icon="mdi:close-circle" class="w-5 h-5 text-red-600" />
            {:else}
              <Icon icon="mdi:information" class="w-5 h-5 text-gray-600" />
            {/if}
          </div>
          <div class="text-sm leading-tight">
            <div class="font-medium text-gray-900 dark:text-gray-100">{toastMessage}</div>
          </div>
          <button
            class="ml-2 text-xs text-gray-500 hover:text-gray-700"
            onclick={() => {
              showToast = false;
              if (toastTimeoutId) {
                clearTimeout(toastTimeoutId);
                toastTimeoutId = null;
              }
            }}
            aria-label="Dismiss toast">×</button
          >
        </div>
      {/if}

      <!-- Apply Button -->
      <div>
        <button
          type="button"
          class="px-2 py-1.5 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:ring-offset-2 transition-colors"
          onclick={apply}
          disabled={applying || hasNameErrors}
        >
          {applying ? 'Applying...' : 'Apply'}
        </button>
      </div>
    </div>

    <!-- Configuration File Path Info (Collapsible) -->
    <div
      class="bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
    >
      <div
        role="button"
        tabindex="0"
        onclick={() => (showConfigFile = !showConfigFile)}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            showConfigFile = !showConfigFile;
          }
        }}
        class="w-full flex items-center justify-between p-3 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      >
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Configuration File</h4>
        <svg
          class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform"
          class:rotate-180={showConfigFile}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"
          ></path>
        </svg>
      </div>
      {#if showConfigFile}
        <div class="px-4 pb-4 border-t border-gray-200 dark:border-gray-700 space-y-1">
          <div class="flex flex-wrap items-center gap-2 text-sm pt-3">
            <span class="font-medium text-gray-600 dark:text-gray-400">Path:</span>
            <span class="text-gray-800 dark:text-gray-200 font-mono text-xs break-all"
              >{configData.pathInfo?.path || configData.path || 'Not set'}</span
            >
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- IndicoConfig component for adding new Indico sources -->
<IndicoConfig
  bind:open={indicoDialogOpen}
  existingNames={(configData?.dataSources || []).map((ds) => ds.name)}
  placeholders={indicoDataSourcePlaceholders}
  saving={applying}
  {apiTokens}
  onCreate={handleCreateIndico}
  onCancel={cancelCreateIndico}
/>
