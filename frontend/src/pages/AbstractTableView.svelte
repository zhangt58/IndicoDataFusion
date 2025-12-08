<script>
  import VirtualDataTable from '../components/VirtualDataTable.svelte';
  import TypeBadge from './TypeBadge.svelte';
  import AbstractDetailsDialog from './AbstractDetailsDialog.svelte';
  import TrackDetailsDialog from './TrackDetailsDialog.svelte';
  import { 
    getTableItems, 
    createDataTableOptions,
    rowRender
  } from './AbstractTableItem.js';
  import DataTableControls from '../components/DataTableControls.svelte';

  /** @type {Array} */
  export let abstractData = [];

  // Abstract dialog state
  let showAbstractDialog = false;
  let selectedAbstract = null;

  // Track dialog state
  let showTrackDialog = false;
  let selectedTracks = [];

  // Simple client-side controls (search/sort/pagination)
  let searchQuery = '';
  let perPage = 25;
  let currentPage = 1;
  let sortKey = null; // e.g. 'Title' or 'Score'
  let sortDir = 'asc'; // 'asc' | 'desc'

  // Map of visible column keys in the data objects
  const visibleKeys = ['ID','Title','State','Submitter','Affiliation','Track','Type','Score','Submitted','Authors'];

  // Collect all unique tracks from all abstracts
  $: allAvailableTracks = abstractData.reduce((acc, abstract) => {
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
  }, []);

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

  $: tableItems = getTableItems(abstractData);
  $: dataTableOptions = createDataTableOptions();

  // Filtering
  $: filteredItems = tableItems.filter(item => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return visibleKeys.some(k => String(item[k] ?? '').toLowerCase().includes(q));
  });

  // Sorting
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

    // fallback string compare
    const sa = String(va ?? '').toLowerCase();
    const sb = String(vb ?? '').toLowerCase();
    return sa < sb ? -1 : sa > sb ? 1 : 0;
  }

  $: sortedItems = (() => {
    if (!sortKey) return filteredItems;
    const copy = filteredItems.slice();
    copy.sort((a,b) => {
      const res = compare(a,b,sortKey);
      return sortDir === 'asc' ? res : -res;
    });
    return copy;
  })();

  // expose total items for DataTableControls (mirrors ContributionTableView)
  $: totalItems = sortedItems.length;

  // Pagination
  $: totalPages = Math.max(1, Math.ceil(sortedItems.length / perPage));
  $: currentPage = Math.min(currentPage, totalPages);
  $: paginatedItems = sortedItems.slice((currentPage-1)*perPage, currentPage*perPage);

  // For VirtualList we pass the paginatedItems so the visible window is virtualized per page
  $: visibleItems = paginatedItems;

  // Helper functions for pagination controls
  function setSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
  }

  // Helper action: apply row-render-like DOM updates to a rendered <tr>
  // Uses the original `rowRender` helper when available, falling back to manual updates.
  function applyRowRender(node, payload) {
    // payload is { item, index }
    let { item, index } = payload || {};

    function buildSyntheticRow(it) {
      // Build a synthetic `row` expected by the original rowRender(row, tr, index)
      // row.cells should be an array of objects where .data contains the column value
      const cells = [];
      // Ensure indices up to 12 exist per original mapping
      // 0 ID,1 Title,2 State,3 Submitter,4 Affiliation,5 Track,6 TrackFull,7 TrackType,8 Type,9 Score,10 Submitted,11 Authors,12 AuthorsTooltip
      cells[0] = { data: it.ID ?? '' };
      cells[1] = { data: it.Title ?? '' };
      cells[2] = { data: it.State ?? '' };
      cells[3] = { data: it.Submitter ?? '' };
      cells[4] = { data: it.Affiliation ?? '' };
      cells[5] = { data: it.Track ?? '' };
      cells[6] = { data: it.TrackFull ?? '[]' };
      cells[7] = { data: it.TrackType ?? '' };
      cells[8] = { data: it.Type ?? '' };
      cells[9] = { data: it.Score ?? '' };
      cells[10] = { data: it.Submitted ?? '' };
      cells[11] = { data: it.Authors ?? '' };
      cells[12] = { data: it.AuthorsTooltip ?? '' };
      return { cells };
    }

    // Build a plain virtual `tr` whose structure matches what rowRender expects
    function buildVirtualTr(visibleCount) {
      const vtr = { childNodes: [] };
      for (let i = 0; i < visibleCount; i++) {
        vtr.childNodes[i] = { childNodes: [ { attributes: {} } ] };
      }
      return vtr;
    }

    function apply(it) {
      // Prefer calling the original rowRender if available
      try {
        if (typeof rowRender === 'function') {
          const synthetic = buildSyntheticRow(it);
          const visibleCount = Math.max((node.children && node.children.length) || 0, (synthetic.cells || []).length || 0);
          const vtr = buildVirtualTr(visibleCount);
          // rowRender will write attributes into vtr.childNodes[x].childNodes[0].attributes
          rowRender(synthetic, vtr, index);

          // copy attributes from the virtual tr into the real DOM
          for (let i = 0; i < vtr.childNodes.length; i++) {
            const vcell = vtr.childNodes[i];
            if (!vcell || !vcell.childNodes || !vcell.childNodes[0]) continue;
            const attrs = vcell.childNodes[0].attributes || {};
            const td = node.children[i];
            if (!td) continue;
            const el = td.firstElementChild || (td.querySelector && td.querySelector('*')) || td;
            for (const name in attrs) {
              if (!Object.prototype.hasOwnProperty.call(attrs, name)) continue;
              try { el.setAttribute(name, attrs[name]); } catch (e) { /* ignore */ }
            }
          }
          return;
        }
      } catch (err) {
        console.error('rowRender call failed, falling back to manual apply:', err);
      }

      // Fallback: manual DOM updates similar to previous implementation
      try {
        // Title link
        const titleCell = node.children[1];
        const titleAnchor = titleCell && (titleCell.querySelector('.title-link') || titleCell.querySelector('a'));
        if (titleAnchor) {
          if (it.ID != null) titleAnchor.setAttribute('data-id', String(it.ID));
          if (it.Title != null) titleAnchor.setAttribute('data-title', String(it.Title));
        }

        // Track link
        const trackCell = node.children[5];
        const trackAnchor = trackCell && (trackCell.querySelector('.track-link') || trackCell.querySelector('a'));
        if (trackAnchor) {
          if (it.TrackFull != null) trackAnchor.setAttribute('data-tracks', String(it.TrackFull));
          trackAnchor.classList.remove('track-accepted', 'track-reviewed');
          if (it.TrackType === 'accepted') trackAnchor.classList.add('track-accepted');
          else trackAnchor.classList.add('track-reviewed');
        }

        // Authors tooltip
        const authorsCell = node.children[9];
        const authorsSpan = authorsCell && (authorsCell.querySelector('.authors-cell') || authorsCell.querySelector('span'));
        if (authorsSpan && it.AuthorsTooltip != null) {
          authorsSpan.setAttribute('title', String(it.AuthorsTooltip));
        }
      } catch (err) {
        console.error('applyRowRender fallback error', err);
      }
    }

    // initial apply
    apply(item);

    return {
      update(newPayload) {
        ({ item, index } = newPayload || {});
        apply(item);
      },
      destroy() {
        // nothing to cleanup
      }
    };
  }
