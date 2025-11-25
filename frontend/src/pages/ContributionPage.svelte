<script>
  import { onMount } from 'svelte';
  import { GetContributionData } from '../../wailsjs/go/backend/IndicoClient';

  let loading = false;
  let contributionData = [];
  let error = null;

  onMount(async () => {
    loading = true;
    try {
      contributionData = (await GetContributionData()) || [];
    } catch (e) {
      console.error('GetContributionData failed', e);
      contributionData = [];
      error = e;
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="p-6 text-center">Loading contributions...</div>
{:else if error}
  <div class="p-6 text-center text-red-600">Failed to load contributions.</div>
{:else}
  <h2 class="fixed bg-indigo-300 top-2 left-2 shadow-md px-2 py-1 rounded-sm text-xl font-semibold text-gray-700 dark:text-gray-200">Contributions ({contributionData.length})</h2>
  <section class="space-y-4 mt-8">
    {#each contributionData as contribution}
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
        <h3 class="text-xl font-bold text-gray-800 dark:text-white mb-2">{contribution.title}</h3>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">Contributor: {contribution.contributor}</p>
        <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
          <span>Type: {contribution.type}</span>
          <span>Submitted: {contribution.submittedAt}</span>
          <span class="ml-4">Status: {contribution.status}</span>
        </div>
      </div>
    {/each}
  </section>
{/if}
