<script>
  // Removed detailed chart imports per request — keep timeline
  import ReviewTimeline from './ReviewTimeline.svelte';
  import ReviewTrackProgress from './ReviewTrackProgress.svelte';

  // Props: expects an array of abstracts (same shape as ReviewChartView)
  let { abstractData = [], height = '28vh' } = $props();

  // Helper: aggregate per-track counts for assigned abstracts
  function aggregatePerTrack(data) {
    // Use a Map keyed by track label -> { reviewed, remaining }
    const m = new Map();
    for (const a of data || []) {
      if (!a || !a.is_my_review) continue;
      // Determine track label: prefer first reviewed_for_tracks entry, fallback to accepted or Unknown
      let trackLabel = 'Unknown';
      if (a.reviewed_for_tracks && a.reviewed_for_tracks.length > 0) {
        const t = a.reviewed_for_tracks[0];
        trackLabel = t?.title || t?.code || trackLabel;
      } else if (a.accepted_track) {
        trackLabel = a.accepted_track.title || a.accepted_track.code || trackLabel;
      }

      const hasSubmitted = a.my_review != null;
      const cur = m.get(trackLabel) || { reviewed: 0, remaining: 0 };
      if (hasSubmitted) cur.reviewed += 1;
      else cur.remaining += 1;
      m.set(trackLabel, cur);
    }
    return m;
  }

  const perTrackMap = $derived.by(() => aggregatePerTrack(abstractData || []));

  // Build arrays for TrackProgressChart: labels, reviewedSeries, remainingSeries
  const trackLabels = $derived.by(() => Array.from(perTrackMap.keys()));
  const reviewedSeries = $derived.by(() => Array.from(perTrackMap.values()).map((v) => v.reviewed));
  const remainingSeries = $derived.by(() =>
    Array.from(perTrackMap.values()).map((v) => v.remaining),
  );

  // Collect "my" submitted reviews from abstracts (used for timeline)
  function collectMyReviews(data) {
    const out = [];
    for (const abstract of data || []) {
      if (abstract && abstract.my_review) {
        out.push({
          ...abstract.my_review,
          abstract_id: abstract.friendly_id,
          abstract_title: abstract.title,
        });
      }
    }
    return out;
  }

  // Derived counts: reviewed (submitted), to-be-reviewed (assigned but not submitted), total assigned
  const reviewedCount = $derived.by(() => {
    let c = 0;
    for (const a of abstractData || []) {
      if (a && a.is_my_review === true && a.my_review != null) c += 1;
    }
    return c;
  });

  const toReviewCount = $derived.by(() => {
    let c = 0;
    for (const a of abstractData || []) {
      if (a && a.is_my_review === true && (a.my_review == null || a.my_review === undefined))
        c += 1;
    }
    return c;
  });

  const totalAssigned = $derived.by(() => {
    let c = 0;
    for (const a of abstractData || []) {
      if (a && a.is_my_review === true) c += 1;
    }
    return c;
  });

  const myReviews = $derived.by(() => collectMyReviews(abstractData || []));
</script>

<div class="p-2">
  <div class="grid grid-cols-3 gap-2 mb-2 bg-gray-50 dark:bg-gray-800 rounded-lg p-2 items-center">
    <div class="text-center">
      <div class="text-lg font-semibold text-green-600 dark:text-green-400">{reviewedCount}</div>
      <div class="text-[10px] uppercase text-gray-500 dark:text-gray-400 tracking-wide">
        Reviewed
      </div>
    </div>
    <div class="text-center">
      <div class="text-lg font-semibold text-amber-600 dark:text-amber-400">{toReviewCount}</div>
      <div class="text-[10px] uppercase text-gray-500 dark:text-gray-400 tracking-wide">
        To Review
      </div>
    </div>
    <div class="text-center">
      <div class="text-lg font-semibold text-blue-600 dark:text-blue-400">{totalAssigned}</div>
      <div class="text-[10px] uppercase text-gray-500 dark:text-gray-400 tracking-wide">
        Assigned
      </div>
    </div>
  </div>

  {#if totalAssigned > 0}
    <!-- Show a short message and timeline of submitted reviews -->
    <div class="mb-2">
      <!-- Progress bars per track -->
      <div class="mb-3">
        <ReviewTrackProgress
          labels={trackLabels}
          {reviewedSeries}
          {remainingSeries}
          height={'20vh'}
          colors={['#10B981', '#E5E7EB']}
          title=""
        />
      </div>
      {#if reviewedCount > 0}
        <ReviewTimeline reviews={myReviews} {height} />
      {:else}
        <div class="text-sm text-gray-600 dark:text-gray-400 text-center py-6">
          You have assigned reviews but none submitted yet.
        </div>
      {/if}
    </div>
  {:else}
    <div class="text-sm text-gray-500 text-center py-8">You have no assigned reviews.</div>
  {/if}
</div>
