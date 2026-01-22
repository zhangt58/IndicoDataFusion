<script>
  import { Modal } from 'flowbite-svelte';
  import {
    CloseOutline,
    MessagesOutline,
    ChevronLeftOutline,
    ChevronRightOutline,
  } from 'flowbite-svelte-icons';
  import AbstractReview from '../components/AbstractReview.svelte';
  import { onMount, onDestroy } from 'svelte';

  let {
    open = $bindable(false),
    reviews = [],
    abstractTitle = '',
    onAffiliationClick = null,
  } = $props();

  let currentIndex = $state(0);
  const currentReview = $derived(reviews && reviews.length > 0 ? reviews[currentIndex] : null);
  const hasPrevious = $derived(currentIndex > 0);
  const hasNext = $derived(reviews && currentIndex < reviews.length - 1);

  function goToPrevious() {
    if (hasPrevious) currentIndex--;
  }

  function goToNext() {
    if (hasNext) currentIndex++;
  }

  // Keyboard navigation for the dialog
  function isTypingElement(el) {
    if (!el) return false;
    const tag = el.tagName;
    return tag === 'INPUT' || tag === 'TEXTAREA' || el.isContentEditable;
  }

  function handleKeydown(e) {
    if (e.altKey || e.ctrlKey || e.metaKey) return;
    const active = document.activeElement;
    if (isTypingElement(active)) return;

    const key = e.key;
    if (key === 'ArrowLeft' || key === 'k') {
      goToPrevious();
    } else if (key === 'ArrowRight' || key === 'j') {
      goToNext();
    }
  }

  onMount(() => window.addEventListener('keydown', handleKeydown));
  onDestroy(() => window.removeEventListener('keydown', handleKeydown));

  function closeDialog() {
    open = false;
    currentIndex = 0;
  }

  $effect(() => {
    if (open && reviews && reviews.length > 0 && currentIndex >= reviews.length) {
      currentIndex = 0;
    }
  });

  const stats = $derived({
    total: reviews ? reviews.length : 0,
    accept: reviews ? reviews.filter((r) => r.proposed_action === 'accept').length : 0,
    reject: reviews ? reviews.filter((r) => r.proposed_action === 'reject').length : 0,
    changeTracks: reviews ? reviews.filter((r) => r.proposed_action === 'change_tracks').length : 0,
    duplicate: reviews
      ? reviews.filter((r) => r.proposed_action === 'mark_as_duplicate').length
      : 0,
  });
</script>

<Modal bind:open size="lg" dismissable={false}>
  <div class="flex justify-between items-start mb-2">
    <div class="flex items-center gap-2 flex-1 min-w-0">
      <MessagesOutline class="w-6 h-6 text-blue-600 dark:text-blue-400 shrink-0" />
      <div class="flex-1 min-w-0">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Abstract Reviews</h3>
        {#if abstractTitle}
          <p class="text-sm text-gray-600 dark:text-gray-400 truncate" title={abstractTitle}>
            {abstractTitle}
          </p>
        {/if}
      </div>
    </div>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-2 inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white shrink-0"
      onclick={closeDialog}
    >
      <CloseOutline class="w-5 h-5" />
    </button>
  </div>

  {#if reviews && reviews.length > 0}
    <div
      class="bg-gray-50 dark:bg-gray-700 rounded-lg p-3 mb-2 grid grid-cols-2 md:grid-cols-5 gap-2"
    >
      <div class="text-center">
        <p class="text-xs text-gray-600 dark:text-gray-400">Total</p>
        <p class="text-lg font-bold text-gray-800 dark:text-white">{stats.total}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-green-600 dark:text-green-400">Accept</p>
        <p class="text-lg font-bold text-green-700 dark:text-green-300">{stats.accept}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-red-600 dark:text-red-400">Reject</p>
        <p class="text-lg font-bold text-red-700 dark:text-red-300">{stats.reject}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-blue-600 dark:text-blue-400">Change Tracks</p>
        <p class="text-lg font-bold text-blue-700 dark:text-blue-300">{stats.changeTracks}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-orange-600 dark:text-orange-400">Duplicate</p>
        <p class="text-lg font-bold text-orange-700 dark:text-orange-300">{stats.duplicate}</p>
      </div>
    </div>

    <div
      class="flex justify-between items-center mb-2 pb-2 border-b border-gray-200 dark:border-gray-700"
    >
      <button
        type="button"
        class="flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        onclick={goToPrevious}
        disabled={!hasPrevious}
      >
        <ChevronLeftOutline class="w-4 h-4" />
        Previous
      </button>
      <span class="text-sm font-semibold text-gray-700 dark:text-gray-300"
        >Review {currentIndex + 1} of {reviews.length}</span
      >
      <button
        type="button"
        class="flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        onclick={goToNext}
        disabled={!hasNext}
      >
        Next
        <ChevronRightOutline class="w-4 h-4" />
      </button>
    </div>

    <div class="max-h-[60vh] overflow-y-auto">
      <AbstractReview review={currentReview} {onAffiliationClick} />
    </div>

    {#if reviews.length > 1}
      <div
        class="flex justify-center gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700"
      >
        {#each reviews as review, index}
          <button
            type="button"
            class="w-2 h-2 rounded-full transition-all {index === currentIndex
              ? 'bg-blue-600 dark:bg-blue-400 w-6'
              : 'bg-gray-300 dark:bg-gray-600 hover:bg-gray-400 dark:hover:bg-gray-500'}"
            onclick={() => (currentIndex = index)}
            aria-label="Go to review {index + 1}"
            title="Review {index + 1}"
          ></button>
        {/each}
      </div>
    {/if}
  {:else}
    <div class="text-center py-8">
      <MessagesOutline class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-3" />
      <p class="text-gray-500 dark:text-gray-400">No reviews available for this abstract.</p>
    </div>
  {/if}
</Modal>
