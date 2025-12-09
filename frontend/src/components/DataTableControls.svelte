<script>
  import { Pagination, Search, Badge } from 'flowbite-svelte';
  import { ChevronLeftOutline, ChevronRightOutline } from 'flowbite-svelte-icons';
  import { onMount, onDestroy } from 'svelte';

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

  // visiblePages: number of middle numeric buttons to show (not counting first/last)
  const visiblePages = 5;

  // stable id to query the pagination DOM if needed (fallback)
  const paginationId = `pagination-${Math.random().toString(36).slice(2,9)}`;

  // pages array for Pagination component (condensed with ellipses)
  const pages = $derived(() => {
    const total = totalPages();
    const current = Math.min(Math.max(1, Math.floor(currentPage)), total);

    // If small number of pages, show all
    if (total <= visiblePages + 2) {
      return Array.from({ length: total }, (_, i) => ({ name: String(i + 1), active: i + 1 === current }));
    }

    const pagesArr = [];
    // always show first
    pagesArr.push({ name: '1', active: current === 1 });

    const half = Math.floor(visiblePages / 2);
    let start = Math.max(2, current - half);
    let end = Math.min(total - 1, current + half);

    // adjust window when near the boundaries
    if (current - 1 <= half) {
      start = 2;
      end = 1 + visiblePages;
    }
    if (total - current <= half) {
      end = total - 1;
      start = total - visiblePages;
    }

    if (start > 2) {
      // compute left jump (midpoint of hidden range 2..start-1)
      const leftHiddenStart = 2;
      const leftHiddenEnd = start - 1;
      const leftJump = Math.max(leftHiddenStart, Math.floor((leftHiddenStart + leftHiddenEnd) / 2));
      pagesArr.push({ name: '…', active: false, ellipsis: 'left', jump: leftJump });
    }

    for (let i = start; i <= end; i++) pagesArr.push({ name: String(i), active: i === current });

    if (end < total - 1) {
      // compute right jump (midpoint of hidden range end+1..total-1)
      const rightHiddenStart = end + 1;
      const rightHiddenEnd = total - 1;
      const rightJump = Math.min(rightHiddenEnd, Math.ceil((rightHiddenStart + rightHiddenEnd) / 2));
      pagesArr.push({ name: '…', active: false, ellipsis: 'right', jump: rightJump });
    }

    // always show last
    pagesArr.push({ name: String(total), active: current === total });

    return pagesArr;
  });

  // previous / next callbacks for Pagination
  function previous() {
    goTo(currentPage - 1);
  }
  function next() {
    goTo(currentPage + 1);
  }

  // DOM reference to the pagination container for decorating ellipsis buttons
  let paginationContainerEl = null;

  function decorateEllipsis() {
    if (!paginationContainerEl) return;
    const pageList = pages();
    // find the <ul> inside the pagination container
    let ul = null;
    if (paginationContainerEl && typeof paginationContainerEl.querySelector === 'function') {
      ul = paginationContainerEl.querySelector('ul');
    } else {
      // fallback: query by data attribute on the wrapper
      ul = document.querySelector(`[data-pagination-id="${paginationId}"] ul`);
    }
    if (!ul) return;
    // Ensure the <ul> has Tailwind utility classes for layout
    ul.classList.remove('flex','gap-0.5','p-0','m-0','list-none','items-center');
    ul.classList.add('flex','gap-0.5','p-0','m-0','list-none','items-center');
    const lis = Array.from(ul.children || []);
    // Build a list of page-button <li>s (skip prev/next controls) by checking their textContent
    const pageLis = [];
    for (const li of lis) {
      const btn = li.querySelector('button, a');
      if (!btn) continue;
      const text = (btn.textContent || '').trim();
      // consider numeric pages and ellipsis as page buttons
      if (/^\d+$/.test(text) || text === '…') pageLis.push({ li, btn, text });
    }

    // Now map pageLis to pageList entries in order
    pageLis.forEach(({ li, btn, text }, idx) => {
      const pageObj = pageList[idx];
      if (!pageObj) return;
      // clear previous decorations
      btn.classList.remove('ellipsis-page');
      btn.removeAttribute('data-jump');
      btn.removeAttribute('data-ellipsis');
      btn.removeAttribute('title');

      // Apply Tailwind utility classes instead of inline styles
      // Remove any previously-applied utility classes we manage
      btn.classList.remove(
        'px-2','py-1','text-sm','border','border-gray-300','bg-transparent','min-w-[2rem]',
        'inline-flex','items-center','justify-center','transition-colors','hover:bg-gray-100',
        'bg-green-100','text-green-600','border-transparent','opacity-60','cursor-default','pointer-events-none',
        'tw-ellipsis'
      );

      // Base Tailwind classes for pagination buttons
      btn.classList.add(
        'px-2','py-1','text-sm','border','border-gray-300','bg-transparent','min-w-[2rem]',
        'inline-flex','items-center','justify-center','transition-colors','hover:bg-gray-100'
      );

      // active page styling
      if (pageObj.active) {
        btn.classList.add('bg-green-100','text-green-600','border-transparent');
        btn.setAttribute('aria-current', 'page');
      } else {
        btn.removeAttribute('aria-current');
        btn.classList.remove('bg-green-100','text-green-600','border-transparent');
      }

      // disabled styling
      if (btn.disabled || btn.getAttribute('aria-disabled') === 'true') {
        btn.classList.add('opacity-60','cursor-default','pointer-events-none');
      } else {
        btn.classList.remove('opacity-60','cursor-default','pointer-events-none');
      }

      if (pageObj.name === '…' && typeof pageObj['jump'] === 'number') {
        btn.setAttribute('data-jump', String(pageObj['jump']));
        btn.setAttribute('data-ellipsis', 'true');
        // restore native tooltip for compatibility (browser tooltip)
        btn.setAttribute('title', `Jump to page ${pageObj['jump']}`);
        // visually indicate ellipsis (smaller text) via Tailwind
        btn.classList.add('text-sm','opacity-90');
      }
    });
  }

  // run decoration whenever pages or currentPage change
  $effect(() => {
    pages();
    currentPage;
    // schedule decoration after the browser has painted the updated DOM
    requestAnimationFrame(() => requestAnimationFrame(() => decorateEllipsis()));
  });

  // Re-apply decorations whenever the pagination DOM mutates (childList changes)
  let _observer = null;
  onMount(() => {
    if (!paginationContainerEl) return;
    const run = () => requestAnimationFrame(() => requestAnimationFrame(() => decorateEllipsis()));
    _observer = new MutationObserver(run);
    const root = paginationContainerEl instanceof Element ? paginationContainerEl : document.querySelector(`[data-pagination-id="${paginationId}"]`);
    if (root) _observer.observe(root, { childList: true, subtree: true, attributes: true });
    // initial run
    run();
  });

  onDestroy(() => {
    _observer?.disconnect();
    _observer = null;
  });

  // handle clicks inside Pagination to detect page button clicks
  function handlePaginationClick(e) {
    const el = e.target instanceof Element ? e.target.closest('button, a') : null;
    if (!el) return;
    // find the parent li index so we can map to pages()
    const li = el.closest('li');
    if (!li) return;
    const ul = li.parentElement;
    if (!ul) return;
    const lis = Array.from(ul.children);
    // Find prev/next by detecting non-numeric/non-ellipsis buttons
    const pageList = pages();
    // Build ordered list of page-button lis (numeric and ellipsis)
    const pageLis = [];
    for (const item of lis) {
      const btn = item.querySelector('button, a');
      if (!btn) continue;
      const text = (btn.textContent || '').trim();
      if (/^\d+$/.test(text) || text === '…') pageLis.push({ li: item, btn, text });
    }

    // If clicked on prev/next controls, detect by checking whether clicked li is not in pageLis
    const clickedIsPage = pageLis.some(p => p.li === li);
    if (!clickedIsPage) {
      // determine if it was prev (first li) or next (last li)
      const firstLi = lis[0];
      const lastLi = lis[lis.length - 1];
      if (li === firstLi) {
        previous();
      } else if (li === lastLi) {
        next();
      }
      return;
    }

    // Map clicked li to page index
    const pageIdx = pageLis.findIndex(p => p.li === li);
    if (pageIdx < 0 || pageIdx >= pageList.length) return;
    const pageObj = pageList[pageIdx];
    if (!pageObj) return;
    if (pageObj.name === '…' && typeof pageObj['jump'] === 'number') {
      goTo(pageObj['jump']);
      return;
    }
    const n = Number(pageObj.name);
    if (!Number.isNaN(n)) goTo(n);
  }

  // No external tooltip component in use; we rely on native title attribute for ellipses
