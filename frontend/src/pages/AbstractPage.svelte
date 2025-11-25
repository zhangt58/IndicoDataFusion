<script>
  import { onMount } from 'svelte';
  import { GetAbstractData } from '../../wailsjs/go/backend/IndicoClient';

  let loading = false;
  let abstractData = [];
  let error = null;

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
  <section class="space-y-4">
    <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-200 mb-4">Abstracts</h2>
    {#each abstractData as abstract}
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
        <h3 class="text-xl font-bold text-gray-800 dark:text-white mb-2">{abstract.title}</h3>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">By: {abstract.author}</p>
        <p class="text-gray-700 dark:text-gray-300 mb-3">{abstract.description}</p>
        <div class="flex gap-2 flex-wrap">
          {#each abstract.keywords as keyword}
            <span class="px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full text-sm">{keyword}</span>
          {/each}
        </div>
      </div>
    {/each}
  </section>
{/if}
