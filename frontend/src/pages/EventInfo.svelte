<script>
  import { onMount } from 'svelte';
  import { GetEventData } from '../../wailsjs/go/backend/IndicoClient';

  let loading = false;
  let eventInfo = null;
  let error = null;

  onMount(async () => {
    loading = true;
    try {
      eventInfo = await GetEventData();
    } catch (e) {
      console.error('GetEventData failed', e);
      error = e;
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="p-6 text-center">Loading event information...</div>
{:else if error}
  <div class="p-6 text-center text-red-600">Failed to load event information: {error}</div>
{:else if eventInfo}
  <div class="max-w-4xl mx-auto mt-8">
    <!-- Event Header -->
    <div class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-t-lg p-6 text-white">
      <div class="flex items-center gap-2 mb-2">
        <span class="px-3 py-1 bg-white/20 rounded-full text-sm font-medium">
          {eventInfo.category || 'Conference'}
        </span>
        <span class="text-sm opacity-80">ID: {eventInfo.id}</span>
      </div>
      <h1 class="text-2xl md:text-3xl font-bold">{eventInfo.title}</h1>
    </div>

    <!-- Event Details Card -->
    <div class="bg-white dark:bg-gray-800 rounded-b-lg shadow-lg border border-gray-200 dark:border-gray-700">
      <!-- Date and Location -->
      <div class="p-6 border-b border-gray-200 dark:border-gray-700">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Dates -->
          <div class="flex items-start gap-3">
            <div class="p-2 bg-indigo-100 dark:bg-indigo-900 rounded-lg">
              <svg class="w-6 h-6 text-indigo-600 dark:text-indigo-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-semibold text-gray-600 dark:text-gray-400">Date</p>
              <p class="text-gray-800 dark:text-gray-200">{eventInfo.startDate || 'N/A'}</p>
              <p class="text-gray-600 dark:text-gray-400 text-sm">to {eventInfo.endDate || 'N/A'}</p>
            </div>
          </div>

          <!-- Location -->
          <div class="flex items-start gap-3">
            <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
              <svg class="w-6 h-6 text-green-600 dark:text-green-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
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
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">About the Event</h2>
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
