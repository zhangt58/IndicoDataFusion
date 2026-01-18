<script>
  import TypeBadge from './TypeBadge.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import AffiliationDialog from '../components/AffiliationDialog.svelte';
  import AffiliationBadge from '../components/AffiliationBadge.svelte';

  let { abstract = {} } = $props();

  // Dialog state
  let showAffiliationDialog = $state(false);
  let selectedAffiliation = $state(null);

  // Handle affiliation click
  function handleAffiliationClick(affiliation) {
    selectedAffiliation = affiliation;
    showAffiliationDialog = true;
  }
</script>

<div
  class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700"
>
  <!-- Title and Status -->
  <div class="flex justify-between items-start mb-3">
    <div class="flex-1">
      <h3 class="text-xl font-bold text-gray-800 dark:text-white">{abstract.title}</h3>
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        ID: {abstract.id}
        {#if abstract.friendly_id}(#{abstract.friendly_id}){/if}
      </p>
    </div>
    <span
      class="ml-4 px-3 py-1 rounded-full text-xs font-semibold
      {abstract.state === 'accepted'
        ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200'
        : abstract.state === 'rejected'
          ? 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200'
          : 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200'}"
    >
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
      </p>
      {#if abstract.submitter.affiliation}
        <AffiliationBadge
          affiliation={abstract.submitter.affiliation}
          onclick={handleAffiliationClick}
          className="text-gray-600 dark:text-gray-400"
        />
      {/if}
      <p class="text-xs text-gray-500">{abstract.submitted_dt}</p>
    </div>
  {/if}

  <!-- Tracks -->
  {#if abstract.reviewed_for_tracks && abstract.reviewed_for_tracks.length > 0}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">
        Reviewed for tracks:
      </p>
      <div class="flex gap-2 flex-wrap">
        {#each abstract.reviewed_for_tracks as track}
          <TrackBadge text={track.title} className="track-reviewed" />
        {/each}
      </div>
    </div>
  {/if}

  {#if abstract.accepted_track}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Accepted track:</p>
      <TrackBadge text={abstract.accepted_track.title} className="track-accepted" />
    </div>
  {/if}

  <!-- Contribution Type -->
  {#if abstract.accepted_contrib_type}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Type:</p>
      <TypeBadge text={abstract.accepted_contrib_type.name} />
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
        {#if abstract.judge.affiliation}
          <AffiliationBadge
            affiliation={abstract.judge.affiliation}
            onclick={handleAffiliationClick}
            className="text-gray-600 dark:text-gray-400"
          />
        {/if}
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
              {person.first_name}
              {person.last_name}
              {#if person.is_speaker}
                <span class="ml-1 text-xs">🎤</span>
              {/if}
            </span>
            {#if person.author_type === 'primary'}
              <span class="ml-1 text-xs text-blue-600 dark:text-blue-300">(Primary)</span>
            {/if}
            {#if person.affiliation}
              <div class="mt-1">
                <AffiliationBadge
                  affiliation={person.affiliation}
                  onclick={handleAffiliationClick}
                  className="text-blue-600 dark:text-blue-400"
                />
              </div>
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

<!-- Affiliation Details Dialog -->
<AffiliationDialog bind:open={showAffiliationDialog} affiliation={selectedAffiliation} />
