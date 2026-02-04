<script>
  import { Modal } from 'flowbite-svelte';
  import ReviewerCard from './ReviewerCard.svelte';
  import AffiliationDialog from './AffiliationDialog.svelte';
  import Icon from '@iconify/svelte';
  import { OpenSafeURL } from '../../wailsjs/go/main/App';

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

  // Derived dashboard URL from reviewer.avatar_url (nullable)
  const dashboardURL = $derived(() => {
    if (!reviewer) return null;
    const av = reviewer.avatar_url;
    if (!av) return null;
    try {
      const u = new URL(av, window.location.origin);
      // find '/user' in the pathname
      const idx = u.pathname.indexOf('/user');
      let basePath = u.origin;
      if (idx !== -1) {
        basePath += u.pathname.slice(0, idx);
      }
      if (!reviewer.id) return null;
      return basePath + '/user/' + reviewer.id + '/dashboard';
    } catch (e) {
      return null;
    }
  });

  async function openDashboard() {

    const url = dashboardURL();
    console.log(url);
    if (!url) return;
    try {
      await OpenSafeURL(url);
    } catch (e) {
      console.error('OpenSafeURL failed', e);
    }
  }
</script>

<Modal bind:open size="md" dismissable={false} class="max-w-xl mx-auto">
  <div class="flex justify-between items-start mb-2">
    <div class="flex items-center gap-2">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Reviewer Details</h3>
    </div>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <Icon icon="mdi:close" class="w-5 h-5" />
    </button>
  </div>

  {#if reviewer}
    <div class="space-y-4">
      <ReviewerCard reviewer={reviewer} showEmail={true} onAffiliationClick={handleAffiliationClick} />

      <div class="pt-2 border-t border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
          {#if reviewer.identifier}
            <div class="truncate font-mono">Identifier: {reviewer.identifier}</div>
          {/if}

          {#if dashboardURL()}
            <button
              type="button"
              class="ml-auto hover:text-blue-500 hover:dark:text-blue-100 inline-flex items-center gap-1 cursor-pointer"
              onclick={openDashboard}
              aria-label="Open reviewer dashboard in browser"
              title="Open reviewer dashboard in browser"
            >
              <Icon icon="mdi:open-in-new" class="w-4 h-4" />
            </button>
          {/if}
        </div>
      </div>
    </div>
  {:else}
    <div class="text-center py-6 text-gray-500 dark:text-gray-400">No reviewer selected</div>
  {/if}
</Modal>

<!-- Nested affiliation dialog -->
<AffiliationDialog bind:open={affiliationOpen} affiliation={affiliation} />
