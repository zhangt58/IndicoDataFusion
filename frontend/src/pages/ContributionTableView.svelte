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
  let perPage = 25;  // Fixed value, not user-configurable
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
<div class="flex flex-col overflow-hidden mt-8 px-4" style="height: calc(100vh - 8rem);">
  <!-- Sticky Controls at top -->
  <div class="sticky top-0 z-10 bg-transparent dark:bg-gray-800 py-1 border-b border-gray-200 dark:border-gray-700 mb-2 flex-shrink-0">
    <DataTableControls
      search={searchQuery}
      currentPage={currentPage}
      perPage={perPage}
      {totalItems}
      pagechange={(payload) => { currentPage = payload.currentPage }}
      searchchange={(payload) => { searchQuery = payload.search }}
    />
  </div>

  <!-- Table area: scrollable content -->
  <section class="flex-1 overflow-hidden flex flex-col min-h-0" on:click={handleTableClick}>
    <VirtualDataTable
      items={visibleItems}
      {visibleKeys}
      bind:sortKey
      bind:sortDir
      className="datatable-table"
      style="width:100%;height:100%;"
      on:sort={(e) => setSort(e.detail)}
      colWidths={{ ID: '6%', Code: '8%', Title: '36%', Type: '6%', Session: '8%', Track: '8%', Start: '8%', Duration: '6%', Room: '6%', Speakers: '8%' }}
    >
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
    display: inline-flex !important;
    align-items: flex-start !important;
    justify-content: flex-start !important;
    text-align: left !important;
    padding: 0 !important;
    background: transparent !important;
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

  /* Speakers cell with tooltip */
  :global(.speakers-cell) {
    cursor: help;
  }
</style>
