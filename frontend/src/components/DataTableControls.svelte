<script>
  import { Button, PaginationNav, Dropdown, DropdownItem, Search, Badge } from 'flowbite-svelte';
  import { ChevronDownOutline } from 'flowbite-svelte-icons';

  // Declare props via $props() per Svelte v5 migration. Use names expected by parent components.
  let {
    search = '',
    currentPage = 1,
    perPage = 25,
    totalItems = 0,
    perPageOptions = [10, 25, 50, 100],
    perpagechange = () => {},
    pagechange = () => {},
    searchchange = () => {}
  } = $props();

  // totalPages derived from totalItems and perPage (runes-friendly)
  const totalPages = $derived(() => Math.max(1, Math.ceil(totalItems / perPage)));

  // clamp currentPage into valid range whenever currentPage or totalPages change
  $effect(() => {
    currentPage = Math.min(Math.max(1, +currentPage || 1), totalPages());
  });

  function goTo(p) {
    const np = Math.min(Math.max(1, Math.floor(p)), totalPages());
    if (np !== currentPage) {
      currentPage = np;
      // call the parent-provided callback prop, if any
      pagechange?.({ currentPage });
    }
  }

  function setPerPage(n) {
    perPage = n;
    currentPage = 1;
    perpagechange?.({ perPage });
    pagechange?.({ currentPage });
  }

  // emit search changes for parent when search updates (bind also works)
  $effect(() => {
    if (typeof search !== 'undefined') searchchange?.({ search });
  });

</script>

<div class="flex items-center w-full gap-3 px-0 py-1">
  <!-- Search field on the left -->
  <div class="flex-1">
    <Search size="sm" bind:value={search} placeholder="Search..." clearable
            clearableOnClick={() => { search = ''; searchchange?.({ search }); }} />
  </div>
  <div class="flex-1">
    <Badge class="text-md rounded" color="gray">Total: {totalItems}</Badge>
  </div>

  <PaginationNav
    currentPage={currentPage}
    totalPages={totalPages()}
    visiblePages={5}
    onPageChange={(p) => goTo(p)}
    showIcons={true}
    ariaLabel="Pagination"
    classes={{ active: "bg-green-100 dark:bg-green-700 text-green-600 dark:text-white rounded-sm" }}
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
