<script>
  import { Modal } from 'flowbite-svelte';
  import { CloseOutline } from 'flowbite-svelte-icons';
  import AbstractCardItem from './AbstractCardItem.svelte';

  /** @type {boolean} */
  export let open = false;

  /** @type {Object|null} */
  export let abstract = null;

  // Close dialog
  function closeDialog() {
    open = false;
  }
</script>

<Modal bind:open={open} size="xl" dismissable={false} class="abstract-dialog">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Abstract Details</h3>
    <button 
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      on:click={closeDialog}
    >
      <CloseOutline class="shrink-0 h-6 w-6" />
    </button>
  </div>
  {#if abstract}
    <AbstractCardItem {abstract} />
  {/if}
</Modal>
