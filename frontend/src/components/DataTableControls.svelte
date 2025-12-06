<script>
  import { createEventDispatcher } from 'svelte';
  import { Button, PaginationNav, Dropdown, DropdownItem, Search, Badge } from 'flowbite-svelte';
  import { ChevronDownOutline } from 'flowbite-svelte-icons';
  export let search = '';
  export let currentPage = 1;
  export let perPage = 25;
  export let totalItems = 0;
  export let perPageOptions = [10, 25, 50, 100];

  const dispatch = createEventDispatcher();

  $: totalPages = Math.max(1, Math.ceil(totalItems / perPage));
  $: currentPage = Math.min(Math.max(1, +currentPage || 1), totalPages);

  $: startIndex = totalItems === 0 ? 0 : (currentPage - 1) * perPage + 1;
  $: endIndex = Math.min(totalItems, currentPage * perPage);

  function goTo(p) {
    const np = Math.min(Math.max(1, Math.floor(p)), totalPages);
    if (np !== currentPage) {
      currentPage = np;
      dispatch('pagechange', { currentPage });
    }
  }

  function setPerPage(n) {
    perPage = n;
    currentPage = 1;
    dispatch('perpagechange', { perPage });
    dispatch('pagechange', { currentPage });
  }

  // emit search changes for parent when search updates (bind also works)
  $: if (typeof search !== 'undefined') dispatch('searchchange', { search });

  // compute visible page numbers around current page
  function pageRange(windowSize = 5) {
    const pages = [];
    if (totalPages <= windowSize) {
      for (let i = 1; i <= totalPages; i++) pages.push(i);
      return pages;
    }
    const half = Math.floor(windowSize / 2);
    let start = Math.max(1, currentPage - half);
    let end = Math.min(totalPages, currentPage + half);
    if (start === 1) end = start + windowSize - 1;
    if (end === totalPages) start = end - windowSize + 1;
    for (let i = start; i <= end; i++) pages.push(i);
    return pages;
  }

  // cache pages so template doesn't call pageRange() repeatedly
  $: pages = pageRange();

  // visible pages for reference (PaginationNav computes internally; pages kept for possible use)
</script>

<div class="flex items-center w-full gap-3 px-0 py-1">
  <!-- Search field on the left -->
  <div class="flex-1">
    <Search size="sm" bind:value={search} placeholder="Search..." clearable />
  </div>
  <div class="flex-1">
    <Badge class="text-md rounded" color="gray">Total: {totalItems}</Badge>
  </div>

  <PaginationNav
    currentPage={currentPage}
    totalPages={totalPages}
    visiblePages={5}
    onPageChange={(p) => goTo(p)}
    showIcons={true}
    ariaLabel="Pagination"
    classes={{ active: "bg-green-100 dark:bg-green-700 text-green-600 dark:text-white" }}
  />

  <div class="flex items-center gap-2">
    <Button slot="trigger" color="light" class="pl-3 pr-0 py-1 border rounded-md text-md bg-white text-gray-700 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200">
        {perPage} / page
        <ChevronDownOutline class="ms-2 h-4 w-4 text-gray-700 dark:text-gray-200" />
    </Button>
    <Dropdown simple placement="top" offset={6}>
      {#each perPageOptions as opt}
        <DropdownItem onclick={() => setPerPage(opt)}>{opt} per page</DropdownItem>
      {/each}
    </Dropdown>
  </div>
</div>
