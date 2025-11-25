<script>
  import { Table } from '@flowbite-svelte-plugins/datatable';

  /** @type {Array} */
  export let abstractData = [];

  // Transform abstract data for the table view
  function getTableItems(data) {
    return data.map(abstract => ({
      ID: abstract.friendly_id || abstract.id,
      Title: abstract.title || '',
      State: abstract.state || '',
      Submitter: abstract.submitter?.full_name || '',
      Affiliation: abstract.submitter?.affiliation || '',
      Track: abstract.accepted_track?.title || abstract.reviewed_for_tracks?.[0]?.title || '',
      Type: abstract.accepted_contrib_type?.name || '',
      Score: abstract.score ?? '',
      Submitted: abstract.submitted_dt || '',
      Authors: abstract.persons?.map(p => `${p.first_name} ${p.last_name}`).join(', ') || ''
    }));
  }

  // DataTable options
  const dataTableOptions = {
    searchable: true,
    sortable: true,
    paging: true,
    perPage: 20,
    perPageSelect: [10, 20, 50, 100]
  };

  $: tableItems = getTableItems(abstractData);
</script>

<section class="mt-12 p-4">
  <Table items={tableItems} {dataTableOptions} />
</section>
