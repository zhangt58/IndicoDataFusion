<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import {
    GetCacheStats,
    GetCacheEntries,
    RefreshCache,
    ClearCache,
    DeleteCacheEntry,
    OpenCacheDirectory,
    GetStructuredConfigUI,
    ApplyStructuredConfigUI,
  } from '../../wailsjs/go/main/App';
  import { Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';

  let cacheStats = $state(null);
  /** @type {Record<string, any[]>} */
  let cacheEntries = $state({});
  let loading = $state(true);
  let refreshing = $state({});
  let errorMsg = $state('');
  let successMsg = $state('');
  let showClearConfirm = $state(false);
  let expandedDataSources = $state({});

  // Cache configuration state
  let configData = $state(null);
  let showCacheConfig = $state(false);
  let applyingConfig = $state(false);

  async function loadCacheInfo() {
    try {
      loading = true;
      cacheStats = await GetCacheStats();
      cacheEntries = await GetCacheEntries();

      // Load cache configuration
      const cfg = await GetStructuredConfigUI();
      if (!cfg.cache) {
        cfg.cache = {
          ttl: '24h',
          maxSize: '100MB',
          cacheDir: '',
        };
      }
      configData = cfg;

      // Only expand the active data source
      if (cacheEntries && typeof cacheEntries === 'object') {
        const activeDataSource = cacheStats?.data_source_name;
        expandedDataSources = {};
        Object.keys(cacheEntries).forEach((dataSourceName) => {
          // Only expand if it's the active data source
          expandedDataSources[dataSourceName] = dataSourceName === activeDataSource;
        });
        expandedDataSources = { ...expandedDataSources };
      }

      loading = false;
    } catch (e) {
      errorMsg = `Failed to load cache info: ${e}`;
      loading = false;
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    const date = new Date(timestamp);
    return date.toLocaleString();
  }

  function formatTimeAgo(timestamp) {
    if (!timestamp) return '';
    const now = new Date();
    const date = new Date(timestamp);
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  }

  onMount(async () => {
    await loadCacheInfo();

    // Listen for cache update events from backend
    EventsOn('cache:updated', () => {
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
      setTimeout(() => {
        successMsg = '';
      }, 3000);
    } catch (e) {
      errorMsg = `Failed to refresh ${key}: ${e}`;
    } finally {
      refreshing[key] = false;
    }
  }

  async function handleDeleteEntry(key) {
    errorMsg = '';
    successMsg = '';
    try {
      await DeleteCacheEntry(key);
      successMsg = `Cache entry deleted: ${key}`;
      setTimeout(() => {
        successMsg = '';
      }, 3000);
      await loadCacheInfo();
    } catch (e) {
      errorMsg = `Failed to delete ${key}: ${e}`;
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
      setTimeout(() => {
        successMsg = '';
      }, 3000);
      await loadCacheInfo();
    } catch (e) {
      errorMsg = `Failed to clear cache: ${e}`;
    }
  }

  function cancelClearCache() {
    showClearConfirm = false;
  }

  async function handleOpenCacheDirectory() {
    errorMsg = '';
    try {
      await OpenCacheDirectory();
    } catch (e) {
      errorMsg = `Failed to open cache directory: ${e}`;
    }
  }

  async function applyCacheConfig() {
    errorMsg = '';
    successMsg = '';
    applyingConfig = true;
    try {
      await ApplyStructuredConfigUI(configData);
      successMsg = 'Cache configuration applied successfully';
      setTimeout(() => {
        successMsg = '';
      }, 3000);
      // Reload cache info to reflect changes
      await loadCacheInfo();
    } catch (e) {
      errorMsg = `Failed to apply cache configuration: ${e}`;
    } finally {
      applyingConfig = false;
    }
  }

  function getCacheKeyLabel(key) {
    const labels = {
      event_info: 'Event Information',
      abstracts: 'Abstracts',
      contributions: 'Contributions',
    };
    return labels[key] || key;
  }

  // Reactive transform: compute per-entry expiry flags so template can rely on reactive fields
  let groupedEntries = $derived(
    (() => {
      const out = {};
      if (!cacheEntries || typeof cacheEntries !== 'object') return out;

      function parseRawExpiry(raw) {
        if (raw == null) return { raw: null, date: null };
        // Already a Date
        if (raw instanceof Date) return { raw, date: raw };
        // String (ISO/RFC3339)
        if (typeof raw === 'string') {
          const d = new Date(raw);
          return { raw, date: isNaN(d.getTime()) ? null : d };
        }
        // Number (seconds or milliseconds)
        if (typeof raw === 'number') {
          let ms = raw;
          // if it looks like seconds (10 digits), convert to ms
          if (raw > 0 && raw < 1e12) ms = raw * 1000;
          const d = new Date(ms);
          return { raw, date: isNaN(d.getTime()) ? null : d };
        }
        // Object: try common shapes (e.g., {seconds, nanos} or {sec, nsec} or {Time: '...'} )
        if (typeof raw === 'object') {
          try {
            if (raw.Time && typeof raw.Time === 'string') {
              const d = new Date(raw.Time);
              return { raw, date: isNaN(d.getTime()) ? null : d };
            }
            const secs = raw.seconds ?? raw.secs ?? raw.sec ?? raw.Sec ?? raw.Seconds;
            const nanos = raw.nanos ?? raw.nanoseconds ?? raw.Nanos ?? raw.Nanoseconds ?? 0;
            if (typeof secs === 'number') {
              const ms = secs * 1000 + Math.floor((nanos || 0) / 1e6);
              const d = new Date(ms);
              return { raw, date: isNaN(d.getTime()) ? null : d };
            }
            // Fallback: try to find ISO substring
            const s = JSON.stringify(raw);
            const match = s.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?/);
            if (match) {
              const d = new Date(match[0]);
              return { raw, date: isNaN(d.getTime()) ? null : d };
            }
          } catch (e) {
            // ignore
          }
        }
        return { raw, date: null };
      }

      Object.entries(cacheEntries).forEach(([dsName, entries]) => {
        // Make a shallow copy and sort by the entry.key (ascending, case-insensitive, numeric-aware)
        const sorted = [...(entries || [])].sort((a, b) => {
          const ka = a && a.key != null ? String(a.key) : '';
          const kb = b && b.key != null ? String(b.key) : '';
          return ka.localeCompare(kb, undefined, { numeric: true, sensitivity: 'base' });
        });

        out[dsName] = sorted.map((entry) => {
          // Accept both camelCase and snake_case serialization from backend
          const rawExpiryCandidate =
            entry && (entry.expiresAt ?? entry.expires_at ?? entry.expires_at_string ?? null);
          const { date: expiryDate } = parseRawExpiry(rawExpiryCandidate);
          const isZeroTime = expiryDate && expiryDate.getFullYear && expiryDate.getFullYear() === 1;
          // Coerce to booleans so template sees true/false instead of null
          const hasExpiry = !!(expiryDate && !isNaN(expiryDate.getTime()) && !isZeroTime);
          const isExpired = !!(hasExpiry && expiryDate.getTime() < Date.now());
          // Normalize expiresAt on returned entry so template formatTimestamp(entry.expiresAt) works
          const normalizedExpiresAt = expiryDate ? expiryDate.toISOString() : null;
          return {
            ...entry,
            expiresAt: normalizedExpiresAt,
            hasExpiry,
            isExpired,
            __expiryDate: expiryDate,
          };
        });
      });

      // Temporary debug: show a compact sample of computed flags to diagnose null/undefined
      try {
        const sample = Object.fromEntries(
          Object.entries(out).map(([ds, entries]) => [
            ds,
            entries.map((e) => ({
              key: e.key,
              rawExpiry: e.expiresAt,
              hasExpiry: e.hasExpiry,
              isExpired: e.isExpired,
            })),
          ]),
        );
        console.debug('groupedEntries sample:', sample);
      } catch (err) {
        // ignore serialization errors
      }

      return out;
    })(),
  );

  // Reactive: sorted array of [dataSourceName, entries] sorted by data source name
  let groupedEntriesList = $derived(
    Object.entries(groupedEntries).sort((a, b) => {
      const ka = a && a[0] != null ? String(a[0]) : '';
      const kb = b && b[0] != null ? String(b[0]) : '';
      return ka.localeCompare(kb, undefined, { numeric: true, sensitivity: 'base' });
    }),
  );
</script>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  <div class="mb-2">
    <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Cache Management</h2>
  </div>

  {#if loading}
    <div class="flex items-center justify-center p-2">
      <div class="text-center">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"
        ></div>
        <p class="text-gray-600 dark:text-gray-400">Loading cache information...</p>
      </div>
    </div>
  {:else}
    <!-- Cache Statistics -->
    <div
      class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700"
    >
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Cache Statistics</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-1">
        <div class="bg-gray-200 dark:bg-gray-700 rounded-md p-2">
          <div class="text-sm text-gray-500 dark:text-gray-400">Status</div>
          <div
            class="text-xl font-bold {cacheStats?.enabled
              ? 'text-green-600 dark:text-green-400'
              : 'text-red-600 dark:text-red-400'}"
          >
            {cacheStats?.enabled ? 'Enabled' : 'Disabled'}
          </div>
        </div>
        <div class="bg-gray-200 dark:bg-gray-700 rounded-md p-2">
          <div class="text-sm text-gray-500 dark:text-gray-400">Entries</div>
          <div class="text-xl font-bold text-indigo-600 dark:text-indigo-400">
            {cacheStats?.entries || 0}
          </div>
        </div>
        <div class="bg-gray-200 dark:bg-gray-700 rounded-md p-2">
          <div class="text-sm text-gray-500 dark:text-gray-400">Size</div>
          <div class="text-lg font-bold text-blue-600 dark:text-blue-400">
            {cacheStats?.current_size_mb || '0 MB'}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            Max: {cacheStats?.max_size_mb || 'N/A'}
          </div>
        </div>
        <div class="bg-gray-200 dark:bg-gray-700 rounded-md p-2">
          <div class="text-sm text-gray-500 dark:text-gray-400">TTL</div>
          <div class="text-lg font-bold text-purple-600 dark:text-purple-400">
            {cacheStats?.ttl || 'N/A'}
          </div>
        </div>
      </div>
      <div
        class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-700 grid grid-cols-1 md:grid-cols-2 gap-2"
      >
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
            <div class="text-xs font-mono text-gray-800 dark:text-gray-200 break-all mt-1">
              {cacheStats.cache_dir}
            </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Cached Data Entries (Grouped by Data Source) -->
    {#if cacheEntries && Object.keys(cacheEntries).length > 0}
      <div class="space-y-1">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Cached Data</h3>
        </div>

        {#each groupedEntriesList as [dataSourceName, entries]}
          <div
            class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden"
          >
            <!-- Data Source Header -->
            <button
              type="button"
              onclick={() => {
                expandedDataSources[dataSourceName] = !expandedDataSources[dataSourceName];
                expandedDataSources = { ...expandedDataSources };
              }}
              class="w-full flex items-center justify-between p-2 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              <span class="inline-flex items-center gap-2">
                <span class="text-base font-semibold text-gray-900 dark:text-gray-100"
                  >{dataSourceName}</span
                >
                <span
                  class="px-2 py-0.5 text-xs rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200"
                >
                  {entries.length}
                  {entries.length === 1 ? 'entry' : 'entries'}
                </span>
              </span>
              <svg
                class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedDataSources[
                  dataSourceName
                ]
                  ? 'rotate-180'
                  : ''}"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                ></path>
              </svg>
            </button>

            <!-- Data Source Content -->
            {#if expandedDataSources[dataSourceName]}
              <div class="border-t border-gray-200 dark:border-gray-700">
                <div class="divide-y divide-gray-200 dark:divide-gray-700">
                  {#each entries as entry (entry.key)}
                    <div
                      class="p-2 {entry.isExpired
                        ? 'bg-red-50/50 dark:bg-red-900/10'
                        : ''} hover:bg-gray-200 dark:hover:bg-gray-700"
                    >
                      <div class="flex items-start justify-between">
                        <div class="flex items-start gap-2 flex-1">
                          <div
                            class="w-2 h-2 {entry.isExpired
                              ? 'bg-red-500'
                              : 'bg-green-500'} rounded-full mt-2"
                          ></div>
                          <div class="flex-1 min-w-0">
                            <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                              {getCacheKeyLabel(entry.key)}
                            </div>
                            <div class="mt-1 space-y-0.5">
                              <div class="text-xs text-gray-600 dark:text-gray-400">
                                <span class="font-normal">Last Updated:</span>
                                {formatTimestamp(entry.timestamp)}
                                <span class="ml-2 text-gray-500 dark:text-gray-500"
                                  >({formatTimeAgo(entry.timestamp)})</span
                                >
                              </div>
                              {#if entry.hasExpiry}
                                {#if entry.isExpired}
                                  <div class="text-xs text-red-600 dark:text-red-400">
                                    <span
                                      class="px-2 py-0.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded text-xs font-semibold"
                                      >EXPIRED</span
                                    >
                                  </div>
                                {:else}
                                  <div class="text-xs text-gray-600 dark:text-gray-400">
                                    <span class="font-normal">Expires:</span>
                                    {formatTimestamp(entry.expiresAt)}
                                  </div>
                                {/if}
                              {/if}
                            </div>
                          </div>
                        </div>
                        <div class="flex gap-2 ml-4">
                          <button
                            type="button"
                            onclick={() => handleRefresh(entry.key)}
                            disabled={refreshing[entry.key]}
                            class="px-2 py-1 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
                            title="Refresh from API"
                          >
                            {refreshing[entry.key] ? 'Refreshing...' : 'Refresh'}
                          </button>
                          <button
                            type="button"
                            onclick={() => handleDeleteEntry(entry.key)}
                            class="px-2 py-1 rounded bg-red-600 text-white text-sm hover:bg-red-700 transition-colors whitespace-nowrap"
                            title="Delete this cache entry"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {:else}
      <div
        class="bg-gray-50 dark:bg-gray-800 rounded-lg p-8 text-center border border-gray-200 dark:border-gray-700"
      >
        <Icon
          icon="mdi:database-off"
          class="w-16 h-16 mx-auto text-gray-400 dark:text-gray-500 mb-4"
        />
        <p class="text-gray-600 dark:text-gray-400 text-lg">No cached data available</p>
        <p class="text-gray-500 dark:text-gray-500 text-sm mt-2">
          Data will be cached when you fetch it from the API
        </p>
      </div>
    {/if}

    <!-- Actions -->
    <div
      class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-2 border border-gray-200 dark:border-gray-700"
    >
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Actions</h3>
      <div class="flex gap-2">
        <button
          type="button"
          onclick={handleOpenCacheDirectory}
          disabled={!cacheStats?.cache_dir}
          class="flex-1 px-3 py-2 rounded-lg bg-gray-600 text-white font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
          title="Open cache directory in file browser"
        >
          <Icon icon="mdi:folder-open" class="w-5 h-5" />
          Open Directory
        </button>
        <button
          type="button"
          onclick={handleClearAll}
          class="flex-1 px-3 py-2 rounded-lg bg-red-600 text-white font-medium hover:bg-red-700 transition-colors flex items-center justify-center gap-2"
        >
          <Icon icon="mdi:trash-can" class="w-5 h-5" />
          Clear All Cache
        </button>
      </div>
    </div>

    <!-- Messages -->
    {#if errorMsg}
      <div
        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3"
      >
        <p class="text-sm text-red-600 dark:text-red-400">{errorMsg}</p>
      </div>
    {/if}
    {#if successMsg}
      <div
        class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-3"
      >
        <p class="text-sm text-green-600 dark:text-green-400">{successMsg}</p>
      </div>
    {/if}

    <!-- Cache Configuration (collapsible) -->
    {#if configData?.cache}
      <div
        class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <button
          type="button"
          onclick={() => (showCacheConfig = !showCacheConfig)}
          class="w-full flex items-center justify-between p-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          <span class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Cache Configuration
          </span>
          <Icon
            icon="mdi:chevron-down"
            class={`w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform ${showCacheConfig ? 'rotate-180' : ''}`}
          />
        </button>

        {#if showCacheConfig}
          <div class="border-t border-gray-200 dark:border-gray-700 p-3">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-3">
              <div>
                <label
                  for="cache-ttl"
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  TTL (Time-To-Live)
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
                  class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 24h, 1h30m, 30m</p>
              </div>
              <div>
                <label
                  for="cache-maxsize"
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  Max Size
                  <span
                    class="text-xs text-gray-500 dark:text-gray-400 ml-1"
                    title="Maximum cache size - oldest entries evicted when limit reached">ⓘ</span
                  >
                </label>
                <input
                  id="cache-maxsize"
                  type="text"
                  bind:value={configData.cache.maxSize}
                  placeholder="100MB"
                  class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 100MB, 1GB, 500MB</p>
              </div>
              <div>
                <label
                  for="cache-dir"
                  class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  Cache Directory
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
                  class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1.5 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Leave empty for default</p>
              </div>
            </div>
            <div class="flex justify-end">
              <button
                type="button"
                onclick={applyCacheConfig}
                disabled={applyingConfig}
                class="px-3 py-1.5 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {applyingConfig ? 'Applying...' : 'Apply'}
              </button>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<!-- Confirmation Modal for Clear Cache -->
<Modal bind:open={showClearConfirm} size="sm" autoclose={false}>
  <div class="text-center">
    <Icon
      icon="mdi:alert-circle"
      class="mx-auto mb-4 text-gray-400 w-12 h-12 dark:text-gray-200"
      aria-hidden="true"
    />
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
      Are you sure you want to clear all cached data? This will remove the cache file.
    </h3>
    <div class="flex justify-center gap-2">
      <button
        type="button"
        onclick={confirmClearCache}
        class="text-white bg-red-600 hover:bg-red-700 focus:ring-4 focus:outline-none focus:ring-red-300 dark:focus:ring-red-800 font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center"
      >
        Yes, clear it
      </button>
      <button
        type="button"
        onclick={cancelClearCache}
        class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2.5 hover:text-gray-900 focus:z-10 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
      >
        No, cancel
      </button>
    </div>
  </div>
</Modal>
