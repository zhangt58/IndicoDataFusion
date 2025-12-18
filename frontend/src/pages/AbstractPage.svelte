<script>
  import { onMount, onDestroy } from 'svelte';
  import { GridOutline, CreditCardOutline, RefreshOutline } from 'flowbite-svelte-icons';
  import { GetAbstracts, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { createCachePage } from '../utils/cacheUtils.js';
  import AbstractCardView from './AbstractCardView.svelte';
  import AbstractTableView from './AbstractTableView.svelte';
  import LoadErrorHint from './LoadErrorHint.svelte';

  let loading = false;
  let refreshing = false;
  let abstractData = [];
  let error = null;
  let viewMode = 'card';
  let isTestMode = false;
  let cacheExpired = false;

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
    (v) => { refreshing = v; },
    (err) => { error = err; }
  );

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
  <div class="p-6 text-center">Loading abstracts...</div>
{:else if error}
  <LoadErrorHint {error} message="Failed to load abstracts." />
{:else}
  <div class="fixed bg-sky-300 dark:bg-sky-800 top-12 left-2 shadow-md px-2 py-1 rounded-sm flex items-center gap-2 z-999">
    <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">Abstracts ({abstractData.length})</h2>
    <div class="flex gap-1 ml-2">
      {#if !isTestMode}
        <div class="relative">
          <button
            onclick={() => handleRefresh()}
            disabled={refreshing}
            class="p-1.5 rounded transition-colors hover:bg-sky-100 disabled:opacity-50"
            title={cacheExpired ? "Cache expired - Click to refresh" : "Refresh from API"}
          >
            <RefreshOutline class={`shrink-0 h-6 w-6 ${refreshing ? 'animate-spin' : ''}`} />
          </button>
          {#if cacheExpired && !refreshing}
            <span class="absolute -top-1 -right-1 flex h-3 w-3">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3 bg-red-500" title="Cache expired"></span>
            </span>
          {/if}
        </div>
      {/if}
      <button
        onclick={() => viewMode = 'card'}
        class="p-1.5 rounded transition-colors {viewMode === 'card' ? 'bg-sky-400' : 'hover:bg-sky-100'}"
        title="Card View"
      >
        <CreditCardOutline class="shrink-0 h-6 w-6" />
      </button>
      <button
        onclick={() => viewMode = 'table'}
        class="p-1.5 rounded transition-colors {viewMode === 'table' ? 'bg-sky-400' : 'hover:bg-sky-100'}"
        title="Table View"
      >
        <GridOutline class="shrink-0 h-6 w-6" />
      </button>
    </div>
  </div>

  <div class="max-w-full overflow-x-auto overflow-y-hidden rounded-md
              mt-8 mb-4 h-[calc(100vh-9rem)]">
  {#if viewMode === 'table'}
    <AbstractTableView {abstractData} />
  {:else}
    <AbstractCardView {abstractData} />
  {/if}
  </div>
{/if}
