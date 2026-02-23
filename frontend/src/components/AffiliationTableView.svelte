<script>
  import { DataTable, DataTableControls } from '@zhangt58/svelte-vtable';
  import AbstractDetailsDialog from '../pages/AbstractDetailsDialog.svelte';
  import AbstractTooltip from './AbstractTooltip.svelte';

  // Props: array of abstract objects (same shape returned by GetAbstracts)
  let { abstractData = [] } = $props();

  // Table controls state
  let searchQuery = $state('');
  let currentPage = $state(1);
  let perPage = $state(25);
  let sortKey = $state(null); // e.g. 'Affiliation' | 'Count'
  let sortDir = $state('desc'); // 'asc' | 'desc'
  let activeFilters = $state({});
  // Build columnFilters from affiliationData so DataTableControls can render per-column filters
  function getUniqueValuesWithCounts(items, key) {
    const counts = {};
    (items || []).forEach((it) => {
      const val = it && it[key];
      if (Array.isArray(val)) {
        val.forEach((v) => {
          const s = String(v ?? '');
          counts[s] = (counts[s] || 0) + 1;
        });
      } else {
        const s = String(val ?? '');
        counts[s] = (counts[s] || 0) + 1;
      }
    });
    const uniqueValues = Object.keys(counts).sort();
    return { uniqueValues, counts };
  }

  // columnFilters should be a callable derived rune (like other components)
  let columnFilters = $derived(() => {
    return columns.map((c) => {
      const { uniqueValues, counts } = getUniqueValuesWithCounts(affiliationData || [], c.title);
      return { key: c.title, label: c.title, uniqueValues, counts };
    });
  });

  function handleFilterChange({ allFilters }) {
    activeFilters = { ...allFilters };
    currentPage = 1;
  }

  function matchesFilters(item, filters) {
    for (const [columnKey, selectedValues] of Object.entries(filters)) {
      if (!selectedValues || selectedValues.length === 0) continue;
      const itemValue = item[columnKey];
      if (Array.isArray(itemValue)) {
        const itemStrings = itemValue.map((v) => String(v ?? ''));
        if (!selectedValues.some((val) => itemStrings.includes(val))) return false;
      } else {
        const itemStr = String(itemValue ?? '');
        if (!selectedValues.includes(itemStr)) return false;
      }
    }
    return true;
  }

  // Columns definition
  const columns = [
    { id: 'Affiliation', title: 'Affiliation', stretch: 6 },
    { id: 'Count', title: 'Count', stretch: 1 },
    { id: 'Country', title: 'Country', stretch: 2 },
    { id: 'Continent', title: 'Continent', stretch: 2 },
    { id: 'Abstracts', title: 'Abstracts', stretch: 3 },
  ];

  const visibleKeys = columns.map((c) => c.title);

  // Build mappedColumns for rowSnippet rendering
  const mappedColumns = columns.map((c) => ({
    id: c.id,
    title: c.title,
    nowrap: false,
    stretch: c.stretch,
  }));

  // Build column widths mapping to pass to DataTable (using stretch weights)
  const colWidths = mappedColumns.reduce((acc, c) => {
    acc[c.title] = c.stretch;
    return acc;
  }, {});

  // Helpers: collect affiliation entries from abstracts (similar to AbstractChartView)
  function collectAffiliations(data) {
    const items = [];
    for (const a of data || []) {
      if (Array.isArray(a.persons)) {
        for (const p of a.persons) {
          if (p.affiliation) {
            const affiliationStr =
              typeof p.affiliation === 'string'
                ? p.affiliation
                : p.affiliation.name || p.affiliation.raw || String(p.affiliation);

            items.push({
              raw: affiliationStr,
              affiliation: affiliationStr,
              country_name: p.affiliation?.country_name || p.country_name || '',
              continent: p.affiliation?.continent || p.continent || '',
              abstract_id: a.id || a.ID || null,
              abstract_title: a.title || a.Title || '',
            });
          }
        }
      }
    }
    return items;
  }

  // Minimal mapping country -> continent used as fallback
  const countryToContinent = {
    'United States': 'North America',
    USA: 'North America',
    Canada: 'North America',
    Mexico: 'North America',
    China: 'Asia',
    Japan: 'Asia',
    'South Korea': 'Asia',
    India: 'Asia',
    Germany: 'Europe',
    France: 'Europe',
    Italy: 'Europe',
    Spain: 'Europe',
    Switzerland: 'Europe',
    'United Kingdom': 'Europe',
    Netherlands: 'Europe',
    Sweden: 'Europe',
    Belgium: 'Europe',
    Austria: 'Europe',
    Poland: 'Europe',
    Russia: 'Europe',
    Australia: 'Oceania',
    Brazil: 'South America',
    Argentina: 'South America',
  };

  function guessContinentFromCountry(country) {
    if (!country) return 'Other';
    return countryToContinent[country] || 'Other';
  }

  // Precompute a lookup map of abstracts with memoized speakerName/snippet/friendlyId
  let abstractMap = $derived.by(() => {
    const map = {};
    for (const a of abstractData || []) {
      const rawId = a.id ?? '';
      const id = String(rawId);

      // compute friendly id display
      const native = a.friendly_id ?? null;
      const friendlyId = native ? `#${native}` : `#${id}`;

      // compute speaker name (prefer Persons / persons and is_speaker)
      const personsField = a.persons ?? [];
      let speaker = null;
      if (Array.isArray(personsField) && personsField.length > 0) {
        speaker =
          personsField.find((p) => p?.is_speaker === true || p?.is_speaker === 'true') ||
          personsField[0];
      }
      let speakerName = '';
      if (speaker) {
        const first = speaker.first_name ?? '';
        const last = speaker.last_name ?? '';
        speakerName =
          first || last ? `${first}${first && last ? ' ' : ''}${last}`.trim() : speaker.name || '';
      }

      // compute snippet (first ~200 chars)
      const rawText = a.content ?? '';
      const rawStr = String(rawText || '');
      const snippet =
        rawStr.replace(/\s+/g, ' ').trim().slice(0, 200) + (rawStr.length > 200 ? '…' : '');

      // store a shallow copy with memoized fields to avoid re-computing
      map[id] = { ...a, _speakerName: speakerName, _snippet: snippet, _friendlyId: friendlyId };
    }
    return map;
  });

  // Derived: aggregate affiliations into unique rows with counts and examples
  const affiliationData = $derived.by(() => {
    const items = collectAffiliations(abstractData || []);
    const m = new Map();
    for (const it of items) {
      const key = it.affiliation || it.raw || '';
      if (!key) continue;
      if (!m.has(key)) {
        m.set(key, {
          affiliation: key,
          count: 0,
          country_name: it.country_name || '',
          continent: it.continent || guessContinentFromCountry(it.country_name),
          examples: new Set(),
        });
      }
      const e = m.get(key);
      e.count += 1;
      if (!e.country_name && it.country_name) e.country_name = it.country_name;
      if (!e.continent && it.continent)
        e.continent = it.continent || guessContinentFromCountry(it.country_name);
      if (it.abstract_id) e.examples.add(String(it.abstract_id));
    }

    // Convert to array of table rows (keep full list of abstract IDs)
    const arr = Array.from(m.values()).map((v) => ({
      Affiliation: v.affiliation,
      Count: v.count,
      Country: v.country_name || '',
      Continent: v.continent || '',
      AbstractIds: Array.from(v.examples),
    }));

    // Default sort by count desc
    arr.sort((a, b) => b.Count - a.Count);
    return arr;
  });

  // Paged, filtered, sorted view
  const paged = $derived.by(() => {
    let arr = affiliationData || [];

    // Simple search across multiple fields
    if (searchQuery && String(searchQuery).trim() !== '') {
      const q = String(searchQuery).toLowerCase();
      arr = arr.filter((it) => {
        const abstractsText = Array.isArray(it.AbstractIds) ? it.AbstractIds.join(' ') : '';
        return (
          (it.Affiliation && String(it.Affiliation).toLowerCase().includes(q)) ||
          (it.Country && String(it.Country).toLowerCase().includes(q)) ||
          (it.Continent && String(it.Continent).toLowerCase().includes(q)) ||
          (abstractsText && abstractsText.toLowerCase().includes(q))
        );
      });
    }

    // Filtering by active filters
    if (Object.keys(activeFilters).length > 0) {
      arr = arr.filter((it) => matchesFilters(it, activeFilters));
    }

    // Sorting by key if requested
    if (sortKey) {
      arr = [...arr].sort((a, b) => {
        const ak = a[sortKey];
        const bk = b[sortKey];
        if (typeof ak === 'number' && typeof bk === 'number') {
          return sortDir === 'asc' ? ak - bk : bk - ak;
        }
        return sortDir === 'asc'
          ? String(ak || '').localeCompare(String(bk || ''))
          : String(bk || '').localeCompare(String(ak || ''));
      });
    }

    const total = arr.length;
    const start = Math.max(0, (currentPage - 1) * perPage);
    const pageItems = arr.slice(start, start + perPage);
    return { items: pageItems, total };
  });

  // convenience binds used in template
  const totalItems = $derived(paged.total);
  const visibleItems = $derived(paged.items);

  // Sorting callback from DataTable
  function onSort(payload) {
    // DataTable may call sortCallback with a string key or an object { key, dir }
    const key = typeof payload === 'string' ? payload : (payload?.key ?? null);
    if (!key) return;
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
    currentPage = 1;
  }

  // Abstract details dialog state (for clicking abstracts in the table)
  let showAbstractDialog = $state(false);
  let selectedAbstract = $state(null);
  let selectedAbstractId = $state(null);
  let selectedAffiliationAbstractIds = $state([]); // full list of abstract ids for current affiliation
  let currentDialogIndex = $state(-1);

  function findAbstractById(id) {
    const sid = String(id);
    return (abstractData || []).find((a) => String(a.id) === sid || String(a.ID) === sid);
  }

  // Friendly ID helpers
  function getFriendlyId(id) {
    const a = findAbstractById(id);
    if (!a) return `#${String(id)}`;
    const native = a.friendly_id ?? null;
    return native ? `#${native}` : `#${String(id)}`;
  }

  function openAbstract(id, affiliationList = null) {
    const sid = String(id);
    selectedAbstractId = sid;
    selectedAbstract = findAbstractById(sid);
    selectedAffiliationAbstractIds = affiliationList ? affiliationList.map((x) => String(x)) : [];
    currentDialogIndex = selectedAffiliationAbstractIds.findIndex((x) => String(x) === sid);
    if (selectedAbstract) showAbstractDialog = true;
  }

  function navigateDialog(direction) {
    if (
      !Array.isArray(selectedAffiliationAbstractIds) ||
      selectedAffiliationAbstractIds.length === 0
    )
      return;
    let nextIndex = direction === 'prev' ? currentDialogIndex - 1 : currentDialogIndex + 1;
    if (nextIndex < 0 || nextIndex >= selectedAffiliationAbstractIds.length) return;
    const nextId = selectedAffiliationAbstractIds[nextIndex];
    selectedAbstractId = String(nextId);
    selectedAbstract = findAbstractById(nextId);
    currentDialogIndex = nextIndex;
  }
