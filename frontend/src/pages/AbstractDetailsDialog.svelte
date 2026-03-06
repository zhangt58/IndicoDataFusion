<script>
  import { Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import AbstractCardItem from './AbstractCardItem.svelte';
  import { onMount, onDestroy } from 'svelte';

  /**
   * open        - controls modal visibility
   * abstract    - the currently displayed abstract object
   * isMyReview  - whether the abstract belongs to the user's review track
   * currentIndex - index of the current abstract in the sorted/filtered list
   * totalCount   - total number of abstracts in the sorted/filtered list
   * onNavigate  - callback(direction: 'prev'|'next') to request navigation
   */
  let {
    open = $bindable(false),
    abstract = $bindable(null),
    isMyReview = false,
    currentIndex = 0,
    totalCount = 0,
    onNavigate = null,
    visibilityConfig = null,
  } = $props();

  // Close dialog
  function closeDialog() {
    open = false;
  }

  // Handle refresh from AbstractCardItem - ensure the bindable abstract is updated
  function handleRefresh(refreshedAbstract) {
    abstract = refreshedAbstract;
  }

  function goPrev() {
    onNavigate && onNavigate('prev');
  }

  function goNext() {
    onNavigate && onNavigate('next');
  }

  // Keyboard navigation — registered in bubble phase so capture-phase handlers in child
  // dialogs (AbstractReviewsDialog, AbstractReviewFormDialog) can call stopPropagation first.
  function handleKeydown(e) {
    if (!open) return;
    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      e.stopPropagation();
      goPrev();
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      e.stopPropagation();
      goNext();
    }
  }

  onMount(() => window.addEventListener('keydown', handleKeydown, false));
  onDestroy(() => window.removeEventListener('keydown', handleKeydown, false));
</script>

<Modal bind:open size="xl" dismissable={false}>
  <div class="grid grid-cols-3 items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Abstract Details</h3>
    <div class="flex items-center justify-center gap-1">
      {#if totalCount > 1}
        <button
          type="button"
          title="Previous abstract (←)"
          disabled={currentIndex <= 0}
          onclick={goPrev}
          class="text-gray-500 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
        >
          <Icon icon="mdi:chevron-left" class="shrink-0 h-6 w-6" />
        </button>
      {/if}
      {#if totalCount > 0}
        <span class="text-sm text-gray-500 dark:text-gray-400 tabular-nums">
          {currentIndex + 1} / {totalCount}
        </span>
      {/if}
      {#if totalCount > 1}
        <button
          type="button"
          title="Next abstract (→)"
          disabled={currentIndex >= totalCount - 1}
          onclick={goNext}
          class="text-gray-500 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
        >
          <Icon icon="mdi:chevron-right" class="shrink-0 h-6 w-6" />
        </button>
      {/if}
    </div>
    <div class="flex justify-end">
      <button
        type="button"
        class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
        onclick={closeDialog}
      >
        <Icon icon="mdi:close" class="shrink-0 h-6 w-6" />
      </button>
    </div>
  </div>
  {#if abstract}
    <AbstractCardItem bind:abstract onRefresh={handleRefresh} {isMyReview} {visibilityConfig} />
  {/if}
</Modal>
