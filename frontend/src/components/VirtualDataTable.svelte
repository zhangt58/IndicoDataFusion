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
  // New prop: colWidths can be an object mapping key->css width (e.g. '30%')
  // or an array of widths matching visibleKeys order. If not provided, defaults are used.
  export let colWidths = {};

  const dispatch = createEventDispatcher();

  function handleSort(key) {
    dispatch('sort', key);
  }

  // Compute column width by key (fallback to even distribution)
  function colWidth(key, index) {
    // If caller provided colWidths, support both object and array forms
    try {
      if (colWidths) {
        if (Array.isArray(colWidths) && colWidths[index] !== undefined) {
          const v = colWidths[index];
          return typeof v === 'number' ? `${v}%` : String(v);
        }
        if (typeof colWidths === 'object' && colWidths[key] !== undefined) {
          const v = colWidths[key];
          return typeof v === 'number' ? `${v}%` : String(v);
        }
      }
    } catch (err) {
      // ignore and fall through to defaults
    }

    // Map common keys to preferred widths (matching AbstractTableView rules)
    const map = {
      'ID': '6%',
      'Title': '30%',
      'State': '8%',
      'Submitter': '12%',
      'Affiliation': '12%',
      'Track': '8%',
      'Type': '6%',
      'Score': '5%',
      'Submitted': '7%',
      'Authors': '6%'
    };
    if (map[key] !== undefined) return map[key];
    // fallback: evenly split remaining width
    const n = Math.max(1, visibleKeys.length);
    const pct = Math.floor(100 / n);
    return pct + '%';
  }
</script>

