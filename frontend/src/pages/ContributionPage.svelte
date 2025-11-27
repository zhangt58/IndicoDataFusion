<script>
  import { onMount } from 'svelte';
  import { GetContributions } from '../../wailsjs/go/main/App';
  import { LayoutGrid, CreditCard } from '@lucide/svelte';
  import ContributionCardView from './ContributionCardView.svelte';
  import ContributionTableView from './ContributionTableView.svelte';

  let loading = false;
  let contributionData = [];
  let error = null;
  let viewMode = 'card'; // 'card' or 'table'

  onMount(async () => {
    loading = true;
    try {
      contributionData = (await GetContributions()) || [];
    } catch (e) {
      console.error('GetContributions failed', e);
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
  <div class="p-6 text-center text-red-600">Failed to load contributions: {error}</div>
{:else}
  <div class="fixed bg-indigo-300 dark:bg-indigo-800 top-2 left-2 shadow-md px-2 py-1 rounded-sm flex items-center gap-2 z-10">
    <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">Contributions ({contributionData.length})</h2>
    <div class="flex gap-1 ml-2">
      <button
        onclick={() => viewMode = 'card'}
        class="p-1.5 rounded transition-colors {viewMode === 'card' ? 'bg-indigo-400' : 'hover:bg-indigo-100'}"
        title="Card View"
      >
        <CreditCard size={18} />
      </button>
      <button
        onclick={() => viewMode = 'table'}
        class="p-1.5 rounded transition-colors {viewMode === 'table' ? 'bg-indigo-400' : 'hover:bg-indigo-100'}"
        title="Table View"
      >
        <LayoutGrid size={18} />
      </button>
    </div>
  </div>

  {#if viewMode === 'table'}
    <ContributionTableView {contributionData} />
  {:else}
    <ContributionCardView {contributionData} />
  {/if}
{/if}
