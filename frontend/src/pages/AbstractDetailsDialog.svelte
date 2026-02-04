<script>
  import { Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import AbstractCardItem from './AbstractCardItem.svelte';

  let {
    open = $bindable(false),
    abstract = $bindable(null),
    isMyReview = false
  } = $props();

  // Close dialog
  function closeDialog() {
    open = false;
  }

  // Handle refresh from AbstractCardItem - ensure the bindable abstract is updated
  function handleRefresh(refreshedAbstract) {
    // Update the bindable prop to trigger reactivity up the chain
    abstract = refreshedAbstract;
  }
</script>

<Modal bind:open size="xl" dismissable={false}>
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Abstract Details</h3>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <Icon icon="mdi:close" class="shrink-0 h-6 w-6" />
    </button>
  </div>
  {#if abstract}
    <AbstractCardItem bind:abstract onRefresh={handleRefresh} isMyReview={isMyReview} />
  {/if}
</Modal>
