<script>
  import Icon from '@iconify/svelte';
  import { GetReviewTracks } from '../../wailsjs/go/main/App';
  import { voteStatsStore } from '../lib/voteStats.svelte.js';

  let {
    open = $bindable(false),
    onFilterTrackByID = null,
    // Reference to the button element that opened the panel
    buttonElement = null,
  } = $props();

  let loading = $state(false);
  let reviewTracks = $state([]);
  let error = $state(null);
  // store the selected review track id
  let selectedTrackID = $state(null);
  let collapsed = $state(false);

  // Dragging state
  let isDragging = $state(false);
  let dragOffsetX = $state(0);
  let dragOffsetY = $state(0);
  let posX = $state(0);
  let posY = $state(0);
  let panelElement = $state(null);

  // count tracks with links
  let validTrackCount = $derived.by(() => {
    let total = 0;
    for (const track of reviewTracks) {
      if (track.link) {
        total += 1;
      }
    }
    return total;
  });

  // Fetch review tracks and vote stats via shared store
  async function fetchReviewData() {
    loading = true;
    error = null;
    try {
      const [tracksData] = await Promise.all([GetReviewTracks(), voteStatsStore.refresh()]);
      // Backend already filters tracks with <a> defined
      reviewTracks = tracksData?.tracks || [];
    } catch (err) {
      error = err;
      reviewTracks = [];
    } finally {
      loading = false;
    }
  }

  /** Returns the TrackVoteStats for a given track_id, or null. */
  function trackVotes(trackID) {
    if (!voteStatsStore.data?.per_track) return null;
    return voteStatsStore.data.per_track[trackID] ?? null;
  }

  // Handle track filter click - toggle track selection
  function handleTrackClick(track) {
    const trackId = track.track_id ?? null;

    // Toggle: if already selected, deselect (set to null), otherwise select
    if (selectedTrackID === trackId) {
      selectedTrackID = null;
    } else {
      selectedTrackID = trackId;
    }

    // Always call the callback to update the parent (AbstractPage)
    if (onFilterTrackByID) {
      onFilterTrackByID(selectedTrackID);
    }
  }

  // Mouse drag handlers
  function handleMouseDown(e) {
    if (collapsed || !panelElement) return;
    // Only allow dragging from header area
    const target = e.target;
    if (!target.closest('.drag-handle')) return;

    isDragging = true;
    dragOffsetX = e.clientX - posX;
    dragOffsetY = e.clientY - posY;
    e.preventDefault();
  }

  function handleMouseMove(e) {
    if (!isDragging) return;
    posX = e.clientX - dragOffsetX;
    posY = e.clientY - dragOffsetY;
  }

  function handleMouseUp() {
    isDragging = false;
  }

  // Toggle collapse
  function toggleCollapse() {
    collapsed = !collapsed;
  }

  // Close panel
  function close() {
    open = false;
    collapsed = false;
  }

  $effect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
    } else {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    }
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  });

  // Position the panel to the right of the button when opened
  $effect(() => {
    if (open && buttonElement) {
      const rect = buttonElement.getBoundingClientRect();
      posX = rect.right + 20;
      posY = rect.top - 2;
    }
  });
</script>

