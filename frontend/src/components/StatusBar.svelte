<script>
  import { onMount, onDestroy } from 'svelte';
  import { DarkMode } from 'flowbite-svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { GetAppInfo, ReviewMode } from '../../wailsjs/go/main/App';

  let dataSource = $state('');
  let loading = $state(true);
  let appName = $state('');
  let appVersion = $state('');
  let reviewMode = $state(false);

  // Live clock state
  let currentTime = $state('');
  let _clockTimer = null;

  function formatDateTime(d) {
    // YYYY-MM-DD HH:MM:SS
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  }

  function openConfigSettings() {
    try {
      window.dispatchEvent(new CustomEvent('open:settings', { detail: { tab: 'config' } }));
    } catch (e) {
      console.error('Failed to open settings (config tab):', e);
    }
  }

  function openAbout() {
    try {
      window.dispatchEvent(new CustomEvent('open:settings', { detail: { tab: 'about' } }));
    } catch (e) {
      console.error('Failed to open About dialog from StatusBar:', e);
    }
  }

  function handleAppDatasource(ev) {
    try {
      const v = ev && ev.detail ? ev.detail : ev;
      if (typeof v === 'string') {
        dataSource = v;
      } else if (v && v.data_source_name) {
        dataSource = v.data_source_name;
      } else if (v && v.name) {
        dataSource = v.name;
      }
    } catch (e) {
      console.debug('app:datasource handler error', e);
    }
    // Re-fetch reviewMode whenever the active data source changes
    ReviewMode()
      .then((v) => (reviewMode = v))
      .catch((e) => console.debug('Failed to refresh review mode on datasource change', e));
  }

  function handleAppReviewMode(ev) {
    try {
      const v = ev && ev.detail !== undefined ? ev.detail : ev;
      reviewMode = !!v;
    } catch (e) {
      console.debug('app:reviewmode handler error', e);
    }
  }

  onMount(async () => {
    // initialize clock
    currentTime = formatDateTime(new Date());
    try {
      _clockTimer = setInterval(() => {
        currentTime = formatDateTime(new Date());
      }, 1000);
    } catch (e) {
      console.debug('Failed to start clock timer', e);
    }

    // Check review mode
    try {
      reviewMode = await ReviewMode();
    } catch (e) {
      console.debug('Failed to check review mode', e);
    }

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
      EventsOn('app:reviewmode', handleAppReviewMode);
    } catch (e) {
      console.debug('Failed to subscribe to app:reviewmode in StatusBar', e);
    }

    try {
      const info = await GetAppInfo();
      if (info) {
        // Try a few plausible field names for an active data source in AppInfo
        const ds = info.dataSource ?? null;
        if (ds && !dataSource) {
          dataSource = ds;
        }

        // app name/version for display in status bar
        if (info.name) appName = info.name;
        if (info.version) appVersion = info.version;
      }
    } catch (e) {
      // Not fatal; backend may not populate AppInfo with data source.
      console.debug('GetAppInfo failed or did not include data source', e);
    }
  });

  onDestroy(() => {
    try {
      EventsOff('app:datasource');
    } catch (e) {
      /* ignore */
    }
    try {
      EventsOff('app:reviewmode');
    } catch (e) {
      /* ignore */
    }
    try {
      if (_clockTimer) clearInterval(_clockTimer);
    } catch (e) {
      /* ignore */
    }
  });

  // tooltip text reactive
  let tooltipText = $derived(
    dataSource ? `Active Data Source: ${dataSource}` : 'No active data source',
  );
</script>

<!-- Fixed status bar at the bottom of the window -->
<div
  class="fixed bottom-0 left-0 right-0 h-10 rounded-b-xl bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 text-sm text-gray-700 dark:text-gray-300 flex items-center justify-between z-40 shadow-md dark:shadow-black/40"
  style="--wails-draggable: no-drag; padding-left:0.5rem; padding-right:0.5rem; z-index: 10000; box-shadow: 0 -8px 24px -16px rgba(0,0,0,0.6), 0 1px 0 rgba(0,0,0,0.04);"
>
  {#if loading}
    <span class="px-3">Loading…</span>
  {:else}
    <!-- Left: App name and version -->
    <div class="flex items-center gap-2 pr-3 justify-start min-w-2xs">
      <DarkMode class="px-1 py-1 rounded dark:text-primary-600" />
      {#if appName}
        <button
          class="text-sm text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-300 transition-colors px-2 py-1 rounded focus:outline-none"
          onclick={openAbout}
          title={appVersion ? `${appName} - ${appVersion}` : appName}
          aria-label="Open About dialog"
        >
          <span class="font-semibold">{appName}</span>{#if appVersion}<span class="text-sm ml-1"
              >{`${appVersion}`}</span
            >{/if}
        </button>
      {/if}
    </div>
    {#if dataSource}
      <div
        class="flex items-center gap-3 max-w-[70%] mx-auto px-2"
        role="group"
        aria-label={`Data source in use: ${dataSource}`}
      >
        {#if reviewMode}
          <span
            class="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200 text-xs font-semibold shadow-sm"
            title="Review Mode: Limited data visibility for reviewers"
          >
            <svg
              class="w-3 h-3"
              fill="currentColor"
              viewBox="0 0 20 20"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                clip-rule="evenodd"
              ></path>
            </svg>
            Review Mode
          </span>
        {/if}
        <span class="text-sm text-gray-500 dark:text-gray-400">Data source:</span>

        <button
          class="inline-flex items-center gap-2 truncate max-w-[60%] text-sm font-semibold px-3 py-1 rounded-md bg-indigo-50 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-200 hover:bg-indigo-200 dark:hover:bg-indigo-700 transition-colors duration-150 shadow-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 relative"
          title={tooltipText}
          aria-label={`Open Config settings for ${dataSource}`}
          onclick={openConfigSettings}
          onkeydown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              openConfigSettings();
            }
          }}
        >
          <span class="truncate">{dataSource}</span>
        </button>
      </div>
    {:else}
      <span class="text-gray-500">No data source configured</span>
    {/if}

    <!-- Right: Live clock -->
    <div class="flex items-center gap-2 pl-3 justify-end min-w-2xs">
      <span
        class="text-sm font-mono text-gray-500 dark:text-gray-400"
        title={new Date().toString()}
        aria-label="Current date and time">{currentTime}</span
      >
    </div>
  {/if}
</div>