</script>

<div class="flex items-center w-full gap-3 px-0 py-1">
  <!-- Search field on the left -->
  <div class="flex-1">
    <Search size="sm" bind:value={search} placeholder="Search..." clearable
            clearableOnClick={() => { search = ''; searchchange?.({ search }); }} />
  </div>

  <!-- Range count badge -->
  <div class="flex-1">
    <Badge rounded color="gray">
      Showing {startItem()} to {endItem()} of {totalItems}
    </Badge>
  </div>

  <!-- Pagination navigation: use Pagination with chevrons and page buttons -->
  <div bind:this={paginationContainerEl} data-pagination-id={paginationId} class="flex items-center gap-0.5" aria-label="Pagination">
    <Pagination
      pages={pages()}
      {previous}
      {next}
      ariaLabel="Pagination"
      onclick={handlePaginationClick}
    >
      {#snippet prevContent()}
        <span class="sr-only">Previous</span>
        <ChevronLeftOutline class="shrink-0 h-6 w-6" />
      {/snippet}

      {#snippet nextContent()}
        <span class="sr-only">Next</span>
        <ChevronRightOutline class="shrink-0 h-6 w-6" />
      {/snippet}
    </Pagination>
  </div>
</div>

<style>
  /* Force the active/current page to use the green tokens (override flowbite's primary/orange).
     Intentionally specific to reliably win over third-party rules. */
  :global([data-pagination-id] ul li [aria-current="page"]) {
    background-color: #dcfce7 !important; /* bg-green-100 */
    color: #16a34a !important;           /* text-green-600 */
    border-color: transparent !important;
  }

  /* Ensure Prev/Next controls visually match the numeric buttons' border and rhythm
     (keeps border, radius and min-width consistent). This is minimal and won't
     override hover/text colors managed by Tailwind classes applied at runtime. */
  :global([data-pagination-id] ul li:first-child > button),
  :global([data-pagination-id] ul li:first-child > a),
  :global([data-pagination-id] ul li:last-child > button),
  :global([data-pagination-id] ul li:last-child > a) {
    border: 1px solid #d1d5db; /* border-gray-300 */
    min-width: 2rem; /* align with numeric buttons' min width */
    background: transparent;
    padding: 0.25rem 0.5rem; /* keep comfortable hit area */
  }

  /* Light mode: hover/focus for Prev/Next match numeric button hover */
  :global([data-pagination-id] ul li:first-child > button:hover),
  :global([data-pagination-id] ul li:first-child > a:hover),
  :global([data-pagination-id] ul li:last-child > button:hover),
  :global([data-pagination-id] ul li:last-child > a:hover),
  :global([data-pagination-id] ul li:first-child > button:focus),
  :global([data-pagination-id] ul li:first-child > a:focus),
  :global([data-pagination-id] ul li:last-child > button:focus),
  :global([data-pagination-id] ul li:last-child > a:focus) {
    background-color: #f3f4f6; /* tailwind gray-100 */
    outline: none;
  }

  /* Ellipsis styling: make clickable and visually lighter (JS sets data-ellipsis/data-jump) */
  :global([data-pagination-id] ul li [data-ellipsis="true"]) {
    cursor: pointer;
    background: transparent !important;
    color: inherit;
    opacity: 0.9;
    font-size: 0.875rem;
    padding: 0.25rem 0.5rem;
  }

  /* Chevron sizing for Prev/Next so icons sit comfortably in the buttons */
  :global([data-pagination-id] ul li:first-child svg),
  :global([data-pagination-id] ul li:last-child svg) {
    height: 1.25rem;
    width: 1.25rem;
    display: block;
  }

  /* Dark mode: tone down chevron icon brightness for prev/next controls */
  :global(.dark [data-pagination-id] ul li:first-child svg),
  :global(.dark [data-pagination-id] ul li:last-child svg) {
    color: #9ca3af; /* gray-400 */
    stroke: currentColor;
    fill: none;
    opacity: 0.95;
  }

  /* Dark mode: adjust Prev/Next button border & hover to match dark theme */
  :global(.dark [data-pagination-id] ul li:first-child > button),
  :global(.dark [data-pagination-id] ul li:first-child > a),
  :global(.dark [data-pagination-id] ul li:last-child > button),
  :global(.dark [data-pagination-id] ul li:last-child > a) {
    border-color: #374151; /* tailwind gray-700 */
    background: transparent;
  }
  :global(.dark [data-pagination-id] ul li:first-child > button:hover),
  :global(.dark [data-pagination-id] ul li:first-child > a:hover),
  :global(.dark [data-pagination-id] ul li:last-child > button:hover),
  :global(.dark [data-pagination-id] ul li:last-child > a:hover),
  :global(.dark [data-pagination-id] ul li:first-child > button:focus),
  :global(.dark [data-pagination-id] ul li:first-child > a:focus),
  :global(.dark [data-pagination-id] ul li:last-child > button:focus),
  :global(.dark [data-pagination-id] ul li:last-child > a:focus) {
    background-color: rgba(255,255,255,0.03);
    border-color: #4b5563; /* tailwind gray-600 on hover */
    outline: none;
  }
</style>
