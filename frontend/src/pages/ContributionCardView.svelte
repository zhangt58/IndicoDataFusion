<script>
  import ContributionCardItem from './ContributionCardItem.svelte';
  import { VirtualList } from 'svelte-virtuallists';

  /** @type {Array} */
  export let contributionData = [];
  // Height for the virtual list container. Default to full viewport.
  export let listHeight = '100vh';
</script>

{#if contributionData && contributionData.length > 0}
  <div class="space-y-4 mt-8" style={`height:${listHeight}; display:flex; flex-direction:column;`}>
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