{#if open}
  <div
    bind:this={panelElement}
    class="fixed bg-sky-200/20 dark:bg-sky-900/20 backdrop-blur-xs shadow-md rounded-md z-1000 border border-sky-400 dark:border-sky-700 max-w-md"
    class:cursor-move={isDragging}
    style="left: {posX}px; top: {posY}px;"
  >
    {#if collapsed}
      <!-- Collapsed: show just icon -->
      <button
        onclick={toggleCollapse}
        class="p-2 hover:bg-sky-300 dark:hover:bg-sky-800 rounded-md transition-colors"
        title="Expand My Reviews"
      >
        <Icon icon="mdi:clipboard-list" class="w-5 h-5 text-gray-800 dark:text-gray-100" />
      </button>
    {:else}
      <!-- Expanded view -->
      <div class="px-2 py-2">
        <div
          class="flex items-center justify-between mb-2 drag-handle cursor-move"
          onmousedown={handleMouseDown}
          role="button"
          tabindex="0"
        >
          <h3
            class="text-sm font-semibold text-gray-800 dark:text-gray-100 flex items-center gap-2"
          >
            <Icon icon="mdi:clipboard-list" class="w-5 h-5" />
            My Review Tracks
          </h3>
          <div class="flex items-center gap-1">
            <button
              onclick={toggleCollapse}
              class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1"
              title="Collapse"
            >
              <Icon icon="mdi:chevron-up" class="w-4 h-4" />
            </button>
            <button
              onclick={close}
              class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100"
              title="Close"
            >
              ✕
            </button>
          </div>
        </div>

        <div class="mb-2">
          <button
            onclick={fetchReviewData}
            disabled={loading}
            class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white rounded text-xs font-medium transition-colors flex items-center gap-1"
          >
            <Icon icon="mdi:filter" class="w-4 h-4" />
            {loading ? 'Fetching...' : 'Fetch My Reviews'}
          </button>
        </div>

        {#if error}
          <div class="text-red-600 dark:text-red-400 text-xs mb-2">
            Error: {error.message || String(error)}
          </div>
        {/if}

        {#if reviewTracks.length > 0}
          <div class="font-medium text-xs text-gray-600 dark:text-gray-400 mb-1">
            {validTrackCount} track(s) assigned:
          </div>

          <!-- Scrollable area covering vote stats + track list -->
          <div class="overflow-y-auto max-h-96 pr-1 space-y-1">
            <!-- Vote summary bar (shown when voteStatsStore.data is available) -->
            {#if voteStatsStore.data}
              <div
                class="px-2 py-1 rounded bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-700"
              >
                <div class="flex items-center gap-1.5 mb-1">
                  <Icon icon="mdi:vote" class="w-3.5 h-3.5 text-indigo-600 dark:text-indigo-400" />
                  <span class="text-xs font-semibold text-indigo-700 dark:text-indigo-300">
                    Votes: {voteStatsStore.data.total_cast} cast · {voteStatsStore.data.max_votes} max
                    per track
                  </span>
                </div>
                <div class="flex flex-col gap-0.5">
                  {#each reviewTracks.filter((t) => t.link) as track}
                    {@const tv = trackVotes(track.track_id)}
                    {#if tv}
                      {@const pct = tv.votes_max > 0 ? (tv.votes_cast / tv.votes_max) * 100 : 0}
                      {@const overLimit = tv.votes_cast >= tv.votes_max}
                      <div class="flex items-center gap-1.5">
                        <span
                          class="text-[0.65rem] text-gray-600 dark:text-gray-400 truncate max-w-28"
                          title={track.name}>{track.name}</span
                        >
                        <div
                          class="flex-1 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden"
                        >
                          <div
                            class="h-full rounded-full transition-all {overLimit
                              ? 'bg-red-500'
                              : pct >= 67
                                ? 'bg-amber-400'
                                : 'bg-green-500'}"
                            style="width: {Math.min(pct, 100)}%"
                          ></div>
                        </div>
                        <span
                          class="text-[0.65rem] font-mono font-semibold shrink-0 {overLimit
                            ? 'text-red-600 dark:text-red-400'
                            : 'text-gray-600 dark:text-gray-400'}"
                        >
                          {tv.votes_cast}/{tv.votes_max}
                        </span>
                      </div>
                    {/if}
                  {/each}
                </div>
              </div>
            {/if}

            <div class="flex flex-col gap-1">
              {#each reviewTracks as track}
                {#if track.link}
                  {@const tv = trackVotes(track.track_id)}
                  {@const overLimit = tv && tv.votes_cast >= tv.votes_max}
                  <button
                    onclick={() => handleTrackClick(track)}
                    class="relative px-5 py-1 text-xs rounded transition-colors {selectedTrackID ===
                    track.track_id
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600'}"
                    title="{track.name}{tv ? ` — ${tv.votes_cast}/${tv.votes_max} votes cast` : ''}"
                  >
                    {#if track.abstract_count > 0}
                      <span
                        class="absolute -top-1 -left-0.5 flex items-center justify-center min-w-4 h-4 px-1 bg-red-500 text-white text-[0.55rem] rounded-md border border-white dark:border-gray-800"
                      >
                        {track.abstract_count}
                      </span>
                    {/if}
                    {track.name}
                    {#if tv}
                      <span
                        class="ml-2 inline-flex items-center gap-0.5 text-[0.6rem] font-semibold px-1 rounded {overLimit
                          ? 'bg-red-200 text-red-700 dark:bg-red-800/50 dark:text-red-300'
                          : 'bg-green-100 text-green-700 dark:bg-green-800/50 dark:text-green-300'}"
                        title="Votes cast / max per track"
                      >
                        <Icon icon="mdi:vote" class="w-2.5 h-2.5" />
                        {tv.votes_cast}/{tv.votes_max}
                      </span>
                    {/if}
                  </button>
                {:else}
                  <span
                    class="relative px-2 py-1 text-xs rounded bg-gray-300 dark:bg-gray-600 text-gray-500 dark:text-gray-400 cursor-not-allowed"
                    title={track.name}
                  >
                    {track.name}
                  </span>
                {/if}
              {/each}
            </div>
            <!-- end scrollable area -->
          </div>
        {:else if !loading}
          <div class="text-xs text-gray-500 dark:text-gray-400">
            Click "Fetch My Reviews" to load your assigned review tracks.
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
