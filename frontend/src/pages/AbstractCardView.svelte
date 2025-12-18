<script>
  import AbstractCardItem from './AbstractCardItem.svelte';
  import { VirtualList } from 'svelte-virtuallists';

  let { abstractData = [] } = $props();
</script>

{#if abstractData && abstractData.length > 0}
  <!-- Full-viewport container: makes the VirtualList fill the viewport exactly -->
  <div class="flex flex-col h-screen" style="height: calc(100vh - 8rem);">
    <VirtualList items={abstractData}>
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
