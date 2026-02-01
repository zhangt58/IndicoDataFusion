<script>
  import { onMount, onDestroy } from 'svelte';
  import { GetEventInfo, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
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

  // Helper to format file size
  function formatFileSize(bytes) {
    if (!bytes || bytes === 0) return '';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  }

  // Deduplicate attachments by title, keeping the latest modified
  function deduplicateAttachments(attachments) {
    if (!attachments || attachments.length === 0) return [];

    const titleMap = new Map();

    for (const attachment of attachments) {
      const rawKey = attachment.title || attachment.filename || 'untitled';
      let key = String(rawKey).trim().toLowerCase();
      const existing = titleMap.get(key);

      if (!existing) {
        titleMap.set(key, attachment);
      } else {
        // Compare modified dates, keep the latest
        const existingDate = existing.modified_dt ? new Date(existing.modified_dt) : new Date(0);
        const currentDate = attachment.modified_dt ? new Date(attachment.modified_dt) : new Date(0);

        if (currentDate > existingDate) {
          titleMap.set(key, attachment);
        }
      }
    }

    return Array.from(titleMap.values());
  }

  // Get file type icon and color based on content type or filename
  function getFileIcon(attachment) {
    const contentType = attachment.content_type || '';
    const filename = attachment.filename || attachment.title || '';
    const ext = filename.split('.').pop()?.toLowerCase() || '';

    // PDF files
    if (contentType.includes('pdf') || ext === 'pdf') {
      return {
        icon: 'pdf',
        color: 'text-red-600 dark:text-red-400',
        bgColor: 'bg-red-50 dark:bg-red-900/20'
      };
    }

    // Image files
    if (contentType.includes('image') || ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'].includes(ext)) {
      return {
        icon: 'image',
        color: 'text-green-600 dark:text-green-400',
        bgColor: 'bg-green-50 dark:bg-green-900/20'
      };
    }

    // Word documents
    if (contentType.includes('word') || contentType.includes('msword') || ['doc', 'docx'].includes(ext)) {
      return {
        icon: 'document',
        color: 'text-blue-600 dark:text-blue-400',
        bgColor: 'bg-blue-50 dark:bg-blue-900/20'
      };
    }

    // Excel/spreadsheet files
    if (contentType.includes('excel') || contentType.includes('spreadsheet') || ['xls', 'xlsx', 'csv'].includes(ext)) {
      return {
        icon: 'table',
        color: 'text-emerald-600 dark:text-emerald-400',
        bgColor: 'bg-emerald-50 dark:bg-emerald-900/20'
      };
    }

    // PowerPoint files
    if (contentType.includes('presentation') || contentType.includes('powerpoint') || ['ppt', 'pptx'].includes(ext)) {
      return {
        icon: 'presentation',
        color: 'text-orange-600 dark:text-orange-400',
        bgColor: 'bg-orange-50 dark:bg-orange-900/20'
      };
    }

    // Archive files
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
      return {
        icon: 'archive',
        color: 'text-purple-600 dark:text-purple-400',
        bgColor: 'bg-purple-50 dark:bg-purple-900/20'
      };
    }

    // Video files
    if (contentType.includes('video') || ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv'].includes(ext)) {
      return {
        icon: 'video',
        color: 'text-pink-600 dark:text-pink-400',
        bgColor: 'bg-pink-50 dark:bg-pink-900/20'
      };
    }

    // Default file icon
    return {
      icon: 'file',
      color: 'text-gray-600 dark:text-gray-400',
      bgColor: 'bg-gray-50 dark:bg-gray-700/50'
    };
  }

  // Handle attachment click
  async function openAttachment(url) {
    if (!url) return;
    try {
      await BrowserOpenURL(url);
    } catch (e) {
      console.error('Failed to open attachment URL:', e);
    }
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
      style={eventInfo.folders && eventInfo.folders.length > 0 ? 'max-height: 40vh; overflow-y: auto;' : ''}
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
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
            <svg
              class="inline-block w-5 h-5 mr-2 -mt-1"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
              />
            </svg>
            Materials & Attachments
          </h2>

          {#each eventInfo.folders as folder}
            {@const dedupedAttachments = deduplicateAttachments(folder.attachments)}
            {#if dedupedAttachments.length > 0}
              <div class="mb-6 last:mb-0">
                <!-- Folder Header -->
                <div class="flex items-center gap-2 mb-3">
                  <svg
                    class="w-5 h-5 text-amber-600 dark:text-amber-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                    />
                  </svg>
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
                      <svg
                        class="w-4 h-4 text-red-600 dark:text-red-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                        />
                      </svg>
                    </span>
                  {/if}
                </div>

                {#if folder.description}
                  <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">{folder.description}</p>
                {/if}

                <!-- Attachments Grid - Horizontal Layout -->
                <div class="flex flex-wrap gap-2">
                  {#each dedupedAttachments as attachment}
                    {@const fileInfo = getFileIcon(attachment)}
                    <button
                      onclick={() => openAttachment(attachment.download_url)}
                      disabled={!attachment.download_url}
                      class="group relative flex flex-col items-center justify-center w-24 h-24 rounded-lg border-2 border-gray-200 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 transition-all {fileInfo.bgColor} {!attachment.download_url ? 'opacity-50 cursor-not-allowed' : 'hover:shadow-lg cursor-pointer'}"
                      title={attachment.title || attachment.filename || 'Untitled'}
                    >
                      <!-- File Type Icon -->
                      <span class="mb-1 block">
                        {#if fileInfo.icon === 'pdf'}
                          <!-- PDF Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z"/>
                            <path d="M14 2v6h6M9.5 12.5v5M11.5 12.5v5M9.5 15h2" stroke="currentColor" stroke-width="1" fill="none"/>
                          </svg>
                        {:else if fileInfo.icon === 'image'}
                          <!-- Image Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                          </svg>
                        {:else if fileInfo.icon === 'document'}
                          <!-- Word Document Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z"/>
                            <path d="M14 2v6h6M8 13h8M8 17h8M8 9h2" stroke="white" stroke-width="1"/>
                          </svg>
                        {:else if fileInfo.icon === 'table'}
                          <!-- Excel/Table Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                          </svg>
                        {:else if fileInfo.icon === 'presentation'}
                          <!-- PowerPoint Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6"/>
                          </svg>
                        {:else if fileInfo.icon === 'archive'}
                          <!-- Archive/ZIP Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"/>
                          </svg>
                        {:else if fileInfo.icon === 'video'}
                          <!-- Video Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                          </svg>
                        {:else}
                          <!-- Default File Icon -->
                          <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                          </svg>
                        {/if}
                      </span>

                      <!-- File Name (truncated) -->
                      <span class="text-xs text-center {fileInfo.color} font-medium truncate w-full px-1">
                        {(attachment.title || attachment.filename || 'Untitled').split('.')[0].substring(0, 12)}
                        {(attachment.title || attachment.filename || '').length > 12 ? '...' : ''}
                      </span>

                      <!-- File Size Badge -->
                      {#if attachment.size}
                        <span class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                          {formatFileSize(attachment.size)}
                        </span>
                      {/if}

                      <!-- Protected Badge -->
                      {#if attachment.is_protected}
                        <span class="absolute top-1 right-1 block">
                          <svg class="w-3 h-3 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
                          </svg>
                        </span>
                      {/if}

                      <!-- Hover Tooltip -->
                      <span class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 dark:bg-gray-700 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10 block">
                        {attachment.title || attachment.filename || 'Untitled'}
                        {#if attachment.content_type}
                          <br><span class="text-gray-300">{attachment.content_type}</span>
                        {/if}
                      </span>
                    </button>
                  {/each}
                </div>
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
