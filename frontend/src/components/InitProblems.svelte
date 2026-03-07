<script>
  import { onMount, onDestroy } from 'svelte';
  import Icon from '@iconify/svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { GetInitProblems } from '../../wailsjs/go/main/App';

  let problems = $state([]);
  let show = $state(true);

  // Classify a problem string into a display kind
  function classify(p) {
    const s = p.toLowerCase();
    if (/token|api.?key|unauthorized|401|403/.test(s)) return 'auth';
    if (/not found|404|event.?id|base.?url/.test(s)) return 'config';
    if (/timeout|deadline/.test(s)) return 'timeout';
    if (/network|dial|connect|no such host/.test(s)) return 'network';
    return 'warning';
  }

  const kindMeta = {
    auth: { icon: 'mdi:key-alert', color: 'text-red-500 dark:text-red-400' },
    config: { icon: 'mdi:cog-off', color: 'text-orange-500 dark:text-orange-400' },
    timeout: { icon: 'mdi:timer-alert', color: 'text-yellow-500 dark:text-yellow-300' },
    network: { icon: 'mdi:lan-disconnect', color: 'text-orange-500 dark:text-orange-400' },
    warning: { icon: 'mdi:alert', color: 'text-yellow-500 dark:text-yellow-300' },
  };

  // Open the Settings modal and switch to the Configuration tab
  function openSettingsConfig() {
    try {
      window.dispatchEvent(new CustomEvent('open:settings', { detail: { tab: 'config' } }));
    } catch (e) {
      console.error('Failed to open settings (config):', e);
    }
  }

  async function loadProblems() {
    try {
      const p = await GetInitProblems();
      problems = p || [];
    } catch (e) {
      console.error('GetInitProblems failed', e);
      problems = [];
    }
  }

  function handleEvent(...data) {
    // Wails passes event payload as args; we expect an array of strings
    try {
      const payload = data && data.length > 0 ? data[0] : [];
      problems = payload || [];
      // Re-show the banner if new problems arrive
      if (problems.length > 0) show = true;
    } catch (e) {
      console.error('app:initproblems handler', e);
    }
  }

  let off;
  onMount(() => {
    loadProblems();
    off = EventsOn('app:initproblems', handleEvent);
  });
  onDestroy(() => {
    if (off) EventsOff('app:initproblems');
  });
</script>

{#if show && problems && problems.length > 0}
  <div class="fixed top-24 right-12 z-50 w-96 max-h-[80vh] flex flex-col">
    <div
      class="rounded-xl border border-amber-300 dark:border-amber-700 bg-white dark:bg-gray-900 shadow-lg overflow-hidden flex flex-col"
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between gap-2 bg-amber-50 dark:bg-amber-950 px-4 py-3 border-b border-amber-200 dark:border-amber-800"
      >
        <div class="flex items-center gap-2">
          <Icon
            icon="mdi:alert-circle-outline"
            class="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0"
          />
          <span class="text-sm font-semibold text-amber-800 dark:text-amber-200">
            Configuration {problems.length === 1 ? 'issue' : `issues (${problems.length})`}
          </span>
        </div>
        <button
          onclick={() => (show = false)}
          aria-label="Dismiss"
          class="p-1 rounded hover:bg-amber-100 dark:hover:bg-amber-900 text-amber-700 dark:text-amber-300 transition-colors"
        >
          <Icon icon="mdi:close" class="w-4 h-4" />
        </button>
      </div>

      <!-- Problem list -->
      <ul class="overflow-y-auto divide-y divide-amber-100 dark:divide-amber-900">
        {#each problems as p}
          {@const kind = classify(p)}
          {@const meta = kindMeta[kind]}
          <li class="flex items-start gap-3 px-4 py-3">
            <Icon icon={meta.icon} class="w-4 h-4 mt-0.5 shrink-0 {meta.color}" />
            <span class="text-xs text-gray-700 dark:text-gray-200 break-words leading-snug"
              >{p}</span
            >
          </li>
        {/each}
      </ul>

      <!-- Footer / action -->
      <div
        class="flex items-center justify-between gap-2 px-4 py-3 border-t border-amber-100 dark:border-amber-900 bg-amber-50/60 dark:bg-amber-950/40"
      >
        <span class="text-xs text-gray-500 dark:text-gray-400">
          Manage tokens in Configuration → Advanced → API Tokens
        </span>
        <div class="flex items-center gap-2 shrink-0">
          <button
            onclick={() => window.dispatchEvent(new CustomEvent('open:setup-wizard'))}
            aria-label="Open Setup Wizard"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-600 text-white text-xs font-medium transition-colors"
          >
            <Icon icon="mdi:auto-fix" class="w-3.5 h-3.5" />
            Wizard
          </button>
          <button
            onclick={openSettingsConfig}
            aria-label="Open Settings — Configuration"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium transition-colors"
          >
            <Icon icon="mdi:cog" class="w-3.5 h-3.5" />
            Open
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
