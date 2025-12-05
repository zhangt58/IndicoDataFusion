<script>
  import AbstractCardItem from './AbstractCardItem.svelte';
  import { VirtualList } from 'svelte-virtuallists';

  /** @type {Array} */
  export let abstractData = [];
  // Height of the virtual list container. Can be overridden by parent.
  // Keep default as '100vh' for full viewport fill; parent can provide 'calc(100vh - 64px)' if there's a header.
  export let listHeight = '100vh';
</script>

{#if abstractData && abstractData.length > 0}
  <!-- Full-viewport container: makes the VirtualList fill the viewport exactly -->
  <div class="space-y-4 mt-8" style={`height:${listHeight}; display:flex; flex-direction:column;`}>
    <VirtualList items={abstractData} style={`width:100%;height:100%;`}>
      {#snippet vl_slot({ index, item })}
        <div class="mb-4">
          <AbstractCardItem abstract={item} />
        </div>
      {/snippet}
    </VirtualList>
  </div>
{:else}
  <div class="p-4 text-center text-slate-500">No abstracts to display.</div>
{/if}
