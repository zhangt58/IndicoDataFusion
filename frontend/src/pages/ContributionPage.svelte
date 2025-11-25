<script>
  import { onMount } from 'svelte';
  import { GetContributionData } from '../../wailsjs/go/backend/IndicoClient';

  let loading = false;
  let contributionData = [];
  let error = null;

  onMount(async () => {
    loading = true;
    try {
      contributionData = (await GetContributionData()) || [];
    } catch (e) {
      console.error('GetContributionData failed', e);
      contributionData = [];
      error = e;
    } finally {
      loading = false;
    }
  });

  function formatDate(dateInfo) {
    if (!dateInfo) return '';
    const { date, time, tz } = dateInfo;
    return tz ? `${date} ${time} (${tz})` : `${date} ${time}`;
  }
</script>

{#if loading}
  <div class="p-6 text-center">Loading contributions...</div>
{:else if error}
  <div class="p-6 text-center text-red-600">Failed to load contributions: {error}</div>
{:else}
  <h2 class="fixed bg-indigo-300 top-2 left-2 shadow-md px-2 py-1 rounded-sm text-xl font-semibold text-gray-700 dark:text-gray-200">Contributions ({contributionData.length})</h2>
  <section class="space-y-4 mt-8">
    {#each contributionData as contribution}
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
        <!-- Title and Code -->
        <div class="flex justify-between items-start mb-3">
          <div class="flex-1">
            <h3 class="text-xl font-bold text-gray-800 dark:text-white">{contribution.title}</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              ID: {contribution.id} {#if contribution.friendly_id}(#{contribution.friendly_id}){/if}
              {#if contribution.code}
                <span class="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded">Code: {contribution.code}</span>
              {/if}
            </p>
          </div>
          {#if contribution.type}
            <span class="ml-4 px-3 py-1 rounded-full text-xs font-semibold bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200">
              {contribution.type}
            </span>
          {/if}
        </div>

        <!-- Description -->
        {#if contribution.description}
          <div class="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded">
            <p class="text-sm text-gray-700 dark:text-gray-300">{contribution.description}</p>
          </div>
        {/if}

        <!-- Session and Track -->
        <div class="mb-3 flex gap-4 flex-wrap">
          {#if contribution.session}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Session:</p>
              <span class="text-sm px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded">
                {contribution.session}
              </span>
            </div>
          {/if}

          {#if contribution.track}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Track:</p>
              <span class="text-sm px-2 py-1 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded">
                {contribution.track}
              </span>
            </div>
          {/if}
        </div>

        <!-- Date, Time, and Location -->
        <div class="mb-3 grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
          {#if contribution.startDate}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Start:</p>
              <p class="text-gray-700 dark:text-gray-300">{formatDate(contribution.startDate)}</p>
            </div>
          {/if}

          {#if contribution.endDate}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">End:</p>
              <p class="text-gray-700 dark:text-gray-300">{formatDate(contribution.endDate)}</p>
            </div>
          {/if}

          {#if contribution.duration}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Duration:</p>
              <p class="text-gray-700 dark:text-gray-300">{contribution.duration} minutes</p>
            </div>
          {/if}

          {#if contribution.location}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Location:</p>
              <p class="text-gray-700 dark:text-gray-300">{contribution.location}</p>
            </div>
          {/if}

          {#if contribution.room || contribution.roomFullname}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Room:</p>
              <p class="text-gray-700 dark:text-gray-300">{contribution.roomFullname || contribution.room}</p>
            </div>
          {/if}

          {#if contribution.board_number}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Board Number:</p>
              <p class="text-gray-700 dark:text-gray-300">{contribution.board_number}</p>
            </div>
          {/if}
        </div>

        <!-- Speakers -->
        {#if contribution.speakers && contribution.speakers.length > 0}
          <div class="mb-3">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Speakers:</p>
            <div class="flex gap-2 flex-wrap">
              {#each contribution.speakers as speaker}
                <div class="px-3 py-1 bg-blue-50 dark:bg-blue-900 rounded text-sm">
                  <span class="font-medium text-blue-800 dark:text-blue-200">
                    🎤 {speaker.fullName || `${speaker.first_name} ${speaker.last_name}`}
                  </span>
                  {#if speaker.affiliation}
                    <div class="text-xs text-blue-600 dark:text-blue-400">{speaker.affiliation}</div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Primary Authors -->
        {#if contribution.primaryauthors && contribution.primaryauthors.length > 0}
          <div class="mb-3">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Primary Authors:</p>
            <div class="flex gap-2 flex-wrap">
              {#each contribution.primaryauthors as author}
                <div class="px-3 py-1 bg-amber-50 dark:bg-amber-900 rounded text-sm">
                  <span class="font-medium text-amber-800 dark:text-amber-200">
                    {author.fullName || `${author.first_name} ${author.last_name}`}
                  </span>
                  {#if author.affiliation}
                    <div class="text-xs text-amber-600 dark:text-amber-400">{author.affiliation}</div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Co-Authors -->
        {#if contribution.coauthors && contribution.coauthors.length > 0}
          <div class="mb-3">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Co-Authors ({contribution.coauthors.length}):</p>
            <div class="flex gap-2 flex-wrap">
              {#each contribution.coauthors as coauthor}
                <div class="px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">
                  <span class="text-gray-800 dark:text-gray-200">
                    {coauthor.fullName || `${coauthor.first_name} ${coauthor.last_name}`}
                  </span>
                  {#if coauthor.affiliation}
                    <div class="text-xs text-gray-600 dark:text-gray-400">{coauthor.affiliation}</div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Keywords and References -->
        {#if (contribution.keywords && contribution.keywords.length > 0) || (contribution.references && contribution.references.length > 0)}
          <div class="mb-3 flex gap-4">
            {#if contribution.keywords && contribution.keywords.length > 0}
              <div>
                <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Keywords:</p>
                <div class="flex gap-1 flex-wrap">
                  {#each contribution.keywords as keyword}
                    <span class="px-2 py-0.5 bg-teal-100 dark:bg-teal-900 text-teal-800 dark:text-teal-200 rounded text-xs">
                      {keyword}
                    </span>
                  {/each}
                </div>
              </div>
            {/if}

            {#if contribution.references && contribution.references.length > 0}
              <div>
                <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">References: {contribution.references.length}</p>
              </div>
            {/if}
          </div>
        {/if}

        <!-- Materials and Folders -->
        {#if (contribution.material && contribution.material.length > 0) || (contribution.folders && contribution.folders.length > 0)}
          <div class="mb-3">
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">
              {#if contribution.material && contribution.material.length > 0}
                Materials: {contribution.material.length}
              {/if}
              {#if contribution.folders && contribution.folders.length > 0}
                {#if contribution.material && contribution.material.length > 0} | {/if}
                Folders: {contribution.folders.length}
              {/if}
            </p>
          </div>
        {/if}

        <!-- Link to Indico -->
        {#if contribution.url}
          <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
            <a href={contribution.url} target="_blank" rel="noopener noreferrer" 
               class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
              View on Indico →
            </a>
          </div>
        {/if}
      </div>
    {/each}
  </section>
{/if}
