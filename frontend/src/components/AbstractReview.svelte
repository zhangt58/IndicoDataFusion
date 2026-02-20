<script>
  import Icon from '@iconify/svelte';
  import AffiliationBadge from './AffiliationBadge.svelte';
  import TrackBadge from '../pages/TrackBadge.svelte';
  import TypeBadge from '../pages/TypeBadge.svelte';
  import { formatDate, formatRatingValue, ACTION_STYLES } from '../lib/reviewUtils.js';

  let { review = null, onAffiliationClick = null } = $props();

  // Handle affiliation click
  function handleAffiliationClick(affiliation) {
    if (onAffiliationClick) {
      onAffiliationClick(affiliation);
    }
  }

  const actionStyle = $derived(review ? (ACTION_STYLES[review.proposed_action] ?? null) : null);
</script>

{#if review}
  <div
    class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-2 space-y-1"
  >
    <!-- Header: Reviewer and Action -->
    <div class="flex justify-between items-start gap-2">
      <div class="flex items-start gap-2 flex-1">
        {#if review.user && review.user.avatar_url}
          <img
            src={review.user.avatar_url}
            alt={`Avatar of ${review.user.full_name || 'reviewer'}`}
            class="w-8 h-8 rounded-full object-cover shrink-0 mt-1"
          />
        {:else}
          <Icon
            icon="mdi:account-circle"
            class="w-8 h-8 text-gray-500 dark:text-gray-400 shrink-0 mt-1"
          />
        {/if}
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-gray-800 dark:text-white">
            {review.user.full_name}
          </p>
          {#if review.user.title}
            <p class="text-xs text-gray-500 dark:text-gray-400">{review.user.title}</p>
          {/if}
          {#if review.user.affiliation}
            <AffiliationBadge
              affiliation={review.user.affiliation}
              onclick={handleAffiliationClick}
              className="text-xs text-blue-600 dark:text-blue-400 mt-1"
            />
          {/if}
        </div>
      </div>

      <!-- Proposed Action Badge -->
      {#if actionStyle}
        <div class="flex items-center gap-2 px-2 py-1 rounded-lg border {actionStyle.badgeClass}">
          <Icon icon={actionStyle.icon} class="w-4 h-4" />
          <span class="text-xs font-semibold">{actionStyle.label}</span>
        </div>
      {/if}
    </div>

    <!-- Track -->
    <div class="flex items-center gap-2 text-xs">
      <span class="text-gray-600 dark:text-gray-400 font-semibold">Track:</span>
      <TrackBadge text={review.track.title} type="reviewed" />
    </div>

    <!-- Timestamps -->
    <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
      <div class="flex items-center gap-1">
        <Icon icon="mdi:calendar" class="w-4 h-4" />
        <span>Created: {formatDate(review.created_dt)}</span>
      </div>
      {#if review.modified_dt}
        <div class="flex items-center gap-1">
          <Icon icon="mdi:calendar" class="w-4 h-4" />
          <span>Modified: {formatDate(review.modified_dt)}</span>
        </div>
      {/if}
    </div>

    <!-- Comment -->
    {#if review.comment && review.comment.trim()}
      <div class="bg-gray-50 dark:bg-gray-700 rounded p-2">
        <div class="flex items-center gap-1 mb-1">
          <Icon icon="mdi:message-text" class="w-4 h-4 text-gray-500 dark:text-gray-400" />
          <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Comment</span>
        </div>
        <p class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap wrap-break-word">
          {review.comment}
        </p>
      </div>
    {/if}

    <!-- Ratings -->
    {#if review.ratings && review.ratings.length > 0}
      <div class="bg-gray-50 dark:bg-gray-700 rounded p-2">
        <div class="flex items-center gap-1 mb-1">
          <Icon icon="mdi:star-outline" class="w-4 h-4 text-yellow-500" />
          <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Ratings</span>
        </div>
        <div class="space-y-1">
          {#each review.ratings as rating}
            <div class="flex items-start justify-between text-sm gap-1">
              <div class="flex-1 min-w-0">
                {#if rating.question_details}
                  <div class="text-gray-700 dark:text-gray-300 font-medium">
                    {rating.question_details.title}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                    Question {rating.question}
                  </div>
                {:else}
                  <span class="text-gray-600 dark:text-gray-400">Question {rating.question}:</span>
                {/if}
              </div>
              <span class="font-semibold text-gray-800 dark:text-white shrink-0">
                {formatRatingValue(rating.value)}
              </span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Proposed Contribution Type -->
    {#if review.proposed_contrib_type}
      <div class="flex items-center gap-2 text-xs">
        <span class="text-gray-600 dark:text-gray-400 font-semibold">Proposed Type:</span>
        <TypeBadge text={review.proposed_contrib_type.name} />
      </div>
    {/if}

    <!-- Proposed Tracks (for change_tracks action) -->
    {#if review.proposed_tracks && review.proposed_tracks.length > 0}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Proposed Tracks:</p>
        <div class="flex gap-2 flex-wrap">
          {#each review.proposed_tracks as track}
            <TrackBadge text={track.title} type="proposed" />
          {/each}
        </div>
      </div>
    {/if}

    <!-- Proposed Related Abstract (for mark_as_duplicate action) -->
    {#if review.proposed_related_abstract}
      <div
        class="bg-orange-50 dark:bg-orange-900/20 rounded p-2 border border-orange-200 dark:border-orange-800"
      >
        <p class="text-xs font-semibold text-orange-800 dark:text-orange-300 mb-1">Duplicate of:</p>
        <p class="text-sm text-gray-700 dark:text-gray-300">
          <span class="font-semibold">#{review.proposed_related_abstract.friendly_id}</span>
          - {review.proposed_related_abstract.title}
        </p>
      </div>
    {/if}

    <!-- Review ID (small footer) -->
    <div
      class="text-xs text-gray-400 dark:text-gray-500 pt-2 border-t border-gray-200 dark:border-gray-700"
    >
      Review ID: {review.id}
    </div>
  </div>
{:else}
  <div class="p-4 text-center text-gray-500 dark:text-gray-400 text-sm">
    No review data available
  </div>
{/if}
