<script>
  import Icon from '@iconify/svelte';

  // Props via Svelte 5 rune-style $props()
  let {
    hasReview = false,
    disabled = false,
    title = undefined,
    ariaLabel = undefined,
    onClick = null,
  } = $props();

  function handleClick(e) {
    // Keep table/card behavior: stop propagation on inner click
    try {
      e.stopPropagation();
    } catch (err) {}
    if (disabled) return;
    // call prop callback if provided
    if (typeof onClick === 'function') onClick(e);
  }
</script>

<button
  type="button"
  class={hasReview
    ? 'px-2 py-1 text-xs rounded bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-200 font-semibold flex items-center gap-1 hover:bg-green-200 dark:hover:bg-green-800 transition-colors cursor-pointer'
    : 'px-2 py-1 text-xs rounded bg-purple-100 dark:bg-purple-900 text-purple-700 dark:text-purple-200 font-semibold flex items-center gap-1 hover:bg-purple-200 dark:hover:bg-purple-800 transition-colors cursor-pointer'}
  onclick={handleClick}
  {disabled}
  title={title ?? (hasReview ? 'Click to update your review' : 'Click to submit your review')}
  aria-label={ariaLabel ?? (hasReview ? 'Update review' : 'Submit review')}
>
  <Icon icon={hasReview ? 'mdi:clipboard-check' : 'mdi:clipboard-list'} class="w-3 h-3" />
  {#if hasReview}
    Update Review
  {:else}
    Submit Review
  {/if}
</button>
