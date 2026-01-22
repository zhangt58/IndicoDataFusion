<script>
  import { onMount, onDestroy } from 'svelte';
  import { Toggle } from 'flowbite-svelte';

  // Component: AbstractSubmissionTrend
  // Props:
  // - submittedTimes: Array<string> | Array<{submitted_dt: string}>  -- timestamps or abstract objects
  // - title: string
  // - height: string (css height)
  // - color: string (chart color)

  let {
    submittedTimes = [],
    title = 'Submissions over time',
    height = '40vh',
    color = '#4f46e5',
  } = $props();

  let selectedInterval = $state('week');
  // Toggle: per-bucket (false) or cumulative (true)
  let isCumulative = $state(false);

  let container = $state(null);
  let chart = null;
  let echarts = null;

  function toDate(s) {
    if (!s) return null;
    // Support either raw timestamp strings or objects with submitted_dt
    if (typeof s === 'string') return new Date(s);
    if (typeof s === 'object') {
      if (s.submitted_dt) return new Date(s.submitted_dt);
    }
    return null;
  }

  function startOfDay(d) {
    return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
  }

  function startOfWeek(d) {
    // ISO week starting Monday
    const day = d.getUTCDay(); // 0 (Sun) .. 6
    const diff = (day + 6) % 7; // days since Monday
    const dt = new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
    dt.setUTCDate(dt.getUTCDate() - diff);
    return dt;
  }

  function startOfMonth(d) {
    return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), 1));
  }

  function addInterval(d, i) {
    const dt = new Date(d.getTime());
    if (selectedInterval === 'day') {
      dt.setUTCDate(dt.getUTCDate() + i);
    } else if (selectedInterval === 'week') {
      dt.setUTCDate(dt.getUTCDate() + i * 7);
    } else {
      // month
      dt.setUTCMonth(dt.getUTCMonth() + i);
    }
    return dt;
  }

  function formatLabel(d) {
    if (selectedInterval === 'day') {
      return d.toISOString().slice(0, 10);
    } else if (selectedInterval === 'week') {
      // label as YYYY-MM-DD (week start)
      return d.toISOString().slice(0, 10);
    }
    // month
    return d.toISOString().slice(0, 7);
  }

  function buildBuckets(times) {
    const dates = (times || [])
      .map(toDate)
      .filter(Boolean)
      .map((d) => new Date(d.getTime()));

    if (!dates.length) return { labels: [], series: [] };

    // Normalize to UTC starts
    let startFn = startOfDay;
    if (selectedInterval === 'week') startFn = startOfWeek;
    if (selectedInterval === 'month') startFn = startOfMonth;

    let min = startFn(dates[0]);
    let max = startFn(dates[0]);
    for (const d of dates) {
      const s = startFn(d);
      if (s < min) min = s;
      if (s > max) max = s;
    }

    // Build bucket map
    const buckets = new Map();
    let idx = 0;
    const labels = [];
    let cur = new Date(min.getTime());
    while (cur <= max) {
      const key = cur.toISOString();
      buckets.set(key, 0);
      labels.push(formatLabel(cur));
      idx += 1;
      cur = addInterval(min, idx);
    }

    // Count
    for (const d of dates) {
      const s = startFn(d);
      const key = s.toISOString();
      if (buckets.has(key)) buckets.set(key, buckets.get(key) + 1);
    }

    const series = Array.from(buckets.values());
    return { labels, series };
  }

  function buildOption(labels, series) {
    // Show a visible slider when there are many labels; always enable inside zoom (mouse wheel / drag)
    const showSlider = (labels || []).length > 20;
    const sliderEnd = Math.min(100, Math.round((20 / Math.max(1, (labels || []).length)) * 100));

    // Determine label rotation and provide enough bottom padding so slider sits below labels
    const labelRotate = (labels || []).length > 10 ? 45 : 0;
    // If rotated, increase bottom reserve to accommodate rotated labels and slider; otherwise smaller reserve
    const gridBottom = labelRotate ? 110 : 88;

    const dataZoom = [
      { type: 'inside', xAxisIndex: 0, filterMode: 'filter' },
      ...(showSlider
        ? [
            {
              type: 'slider',
              xAxisIndex: 0,
              start: 0,
              end: sliderEnd,
              handleSize: '100%',
              // position slider below x-axis labels; grid.bottom must leave room
              bottom: 12,
              // smaller height keeps it compact under labels
              height: 28,
            },
          ]
        : []),
    ];

    return {
      title: {
        text: title || '',
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600 },
      },
      toolbox: {
        show: true,
        right: 12,
        feature: {
          restore: { title: 'Reset Zoom' },
          saveAsImage: { title: 'Save as Image' },
        },
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
      },
      xAxis: {
        type: 'category',
        data: labels,
        axisLabel: { fontFamily: 'Inter, sans-serif', rotate: labelRotate },
      },
      yAxis: {
        type: 'value',
        axisLabel: { fontFamily: 'Inter, sans-serif' },
      },
      // Attach the dataZoom array so the slider is rendered (slider positioned via `bottom`)
      dataZoom,
      series: [
        {
          name: isCumulative ? 'Cumulative submissions' : 'Submissions',
          type: isCumulative ? 'line' : 'bar',
          smooth: true,
          data: series,
          lineStyle: { color },
          itemStyle: { color },
          areaStyle: isCumulative ? {} : undefined,
        },
      ],
      animation: true,
      // Ensure chart container has room at bottom for labels+slider
      grid: { left: '6%', right: '6%', top: 60, bottom: gridBottom },
    };
  }

  function setOptions() {
    if (!chart) return;
    const { labels, series } = buildBuckets(submittedTimes);

    // If cumulative mode selected, transform series into cumulative sums
    let finalSeries = series;
    if (isCumulative) {
      const out = [];
      let acc = 0;
      for (const v of series) {
        acc += v || 0;
        out.push(acc);
      }
      finalSeries = out;
    }

    const opt = buildOption(labels, finalSeries);
    chart.setOption(opt, { notMerge: true });
  }

  function resize() {
    chart?.resize();
  }

  onMount(async () => {
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

  $effect(() => {
    // Reference these reactive values so Svelte re-runs this effect when they change
    isCumulative;
    selectedInterval;
    submittedTimes;
    if (chart) setOptions();
  });
</script>

{#if submittedTimes && submittedTimes.length}
  <div class="flex items-center gap-2 mb-2">
    <div class="text-sm text-gray-600 dark:text-gray-300 mr-2">Interval:</div>
    <div class="inline-flex rounded-md border bg-white dark:bg-gray-800">
      <button
        type="button"
        class="px-3 py-1 text-sm rounded-l-md focus:outline-none {selectedInterval === 'day'
          ? 'bg-indigo-600 text-white'
          : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'}"
        onclick={() => {
          selectedInterval = 'day';
          setOptions();
        }}
        aria-pressed={selectedInterval === 'day'}
      >
        Day
      </button>
      <button
        type="button"
        class="px-3 py-1 text-sm focus:outline-none {selectedInterval === 'week'
          ? 'bg-indigo-600 text-white'
          : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'}"
        onclick={() => {
          selectedInterval = 'week';
          setOptions();
        }}
        aria-pressed={selectedInterval === 'week'}
      >
        Week
      </button>
      <button
        type="button"
        class="px-3 py-1 text-sm rounded-r-md focus:outline-none {selectedInterval === 'month'
          ? 'bg-indigo-600 text-white'
          : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'}"
        onclick={() => {
          selectedInterval = 'month';
          setOptions();
        }}
        aria-pressed={selectedInterval === 'month'}
      >
        Month
      </button>
    </div>

    <!-- Single toggle for Cumulative mode using Flowbite Toggle -->
    <div class="ml-4 inline-flex items-center gap-3">
      <label class="flex items-center gap-2">
        <span class="text-sm text-gray-600 dark:text-gray-300">Cumulative</span>
        <Toggle color={'indigo'} bind:checked={isCumulative} aria-label="Toggle cumulative mode" />
      </label>
    </div>
  </div>
  <div
    bind:this={container}
    role="img"
    aria-label={title}
    style="height: {height}; width: 100%"
  ></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
{/if}
