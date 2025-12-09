<script>
  import ContributionCardItem from './ContributionCardItem.svelte';
  import { VirtualList } from 'svelte-virtuallists';

  /** @type {Array} */
  export let contributionData = [];
</script>

{#if contributionData && contributionData.length > 0}
  <div class="flex flex-col mt-8 h-screen" style="height: calc(100vh - 8rem);">
    <VirtualList items={contributionData} style={`width:100%;height:100%;`}>
      {#snippet vl_slot({ index, item })}
        <div class="mb-4">
          <ContributionCardItem contribution={item} />
        </div>
      {/snippet}
    </VirtualList>
  </div>
{:else}
  <div class="p-4 text-center text-slate-500">No contributions to display.</div>
{/if}