</script>

<!-- Controls + table wrapped in a viewport-height flex container -->
<div class="space-y-4 mt-8" style="height:100vh; display:flex; flex-direction:column;">
  <!-- Controls: search, perPage, (removed Sort-by select) -->
  <div class="flex items-center gap-4 p-2">
    <!-- DataTableControls uses Svelte v5 callback-prop API: pass values and callbacks
         instead of Svelte v3-style events/bind. -->
    <DataTableControls
      search={searchQuery}
      currentPage={currentPage}
      perPage={perPage}
      {totalItems}
      perPageOptions={[10,25,50,100]}
      perpagechange={(payload) => { perPage = payload.perPage }}
      pagechange={(payload) => { currentPage = payload.currentPage }}
      searchchange={(payload) => { searchQuery = payload.search }}
    />
  </div>

   <!-- Table area: grow to fill remaining viewport space -->
   <section class="mt-2 p-4 abstract-table-view" style="flex:1;display:flex;flex-direction:column;overflow:hidden;">
     <VirtualDataTable items={visibleItems} {visibleKeys} bind:sortKey bind:sortDir className="datatable-table" style="width:100%;height:100%" on:sort={(e) => setSort(e.detail)}>
       <svelte:fragment slot="default" let:item let:index>
         <tr use:applyRowRender={{ item, index }}>
           <td>{item.ID}</td>
           <td>
             <button type="button" class="title-link" on:click={() => openAbstract(item.ID)} data-id={item.ID} data-title={item.Title}>{item.Title}</button>
           </td>
           <td>
             {#if item.State}
               <span class={item.State.toLowerCase() === 'accepted' ? 'state-badge state-accepted' : (item.State.toLowerCase() === 'rejected' ? 'state-badge state-rejected' : 'state-badge state-other')}>{item.State}</span>
             {/if}
           </td>
           <td>{item.Submitter}</td>
           <td>{item.Affiliation}</td>
           <td>
             {#if item.Track}
               <button type="button" class={'track-badge ' + (item.TrackType === 'accepted' ? 'track-accepted' : 'track-reviewed') + ' track-link'} on:click={() => openTrack(item.TrackFull)} data-tracks={item.TrackFull}>{item.Track}</button>
             {/if}
           </td>
           <td>
             {#if item.Type}
               <TypeBadge text={item.Type} />
             {/if}
           </td>
           <td>{item.Score}</td>
           <td>{item.Submitted}</td>
           <td>
             {#if item.Authors}
               <span class="authors-cell" title={item.AuthorsTooltip}>{item.Authors}</span>
             {/if}
           </td>
         </tr>
       </svelte:fragment>
     </VirtualDataTable>
    </section>
  </div>

 <!-- Abstract Detail Dialog -->
 <AbstractDetailsDialog bind:open={showAbstractDialog} abstract={selectedAbstract} />

 <!-- Track Details Dialog -->
 <TrackDetailsDialog bind:open={showTrackDialog} tracks={selectedTracks} allTracks={allAvailableTracks} />

<style>
  /* Title link styling */
  :global(.title-link) {
    color: #0d6efd;
    text-decoration: none;
    cursor: pointer;
    /* Make buttons behave like left-aligned links and align content to the top-left */
    display: inline-flex !important;
    align-items: flex-start !important; /* vertical alignment inside the button */
    justify-content: flex-start !important; /* horizontal alignment */
    text-align: left !important; /* ensure multi-line text is left-aligned */
    padding: 0.0rem !important; /* remove extra button padding that can change alignment */
    background: transparent !important; /* look like a link */
    border: none !important;
  }

  :global(.title-link:hover) {
    color: #0a58ca;
    text-decoration: underline;
  }

  :global(.dark .title-link) {
    color: #60a5fa;
  }

  :global(.dark .title-link:hover) {
    color: #93c5fd;
  }

  /* State badge styling - matches card view */
  :global(.state-badge) {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: capitalize;
  }

  :global(.state-accepted) {
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  :global(.state-rejected) {
    background-color: #fee2e2 !important;
    color: #991b1b !important;
  }

  :global(.state-other) {
    background-color: #fef3c7 !important;
    color: #92400e !important;
  }

  /* Track badge styling - matches card view */

  :global(.track-link) {
    cursor: pointer;
    text-decoration: none;
  }

  :global(.track-link:hover) {
    text-decoration: underline;
  }

  /* Authors cell with tooltip */
  :global(.authors-cell) {
    cursor: help;
  }

  :global(.track-accepted) {
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  :global(.track-reviewed) {
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }

  .abstract-table-view :global(.datatable-table) {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.95rem;
  }

  .abstract-table-view :global(.datatable-table thead th) {
    background-color: #f8f9fa;
    border-bottom: 2px solid #dee2e6;
    padding: 0.3rem 0.5rem;
    text-align: left;
    font-weight: 600;
    white-space: nowrap;
  }

  /* add sticky header styles */
  .abstract-table-view :global(.datatable-table thead) {
    /* ensure thead is treated as header group inside the scroll container */
    display: table-header-group;
  }

  .abstract-table-view :global(.datatable-table thead th) {
    position: sticky;
    top: 0;
    z-index: 20;
    /* keep header background to cover rows when sticky */
    background-color: var(--tbl-head-bg, #f8f9fa);
  }

  .abstract-table-view :global(.datatable-table tbody tr:nth-child(odd)) {
    background-color: rgba(0, 0, 0, 0.02);
  }

  .abstract-table-view :global(.datatable-table tbody tr:hover) {
    background-color: rgba(0, 0, 0, 0.075);
  }

  .abstract-table-view :global(.datatable-table tbody td) {
    padding: 0.2rem 0.5rem;
    border-top: 1px solid #dee2e6;
    vertical-align: middle;
    text-align: left;
  }

  /* Compact styling */
  .abstract-table-view :global(.datatable-table.compact tbody td) {
    padding: 0.1rem 0.3rem;
  }

  /* Search input styling */
  .abstract-table-view :global(.datatable-input) {
    padding: 0.375rem 0.75rem;
    font-size: 0.95rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  .abstract-table-view :global(.datatable-input:focus) {
    border-color: #86b7fe;
    outline: 0;
    box-shadow: 0 0 0.1rem 0.2rem rgba(13, 110, 253, 0.25);
  }

  /* Column filter row styling */
  .abstract-table-view :global(.search-filtering-row) {
    background-color: #f8f9fa;
  }

  .abstract-table-view :global(.search-filtering-row th) {
    padding: 0.1rem 0.1rem;
    border-bottom: 1px solid #dee2e6;
  }

  .abstract-table-view :global(.column-filter) {
    width: 95%;
    padding: 0.2rem 0.25rem;
    font-size: 0.8rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  .abstract-table-view :global(.column-filter::placeholder) {
    color: #adb5bd;
    font-size: 0.75rem;
    font-style: italic;
  }

  /* Pagination styling */
  .abstract-table-view :global(.datatable-pagination) {
    display: flex;
    gap: 0.1rem;
    margin-top: 0.6rem;
  }

  .abstract-table-view :global(.datatable-pagination li a),
  .abstract-table-view :global(.datatable-pagination li button) {
    padding: 0.3rem 0.7rem;
    border: 1px solid #dee2e6;
    border-radius: 0;
    background-color: #fff;
    color: #0d6efd;
    text-decoration: none;
  }

  .abstract-table-view :global(.datatable-pagination li a:hover),
  .abstract-table-view :global(.datatable-pagination li button:hover) {
    background-color: #e9ecef;
  }

  .abstract-table-view :global(.datatable-pagination .datatable-active a),
  .abstract-table-view :global(.datatable-pagination .datatable-active button) {
    background-color: #0d6efd;
    border-color: #0d6efd;
    color: #fff;
  }

  /* Per page select styling */
  .abstract-table-view :global(.datatable-selector) {
    padding: 0.375rem 2rem 0.375rem 0.75rem;
    font-size: 0.875rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
    background-color: #fff;
  }

  /* Dark mode support */
  :global(.dark .state-accepted) {
    background-color: #14532d !important;
    color: #bbf7d0 !important;
  }

  :global(.dark .state-rejected) {
    background-color: #7f1d1d !important;
    color: #fecaca !important;
  }

  :global(.dark .state-other) {
    background-color: #78350f !important;
    color: #fde68a !important;
  }

  :global(.dark .track-accepted) {
    background-color: #14532d !important;
    color: #bbf7d0 !important;
  }

  :global(.dark .track-reviewed) {
    background-color: #581c87 !important;
    color: #e9d5ff !important;
  }

  :global(.dark) .abstract-table-view :global(.datatable-table thead th) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .abstract-table-view :global(.datatable-table tbody tr:nth-child(odd)) {
    background-color: rgba(255, 255, 255, 0.02);
  }

  :global(.dark) .abstract-table-view :global(.datatable-table tbody tr:hover) {
    background-color: rgba(255, 255, 255, 0.075);
  }

  :global(.dark) .abstract-table-view :global(.datatable-table tbody td) {
    border-color: #4b5563;
    color: #e5e7eb;
  }

  :global(.dark) .abstract-table-view :global(.datatable-input) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .abstract-table-view :global(.search-filtering-row) {
    background-color: #1f2937;
  }

  :global(.dark) .abstract-table-view :global(.search-filtering-row th) {
    border-color: #4b5563;
  }

  :global(.dark) .abstract-table-view :global(.column-filter) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .abstract-table-view :global(.column-filter::placeholder) {
    color: #9ca3af;
  }

  :global(.dark) .abstract-table-view :global(.datatable-selector) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .abstract-table-view :global(.datatable-pagination li a),
  :global(.dark) .abstract-table-view :global(.datatable-pagination li button) {
    background-color: #374151;
    border-color: #4b5563;
    color: #60a5fa;
  }

  :global(.dark) .abstract-table-view :global(.datatable-pagination .datatable-active a),
  :global(.dark) .abstract-table-view :global(.datatable-pagination .datatable-active button) {
    background-color: #2563eb;
    border-color: #2563eb;
    color: #fff;
  }
</style>
