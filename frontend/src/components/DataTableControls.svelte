<script>
  import { PaginationNav, Search, Badge } from 'flowbite-svelte';

  // Props for search, pagination (but no per-page control)
  let {
    search = '',
    currentPage = 1,
    perPage = 25,  // Fixed value, not user-configurable
    totalItems = 0,
    pagechange = () => {},
    searchchange = () => {}
  } = $props();

  // totalPages derived from totalItems and perPage
  const totalPages = $derived(() => Math.max(1, Math.ceil(totalItems / perPage)));

  // clamp currentPage into valid range whenever currentPage or totalPages change
  $effect(() => {
    currentPage = Math.min(Math.max(1, +currentPage || 1), totalPages());
  });

  function goTo(p) {
    const np = Math.min(Math.max(1, Math.floor(p)), totalPages());
    if (np !== currentPage) {
      currentPage = np;
      pagechange?.({ currentPage });
    }
  }

  // emit search changes for parent when search updates
  $effect(() => {
    if (typeof search !== 'undefined') searchchange?.({ search });
  });

  // Calculate range for display
  const startItem = $derived(() => totalItems === 0 ? 0 : (currentPage - 1) * perPage + 1);
  const endItem = $derived(() => Math.min(currentPage * perPage, totalItems));

</script>

<div class="flex items-center w-full gap-3 px-0 py-1">
  <!-- Search field on the left -->
  <div class="flex-1">
    <Search size="sm" bind:value={search} placeholder="Search..." clearable
            clearableOnClick={() => { search = ''; searchchange?.({ search }); }} />
  </div>

  <!-- Range count badge -->
  <div class="flex-1">
    <Badge class="text-md rounded" color="gray">
      Showing {startItem()} to {endItem()} of {totalItems}
    </Badge>
  </div>

  <!-- Pagination navigation (no per-page dropdown) -->
  <PaginationNav
    currentPage={currentPage}
    totalPages={totalPages()}
    visiblePages={5}
    onPageChange={(p) => goTo(p)}
    showIcons={true}
    ariaLabel="Pagination"
    classes={{ active: "bg-green-100 dark:bg-green-700 text-green-600 dark:text-white rounded-sm" }}
  />
</div>
