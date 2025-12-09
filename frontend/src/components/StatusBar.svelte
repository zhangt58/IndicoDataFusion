<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { GetAppInfo } from '../../wailsjs/go/main/App';

  let dataSource = '';
  let loading = true;

  function openConfigSettings() {
    try {
      window.dispatchEvent(new CustomEvent('open:settings', { detail: { tab: 'config' } }));
    } catch (e) {
      console.error('Failed to open settings (config tab):', e);
    }
  }

  function handleAppDatasource(ev) {
    try {
      // runtime may pass the payload as ev.detail or raw value
      const v = ev && ev.detail ? ev.detail : ev;
      if (typeof v === 'string') {
        dataSource = v;
      }
      else if (v && v.data_source_name) {
        dataSource = v.data_source_name;
      }
      else if (v && v.name) {
        dataSource = v.name;
      }
    } catch (e) {
      console.debug('app:datasource handler error', e);
    }
  }

  onMount(async () => {
    // No blocking cache lookup here; rely primarily on backend events, but try to
    // initialize from AppInfo if it contains the active data source (robust to
    // different backend shapes).
    loading = false;

    try {
      EventsOn('app:datasource', handleAppDatasource);
    } catch (e) {
      console.debug('Failed to subscribe to events in StatusBar', e);
    }

    try {
      const info = await GetAppInfo();
      if (info) {
        // Try a few plausible field names for an active data source in AppInfo
        const ds = info.dataSource ?? null;
        if (ds && !dataSource) {
          dataSource = ds;
        }
      }
    } catch (e) {
      // Not fatal; backend may not populate AppInfo with data source.
      console.debug('GetAppInfo failed or did not include data source', e);
    }
  });

  onDestroy(() => {
    try { EventsOff('app:datasource'); } catch (e) { /* ignore */ }
  });

  // tooltip text reactive
  $: tooltipText = dataSource ? `${dataSource}` : '';
</script>

<!-- Fixed status bar at the bottom of the window -->
<div class="fixed bottom-0 left-0 right-0 h-10 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 text-sm text-gray-700 dark:text-gray-300 flex items-center justify-center z-40" style="--wails-draggable: no-drag; padding-left:0.5rem; padding-right:0.5rem;">
    {#if loading}
      <span class="px-3">Loading…</span>
    {:else}
      {#if dataSource}
      <div
        class="flex items-center gap-3 max-w-[90%] mx-auto px-2"
        role="group"
        aria-label={`Data source in use: ${dataSource}`}
      >
        <span class="text-sm text-gray-500 dark:text-gray-400">Data source:</span>

        <button
          class="inline-flex items-center gap-2 truncate max-w-[60%] text-sm font-semibold px-3 py-1 rounded-md bg-indigo-50 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-200 hover:bg-indigo-200 dark:hover:bg-indigo-700 transition-colors duration-150 shadow-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 relative"
           title={tooltipText}
           aria-label={`Open Config settings for ${dataSource}`}
           on:click={openConfigSettings}
           on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); openConfigSettings(); } }}
         >
          <span class="truncate">{dataSource}</span>
         </button>
        </div>
       {:else}
         <span class="text-gray-500">No data source configured</span>
       {/if}
     {/if}
   </div>

<style>
  /* Ensure the status bar does not cover important UI when present; app content may need bottom padding */
 </style>
