<script>
  export let text = '';
  export let as = 'span'; // 'span' or 'button'
  export let className = '';

</script>

{#if as === 'button'}
  <button class={"type-badge ml-4 px-3 py-1 rounded-full text-xs font-normal bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 " + className} {...$$restProps}>
    {text}
  </button>
{:else}
  <span class={"type-badge ml-4 px-3 py-1 rounded-full text-xs font-normal bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 " + className} {...$$restProps}>
    {text}
  </span>
{/if}

<style>
  .type-badge {
      /*display: inline-flex !important;*/
      display: inline-block;
      align-items: flex-start !important; /* vertical alignment inside the button */
      justify-content: flex-start !important; /* horizontal alignment */
      text-align: left !important; /* ensure multi-line text is left-aligned */
      padding: 0.125rem 0.5rem;
      border-radius: 0.25rem;
      font-size: 0.75rem;
      background-color: #e0e7ff !important;
      color: #3730a3 !important;
  }
  :global(.dark .type-badge) {
      background-color: #3730a3 !important;
      color: #e0e7ff !important;
  }
</style>