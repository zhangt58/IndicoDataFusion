<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    GridOutline,
    CreditCardOutline,
    RefreshOutline,
    ChartOutline,
    ClipboardListOutline,
  } from 'flowbite-svelte-icons';
  import { GetAbstracts, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { createCachePage } from '../utils/cacheUtils.js';
  import AbstractCardView from './AbstractCardView.svelte';
  import AbstractTableView from './AbstractTableView.svelte';
  import AbstractChartView from './AbstractChartView.svelte';
  import LoadErrorHint from './LoadErrorHint.svelte';
  import AbstractMyReviews from './AbstractMyReviews.svelte';

  let loading = $state(false);
  let refreshing = $state(false);
  let abstractData = $state([]);
  let error = $state(null);
  let viewMode = $state('card');
  let isTestMode = $state(false);
  let cacheExpired = $state(false);
  let showReviewPanel = $state(false);
  let selectedReviewTrackID = $state(null);
  let reviewButton = $state(null);

  async function loadData() {
    loading = true;
    error = null;
    try {
      abstractData = (await GetAbstracts()) || [];
    } catch (e) {
      console.error('GetAbstracts failed', e);
      abstractData = [];
      error = e;
    } finally {
      loading = false;
    }
  }

  const { handleRefresh, handleCacheEvent } = createCachePage(
    'abstracts',
    loadData,
    (v) => {
      refreshing = v;
    },
    (err) => {
      error = err;
    },
  );

  // Handle track filter from review panel
  function handleFilterTrackByID(trackID) {
    // Switch to table view for better filtering visualization
    viewMode = 'table';
    // Set and pass the selected review track ID to AbstractTableView
    selectedReviewTrackID = trackID;
  }

  onMount(async () => {
    try {
      isTestMode = await IsTestMode();
    } catch (e) {
      console.error('Failed to check test mode', e);
    }

    await loadData();

    // Get current data source name from cache stats
    let currentDataSource = null;
    try {
      const stats = await GetCacheStats();
      currentDataSource = stats?.data_source_name || null;
    } catch (e) {
      currentDataSource = null;
    }

    EventsOn('cache:updated', (...data) => {
      const ev = (data && data.length ? data[0] : data) || {};

      if (ev.data_source_name && currentDataSource && ev.data_source_name !== currentDataSource) {
        return;
      }

      // Handle expired notification from backend goroutine
      if (ev.action === 'expired' && ev.key === 'abstracts') {
        cacheExpired = true;
        return;
      }

      // Handle refresh/delete/clear actions
      if (ev.action === 'refreshed' && ev.key === 'abstracts') {
        cacheExpired = false;
      }

      handleCacheEvent(ev);
    });
  });

  onDestroy(() => {
    EventsOff('cache:updated');
  });
</script>

{#if loading}
  <div class="flex items-center justify-center py-12">
    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-400"></div>
    <span class="ml-3 text-2xl text-gray-600 dark:text-gray-400">Loading...</span>
  </div>
{:else if error}
  <LoadErrorHint {error} message="Failed to load abstracts." />
{:else}
  <div
    class="fixed bg-sky-300 dark:bg-sky-800 top-12 left-2 shadow-md px-2 py-1 rounded-sm flex items-center gap-2 z-999"
  >
    <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">
      Abstracts ({abstractData.length})
    </h2>
    <div class="flex gap-1 ml-2">
      {#if !isTestMode}
        <div class="relative">
          <button
            onclick={() => handleRefresh()}
            disabled={refreshing}
            class="p-1.5 rounded transition-colors hover:bg-sky-100 disabled:opacity-50"
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
      <button
        onclick={() => (viewMode = 'card')}
        class="p-1.5 rounded transition-colors {viewMode === 'card'
          ? 'bg-sky-400'
          : 'hover:bg-sky-100'}"
        title="Card View"
      >
        <CreditCardOutline class="shrink-0 h-6 w-6" />
      </button>
      <button
        onclick={() => (viewMode = 'table')}
        class="p-1.5 rounded transition-colors {viewMode === 'table'
          ? 'bg-sky-400'
          : 'hover:bg-sky-100'}"
        title="Table View"
      >
        <GridOutline class="shrink-0 h-6 w-6" />
      </button>
      <button
        onclick={() => (viewMode = 'chart')}
        class="p-1.5 rounded transition-colors {viewMode === 'chart'
          ? 'bg-sky-400'
          : 'hover:bg-sky-100'}"
        title="Chart View"
        aria-label="Chart View"
        disabled={loading || abstractData.length === 0}
      >
        <ChartOutline class="shrink-0 h-6 w-6" />
      </button>
      <button
        bind:this={reviewButton}
        onclick={() => (showReviewPanel = !showReviewPanel)}
        class="p-1.5 rounded transition-colors {showReviewPanel
          ? 'bg-sky-400'
          : 'hover:bg-sky-100'}"
        title="My Reviews"
        aria-label="My Reviews"
      >
        <ClipboardListOutline class="shrink-0 h-6 w-6" />
      </button>
    </div>
  </div>

  <!-- Review Panel -->
  <AbstractMyReviews
    bind:open={showReviewPanel}
    onFilterTrackByID={handleFilterTrackByID}
    buttonElement={reviewButton}
  />

  <div
    class="max-w-full overflow-x-auto overflow-y-hidden rounded-md
              mt-8 mb-4 h-[calc(100vh-9rem)]"
  >
    {#if viewMode === 'table'}
      <AbstractTableView
        bind:abstractData
        selectedReviewTrackID={selectedReviewTrackID}
      />
    {:else if viewMode === 'chart'}
      {#if abstractData && abstractData.length > 0}
        <AbstractChartView {abstractData} />
      {:else}
        <div class="p-4 text-center text-slate-500">No abstracts to display for chart.</div>
      {/if}
    {:else}
      <AbstractCardView bind:abstractData />
    {/if}
  </div>
{/if}
