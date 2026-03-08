<script>
  import Icon from '@iconify/svelte';
  import { Select, Radio } from 'flowbite-svelte';
  import {
    SubmitAbstractReview,
    UpdateAbstractReview,
    GetAbstracts,
  } from '../../wailsjs/go/main/App';
  import ReviewerDialog from './ReviewerDialog.svelte';
  import TrackBadge from '../pages/TrackBadge.svelte';
  import TypeBadge from '../pages/TypeBadge.svelte';
  import StateBadge from '../pages/StateBadge.svelte';
  import AffiliationBadge from './AffiliationBadge.svelte';
  import AffiliationDialog from './AffiliationDialog.svelte';
  import { getAllTracks } from '../pages/AbstractTableItem.js';
  import { formatDate, ACTION_STYLES } from '../lib/reviewUtils.js';
  import { voteStatsStore } from '../lib/voteStats.svelte.js';

  /**
   * Props:
   *   abstract      - AbstractData object (required)
   *   reviewTrack   - Track {id, title, code} or ReviewTrack {track_id, name} for new reviews
   *   onSuccess     - callback on successful submission
   *   onCancel      - callback on cancel
   */
  let { abstract = null, reviewTrack = null, onSuccess = null, onCancel = null } = $props();

  // ── Form state ──────────────────────────────────────────────────────────────
  let isSubmitting = $state(false);
  let error = $state(null);

  // question ratings map: questionID (int) → value (0 | 1)  stored as string for flowbite Select
  let questionRatings = $state({});
  // For priority questions: track which one (if any) is selected as "yes" (only one can be yes)
  let prioritySelection = $state('none'); // 'none' | 'first' | 'second'
  let proposedAction = $state('accept');
  let proposedContribTypeName = $state('');
  let proposedTrackIDs = $state([]);
  let proposedRelatedAbstractID = $state('');
  let relatedAbstractSearch = $state('');
  let comment = $state('');
  let selectedTrackID = $state(0); // for new reviews with multiple available tracks

  // Abstract list for related-abstract picker
  let allAbstracts = $state([]);
  let loadingAbstracts = $state(false);

  // Reviewer dialog
  let showReviewerDialog = $state(false);
  // Affiliation dialog for primary author
  let showAffiliationDialog = $state(false);

  // ── Derived ─────────────────────────────────────────────────────────────────
  const isEditMode = $derived(abstract?.my_review != null);
  const currentReview = $derived(abstract?.my_review);
  const questions = $derived(abstract?.questions || {});
  const contribTypes = $derived(abstract?.contrib_types || {});
  const availableTracks = $derived(abstract?.reviewed_for_tracks || []);

  const sortedQuestions = $derived(
    Object.entries(questions)
      .map(([qID, q]) => ({ id: parseInt(qID), ...q }))
      .sort((a, b) => (a.position ?? 0) - (b.position ?? 0)),
  );

  // Separate priority questions from others
  const priorityQuestions = $derived.by(() => {
    const first = sortedQuestions.find((q) => q.title?.toLowerCase() === 'first priority');
    const second = sortedQuestions.find((q) => q.title?.toLowerCase() === 'second priority');
    return { first, second };
  });

  const nonPriorityQuestions = $derived(
    sortedQuestions.filter(
      (q) =>
        q.title?.toLowerCase() !== 'first priority' && q.title?.toLowerCase() !== 'second priority',
    ),
  );

  // Sync prioritySelection with questionRatings for priority questions
  $effect(() => {
    if (priorityQuestions.first && priorityQuestions.second) {
      const firstID = priorityQuestions.first.id;
      const secondID = priorityQuestions.second.id;

      if (prioritySelection === 'first') {
        questionRatings[firstID] = '1';
        questionRatings[secondID] = '0';
      } else if (prioritySelection === 'second') {
        questionRatings[firstID] = '0';
        questionRatings[secondID] = '1';
      } else {
        questionRatings[firstID] = '0';
        questionRatings[secondID] = '0';
      }
    }
  });

  const contribTypeOptions = $derived(
    Object.entries(contribTypes).map(([name, id]) => ({ name, id })),
  );

  // Resolve track ID for new review submission
  const resolvedNewTrackID = $derived.by(() => {
    if (selectedTrackID > 0) return selectedTrackID;
    if (reviewTrack) return reviewTrack.id ?? reviewTrack.track_id ?? 0;
    if (availableTracks.length > 0) return availableTracks[0].id;
    return 0;
  });

  const effectiveTrackID = $derived(
    isEditMode ? (currentReview?.track?.id ?? 0) : resolvedNewTrackID,
  );

  const effectiveTrackTitle = $derived.by(() => {
    if (isEditMode && currentReview?.track) return currentReview.track.title;
    if (reviewTrack) return reviewTrack.title ?? reviewTrack.name ?? '';
    if (availableTracks.length > 0) {
      const t = availableTracks.find((t) => t.id === resolvedNewTrackID);
      return t?.title ?? availableTracks[0]?.title ?? '';
    }
    return '';
  });

  // All unique tracks from loaded abstracts (for change_tracks picker)
  const allLoadedTracks = $derived.by(() => {
    const tracks = getAllTracks(allAbstracts);
    // Sort tracks by ID (numerically)
    return tracks.sort((a, b) => {
      if (a.id == null && b.id == null) return 0;
      if (a.id == null) return 1;
      if (b.id == null) return -1;
      return a.id - b.id;
    });
  });

  // Filtered abstract list for related-abstract search
  const filteredAbstracts = $derived.by(() => {
    const q = relatedAbstractSearch.trim().toLowerCase();

    // For mark_as_duplicate and merge: only show abstracts assigned to the reviewer and exclude current
    let abstracts = allAbstracts;
    if (proposedAction === 'mark_as_duplicate' || proposedAction === 'merge') {
      abstracts = allAbstracts.filter((a) => {
        // Exclude the current abstract being reviewed
        if (a.id === abstract?.id) return false;
        // Only include abstracts assigned to this reviewer
        return a.is_my_review === true || a.my_review != null;
      });
    }

    if (!q) return abstracts.slice(0, 60);
    return abstracts
      .filter((a) => String(a.friendly_id).includes(q) || a.title?.toLowerCase().includes(q))
      .slice(0, 60);
  });

  // Currently selected related abstract object (for display)
  // proposedRelatedAbstractID stores the database ID (a.id), not friendly_id
  const selectedRelatedAbstract = $derived.by(() => {
    if (!proposedRelatedAbstractID) return null;
    const dbID = parseInt(proposedRelatedAbstractID);
    return allAbstracts.find((a) => a.id === dbID) ?? null;
  });

  // Primary author (first primary or first person)
  const primaryAuthor = $derived.by(() => {
    if (!abstract?.persons || abstract.persons.length === 0) return null;
    const primary = abstract.persons.find((p) => p.author_type === 'primary');
    return primary ?? abstract.persons[0];
  });

  // Sorted available tracks (by ID)
  const sortedAvailableTracks = $derived.by(() => {
    if (!availableTracks || availableTracks.length === 0) return [];
    return [...availableTracks].sort((a, b) => (a.id ?? 0) - (b.id ?? 0));
  });

  // Sorted reviewed for tracks (by ID)
  const sortedReviewedForTracks = $derived.by(() => {
    if (!abstract?.reviewed_for_tracks || abstract.reviewed_for_tracks.length === 0) return [];
    return [...abstract.reviewed_for_tracks].sort((a, b) => (a.id ?? 0) - (b.id ?? 0));
  });

  // Form validation: check if form is valid for submission
  const isFormValid = $derived.by(() => {
    // Must have a track ID
    if (!effectiveTrackID) return false;

    // Action-specific validation
    if (proposedAction === 'change_tracks' && proposedTrackIDs.length === 0) return false;
    if (
      (proposedAction === 'mark_as_duplicate' || proposedAction === 'merge') &&
      !proposedRelatedAbstractID.trim()
    )
      return false;

    return true;
  });

  // Check if form has changes (for edit mode)
  const hasChanges = $derived.by(() => {
    if (!isEditMode || !currentReview) return true; // always allow submit for new reviews

    // Check ratings changes
    for (const rating of currentReview.ratings ?? []) {
      const currentValue =
        typeof rating.value === 'boolean'
          ? rating.value
            ? '1'
            : '0'
          : String(Number(rating.value) || 0);
      const formValue = questionRatings[rating.question] ?? '0';
      if (currentValue !== formValue) return true;
    }

    // Check proposed action
    if ((currentReview.proposed_action || 'accept') !== proposedAction) return true;

    // Check proposed contribution type
    const currentContribTypeName = currentReview.proposed_contrib_type?.name ?? '';
    if (currentContribTypeName !== proposedContribTypeName) return true;

    // Check proposed tracks (for change_tracks action)
    const currentTrackIDs = (currentReview.proposed_tracks ?? [])
      .map((t) => t.id)
      .sort((a, b) => a - b);
    const formTrackIDs = [...proposedTrackIDs].sort((a, b) => a - b);
    if (
      currentTrackIDs.length !== formTrackIDs.length ||
      !currentTrackIDs.every((id, idx) => id === formTrackIDs[idx])
    )
      return true;

    // Check proposed related abstract
    const currentRelatedID = currentReview.proposed_related_abstract
      ? String(currentReview.proposed_related_abstract.id || '')
      : '';
    if (currentRelatedID !== proposedRelatedAbstractID) return true;

    // Check comment
    if ((currentReview.comment || '') !== comment) return true;

    return false;
  });

  // Submit button should be enabled only if form is valid and has changes (for edit mode)
  const canSubmit = $derived(isFormValid && hasChanges && !isSubmitting);

  // Vote stats for the effective track (from shared store)
  const currentTrackVotes = $derived.by(() => {
    if (!voteStatsStore.data?.per_track || !effectiveTrackID) return null;
    return voteStatsStore.data.per_track[effectiveTrackID] ?? null;
  });

  // Whether the current priority selection casts a vote (first or second = yes)
  const castingVote = $derived(prioritySelection === 'first' || prioritySelection === 'second');

  // ── Initialisation ──────────────────────────────────────────────────────────
  let _initialized = $state(false);

  $effect(() => {
    if (!abstract) return;

    if (isEditMode && currentReview) {
      const newRatings = {};
      for (const rating of currentReview.ratings ?? []) {
        newRatings[rating.question] =
          typeof rating.value === 'boolean'
            ? rating.value
              ? '1'
              : '0'
            : String(Number(rating.value) || 0);
      }
      for (const qID of Object.keys(questions)) {
        const id = parseInt(qID);
        if (!(id in newRatings)) newRatings[id] = '0';
      }
      questionRatings = newRatings;

      // Set prioritySelection based on ratings
      if (priorityQuestions.first && priorityQuestions.second) {
        const firstVal = newRatings[priorityQuestions.first.id] === '1';
        const secondVal = newRatings[priorityQuestions.second.id] === '1';
        if (firstVal) {
          prioritySelection = 'first';
        } else if (secondVal) {
          prioritySelection = 'second';
        } else {
          prioritySelection = 'none';
        }
      }

      proposedAction = currentReview.proposed_action || 'accept';
      proposedContribTypeName = currentReview.proposed_contrib_type?.name ?? '';
      proposedTrackIDs = currentReview.proposed_tracks?.map((t) => t.id) ?? [];
      proposedRelatedAbstractID = currentReview.proposed_related_abstract
        ? String(currentReview.proposed_related_abstract.id || '')
        : '';
      relatedAbstractSearch = '';
      comment = currentReview.comment || '';
    } else if (!_initialized) {
      const newRatings = {};
      for (const qID of Object.keys(questions)) {
        newRatings[parseInt(qID)] = '0';
      }
      questionRatings = newRatings;
      prioritySelection = 'none';
      proposedAction = 'accept';
      proposedContribTypeName = '';
      proposedTrackIDs = [];
      proposedRelatedAbstractID = '';
      relatedAbstractSearch = '';
      comment = '';
      if (sortedAvailableTracks.length > 0) {
        selectedTrackID = reviewTrack?.id ?? reviewTrack?.track_id ?? sortedAvailableTracks[0].id;
      }
    }
    _initialized = true;
  });

  // Lazily load the full abstract list for change_tracks, mark_as_duplicate, merge
  $effect(() => {
    if (
      (proposedAction === 'mark_as_duplicate' ||
        proposedAction === 'merge' ||
        proposedAction === 'change_tracks') &&
      allAbstracts.length === 0 &&
      !loadingAbstracts
    ) {
      loadingAbstracts = true;
      GetAbstracts()
        .then((list) => {
          allAbstracts = list ?? [];
        })
        .catch((e) => console.error('Failed to load abstracts for picker:', e))
        .finally(() => {
          loadingAbstracts = false;
        });
    }
  });

  // Load vote stats once when the form mounts if not already populated in the shared store
  $effect(() => {
    if (voteStatsStore.data !== null || voteStatsStore.loading) return;
    voteStatsStore.refresh();
  });

  // ── Helpers ─────────────────────────────────────────────────────────────────
  function toggleProposedTrack(trackID) {
    proposedTrackIDs = proposedTrackIDs.includes(trackID)
      ? proposedTrackIDs.filter((id) => id !== trackID)
      : [...proposedTrackIDs, trackID];
  }

  function selectRelatedAbstract(a) {
    // Store the database ID (a.id), not the friendly_id
    proposedRelatedAbstractID = String(a.id);
    relatedAbstractSearch = '';
  }

  function clearRelatedAbstract() {
    proposedRelatedAbstractID = '';
    relatedAbstractSearch = '';
  }

  // ── Submission ───────────────────────────────────────────────────────────────
  async function handleSubmit() {
    if (isSubmitting) return;
    error = null;
    isSubmitting = true;
    try {
      // Derive first/second priority values by title from question map
      let firstPriorityValue = 0;
      let secondPriorityValue = 0;
      for (const [qID, q] of Object.entries(questions)) {
        const title = q.title?.toLowerCase() ?? '';
        const id = parseInt(qID);
        if (title === 'first priority') firstPriorityValue = parseInt(questionRatings[id] ?? '0');
        if (title === 'second priority') secondPriorityValue = parseInt(questionRatings[id] ?? '0');
      }

      const contribTypeID =
        proposedContribTypeName && contribTypes[proposedContribTypeName]
          ? contribTypes[proposedContribTypeName]
          : null;
      const relatedAbstractID = proposedRelatedAbstractID.trim()
        ? parseInt(proposedRelatedAbstractID.trim())
        : null;

      if (isEditMode) {
        await UpdateAbstractReview(
          abstract.id,
          currentReview.id,
          effectiveTrackID,
          firstPriorityValue,
          secondPriorityValue,
          proposedAction,
          contribTypeID,
          proposedTrackIDs.length > 0 ? proposedTrackIDs : null,
          relatedAbstractID,
          comment,
        );
      } else {
        await SubmitAbstractReview(
          abstract.id,
          effectiveTrackID,
          firstPriorityValue,
          secondPriorityValue,
          proposedAction,
          contribTypeID,
          proposedTrackIDs.length > 0 ? proposedTrackIDs : null,
          relatedAbstractID,
          comment,
        );
      }
      if (onSuccess) onSuccess();
      // Vote stats refresh is handled by the caller (AbstractReviewFormDialog) after
      // RefreshAbstractByID completes, ensuring GetVoteStats reads the updated cache.
    } catch (err) {
      error = err.message || String(err);
    } finally {
      isSubmitting = false;
    }
  }

  function handleCancel() {
    if (onCancel) onCancel();
  }
