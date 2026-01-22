<script>
  import { RefreshAbstractByID } from '../../wailsjs/go/main/App';
  import { DataTable, DataTableControls } from '@zhangt58/svelte-vtable';
  import TypeBadge from './TypeBadge.svelte';
  import AbstractDetailsDialog from './AbstractDetailsDialog.svelte';
  import TrackDetailsDialog from './TrackDetailsDialog.svelte';
  import AffiliationDialog from '../components/AffiliationDialog.svelte';
  import AffiliationBadge from '../components/AffiliationBadge.svelte';
  import TitleButton from '../components/TitleButton.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import StateBadge from './StateBadge.svelte';
  import { getTableItems, getShortTrackName } from './AbstractTableItem.js';

  let { abstractData = $bindable([]) } = $props();

  // Abstract dialog state
  let showAbstractDialog = $state(false);
  let selectedAbstract = $state(null);
  let selectedAbstractId = $state(null);
  let lastSyncedAbstract = $state(null); // Track last synced version to prevent infinite loops

  // Track dialog state
  let showTrackDialog = $state(false);
  let selectedTracks = $state([]);

  // Affiliation dialog state
  let showAffiliationDialog = $state(false);
  let selectedAffiliation = $state(null);

  // Refresh state - track which rows are currently refreshing
  let refreshingIds = $state(new Set());

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
    { id: 'Authors', title: 'Authors', stretch: 2 },
    { id: 'FirstPriority', title: 'First Priority', stretch: 1 },
    { id: 'SecondPriority', title: 'Second Priority', stretch: 1 },
    { id: 'Refresh', title: 'Refresh', stretch: 1 },

    // New columns
    { id: 'AcceptedTrack', title: 'Accepted Track', stretch: 2 },
    { id: 'AcceptedContribType', title: 'Accepted Type', stretch: 2 },
    { id: 'SubmittedContribType', title: 'Submitted Type', stretch: 2 },
    { id: 'ReviewedForTracks', title: 'Reviewed Tracks', stretch: 2 },
    { id: 'SubmittedForTracks', title: 'Submitted Tracks', stretch: 2 },
  ];

  // Map of visible column keys in the data objects (header titles)
  const visibleKeys = columns.map((c) => c.title);

  // Build mappedColumns for rowSnippet rendering
  const mappedColumns = columns.map((c) => ({
    id: c.id,
    title: c.title,
    nowrap: false,
    stretch: c.stretch,
  }));

  // Build column widths mapping to pass to DataTable (using stretch weights)
  const colWidths = mappedColumns.reduce((acc, c) => {
    acc[c.title] = c.stretch;
    return acc;
  }, {});

  // Collect all unique tracks from all abstracts
  let allAvailableTracks = $derived(
    abstractData.reduce((acc, abstract) => {
      if (abstract.accepted_track && !acc.some((t) => t.title === abstract.accepted_track.title)) {
        acc.push({ title: abstract.accepted_track.title, type: 'accepted' });
      }
      if (abstract.reviewed_for_tracks) {
        abstract.reviewed_for_tracks.forEach((track) => {
          if (!acc.some((t) => t.title === track.title)) {
            acc.push({ title: track.title, type: 'reviewed' });
          }
        });
      }
      return acc;
    }, []),
  );

  // Find abstract by ID (database ID)
  function findAbstractById(id) {
    // Use strict string comparison to avoid type coercion warnings
    const sid = String(id);
    return abstractData.find((a) => String(a.id) === sid);
  }

  // Open abstract details by id (used by title button)
  function openAbstract(id) {
    const sid = String(id);
    selectedAbstractId = sid;
    selectedAbstract = findAbstractById(sid);
    lastSyncedAbstract = selectedAbstract; // Initialize with the original
    if (selectedAbstract) {
      showAbstractDialog = true;
    } else {
      console.warn('[AbstractTableView] Abstract not found for ID:', id);
    }
  }

  // Sync changes from dialog back to abstractData array
  // This effect runs when selectedAbstract changes (e.g., after refresh in dialog)
  $effect(() => {
    // Guard: skip if no abstract selected or dialog closed
    if (!selectedAbstract || !selectedAbstractId) {
      lastSyncedAbstract = null;
      return;
    }

    // Prevent infinite loop: only sync if the abstract actually changed
    // (not just the array reference)
    if (lastSyncedAbstract === selectedAbstract) {
      return;
    }

    const index = abstractData.findIndex((a) => String(a.id) === String(selectedAbstractId));

    if (index !== -1) {
      // Update the array element and trigger reactivity
      abstractData[index] = selectedAbstract;
      abstractData = [...abstractData];
      // Track this version to prevent re-syncing on next effect run
      lastSyncedAbstract = selectedAbstract;
    } else {
      console.warn(
        '[AbstractTableView] Could not find abstract in array to update:',
        selectedAbstractId,
      );
    }
  });

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

  // Open affiliation dialog from AffiliationFull data
  function openAffiliation(affiliationFull) {
    try {
      if (!affiliationFull) {
        selectedAffiliation = null;
        return;
      }
      selectedAffiliation =
        typeof affiliationFull === 'string' ? JSON.parse(affiliationFull) : affiliationFull;
      if (selectedAffiliation) showAffiliationDialog = true;
    } catch (e) {
      console.error('Failed to parse affiliationFull:', e);
      selectedAffiliation = null;
    }
  }

  // Handle row refresh
  async function handleRowRefresh(abstractId) {
    if (refreshingIds.has(abstractId)) return;

    refreshingIds = new Set([...refreshingIds, abstractId]);

    try {
      const refreshedAbstract = await RefreshAbstractByID(abstractId);

      if (refreshedAbstract) {
        // Update the abstract in the abstractData array directly
        const index = abstractData.findIndex((a) => a.id === refreshedAbstract.id);

        if (index !== -1) {
          // Create a new array with the updated abstract to trigger reactivity
          abstractData = [
            ...abstractData.slice(0, index),
            refreshedAbstract,
            ...abstractData.slice(index + 1),
          ];
        } else {
          // If not found, add it to the array
          abstractData = [...abstractData, refreshedAbstract];
        }
      }
    } catch (err) {
      alert('Failed to refresh abstract: ' + (err && err.message ? err.message : String(err)));
    } finally {
      const newSet = new Set(refreshingIds);
      newSet.delete(abstractId);
      refreshingIds = newSet;
    }
  }

  // Build table items and options
  let tableItems = $derived(getTableItems(abstractData));

  // columnFilters derived from tableItems (for DataTableControls/DataTableFilters)
  function getUniqueValuesWithCounts(items, key) {
    // Use the table item property key (column id) to extract unique values
    const counts = {};
    (items || []).forEach((it) => {
      const val = it && it[key];
      if (Array.isArray(val)) {
        val.forEach((v) => {
          const s = String(v ?? '');
          counts[s] = (counts[s] || 0) + 1;
        });
      } else {
        const s = String(val ?? '');
        counts[s] = (counts[s] || 0) + 1;
      }
    });
    const uniqueValues = Object.keys(counts).sort();
    return { uniqueValues, counts };
  }

  // Build columnFilters keyed by column id so DataTableFilters maps to the table item fields
  let columnFilters = $derived(
    columns.map((c) => {
      const { uniqueValues, counts } = getUniqueValuesWithCounts(tableItems || [], c.id);
      return { key: c.id, label: c.title, uniqueValues, counts };
    }),
  );

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
        const itemStrings = itemValue.map((v) => String(v ?? ''));
        if (!selectedValues.some((val) => itemStrings.includes(val))) return false;
      } else {
        const itemStr = String(itemValue ?? '');
        if (!selectedValues.includes(itemStr)) return false;
      }
    }
    return true;
  }

  // Filtering (apply active column filters first, then search)
  let filteredItems = $derived(
    tableItems.filter((item) => {
      if (Object.keys(activeFilters).length > 0) {
        if (!matchesFilters(item, activeFilters)) return false;
      }
      if (!searchQuery) return true;
      const q = searchQuery.toLowerCase();
      // Search across actual table item fields (use column ids)
      return mappedColumns.some((col) =>
        String(item[col.id] ?? '')
          .toLowerCase()
          .includes(q),
      );
    }),
  );

  // Sorting helper (handles ID numeric, Score numeric, Submitted timestamp, Track numeric extraction, and priority columns)
  function compare(a, b, key) {
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

    // numeric sort for First Priority and Second Priority
    if (key === 'First Priority' || key === 'Second Priority') {
      let na, nb;
      if (key === 'First Priority') {
        na = Number(a.FirstPriority);
        nb = Number(b.FirstPriority);
      } else {
        na = Number(a.SecondPriority);
        nb = Number(b.SecondPriority);
      }
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

  let sortedItems = $derived(
    (() => {
      if (!sortKey) return filteredItems;
      const copy = filteredItems.slice();
      copy.sort((a, b) => {
        const res = compare(a, b, sortKey);
        return sortDir === 'asc' ? res : -res;
      });
      return copy;
    })(),
  );

  // expose total items for DataTableControls
  let totalItems = $derived(sortedItems.length);

  // Pagination
  let totalPages = $derived(Math.max(1, Math.ceil(sortedItems.length / perPage)));

  $effect(() => {
    if (currentPage > totalPages) {
      currentPage = totalPages;
    }
  });

  let paginatedItems = $derived(
    sortedItems.slice((currentPage - 1) * perPage, currentPage * perPage),
  );

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
    onclick={() => {
      try {
        select && select();
      } catch (e) {}
    }}
    tabindex="0"
    class:selected-row={selected && String(selected.ID) === String(item.ID)}
    aria-selected={selected && String(selected.ID) === String(item.ID)}
  >
    {#each mappedColumns as col}
      <td class={col.nowrap ? 'nowrap' : ''}>
        {#if col.id === 'ID'}
          {item.ID}
        {:else if col.id === 'Title'}
          <TitleButton
            onclick={() => openAbstract(item.DatabaseID)}
            data-id={item.ID}
            data-title={item.Title}>{item.Title}</TitleButton
          >
        {:else if col.id === 'State'}
          {#if item.State}
            <StateBadge state={item.State} />
          {/if}
        {:else if col.id === 'Affiliation'}
          {#if item.Affiliation && item.AffiliationFull}
            {@const affiliationObj = JSON.parse(item.AffiliationFull)}
            <AffiliationBadge
              affiliation={affiliationObj}
              onclick={() => openAffiliation(item.AffiliationFull)}
              showCity={false}
              className="text-gray-600 dark:text-gray-400"
            />
          {:else if item.Affiliation}
            {item.Affiliation}
          {/if}
        {:else if col.id === 'Track'}
          {#if item.Track}
            <TrackBadge
              text={item.Track}
              type={item.TrackType}
              onclick={() => openTrack(item.TrackFull)}
              data-tracks={item.TrackFull}
            />
          {/if}
        {:else if col.id === 'AcceptedTrack'}
          {#if item.AcceptedTrack}
            <TrackBadge text={getShortTrackName(item.AcceptedTrack)} type="accepted" onclick={() => openTrack({ title: item.AcceptedTrack, type: 'accepted' })} />
          {/if}
        {:else if col.id === 'AcceptedContribType'}
          {#if item.AcceptedContribType}
            <TypeBadge text={item.AcceptedContribType} />
          {/if}
        {:else if col.id === 'SubmittedContribType'}
          {#if item.SubmittedContribType}
            <TypeBadge text={item.SubmittedContribType} />
          {/if}
        {:else if col.id === 'ReviewedForTracks'}
          {#if item.ReviewedForTracks && item.ReviewedForTracks.length > 0}
            <div class="flex gap-1 flex-wrap">
              {#each item.ReviewedForTracks as rt}
                <TrackBadge text={getShortTrackName(rt)} type="reviewed" onclick={() => openTrack({ title: rt, type: 'reviewed' })} />
              {/each}
            </div>
          {/if}
        {:else if col.id === 'SubmittedForTracks'}
          {#if item.SubmittedForTracks && item.SubmittedForTracks.length > 0}
            <div class="flex gap-1 flex-wrap">
              {#each item.SubmittedForTracks as st}
                <TrackBadge text={getShortTrackName(st)} type="reviewed" onclick={() => openTrack({ title: st, type: 'reviewed' })} />
              {/each}
            </div>
          {/if}
        {:else if col.id === 'Type'}
          {#if item.Type}
            <TypeBadge text={item.Type} />
          {/if}
        {:else if col.id === 'Authors'}
          {#if item.Authors}
            <span class="cursor-help" title={item.AuthorsTooltip}>{item.Authors}</span>
          {/if}
        {:else if col.id === 'Refresh'}
          {@const isRefreshing = refreshingIds.has(item.DatabaseID)}
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              handleRowRefresh(item.DatabaseID);
            }}
            disabled={isRefreshing}
            aria-label={isRefreshing
              ? `Refreshing abstract ${item.ID}`
              : `Refresh abstract ${item.ID}`}
            title="Refresh this abstract"
            class="px-2 py-1 text-base rounded-md bg-sky-100 text-sky-800 dark:bg-sky-800 dark:text-sky-100 border-0 cursor-pointer transition-all duration-150 ease-in-out hover:bg-sky-200 dark:hover:bg-sky-700 disabled:bg-gray-200 dark:disabled:bg-gray-700 disabled:text-gray-400 dark:disabled:text-gray-400 disabled:cursor-not-allowed"
            class:animate-spin={isRefreshing}
          >
            ↻
          </button>
        {:else}
          {item[col.id]}
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}

<div class="flex flex-col overflow-auto px-1" style="height:calc(100vh - 8rem);">
  <div
    class="sticky top-0 z-10 bg-transparent
              px-2 py-2 rounded-md border-gray-200 dark:border-gray-700
              mb-2 mt-2 shrink-0 shadow-md dark:shadow-black/40"
  >
    <DataTableControls
      search={searchQuery}
      {currentPage}
      bind:perPage
      {totalItems}
      pagechange={(payload) => {
        currentPage = payload.currentPage;
      }}
      searchchange={(payload) => {
        searchQuery = payload.search;
      }}
      {columnFilters}
      {activeFilters}
      filterChange={handleFilterChange}
      filtersVisible={false}
    />
  </div>

  <section class="flex-1 overflow-auto flex flex-col max-h-screen min-h-0">
    <DataTable
      items={visibleItems}
      {visibleKeys}
      {sortKey}
      {sortDir}
      sortCallback={onSort}
      className="datatable-table w-full mt-0.5 mb-2 overflow-auto min-h-0"
      {colWidths}
      virtualize={false}
      {rowSnippet}
    />
  </section>
</div>

<!-- Abstract Detail Dialog -->
<AbstractDetailsDialog bind:open={showAbstractDialog} bind:abstract={selectedAbstract} />

<!-- Track Details Dialog -->
<TrackDetailsDialog
  bind:open={showTrackDialog}
  tracks={selectedTracks}
  allTracks={allAvailableTracks}
/>

<!-- Affiliation Details Dialog -->
<AffiliationDialog bind:open={showAffiliationDialog} affiliation={selectedAffiliation} />
