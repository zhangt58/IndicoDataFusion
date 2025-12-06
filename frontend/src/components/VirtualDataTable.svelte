<script>
  import { createEventDispatcher } from 'svelte';
  import { VirtualList } from 'svelte-virtuallists';

  export let items = [];
  export let visibleKeys = [];
  export let sortKey = null;
  export let sortDir = 'asc';
  export let className = '';
  export let style = '';
  export let emptyMessage = 'No items to display.';

  const dispatch = createEventDispatcher();

  function handleSort(key) {
    dispatch('sort', key);
  }
</script>

<section class="virtual-data-table {className}" style={style}>
  {#if items && items.length > 0}
    <VirtualList items={items} isTable class="datatable-table" style="width:100%;height:100%">
      {#snippet header()}
        <thead>
          <tr>
            {#each visibleKeys as key}
              <th class="cursor-pointer select-none" on:click={() => handleSort(key)} aria-sort={sortKey === key ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}>
                <div style="display:inline-flex;align-items:center;gap:0.25rem;">
                  <span>{key}</span>
                  {#if sortKey === key}
                    <span aria-hidden="true">{sortDir === 'asc' ? '▲' : '▼'}</span>
                  {/if}
                </div>
              </th>
            {/each}
          </tr>
        </thead>
      {/snippet}

      {#snippet vl_slot({ index, item })}
        <slot {item} {index} />
      {/snippet}
    </VirtualList>
  {:else}
    <div class="p-4 text-center text-slate-500">{emptyMessage}</div>
  {/if}
</section>

<style>
  .virtual-data-table { display:block; }
  .datatable-table { width:100%; border-collapse:collapse; }
  .datatable-table thead th { text-align:left; }
</style>

