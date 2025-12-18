<script>
  let { text = '', as = 'span', className = '', ...rest } = $props();
</script>

{#if as === 'button'}
  <button class={"session-badge text-sm px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded " + className} {...rest}>
    {text}
  </button>
{:else}
  <span class={"session-badge text-sm px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded " + className} {...rest}>
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
    transition: background-color 120ms ease, transform 100ms ease, box-shadow 120ms ease;
    cursor: default;
  }
  .session-badge:hover,
  .session-badge:focus {
    /* Slightly darker background and subtle lift to indicate interactivity (similar to track hover feel) */
    filter: brightness(0.95);
    transform: translateY(-1px);
    box-shadow: 0 1px 4px rgba(16,24,40,0.06);
  }
  .session-badge:active {
    transform: translateY(0);
  }

  /* When used as interactive link/button, parent may pass a class like 'session-link' or 'interactive'. Allow the element itself to adopt pointer cursor when so used */
  .session-badge.session-link,
  .session-badge.interactive {
    cursor: pointer;
  }

  :global(.dark .session-badge) {
      background-color: #581c87 !important;
      color: #e9d5ff !important;
  }
</style>
