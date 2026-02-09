<script>
  import { DataTable, DataTableControls } from '@zhangt58/svelte-vtable';
  import Icon from '@iconify/svelte';
  import IndicoConfig from './IndicoConfig.svelte';

  let {
    dataSources = [],
    currentActiveIndex = -1,
    apiTokens = [],
    existingNames = [],
    existingTags = [],
    existingBaseUrls = [],
    onActivate = () => {},
    onDelete = () => {},
    onToggleFavorite = () => {},
    onEditName = () => {},
    onUpdate = () => {},
  } = $props();

  let searchQuery = $state('');
  let perPage = $state(25);
  let currentPage = $state(1);
  let sortKey = $state(null);
  let sortDir = $state('asc');
  let activeFilters = $state({});

  let showEditDialog = $state(false);
  let editingIndex = $state(-1);
  let editingData = $state(null);

  const columns = [
    { id: 'Edit', title: 'Edit', stretch: 0.5, filterable: false },
    { id: 'Name', title: 'Name', stretch: 3 },
    { id: 'Type', title: 'Type', stretch: 1 },
    { id: 'BaseURL', title: 'Base URL', stretch: 3 },
    { id: 'EventID', title: 'ID', stretch: 1 },
    { id: 'APIToken', title: 'Token', stretch: 2 },
    { id: 'Timeout', title: 'Timeout', stretch: 1 },
    { id: 'Description', title: 'Description', stretch: 2 },
    { id: 'Tags', title: 'Tags', stretch: 2 },
    { id: 'Favorite', title: 'Fav', stretch: 0.5 },
    { id: 'Active', title: 'Act', stretch: 0.5 },
    { id: 'Delete', title: 'Del', stretch: 0.5, filterable: false },
  ];

  const visibleKeys = columns.map((c) => c.title);
  const filterableColumns = columns.filter((c) => c.filterable !== false).map((c) => c.id);
  const colWidths = columns.reduce((acc, c) => ({ ...acc, [c.title]: c.stretch }), {});

  // Transform data sources into table rows
  let tableItems = $derived(
    dataSources.map((ds, i) => ({
      _index: i,
      _raw: ds,
      Name: ds.name || `Source ${i}`,
      Type: ds.type === 'indico' ? 'API' : 'TEST',
      BaseURL: ds.indico?.baseUrl || ds.test?.dataDir || '',
      EventID: ds.indico?.eventId?.toString() || '',
      APIToken: ds.indico?.apiTokenName || '',
      Timeout: ds.indico?.timeout || '',
      Description: ds.description || '',
      Tags: ds.tags || [],
      Favorite: !!ds.favorite,
      Active: currentActiveIndex === i,
    })),
  );

  // Prepare column filters for DataTableFilters component
  let columnFilters = $derived(() => {
    const filters = [];

    for (const col of columns) {
      if (!filterableColumns.includes(col.id)) continue;

      const uniqueValues = new Set();
      const counts = {};
      tableItems.forEach((item) => {
        const value = item[col.id];
        if (Array.isArray(value)) {
          value.forEach((v) => {
            uniqueValues.add(v);
            counts[v] = (counts[v] || 0) + 1;
          });
        } else if (value !== undefined && value !== null && value !== '') {
          const s = String(value);
          uniqueValues.add(s);
          counts[s] = (counts[s] || 0) + 1;
        }
      });

      if (uniqueValues.size > 0) {
        filters.push({
          key: col.id,
          label: col.title,
          uniqueValues: Array.from(uniqueValues).sort(),
          counts: counts,
        });
      }
    }

    return filters;
  });

  // Filter by search query and active filters
  let filteredItems = $derived(
    tableItems.filter((it) => {
      // apply active column filters first
      if (Object.keys(activeFilters).length > 0) {
        for (const [columnKey, selectedValues] of Object.entries(activeFilters)) {
          if (!selectedValues || selectedValues.length === 0) continue;
          const itemValue = it[columnKey];
          if (Array.isArray(itemValue)) {
            // For array values (like Tags), check if any match the selected values
            if (!itemValue.some((val) => selectedValues.includes(val))) {
              return false;
            }
          } else {
            // For single values, check for direct match
            if (!selectedValues.includes(String(itemValue))) {
              return false;
            }
          }
        }
      }

      // Search filter
      if (searchQuery) {
        const q = searchQuery.toLowerCase();
        const searchMatch = [
          it.Name,
          it.Type,
          it.Description,
          it.BaseURL,
          it.EventID,
          it.Tags.join(' '),
        ].some((val) => String(val).toLowerCase().includes(q));
        if (!searchMatch) return false;
      }

      return true;
    }),
  );

  // Sort items
  let sortedItems = $derived.by(() => {
    if (!sortKey) return filteredItems;

    const sorted = [...filteredItems].sort((a, b) => {
      const va = a[sortKey] ?? '';
      const vb = b[sortKey] ?? '';
      if (typeof va === 'string' && typeof vb === 'string') return va.localeCompare(vb);
      return va < vb ? -1 : va > vb ? 1 : 0;
    });

    return sortDir === 'desc' ? sorted.reverse() : sorted;
  });

  // Pagination
  let totalItems = $derived(sortedItems.length);
  let visibleItems = $derived(
    sortedItems.slice((currentPage - 1) * perPage, currentPage * perPage),
  );

  function setSort(key) {
    sortDir = sortKey === key && sortDir === 'asc' ? 'desc' : 'asc';
    sortKey = key;
  }

  function openEditDialog(index) {
    editingIndex = index;
    const ds = dataSources[index];
    if (!ds || ds.type !== 'indico') return;

    // Prepare initial data for IndicoConfig
    editingData = {
      name: ds.name || '',
      baseUrl: ds.indico?.baseUrl || '',
      eventId: ds.indico?.eventId || 0,
      apiTokenName: ds.indico?.apiTokenName || '',
      apiToken: ds.indico?.apiTokenName || '',
      timeout: ds.indico?.timeout || '60s',
      description: ds.description || '',
      tags: ds.tags || [],
      favorite: ds.favorite || false,
    };

    showEditDialog = true;
  }

  function handleEditSave(event) {
    const payload = event.detail || event;
    if (editingIndex < 0) return;

    const ds = dataSources[editingIndex];
    const updatedSource = {
      ...ds,
      name: payload.name?.trim() || ds.name,
      description: payload.description || ds.description || '',
      tags: payload.tags || ds.tags || [],
      favorite: payload.favorite !== undefined ? payload.favorite : ds.favorite,
    };

    if (updatedSource.type === 'indico') {
      updatedSource.indico = {
        baseUrl: payload.baseUrl?.trim() || ds.indico?.baseUrl || '',
        eventId: payload.eventId ? parseInt(String(payload.eventId), 10) : ds.indico?.eventId || 0,
        apiTokenName: payload.apiTokenName || payload.apiToken || ds.indico?.apiTokenName || '',
        timeout: payload.timeout || ds.indico?.timeout || '60s',
      };
    }

    onUpdate(editingIndex, updatedSource);
    showEditDialog = false;
    editingIndex = -1;
  }

  function handleEditCancel() {
    showEditDialog = false;
    editingIndex = -1;
    editingData = null;
  }
