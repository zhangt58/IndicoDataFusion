<script>
  import AbstractCardItem from './AbstractCardItem.svelte';

  /** @type {Array} */
  export let abstractData = [];
</script>

<section class="space-y-4 mt-8">
  {#each abstractData as abstract}
    <AbstractCardItem {abstract} />
  {/each}
</section>
