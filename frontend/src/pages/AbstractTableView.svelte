<script>
  import { DataTable, DataTableControls } from '@zhangt58/svelte-vtable';
  import TypeBadge from './TypeBadge.svelte';
  import AbstractDetailsDialog from './AbstractDetailsDialog.svelte';
  import TrackDetailsDialog from './TrackDetailsDialog.svelte';
  import TitleLink from '../components/TitleLink.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import {
    getTableItems, 
  } from './AbstractTableItem.js';

  let { abstractData = [] } = $props();

  // Abstract dialog state
  let showAbstractDialog = $state(false);
  let selectedAbstract = $state(null);

  // Track dialog state
  let showTrackDialog = $state(false);
  let selectedTracks = $state([]);

  // Simple client-side controls (search/sort/pagination)
  // We will use the event-based API like the example: DataTableControls emits pagechange/searchchange
  let searchQuery = $state('');
  let perPage = $state(25);
  let currentPage = $state(1);
  let sortKey = $state(null); // e.g. 'Title' or 'Score'
  let sortDir = $state('asc'); // 'asc' | 'desc'

  // Column filters state (added) - map of { headerTitle: [selectedValues] }
  let activeFilters = $state({});

  // Columns definition (id matches keys in tableItems)
  const columns = [
    { id: 'ID', title: 'ID', stretch: 1 },
    { id: 'Title', title: 'Title', stretch: 6 },
    { id: 'State', title: 'State', stretch: 2 },
    { id: 'Submitter', title: 'Submitter', stretch: 3 },
    { id: 'Affiliation', title: 'Affiliation', stretch: 3 },
    { id: 'Track', title: 'Track', stretch: 2 },
    { id: 'Type', title: 'Type', stretch: 1 },
    { id: 'Score', title: 'Score', stretch: 1 },
    { id: 'Submitted', title: 'Submitted', stretch: 2 },
    { id: 'Authors', title: 'Authors', stretch: 2 }
  ];

  // Map of visible column keys in the data objects (header titles)
  const visibleKeys = columns.map(c => c.title);

  // Build mappedColumns for rowSnippet rendering
  const mappedColumns = columns.map(c => ({ id: c.id, title: c.title, nowrap: false, stretch: c.stretch }));

  // Build column widths mapping to pass to DataTable (using stretch weights)
  const colWidths = mappedColumns.reduce((acc, c) => {
    acc[c.title] = c.stretch;
    return acc;
  }, {});

  // Collect all unique tracks from all abstracts
  let allAvailableTracks = $derived(abstractData.reduce((acc, abstract) => {
    if (abstract.accepted_track && !acc.some(t => t.title === abstract.accepted_track.title)) {
      acc.push({ title: abstract.accepted_track.title, type: 'accepted' });
    }
    if (abstract.reviewed_for_tracks) {
      abstract.reviewed_for_tracks.forEach(track => {
        if (!acc.some(t => t.title === track.title)) {
          acc.push({ title: track.title, type: 'reviewed' });
        }
      });
    }
    return acc;
  }, []));

  // Find abstract by ID
  function findAbstractById(id) {
    // Use strict string comparison to avoid type coercion warnings
    const sid = String(id);
    return abstractData.find(a => String(a.friendly_id || a.id) === sid);
  }

  // Open abstract details by id (used by title button)
  function openAbstract(id) {
    const sid = String(id);
    selectedAbstract = findAbstractById(sid);
    if (selectedAbstract) showAbstractDialog = true;
  }

  // Open track dialog from TrackFull data (may be JSON string or array)
  function openTrack(trackFull) {
    try {
      if (!trackFull) {
        selectedTracks = [];
        return;
      }
      const parsed = typeof trackFull === 'string' ? JSON.parse(trackFull) : trackFull;
      selectedTracks = Array.isArray(parsed) ? parsed : [parsed];
      if (selectedTracks.length > 0) showTrackDialog = true;
    } catch (e) {
      console.error('Failed to parse trackFull:', e);
      selectedTracks = [];
    }
  }

  // Build table items and options
  let tableItems = $derived(getTableItems(abstractData));

  // columnFilters derived from tableItems (for DataTableControls/DataTableFilters)
  function getUniqueValuesWithCounts(items, header) {
    const counts = {};
    (items || []).forEach(it => {
      const val = it && it[header];
      if (Array.isArray(val)) {
        val.forEach(v => { const s = String(v ?? ''); counts[s] = (counts[s] || 0) + 1; });
      } else {
        const s = String(val ?? ''); counts[s] = (counts[s] || 0) + 1;
      }
    });
    const uniqueValues = Object.keys(counts).sort();
    return { uniqueValues, counts };
  }

  let columnFilters = $derived(columns.map(c => {
    const { uniqueValues, counts } = getUniqueValuesWithCounts(tableItems || [], c.title);
    return { key: c.title, label: c.title, uniqueValues, counts };
  }));

  // Handle filter changes from DataTableControls/DataTableFilters
  function handleFilterChange({ allFilters }) {
    activeFilters = { ...allFilters };
    currentPage = 1;
  }

  // Helper to check if an item matches activeFilters
  function matchesFilters(item, filters) {
    for (const [columnKey, selectedValues] of Object.entries(filters)) {
      if (!selectedValues || selectedValues.length === 0) continue;
      const itemValue = item[columnKey];
      if (Array.isArray(itemValue)) {
        const itemStrings = itemValue.map(v => String(v ?? ''));
        if (!selectedValues.some(val => itemStrings.includes(val))) return false;
      } else {
        const itemStr = String(itemValue ?? '');
        if (!selectedValues.includes(itemStr)) return false;
      }
    }
    return true;
  }

  // Filtering (apply active column filters first, then search)
  let filteredItems = $derived(tableItems.filter(item => {
    if (Object.keys(activeFilters).length > 0) {
      if (!matchesFilters(item, activeFilters)) return false;
    }
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return visibleKeys.some(k => String(item[k] ?? '').toLowerCase().includes(q));
  }));

  // Sorting helper (handles ID numeric, Score numeric, Submitted timestamp, and Track numeric extraction)
  function compare(a,b,key) {
    const va = a[key];
    const vb = b[key];

    // Special-case ID: numeric sort using IDNumber
    if (key === 'ID') {
      const na = a.IDNumber != null ? Number(a.IDNumber) : NaN;
      const nb = b.IDNumber != null ? Number(b.IDNumber) : NaN;
      if (isNaN(na) && isNaN(nb)) return 0;
      if (isNaN(na)) return -1;
      if (isNaN(nb)) return 1;
      return na - nb;
    }

    // numeric sort for Score
    if (key === 'Score') {
      const na = Number(va === '' ? NaN : va);
      const nb = Number(vb === '' ? NaN : vb);
      if (isNaN(na) && isNaN(nb)) return 0;
      if (isNaN(na)) return -1;
      if (isNaN(nb)) return 1;
      return na - nb;
    }

    // Special-case Submitted: sort by SubmittedMillis timestamp
    if (key === 'Submitted') {
      const ta = a.SubmittedMillis != null ? Number(a.SubmittedMillis) : NaN;
      const tb = b.SubmittedMillis != null ? Number(b.SubmittedMillis) : NaN;
      if (isNaN(ta) && isNaN(tb)) return 0;
      if (isNaN(ta)) return -1;
      if (isNaN(tb)) return 1;
      return ta - tb;
    }

    // Special-case Track: extract number from strings like "MC10" or "MC10: description"
    if (key === 'Track') {
      const extractTrackNumber = (str) => {
        if (!str) return NaN;
        const match = String(str).match(/\d+/);
        return match ? Number(match[0]) : NaN;
      };
      const na = extractTrackNumber(va);
      const nb = extractTrackNumber(vb);
      if (isNaN(na) && isNaN(nb)) {
        // fallback to string comparison if no numbers found
        const sa = String(va ?? '').toLowerCase();
        const sb = String(vb ?? '').toLowerCase();
        return sa < sb ? -1 : sa > sb ? 1 : 0;
      }
      if (isNaN(na)) return -1;
      if (isNaN(nb)) return 1;
      return na - nb;
    }

    // fallback string compare
    const sa = String(va ?? '').toLowerCase();
    const sb = String(vb ?? '').toLowerCase();
    return sa < sb ? -1 : sa > sb ? 1 : 0;
  }

  let sortedItems = $derived((() => {
    if (!sortKey) return filteredItems;
    const copy = filteredItems.slice();
    copy.sort((a,b) => {
      const res = compare(a,b,sortKey);
      return sortDir === 'asc' ? res : -res;
    });
    return copy;
  })());

  // expose total items for DataTableControls
  let totalItems = $derived(sortedItems.length);

  // Pagination
  let totalPages = $derived(Math.max(1, Math.ceil(sortedItems.length / perPage)));

  $effect(() => {
    if (currentPage > totalPages) {
      currentPage = totalPages;
    }
  });

  let paginatedItems = $derived(sortedItems.slice((currentPage-1)*perPage, currentPage*perPage));

  // For VirtualList we pass the paginatedItems so the visible window is virtualized per page
  let visibleItems = $derived(paginatedItems);

  // sort callback used by DataTable (event-based)
  function onSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
    currentPage = 1;
  }
