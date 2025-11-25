<script>
  import { Modal } from 'flowbite-svelte';
  import { X } from '@lucide/svelte';
  import { getShortTrackName } from './AbstractTableItem.js';

  /** @type {boolean} */
  export let open = false;

  /** @type {Array<{title: string, type: string}>} */
  export let tracks = [];

  /** @type {Array<{title: string, type: string}>} */
  export let allTracks = [];

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
  $: uniqueTracks = allTracks
    .filter((track, index, self) => 
      index === self.findIndex(t => t.title === track.title)
    )
    .sort(sortTracks);

  // Check if a track is one of the current abstract's tracks
  function isCurrentTrack(trackTitle) {
    return tracks.some(t => t.title === trackTitle);
  }

  // Get the type of the current track (for highlighting color)
  function getCurrentTrackType(trackTitle) {
    const found = tracks.find(t => t.title === trackTitle);
    return found?.type || null;
  }
</script>

<Modal bind:open={open} size="md" dismissable={false} class="track-dialog">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Track Details</h3>
    <button 
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      on:click={closeDialog}
    >
      <X size={20} />
    </button>
  </div>
  
  {#if tracks.length > 0}
    <div class="space-y-3">
      {#each tracks as track}
        <div class="p-3 rounded-lg border {track.type === 'accepted' ? 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-800' : 'bg-purple-50 dark:bg-purple-900/30 border-purple-200 dark:border-purple-800'}">
          <div class="flex items-center gap-2">
            <span class="px-2 py-0.5 text-xs font-medium rounded {track.type === 'accepted' ? 'bg-green-100 dark:bg-green-800 text-green-800 dark:text-green-200' : 'bg-purple-100 dark:bg-purple-800 text-purple-800 dark:text-purple-200'}">
              {track.type === 'accepted' ? 'Accepted' : 'Reviewed'}
            </span>
          </div>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">{track.title}</p>
        </div>
      {/each}
    </div>
  {:else}
    <p class="text-gray-500 dark:text-gray-400">No track information available.</p>
  {/if}

  <!-- All Tracks List Section -->
  {#if allTracks.length > 0}
    <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">All Available Tracks</h4>
      <ul class="space-y-2">
        {#each uniqueTracks as track}
          {@const isCurrent = isCurrentTrack(track.title)}
          {@const currentType = getCurrentTrackType(track.title)}
          <li class="text-sm pl-2 border-l-2 {isCurrent 
            ? (currentType === 'accepted' 
              ? 'border-green-400 text-green-700 dark:text-green-300 font-semibold' 
              : 'border-purple-400 text-purple-700 dark:text-purple-300 font-semibold')
            : 'border-blue-400 text-blue-700 dark:text-blue-300'}">
            {track.title}
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</Modal>
