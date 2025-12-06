<script>
  export let text = '';
  export let as = 'button';
  export let className = '';
</script>

{#if as === 'button'}
  <button class={"track-badge " + className} {...$$restProps}>{text}</button>
{:else}
  <span class={"track-badge " + className} {...$$restProps}>{text}</span>
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

    :global(.dark .track-badge) {
        background-color: #14532d !important;
        color: #bbf7d0 !important;
    }
</style>