</script>

<div class="flex flex-col h-full">
  <div
    class="sticky top-0 z-10 bg-transparent px-2 py-2 mb-2 mt-2 shrink-0 shadow-md dark:shadow-black/40"
  >
    <DataTableControls
      search={searchQuery}
      {currentPage}
      bind:perPage
      {totalItems}
      pagechange={(payload) => {
        currentPage = payload.currentPage;
      }}
      searchchange={(payload) => {
        searchQuery = payload.search;
      }}
      columnFilters={columnFilters()}
      {activeFilters}
      filterChange={handleFilterChange}
      filtersVisible={false}
    />
  </div>

  <section class="flex-1 overflow-auto flex flex-col max-h-screen min-h-0">
    <DataTable
      items={visibleItems}
      {visibleKeys}
      {sortKey}
      {sortDir}
      sortCallback={onSort}
      className="w-full mt-0.5 mb-4 overflow-auto min-h-0"
      {colWidths}
      virtualize={false}
      {rowSnippet}
    />
  </section>
</div>

<!-- Row snippet -->
{#snippet rowSnippet({ item, index, select, selected })}
  <tr
    onclick={() => {
      try {
        select && select();
      } catch (e) {}
    }}
    tabindex="0"
    class:selected-row={selected && String(selected.Affiliation) === String(item.Affiliation)}
    aria-selected={selected && String(selected.Affiliation) === String(item.Affiliation)}
  >
    {#each mappedColumns as col}
      <td class={col.nowrap ? 'nowrap' : ''}>
        {#if col.id === 'Affiliation'}
          <div class="wrap-break-word">{item.Affiliation}</div>
        {:else if col.id === 'Count'}
          {item.Count}
        {:else if col.id === 'Country'}
          {item.Country}
        {:else if col.id === 'Continent'}
          {item.Continent}
        {:else if col.id === 'Abstracts'}
          <span class="text-sm text-gray-600">
            {#if Array.isArray(item.AbstractIds) && item.AbstractIds.length > 0}
              {#each item.AbstractIds.slice(0, 5) as abstractId, i}
                <button
                  type="button"
                  onclick={(e) => {
                    e.stopPropagation();
                    openAbstract(abstractId, item.AbstractIds);
                  }}
                  class="text-blue-600 hover:underline bg-transparent p-0 m-0 align-baseline"
                  aria-label={'Open abstract ' +
                    (findAbstractById(abstractId)?.title || abstractId)}
                >
                  {getFriendlyId(abstractId)}
                </button>
                <AbstractTooltip abstract={abstractMap[abstractId]} />
                {#if i < Math.min(4, item.AbstractIds.length - 1)}{', '}{/if}
              {/each}
              {#if item.AbstractIds.length > 5}
                , ...+{item.AbstractIds.length - 5}
              {/if}
            {/if}
          </span>
        {/if}
      </td>
    {/each}
  </tr>
{/snippet}

{#if showAbstractDialog && selectedAbstract}
  <AbstractDetailsDialog
    bind:open={showAbstractDialog}
    bind:abstract={selectedAbstract}
    currentIndex={currentDialogIndex}
    totalCount={selectedAffiliationAbstractIds.length}
    onNavigate={(direction) => navigateDialog(direction)}
  />
{/if}
