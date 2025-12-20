<script>
  // Reusable Reveal dialog for displaying a secret token
  let {
    open = $bindable(false),
    loading = false,
    name = '',
    token = '',
    error = '',
    onClose = () => {},
  } = $props();

  function close() {
    open = false;
    try {
      onClose();
    } catch (e) {
      console.error('RevealDialog onClose error', e);
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <div
      class="absolute inset-0 bg-black/40"
      role="button"
      tabindex="0"
      aria-label="Close dialog"
      onclick={close}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar' || e.key === 'Escape') {
          e.preventDefault();
          close();
        }
      }}
    ></div>
    <div
      class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
      role="dialog"
      aria-modal="true"
      tabindex="0"
    >
      <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
        Reveal API Token — {name ?? ''}
      </h4>

      {#if loading}
        <p class="text-sm text-gray-600 dark:text-gray-400">Loading…</p>
      {:else if error}
        <p class="text-sm text-red-600">Error: {error}</p>
      {:else}
        <div class="mb-2">
          <div class="block text-xs text-gray-600 dark:text-gray-400 mb-1">Token value</div>
          <div class="font-mono wrap-break-word p-2 rounded bg-gray-100 dark:bg-gray-700 text-sm">
            {token}
          </div>
        </div>
        <div class="flex justify-end gap-2">
          <button class="px-2 py-1 rounded bg-gray-200" onclick={close}>Close</button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* keep 2-space indentation for frontend files */
</style>
