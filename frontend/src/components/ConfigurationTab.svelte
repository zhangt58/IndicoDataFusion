<script>
  import { onMount } from 'svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';

  let configData = null;
  let loading = true;
  let applying = false;
  let applyError = '';
  let applySuccess = '';
  let expandedSources = {};
  let currentActive = '';
  let selectedActive = '';

  async function loadConfig() {
    try {
      configData = await GetStructuredConfigUI();
      // Initialize all sources as collapsed
      configData.dataSources.forEach(ds => {
        expandedSources[ds.name] = false;
      });
      selectedActive = configData.activeDataSourceName;
      currentActive = configData.activeDataSourceName;
      loading = false;
    } catch (e) {
      loading = false;
      applyError = `Failed to load config: ${e}`;
    }
  }

  onMount(loadConfig);

  function toggleSource(name) {
    expandedSources[name] = !expandedSources[name];
  }

  async function apply() {
    applyError = '';
    applySuccess = '';
    applying = true;
    try {
      configData.activeDataSourceName = selectedActive;
      await ApplyStructuredConfigUI(configData);
      currentActive = selectedActive;
      applySuccess = 'Configuration applied successfully';
    } catch (e) {
      applyError = `Failed to apply configuration: ${e}`;
    } finally {
      applying = false;
    }
  }

  function getDataSourceNames() {
    return configData?.dataSources.map(ds => ds.name) || [];
  }
</script>

<div class="p-4 space-y-4 max-w-5xl mx-auto">
  {#if loading}
    <div class="flex items-center justify-center p-8">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading configuration...</p>
      </div>
    </div>
  {:else}
    <!-- Active Data Source Selector -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Active Data Source</h3>
      <div class="flex items-center gap-3">
        <label for="active-source" class="text-sm font-medium text-gray-700 dark:text-gray-300">Select Data Source:</label>
        <select
          id="active-source"
          bind:value={selectedActive}
          class="flex-1 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {#each getDataSourceNames() as sourceName}
            <option value={sourceName}>{sourceName}</option>
          {/each}
        </select>
      </div>
      <div class="mt-2">
        <span class="text-sm text-green-600 dark:text-green-400">Current Active: <strong>{currentActive}</strong></span>
      </div>
    </div>

    <!-- Data Sources -->
    <div class="space-y-3">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Data Sources</h3>

      {#each configData.dataSources as dataSource (dataSource.name)}
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden">
          <!-- Header -->
          <button
            type="button"
            on:click={() => toggleSource(dataSource.name)}
            class="w-full flex items-center justify-between p-4 hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
          >
            <div class="flex items-center gap-3">
              <span class="text-base font-semibold text-gray-900 dark:text-gray-100">{dataSource.name}</span>
              <span class="px-2 py-0.5 text-xs rounded-full {dataSource.type === 'indico' ? 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200' : 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200'}">
                {dataSource.type === 'indico' ? 'API' : 'Test Data'}
              </span>
              {#if currentActive === dataSource.name}
                <span class="px-2 py-0.5 text-xs rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200">Active</span>
              {/if}
            </div>
            <svg class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedSources[dataSource.name] ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
            </svg>
          </button>

          <!-- Content -->
          {#if expandedSources[dataSource.name]}
            <div class="px-4 pb-4 pt-2 border-t border-gray-200 dark:border-gray-700 space-y-3">
              {#if dataSource.type === 'indico' && dataSource.indico}
                <!-- Indico API Configuration -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL</label>
                    <input
                      type="text"
                      bind:value={dataSource.indico.baseUrl}
                      placeholder="https://indico.example.org"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event ID</label>
                    <input
                      type="number"
                      bind:value={dataSource.indico.eventId}
                      placeholder="123"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div class="md:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Token</label>
                    <input
                      type="password"
                      bind:value={dataSource.indico.apiToken}
                      placeholder="indp_..."
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
                    <input
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
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="md:col-span-2">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Data Directory</label>
                    <input
                      type="text"
                      bind:value={dataSource.test.dataDir}
                      placeholder="./testdata"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event Info File</label>
                    <input
                      type="text"
                      bind:value={dataSource.test.eventInfo}
                      placeholder="info.json"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Abstracts File</label>
                    <input
                      type="text"
                      bind:value={dataSource.test.abstracts}
                      placeholder="abstracts.json"
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Contributions File</label>
                    <input
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

    <!-- Status Messages -->
    {#if applyError}
      <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
        <p class="text-sm text-red-600 dark:text-red-400">{applyError}</p>
      </div>
    {/if}
    {#if applySuccess}
      <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-3">
        <p class="text-sm text-green-600 dark:text-green-400">{applySuccess}</p>
      </div>
    {/if}

    <!-- Apply Button -->
    <div class="flex justify-end">
      <button
        type="button"
        class="px-6 py-2 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:ring-offset-2 transition-colors"
        on:click={apply}
        disabled={applying}
      >
        {applying ? 'Applying...' : 'Apply Configuration'}
      </button>
    </div>

    <!-- Configuration File Path Info -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Configuration File</h4>
      <div class="space-y-1">
        <div class="flex flex-wrap items-center gap-2 text-sm">
          <span class="font-medium text-gray-600 dark:text-gray-400">Path:</span>
          <span class="text-gray-800 dark:text-gray-200 font-mono text-xs break-all">{configData.pathInfo?.path}</span>
        </div>
        <div class="flex items-center gap-2 text-xs">
          <span class="text-gray-500 dark:text-gray-400">Source:</span>
          <span class="px-2 py-0.5 rounded-full {configData.pathInfo?.fromEnv ? 'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}">
            {configData.pathInfo?.fromEnv ? `Environment (${configData.pathInfo?.envVarName})` : 'Default Location'}
          </span>
        </div>
      </div>
    </div>
  {/if}
</div>
