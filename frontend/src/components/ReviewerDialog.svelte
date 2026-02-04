<script>
  import { Modal } from 'flowbite-svelte';
  import { CloseOutline } from 'flowbite-svelte-icons';
  import ReviewerCard from './ReviewerCard.svelte';
  import AffiliationDialog from './AffiliationDialog.svelte';

  let { open = $bindable(false), reviewer = null } = $props();

  // Local state for nested affiliation dialog
  let affiliationOpen = $state(false);
  let affiliation = $state(null);

  function closeDialog() {
    open = false;
  }

  function handleAffiliationClick(aff) {
    affiliation = aff;
    affiliationOpen = true;
  }
</script>

<Modal bind:open size="md" dismissable={false} class="max-w-xl mx-auto">
  <div class="flex justify-between items-start mb-4">
    <div class="flex items-center gap-2">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Reviewer Details</h3>
    </div>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <CloseOutline class="w-5 h-5" />
    </button>
  </div>

  {#if reviewer}
    <div class="space-y-4">
      <ReviewerCard reviewer={reviewer} showEmail={true} onAffiliationClick={handleAffiliationClick} />

      <div class="pt-2 border-t border-gray-200 dark:border-gray-700">
        <div class="text-sm text-gray-600 dark:text-gray-400">Identifier: {reviewer.identifier || 'N/A'}</div>
        {#if reviewer.id}
          <div class="text-sm text-gray-600 dark:text-gray-400">ID: {reviewer.id}</div>
        {/if}
      </div>
    </div>
  {:else}
    <div class="text-center py-6 text-gray-500 dark:text-gray-400">No reviewer selected</div>
  {/if}
</Modal>

<!-- Nested affiliation dialog -->
<AffiliationDialog bind:open={affiliationOpen} affiliation={affiliation} />