<section class="virtual-data-table {className}" style={style}>
  {#if items && items.length > 0}
    <VirtualList items={items} isTable class="datatable-table" style="width:100%;height:100%;overflow:auto;">
      {#snippet header()}
        <!-- colgroup enforces the column widths so table-layout: fixed distributes as intended -->
        <colgroup>
          {#each visibleKeys as key, i}
            <col style="width: {colWidth(key, i)}" />
          {/each}
        </colgroup>
        <thead class="sticky-header">
          <tr>
            {#each visibleKeys as key, i}
              <th style="width: {colWidth(key, i)}" class="cursor-pointer select-none" on:click={() => handleSort(key)}
                  aria-sort={sortKey === key ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}>
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

      <!-- svelte-ignore unused-function -->
      {#snippet vl_slot({ index, item })}
        <slot {item} {index} />
      {/snippet}
    </VirtualList>
  {:else}
    <div class="p-4 text-center text-slate-500">{emptyMessage}</div>
  {/if}
</section>

<style>
  .virtual-data-table { 
    display: block; 
    min-width: 0;
    height: 100%;
    overflow: auto;
  }
  
  :global(.datatable-table) {
    display: table;
    width: 100%;
    min-width: 0;
    border-collapse: collapse;
    box-sizing: border-box;
    font-size: 0.875rem; /* 14px - smaller, more compact */
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
    line-height: 1.4;
    border: none;
  }

  /* Sticky header support */
  :global(.datatable-table) thead.sticky-header {
    position: sticky;
    top: 0;
    z-index: 20;
  }

  /* Header visual styles - only bottom border */
  :global(.datatable-table) thead th {
    background-color: var(--tbl-head-bg, #f8f9fa);
    font-size: 0.8125rem; /* 13px */
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    border-top: none;
    border-left: none;
    border-right: none;
    border-bottom: 2px solid #dee2e6;
    padding: 0.5rem 0.75rem;
    text-align: left;
    white-space: normal; /* Allow header text to wrap */
    overflow: visible;
    box-shadow: 0 2px 2px -1px rgba(0, 0, 0, 0.1); /* Keep border visible on scroll */
  }

  /* Row striping - alternating colors */
  /* VirtualList may wrap rows, so we target tr directly */
  :global(.datatable-table tbody tr:nth-of-type(even)) {
    background-color: #ffffff;
  }

  :global(.datatable-table tbody tr:nth-of-type(odd)) {
    background-color: #f8f9fa;
  }

  /* Also ensure tr elements themselves have background */
  :global(.datatable-table tbody tr) {
    background-color: inherit;
  }

  /* Hover effect */
  :global(.datatable-table tbody tr:hover) {
    background-color: #e7f1ff !important;
    transition: background-color 0.15s ease-in-out;
  }

  /* Body cell defaults - only bottom border */
  :global(.datatable-table tbody td) {
    padding: 0.5rem 0.75rem;
    border-top: none;
    border-left: none;
    border-right: none;
    border-bottom: 1px solid #e9ecef;
    vertical-align: middle;
    text-align: left;
    font-size: 0.875rem; /* 14px */
    overflow: hidden;
    white-space: normal; /* Allow wrapping by default */
    word-wrap: break-word;
    box-sizing: border-box;
  }

  /* No-wrap variant for cells that should truncate */
  :global(.datatable-table tbody td.nowrap) {
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  /* Compact variant */
  :global(.datatable-table.compact tbody td) {
    padding: 0.1rem 0.3rem;
  }

  /* Header content alignment helper */
  :global(.datatable-table thead th > div) {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
  }

  /* Allow certain long cells to wrap fully if explicitly requested */
  :global(.datatable-table td.wrap) {
    white-space: normal;
    word-break: break-word;
  }

  /* Dark mode variants for the datatable */
  :global(.dark) :global(.datatable-table thead th) {
    background-color: #1f2937;
    border-bottom-color: #374151;
    color: #f9fafb;
    box-shadow: 0 2px 2px -1px rgba(0, 0, 0, 0.3); /* Darker shadow for dark mode */
  }

  :global(.dark) :global(.datatable-table tbody tr:nth-of-type(even)) {
    background-color: #111827;
  }

  :global(.dark) :global(.datatable-table tbody tr:nth-of-type(odd)) {
    background-color: #1f2937;
  }

  :global(.dark) :global(.datatable-table tbody tr:hover) {
    background-color: #374151 !important;
  }

  :global(.dark) :global(.datatable-table tbody td) {
    border-bottom-color: #374151;
    color: #e5e7eb;
  }

  /* Shared control styles (inputs, pagination, column filters) used by page components */
  :global(.datatable-input) {
    padding: 0.375rem 0.75rem;
    font-size: 0.95rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  :global(.datatable-input:focus) {
    border-color: #86b7fe;
    outline: 0;
    box-shadow: 0 0 0.1rem 0.2rem rgba(13, 110, 253, 0.25);
  }

  :global(.search-filtering-row) {
    background-color: #f8f9fa;
  }

  :global(.search-filtering-row th) {
    padding: 0.1rem 0.1rem;
    border-bottom: 1px solid #dee2e6;
  }

  :global(.column-filter) {
    width: 95%;
    padding: 0.2rem 0.25rem;
    font-size: 0.8rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  :global(.column-filter::placeholder) {
    color: #adb5bd;
    font-size: 0.75rem;
    font-style: italic;
  }

  :global(.datatable-pagination) {
    display: flex;
    gap: 0.1rem;
    margin-top: 0.6rem;
  }

  :global(.datatable-pagination li a),
  :global(.datatable-pagination li button) {
    padding: 0.3rem 0.7rem;
    border: 1px solid #dee2e6;
    border-radius: 0;
    background-color: #fff;
    color: #0d6efd;
    text-decoration: none;
  }

  :global(.datatable-pagination li a:hover),
  :global(.datatable-pagination li button:hover) {
    background-color: #e9ecef;
  }

  :global(.datatable-pagination .datatable-active a),
  :global(.datatable-pagination .datatable-active button) {
    background-color: #0d6efd;
    border-color: #0d6efd;
    color: #fff;
  }

  :global(.datatable-selector) {
    padding: 0.375rem 2rem 0.375rem 0.75rem;
    font-size: 0.875rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
    background-color: #fff;
  }

  /* Title link styling - common for both Abstract and Contribution views */
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
    font-size: 0.875rem;
    font-weight: 400;
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

  /* Badge sizing adjustments for compact table */
  :global(.datatable-table .badge) {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
  }

  /* Badges and small helpers used inside table cells */
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

  :global(.track-link) {
    cursor: pointer;
    text-decoration: none;
  }

  :global(.track-link:hover) {
    text-decoration: underline;
  }

  :global(.track-accepted) {
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  :global(.track-reviewed) {
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }

  :global(.authors-cell) {
    cursor: help;
  }
</style>
