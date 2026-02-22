<script>
  /**
   * ReviewMatrixView
   *
   * Props:
   *   abstracts             – Array<AbstractData>
   *   firstPriorityWeight   – number  weight applied to first_priority  (default 1)
   *   secondPriorityWeight  – number  weight applied to second_priority (default 1)
   *   columns               – number  grid columns  (default 0 = auto)
   *   title                 – string  optional heading
   *   onAbstractUpdated     – callback(refreshedAbstract)
   */

  import AbstractReviewsDialog from './AbstractReviewsDialog.svelte';
  import AbstractReviewFormDialog from './AbstractReviewFormDialog.svelte';
  import Icon from '@iconify/svelte';
  import { ACTION_STYLES } from '../lib/reviewUtils.js';

  let {
    abstracts = [],
    firstPriorityWeight = $bindable(1),
    secondPriorityWeight = $bindable(1),
    columns = 0,
    title = 'Review Matrix',
    onAbstractUpdated = null,
  } = $props();

  // ── action colour palette ─────────────────────────────────────────────────
  const ACTION_COLORS = {
    accept: { bg: '#dcfce7', border: '#16a34a', text: '#15803d' },
    reject: { bg: '#fee2e2', border: '#dc2626', text: '#b91c1c' },
    change_tracks: { bg: '#dbeafe', border: '#2563eb', text: '#1d4ed8' },
    mark_as_duplicate: { bg: '#ffedd5', border: '#ea580c', text: '#c2410c' },
    merge: { bg: '#f3e8ff', border: '#9333ea', text: '#7e22ce' },
    none: { bg: '#f3f4f6', border: '#d1d5db', text: '#6b7280' },
    mixed: { bg: '#fef9c3', border: '#ca8a04', text: '#854d0e' },
  };

  // "Mine" is indicated by a sky-blue dot only (no outline border)
  const MY_DOT_COLOR = '#0ea5e9'; // sky-500  – assigned but not yet reviewed
  const DONE_DOT_COLOR = '#16a34a'; // green-600 – already reviewed by me

  // ── dialog state ──────────────────────────────────────────────────────────
  let showReviewsDialog = $state(false);
  let showReviewFormDialog = $state(false);
  let selectedAbstract = $state(null);

  // ── controls state ────────────────────────────────────────────────────────
  let showWeightControls = $state(false);
  let sortByScore = $state(false);

  // ── helpers ───────────────────────────────────────────────────────────────

  function consensusAction(abstract) {
    const reviews = abstract.reviews;
    if (!reviews || reviews.length === 0) return 'none';
    const actions = reviews.map((r) => r.proposed_action).filter(Boolean);
    if (actions.length === 0) return 'none';
    const counts = {};
    for (const a of actions) counts[a] = (counts[a] || 0) + 1;
    const sorted = Object.entries(counts).sort((a, b) => b[1] - a[1]);
    if (sorted.length > 1 && sorted[0][1] === sorted[1][1]) return 'mixed';
    return sorted[0][0];
  }

  function calcScore(abstract) {
    const fp = abstract.first_priority || 0;
    const sp = abstract.second_priority || 0;
    return firstPriorityWeight * fp + secondPriorityWeight * sp;
  }

  function fmtScore(score) {
    return score % 1 === 0 ? String(Math.round(score)) : score.toFixed(1);
  }

  function primaryTrack(abstract) {
    if (abstract.reviewed_for_tracks?.length > 0)
      return abstract.reviewed_for_tracks[0].title || '';
    if (abstract.submitted_for_tracks?.length > 0)
      return abstract.submitted_for_tracks[0].title || '';
    return '';
  }

  // ── sorted + grouped ──────────────────────────────────────────────────────

  const groupedByTrack = $derived.by(() => {
    const sorted = [...(abstracts || [])].sort((a, b) => {
      const ta = primaryTrack(a).toLowerCase();
      const tb = primaryTrack(b).toLowerCase();
      if (ta < tb) return -1;
      if (ta > tb) return 1;
      return (a.id || 0) - (b.id || 0);
    });

    const groups = [];
    const seen = new Map();
    for (const abstract of sorted) {
      const track = primaryTrack(abstract) || '(No Track)';
      if (!seen.has(track)) {
        const g = { trackTitle: track, abstracts: [] };
        seen.set(track, g);
        groups.push(g);
      }
      seen.get(track).abstracts.push(abstract);
    }

    // Optionally sort each group by descending weighted score
    if (sortByScore) {
      for (const g of groups) {
        g.abstracts = [...g.abstracts].sort((a, b) => calcScore(b) - calcScore(a));
      }
    }
    return groups;
  });

  const totalAbstracts = $derived((abstracts || []).length);
  const myReviewCount = $derived((abstracts || []).filter((a) => a.is_my_review).length);
  const reviewedCount = $derived((abstracts || []).filter((a) => a.my_review != null).length);

  const effectiveColumns = $derived.by(() => {
    if (columns > 0) return columns;
    const n = totalAbstracts;
    if (n <= 50) return 10;
    if (n <= 120) return 15;
    if (n <= 250) return 20;
    return 25;
  });

  // ── click handlers ────────────────────────────────────────────────────────

  function openAbstract(abstract) {
    selectedAbstract = abstract;
    if (abstract.is_my_review) {
      showReviewFormDialog = true;
    } else {
      showReviewsDialog = true;
    }
  }

  function handleAbstractUpdated(refreshed) {
    selectedAbstract = refreshed;
    if (typeof onAbstractUpdated === 'function') onAbstractUpdated(refreshed);
  }

  // ── legend ────────────────────────────────────────────────────────────────
  const legendItems = [
    ...Object.entries(ACTION_STYLES).map(([key, style]) => ({
      key,
      label: style.label,
      color: ACTION_COLORS[key] ?? ACTION_COLORS.none,
    })),
    { key: 'mixed', label: 'Mixed', color: ACTION_COLORS.mixed },
    { key: 'none', label: 'No Review', color: ACTION_COLORS.none },
  ];
