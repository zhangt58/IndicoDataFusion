<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { GetCacheStats, GetCacheKeys, RefreshCache, ClearCache, IsTestMode } from '../../wailsjs/go/main/App';
  import { Modal } from 'flowbite-svelte';

  let cacheStats = null;
  let cacheKeys = [];
  let loading = true;
  let refreshing = {};
  let errorMsg = '';
  let successMsg = '';
  let isTestMode = false;
  let showClearConfirm = false;

  async function loadCacheInfo() {
    try {
      loading = true;
      cacheStats = await GetCacheStats();
      cacheKeys = await GetCacheKeys();
      loading = false;
    } catch (e) {
      errorMsg = `Failed to load cache info: ${e}`;
      loading = false;
    }
  }

  onMount(async () => {
    // Check if in test mode
    try {
      isTestMode = await IsTestMode();
    } catch (e) {
      console.error('Failed to check test mode', e);
    }

    loadCacheInfo();

    // Listen for cache update events from backend
    EventsOn('cache:updated', (data) => {
      console.log('Cache updated:', data);
      successMsg = `Cache ${data.action}: ${data.key || 'all'}`;
      setTimeout(() => { successMsg = ''; }, 3000);
      loadCacheInfo();
    });
  });

  onDestroy(() => {
    EventsOff('cache:updated');
  });

  async function handleRefresh(key) {
    errorMsg = '';
    successMsg = '';
    refreshing[key] = true;
    try {
      await RefreshCache(key);
      successMsg = `Cache refreshed: ${key}`;
      setTimeout(() => { successMsg = ''; }, 3000);
    } catch (e) {
      errorMsg = `Failed to refresh ${key}: ${e}`;
    } finally {
      refreshing[key] = false;
    }
  }

  function handleClearAll() {
    showClearConfirm = true;
  }

  async function confirmClearCache() {
    showClearConfirm = false;
    errorMsg = '';
    successMsg = '';
    try {
      await ClearCache();
      successMsg = 'All cache cleared and file removed';
      setTimeout(() => { successMsg = ''; }, 3000);
      await loadCacheInfo();
    } catch (e) {
      errorMsg = `Failed to clear cache: ${e}`;
    }
  }

  function cancelClearCache() {
    showClearConfirm = false;
  }

  function getCacheKeyLabel(key) {
    const labels = {
      'event_info': 'Event Information',
      'abstracts': 'Abstracts',
      'contributions': 'Contributions'
    };
    return labels[key] || key;
  }
