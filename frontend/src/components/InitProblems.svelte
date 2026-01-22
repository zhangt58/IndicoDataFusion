<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
  import { GetInitProblems } from '../../wailsjs/go/main/App';

  let problems = $state([]);
  let show = $state(true);

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
  <div class="fixed top-16 right-4 z-50 w-96">
    <div
      class="bg-yellow-50 dark:bg-yellow-900 border border-yellow-400 dark:border-yellow-700 rounded shadow p-3 text-sm text-yellow-900 dark:text-yellow-100"
    >
      <div class="flex justify-between items-start">
        <div class="font-semibold">Configuration issues</div>
        <button class="text-xs px-2 py-1" aria-label="Dismiss" onclick={() => (show = false)}
          >Dismiss</button
        >
      </div>
      <ul class="mt-2 space-y-2">
        {#each problems as p}
          <li class="flex items-start justify-between gap-2">
            <div class="flex-1">
              <div class="text-xs wrap-break-word">{p}</div>
            </div>
          </li>
        {/each}
      </ul>
      <div class="mt-2 text-xs text-gray-700 dark:text-gray-200 flex items-center justify-between">
        <div>You can manage tokens in Configuration → Advanced → API Tokens</div>
        <div>
          <button
            class="text-xs px-2 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700"
            aria-label="Open Settings — Configuration"
            onclick={openSettingsConfig}
          >
            Open
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}