<script>
  let { text = '', as = 'button', className = '', ...rest } = $props();
</script>

{#if as === 'button'}
  <button class={'track-badge ' + className} {...rest}>{text}</button>
{:else}
  <span class={'track-badge ' + className} {...rest}>{text}</span>
{/if}

<style>
  .track-badge {
    /*display: inline-flex !important;*/
    display: inline-block;
    align-items: flex-start !important; /* vertical alignment inside the button */
    justify-content: flex-start !important; /* horizontal alignment */
    text-align: left !important; /* ensure multi-line text is left-aligned */
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  /* Variant styles when parent passes variant classes (e.g. track-accepted or track-reviewed) */
  .track-badge.track-accepted {
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  .track-badge.track-reviewed {
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }

  :global(.dark) .track-badge {
    background-color: #14532d !important;
    color: #bbf7d0 !important;
  }
</style>