</script>

<!-- Row snippet for AbstractTableView (moved out of <script>) -->
{#snippet rowSnippet({ item, index, select, selected })}
  <tr
    onclick={() => { try { select && select(); } catch (e) {} }}
    tabindex="0"
    class="cursor-pointer"
    class:selected-row={selected && String(selected.ID) === String(item.ID)}
    aria-selected={selected && String(selected.ID) === String(item.ID)}
  >
    {#each mappedColumns as col}
      <td class={col.nowrap ? 'nowrap' : ''}>
        {#if col.id === 'ID'}
          {item.ID}
        {:else if col.id === 'Title'}
          <TitleLink as="button" onclick={() => openAbstract(item.ID)} data-id={item.ID} data-title={item.Title}>{item.Title}</TitleLink>
        {:else if col.id === 'State'}
          {#if item.State}
            <span class={item.State.toLowerCase() === 'accepted' ? 'state-badge state-accepted' : (item.State.toLowerCase() === 'rejected' ? 'state-badge state-rejected' : 'state-badge state-other')}>{item.State}</span>
          {/if}
        {:else if col.id === 'Track'}
          {#if item.Track}
            <TrackBadge text={item.Track} as="button" className={(item.TrackType === 'accepted' ? 'track-accepted' : 'track-reviewed') + ' track-link'} onclick={() => openTrack(item.TrackFull)} {...{ 'data-tracks': item.TrackFull }} />
          {/if}
        {:else if col.id === 'Type'}
          {#if item.Type}
            <TypeBadge text={item.Type} />
          {/if}
        {:else if col.id === 'Authors'}
          {#if item.Authors}
            <span class="authors-cell" title={item.AuthorsTooltip}>{item.Authors}</span>
          {/if}
        {:else}
          {item[col.id]}
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}

<div class="flex flex-col overflow-auto px-1" style="height:calc(100vh - 8rem);">
  <div class="sticky top-0 z-10 bg-transparent
              px-2 py-2 rounded-md border-gray-200 dark:border-gray-700
              mb-2 mt-2 shrink-0 shadow-md dark:shadow-black/40">
    <DataTableControls
      search={searchQuery}
      currentPage={currentPage}
      bind:perPage={perPage}
      {totalItems}
      pagechange={(payload) => { currentPage = payload.currentPage }}
      searchchange={(payload) => { searchQuery = payload.search }}
      columnFilters={columnFilters}
      activeFilters={activeFilters}
      filterChange={handleFilterChange}
      filtersVisible={false}
    />
  </div>

  <section class="flex-1 overflow-auto flex flex-col max-h-screen min-h-0">
    <DataTable
      items={visibleItems}
      {visibleKeys}
      sortKey={sortKey}
      sortDir={sortDir}
      sortCallback={onSort}
      className="datatable-table w-full mt-0.5 mb-2 overflow-auto min-h-0"
      colWidths={colWidths}
      virtualize={false}
      rowSnippet={rowSnippet}
    />
  </section>
</div>

 <!-- Abstract Detail Dialog -->
 <AbstractDetailsDialog bind:open={showAbstractDialog} abstract={selectedAbstract} />

 <!-- Track Details Dialog -->
 <TrackDetailsDialog bind:open={showTrackDialog} tracks={selectedTracks} allTracks={allAvailableTracks} />

<style>
  /* Component-specific styling for AbstractTableView */
  
  /* State badge styling - specific to AbstractTableView (scoped) */
  .state-badge {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-weight: 500;
  }

  /* State variants (scoped to this component) */
  .state-accepted {
    background-color: #dcfce7;
    color: #166534;
  }

  .state-rejected {
    background-color: #fee2e2;
    color: #991b1b;
  }

  .state-other {
    background-color: #fef3c7;
    color: #92400e;
  }

  /* Authors cell with tooltip (scoped) */
  .authors-cell {
    cursor: help;
  }

  /* Other table/badge helpers moved to shared CSS */
</style>
