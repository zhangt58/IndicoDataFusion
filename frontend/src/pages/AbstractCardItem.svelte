<script>
  import { RefreshAbstractByID } from '../../wailsjs/go/main/App';
  import { ClipboardListOutline } from 'flowbite-svelte-icons';
  import { OpenSafeURL } from '../../wailsjs/go/main/App';
  import Icon from '@iconify/svelte';
  import TypeBadge from './TypeBadge.svelte';
  import TrackBadge from './TrackBadge.svelte';
  import StateBadge from './StateBadge.svelte';
  import AffiliationDialog from '../components/AffiliationDialog.svelte';
  import AffiliationBadge from '../components/AffiliationBadge.svelte';
  import AbstractReviewsDialog from '../components/AbstractReviewsDialog.svelte';

  let {
    abstract = $bindable({}),
    onRefresh = null,
    isMyReview = false
  } = $props();

  // Dialog state
  let showAffiliationDialog = $state(false);
  let selectedAffiliation = $state(null);
  let showReviewsDialog = $state(false);
  // Add state to control the raw JSON collapsible
  let showRawJson = $state(false);

  // copied-state for copy button
  let showCopied = $state(false);

  // refresh state
  let isRefreshing = $state(false);

  // Handle affiliation click
  function handleAffiliationClick(affiliation) {
    selectedAffiliation = affiliation;
    showAffiliationDialog = true;
  }

  // Handle reviews click
  function handleReviewsClick() {
    showReviewsDialog = true;
  }

  // Handle refresh
  async function handleRefresh() {
    if (isRefreshing || !abstract.id) return;

    isRefreshing = true;

    try {
      const refreshed = await RefreshAbstractByID(abstract.id);
      // Update the bindable prop directly - this will propagate up through AbstractCardView to AbstractPage
      abstract = refreshed;
      // Also notify parent via callback if provided (for backwards compatibility)
      if (typeof onRefresh === 'function') {
        onRefresh(refreshed);
      }
    } catch (err) {
      alert('Failed to refresh abstract: ' + (err && err.message ? err.message : String(err)));
    } finally {
      isRefreshing = false;
    }
  }

  // Use precomputed aggregated ratings from backend
  const firstPriorityTotal = $derived(abstract.first_priority || 0);
  const secondPriorityTotal = $derived(abstract.second_priority || 0);

  async function copyRawJson() {
    // prevent details toggle (additional precaution)
    try {
      const text = JSON.stringify(abstract, null, 2);

      if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        showCopied = true;
        setTimeout(() => (showCopied = false), 2000);
      }
    } catch (err) {
      // best-effort: inform user
      alert('Copy failed: ' + (err && err.message ? err.message : String(err)));
    }
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
    <div class="ml-4 flex items-center gap-2">
      {#if isMyReview}
        <span
          class="px-2 py-1 text-xs rounded bg-purple-100 dark:bg-purple-900 text-purple-700 dark:text-purple-200 font-semibold flex items-center gap-1"
          title="This abstract is on your review track"
        >
          <ClipboardListOutline class="w-3 h-3" />
          My Review
          <a
            href={abstract.review_url}
            onclick={() => OpenSafeURL(abstract.review_url)}
            target="_blank"
            rel="noopener noreferrer"
            class="ml-1 inline-flex items-center"
            title="Open review page"
            aria-label="Open review page"
          >
            <Icon icon="mdi:open-in-new" class="w-3 h-3 text-blue-600 dark:text-blue-300" aria-hidden="true" />
          </a>
        </span>
      {/if}
      <button
        type="button"
        onclick={handleRefresh}
        disabled={isRefreshing}
        class="px-2 py-1 text-xs rounded transition-colors duration-150
          {isRefreshing
          ? 'bg-gray-300 dark:bg-gray-600 text-gray-500 dark:text-gray-400 cursor-not-allowed'
          : 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-200 hover:bg-blue-200 dark:hover:bg-blue-800'}"
        title="Refresh abstract data"
      >
        {isRefreshing ? '↻ Refreshing...' : '↻ Refresh'}
      </button>
      <StateBadge state={abstract.state} className="font-semibold" />
    </div>
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
          <TrackBadge text={track.title ?? track.code} type="reviewed" />
        {/each}
      </div>
    </div>
  {/if}

  {#if abstract.submitted_for_tracks && abstract.submitted_for_tracks.length > 0}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">
        Submitted for tracks:
      </p>
      <div class="flex gap-2 flex-wrap">
        {#each abstract.submitted_for_tracks as track}
          <TrackBadge text={track.title ?? track.code} type="reviewed" />
        {/each}
      </div>
    </div>
  {/if}

  {#if abstract.accepted_track}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Accepted track:</p>
      <TrackBadge text={abstract.accepted_track.title ?? abstract.accepted_track.code} type="accepted" />
    </div>
  {/if}

  <!-- Contribution Type -->
  {#if abstract.accepted_contrib_type}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Type:</p>
      <TypeBadge text={abstract.accepted_contrib_type.name} />
    </div>
  {/if}

  {#if abstract.submitted_contrib_type}
    <div class="mb-3">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Submitted type:</p>
      <TypeBadge text={abstract.submitted_contrib_type.name} />
    </div>
  {/if}

  <!-- Score and Judge -->
  <div class="flex gap-4">
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

  <!-- Ratings Summary -->
  {#if abstract.reviews && abstract.reviews.length > 0 && (firstPriorityTotal > 0 || secondPriorityTotal > 0)}
    <div class="mt-1">
      <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Ratings:</p>
      <div class="text-sm flex gap-4 items-center">
        {#if firstPriorityTotal > 0}
          <div class="flex items-center gap-2">
            <span class="text-gray-600 dark:text-gray-400">First Priority:</span>
            <span class="font-semibold text-blue-600 dark:text-blue-400">
              {firstPriorityTotal}
            </span>
          </div>
        {/if}
        {#if secondPriorityTotal > 0}
          <div class="flex items-center gap-2">
            <span class="text-gray-600 dark:text-gray-400">Second Priority:</span>
            <span class="font-semibold text-green-600 dark:text-green-400">
              {secondPriorityTotal}
            </span>
          </div>
        {/if}
      </div>
    </div>
  {/if}

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
    <div class="flex gap-4 items-center">
      <span>Modified: {abstract.modified_dt}</span>
      {#if abstract.custom_fields && abstract.custom_fields.length > 0}
        <span>Custom Fields: {abstract.custom_fields.length}</span>
      {/if}
      {#if abstract.reviews && abstract.reviews.length > 0}
        <button
          type="button"
          class="text-blue-600 dark:text-blue-400 hover:underline cursor-pointer font-semibold"
          onclick={handleReviewsClick}
        >
          Reviews: {abstract.reviews.length}
        </button>
      {/if}
    </div>
  </div>

  <!-- Raw JSON (collapsible) -->
  <div class="mt-2">
    <details class="bg-gray-50 dark:bg-gray-700 rounded p-1" bind:open={showRawJson}>
      <summary
        class="cursor-pointer text-sm font-semibold text-gray-700 dark:text-gray-300 flex justify-between items-center"
      >
        <span>Raw abstract JSON</span>
        <span class="inline-flex items-center gap-1">
          <span class="relative inline-flex items-center">
            <button
              onclick={copyRawJson}
              class="text-xs px-1 py-1 rounded bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 transition-colors duration-150 hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-gray-900 dark:hover:text-white focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-600"
              aria-label="Copy raw JSON to clipboard"
              aria-describedby="copy-tooltip"
              title="Copy JSON"
            >
              {showCopied ? 'Copied' : 'Copy'}
            </button>
          </span>
          <span class="text-xs text-gray-500">{showRawJson ? 'Hide' : 'Show'}</span>
        </span>
      </summary>
      <pre
        class="mt-1 overflow-auto text-xs text-gray-700 dark:text-gray-300"
        style="max-height:360px;white-space:pre-wrap;">
{JSON.stringify(abstract, null, 2)}
    </pre>
    </details>
  </div>
</div>

<!-- Affiliation Details Dialog -->
<AffiliationDialog bind:open={showAffiliationDialog} affiliation={selectedAffiliation} />

<!-- Abstract Reviews Dialog -->
<AbstractReviewsDialog
  bind:open={showReviewsDialog}
  reviews={abstract.reviews || []}
  abstractTitle={abstract.title}
  onAffiliationClick={handleAffiliationClick}
/>