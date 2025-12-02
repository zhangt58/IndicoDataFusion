<script>
  import { onMount } from 'svelte';
  export let error = null;
  export let message = null;
  export let title = 'Failed to load information';

  function openSettings() {
    // Dispatch a window-level event so the top-level App can open the Settings modal
    try {
      window.dispatchEvent(new CustomEvent('open:settings'));
    } catch (e) {
      console.warn('Could not dispatch open:settings event', e);
    }
  }

  onMount(() => {
    // Keep the raw error off the UI but available in the console for debugging
    if (error) {
      console.debug('LoadErrorHint internal error:', error);
    }
  });
</script>

<div class="p-6 text-center">
  <div class="inline-block max-w-xl text-left">
    <h2 class="text-lg font-semibold text-red-600 mb-2">{title}</h2>
    {#if message}
      <p class="text-sm text-red-500 mb-2">{message}</p>
    {/if}
    <!-- Intentionally not showing detailed error text to avoid exposing raw errors -->

    <div class="bg-yellow-50 border border-yellow-200 rounded-md p-4 mb-4">
      <p class="text-sm text-yellow-800 mb-3">
        Hint: Click the top <strong>Settings</strong> button and modify the data sources in <strong>ConfigurationTab</strong> to point the app at a valid Indico data source.
      </p>
      <div class="flex gap-2">
        <button on:click={openSettings} class="px-3 py-1 bg-yellow-600 text-white rounded hover:bg-yellow-700">Open Settings</button>
      </div>
    </div>
  </div>
</div>

<style>
  /* keep styling minimal and tailwind-friendly; fallback styles for non-tailwind environments */
  .bg-yellow-50 { background-color: #FFFBEB; }
  .border-yellow-200 { border-color: #FDE68A; }
  .text-yellow-800 { color: #92400E; }
  .rounded-md { border-radius: 0.375rem; }
</style>
