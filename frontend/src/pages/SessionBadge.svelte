<script>
  export let text = '';
  export let as = 'span';
  export let className = '';
</script>

{#if as === 'button'}
  <button class={"session-badge text-sm px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded " + className} {...$$restProps}>
    {text}
  </button>
{:else}
  <span class={"session-badge text-sm px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded " + className} {...$$restProps}>
    {text}
  </span>
{/if}

<style>
  .session-badge {
    display: inline-flex !important;
    align-items: flex-start !important; /* vertical alignment inside the button */
    justify-content: flex-start !important; /* horizontal alignment */
    text-align: left !important; /* ensure multi-line text is left-aligned */
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }
  :global(.dark .session-badge) {
      background-color: #581c87 !important;
      color: #e9d5ff !important;
  }
</style>