</script>

<!-- Root card — same container as AbstractReview -->
<div
  class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-2 space-y-2"
>
  <!-- ══════════════════════════════════════════════════════════════════════════
       ABSTRACT INFORMATION SECTION
       ══════════════════════════════════════════════════════════════════════════ -->
  <div
    class="bg-linear-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-lg border border-blue-200 dark:border-blue-800 p-2 space-y-1.5"
  >
    <!-- Abstract ID, State, and Title -->
    <div class="flex items-center justify-between gap-2 mb-1">
      <div class="flex items-center gap-2">
        <span class="text-sm font-bold text-blue-700 dark:text-blue-300">
          #{abstract?.friendly_id}
        </span>
        <h4 class="text-sm font-semibold text-gray-800 dark:text-white mt-0.5 leading-snug">
          {abstract?.title}
        </h4>
      </div>
      <div class="flex items-center gap-1">
        {#if abstract?.state}
          <StateBadge state={abstract.state} />
        {/if}
        {#if abstract?.submitted_contrib_type}
          <TypeBadge text={abstract.submitted_contrib_type.name} />
        {/if}
      </div>
    </div>

    <!-- Submitted For Tracks -->
    {#if abstract?.submitted_for_tracks && abstract.submitted_for_tracks.length > 0}
      <div class="flex items-center gap-2 text-xs">
        <span class="text-gray-600 dark:text-gray-400 font-semibold shrink-0 pt-0.5"
          >Submitted:</span
        >
        <div class="flex flex-wrap gap-1">
          {#each abstract.submitted_for_tracks.sort((a, b) => (a.id ?? 0) - (b.id ?? 0)) as track (track.id)}
            <TrackBadge text={track.title || track.code} type="" />
          {/each}
        </div>
      </div>
    {/if}

    <!-- Reviewed For Tracks -->
    {#if sortedReviewedForTracks.length > 0}
      <div class="flex items-center gap-2 text-xs">
        <span class="text-gray-600 dark:text-gray-400 font-semibold shrink-0 pt-0.5">Reviewed:</span
        >
        <div class="flex flex-wrap gap-1">
          {#each sortedReviewedForTracks as track (track.id)}
            <TrackBadge text={track.title} type="reviewed" />
          {/each}
        </div>
      </div>
    {/if}

    <!-- Primary Author with Affiliation -->
    {#if primaryAuthor}
      <div
        class="flex items-center gap-2 text-xs pt-1 border-t border-blue-200 dark:border-blue-800"
      >
        <span class="text-gray-600 dark:text-gray-400 font-semibold shrink-0 pt-0.5">
          <Icon icon="mdi:account" class="w-3.5 h-3.5 inline" />
          Author:
        </span>
        <div class="flex-1">
          <div class="font-medium text-gray-800 dark:text-gray-200">
            {primaryAuthor.first_name}
            {primaryAuthor.last_name}
            {#if primaryAuthor.author_type === 'primary'}
              <span class="text-blue-600 dark:text-blue-400">(Primary)</span>
            {/if}
            {#if primaryAuthor.is_speaker}
              <span class="ml-1">🎤</span>
            {/if}
          </div>
          {#if primaryAuthor.affiliation}
            <AffiliationBadge
              affiliation={primaryAuthor.affiliation}
              onclick={() => (showAffiliationDialog = true)}
            />
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <!-- ══════════════════════════════════════════════════════════════════════════
       MY LAST REVIEW (Edit mode only)
       ══════════════════════════════════════════════════════════════════════════ -->
  {#if isEditMode && currentReview}
    <div
      class="bg-amber-50 dark:bg-amber-900/20 rounded-lg border border-amber-200 dark:border-amber-800 p-2 space-y-2"
    >
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <Icon icon="mdi:pencil-box-outline" class="w-4 h-4 text-amber-600 dark:text-amber-400" />
          <span class="text-sm font-semibold text-amber-900 dark:text-amber-200"
            >My Last Review</span
          >
        </div>
        <!-- Proposed action badge -->
        <div
          class="flex items-center gap-1.5 px-2 py-1 rounded-lg border text-xs font-semibold {ACTION_STYLES[
            currentReview.proposed_action
          ].badgeClass}"
        >
          <Icon icon={ACTION_STYLES[currentReview.proposed_action].icon} class="w-3.5 h-3.5" />
          {ACTION_STYLES[currentReview.proposed_action].label || currentReview.proposed_action}
        </div>
      </div>

      <!-- Ratings summary -->
      {#if currentReview.ratings && currentReview.ratings.length > 0}
        <div class="flex flex-wrap gap-2 text-xs">
          {#each currentReview.ratings as rating (rating.question)}
            {@const question = questions[rating.question]}
            {@const value =
              typeof rating.value === 'boolean'
                ? rating.value
                  ? 1
                  : 0
                : Number(rating.value) || 0}
            <div
              class="flex items-center gap-1 px-2 py-0.5 rounded {value === 1
                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                : 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300'}"
            >
              <span class="font-medium">{question?.title || `Q${rating.question}`}:</span>
              <span class="font-semibold">{value === 1 ? 'Yes' : 'No'}</span>
            </div>
          {/each}
        </div>
      {/if}

      <!-- Timestamps -->
      <div class="flex flex-wrap items-center gap-3 text-xs text-gray-600 dark:text-gray-400">
        <div class="flex items-center gap-1">
          <Icon icon="mdi:calendar-clock" class="w-3.5 h-3.5" />
          <span>Created: {formatDate(currentReview.created_dt)}</span>
        </div>
        {#if currentReview.modified_dt && currentReview.modified_dt !== currentReview.created_dt}
          <div class="flex items-center gap-1">
            <Icon icon="mdi:calendar-edit" class="w-3.5 h-3.5" />
            <span>Modified: {formatDate(currentReview.modified_dt)}</span>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- ══════════════════════════════════════════════════════════════════════════
       REVIEW FORM HEADER
       ══════════════════════════════════════════════════════════════════════════ -->
  <div
    class="flex items-center justify-between gap-2 pt-2 border-t border-gray-200 dark:border-gray-700"
  >
    <div class="flex items-center gap-2">
      <Icon
        icon={isEditMode ? 'mdi:pencil-circle' : 'mdi:clipboard-edit-outline'}
        class="w-5 h-5 text-blue-600 dark:text-blue-400"
      />
      <h5 class="text-base font-semibold text-gray-900 dark:text-white">
        {isEditMode ? 'Update' : 'New'}
      </h5>
    </div>

    <!-- Track selection (new reviews with multiple tracks) -->
    {#if !isEditMode && sortedAvailableTracks.length > 1}
      <div class="flex items-center gap-2">
        <span class="text-xs text-gray-600 dark:text-gray-400 font-medium">Track:</span>
        <div class="flex flex-wrap gap-1">
          {#each sortedAvailableTracks as track (track.id)}
            <button
              type="button"
              onclick={() => (selectedTrackID = track.id)}
              aria-pressed={resolvedNewTrackID === track.id}
            >
              <TrackBadge
                text={track.title}
                type={resolvedNewTrackID === track.id ? 'reviewed' : ''}
              />
            </button>
          {/each}
        </div>
      </div>
    {:else}
      <TrackBadge text={effectiveTrackTitle || '—'} type="reviewed" />
    {/if}
  </div>

  <!-- ══════════════════════════════════════════════════════════════════════════
       RATINGS SECTION
       ══════════════════════════════════════════════════════════════════════════ -->
  {#if sortedQuestions.length > 0}
    <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-2">
      <div class="flex items-center gap-2 mb-0.5">
        <Icon icon="mdi:star-outline" class="w-4 h-4 text-yellow-500" />
        <span class="text-sm font-semibold text-gray-700 dark:text-gray-300">Ratings</span>

        <!-- Vote counter pill for the current track -->
        {#if currentTrackVotes}
          {@const overLimit = currentTrackVotes.votes_cast >= currentTrackVotes.votes_max}
          {@const wouldExceed =
            castingVote &&
            !isEditMode &&
            currentTrackVotes.votes_cast >= currentTrackVotes.votes_max}
          <span
            class="ml-auto flex items-center gap-1 px-1.5 py-0.5 rounded text-[0.65rem] font-semibold border {overLimit
              ? 'bg-red-100 border-red-300 text-red-700 dark:bg-red-900/30 dark:border-red-600 dark:text-red-300'
              : currentTrackVotes.votes_left <= 1
                ? 'bg-amber-100 border-amber-300 text-amber-700 dark:bg-amber-900/30 dark:border-amber-600 dark:text-amber-300'
                : 'bg-green-100 border-green-300 text-green-700 dark:bg-green-900/30 dark:border-green-600 dark:text-green-300'}"
            title="Votes cast for this track: {currentTrackVotes.votes_cast} / {currentTrackVotes.votes_max} max"
          >
            <Icon icon="mdi:vote" class="w-3 h-3" />
            {currentTrackVotes.votes_cast}/{currentTrackVotes.votes_max} votes
            {#if overLimit}
              <Icon icon="mdi:alert" class="w-3 h-3" />
            {:else}
              · {currentTrackVotes.votes_left} left
            {/if}
          </span>
          {#if wouldExceed}
            <span class="text-[0.65rem] text-red-600 dark:text-red-400 font-medium">
              ⚠ Exceeds track limit
            </span>
          {/if}
        {/if}
      </div>
      <div class="space-y-0.5">
        <!-- Priority Questions with Radio buttons (only one can be Yes) -->
        {#if priorityQuestions.first && priorityQuestions.second}
          <div
            class="bg-white dark:bg-gray-800 rounded p-2 space-y-1.5 border border-gray-200 dark:border-gray-600"
          >
            <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">
              Select priority (only one can be Yes):
            </p>
            <div class="flex flex-col gap-2">
              <Radio name="priority" value="none" bind:group={prioritySelection} class="text-sm">
                <span class="text-sm text-gray-700 dark:text-gray-300"> Neither (both No) </span>
              </Radio>
              <Radio name="priority" value="first" bind:group={prioritySelection} class="text-sm">
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  <span
                    class="text-xs text-gray-500 dark:text-gray-400 font-mono rounded-sm border p-0.5"
                  >
                    Q-{priorityQuestions.first.id}
                  </span>
                  {' '}{priorityQuestions.first.title}
                </span>
              </Radio>
              <Radio name="priority" value="second" bind:group={prioritySelection} class="text-sm">
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  <span
                    class="text-xs text-gray-500 dark:text-gray-400 font-mono rounded-sm border p-0.5"
                  >
                    Q-{priorityQuestions.second.id}
                  </span>
                  {' '}{priorityQuestions.second.title}
                </span>
              </Radio>
            </div>
          </div>
        {/if}

        <!-- Non-Priority Questions with Select -->
        {#each nonPriorityQuestions as q (q.id)}
          <div class="flex items-center justify-between gap-2">
            <div class="flex-1 min-w-0">
              <label
                for="question-{q.id}"
                class="text-sm text-gray-700 dark:text-gray-300 font-medium cursor-pointer"
              >
                <span
                  class="text-xs text-gray-500 dark:text-gray-400 font-mono rounded-sm border p-0.5"
                  >Q-{q.id}</span
                >
                {' '}{q.title}
              </label>
            </div>
            <!-- flowbite-svelte Select for rating value -->
            <Select
              id="question-{q.id}"
              size="sm"
              underline
              bind:value={questionRatings[q.id]}
              items={[
                { value: '0', name: 'No' },
                { value: '1', name: 'Yes' },
              ]}
              class="w-20 shrink-0 font-semibold text-sm
                {(questionRatings[q.id] ?? '0') === '1'
                ? '!border-green-400 dark:!border-green-600 !text-green-700 dark:!text-green-300'
                : '!border-red-300 dark:!border-red-700 !text-red-700 dark:!text-red-300'}"
              aria-label={`Rating for ${q.title}`}
            />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- ── Proposed Action — badge toggles matching AbstractReview badge style ─ -->
  <div class="bg-gray-50 dark:bg-gray-700 rounded p-2">
    <div class="flex items-center gap-1 mb-1.5">
      <Icon
        icon="mdi:checkbox-marked-circle-outline"
        class="w-4 h-4 text-gray-500 dark:text-gray-400"
      />
      <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Proposed Action</span>
    </div>
    <div class="flex flex-wrap gap-2" role="group" aria-label="Proposed action">
      {#each Object.entries(ACTION_STYLES) as [value, style] (value)}
        <button
          type="button"
          onclick={() => (proposedAction = value)}
          aria-pressed={proposedAction === value}
          class="flex items-center gap-1.5 px-2 py-1 rounded-lg border text-xs font-semibold transition-all
            {proposedAction === value
            ? style.badgeClass + ' ring-2 ring-offset-1 ring-current'
            : 'bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400 border-gray-200 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-400'}"
        >
          <Icon icon={style.icon} class="w-3.5 h-3.5" />
          {style.label}
        </button>
      {/each}
    </div>
  </div>

  <!-- ── Contribution Type — only for accept (not reject) ─────────────────── -->
  {#if proposedAction === 'accept'}
    <div class="flex items-center gap-2 text-xs flex-wrap">
      <div class="flex items-center gap-1">
        {#if abstract?.indico_url}
          <a
            href={abstract.indico_url}
            target="_blank"
            rel="noopener noreferrer"
            title="The full list of contribution types is only available on the original review web page, click to open."
            class="inline-flex items-center"
            aria-label="Open original abstract page (Indico)"
          >
            <Icon
              icon="mdi:information-outline"
              class="w-3.5 h-3.5 text-gray-500 dark:text-gray-400 cursor-pointer"
              aria-hidden="false"
            />
          </a>
        {/if}
        <span class="text-gray-600 dark:text-gray-400 font-semibold">Proposed Type:</span>
      </div>
      <div class="flex flex-wrap gap-1">
        <button
          type="button"
          onclick={() => (proposedContribTypeName = '')}
          aria-pressed={proposedContribTypeName === ''}
          class="px-2 py-1 rounded-sm text-xs border transition-colors
            {proposedContribTypeName === ''
            ? 'bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 border-indigo-300 dark:border-indigo-700 ring-1 ring-indigo-400'
            : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-200 dark:border-gray-600 hover:border-gray-400'}"
        >
          — None —
        </button>
        {#each contribTypeOptions as opt (opt.name)}
          <button
            type="button"
            onclick={() => (proposedContribTypeName = opt.name)}
            aria-pressed={proposedContribTypeName === opt.name}
            class="transition-all {proposedContribTypeName === opt.name
              ? 'ring-2 ring-offset-1 ring-indigo-500'
              : 'opacity-70 hover:opacity-100'}"
          >
            <TypeBadge text={opt.name} />
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- ── Proposed Tracks — list view with TrackBadge for each track ─────────
       Uses getAllTracks across all abstracts to show the full conference track list -->
  {#if proposedAction === 'change_tracks'}
    <div class="bg-gray-50 dark:bg-gray-700 rounded p-2 space-y-1.5">
      <div class="flex items-center gap-1">
        <Icon icon="mdi:tag-multiple-outline" class="w-4 h-4 text-blue-500" />
        <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
          Proposed Tracks
          <span class="font-normal text-gray-400 dark:text-gray-500">(select one or more)</span>
        </span>
        {#if loadingAbstracts}
          <Icon icon="mdi:loading" class="w-3.5 h-3.5 animate-spin text-gray-400 ml-1" />
        {/if}
      </div>

      <!-- Track list — one row per track, same layout as abstract track rows -->
      {#if allLoadedTracks.length === 0 && !loadingAbstracts}
        <!-- Fallback: only available_for_review tracks -->
        <div class="flex flex-col gap-0.5">
          {#each availableTracks as track (track.id)}
            {@const selected = proposedTrackIDs.includes(track.id)}
            <button
              type="button"
              onclick={() => toggleProposedTrack(track.id)}
              aria-pressed={selected}
              class="flex items-center justify-between px-2 py-1 rounded border text-xs transition-colors
                {selected
                ? 'bg-yellow-50 dark:bg-yellow-900/30 border-yellow-400 dark:border-yellow-600'
                : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-600 hover:border-gray-400'}"
            >
              <TrackBadge text={track.title} type={selected ? 'proposed' : ''} />
              {#if selected}
                <Icon
                  icon="mdi:check-circle"
                  class="w-4 h-4 text-yellow-600 dark:text-yellow-400 shrink-0"
                />
              {/if}
            </button>
          {/each}
        </div>
      {:else}
        <div class="flex flex-col gap-1 max-h-40 overflow-y-auto">
          {#each allLoadedTracks as track (track.id ?? track.label)}
            {@const selected = proposedTrackIDs.includes(track.id)}
            <button
              type="button"
              onclick={() => toggleProposedTrack(track.id)}
              aria-pressed={selected}
              class="flex items-center justify-between px-2 py-1 rounded border text-xs transition-colors
                {selected
                ? 'bg-yellow-50 dark:bg-yellow-900/30 border-yellow-400 dark:border-yellow-600'
                : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-600 hover:border-gray-400'}"
            >
              <TrackBadge text={track.title || track.label} type={selected ? 'proposed' : ''} />
              {#if selected}
                <Icon
                  icon="mdi:check-circle"
                  class="w-4 h-4 text-yellow-600 dark:text-yellow-400 shrink-0"
                />
              {/if}
            </button>
          {/each}
        </div>
      {/if}

      {#if proposedTrackIDs.length === 0}
        <p class="text-xs text-orange-600 dark:text-orange-400 flex items-center gap-1">
          <Icon icon="mdi:alert-outline" class="w-3.5 h-3.5" />
          Select at least one track.
        </p>
      {:else}
        <!-- Selected tracks summary -->
        <div class="flex items-center flex-wrap gap-1 pt-0.5">
          <span class="text-xs text-gray-500 dark:text-gray-400">Selected:</span>
          {#each allLoadedTracks.filter( (t) => proposedTrackIDs.includes(t.id), ) as t (t.id ?? t.label)}
            <TrackBadge text={t.title || t.label} type="proposed" />
          {/each}
          {#each availableTracks.filter((t) => proposedTrackIDs.includes(t.id) && !allLoadedTracks.some((lt) => lt.id === t.id)) as t (t.id)}
            <TrackBadge text={t.title} type="proposed" />
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- ── Related Abstract — single-line list: #ID · state · title ─────────── -->
  {#if proposedAction === 'mark_as_duplicate' || proposedAction === 'merge'}
    <div
      class="bg-orange-50 dark:bg-orange-900/20 rounded p-2 border border-orange-200 dark:border-orange-800 space-y-1"
    >
      <div class="flex items-center gap-1">
        <p class="text-xs font-semibold text-orange-800 dark:text-orange-300">
          {proposedAction === 'merge' ? 'Merge into Abstract:' : 'Duplicate of:'}
        </p>
        {#if loadingAbstracts}
          <Icon icon="mdi:loading" class="w-3.5 h-3.5 animate-spin text-orange-500" />
        {/if}
      </div>

      <!-- Selected abstract display (same style as AbstractReview proposed_related_abstract) -->
      {#if selectedRelatedAbstract}
        <div
          class="flex items-center justify-between gap-2 px-2 py-1 rounded bg-white dark:bg-gray-800 border border-orange-200 dark:border-orange-700"
        >
          <div class="flex items-center gap-1.5 flex-1 min-w-0">
            <span class="text-xs font-semibold text-orange-700 dark:text-orange-300 shrink-0">
              #{selectedRelatedAbstract.friendly_id}
            </span>
            {#if selectedRelatedAbstract.state}
              <StateBadge state={selectedRelatedAbstract.state} />
            {/if}
            <span
              class="text-sm text-gray-700 dark:text-gray-300 truncate"
              title={selectedRelatedAbstract.title}
            >
              {selectedRelatedAbstract.title}
            </span>
          </div>
          <button
            type="button"
            onclick={clearRelatedAbstract}
            class="shrink-0 text-gray-400 hover:text-red-500 transition-colors"
            aria-label="Clear selection"
          >
            <Icon icon="mdi:close-circle" class="w-4 h-4" />
          </button>
        </div>
      {:else if proposedRelatedAbstractID}
        <div
          class="flex items-center justify-between gap-2 px-2 py-1 rounded bg-white dark:bg-gray-800 border border-orange-200 dark:border-orange-700"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300"
            >Abstract #{proposedRelatedAbstractID}</span
          >
          <button
            type="button"
            onclick={clearRelatedAbstract}
            class="shrink-0 text-gray-400 hover:text-red-500 transition-colors"
            aria-label="Clear selection"
          >
            <Icon icon="mdi:close-circle" class="w-4 h-4" />
          </button>
        </div>
      {/if}

      <!-- Search box -->
      <div
        class="flex items-center gap-1.5 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
      >
        <Icon icon="mdi:magnify" class="w-4 h-4 text-gray-400 shrink-0" />
        <input
          type="text"
          bind:value={relatedAbstractSearch}
          placeholder="Search by ID or title…"
          class="flex-1 text-sm bg-transparent outline-none text-gray-900 dark:text-gray-100 placeholder-gray-400"
          aria-label="Search abstracts"
        />
      </div>

      <!-- Single-line list: #ID · title · StateBadge ──────────────────────────── -->
      {#if relatedAbstractSearch.trim() || (!selectedRelatedAbstract && !proposedRelatedAbstractID)}
        <div
          class="flex flex-col max-h-36 overflow-y-auto border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-gray-800 divide-y divide-gray-100 dark:divide-gray-700 shadow-sm"
        >
          {#if filteredAbstracts.length === 0}
            <p class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 text-center">
              {loadingAbstracts ? 'Loading…' : 'No results'}
            </p>
          {:else}
            {#each filteredAbstracts as a (a.id)}
              <button
                type="button"
                onclick={() => selectRelatedAbstract(a)}
                class="flex items-center gap-1.5 px-2 py-1 text-left hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-colors w-full"
              >
                <!-- #ID -->
                <span
                  class="text-xs font-semibold text-orange-700 dark:text-orange-300 shrink-0 w-8 text-right"
                >
                  #{a.friendly_id}
                </span>
                <!-- full title — truncated to one line -->
                <span
                  class="text-sm text-gray-700 dark:text-gray-300 truncate flex-1 min-w-0"
                  title={a.title}
                >
                  {a.title}
                </span>
                <!-- state badge at the end -->
                {#if a.state}
                  <StateBadge state={a.state} />
                {/if}
              </button>
            {/each}
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  <!-- ── Comment — same style as AbstractReview comment section ───────────── -->
  <div class="bg-gray-50 dark:bg-gray-700 rounded p-2">
    <div class="flex items-center gap-1 mb-1">
      <Icon icon="mdi:message-text" class="w-4 h-4 text-gray-500 dark:text-gray-400" />
      <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">Comment</span>
      <span class="text-xs text-gray-400 dark:text-gray-500">(optional)</span>
    </div>
    <textarea
      id="review-comment"
      bind:value={comment}
      rows="3"
      placeholder="Enter your review comment…"
      class="w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 resize-y min-h-20 whitespace-pre-wrap"
    ></textarea>
  </div>

  <!-- ── Error ─────────────────────────────────────────────────────────────── -->
  {#if error}
    <div
      class="p-2 bg-red-50 dark:bg-red-900/20 border border-red-300 dark:border-red-800 rounded text-sm text-red-700 dark:text-red-400 flex items-start gap-2"
      role="alert"
    >
      <Icon icon="mdi:alert-circle-outline" class="w-5 h-5 shrink-0 mt-0.5" />
      <span>{error}</span>
    </div>
  {/if}

  <!-- Validation messages for incomplete form -->
  {#if !isFormValid && !error}
    <div
      class="p-2 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-800 rounded text-sm text-yellow-800 dark:text-yellow-300 flex items-start gap-2"
      role="alert"
    >
      <Icon icon="mdi:information-outline" class="w-5 h-5 shrink-0 mt-0.5" />
      <div>
        {#if proposedAction === 'change_tracks' && proposedTrackIDs.length === 0}
          <span>Please select at least one proposed track.</span>
        {:else if (proposedAction === 'mark_as_duplicate' || proposedAction === 'merge') && !proposedRelatedAbstractID.trim()}
          <span>Please select the related abstract.</span>
        {:else if !effectiveTrackID}
          <span>No review track available. Please select a track.</span>
        {:else}
          <span>Please complete all required fields.</span>
        {/if}
      </div>
    </div>
  {/if}

  <!-- No changes warning for edit mode -->
  {#if isEditMode && !hasChanges && !error}
    <div
      class="p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-300 dark:border-blue-800 rounded text-sm text-blue-700 dark:text-blue-300 flex items-start gap-2"
      role="alert"
    >
      <Icon icon="mdi:information-outline" class="w-5 h-5 shrink-0 mt-0.5" />
      <span>No changes detected. Modify the review to enable the Update button.</span>
    </div>
  {/if}

  <!-- ── Submit / Cancel ───────────────────────────────────────────────────── -->
  <div class="flex justify-end gap-2 pt-1 border-t border-gray-200 dark:border-gray-700">
    <button
      type="button"
      onclick={handleCancel}
      disabled={isSubmitting}
      class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    >
      Cancel
    </button>
    <button
      type="button"
      onclick={handleSubmit}
      disabled={!canSubmit}
      title={!canSubmit && !isSubmitting
        ? isEditMode && !hasChanges
          ? 'No changes to save'
          : 'Please complete all required fields'
        : ''}
      class="px-4 py-1.5 text-sm font-medium text-white rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1.5
        {isEditMode ? 'bg-amber-600 hover:bg-amber-700' : 'bg-blue-600 hover:bg-blue-700'}"
    >
      {#if isSubmitting}
        <Icon icon="mdi:loading" class="w-4 h-4 animate-spin" />
        {isEditMode ? 'Updating…' : 'Submitting…'}
      {:else}
        <Icon icon={isEditMode ? 'mdi:pencil-check' : 'mdi:send'} class="w-4 h-4" />
        {isEditMode ? 'Update' : 'Submit'}
      {/if}
    </button>
  </div>
</div>

<AffiliationDialog bind:open={showAffiliationDialog} affiliation={primaryAuthor.affiliation} />

<!-- ReviewerDialog — opened when clicking the ReviewerCard -->
{#if isEditMode && currentReview?.user}
  <ReviewerDialog bind:open={showReviewerDialog} reviewer={currentReview.user} />
{/if}
