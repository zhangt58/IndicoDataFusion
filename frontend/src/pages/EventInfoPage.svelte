<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetEventInfo, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { formatDate } from '../utils/dateUtils.js';
  import { createCachePage } from '../utils/cacheUtils.js';
  import { RefreshOutline } from 'flowbite-svelte-icons';
  import LoadErrorHint from './LoadErrorHint.svelte';

  let loading = $state(false);
  let refreshing = $state(false);
  let error = $state(null);
  let errorString = $state(null);
  let isTestMode = $state(false);
  let cacheExpired = $state(false);

  let eventInfo = $state(null);
  let currentDataSource = null;

  async function loadData() {
    loading = true;
    error = null;
    try {
      eventInfo = await GetEventInfo();
    } catch (e) {
      console.error('GetEventInfo failed', e);
      error = e;
      errorString = 'Failed to load event information.';
    } finally {
      loading = false;
    }
  }

  const { handleRefresh, handleCacheEvent } = createCachePage(
    'event_info',
    loadData,
    (v) => {
      refreshing = v;
    },
    (err) => {
      error = err;
    },
  );

  onMount(async () => {
    try {
      isTestMode = await IsTestMode();
    } catch (e) {
      console.error('Failed to check test mode', e);
    }

    await loadData();

    // Get current data source name from cache stats so we can ignore cache events from other data sources
    try {
      const stats = await GetCacheStats();
      currentDataSource = stats?.data_source_name || null;
    } catch (e) {
      console.warn('Failed to get cache stats for data source name', e);
      currentDataSource = null;
    }

    EventsOn('cache:updated', (...data) => {
      const ev = (data && data.length ? data[0] : data) || {};

      // If the event includes a data_source_name and it doesn't match our current data source, ignore it
      if (ev.data_source_name && currentDataSource && ev.data_source_name !== currentDataSource) {
        return;
      }

      // Handle expired notification from backend goroutine
      if (ev.action === 'expired' && ev.key === 'event_info') {
        cacheExpired = true;
        return;
      }

      // Handle refresh/delete/clear actions
      if (ev.action === 'refreshed' && ev.key === 'event_info') {
        cacheExpired = false;
      }

      handleCacheEvent(ev);
    });
  });

  onDestroy(() => {
    EventsOff('cache:updated');
  });

  // Wrapper to handle 'N/A' for empty dates in EventInfo
  function formatEventDate(dateInfo) {
    return dateInfo ? formatDate(dateInfo) : 'N/A';
  }
</script>

{#if loading}
  <div class="p-6 text-center">Loading event information...</div>
{:else if error}
  <LoadErrorHint {error} message={errorString} title="Failed to load event information" />
{:else if eventInfo}
  <div class="max-w-4xl mx-auto mt-8">
    <!-- Event Header -->
    <div class="bg-linear-to-r from-indigo-500 to-purple-600 rounded-t-lg p-6 text-white">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <span class="px-3 py-1 bg-white/20 rounded-full text-sm font-medium">
            {eventInfo.category || 'Conference'}
          </span>
          <span class="text-sm opacity-80">ID: {eventInfo.id}</span>
        </div>
        {#if !isTestMode}
          <div class="relative">
            <button
              onclick={() => handleRefresh()}
              disabled={refreshing}
              class="p-2 rounded-lg bg-white/20 hover:bg-white/30 transition-colors disabled:opacity-50"
              title={cacheExpired ? 'Cache expired - Click to refresh' : 'Refresh from API'}
            >
              <RefreshOutline class={`shrink-0 h-6 w-6 ${refreshing ? 'animate-spin' : ''}`} />
            </button>
            {#if cacheExpired && !refreshing}
              <span class="absolute -top-1 -right-1 flex h-3 w-3">
                <span
                  class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"
                ></span>
                <span
                  class="relative inline-flex rounded-full h-3 w-3 bg-red-500"
                  title="Cache expired"
                ></span>
              </span>
            {/if}
          </div>
        {/if}
      </div>
      <h1 class="text-2xl md:text-3xl font-bold">{eventInfo.title}</h1>
    </div>

    <!-- Event Details Card -->
    <div
      class="bg-white dark:bg-gray-800 rounded-b-lg shadow-lg border border-gray-200 dark:border-gray-700"
    >
      <!-- Date and Location -->
      <div class="p-6 border-b border-gray-200 dark:border-gray-700">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Dates -->
          <div class="flex items-start gap-3">
            <div class="p-2 bg-indigo-100 dark:bg-indigo-900 rounded-lg">
              <svg
                class="w-6 h-6 text-indigo-600 dark:text-indigo-300"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
            </div>
            <div>
              <p class="text-sm font-semibold text-gray-600 dark:text-gray-400">Date</p>
              <p class="text-gray-800 dark:text-gray-200">{formatEventDate(eventInfo.startDate)}</p>
              <p class="text-gray-600 dark:text-gray-400 text-sm">
                to {formatEventDate(eventInfo.endDate)}
              </p>
            </div>
          </div>

          <!-- Location -->
          <div class="flex items-start gap-3">
            <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
              <svg
                class="w-6 h-6 text-green-600 dark:text-green-300"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </div>
            <div>
              <p class="text-sm font-semibold text-gray-600 dark:text-gray-400">Location</p>
              <p class="text-gray-800 dark:text-gray-200">{eventInfo.location}</p>
              {#if eventInfo.address}
                <p class="text-gray-600 dark:text-gray-400 text-sm">{eventInfo.address}</p>
              {/if}
            </div>
          </div>
        </div>
      </div>

      <!-- Description -->
      {#if eventInfo.description}
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
            About the Event
          </h2>
          <div class="prose dark:prose-invert max-w-none text-gray-700 dark:text-gray-300">
            {@html eventInfo.description}
          </div>
        </div>
      {/if}
    </div>
  </div>
{:else}
  <div class="p-6 text-center text-gray-500">No event information available.</div>
{/if}

<style>
  /* Style for HTML content from description */
  :global(.prose a) {
    color: #4f46e5;
    text-decoration: underline;
  }
  :global(.prose a:hover) {
    color: #4338ca;
  }
  :global(.dark .prose a) {
    color: #818cf8;
  }
  :global(.dark .prose a:hover) {
    color: #a5b4fc;
  }
</style>