</script>

<div class="flex flex-col overflow-auto" style="height:calc(100vh - 10rem);">
  <div class="sticky top-0 z-10 bg-transparent px-2 py-2 mb-2 mt-2">
    <DataTableControls
      search={searchQuery}
      {currentPage}
      bind:perPage
      {totalItems}
      pagechange={(payload) => (currentPage = payload.currentPage)}
      searchchange={(payload) => (searchQuery = payload.search)}
      columnFilters={columnFilters()}
      {activeFilters}
      filterChange={({ allFilters }) => {
        activeFilters = { ...allFilters };
        currentPage = 1;
      }}
      filtersVisible={false}
    />
  </div>

  <section class="flex-1 overflow-auto flex flex-col min-h-0">
    <DataTable
      items={visibleItems}
      {visibleKeys}
      {sortKey}
      {sortDir}
      sortCallback={setSort}
      className="w-full mt-0.5 mb-4 overflow-auto min-h-0"
      {colWidths}
      virtualize={false}
      {rowSnippet}
    />
  </section>
</div>

{#snippet rowSnippet({ item, select, selected })}
  <tr
    onclick={() => select?.()}
    tabindex="0"
    class:selected-row={selected?._index === item._index}
    aria-selected={selected?._index === item._index}
  >
    {#each columns as col}
      <td>
        {#if col.id === 'Edit'}
          <button
            type="button"
            class="p-1 rounded focus:outline-none text-indigo-500 hover:text-indigo-700"
            onclick={(e) => {
              e.stopPropagation();
              openEditDialog(item._index);
            }}
            title="Edit data source"
          >
            <Icon icon="mdi:pencil" class="w-5 h-5" />
          </button>
        {:else if col.id === 'Name'}
          <input
            type="text"
            value={item.Name}
            oninput={(e) => {
              const target = e.target;
              if (target && 'value' in target) {
                onEditName(item._index, target.value);
              }
            }}
            class="w-full bg-transparent border-b border-transparent focus:border-indigo-400 px-1 py-0.5 text-sm"
          />
        {:else if col.id === 'Type'}
          <span
            class="px-2 py-0.5 text-xs rounded-full bg-gray-100 dark:bg-gray-900 text-gray-800 dark:text-gray-200"
          >
            {item.Type}
          </span>
        {:else if col.id === 'BaseURL'}
          <span class="text-xs font-mono" title={item.BaseURL}>{item.BaseURL}</span>
        {:else if col.id === 'EventID'}
          <span class="text-xs">{item.EventID}</span>
        {:else if col.id === 'APIToken'}
          <span class="font-mono text-sm">{item.APIToken}</span>
        {:else if col.id === 'Timeout'}
          <span>{item.Timeout}</span>
        {:else if col.id === 'Description'}
          <span title={item.Description}>{item.Description}</span>
        {:else if col.id === 'Tags'}
          {#if item.Tags.length > 0}
            <div class="flex gap-1 flex-wrap">
              {#each item.Tags as tag}
                <span class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">{tag}</span>
              {/each}
            </div>
          {/if}
        {:else if col.id === 'Favorite'}
          <button
            type="button"
            class="p-1 rounded focus:outline-none"
            onclick={(e) => {
              e.stopPropagation();
              onToggleFavorite(item._raw);
            }}
            aria-pressed={item.Favorite}
            title={item.Favorite ? 'Unmark favorite' : 'Mark favorite'}
          >
            <Icon
              icon={item.Favorite ? 'mdi:star' : 'mdi:star-outline'}
              class="w-5 h-5 text-yellow-500"
            />
          </button>
        {:else if col.id === 'Active'}
          <button
            type="button"
            class="p-1 rounded focus:outline-none"
            onclick={(e) => {
              e.stopPropagation();
              onActivate(item._index);
            }}
            aria-pressed={item.Active}
            title={item.Active ? 'Active data source' : 'Set as active'}
          >
            <Icon
              icon={item.Active
                ? 'mdi:checkbox-marked-circle'
                : 'mdi:checkbox-blank-circle-outline'}
              class="w-5 h-5 text-indigo-500"
            />
          </button>
        {:else if col.id === 'Delete'}
          <button
            type="button"
            class="text-red-500 hover:text-red-700 p-1 rounded focus:outline-none"
            onclick={(e) => {
              e.stopPropagation();
              onDelete(item._index);
            }}
            title="Delete data source"
          >
            <Icon icon="mdi:delete" class="w-5 h-5" />
          </button>
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}

{#if showEditDialog && editingIndex >= 0 && editingData}
  <IndicoConfig
    bind:open={showEditDialog}
    existingNames={existingNames.filter((_, i) => i !== editingIndex)}
    {existingTags}
    {existingBaseUrls}
    {apiTokens}
    initialData={editingData}
    saving={false}
    onCreate={handleEditSave}
    onCancel={handleEditCancel}
  />
{/if}
