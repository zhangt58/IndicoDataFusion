<script>
  import { onMount } from 'svelte';
  import { GetAbstractData } from '../../wailsjs/go/backend/IndicoClient';

  let loading = false;
  let abstractData = [];
  let error = null;

  onMount(async () => {
    loading = true;
    try {
      abstractData = (await GetAbstractData()) || [];
    } catch (e) {
      console.error('GetAbstractData failed', e);
      abstractData = [];
      error = e;
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="p-6 text-center">Loading abstracts...</div>
{:else if error}
  <div class="p-6 text-center text-red-600">Failed to load abstracts.</div>
{:else}
  <section class="space-y-4">
    <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-200 mb-4">Abstracts ({abstractData.length})</h2>
    {#each abstractData as abstract}
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
        <!-- Title and Status -->
        <div class="flex justify-between items-start mb-3">
          <div class="flex-1">
            <h3 class="text-xl font-bold text-gray-800 dark:text-white">{abstract.title}</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              ID: {abstract.id} {#if abstract.friendly_id}(#{abstract.friendly_id}){/if}
            </p>
          </div>
          <span class="ml-4 px-3 py-1 rounded-full text-xs font-semibold
            {abstract.state === 'accepted' ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200' :
             abstract.state === 'rejected' ? 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200' :
             'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200'}">
            {abstract.state}
          </span>
        </div>

        <!-- Content -->
        {#if abstract.content}
          <div class="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded">
            <p class="text-sm text-gray-700 dark:text-gray-300">{abstract.content}</p>
          </div>
        {/if}

        <!-- Submitter -->
        {#if abstract.submitter}
          <div class="mb-3">
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Submitted by:</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {abstract.submitter.full_name}
              {#if abstract.submitter.affiliation}
                <span class="text-xs text-gray-500">({abstract.submitter.affiliation})</span>
              {/if}
            </p>
            <p class="text-xs text-gray-500">{abstract.submitted_dt}</p>
          </div>
        {/if}

        <!-- Tracks -->
        {#if abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0}
          <div class="mb-3">
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Reviewed for tracks:</p>
            <div class="flex gap-2 flex-wrap">
              {#each abstract.reviewed_for_tracks as track}
                <span class="px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded text-xs">
                  {track.title}
                </span>
              {/each}
            </div>
          </div>
        {/if}

        {#if abstract.accepted_track}
          <div class="mb-3">
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Accepted track:</p>
            <span class="px-2 py-1 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded text-xs">
              {abstract.accepted_track.title}
            </span>
          </div>
        {/if}

        <!-- Contribution Type -->
        {#if abstract.accepted_contrib_type}
          <div class="mb-3">
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Type:</p>
            <span class="text-sm text-gray-700 dark:text-gray-300">{abstract.accepted_contrib_type.name}</span>
          </div>
        {/if}

        <!-- Score and Judge -->
        <div class="flex gap-4 mb-3">
          {#if abstract.score !== null && abstract.score !== undefined}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Score:</p>
              <p class="text-sm text-gray-700 dark:text-gray-300">{abstract.score}</p>
            </div>
          {/if}

          {#if abstract.judge}
            <div>
              <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Judge:</p>
              <p class="text-sm text-gray-700 dark:text-gray-300">
                {abstract.judge.full_name}
              </p>
            </div>
          {/if}
        </div>

        <!-- Authors/Persons -->
        {#if abstract.persons && abstract.persons.length > 0}
          <div class="mt-3">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Authors:</p>
            <div class="flex gap-2 flex-wrap">
              {#each abstract.persons as person}
                <div class="px-3 py-1 bg-blue-50 dark:bg-blue-900 rounded text-sm">
                  <span class="font-medium text-blue-800 dark:text-blue-200">
                    {person.first_name} {person.last_name}
                    {#if person.is_speaker}
                      <span class="ml-1 text-xs">🎤</span>
                    {/if}
                  </span>
                  {#if person.author_type === 'primary'}
                    <span class="ml-1 text-xs text-blue-600 dark:text-blue-300">(Primary)</span>
                  {/if}
                  {#if person.affiliation}
                    <div class="text-xs text-blue-600 dark:text-blue-400">{person.affiliation}</div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Metadata -->
        <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600 text-xs text-gray-500">
          <div class="flex gap-4">
            <span>Modified: {abstract.modified_dt}</span>
            {#if abstract.custom_fields && abstract.custom_fields.length > 0}
              <span>Custom Fields: {abstract.custom_fields.length}</span>
            {/if}
            {#if abstract.reviews && abstract.reviews.length > 0}
              <span>Reviews: {abstract.reviews.length}</span>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </section>
{/if}
