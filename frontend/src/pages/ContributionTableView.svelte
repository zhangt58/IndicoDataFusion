<script>
  import { Table } from '@flowbite-svelte-plugins/datatable';
  import ContributionDetailsDialog from './ContributionDetailsDialog.svelte';
  import { 
    getTableItems, 
    createDataTableOptions 
  } from './ContributionTableItem.js';

  /** @type {Array} */
  export let contributionData = [];

  // Contribution dialog state (can be extended later for contribution details dialog)
  let showContributionDialog = false;
  let selectedContribution = null;

  // Find contribution by ID
  function findContributionById(id) {
    return contributionData.find(c => (c.friendly_id || c.id) == id);
  }

  // Handle clicks on the table
  function handleTableClick(event) {
    const target = event.target;
    
    // Handle title link click
    if (target.classList.contains('title-link')) {
      event.preventDefault();
      const contributionId = target.dataset.id;
      console.log('Title clicked, contributionId:', contributionId);
      if (!contributionId) {
        console.warn('No contribution ID found in data-id attribute');
        return;
      }
      selectedContribution = findContributionById(contributionId);
      console.log('Found contribution:', selectedContribution);
      if (selectedContribution) {
        showContributionDialog = true;
      }
    }
  }

  $: tableItems = getTableItems(contributionData);
  $: dataTableOptions = createDataTableOptions();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<section class="mt-12 p-4 contribution-table-view" on:click={handleTableClick}>
  <Table items={tableItems} dataTableOptions={dataTableOptions} />
</section>

<!-- Contribution Detail Dialog -->
<ContributionDetailsDialog bind:open={showContributionDialog} contribution={selectedContribution} />

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

  /* Type badge styling */
  :global(.type-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #e0e7ff !important;
    color: #3730a3 !important;
  }

  /* Session badge styling */
  :global(.session-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #f3e8ff !important;
    color: #6b21a8 !important;
  }

  /* Track badge styling */
  :global(.track-badge) {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    background-color: #dcfce7 !important;
    color: #166534 !important;
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

  /* Dark mode support */
  :global(.dark .type-badge) {
    background-color: #3730a3 !important;
    color: #e0e7ff !important;
  }

  :global(.dark .session-badge) {
    background-color: #581c87 !important;
    color: #e9d5ff !important;
  }

  :global(.dark .track-badge) {
    background-color: #14532d !important;
    color: #bbf7d0 !important;
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
