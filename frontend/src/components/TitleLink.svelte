<script>
  let { as = 'button', className = '', children, ...rest } = $props();
</script>

{#if as === 'button'}
  <button class={'title-link ' + className} {...rest}>
    {@render children()}
  </button>
{:else}
  <a class={'title-link ' + className} {...rest}>
    {@render children()}
  </a>
{/if}

<style>
  .title-link {
    color: #0d6efd;
    text-decoration: none;
    cursor: pointer;
    display: inline-flex !important;
    align-items: flex-start !important;
    justify-content: flex-start !important;
    text-align: left !important;
    padding: 0 !important;
    background: transparent !important;
    border: none !important;
    font-size: 0.875rem;
    font-weight: 400;
  }

  .title-link:hover {
    color: #0a58ca;
    text-decoration: underline;
  }

  :global(.dark) .title-link {
    color: #60a5fa;
  }

  :global(.dark) .title-link:hover {
    color: #93c5fd;
  }
</style>
