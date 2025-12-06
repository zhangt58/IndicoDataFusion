<script>
  import VirtualDataTable from '../components/VirtualDataTable.svelte';
  import ContributionDetailsDialog from './ContributionDetailsDialog.svelte';
  import { 
    getTableItems, 
    createDataTableOptions,
    rowRender
  } from './ContributionTableItem.js';
  import SessionBadge from './SessionBadge.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import TypeBadge from './TypeBadge.svelte';
  import DataTableControls from '../components/DataTableControls.svelte';

  /** @type {Array} */
  export let contributionData = [];

  // Contribution dialog state (can be extended later for contribution details dialog)
  let showContributionDialog = false;
  let selectedContribution = null;

  // Find contribution by ID
  function findContributionById(id) {
    return contributionData.find(c => String(c.friendly_id || c.id) === String(id));
  }

  // Open contribution details by id (used by title button)
  function openContribution(id) {
    const sid = String(id);
    selectedContribution = findContributionById(sid);
    if (selectedContribution) showContributionDialog = true;
  }

  // Handle clicks on the table (keeps existing delegation behavior)
  function handleTableClick(event) {
    const target = event.target;
    
    // Handle title link click
    if (target.classList.contains('title-link')) {
      event.preventDefault();
      const contributionId = target.dataset.id;
      if (!contributionId) {
        console.warn('No contribution ID found in data-id attribute');
        return;
      }
      selectedContribution = findContributionById(contributionId);
      if (selectedContribution) {
        showContributionDialog = true;
      }
    }
  }

  // --- Virtualized table client-side controls (search/sort/pagination) ---
  let searchQuery = '';
  let perPage = 25;
  let currentPage = 1;
  let sortKey = null;
  let sortDir = 'asc';

  // Visible columns for contributions (matches ContributionTableItem.js visibleColumnNames)
  const visibleKeys = ['ID','Code','Title','Type','Session','Track','Start','Duration','Room','Speakers'];

  $: tableItems = getTableItems(contributionData);
  $: dataTableOptions = createDataTableOptions(); // kept for potential external use

  // Filtering
  $: filteredItems = tableItems.filter(item => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return visibleKeys.some(k => String(item[k] ?? '').toLowerCase().includes(q));
  });

  // Sorting (string compare for all columns)
  function compare(a,b,key) {
    const sa = String(a[key] ?? '').toLowerCase();
    const sb = String(b[key] ?? '').toLowerCase();
    if (sa < sb) return -1;
    if (sa > sb) return 1;
    return 0;
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

  // total items available after filtering/sorting
  $: totalItems = sortedItems.length;

  // Pagination
  $: totalPages = Math.max(1, Math.ceil(sortedItems.length / perPage));
  $: currentPage = Math.min(currentPage, totalPages);
  $: paginatedItems = sortedItems.slice((currentPage-1)*perPage, currentPage*perPage);
  $: visibleItems = paginatedItems;

  function setSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
  }

  // Helper action to apply the original rowRender to a rendered <tr>
  function applyRowRender(node, payload) {
    let { item, index } = payload || {};

    function buildSyntheticRow(it) {
      const cells = [];
      // Build cells matching ContributionTableItem.js expected data indices 0..14
      cells[0] = { data: it.ID ?? '' };
      cells[1] = { data: it.Code ?? '' };
      cells[2] = { data: it.Title ?? '' };
      cells[3] = { data: it.Type ?? '' };
      cells[4] = { data: it.Session ?? '' };
      cells[5] = { data: it.Track ?? '' };
      cells[6] = { data: it.StartDate ?? '' };
      cells[7] = { data: it.Duration ?? '' };
      cells[8] = { data: it.Location ?? '' };
      cells[9] = { data: it.Room ?? '' };
      cells[10] = { data: it.Speakers ?? '' };
      cells[11] = { data: it.SpeakersTooltip ?? '' };
      cells[12] = { data: it.Authors ?? '' };
      cells[13] = { data: it.AuthorsTooltip ?? '' };
      cells[14] = { data: it.URL ?? '' };
      return { cells };
    }

    // Build a plain virtual `tr` whose structure matches what rowRender expects
    function buildVirtualTr(visibleCount) {
      const vtr = { childNodes: [] };
      for (let i = 0; i < visibleCount; i++) {
        // Each cell has childNodes; the first child has an attributes object that rowRender will mutate
        vtr.childNodes[i] = { childNodes: [ { attributes: {} } ] };
      }
      return vtr;
    }

    function apply(it) {
      // First try to call the original rowRender but using a virtual tr object
      try {
        if (typeof rowRender === 'function') {
          const synthetic = buildSyntheticRow(it);
          // create a virtual tr with slots matching the visible TDs in the real DOM
          const visibleCount = Math.max((node.children && node.children.length) || 0, (synthetic.cells || []).length || 0);
          const vtr = buildVirtualTr(visibleCount);
          // Call rowRender with the synthetic row and virtual tr. rowRender will write into vtr.childNodes[x].childNodes[0].attributes
          rowRender(synthetic, vtr, index);

          // Copy any attributes set by rowRender from vtr into the real DOM elements
          for (let i = 0; i < vtr.childNodes.length; i++) {
            const vcell = vtr.childNodes[i];
            if (!vcell || !vcell.childNodes || !vcell.childNodes[0]) continue;
            const attrs = vcell.childNodes[0].attributes || {};
            const td = node.children && node.children[i];
            if (!td) continue;
            // Prefer the first element child inside the TD (e.g., button, a, span)
            const el = td.firstElementChild || td.querySelector && td.querySelector('*') || td;
            for (const name in attrs) {
              if (!Object.prototype.hasOwnProperty.call(attrs, name)) continue;
              try {
                el.setAttribute(name, attrs[name]);
              } catch (e) {
                // ignore if setAttribute fails for any reason
              }
            }
          }
          return;
        }
      } catch (err) {
        console.error('rowRender call failed, falling back to manual apply:', err);
      }

      // fallback manual attribute updates
      try {
        const titleCell = node.children[2];
        const titleAnchor = titleCell && (titleCell.querySelector('.title-link') || titleCell.querySelector('a'));
        if (titleAnchor) {
          if (it.ID != null) titleAnchor.setAttribute('data-id', String(it.ID));
          if (it.Title != null) titleAnchor.setAttribute('data-title', String(it.Title));
        }

        const speakersCell = node.children[9];
        const speakersSpan = speakersCell && (speakersCell.querySelector('.speakers-cell') || speakersCell.querySelector('span'));
        if (speakersSpan && it.SpeakersTooltip != null) {
          speakersSpan.setAttribute('title', String(it.SpeakersTooltip));
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
      destroy() {}
    };
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="mt-12 p-4 contribution-table-view">
  <!-- Controls -->
  <div class="flex items-center gap-4 p-2">
    <DataTableControls
      bind:perPage
      bind:currentPage
      bind:search={searchQuery}
      {totalItems}
      perPageOptions={[10,25,50,100]}
      on:perpagechange={(e) => { perPage = e.detail.perPage }}
      on:pagechange={(e) => { currentPage = e.detail.currentPage }}
      on:searchchange={(e) => { searchQuery = e.detail.search }}
    />
  </div>

  <!-- Virtualized table -->
  <section on:click={handleTableClick} style="flex:1;overflow:auto;">
    <VirtualDataTable items={visibleItems} {visibleKeys} bind:sortKey bind:sortDir className="datatable-table" style="width:100%;height:100%" on:sort={(e) => setSort(e.detail)}>
      <svelte:fragment slot="default" let:item let:index>
        <tr use:applyRowRender={{ item, index }}>
          <td>{item.ID}</td>
          <td>{item.Code}</td>
          <td><button type="button" class="title-link" data-id={item.ID} on:click={() => openContribution(item.ID)}>{item.Title}</button></td>
          <td>
            {#if item.Type}
              <TypeBadge text={item.Type} />
            {/if}
          </td>
          <td>
            {#if item.Session}
              <SessionBadge text={item.Session} />
            {/if}
          </td>
          <td>
            {#if item.Track}
              <TrackBadge text={item.Track} className="track-link" {...{ 'data-tracks': item.Track }} />
            {/if}
          </td>
          <td>{item.StartDate}</td>
          <td>{item.Duration}</td>
          <td>{item.Room}</td>
          <td>{#if item.Speakers}<span class="speakers-cell" title={item.SpeakersTooltip}>{item.Speakers}</span>{/if}</td>
        </tr>
      </svelte:fragment>
    </VirtualDataTable>
   </section>
 </div>

<!-- Contribution Detail Dialog -->
<ContributionDetailsDialog bind:open={showContributionDialog} contribution={selectedContribution} />

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

  /* Speakers/Authors cell with tooltip */
  :global(.speakers-cell),
  :global(.authors-cell) {
    cursor: help;
  }

  /* Bootstrap-like table styling */
  .contribution-table-view :global(.datatable-wrapper) {
    font-size: 0.9rem;
  }

  .contribution-table-view :global(.datatable-table) {
    width: 100%;
    border-collapse: collapse;
  }

  .contribution-table-view :global(.datatable-table thead th) {
    background-color: #f8f9fa;
    border-bottom: 2px solid #dee2e6;
    padding: 0.3rem 0.5rem;
    text-align: left;
    font-weight: 600;
    white-space: nowrap;
  }

  .contribution-table-view :global(.datatable-table tbody tr:nth-child(odd)) {
    background-color: rgba(0, 0, 0, 0.02);
  }

  .contribution-table-view :global(.datatable-table tbody tr:hover) {
    background-color: rgba(0, 0, 0, 0.075);
  }

  .contribution-table-view :global(.datatable-table tbody td) {
    padding: 0.2rem 0.5rem;
    border-top: 1px solid #dee2e6;
    vertical-align: middle;
  }

  /* Compact styling */
  .contribution-table-view :global(.datatable-table.compact tbody td) {
    padding: 0.1rem 0.3rem;
  }

  /* Search input styling */
  .contribution-table-view :global(.datatable-input) {
    padding: 0.375rem 0.75rem;
    font-size: 0.95rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  .contribution-table-view :global(.datatable-input:focus) {
    border-color: #86b7fe;
    outline: 0;
    box-shadow: 0 0 0.1rem 0.2rem rgba(13, 110, 253, 0.25);
  }

  /* Column filter row styling */
  .contribution-table-view :global(.search-filtering-row) {
    background-color: #f8f9fa;
  }

  .contribution-table-view :global(.search-filtering-row th) {
    padding: 0.1rem 0.1rem;
    border-bottom: 1px solid #dee2e6;
  }

  .contribution-table-view :global(.column-filter) {
    width: 95%;
    padding: 0.2rem 0.25rem;
    font-size: 0.8rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  .contribution-table-view :global(.column-filter::placeholder) {
    color: #adb5bd;
    font-size: 0.75rem;
    font-style: italic;
  }

  /* Pagination styling */
  .contribution-table-view :global(.datatable-pagination) {
    display: flex;
    gap: 0.1rem;
    margin-top: 0.6rem;
  }

  .contribution-table-view :global(.datatable-pagination li a),
  .contribution-table-view :global(.datatable-pagination li button) {
    padding: 0.3rem 0.7rem;
    border: 1px solid #dee2e6;
    border-radius: 0;
    background-color: #fff;
    color: #0d6efd;
    text-decoration: none;
  }

  .contribution-table-view :global(.datatable-pagination li a:hover),
  .contribution-table-view :global(.datatable-pagination li button:hover) {
    background-color: #e9ecef;
  }

  .contribution-table-view :global(.datatable-pagination .datatable-active a),
  .contribution-table-view :global(.datatable-pagination .datatable-active button) {
    background-color: #0d6efd;
    border-color: #0d6efd;
    color: #fff;
  }

  /* Per page select styling */
  .contribution-table-view :global(.datatable-selector) {
    padding: 0.375rem 2rem 0.375rem 0.75rem;
    font-size: 0.875rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
    background-color: #fff;
  }

  :global(.dark) .contribution-table-view :global(.datatable-table thead th) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .contribution-table-view :global(.datatable-table tbody tr:nth-child(odd)) {
    background-color: rgba(255, 255, 255, 0.02);
  }

  :global(.dark) .contribution-table-view :global(.datatable-table tbody tr:hover) {
    background-color: rgba(255, 255, 255, 0.075);
  }

  :global(.dark) .contribution-table-view :global(.datatable-table tbody td) {
    border-color: #4b5563;
    color: #e5e7eb;
  }

  :global(.dark) .contribution-table-view :global(.datatable-input) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .contribution-table-view :global(.search-filtering-row) {
    background-color: #1f2937;
  }

  :global(.dark) .contribution-table-view :global(.search-filtering-row th) {
    border-color: #4b5563;
  }

  :global(.dark) .contribution-table-view :global(.column-filter) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .contribution-table-view :global(.column-filter::placeholder) {
    color: #9ca3af;
  }

  :global(.dark) .contribution-table-view :global(.datatable-selector) {
    background-color: #374151;
    border-color: #4b5563;
    color: #f9fafb;
  }

  :global(.dark) .contribution-table-view :global(.datatable-pagination li a),
  :global(.dark) .contribution-table-view :global(.datatable-pagination li button) {
    background-color: #374151;
    border-color: #4b5563;
    color: #60a5fa;
  }

  :global(.dark) .contribution-table-view :global(.datatable-pagination .datatable-active a),
  :global(.dark) .contribution-table-view :global(.datatable-pagination .datatable-active button) {
    background-color: #2563eb;
    border-color: #2563eb;
    color: #fff;
  }
</style>
