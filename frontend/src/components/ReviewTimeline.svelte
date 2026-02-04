<script>
  import { onMount, onDestroy } from 'svelte';

  // Props (Rune-style)
  let {
    reviews = [],
    title = 'Review Submissions Over Time',
    height = '50vh',
  } = $props();

  let container = null;
  let chart = null;
  let echarts = null;

  // Build review timeline data
  function buildReviewTimeline(reviewList) {
    const dates = reviewList
      .map((r) => r.created_dt)
      .filter(Boolean)
      .map((d) => new Date(d))
      .sort((a, b) => a - b);

    if (dates.length === 0) return { labels: [], series: [] };

    // Group by day
    const buckets = new Map();
    for (const d of dates) {
      const key = d.toISOString().slice(0, 10);
      buckets.set(key, (buckets.get(key) || 0) + 1);
    }

    const labels = Array.from(buckets.keys()).sort();
    const series = labels.map((k) => buckets.get(k));

    return { labels, series };
  }

  const timelineData = $derived.by(() => {
    return buildReviewTimeline(reviews || []);
  });

  function buildOption() {
    if (!timelineData.labels || timelineData.labels.length === 0) return null;

    const showSlider = timelineData.labels.length > 20;
    const sliderEnd = Math.min(
      100,
      Math.round((20 / Math.max(1, timelineData.labels.length)) * 100)
    );

    return {
      title: {
        text: title,
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600, fontSize: 16 },
      },
      grid: { left: '10%', right: '10%', top: 60, bottom: showSlider ? 80 : 50 },
      tooltip: {
        trigger: 'axis',
        formatter: (params) => {
          const p = params && params[0] ? params[0] : null;
          if (!p) return '';
          return `<strong>${p.name}</strong><br/>Reviews: ${p.value}`;
        },
      },
      xAxis: {
        type: 'category',
        name: 'Date',
        nameLocation: 'middle',
        nameGap: 30,
        data: timelineData.labels,
        axisLabel: {
          fontFamily: 'Inter, sans-serif',
          rotate: 45,
        },
      },
      yAxis: {
        type: 'value',
        name: 'Review Count',
        nameLocation: 'middle',
        nameGap: 40,
        axisLabel: { fontFamily: 'Inter, sans-serif' },
        minInterval: 1,
      },
      dataZoom: [
        { type: 'inside', xAxisIndex: 0 },
        ...(showSlider
          ? [
              {
                type: 'slider',
                xAxisIndex: 0,
                start: 0,
                end: sliderEnd,
                bottom: 10,
                height: 20,
              },
            ]
          : []),
      ],
      series: [
        {
          type: 'line',
          data: timelineData.series,
          smooth: true,
          itemStyle: {
            color: '#4f46e5',
          },
          lineStyle: {
            width: 2,
          },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(79, 70, 229, 0.3)' },
                { offset: 1, color: 'rgba(79, 70, 229, 0.05)' },
              ],
            },
          },
          emphasis: {
            itemStyle: {
              color: '#6366f1',
              borderColor: '#1e1b4b',
              borderWidth: 2,
            },
          },
        },
      ],
    };
  }

  function setOptions() {
    if (!chart) return;
    const opt = buildOption();
    if (opt) {
      chart.setOption(opt, { notMerge: true });
    }
  }

  function resize() {
    chart?.resize();
  }

  onMount(async () => {
    // Dynamic import to avoid SSR issues
    echarts = await import('echarts');
    if (!container) return;
    chart = echarts.init(container);
    setOptions();
    window.addEventListener('resize', resize);
  });

  onDestroy(() => {
    window.removeEventListener('resize', resize);
    if (chart) {
      chart.dispose();
      chart = null;
    }
  });

  // Update chart when props change
  $effect(() => {
    if (chart) setOptions();
  });
</script>

{#if reviews && reviews.length}
  <div
    bind:this={container}
    class="echarts-container"
    role="img"
    aria-label={title}
    style="width:100%; height: {height}"
  ></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-8">No review timeline data available</div>
{/if}
