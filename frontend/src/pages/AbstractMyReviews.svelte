<script>
  import Icon from '@iconify/svelte';
  import { GetReviewTracks } from '../../wailsjs/go/main/App';

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

  // Fetch review tracks
  async function fetchReviewData() {
    loading = true;
    error = null;
    try {
      const tracksData = await GetReviewTracks();
      // Backend already filters tracks with <a> defined
      reviewTracks = tracksData?.tracks || [];
    } catch (err) {
      error = err;
      reviewTracks = [];
    } finally {
      loading = false;
    }
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
      <div class="px-3 py-2">
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
          <div class="flex flex-wrap gap-2">
            {#each reviewTracks as track}
              {#if track.link}
                <button
                  onclick={() => handleTrackClick(track)}
                  class="relative px-2 py-1 text-xs rounded transition-colors {selectedTrackID ===
                  track.track_id
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600'}"
                  title={track.name}
                >
                  {#if track.abstract_count > 0}
                    <span
                      class="absolute -top-2 -left-2.5 flex items-center justify-center min-w-4 h-4 px-1 bg-red-500 text-white text-[0.65rem] font-medium rounded-md border border-white dark:border-gray-800"
                    >
                      {track.abstract_count}
                    </span>
                  {/if}
                  {track.name}
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
        {:else if !loading}
          <div class="text-xs text-gray-500 dark:text-gray-400">
            Click "Fetch My Reviews" to load your assigned review tracks.
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
