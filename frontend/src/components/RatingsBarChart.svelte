<script>
  import { onMount, onDestroy } from 'svelte';
  import Icon from '@iconify/svelte';

  // Props (Rune-style)
  let {
    abstractData = [],
    title = 'Weighted Ratings by Abstract',
    height = '46vh',
    firstPriorityWeight = $bindable(1),
    secondPriorityWeight = $bindable(1),
  } = $props();

  let container = $state(null);
  let chart = $state(null);
  let echarts = $state(null);

  // ── Weight controls UI state ──────────────────────────────────────────────
  let showWeightControls = $state(false);

  // ── dark-mode detection ───────────────────────────────────────────────────
  let isDark = $state(
    typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches,
  );
  $effect(() => {
    if (typeof window === 'undefined') return;
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = (e) => {
      isDark = e.matches;
    };
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  });

  // ── button style helpers ──────────────────────────────────────────────────
  const btnBase =
    'flex items-center gap-1 px-2 py-0.5 rounded border font-medium transition-colors text-xs';
  const btnNeutral = `${btnBase} bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:border-gray-400 dark:hover:border-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700`;
  const btnActive = `${btnBase} bg-sky-100 dark:bg-sky-900/50 border-sky-400 dark:border-sky-500 text-sky-700 dark:text-sky-300`;

  // ── data helpers ──────────────────────────────────────────────────────────

  function getAggRatingsByAbstract(data) {
    const items = [];
    for (const abstract of data || []) {
      const id = abstract.friendly_id;
      const itemTitle = abstract.title || '';
      const firstPriority = abstract.first_priority || 0;
      const secondPriority = abstract.second_priority || 0;
      const w1 = firstPriorityWeight ?? 1;
      const w2 = secondPriorityWeight ?? 1;
      const weightedFirst = firstPriority * w1;
      const weightedSecond = secondPriority * w2;
      const weightedTotal = weightedFirst + weightedSecond;

      if (firstPriority > 0 || secondPriority > 0) {
        items.push({
          id,
          title: itemTitle,
          firstPriority,
          secondPriority,
          weightedFirst,
          weightedSecond,
          weightedTotal,
        });
      }
    }
    return items;
  }

  function buildOption() {
    const ratingsData = getAggRatingsByAbstract(abstractData || []);
    if (ratingsData.length === 0) return null;

    ratingsData.sort((a, b) => a.id - b.id);

    const w1 = firstPriorityWeight ?? 1;
    const w2 = secondPriorityWeight ?? 1;
    const isWeighted = w1 !== 1 || w2 !== 1;

    // echarts theme colours based on dark mode
    const axisColor = isDark ? '#9ca3af' : '#6b7280';
    const splitColor = isDark ? '#374151' : '#f3f4f6';
    const bgTooltip = isDark ? '#1f2937' : '#ffffff';
    const borderColor = isDark ? '#374151' : '#e5e7eb';
    const titleColor = isDark ? '#f9fafb' : '#111827';
    const legendColor = isDark ? '#d1d5db' : '#374151';

    const abstractIds = ratingsData.map((item) => item.id.toString());
    const firstPriorityData = ratingsData.map((item) => ({
      value: item.weightedFirst,
      title: item.title,
      id: item.id,
    }));
    const secondPriorityData = ratingsData.map((item) => ({
      value: item.weightedSecond,
      title: item.title,
      id: item.id,
    }));

    const firstLabel = isWeighted ? `1st Priority ×${w1}` : 'First Priority';
    const secondLabel = isWeighted ? `2nd Priority ×${w2}` : 'Second Priority';
    const chartTitle = isWeighted
      ? `Weighted Ratings  (w1=${w1}, w2=${w2})`
      : title || 'Weighted Ratings by Abstract';

    return {
      backgroundColor: 'transparent',
      grid: { left: '8%', right: '5%', top: 56, bottom: 72 },
      legend: {
        data: [firstLabel, secondLabel],
        top: 26,
        textStyle: { fontFamily: 'Inter, sans-serif', color: legendColor },
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
          shadowStyle: { color: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.04)' },
        },
        backgroundColor: bgTooltip,
        borderColor: borderColor,
        borderWidth: 1,
        padding: [10, 14],
        textStyle: {
          fontFamily: 'Inter, sans-serif',
          fontSize: 12,
          color: isDark ? '#f3f4f6' : '#111827',
        },
        formatter: (params) => {
          if (!params || params.length === 0) return '';
          const dataIndex = params[0].dataIndex;
          const item = ratingsData[dataIndex];
          const titleColor2 = isDark ? '#f9fafb' : '#111827';
          const subColor = isDark ? '#9ca3af' : '#6b7280';
          const divColor = isDark ? '#374151' : '#e5e7eb';

          let html = `<div style="font-weight:700;font-size:13px;color:${titleColor2};margin-bottom:2px">Abstract #${item.id}</div>`;
          html += `<div style="font-size:11px;color:${subColor};margin-bottom:6px;max-width:240px;white-space:normal">${item.title}</div>`;
          params.forEach((param) => {
            if (param.value > 0) {
              html += `<div style="display:flex;align-items:center;gap:6px;margin-bottom:2px">`;
              html += `${param.marker}`;
              html += `<span style="color:${subColor}">${param.seriesName}:</span>`;
              html += `<span style="font-weight:700;color:${titleColor2};margin-left:auto;padding-left:12px">${Number(param.value).toFixed(1)}</span>`;
              html += `</div>`;
            }
          });
          if (isWeighted && item.weightedTotal > 0) {
            html += `<div style="border-top:1px solid ${divColor};margin-top:4px;padding-top:4px;display:flex;justify-content:space-between">`;
            html += `<span style="color:${subColor}">Weighted total</span>`;
            html += `<span style="font-weight:700;color:${titleColor2}">${item.weightedTotal.toFixed(1)}</span>`;
            html += `</div>`;
          }
          // Show raw values when weighted
          if (isWeighted) {
            html += `<div style="border-top:1px solid ${divColor};margin-top:4px;padding-top:4px;font-size:10px;color:${subColor}">`;
            html += `Raw: 1st=${item.firstPriority}, 2nd=${item.secondPriority}`;
            html += `</div>`;
          }
          return html;
        },
      },
      xAxis: {
        type: 'category',
        name: 'Abstract ID',
        nameLocation: 'middle',
        nameGap: 28,
        data: abstractIds,
        nameTextStyle: { color: axisColor, fontFamily: 'Inter, sans-serif' },
        axisLabel: { fontFamily: 'Inter, sans-serif', rotate: 45, color: axisColor },
        axisLine: { lineStyle: { color: axisColor } },
        axisTick: { lineStyle: { color: axisColor } },
      },
      yAxis: {
        type: 'value',
        name: isWeighted ? 'Weighted Rating' : 'Rating',
        nameLocation: 'middle',
        nameGap: 40,
        nameTextStyle: { color: axisColor, fontFamily: 'Inter, sans-serif' },
        axisLabel: { fontFamily: 'Inter, sans-serif', color: axisColor },
        axisLine: { lineStyle: { color: axisColor } },
        splitLine: { lineStyle: { color: splitColor } },
        min: 0,
      },
      dataZoom: [
        { type: 'inside', xAxisIndex: 0 },
        {
          type: 'slider',
          xAxisIndex: 0,
          start: 0,
          end: Math.min(100, (30 / Math.max(1, abstractIds.length)) * 100),
          bottom: 6,
          height: 18,
          fillerColor: isDark ? 'rgba(96,165,250,0.15)' : 'rgba(59,130,246,0.12)',
          borderColor: isDark ? '#4b5563' : '#d1d5db',
          textStyle: { color: axisColor },
          handleStyle: { color: isDark ? '#60a5fa' : '#3b82f6' },
        },
      ],
      series: [
        {
          name: firstLabel,
          type: 'bar',
          stack: 'total',
          data: firstPriorityData,
          itemStyle: { color: '#3B82F6', borderRadius: [0, 0, 0, 0] },
          emphasis: { itemStyle: { color: '#2563EB', borderColor: '#1e3a8a', borderWidth: 1 } },
        },
        {
          name: secondLabel,
          type: 'bar',
          stack: 'total',
          data: secondPriorityData,
          itemStyle: { color: '#10B981', borderRadius: [3, 3, 0, 0] },
          emphasis: { itemStyle: { color: '#059669', borderColor: '#064e3b', borderWidth: 1 } },
        },
      ],
    };
  }

  function setOptions() {
    if (!chart) return;
    const opt = buildOption();
    if (opt) chart.setOption(opt, { notMerge: true });
  }

  function resize() {
    chart?.resize();
  }

  onMount(async () => {
    echarts = await import('echarts');
    if (!container) return;
    chart = echarts.init(container, isDark ? 'dark' : null);
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

  // Re-render whenever any reactive value changes
  $effect(() => {
    // reference all reactive inputs
    void abstractData;
    void firstPriorityWeight;
    void secondPriorityWeight;
    void isDark;
    if (chart) setOptions();
  });
</script>

<!-- ── Controls bar ────────────────────────────────────────────────────────── -->
<div
  class="flex flex-wrap items-center gap-x-4 gap-y-1 px-3 py-1.5
         border-b border-gray-200 dark:border-gray-700
         bg-gray-50 dark:bg-gray-800/60 text-xs select-none"
>
  <span class="font-semibold text-gray-700 dark:text-gray-200 shrink-0">Weighted Ratings</span>

  <span class="text-gray-400 dark:text-gray-500 italic whitespace-nowrap">
    {#if firstPriorityWeight !== 1 || secondPriorityWeight !== 1}
      rating = {firstPriorityWeight}×1st + {secondPriorityWeight}×2nd
    {:else}
      rating = 1st + 2nd priority
    {/if}
  </span>

  <button
    type="button"
    onclick={() => (showWeightControls = !showWeightControls)}
    class="{showWeightControls ? btnActive : btnNeutral} ml-auto"
    aria-expanded={showWeightControls}
    title="Adjust priority weights"
  >
    <Icon icon="mdi:tune-variant" class="w-3.5 h-3.5" />
    <span>Weights</span>
  </button>
</div>

{#if showWeightControls}
  <div
    class="flex flex-wrap items-center gap-x-5 gap-y-1.5 px-3 py-2
           bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 text-xs"
  >
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
  </div>
{/if}

<!-- ── Chart ───────────────────────────────────────────────────────────────── -->
{#if abstractData && abstractData.length}
  <div
    bind:this={container}
    class="echarts-container"
    role="img"
    aria-label={title}
    style="width:100%; height:{height}"
  ></div>
{:else}
  <div class="text-sm text-gray-500 dark:text-gray-400 text-center py-8">
    No rating data available
  </div>
{/if}
