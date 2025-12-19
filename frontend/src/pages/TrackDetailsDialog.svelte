<script>
  import { Modal } from 'flowbite-svelte';
  import { CloseOutline } from 'flowbite-svelte-icons';
  import { getShortTrackName } from './AbstractTableItem.js';

  let { open = $bindable(false), tracks = [], allTracks = [], showTypes = true } = $props();

  // Close dialog
  function closeDialog() {
    open = false;
  }

  // Extract number from MC track name (e.g., "MC1" -> 1, "MC10" -> 10)
  function extractMCNumber(shortName) {
    const match = shortName.match(/^MC(\d+)/i);
    return match ? parseInt(match[1], 10) : null;
  }

  // Sort function: MCs by number, others alphabetically
  function sortTracks(a, b) {
    const shortA = getShortTrackName(a.title);
    const shortB = getShortTrackName(b.title);

    const numA = extractMCNumber(shortA);
    const numB = extractMCNumber(shortB);

    // Both are MCs - sort by number
    if (numA !== null && numB !== null) {
      return numA - numB;
    }
    // Only A is MC - MCs come first
    if (numA !== null) return -1;
    // Only B is MC - MCs come first
    if (numB !== null) return 1;
    // Neither is MC - sort alphabetically
    return shortA.localeCompare(shortB);
  }

  // Get unique full track names for the list, sorted
  let uniqueTracks = $derived(
    allTracks
      .filter((track, index, self) => index === self.findIndex((t) => t.title === track.title))
      .sort(sortTracks),
  );

  // Check if a track is one of the current abstract's tracks
  function isCurrentTrack(trackTitle) {
    return tracks.some((t) => t.title === trackTitle);
  }

  // Get the type of the current track (for highlighting color)
  function getCurrentTrackType(trackTitle) {
    const found = tracks.find((t) => t.title === trackTitle);
    if (!found) return null;
    // only treat explicit 'accepted'/'reviewed' as meaningful types; otherwise return null
    return found.type === 'accepted' || found.type === 'reviewed' ? found.type : null;
  }
</script>

<Modal bind:open size="md" dismissable={false} class="track-dialog">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Track Details</h3>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <CloseOutline class="shrink-0 h-6 w-6" />
    </button>
  </div>

  {#if tracks.length > 0}
    <div class="space-y-3">
      {#each tracks as track}
        {#if track}
          <div
            class="p-3 rounded-lg border bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center gap-2">
              {#if showTypes && (track.type === 'accepted' || track.type === 'reviewed')}
                <span
                  class="px-2 py-0.5 text-xs font-medium rounded {track.type === 'accepted'
                    ? 'bg-green-100 dark:bg-green-800 text-green-800 dark:text-green-200'
                    : 'bg-purple-100 dark:bg-purple-800 text-purple-800 dark:text-purple-200'}"
                >
                  {track.type === 'accepted' ? 'Accepted' : 'Reviewed'}
                </span>
              {/if}
              <p class="text-sm text-gray-700 dark:text-gray-300 m-0">{track.title}</p>
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {:else}
    <p class="text-gray-500 dark:text-gray-400">No track information available.</p>
  {/if}

  <!-- All Tracks List Section -->
  {#if allTracks.length > 0}
    <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
        All Available Tracks
      </h4>
      <ul class="space-y-2">
        {#each uniqueTracks as track}
          {@const isCurrent = isCurrentTrack(track.title)}
          {@const currentType = showTypes ? getCurrentTrackType(track.title) : null}
          {#if isCurrent}
            {#if currentType === 'accepted'}
              <li
                class="text-sm pl-2 border-l-2 border-green-400 text-green-700 dark:text-green-300 font-semibold"
              >
                {track.title}
              </li>
            {:else if currentType === 'reviewed'}
              <li
                class="text-sm pl-2 border-l-2 border-purple-400 text-purple-700 dark:text-purple-300 font-semibold"
              >
                {track.title}
              </li>
            {:else}
              <!-- current track but no meaningful type: neutral highlight -->
              <li
                class="text-sm pl-2 border-l-2 border-gray-400 text-gray-700 dark:text-gray-300 font-semibold"
              >
                {track.title}
              </li>
            {/if}
          {:else}
            <li class="text-sm pl-2 border-l-2 border-blue-400 text-blue-700 dark:text-blue-300">
              {track.title}
            </li>
          {/if}
        {/each}
      </ul>
    </div>
  {/if}
</Modal>
