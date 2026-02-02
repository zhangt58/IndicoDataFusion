<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetEventInfo, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { formatDate } from '../utils/dateUtils.js';
  import { createCachePage } from '../utils/cacheUtils.js';
  import { RefreshOutline } from 'flowbite-svelte-icons';
  import LoadErrorHint from './LoadErrorHint.svelte';
  import AttachmentGrid from '../components/AttachmentGrid.svelte';
  import Icon from '@iconify/svelte';

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
  <div class="flex items-center justify-center py-12">
    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-400"></div>
    <span class="ml-3 text-2xl text-gray-600 dark:text-gray-400">Loading...</span>
  </div>
{:else if error}
  <LoadErrorHint {error} message={errorString} title="Failed to load event information" />
{:else if eventInfo}
  <div class="max-w-full mx-auto mt-8">
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
      class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 mb-1"
      style={eventInfo.folders && eventInfo.folders.length > 0 ? 'max-height: 40vh; overflow-y: auto;' : 'max-height: 65vh; overflow-y: auto;'}
    >
      <!-- Date and Location -->
      <div class="p-6 border-b border-gray-200 dark:border-gray-700">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Dates -->
          <div class="flex items-start gap-3">
            <div class="p-2 bg-indigo-100 dark:bg-indigo-900 rounded-lg">
              <Icon icon="mdi:calendar-month" class="w-6 h-6 text-indigo-600 dark:text-indigo-300" />
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
              <Icon icon="mdi:map-marker" class="w-6 h-6 text-green-600 dark:text-green-300" />
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
        <div class="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-2">
            About the Event
          </h2>
          <div class="prose dark:prose-invert max-w-none text-gray-700 dark:text-gray-300">
            {@html eventInfo.description}
          </div>
        </div>
      {/if}

    </div>

    <!-- Materials & Attachments -->
    {#if eventInfo.folders && eventInfo.folders.length > 0}
      <div
        class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700"
        style="max-height: 35vh; overflow-y: auto;"
      >
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-2">
            <Icon icon="mdi:paperclip" class="inline-block w-5 h-5 mr-2 -mt-1" />
            Materials & Attachments
          </h2>

          {#each eventInfo.folders as folder}
            {#if folder.attachments && folder.attachments.length > 0}
            <div class="mb-6 last:mb-0">
              <div class="flex items-center gap-2 mb-2">
                <Icon icon="mdi:folder" class="w-5 h-5 text-amber-600 dark:text-amber-400" />
                <h3 class="text-base font-medium text-gray-700 dark:text-gray-300">
                  {folder.title || 'Attachments'}
                </h3>
                {#if folder.default_folder}
                  <span
                    class="px-2 py-0.5 text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded"
                  >
                    Default
                  </span>
                {/if}
                {#if folder.is_protected}
                  <span title="Protected">
                    <Icon icon="mdi:lock" class="w-4 h-4 text-red-600 dark:text-red-400" />
                  </span>
                {/if}
              </div>

              {#if folder.description}
                <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">{folder.description}</p>
              {/if}

              <AttachmentGrid attachments={folder.attachments} dedupe={true} />
            </div>
            {/if}
          {/each}
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="p-6 text-center text-gray-500">No event information available.</div>
{/if}