</script>

<!-- ── Header bar: title + summary + legend + controls toggle ────────────── -->
<div
  class="flex flex-wrap items-center gap-x-4 gap-y-1 px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/60 text-xs select-none"
>
  {#if title}
    <span class="font-semibold text-gray-700 dark:text-gray-200 shrink-0">{title}</span>
  {/if}

  <span class="text-gray-500 dark:text-gray-400">
    {totalAbstracts} abstracts · {myReviewCount} assigned · {reviewedCount} reviewed by me
  </span>

  <!-- Legend swatches -->
  <div class="flex flex-wrap items-center gap-2 ml-auto">
    {#each legendItems as item (item.key)}
      <span class="flex items-center gap-1">
        <span
          class="inline-block w-3 h-3 rounded-sm border"
          style="background:{item.color.bg};border-color:{item.color.border}"
        ></span>
        <span class="text-gray-600 dark:text-gray-300">{item.label}</span>
      </span>
    {/each}

    <!-- Mine: shown as a coloured dot (same as cell dot indicator) -->
    <span class="flex items-center gap-1.5">
      <span
        class="relative inline-block w-3 h-3 rounded-sm border"
        style="background:#f3f4f6;border-color:#d1d5db"
      >
        <!-- green dot = reviewed -->
        <span
          class="absolute top-0 right-0 w-1.5 h-1.5 rounded-full -translate-y-px translate-x-px"
          style="background:{DONE_DOT_COLOR}"
        ></span>
      </span>
      <span class="text-gray-600 dark:text-gray-300">Reviewed</span>
    </span>
    <span class="flex items-center gap-1.5">
      <span
        class="relative inline-block w-3 h-3 rounded-sm border"
        style="background:#f3f4f6;border-color:#d1d5db"
      >
        <!-- sky dot = assigned, not yet reviewed -->
        <span
          class="absolute top-0 right-0 w-1.5 h-1.5 rounded-full -translate-y-px translate-x-px"
          style="background:{MY_DOT_COLOR}"
        ></span>
      </span>
      <span class="text-gray-600 dark:text-gray-300">Assigned</span>
    </span>
  </div>

  <!-- Weights toggle -->
  <button
    type="button"
    onclick={() => (showWeightControls = !showWeightControls)}
    class="flex items-center gap-1 px-2 py-0.5 rounded border transition-colors ml-1
           {showWeightControls
      ? 'bg-sky-100 dark:bg-sky-900/40 border-sky-400 text-sky-700 dark:text-sky-300'
      : 'bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-400'}"
    aria-expanded={showWeightControls}
    title="Adjust priority weights"
  >
    <Icon icon="mdi:tune-variant" class="w-3.5 h-3.5" />
    <span>Weights</span>
  </button>
</div>

<!-- ── Weight controls (inline, shown when toggled) ──────────────────────── -->
{#if showWeightControls}
  <div
    class="flex flex-wrap items-center gap-x-5 gap-y-1 px-3 py-2 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 text-xs"
  >
    <!-- 1st priority slider -->
    <label class="flex items-center gap-2 shrink-0">
      <span class="font-medium text-gray-700 dark:text-gray-300 whitespace-nowrap">1st ×</span>
      <input
        type="range"
        min="0"
        max="5"
        step="0.5"
        bind:value={firstPriorityWeight}
        class="w-24 accent-sky-500"
        aria-label="First priority weight"
      />
      <span class="w-5 text-center font-bold text-sky-700 dark:text-sky-300"
        >{firstPriorityWeight}</span
      >
    </label>

    <!-- 2nd priority slider -->
    <label class="flex items-center gap-2 shrink-0">
      <span class="font-medium text-gray-700 dark:text-gray-300 whitespace-nowrap">2nd ×</span>
      <input
        type="range"
        min="0"
        max="5"
        step="0.5"
        bind:value={secondPriorityWeight}
        class="w-24 accent-sky-500"
        aria-label="Second priority weight"
      />
      <span class="w-5 text-center font-bold text-sky-700 dark:text-sky-300"
        >{secondPriorityWeight}</span
      >
    </label>

    <span class="text-gray-400 dark:text-gray-500 italic whitespace-nowrap">
      score = {firstPriorityWeight}×1st + {secondPriorityWeight}×2nd
    </span>

    <!-- Sort-by-score toggle -->
    <button
      type="button"
      onclick={() => (sortByScore = !sortByScore)}
      class="flex items-center gap-1.5 px-2 py-0.5 rounded border transition-colors ml-auto shrink-0
             {sortByScore
        ? 'bg-amber-100 dark:bg-amber-900/40 border-amber-400 text-amber-700 dark:text-amber-300'
        : 'bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-400'}"
      aria-pressed={sortByScore}
      title="Sort each track's abstracts by descending weighted score"
    >
      <Icon icon={sortByScore ? 'mdi:sort-descending' : 'mdi:sort-variant'} class="w-3.5 h-3.5" />
      <span>Sort by score</span>
    </button>
  </div>
{/if}

<!-- ── Matrix grid ─────────────────────────────────────────────────────────── -->
{#if groupedByTrack.length === 0}
  <div
    class="flex flex-col items-center justify-center py-16 gap-2 text-sm text-gray-500 dark:text-gray-400"
  >
    <Icon icon="mdi:table-off" class="w-10 h-10 opacity-40" />
    No abstract data available
  </div>
{:else}
  <div
    class="overflow-y-auto pr-4"
    style="max-height: calc(100vh - 25rem)"
    role="grid"
    aria-label="Abstract review matrix"
  >
    {#each groupedByTrack as group (group.trackTitle)}
      <!-- Track header -->
      <div
        class="sticky top-0 z-10 flex items-center gap-2 px-3 py-1
               bg-gray-100/90 dark:bg-gray-700/90 border-y border-gray-300 dark:border-gray-600
               backdrop-blur-sm"
      >
        <Icon
          icon="mdi:tag-outline"
          class="w-3.5 h-3.5 text-gray-500 dark:text-gray-400 shrink-0"
        />
        <span class="text-xs font-semibold text-gray-700 dark:text-gray-200 truncate">
          {group.trackTitle}
        </span>
        <span class="ml-auto text-xs text-gray-400 dark:text-gray-500 shrink-0">
          {group.abstracts.length} abstract{group.abstracts.length !== 1 ? 's' : ''}
        </span>
      </div>

      <!-- Cells -->
      <div
        class="grid gap-1 p-2"
        style="grid-template-columns: repeat({effectiveColumns}, minmax(0, 1fr))"
      >
        {#each group.abstracts as abstract (abstract.id)}
          {@const action = consensusAction(abstract)}
          {@const colors = ACTION_COLORS[action] ?? ACTION_COLORS.none}
          {@const score = calcScore(abstract)}
          {@const scoreStr = fmtScore(score)}
          {@const ismine = abstract.is_my_review === true}
          {@const hasMyReview = abstract.my_review != null}
          {@const reviewCount = (abstract.reviews || []).length}
          {@const displayId = abstract.friendly_id ?? abstract.id}

          <button
            type="button"
            onclick={() => openAbstract(abstract)}
            class="review-cell relative flex flex-col items-start justify-start rounded p-0.5
                   leading-tight transition-transform duration-100
                   hover:scale-110 hover:z-20
                   focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-400"
            style="
              min-height: 2rem;
              font-size: 10px;
              font-weight: 700;
              background: {colors.bg};
              border: 1.5px solid {colors.border};
              color: {colors.text};
            "
            aria-label="Abstract #{displayId}: {abstract.title}"
            title="#{displayId} – {abstract.title}
Track: {primaryTrack(abstract)}
Action: {ACTION_STYLES[action]?.label ?? action}
Score: {scoreStr}
Reviews: {reviewCount}{ismine
              ? hasMyReview
                ? '\n✓ My review submitted'
                : '\n⊙ Assigned to me'
              : ''}"
          >
            <!-- Top row: ID (left) + mine dot (right) -->
            <span class="w-full flex items-center justify-between leading-none">
              <span class="text-sm font-medium">#{displayId}</span>
              {#if ismine}
                <span
                  class="w-1.5 h-1.5 rounded-full shrink-0"
                  style="background:{hasMyReview ? DONE_DOT_COLOR : MY_DOT_COLOR}"
                  aria-hidden="true"
                ></span>
              {/if}
            </span>

            <!-- Bottom row: review count (left) + weighted score (right, prominent) -->
            <span class="w-full flex items-end justify-between mt-auto leading-none">
              {#if reviewCount > 0}
                <span class="opacity-50 font-normal text-xs" aria-hidden="true">
                  {reviewCount}
                </span>
              {:else}
                <span></span>
              {/if}

              {#if score !== 0}
                <span
                  class="font-black tabular-nums text-lg font-semibold opacity-95 leading-none"
                  style="color:{colors.text};"
                  aria-hidden="true">{scoreStr}</span
                >
              {/if}
            </span>
          </button>
        {/each}
      </div>
    {/each}
  </div>
{/if}

<!-- ── Dialogs ──────────────────────────────────────────────────────────── -->
{#if selectedAbstract}
  <AbstractReviewsDialog
    bind:open={showReviewsDialog}
    reviews={selectedAbstract.reviews || []}
    abstractTitle={selectedAbstract.title}
  />

  <AbstractReviewFormDialog
    bind:open={showReviewFormDialog}
    abstract={selectedAbstract}
    reviewTrack={selectedAbstract.reviewed_for_tracks?.[0] ?? null}
    onAbstractUpdated={handleAbstractUpdated}
  />
{/if}

<style>
  .review-cell {
    cursor: pointer;
    user-select: none;
  }
</style>
