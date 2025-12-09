<script>
  import { onMount, onDestroy } from 'svelte';
  import { GridOutline, CreditCardOutline, RefreshOutline } from 'flowbite-svelte-icons';
  import { GetContributions, IsTestMode, GetCacheStats } from '../../wailsjs/go/main/App';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { createCachePage } from '../utils/cacheUtils.js';
  import ContributionCardView from './ContributionCardView.svelte';
  import ContributionTableView from './ContributionTableView.svelte';
  import LoadErrorHint from './LoadErrorHint.svelte';

  let loading = false;
  let refreshing = false;
  let contributionData = [];
  let error = null;
  let viewMode = 'card';
  let isTestMode = false;
  let cacheExpired = false;

  async function loadData() {
    loading = true;
    error = null;
    try {
      contributionData = (await GetContributions()) || [];
    } catch (e) {
      console.error('GetContributions failed', e);
      contributionData = [];
      error = e;
    } finally {
      loading = false;
    }
  }

  const { handleRefresh, handleCacheEvent } = createCachePage(
    'contributions',
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
      if (ev.action === 'expired' && ev.key === 'contributions') {
        cacheExpired = true;
        return;
      }

      // Handle refresh/delete/clear actions
      if (ev.action === 'refreshed' && ev.key === 'contributions') {
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
  <div class="p-6 text-center">Loading contributions...</div>
{:else if error}
  <LoadErrorHint {error} message="Failed to load contributions." />
{:else}
  <div class="fixed bg-indigo-300 dark:bg-indigo-800 top-12 left-2 shadow-md px-2 py-1 rounded-sm flex items-center gap-2 z-10">
    <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">Contributions ({contributionData.length})</h2>
    <div class="flex gap-1 ml-2">
      {#if !isTestMode}
        <div class="relative">
          <button
            onclick={() => handleRefresh()}
            disabled={refreshing}
            class="p-1.5 rounded transition-colors hover:bg-indigo-100 disabled:opacity-50"
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
        class="p-1.5 rounded transition-colors {viewMode === 'card' ? 'bg-indigo-400' : 'hover:bg-indigo-100'}"
        title="Card View"
      >
        <CreditCardOutline class="shrink-0 h-6 w-6" />
      </button>
      <button
        onclick={() => viewMode = 'table'}
        class="p-1.5 rounded transition-colors {viewMode === 'table' ? 'bg-indigo-400' : 'hover:bg-indigo-100'}"
        title="Table View"
      >
        <GridOutline class="shrink-0 h-6 w-6" />
      </button>
    </div>
  </div>

  {#if viewMode === 'table'}
    <ContributionTableView {contributionData} />
  {:else}
    <ContributionCardView {contributionData} />
  {/if}
{/if}