</script>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  <div class="flex items-center justify-between mb-2">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Cache Management</h2>
    <button
      type="button"
      on:click={loadCacheInfo}
      class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
    </button>
  </div>

  {#if loading}
    <div class="flex items-center justify-center p-2">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading cache information...</p>
      </div>
    </div>
  {:else}
    <!-- Cache Statistics -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Cache Statistics</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2">
        <div class="bg-gray-50 dark:bg-gray-750 rounded-lg p-3">
          <div class="text-sm text-gray-500 dark:text-gray-400">Status</div>
          <div class="text-2xl font-bold {cacheStats?.enabled ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
            {cacheStats?.enabled ? 'Enabled' : 'Disabled'}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-750 rounded-lg p-3">
          <div class="text-sm text-gray-500 dark:text-gray-400">Cached Entries</div>
          <div class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
            {cacheStats?.entries || 0}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-750 rounded-lg p-3">
          <div class="text-sm text-gray-500 dark:text-gray-400">Cache Size</div>
          <div class="text-lg font-bold text-blue-600 dark:text-blue-400">
            {cacheStats?.current_size_mb || '0 MB'}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            Max: {cacheStats?.max_size_mb || 'N/A'}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-750 rounded-lg p-3">
          <div class="text-sm text-gray-500 dark:text-gray-400">TTL</div>
          <div class="text-lg font-bold text-purple-600 dark:text-purple-400">
            {cacheStats?.ttl || 'N/A'}
          </div>
        </div>
      </div>
      <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700 grid grid-cols-1 md:grid-cols-2 gap-3">
        {#if cacheStats?.data_source_name}
          <div>
            <div class="text-sm text-gray-500 dark:text-gray-400">Data Source</div>
            <div class="text-sm font-semibold text-gray-800 dark:text-gray-200 mt-1">
              {cacheStats.data_source_name}
            </div>
          </div>
        {/if}
        {#if cacheStats?.cache_dir}
          <div>
            <div class="text-sm text-gray-500 dark:text-gray-400">Cache Directory</div>
            <div class="text-sm font-mono text-gray-800 dark:text-gray-200 break-all mt-1">
              {cacheStats.cache_dir}
            </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Cached Data Entries -->
    {#if cacheKeys.length > 0}
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700">
        <div class="p-2 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Cached Data</h3>
        </div>
        <div class="divide-y divide-gray-200 dark:divide-gray-700">
          {#each cacheKeys as key}
            <div class="p-2 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-750">
              <div class="flex items-center gap-3">
                <div class="w-2 h-2 bg-green-500 rounded-full"></div>
                <div>
                  <div class="font-medium text-gray-900 dark:text-gray-100">
                    {getCacheKeyLabel(key)}
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400 font-mono">
                    {key}
                  </div>
                </div>
              </div>
              {#if !isTestMode}
                <div class="flex gap-2">
                  <button
                    type="button"
                    on:click={() => handleRefresh(key)}
                    disabled={refreshing[key]}
                    class="px-3 py-1.5 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    {refreshing[key] ? 'Refreshing...' : 'Refresh'}
                  </button>
                </div>
              {:else}
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  (Test data - no refresh needed)
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center border border-gray-200 dark:border-gray-700">
        <svg class="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
        </svg>
        <p class="text-gray-600 dark:text-gray-400 text-lg">No cached data available</p>
        <p class="text-gray-500 dark:text-gray-500 text-sm mt-2">Data will be cached when you fetch it from the API</p>
      </div>
    {/if}

    <!-- Actions -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">Actions</h3>
      <div class="space-y-2">
        <button
          type="button"
          on:click={handleClearAll}
          class="w-full px-4 py-3 rounded-lg bg-red-600 text-white font-medium hover:bg-red-700 transition-colors flex items-center justify-center gap-2"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          Clear All Cache
        </button>
      </div>
    </div>

    <!-- Messages -->
    {#if errorMsg}
      <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
        <p class="text-sm text-red-600 dark:text-red-400">{errorMsg}</p>
      </div>
    {/if}
    {#if successMsg}
      <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-3">
        <p class="text-sm text-green-600 dark:text-green-400">{successMsg}</p>
      </div>
    {/if}

    <!-- Help Information -->
    <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-2">
      <h4 class="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-2">About Cache</h4>
      <div class="text-sm text-blue-800 dark:text-blue-300 space-y-1">
        <p><strong>Refresh:</strong> Force updates the cache by fetching fresh data from the API immediately.</p>
        <p><strong>Clear All:</strong> Removes the cache file completely. Cache will be rebuilt on next data access.</p>
        <p><strong>TTL (Time-To-Live):</strong> Cache entries automatically expire after the configured duration.</p>
        <p><strong>Size Limit:</strong> Oldest entries are automatically evicted when the cache reaches the size limit.</p>
        <p><strong>Data Source:</strong> Cache keys include the data source name to avoid conflicts.</p>
        <p><strong>Auto-Save:</strong> Cache is saved to disk after refreshes and on app shutdown.</p>
        <p><strong>Location:</strong> Cache files are stored in <code>~/.cache/IndicoDataFusion/</code></p>
      </div>
    </div>
  {/if}
</div>

<!-- Confirmation Modal for Clear Cache -->
<Modal bind:open={showClearConfirm} size="sm" autoclose={false}>
  <div class="text-center">
    <svg class="mx-auto mb-4 text-gray-400 w-12 h-12 dark:text-gray-200" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 20">
      <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 11V6m0 8h.01M19 10a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/>
    </svg>
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
      Are you sure you want to clear all cached data? This will remove the cache file.
    </h3>
    <div class="flex justify-center gap-2">
      <button
        type="button"
        on:click={confirmClearCache}
        class="text-white bg-red-600 hover:bg-red-700 focus:ring-4 focus:outline-none focus:ring-red-300 dark:focus:ring-red-800 font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center"
      >
        Yes, clear it
      </button>
      <button
        type="button"
        on:click={cancelClearCache}
        class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2.5 hover:text-gray-900 focus:z-10 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
      >
        No, cancel
      </button>
    </div>
  </div>
</Modal>

<style>
  .dark .bg-gray-750 {
    background-color: #2d3748;
  }
  .dark .hover\:bg-gray-750:hover {
    background-color: #2d3748;
  }
</style>

