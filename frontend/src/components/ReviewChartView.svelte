<script>
  import { Tabs, TabItem } from 'flowbite-svelte';
  import BarChart from './BarChart.svelte';
  import DonutChart from './DonutChart.svelte';
  import RatingsBarChart from './RatingsBarChart.svelte';
  import ReviewTimeline from './ReviewTimeline.svelte';

  // Props: expects an array of abstracts with reviews data
  let { abstractData = [] } = $props();

  // Color schemes for different chart types
  const reviewerColors = [
    '#1E40AF',
    '#7C3AED',
    '#DB2777',
    '#DC2626',
    '#EA580C',
    '#D97706',
    '#65A30D',
    '#059669',
    '#0891B2',
    '#0284C7',
  ];

  const trackColors = [
    '#3B82F6',
    '#8B5CF6',
    '#EC4899',
    '#EF4444',
    '#F97316',
    '#F59E0B',
    '#84CC16',
    '#10B981',
  ];

  const actionColors = {
    accept: '#10B981',
    reject: '#EF4444',
    change: '#F59E0B',
    merge: '#8B5CF6',
    mark_as_duplicate: '#6B7280',
  };

  // Collect all reviews from abstracts
  function collectReviews(data) {
    const reviews = [];
    for (const abstract of data || []) {
      if (Array.isArray(abstract.reviews)) {
        for (const review of abstract.reviews) {
          reviews.push({
            ...review,
            abstract_id: abstract.friendly_id,
            abstract_title: abstract.title,
          });
        }
      }
    }
    return reviews;
  }

  // 1. Count reviews per reviewer
  function aggregateByReviewer(reviews) {
    const map = new Map();
    for (const review of reviews) {
      const name = review.user?.full_name || review.user?.email || 'Unknown';
      map.set(name, (map.get(name) || 0) + 1);
    }
    return map;
  }

  // 2. Review stats by track
  function aggregateByTrack(reviews) {
    const map = new Map();
    for (const review of reviews) {
      const trackName = review.track?.title || 'Unknown Track';
      map.set(trackName, (map.get(trackName) || 0) + 1);
    }
    return map;
  }

  // 3. Review actions distribution
  function aggregateByAction(reviews) {
    const map = new Map();
    for (const review of reviews) {
      const action = review.proposed_action || 'no_action';
      map.set(action, (map.get(action) || 0) + 1);
    }
    return map;
  }

  // Helper: convert Map to chart options
  function buildChartOptions(map, colors) {
    const entries = Array.from(map.entries()).sort((a, b) => b[1] - a[1]);
    const labels = entries.map((e) => e[0]);
    const values = entries.map((e) => e[1]);

    return {
      labels,
      series: values,
      colors: colors.slice(0, labels.length),
    };
  }

  // Helper: keep top N, group rest as "Other"
  function topN(m, n = 10) {
    const arr = Array.from(m.entries()).sort((a, b) => b[1] - a[1]);
    if (arr.length <= n) return new Map(arr);
    const top = arr.slice(0, n);
    const rest = arr.slice(n);
    const otherCount = rest.reduce((s, r) => s + r[1], 0);
    const res = new Map(top);
    if (otherCount > 0) {
      res.set('Other', otherCount);
    }
    return res;
  }

  // Build all chart data using derived state
  const chartData = $derived.by(() => {
    const reviews = collectReviews(abstractData || []);

    const reviewerMap = aggregateByReviewer(reviews);
    const trackMap = aggregateByTrack(reviews);
    const actionMap = aggregateByAction(reviews);

    const reviewerTop = topN(reviewerMap, 15);
    const trackTop = topN(trackMap, 10);

    return {
      reviewer: buildChartOptions(reviewerTop, reviewerColors),
      track: buildChartOptions(trackTop, trackColors),
      action: buildChartOptions(actionMap, Object.values(actionColors)),
      reviews, // Pass all reviews for Timeline component
      totalReviews: reviews.length,
      uniqueReviewers: reviewerMap.size,
      uniqueTracks: trackMap.size,
    };
  });

  const reviewerOptions = $derived(chartData.reviewer);
  const trackOptions = $derived(chartData.track);
  const actionOptions = $derived(chartData.action);
  const allReviews = $derived(chartData.reviews);

  const chartHeight = '50vh';
</script>

<div class="p-2 mb-1">
  <!-- Summary stats bar -->
  <div class="grid grid-cols-3 gap-4 mb-1 bg-gray-50 dark:bg-gray-800 rounded-lg">
    <div class="text-center">
      <div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
        {chartData.totalReviews}
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">Total Reviews</div>
    </div>
    <div class="text-center">
      <div class="text-2xl font-bold text-purple-600 dark:text-purple-400">
        {chartData.uniqueReviewers}
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">Reviewers</div>
    </div>
    <div class="text-center">
      <div class="text-2xl font-bold text-green-600 dark:text-green-400">
        {chartData.uniqueTracks}
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-400">Tracks</div>
    </div>
  </div>

  <!-- Tabs for different visualizations -->
  <Tabs tabStyle="underline">
    <TabItem open title="By Reviewer">
      <div class="p-0.5">
        {#if reviewerOptions && reviewerOptions.series && reviewerOptions.series.length}
          <BarChart
            labels={reviewerOptions.labels}
            series={reviewerOptions.series}
            colors={reviewerOptions.colors}
            title="Reviews per Reviewer"
            height={chartHeight}
          />
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No review data available</div>
        {/if}
      </div>
    </TabItem>

    <TabItem title="By Track">
      <div class="p-0.5">
        {#if trackOptions && trackOptions.series && trackOptions.series.length}
          <div class="flex flex-col md:flex-row gap-0.5">
            <div class="w-full md:w-1/2">
              <DonutChart
                labels={trackOptions.labels}
                series={trackOptions.series}
                colors={trackOptions.colors}
                title="Reviews by Track"
                height={chartHeight}
                legendPosition="bottom"
              />
            </div>
            <div class="w-full md:w-1/2">
              <BarChart
                labels={trackOptions.labels}
                series={trackOptions.series}
                colors={trackOptions.colors}
                title=""
                height={chartHeight}
              />
            </div>
          </div>
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No track data available</div>
        {/if}
      </div>
    </TabItem>

    <TabItem title="By Action">
      <div class="p-0.5">
        {#if actionOptions && actionOptions.series && actionOptions.series.length}
          <DonutChart
            labels={actionOptions.labels}
            series={actionOptions.series}
            colors={actionOptions.colors}
            title="Proposed Actions Distribution"
            height={chartHeight}
            legendPosition="bottom"
          />
        {:else}
          <div class="text-sm text-gray-500 text-center py-8">No action data available</div>
        {/if}
      </div>
    </TabItem>

    <TabItem title="Ratings">
      <div class="p-0.5">
        <RatingsBarChart abstractData={abstractData} height={chartHeight} />
      </div>
    </TabItem>

    <TabItem title="Timeline">
      <div class="p-0.5">
        <ReviewTimeline reviews={allReviews} height={chartHeight} />
      </div>
    </TabItem>
  </Tabs>
</div>
