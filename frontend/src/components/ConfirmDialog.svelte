<script>
  let {
    open = $bindable(false),
    danger = $bindable(false),
    title = 'Confirm',
    message = '',
    confirmLabel = 'Confirm',
    cancelLabel = 'Cancel',
    onConfirm = () => {},
    onCancel = () => {},
  } = $props();

  function doConfirm() {
    // close first so parent UI updates immediately
    open = false;
    try {
      onConfirm();
    } catch (e) {
      // swallow errors from callback to avoid breaking the dialog close
      console.error('ConfirmDialog onConfirm callback error', e);
    }
  }

  function doCancel() {
    open = false;
    try {
      onCancel();
    } catch (e) {
      console.error('ConfirmDialog onCancel callback error', e);
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-40 flex items-center justify-center">
    <div
      class="absolute inset-0 bg-black/40"
      role="button"
      tabindex="0"
      aria-label="Close dialog"
      onclick={doCancel}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar') {
          e.preventDefault();
          doCancel();
        }
      }}
    ></div>

    <div
      role="dialog"
      aria-modal="true"
      tabindex="0"
      onkeydown={(e) => {
        if (e.key === 'Escape') {
          e.stopPropagation();
          doCancel();
        }
      }}
      class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
    >
      <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{title}</h4>
      <p class="text-sm text-gray-700 dark:text-gray-300">{message}</p>
      <div class="mt-4 flex justify-end gap-2">
        <button
          type="button"
          class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm"
          onclick={doCancel}
        >
          {cancelLabel}
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 text-white"
          class:bg-red-600={danger}
          class:bg-indigo-600={!danger}
          onclick={doConfirm}
        >
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}