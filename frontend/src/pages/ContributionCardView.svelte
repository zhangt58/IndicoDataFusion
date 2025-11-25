<script>
  import ContributionCardItem from './ContributionCardItem.svelte';

  /** @type {Array} */
  export let contributionData = [];
</script>

<section class="space-y-4 mt-8">
  {#each contributionData as contribution}
    <ContributionCardItem {contribution} />
  {/each}
</section>
