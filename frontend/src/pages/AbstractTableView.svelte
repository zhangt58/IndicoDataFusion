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
  let perPage = 25;  // Fixed value, not user-configurable
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

  $: sortedItems = (() => {
    if (!sortKey) return filteredItems;
    const copy = filteredItems.slice();
    copy.sort((a,b) => {
      const res = compare(a,b,sortKey);
      return sortDir === 'asc' ? res : -res;
    });
    return copy;
  })();

  // expose total items for DataTableControls
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

<div class="flex flex-col overflow-hidden mt-8 px-1" style="height: calc(100vh - 8rem);">
  <div class="sticky top-0 z-10 bg-transparent py-1 border-b border-gray-200 dark:border-gray-700 mb-2 flex-shrink-0">
    <DataTableControls
      search={searchQuery}
      currentPage={currentPage}
      perPage={perPage}
      {totalItems}
      pagechange={(payload) => { currentPage = payload.currentPage }}
      searchchange={(payload) => { searchQuery = payload.search }}
    />
  </div>

  <section class="flex flex-col flex-1 max-h-screen overflow-hidden">
    <VirtualDataTable
      items={visibleItems}
      {visibleKeys}
      bind:sortKey
      bind:sortDir
      className="datatable-table"
      style="width:100%;height:100%;"
      on:sort={(e) => setSort(e.detail)}
      colWidths={{ ID: '6%', Title: '30%', State: '8%', Submitter: '12%', Affiliation: '12%', Track: '8%', Type: '6%', Score: '5%', Submitted: '7%', Authors: '6%' }}
    >
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
  /* Component-specific styling for AbstractTableView */
  
  /* State badge styling - specific to AbstractTableView */
  :global(.state-badge) {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-weight: 500;
  }

  /* Authors cell with tooltip */
  :global(.authors-cell) {
    cursor: help;
  }
</style>
