<script>
  import { Modal } from 'flowbite-svelte';
  import { CloseOutline } from 'flowbite-svelte-icons';

  /** @type {boolean} */
  export let open = false;

  /** @type {Array<{title?: string, name?: string, items?: any[]}>} */
  export let sessions = [];

  /** @type {Array<{title?: string, name?: string}>} */
  export let allSessions = [];

  // Local UI state (track per-session toggles / expansions by session title)
  let expandedOmitted = {};

  function toggleExpanded(title) {
    expandedOmitted = { ...expandedOmitted, [title]: !expandedOmitted[title] };
  }

  // Close dialog
  function closeDialog() {
    open = false;
  }

  // Unique sorted sessions (alphabetical)
  $: uniqueSessions = allSessions
    .filter((s, i, arr) => i === arr.findIndex(x => x.title === s.title))
    .sort((a,b) => a.title.localeCompare(b.title));

  // Check if a session is current
  function isCurrentSession(title) {
    return sessions.some(s => s.title === title);
  }

  // Helper: safely collect items for a session entry (session object may include `items` as passed from ContributionTableView)
  function collectItemsForSession(s) {
    if (!s) return [];
    if (Array.isArray(s.items)) return s.items;
    // some callers may pass the session as a string-only; no items available
    return [];
  }

  // Aggregate metadata from session items for display (location, time range, rooms, speakers)
  $: sessionAggregates = sessions.map(s => {
    const items = collectItemsForSession(s) || [];

    // helper to compute start millis from various possible fields
    function getItemStartMillis(it) {
      if (!it) return null;
      if (it.StartMillis != null && !isNaN(Number(it.StartMillis))) return Number(it.StartMillis);
      if (it.Start) {
        const p = Date.parse(it.Start);
        if (!isNaN(p)) return p;
      }
      if (it.StartDate) {
        const p = Date.parse(it.StartDate);
        if (!isNaN(p)) return p;
      }
      return null;
    }

    // helper to parse Duration like "HH:MM" into minutes when DurationMinutes is missing
    function parseDurationToMinutes(dur) {
      if (dur == null) return NaN;
      if (typeof dur === 'number') return dur;
      // common formats: "MM", "H:MM", "HH:MM"
      const parts = String(dur).split(':').map(x => Number(x));
      if (parts.length === 1 && !isNaN(parts[0])) return parts[0];
      if (parts.length === 2 && !isNaN(parts[0]) && !isNaN(parts[1])) return parts[0] * 60 + parts[1];
      return NaN;
    }

    // attach computed start to each item and sort; items with no start go last
    const itemsWithStart = items.map(it => ({ it, start: getItemStartMillis(it) }));
    itemsWithStart.sort((a,b) => {
      const sa = a.start;
      const sb = b.start;
      if (sa == null && sb == null) return 0;
      if (sa == null) return 1;
      if (sb == null) return -1;
      return sa - sb;
    });
    // Only include items that have a valid start in the visible list
    const itemsWithStartNonNull = itemsWithStart.filter(x => x.start != null);
    const itemsSorted = itemsWithStartNonNull.map(x => x.it);
    const excludedItems = itemsWithStart.filter(x => x.start == null).map(x => x.it);

    // normalize fields from items: try multiple casing variants
    const locations = new Set();
    const rooms = new Set();
    const speakers = new Set();
    const startMillis = [];
    const endMillis = [];

    // Aggregate only from items that have a start (consistent with the displayed list)
    itemsWithStartNonNull.forEach(({ it, start }) => {
      // possible keys: Location, location, Room, room, StartMillis, Start, StartDate
      const loc = it.Location || it.location || '';
      if (loc) locations.add(loc);
      const room = it.Room || it.room || it.Location || '';
      if (room) rooms.add(room);

      // Speakers may be string or array
      const sp = it.Speakers || it.speakers || it.SpeakersTooltip || '';
      if (sp) {
        if (Array.isArray(sp)) sp.forEach(x => speakers.add(String(x)));
        else if (typeof sp === 'string') {
          // try splitting common separators (comma, semicolon or newline)
          sp.split(/[;,\n]/).map(x => x.trim()).filter(Boolean).forEach(x => speakers.add(x));
        }
      }

      if (start != null) startMillis.push(start);
      const dur = it.DurationMinutes != null ? Number(it.DurationMinutes) : parseDurationToMinutes(it.Duration);
      if (start != null && !isNaN(dur)) endMillis.push(start + (dur * 60 * 1000));
    });

    // compute min start and max end
    const start = startMillis.length ? new Date(Math.min(...startMillis)) : null;
    const end = endMillis.length ? new Date(Math.max(...endMillis)) : null;

    const validCount = itemsWithStartNonNull.length;
    const excludedCount = itemsWithStart.length - itemsWithStartNonNull.length;

    return {
      title: s.title || s.name || String(s),
      items: itemsSorted,
      excludedItems,
      locations: Array.from(locations),
      rooms: Array.from(rooms),
      speakers: Array.from(speakers),
      start,
      end,
      validCount,
      excludedCount
    };
  });

  // Formatting helpers
  function fmtDate(d) {
    if (!d) return null;
    try {
      return d.toLocaleString();
    } catch (e) {
      return d.toISOString();
    }
  }

  // Return the list of contributions to display for a session (include excluded when toggled)
  function getDisplayedItems(sa) {
    // always return only the valid items (those with starts); omitted items are shown in the separate collapsed list
    if (!sa) return [];
    return sa.items || [];
  }

  // Helper to render the contributions summary text (total and omitted if any)
  function contributionsSummary(sa) {
    const total = (sa.validCount || 0) + (sa.excludedCount || 0);
    if (sa.excludedCount > 0) return `${total} total, ${sa.excludedCount} omitted`;
    return `${total} total`;
  }

  // Reactive mapping of omitted-toggle labels per session title.
  // Updates automatically when `sessionAggregates` or `expandedOmitted` change.
  $: omittedToggleLabels = (sessionAggregates || []).reduce((m, sa) => {
    const title = sa?.title || String(sa);
    const count = sa?.excludedCount || 0;
    const expanded = !!expandedOmitted[title];
    m[title] = expanded ? `Hide omitted (${count})` : `Show omitted (${count})`;
    return m;
  }, {});
