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
    accept: {
      bg: '#dcfce7',
      border: '#16a34a',
      text: '#15803d',
      darkBg: '#14532d',
      darkBorder: '#22c55e',
      darkText: '#86efac',
    },
    reject: {
      bg: '#fee2e2',
      border: '#dc2626',
      text: '#b91c1c',
      darkBg: '#450a0a',
      darkBorder: '#ef4444',
      darkText: '#fca5a5',
    },
    change_tracks: {
      bg: '#dbeafe',
      border: '#2563eb',
      text: '#1d4ed8',
      darkBg: '#1e1b4b',
      darkBorder: '#60a5fa',
      darkText: '#93c5fd',
    },
    mark_as_duplicate: {
      bg: '#ffedd5',
      border: '#ea580c',
      text: '#c2410c',
      darkBg: '#431407',
      darkBorder: '#fb923c',
      darkText: '#fdba74',
    },
    merge: {
      bg: '#f3e8ff',
      border: '#9333ea',
      text: '#7e22ce',
      darkBg: '#2e1065',
      darkBorder: '#c084fc',
      darkText: '#d8b4fe',
    },
    none: {
      bg: '#f3f4f6',
      border: '#d1d5db',
      text: '#6b7280',
      darkBg: '#1f2937',
      darkBorder: '#4b5563',
      darkText: '#9ca3af',
    },
  };

  // "Mine" is indicated by a coloured dot only (no outline border)
  const MY_DOT_COLOR = '#0ea5e9'; // sky-500  – assigned but not yet reviewed
  const DONE_DOT_COLOR = '#16a34a'; // green-600 – already reviewed by me

  // ── dialog state ──────────────────────────────────────────────────────────
  let showReviewsDialog = $state(false);
  let showReviewFormDialog = $state(false);
  let selectedAbstract = $state(null);

  // ── controls state ────────────────────────────────────────────────────────
  let showWeightControls = $state(false);
  let sortByScore = $state(false);

  /**
   * Return the proposed action of the most-recently created/modified review.
   * Falls back to majority-vote when timestamps are missing, then 'none'.
   */
  function latestAction(abstract) {
    const reviews = abstract.reviews;
    if (!reviews || reviews.length === 0) return 'none';

    // Sort descending by modified_dt > created_dt
    const sorted = [...reviews].sort((a, b) => {
      const ta = new Date(a.modified_dt ?? a.created_dt ?? 0).getTime();
      const tb = new Date(b.modified_dt ?? b.created_dt ?? 0).getTime();
      return tb - ta;
    });

    const action = sorted[0]?.proposed_action;
    return action && ACTION_COLORS[action] ? action : 'none';
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

  /** Build native-tooltip text for a cell. */
  function cellTooltip(abstract, action, scoreStr, reviewCount, ismine, hasMyReview, displayId) {
    const actionLabel = ACTION_STYLES[action]?.label ?? action;
    const lines = [
      `#${displayId} – ${abstract.title}`,
      `Track: ${primaryTrack(abstract) || '—'}`,
      `Latest action: ${actionLabel}`,
      `Weighted rating: ${scoreStr}`,
      `Reviews: ${reviewCount}`,
    ];
    if (ismine) lines.push(hasMyReview ? '✓ My review submitted' : '⊙ Assigned to me');
    return lines.join('\n');
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
    { key: 'none', label: 'No Review', color: ACTION_COLORS.none },
  ];

  // ── shared button class helpers ───────────────────────────────────────────
  const btnBase =
    'flex items-center gap-1 px-2 py-0.5 rounded border font-medium transition-colors';
  const btnNeutral = `${btnBase} bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:border-gray-400 dark:hover:border-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700`;
  const btnActive = `${btnBase} bg-sky-100 dark:bg-sky-900/50 border-sky-400 dark:border-sky-500 text-sky-700 dark:text-sky-300`;
  const btnSortNeutral = `${btnBase} bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:border-gray-400 dark:hover:border-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700`;
  const btnSortActive = `${btnBase} bg-amber-100 dark:bg-amber-900/40 border-amber-400 dark:border-amber-500 text-amber-700 dark:text-amber-300`;
</script>

<!-- ── Header bar: title + summary + legend + controls toggle ────────────── -->
<div
  class="flex flex-wrap items-center gap-x-4 gap-y-1 px-1 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/60 text-xs select-none"
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
          data-action={item.key}
          style="background:var(--action-bg);border-color:var(--action-border)"
        ></span>
        <span class="text-gray-600 dark:text-gray-300">{item.label}</span>
      </span>
    {/each}
    <!-- Reviewed dot indicator -->
    <span class="flex items-center gap-1.5">
      <span
        class="relative inline-block w-3 h-3 rounded-sm border border-gray-300 dark:border-gray-600 bg-gray-100 dark:bg-gray-800"
      >
        <span
          class="absolute top-0 right-0 w-1.5 h-1.5 rounded-full -translate-y-px translate-x-px"
          style="background:{DONE_DOT_COLOR}"
        ></span>
      </span>
      <span class="text-gray-600 dark:text-gray-300">Reviewed</span>
    </span>
    <!-- Assigned dot indicator -->
    <span class="flex items-center gap-1.5">
      <span
        class="relative inline-block w-3 h-3 rounded-sm border border-gray-300 dark:border-gray-600 bg-gray-100 dark:bg-gray-800"
      >
        <span
          class="absolute top-0 right-0 w-1.5 h-1.5 rounded-full -translate-y-px translate-x-px"
          style="background:{MY_DOT_COLOR}"
        ></span>
      </span>
      <span class="text-gray-600 dark:text-gray-300">Assigned</span>
    </span>
  </div>

  <!-- Weights toggle button -->
  <button
    type="button"
    onclick={() => (showWeightControls = !showWeightControls)}
    class="{showWeightControls ? btnActive : btnNeutral} ml-1"
    aria-expanded={showWeightControls}
    title="Adjust priority weights"
  >
    <Icon icon="mdi:tune-variant" class="w-3.5 h-3.5" />
    <span>Weights</span>
  </button>
</div>

<!-- ── Weight controls row (shown when toggled) ───────────────────────────── -->
{#if showWeightControls}
  <div
    class="flex flex-wrap items-center gap-x-5 gap-y-1.5 px-3 py-2
           bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 text-xs"
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
      weighted rating = {firstPriorityWeight}×1st + {secondPriorityWeight}×2nd
    </span>

    <!-- Sort-by-weighted-rating toggle -->
    <button
      type="button"
      onclick={() => (sortByScore = !sortByScore)}
      class="{sortByScore ? btnSortActive : btnSortNeutral} ml-auto shrink-0"
      aria-pressed={sortByScore}
      title="Sort each track's abstracts by descending weighted rating"
    >
      <Icon icon={sortByScore ? 'mdi:sort-descending' : 'mdi:sort-variant'} class="w-3.5 h-3.5" />
      <span>Sort by rating</span>
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
    class="overflow-y-auto pr-4 px-1"
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
          {@const action = latestAction(abstract)}
          {@const score = calcScore(abstract)}
          {@const scoreStr = fmtScore(score)}
          {@const ismine = abstract.is_my_review === true}
          {@const hasMyReview = abstract.my_review != null}
          {@const reviewCount = (abstract.reviews || []).length}
          {@const displayId = abstract.friendly_id ?? abstract.id}

          <button
            type="button"
            onclick={() => openAbstract(abstract)}
            class="cursor-pointer select-none relative flex flex-col items-start justify-start rounded p-0.5
                   leading-tight transition-transform duration-100
                   hover:scale-110 hover:z-20
                   focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-400"
            data-action={action}
            style="min-height:2rem; font-size:10px; font-weight:700; background:var(--action-bg); border:1.5px solid var(--action-border); color:var(--action-text);"
            aria-label="Abstract #{displayId}: {abstract.title}"
            title={cellTooltip(
              abstract,
              action,
              scoreStr,
              reviewCount,
              ismine,
              hasMyReview,
              displayId,
            )}
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

            <!-- Bottom row: review count (left) + weighted rating (right, prominent) -->
            <span class="w-full flex items-end justify-between mt-auto leading-none">
              {#if reviewCount > 0}
                <span class="opacity-50 font-normal text-xs" aria-hidden="true">{reviewCount}</span>
              {:else}
                <span></span>
              {/if}

              {#if score !== 0}
                <span
                  class="font-black tabular-nums text-lg leading-none opacity-95"
                  style="color:var(--action-text);"
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
  /* Default (light) action colour variables applied by data-action selector */
  [data-action='accept'] {
    --action-bg: #dcfce7;
    --action-border: #16a34a;
    --action-text: #15803d;
  }
  [data-action='reject'] {
    --action-bg: #fee2e2;
    --action-border: #dc2626;
    --action-text: #b91c1c;
  }
  [data-action='change_tracks'] {
    --action-bg: #dbeafe;
    --action-border: #2563eb;
    --action-text: #1d4ed8;
  }
  [data-action='mark_as_duplicate'] {
    --action-bg: #ffedd5;
    --action-border: #ea580c;
    --action-text: #c2410c;
  }
  [data-action='merge'] {
    --action-bg: #f3e8ff;
    --action-border: #9333ea;
    --action-text: #7e22ce;
  }
  [data-action='none'] {
    --action-bg: #f3f4f6;
    --action-border: #d1d5db;
    --action-text: #6b7280;
  }

  :global(.dark) {
    /* Action colour dark overrides */
    /* Accept */
    [data-action='accept'] {
      --action-bg: #14532d;
      --action-border: #22c55e;
      --action-text: #86efac;
    }
    /* Reject */
    [data-action='reject'] {
      --action-bg: #450a0a;
      --action-border: #ef4444;
      --action-text: #fca5a5;
    }
    /* Change Tracks */
    [data-action='change_tracks'] {
      --action-bg: #1e1b4b;
      --action-border: #60a5fa;
      --action-text: #93c5fd;
    }
    /* Mark as Duplicate */
    [data-action='mark_as_duplicate'] {
      --action-bg: #431407;
      --action-border: #fb923c;
      --action-text: #fdba74;
    }
    /* Merge */
    [data-action='merge'] {
      --action-bg: #2e1065;
      --action-border: #c084fc;
      --action-text: #d8b4fe;
    }
    /* None */
    [data-action='none'] {
      --action-bg: #1f2937;
      --action-border: #4b5563;
      --action-text: #9ca3af;
    }

    /* Reviewed/Assigned dots */
    .reviewed-dot {
      background: #16a34a;
    }
    .assigned-dot {
      background: #0ea5e9;
    }
  }
</style>
