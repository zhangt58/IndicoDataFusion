<script>
  import { Table } from '@flowbite-svelte-plugins/datatable';

  /** @type {Array} */
  export let abstractData = [];

  // Helper to format timestamp (remove timezone and microseconds)
  function formatTimestamp(dt) {
    if (!dt) return '';
    // Remove microseconds (.123456) and timezone suffix like +00:00 or Z
    return dt.replace(/\.\d{6}/, '').replace(/([+-]\d{2}:\d{2}|Z)$/, '').replace('T', ' ');
  }

  // Transform abstract data for the table view
  function getTableItems(data) {
    return data.map(abstract => ({
      ID: abstract.friendly_id || abstract.id,
      Title: abstract.title || '',
      State: abstract.state || '',
      Submitter: abstract.submitter?.full_name || '',
      Affiliation: abstract.submitter?.affiliation || '',
      Track: abstract.accepted_track?.title || abstract.reviewed_for_tracks?.[0]?.title || '',
      TrackType: abstract.accepted_track ? 'accepted' : 'reviewed',
      Type: abstract.accepted_contrib_type?.name || '',
      Score: abstract.score ?? '',
      Submitted: formatTimestamp(abstract.submitted_dt),
      Authors: abstract.persons?.map(p => `${p.first_name} ${p.last_name}`).join(', ') || ''
    }));
  }

  // Custom column rendering for State with badge styling
  const renderState = function(data, cell, dataIndex, cellIndex) {
    const stateStr = String(data || '');
    const state = stateStr.toLowerCase();
    let bgClass = 'state-badge state-other';
    if (state === 'accepted') bgClass = 'state-badge state-accepted';
    else if (state === 'rejected') bgClass = 'state-badge state-rejected';
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { class: bgClass },
        childNodes: [{ nodeName: '#text', data: stateStr }]
      }
    ];
  };

  // Custom column rendering for Track with badge styling
  const renderTrack = function(data, cell, dataIndex, cellIndex) {
    const trackStr = String(data || '');
    if (!trackStr) return;
    
    // Default to reviewed style, will be updated based on TrackType
    const bgClass = 'track-badge track-reviewed';
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { class: bgClass },
        childNodes: [{ nodeName: '#text', data: trackStr }]
      }
    ];
  };

  // Custom column rendering for Type with badge styling
  const renderType = function(data, cell, dataIndex, cellIndex) {
    const typeStr = String(data || '');
    if (!typeStr) return;
    
    cell.childNodes = [
      {
        nodeName: 'SPAN',
        attributes: { class: 'type-badge' },
        childNodes: [{ nodeName: '#text', data: typeStr }]
      }
    ];
  };

  // Row render to style track based on TrackType
  const rowRender = function(row, tr, index) {
    // Check TrackType (column 6) to style Track column (column 5)
    const trackType = row.cells[6]?.data;
    if (trackType === 'accepted' && tr.childNodes && tr.childNodes[5]) {
      // Find the track cell and update its class
      const trackCell = tr.childNodes[5];
      if (trackCell.childNodes && trackCell.childNodes[0]?.attributes) {
        trackCell.childNodes[0].attributes.class = 'track-badge track-accepted';
      }
    }
    return tr;
  };

  // DataTable options with column customization
  const dataTableOptions = {
    searchable: true,
    sortable: true,
    paging: true,
    perPage: 25,
    perPageSelect: [10, 25, 50, 100],
    rowRender: rowRender,
    columns: [
      { select: 0, sortable: true, type: 'number' },  // ID
      { select: 1, sortable: true, type: 'string' },  // Title
      { select: 2, render: renderState, sortable: true, type: 'string' },  // State
      { select: 3, sortable: true, type: 'string' },  // Submitter
      { select: 4, sortable: true, type: 'string' },  // Affiliation
      { select: 5, render: renderTrack, sortable: true, type: 'string' },  // Track
      { select: 6, hidden: true, type: 'string' },  // TrackType (hidden helper column)
      { select: 7, render: renderType, sortable: true, type: 'string' },  // Type
      { select: 8, sortable: true, type: 'number' },  // Score
      { select: 9, sortable: true, type: 'string' },  // Submitted
      { select: 10, sortable: true, type: 'string' }  // Authors
    ]
  };

  $: tableItems = getTableItems(abstractData);
</script>

<section class="mt-12 p-4 abstract-table-view">
  <Table items={tableItems} dataTableOptions={dataTableOptions} />
</section>

<style>
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
  :global(.track-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
  }

  :global(.track-accepted) {
    background-color: #dcfce7 !important;
    color: #166534 !important;
  }

  :global(.track-reviewed) {
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }

  /* Type badge styling */
  :global(.type-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #e0f2fe !important;
    color: #0369a1 !important;
  }

  /* Bootstrap-like table styling */
  .abstract-table-view :global(.datatable-wrapper) {
    font-size: 0.875rem;
  }

  .abstract-table-view :global(.datatable-table) {
    width: 100%;
    border-collapse: collapse;
  }

  .abstract-table-view :global(.datatable-table thead th) {
    background-color: #f8f9fa;
    border-bottom: 2px solid #dee2e6;
    padding: 0.75rem;
    text-align: left;
    font-weight: 600;
    white-space: nowrap;
  }

  .abstract-table-view :global(.datatable-table tbody tr:nth-child(odd)) {
    background-color: rgba(0, 0, 0, 0.02);
  }

  .abstract-table-view :global(.datatable-table tbody tr:hover) {
    background-color: rgba(0, 0, 0, 0.075);
  }

  .abstract-table-view :global(.datatable-table tbody td) {
    padding: 0.5rem 0.75rem;
    border-top: 1px solid #dee2e6;
    vertical-align: middle;
  }

  /* Compact styling */
  .abstract-table-view :global(.datatable-table.compact tbody td) {
    padding: 0.3rem 0.5rem;
  }

  /* Search input styling */
  .abstract-table-view :global(.datatable-input) {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    border: 1px solid #ced4da;
    border-radius: 0.25rem;
  }

  .abstract-table-view :global(.datatable-input:focus) {
    border-color: #86b7fe;
    outline: 0;
    box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
  }

  /* Pagination styling */
  .abstract-table-view :global(.datatable-pagination) {
    display: flex;
    gap: 0.25rem;
    margin-top: 1rem;
  }

  .abstract-table-view :global(.datatable-pagination li a),
  .abstract-table-view :global(.datatable-pagination li button) {
    padding: 0.375rem 0.75rem;
    border: 1px solid #dee2e6;
    border-radius: 0.25rem;
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

  :global(.dark .type-badge) {
    background-color: #0c4a6e !important;
    color: #bae6fd !important;
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
