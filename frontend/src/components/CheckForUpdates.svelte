<script>
  import { CheckForNewRelease, OpenSafeURL } from '../../wailsjs/go/main/App';
  import { Modal } from 'flowbite-svelte';

  let checking = $state(false);
  let showModal = $state(false);
  let modalTitle = $state('');
  let modalBody = $state('');
  let modalURL = $state('');

  async function checkUpdates() {
    checking = true;
    showModal = false;
    modalTitle = '';
    modalBody = '';
    modalURL = '';
    try {
      if (!(await CheckForNewRelease())) {
        modalTitle = 'Update Check';
        modalBody = 'No release information returned.';
        modalURL = '';
      } else if ((await CheckForNewRelease()).isNew) {
        modalTitle = 'New Release Available';
        modalBody = `Latest: ${(await CheckForNewRelease()).latestTag}\n\n${(await CheckForNewRelease()).body || ''}`;
        modalURL = (await CheckForNewRelease()).htmlURL || '';
      } else {
        modalTitle = 'Up to Date';
        modalBody = `You are running the latest version.`;
        modalURL = (await CheckForNewRelease()).htmlURL || '';
      }
    } catch (e) {
      const errMsg = e && e.message ? e.message : String(e);
      modalTitle = 'Update Check Failed';
      modalBody = String(errMsg);
      modalURL = '';
    } finally {
      checking = false;
      showModal = true;
    }
  }

  function openURL(u) {
    try {
      OpenSafeURL(u);
    } catch (e) {
      // ignore
    }
  }
</script>

<div class="flex items-center gap-2">
  <button
    type="button"
    onclick={checkUpdates}
    class="text-xs px-2 py-0.5 rounded bg-sky-100 text-sky-700 hover:bg-sky-200 dark:bg-sky-900 dark:text-sky-300 dark:hover:bg-sky-800 border border-sky-200 dark:border-sky-700 focus:outline-none focus:ring-1 focus:ring-sky-200"
    disabled={checking}
    title="Check for new release"
  >
    {#if checking}
      Checking...
    {:else}
      Check
    {/if}
  </button>
</div>

<!-- Modal inside the component -->
<Modal bind:open={showModal} size="sm" onclose={() => (showModal = false)}>
  <div class="p-2">
    <h3 class="text-md font-semibold">{modalTitle}</h3>
    <div class="mt-1 whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">{modalBody}</div>
    <div class="mt-2 flex justify-end gap-2">
      {#if modalURL}
        <button
          type="button"
          onclick={() => openURL(modalURL)}
          class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700">Open Release</button
        >
      {/if}
      <button
        type="button"
        onclick={() => (showModal = false)}
        class="px-3 py-1 rounded bg-indigo-600 text-white">Close</button
      >
    </div>
  </div>
</Modal>
