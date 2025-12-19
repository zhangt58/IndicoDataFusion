<script>
  import { Modal } from 'flowbite-svelte';
  import { CloseOutline } from 'flowbite-svelte-icons';
  import ContributionCardItem from './ContributionCardItem.svelte';

  let { open = $bindable(false), contribution = null } = $props();

  // Close dialog
  function closeDialog() {
    open = false;
  }
</script>

<Modal bind:open size="xl" dismissable={false} class="contribution-dialog">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Contribution Details</h3>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <CloseOutline class="shrink-0 h-6 w-6" />
    </button>
  </div>
  {#if contribution}
    <ContributionCardItem {contribution} />
  {/if}
</Modal>
