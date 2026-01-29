<script>
  import {
    ClipboardListOutline,
    FilterOutline,
    ChevronUpOutline,
  } from 'flowbite-svelte-icons';
  import { GetReviewTracks } from '../../wailsjs/go/main/App';

  let {
    open = $bindable(false),
    onFilterTrack = null,
    buttonElement = null,
  } = $props();

  let loading = $state(false);
  let reviewTracks = $state([]);
  let error = $state(null);
  let selectedTrackName = $state(null);
  let collapsed = $state(false);

  // Dragging state
  let isDragging = $state(false);
  let dragOffsetX = $state(0);
  let dragOffsetY = $state(0);
  let posX = $state(0);
  let posY = $state(0);
  let panelElement = $state(null);

  // Extract track code or short name from full name
  function getTrackCode(trackName) {
    // Try to extract code before colon (e.g., "MC1: Description" -> "MC1")
    const match = trackName.match(/^([^:]+):/);
    if (match) {
      return match[1].trim();
    }
    // Otherwise return the full name
    return trackName;
  }

  // Fetch review tracks
  async function fetchReviewData() {
    loading = true;
    error = null;
    try {
      const tracksData = await GetReviewTracks();
      // Backend already filters tracks with <a> defined
      reviewTracks = tracksData?.tracks || [];
    } catch (err) {
      console.error('Failed to fetch review data:', err);
      error = err;
      reviewTracks = [];
    } finally {
      loading = false;
    }
  }

  // Handle track filter click
  function handleTrackClick(track) {
    selectedTrackName = getTrackCode(track.name);
    if (onFilterTrack) {
      onFilterTrack(selectedTrackName);
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
      posX = rect.right + 50;
      posY = rect.top - 10;
    }
  });
</script>

{#if open}
  <div
    bind:this={panelElement}
    class="fixed bg-sky-200/90 dark:bg-sky-900/90 backdrop-blur-sm shadow-lg rounded-md z-1000 border border-sky-400 dark:border-sky-700 max-w-md {isDragging
      ? 'cursor-move'
      : ''}"
    style="left: {posX}px; top: {posY}px;"
  >
    {#if collapsed}
      <!-- Collapsed: show just icon -->
      <button
        onclick={toggleCollapse}
        class="p-2 hover:bg-sky-300 dark:hover:bg-sky-800 rounded-md transition-colors"
        title="Expand My Reviews"
      >
        <ClipboardListOutline class="w-5 h-5 text-gray-800 dark:text-gray-100" />
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
            <ClipboardListOutline class="w-5 h-5" />
            My Review Tracks
          </h3>
          <div class="flex items-center gap-1">
            <button
              onclick={toggleCollapse}
              class="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1"
              title="Collapse"
            >
              <ChevronUpOutline class="w-4 h-4" />
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
            class="px-3 py-1.5 bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white rounded text-sm font-medium transition-colors flex items-center gap-2"
          >
            <FilterOutline class="w-4 h-4" />
            {loading ? 'Fetching...' : 'Fetch My Reviews'}
          </button>
        </div>

        {#if error}
          <div class="text-red-600 dark:text-red-400 text-xs mb-2">
            Error: {error.message || String(error)}
          </div>
        {/if}

        {#if reviewTracks.length > 0}
          <div class="text-xs text-gray-600 dark:text-gray-400 mb-1">
            {reviewTracks.length} track(s) assigned:
          </div>
          <div class="flex flex-wrap gap-1">
            {#each reviewTracks as track}
              {@const trackCode = getTrackCode(track.name)}
              <button
                onclick={() => handleTrackClick(track)}
                class="px-2 py-1 text-xs rounded transition-colors {selectedTrackName === trackCode
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600'}"
                title={track.name}
              >
                {trackCode}
              </button>
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
