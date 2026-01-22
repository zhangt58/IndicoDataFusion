<script>
  import { DataTable, DataTableControls } from '@zhangt58/svelte-vtable';
  import ContributionDetailsDialog from './ContributionDetailsDialog.svelte';
  import { getTableItems } from './ContributionTableItem.js';
  import SessionBadge from './SessionBadge.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import TypeBadge from './TypeBadge.svelte';
  import TitleLink from '../components/TitleButton.svelte';
  import TrackDetailsDialog from './TrackDetailsDialog.svelte';
  import SessionDetailsDialog from './SessionDetailsDialog.svelte';

  let { contributionData = [] } = $props();

  // Contribution dialog state (can be extended later for contribution details dialog)
  let showContributionDialog = $state(false);
  let selectedContribution = $state(null);

  // Track dialog state
  let showTrackDialog = $state(false);
  let selectedTracks = $state([]);

  // Session dialog state
  let showSessionDialog = $state(false);
  let selectedSessions = $state([]);

  // Column filters state
  let activeFilters = $state({});

  // Aggregate all unique sessions from contributions
  let allAvailableSessions = $derived(
    contributionData.reduce((acc, c) => {
      const title = c.session || c.Session || null;
      if (title && !acc.some((s) => s.title === title)) {
        acc.push({ title });
      }
      return acc;
    }, []),
  );

  // Open session dialog - accepts a string or array/object
  function openSession(sessionFull) {
    try {
      if (!sessionFull) {
        selectedSessions = [];
        return;
      }
      // If it's a simple title string, gather matching table items to show additional fields
      if (typeof sessionFull === 'string') {
        // try parsing JSON first
        let title = sessionFull;
        try {
          const parsed = JSON.parse(sessionFull);
          if (parsed && typeof parsed === 'object' && parsed.title) title = parsed.title;
        } catch (e) {
          // not JSON, keep the raw title
        }
        const matches = tableItems.filter((it) => it.Session === title);
        selectedSessions = [{ title, items: matches }];
      } else if (Array.isArray(sessionFull)) {
        // try to normalize array entries
        selectedSessions = sessionFull.map((s) =>
          typeof s === 'string'
            ? { title: s, items: tableItems.filter((it) => it.Session === s) }
            : s,
        );
      } else if (sessionFull && typeof sessionFull === 'object') {
        // object may contain a title
        const title = sessionFull.title || sessionFull.name || '';
        const matches = title ? tableItems.filter((it) => it.Session === title) : [];
        selectedSessions = [{ ...sessionFull, items: matches }];
      } else {
        selectedSessions = [];
      }

      if (selectedSessions.length > 0) showSessionDialog = true;
    } catch (err) {
      console.error('Failed to open session dialog', err);
      selectedSessions = [];
    }
  }

  // Find contribution by ID
  function findContributionById(id) {
    return contributionData.find((c) => String(c.friendly_id || c.id) === String(id));
  }

  // Open contribution details by id (used by title button)
  function openContribution(id) {
    const sid = String(id);
    selectedContribution = findContributionById(sid);
    if (selectedContribution) showContributionDialog = true;
  }

  // Aggregate all unique tracks from contributions
  let allAvailableTracks = $derived(
    contributionData.reduce((acc, c) => {
      const title = c.track || c.Track || null;
      if (title && !acc.some((t) => t.title === title)) {
        // contribution source doesn't carry accepted/reviewed flag, mark as unknown
        acc.push({ title, type: 'unknown' });
      }
      return acc;
    }, []),
  );

  // Open track dialog - accepts JSON string/array/object or plain title string
  function openTrack(trackFull) {
    try {
      if (!trackFull) {
        selectedTracks = [];
        return;
      }
      // If it's already an array/object, handle accordingly
      if (typeof trackFull === 'string') {
        // Try JSON parse first
        try {
          const parsed = JSON.parse(trackFull);
          selectedTracks = Array.isArray(parsed) ? parsed : [parsed];
        } catch (e) {
          // plain title string
          selectedTracks = [{ title: trackFull, type: 'unknown' }];
        }
      } else if (Array.isArray(trackFull)) {
        selectedTracks = trackFull;
      } else {
        selectedTracks = [trackFull];
      }
      if (selectedTracks.length > 0) showTrackDialog = true;
    } catch (err) {
      console.error('Failed to open track dialog:', err);
      selectedTracks = [];
    }
  }

  // --- Virtualized table client-side controls (search/sort/pagination) ---
  let searchQuery = $state('');
  let perPage = $state(25);
  let currentPage = $state(1);
  let sortKey = $state(null);
  let sortDir = $state('asc');

  // Visible columns for contributions (matches ContributionTableItem.js visibleColumnNames)
  const columns = [
    { id: 'ID', title: 'ID', stretch: 1 },
    { id: 'Code', title: 'Code', stretch: 2 },
    { id: 'Title', title: 'Title', stretch: 6 },
    { id: 'Type', title: 'Type', stretch: 1 },
    { id: 'Session', title: 'Session', stretch: 2 },
    { id: 'Track', title: 'Track', stretch: 2 },
    { id: 'StartDate', title: 'Start', stretch: 2 },
    { id: 'Duration', title: 'Duration', stretch: 1 },
    { id: 'Room', title: 'Room', stretch: 1 },
    { id: 'Speakers', title: 'Speakers', stretch: 2 },
    { id: 'Affiliations', title: 'Affiliations', stretch: 3 },
  ];

  const visibleKeys = columns.map((c) => c.title);

  const mappedColumns = columns.map((c) => ({
    id: c.id,
    title: c.title,
    nowrap: false,
    stretch: c.stretch,
  }));
  const colWidths = mappedColumns.reduce((acc, c) => {
    acc[c.title] = c.stretch;
    return acc;
  }, {});

  let tableItems = $derived(getTableItems(contributionData));

  // columnFilters derived from tableItems
  function getUniqueValuesWithCounts(items, header) {
    const counts = {};
    (items || []).forEach((it) => {
      const val = it && it[header];
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

  let columnFilters = $derived(
    columns.map((c) => {
      const { uniqueValues, counts } = getUniqueValuesWithCounts(tableItems || [], c.title);
      return { key: c.title, label: c.title, uniqueValues, counts };
    }),
  );

  // Filtering
  let filteredItems = $derived(
    tableItems.filter((item) => {
      // apply active column filters first
      if (Object.keys(activeFilters).length > 0) {
        for (const [columnKey, selectedValues] of Object.entries(activeFilters)) {
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
      }

      if (!searchQuery) return true;
      const q = searchQuery.toLowerCase();
      return visibleKeys.some((k) =>
        String(item[k] ?? '')
          .toLowerCase()
          .includes(q),
      );
    }),
  );

  // Sorting (string compare for all columns)
  function compare(a, b, key) {
    // Special-case ID: numeric sort if IDNumber exists
    if (key === 'ID') {
      const na = a.IDNumber != null ? Number(a.IDNumber) : NaN;
      const nb = b.IDNumber != null ? Number(b.IDNumber) : NaN;
      if (isNaN(na) && isNaN(nb)) return 0;
      if (isNaN(na)) return -1;
      if (isNaN(nb)) return 1;
      return na - nb;
    }

    // Special-case Duration: sort by DurationMinutes (number of minutes)
    if (key === 'Duration') {
      const da = a.DurationMinutes != null ? Number(a.DurationMinutes) : NaN;
      const db = b.DurationMinutes != null ? Number(b.DurationMinutes) : NaN;
      if (isNaN(da) && isNaN(db)) return 0;
      if (isNaN(da)) return -1;
      if (isNaN(db)) return 1;
      return da - db;
    }

    // Special-case Start: sort by StartMillis (timestamp)
    if (key === 'Start') {
      const sa = a.StartMillis != null ? Number(a.StartMillis) : NaN;
      const sb = b.StartMillis != null ? Number(b.StartMillis) : NaN;
      if (isNaN(sa) && isNaN(sb)) return 0;
      if (isNaN(sa)) return -1;
      if (isNaN(sb)) return 1;
      return sa - sb;
    }

    // Special-case Track: extract number from strings like "MC10" or "MC10: description"
    if (key === 'Track') {
      const extractTrackNumber = (str) => {
        if (!str) return NaN;
        const match = String(str).match(/\d+/);
        return match ? Number(match[0]) : NaN;
      };
      const na = extractTrackNumber(a[key]);
      const nb = extractTrackNumber(b[key]);
      if (isNaN(na) && isNaN(nb)) {
        // fallback to string comparison if no numbers found
        const sa = String(a[key] ?? '').toLowerCase();
        const sb = String(b[key] ?? '').toLowerCase();
        return sa < sb ? -1 : sa > sb ? 1 : 0;
      }
      if (isNaN(na)) return -1;
      if (isNaN(nb)) return 1;
      return na - nb;
    }

    // fallback: string compare
    const sa = String(a[key] ?? '').toLowerCase();
    const sb = String(b[key] ?? '').toLowerCase();
    if (sa < sb) return -1;
    if (sa > sb) return 1;
    return 0;
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

  // total items available after filtering/sorting
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
  let visibleItems = $derived(paginatedItems);

  function setSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="flex flex-col overflow-auto" style="height:calc(100vh - 8rem);">
  <!-- Sticky Controls at top -->
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
      filterChange={({ allFilters }) => {
        activeFilters = { ...allFilters };
        currentPage = 1;
      }}
      filtersVisible={false}
    />
  </div>

  <section class="flex-1 overflow-auto flex flex-col max-h-screen min-h-0">
    <DataTable
      items={visibleItems}
      {visibleKeys}
      {sortKey}
      {sortDir}
      sortCallback={(k) => setSort(k)}
      className="datatable-table w-full mt-0.5 mb-2 overflow-auto min-h-0"
      {colWidths}
      virtualize={false}
      {rowSnippet}
    />
  </section>
</div>

<!-- Contribution Detail Dialog -->
<ContributionDetailsDialog bind:open={showContributionDialog} contribution={selectedContribution} />

<!-- Track Details Dialog -->
<TrackDetailsDialog
  bind:open={showTrackDialog}
  tracks={selectedTracks}
  allTracks={allAvailableTracks}
  showTypes={false}
/>

<!-- Session Details Dialog -->
<SessionDetailsDialog
  bind:open={showSessionDialog}
  sessions={selectedSessions}
  allSessions={allAvailableSessions}
/>

<!-- Row snippet for ContributionTableView (moved out of <script>) -->
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
        {:else if col.id === 'Code'}
          {item.Code}
        {:else if col.id === 'Title'}
          <TitleLink as="button" data-id={item.ID} onclick={() => openContribution(item.ID)}
            >{item.Title}</TitleLink
          >
        {:else if col.id === 'Type'}
          {#if item.Type}
            <TypeBadge text={item.Type} />
          {/if}
        {:else if col.id === 'Session'}
          {#if item.Session}
            <SessionBadge
              text={item.Session}
              onclick={() => openSession(item.Session)}
              data-session={item.Session}
            />
          {/if}
        {:else if col.id === 'Track'}
          {#if item.Track}
            <TrackBadge
              text={item.Track}
              className="track-link"
              onclick={() => openTrack(item.Track)}
              data-tracks={item.Track}
            />
          {/if}
        {:else if col.id === 'Speakers'}
          {#if item.Speakers}
            <span class="cursor-help" title={item.SpeakersTooltip}>{item.Speakers}</span>
          {/if}
        {:else if col.id === 'Affiliations'}
          {#if item.SpeakersAffiliations}
            <span title={item.SpeakersTooltip}>
              {item.SpeakersAffiliations}
            </span>
          {/if}
        {:else}
          {item[col.id]}
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}