</script>

<Modal bind:open={open} size="lg" dismissable={false} class="session-dialog">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Session Details</h3>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      on:click={closeDialog}
    >
      <CloseOutline class="shrink-0 h-6 w-6" />
    </button>
  </div>

  {#if sessions.length > 0}
    <div class="space-y-4">
      {#each sessionAggregates as sa}
        <div class="p-3 rounded-lg border bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-sm font-semibold text-gray-800 dark:text-gray-200 m-0">
                {sa.title}
              </p>
              {#if sa.start || sa.end}
                <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">
                  {#if sa.start}{fmtDate(sa.start)}{/if}
                  {#if sa.start && sa.end} — {/if}
                  {#if sa.end}{fmtDate(sa.end)}{/if}
                </p>
              {/if}

              <div class="flex items-center gap-3 mt-2">
                {#if sa.locations.length}
                  <p class="text-xs text-gray-600 dark:text-gray-400">Location: {sa.locations.join(', ')}</p>
                {/if}

                {#if sa.rooms.length}
                  <p class="text-xs text-gray-600 dark:text-gray-400">Room(s): {sa.rooms.join(', ')}</p>
                {/if}

                {#if sa.speakers.length}
                  <p class="text-xs text-gray-600 dark:text-gray-400">Speakers: {sa.speakers.join(', ')}</p>
                {/if}
              </div>
            </div>
          </div>

          {#if sa.validCount + sa.excludedCount > 0}
            <div class="mt-3">
              <h5 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Contributions
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">({contributionsSummary(sa)})</span>
              </h5>

              {#if getDisplayedItems(sa).length}
                <ul class="divide-y divide-gray-100 dark:divide-gray-700 max-h-56 overflow-auto">
                  {#each getDisplayedItems(sa) as it}
                    <li class="py-2">
                      <div class="text-sm text-gray-800 dark:text-gray-200">{it.Title || it.title || it.name}</div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">{it.StartDate ? it.StartDate : (it.Start ? it.Start : '')} {#if it.Duration} — {it.Duration}{/if} {#if it.Room} — {it.Room}{/if}</div>
                    </li>
                  {/each}
                </ul>
              {:else}
                <p class="text-xs text-gray-500 dark:text-gray-400">No contributions displayed. Expand omitted to inspect items without start times.</p>
              {/if}
            </div>
          {/if}

          {#if sa.excludedCount > 0}
            <div class="mt-2">
              <button class="text-xs text-blue-600 dark:text-blue-400 hover:underline" type="button" on:click={() => toggleExpanded(sa.title)}>
                {omittedToggleLabels[sa.title] || ''}
              </button>

              {#if expandedOmitted[sa.title]}
                <div class="mt-2 p-2 bg-gray-50 dark:bg-gray-900 rounded">
<!--                  <h6 class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Omitted contributions (no start time)</h6>-->
                  <ul class="text-xs text-gray-700 dark:text-gray-300 space-y-1">
                    {#each sa.excludedItems as oit}
                      <li class="py-1">{oit.Title || oit.title || oit.name} {#if oit.Room} — {oit.Room}{/if}</li>
                    {/each}
                  </ul>
                </div>
              {/if}
            </div>
          {/if}

        </div>
      {/each}
    </div>
  {:else}
    <p class="text-gray-500 dark:text-gray-400">No session information available.</p>
  {/if}

  {#if uniqueSessions.length > 0}
    <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">All Available Sessions</h4>
      <ul class="space-y-2">
        {#each uniqueSessions as s}
          {#if isCurrentSession(s.title)}
            <li class="text-sm pl-2 border-l-2 border-purple-400 text-purple-700 dark:text-purple-300 font-semibold">{s.title}</li>
          {:else}
            <li class="text-sm pl-2 border-l-2 border-blue-400 text-blue-700 dark:text-blue-300">{s.title}</li>
          {/if}
        {/each}
      </ul>
    </div>
  {/if}
</Modal>

<style>
  /* small styles to match other dialogs */
</style>
