<script>
  import { Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import AbstractReviewForm from './AbstractReviewForm.svelte';
  import { RefreshAbstractByID } from '../../wailsjs/go/main/App';
  import { onMount, onDestroy } from 'svelte';

  /**
   * Props:
   *   open            - $bindable boolean controlling modal visibility
   *   abstract        - AbstractData object being reviewed
   *   reviewTrack     - Track {id, title} or ReviewTrack {track_id, name} for new reviews
   *   onAbstractUpdated - callback(refreshedAbstract) after successful submission
   */
  let {
    open = $bindable(false),
    abstract = null,
    reviewTrack = null,
    onAbstractUpdated = null,
  } = $props();

  let isRefreshing = $state(false);
  let refreshError = $state(null);

  const isEditMode = $derived(abstract?.my_review != null);

  function closeDialog() {
    open = false;
    refreshError = null;
  }

  async function handleSuccess() {
    if (!abstract?.id) {
      closeDialog();
      return;
    }

    isRefreshing = true;
    refreshError = null;

    try {
      const refreshed = await RefreshAbstractByID(abstract.id);
      if (onAbstractUpdated) {
        onAbstractUpdated(refreshed);
      }
      closeDialog();
    } catch (err) {
      console.error('Failed to refresh abstract after review submission:', err);
      refreshError =
        'Review submitted, but failed to refresh abstract data. Please refresh manually.';
      // Still notify parent with stale data so the UI isn't stuck
      if (onAbstractUpdated) {
        onAbstractUpdated(abstract);
      }
    } finally {
      isRefreshing = false;
    }
  }

  function handleCancel() {
    closeDialog();
  }

  // Block ArrowLeft/ArrowRight from reaching parent dialogs (e.g. AbstractDetailsDialog)
  // when this dialog is open. Use capture phase so stopPropagation takes effect first.
  function blockArrowKeysForParent(e) {
    if (!open) return;
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.stopPropagation();
    }
  }

  onMount(() => window.addEventListener('keydown', blockArrowKeysForParent, true));
  onDestroy(() => window.removeEventListener('keydown', blockArrowKeysForParent, true));
</script>

<Modal bind:open size="lg" dismissable={false} class="max-w-3xl">
  <!-- Modal Header -->
  <div
    class="flex items-center justify-between gap-3 mb-1 pb-2 border-b border-gray-200 dark:border-gray-600"
  >
    <div class="flex items-center gap-2 min-w-0">
      <Icon
        icon={isEditMode ? 'mdi:clipboard-edit-outline' : 'mdi:clipboard-plus-outline'}
        class="w-6 h-6 shrink-0 {isEditMode
          ? 'text-amber-500 dark:text-amber-400'
          : 'text-blue-500 dark:text-blue-400'}"
      />
      <div class="min-w-0">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white leading-tight">
          {isEditMode ? 'Update Review' : 'Submit Review'}
        </h3>
        {#if abstract?.title}
          <div class="flex items-center gap-2">
            <span class="text-xs">ID: {abstract.id} - </span>
            <p class="text-xs text-gray-500 dark:text-gray-400 truncate" title={abstract.title}>
              {abstract.title}
            </p>
          </div>
        {/if}
      </div>
    </div>

    <div class="flex items-center gap-2 shrink-0">
      <!-- Mode badge -->
      <span
        class="hidden sm:inline-flex text-xs px-2 py-0.5 rounded-full font-semibold
        {isEditMode
          ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300'
          : 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300'}"
      >
        {isEditMode ? 'Edit' : 'New'}
      </span>
      <!-- Close button -->
      <button
        type="button"
        onclick={closeDialog}
        disabled={isRefreshing}
        class="text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg p-1.5 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        aria-label="Close dialog"
      >
        <Icon icon="mdi:close" class="w-5 h-5" />
      </button>
    </div>
  </div>

  <!-- Modal Body -->
  <div class="overflow-y-auto max-h-full pr-1 -mr-1">
    {#if isRefreshing}
      <!-- Refreshing overlay -->
      <div class="flex flex-col items-center justify-center py-16 gap-3">
        <Icon icon="mdi:loading" class="w-10 h-10 animate-spin text-blue-500 dark:text-blue-400" />
        <p class="text-sm text-gray-600 dark:text-gray-400">Saving and refreshing abstract data…</p>
      </div>
    {:else if refreshError}
      <!-- Non-fatal refresh error: show warning but allow close -->
      <div class="p-4 space-y-3">
        <div
          class="flex items-start gap-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-700 rounded-md text-sm text-yellow-800 dark:text-yellow-300"
        >
          <Icon icon="mdi:alert-outline" class="w-5 h-5 shrink-0 mt-0.5" />
          <span>{refreshError}</span>
        </div>
        <div class="flex justify-end">
          <button
            type="button"
            onclick={closeDialog}
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    {:else if abstract}
      <div class="pt-2">
        <AbstractReviewForm
          {abstract}
          {reviewTrack}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    {:else}
      <div class="py-12 text-center text-gray-500 dark:text-gray-400">
        <Icon icon="mdi:file-alert-outline" class="w-10 h-10 mx-auto mb-2 opacity-50" />
        <p class="text-sm">No abstract data available.</p>
      </div>
    {/if}
  </div>
</Modal>
