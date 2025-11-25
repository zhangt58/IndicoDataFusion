<script>
  import { onMount } from 'svelte';
  import { GetAbstractData } from '../../wailsjs/go/backend/IndicoClient';
  import { LayoutGrid, CreditCard } from '@lucide/svelte';
  import AbstractCardView from './AbstractCardView.svelte';
  import AbstractTableView from './AbstractTableView.svelte';

  let loading = false;
  let abstractData = [];
  let error = null;
  let viewMode = 'card'; // 'card' or 'table'

  onMount(async () => {
    loading = true;
    try {
      abstractData = (await GetAbstractData()) || [];
    } catch (e) {
      console.error('GetAbstractData failed', e);
      abstractData = [];
      error = e;
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="p-6 text-center">Loading abstracts...</div>
{:else if error}
  <div class="p-6 text-center text-red-600">Failed to load abstracts.</div>
{:else}
  <div class="fixed bg-sky-300 top-2 left-2 shadow-md px-2 py-1 rounded-sm flex items-center gap-2 z-10">
    <h2 class="text-xl font-semibold text-gray-700 dark:text-gray-50">Abstracts ({abstractData.length})</h2>
    <div class="flex gap-1 ml-2">
      <button
        onclick={() => viewMode = 'card'}
        class="p-1.5 rounded transition-colors {viewMode === 'card' ? 'bg-sky-500 text-white' : 'bg-white text-gray-600 hover:bg-sky-100'}"
        title="Card View"
      >
        <CreditCard size={18} />
      </button>
      <button
        onclick={() => viewMode = 'table'}
        class="p-1.5 rounded transition-colors {viewMode === 'table' ? 'bg-sky-500 text-white' : 'bg-white text-gray-600 hover:bg-sky-100'}"
        title="Table View"
      >
        <LayoutGrid size={18} />
      </button>
    </div>
  </div>

  {#if viewMode === 'table'}
    <AbstractTableView {abstractData} />
  {:else}
    <AbstractCardView {abstractData} />
  {/if}
{/if}
