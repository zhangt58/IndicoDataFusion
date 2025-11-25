<script>
  import { Table } from '@flowbite-svelte-plugins/datatable';
  import AbstractDetailsDialog from './AbstractDetailsDialog.svelte';
  import TrackDetailsDialog from './TrackDetailsDialog.svelte';
  import { 
    getTableItems, 
    createDataTableOptions 
  } from './AbstractTableItem.js';

  /** @type {Array} */
  export let abstractData = [];

  // Abstract dialog state
  let showAbstractDialog = false;
  let selectedAbstract = null;

  // Track dialog state
  let showTrackDialog = false;
  let selectedTracks = [];

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
    return abstractData.find(a => (a.friendly_id || a.id) == id);
  }

  // Handle clicks on the table
  function handleTableClick(event) {
    const target = event.target;
    
    // Handle title link click
    if (target.classList.contains('title-link')) {
      event.preventDefault();
      const abstractId = target.dataset.id;
      selectedAbstract = findAbstractById(abstractId);
      if (selectedAbstract) {
        showAbstractDialog = true;
      }
    }
    
    // Handle track link click
    if (target.classList.contains('track-link')) {
      event.preventDefault();
      const tracksData = target.dataset.tracks;
      try {
        selectedTracks = JSON.parse(tracksData || '[]');
        if (selectedTracks.length > 0) {
          showTrackDialog = true;
        }
      } catch (e) {
        console.error('Failed to parse tracks data:', e);
      }
    }
  }

  $: tableItems = getTableItems(abstractData);
  $: dataTableOptions = createDataTableOptions();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<section class="mt-12 p-4 abstract-table-view" on:click={handleTableClick}>
  <Table items={tableItems} dataTableOptions={dataTableOptions} />
</section>

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
  :global(.track-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
  }

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
