<script>
  import { onMount } from 'svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';

  let configData = null;
  let loading = true;
  let applying = false;
  let applyError = '';
  let applySuccess = '';
  // expandedSources keyed by data-source index to avoid problems when renaming
  let expandedSources = {};
  // track active selection by index so we can rename sources safely
  let currentActiveIndex = 0;
  let selectedActiveIndex = 0;
  let showConfigFile = false;

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
      // Ensure backend activeDataSourceName is set from the currently selected index
      if (configData && configData.dataSources && configData.dataSources[selectedActiveIndex]) {
        configData.activeDataSourceName = configData.dataSources[selectedActiveIndex].name;
      }
      await ApplyStructuredConfigUI(configData);
      currentActiveIndex = selectedActiveIndex;
      applySuccess = 'Configuration applied successfully';
    } catch (e) {
      applyError = `Failed to apply configuration: ${e}`;
    } finally {
      applying = false;
    }
  }
</script>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  {#if loading}
    <div class="flex items-center justify-center p-4">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading configuration...</p>
      </div>
    </div>
  {:else}
    <!-- Active Data Source Selector -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Active Data Source</h3>
      <div class="flex items-center gap-2">
        <label for="active-source" class="text-sm font-medium text-gray-700 dark:text-gray-300">Select Data Source:</label>
        <select
          id="active-source"
          bind:value={selectedActiveIndex}
          class="flex-1 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Data Sources</h3>
      <div class="space-y-2">

        {#each configData.dataSources as dataSource, i (i)}
        <div class="bg-gray-50 dark:bg-gray-750 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
          <!-- Header -->
          <div
            role="button"
            tabindex="0"
            on:click={() => toggleSource(i)}
            on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleSource(i); } }}
            class="w-full flex items-center justify-between p-2 hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
          >
            <span class="flex items-center gap-3">
              <!-- Make the name editable -->
              <input
                id={`ds-name-${i}`}
                type="text"
                bind:value={dataSource.name}
                placeholder="Data source name"
                title="Edit data source name"
                aria-label={`Data source name ${i}`}
                class="text-xl md:text-lg font-semibold text-gray-900 dark:text-gray-100 bg-transparent border-b-2 border-transparent focus:border-indigo-500 px-1 py-0.5 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-sm transition-colors placeholder-gray-400 cursor-text"
              />
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
            <svg class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedSources[i] ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
            </svg>
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
                      placeholder="https://indico.example.org"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-eventId`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label>
                    <input
                      id={`ds-${i}-eventId`}
                      type="number"
                      bind:value={dataSource.indico.eventId}
                      placeholder="123"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div class="md:col-span-2">
                    <label for={`ds-${i}-apiToken`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label>
                    <input
                      id={`ds-${i}-apiToken`}
                      type="password"
                      bind:value={dataSource.indico.apiToken}
                      placeholder="indp_..."
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-timeout`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
                    <input
                      id={`ds-${i}-timeout`}
                      type="text"
                      bind:value={dataSource.indico.timeout}
                      placeholder="15s"
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
                      placeholder="./testdata"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-eventInfo`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event Info File</label>
                    <input
                      id={`ds-${i}-test-eventInfo`}
                      type="text"
                      bind:value={dataSource.test.eventInfo}
                      placeholder="info.json"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-abstracts`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Abstracts File</label>
                    <input
                      id={`ds-${i}-test-abstracts`}
                      type="text"
                      bind:value={dataSource.test.abstracts}
                      placeholder="abstracts.json"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label for={`ds-${i}-test-contribs`} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Contributions File</label>
                    <input
                      id={`ds-${i}-test-contribs`}
                      type="text"
                      bind:value={dataSource.test.contribs}
                      placeholder="contribs.json"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                </div>
              {/if}
            </div>
          {/if}
        </div>
        {/each}
      </div>
    </div>

    <!-- Cache Configuration -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Cache Configuration</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
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
        disabled={applying}
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
        class="w-full flex items-center justify-between p-3 hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
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
