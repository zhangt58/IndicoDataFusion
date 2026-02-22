<script>
  import { onMount, onDestroy } from 'svelte';

  // Props (Rune-style)
  let {
    abstractData = [],
    title = 'Ratings by Abstract (First & Second Priority)',
    height = '50vh',
    firstPriorityWeight = 1,
    secondPriorityWeight = 1,
  } = $props();

  let container = $state(null);
  let chart = $state(null);
  let echarts = $state(null);

  // Aggregated ratings by abstract friendly ID (using pre-computed fields)
  function getAggRatingsByAbstract(data) {
    const items = [];
    for (const abstract of data || []) {
      const id = abstract.friendly_id;
      const title = abstract.title || '';
      // Use pre-computed first_priority and second_priority fields
      const firstPriority = abstract.first_priority || 0;
      const secondPriority = abstract.second_priority || 0;
      const w1 = firstPriorityWeight ?? 1;
      const w2 = secondPriorityWeight ?? 1;
      const weightedFirst = firstPriority * w1;
      const weightedSecond = secondPriority * w2;
      const weightedTotal = weightedFirst + weightedSecond;

      // Include if either priority exists
      if (firstPriority > 0 || secondPriority > 0) {
        items.push({
          id,
          title,
          firstPriority,
          secondPriority,
          weightedFirst,
          weightedSecond,
          weightedTotal,
          avg:
            (firstPriority + secondPriority) /
            ((firstPriority > 0 ? 1 : 0) + (secondPriority > 0 ? 1 : 0)),
        });
      }
    }
    return items;
  }

  function buildOption() {
    const ratingsData = getAggRatingsByAbstract(abstractData || []);
    if (ratingsData.length === 0) return null;

    // Sort by abstract ID
    ratingsData.sort((a, b) => a.id - b.id);

    const w1 = firstPriorityWeight ?? 1;
    const w2 = secondPriorityWeight ?? 1;
    const isWeighted = w1 !== 1 || w2 !== 1;

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
      ? `Weighted Ratings (w1=${w1}, w2=${w2})`
      : title || 'Ratings by Abstract (First & Second Priority)';

    return {
      title: {
        text: chartTitle,
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600, fontSize: 16 },
      },
      grid: { left: '10%', right: '10%', top: 60, bottom: 80 },
      legend: {
        data: [firstLabel, secondLabel],
        top: 35,
        textStyle: { fontFamily: 'Inter, sans-serif' },
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params) => {
          if (!params || params.length === 0) return '';
          const dataIndex = params[0].dataIndex;
          const item = ratingsData[dataIndex];

          let html = `<strong>Abstract #${item.id}</strong><br/>`;
          html += `<span style="font-size:12px;color:#666">${item.title}</span><br/>`;
          html += `<div style="margin-top:6px">`;
          params.forEach((param) => {
            if (param.value > 0) {
              html += `${param.marker} ${param.seriesName}: <strong>${param.value.toFixed(1)}</strong><br/>`;
            }
          });
          if (isWeighted) {
            html += `<hr style="margin:4px 0;border-color:#eee"/>`;
            html += `Weighted Total: <strong>${item.weightedTotal.toFixed(1)}</strong><br/>`;
          }
          html += `</div>`;
          return html;
        },
      },
      xAxis: {
        type: 'category',
        name: 'Abstract ID',
        nameLocation: 'middle',
        nameGap: 30,
        data: abstractIds,
        axisLabel: { fontFamily: 'Inter, sans-serif', rotate: 45 },
      },
      yAxis: {
        type: 'value',
        name: isWeighted ? 'Weighted Score' : 'Rating Score',
        nameLocation: 'middle',
        nameGap: 50,
        axisLabel: { fontFamily: 'Inter, sans-serif' },
        min: 0,
      },
      dataZoom: [
        { type: 'inside', xAxisIndex: 0 },
        {
          type: 'slider',
          xAxisIndex: 0,
          start: 0,
          end: Math.min(100, (30 / abstractIds.length) * 100),
          bottom: 10,
          height: 20,
        },
      ],
      series: [
        {
          name: firstLabel,
          type: 'bar',
          stack: 'total',
          data: firstPriorityData,
          itemStyle: { color: '#3B82F6' },
          emphasis: { itemStyle: { color: '#2563EB', borderColor: '#1e3a8a', borderWidth: 2 } },
        },
        {
          name: secondLabel,
          type: 'bar',
          stack: 'total',
          data: secondPriorityData,
          itemStyle: { color: '#10B981' },
          emphasis: { itemStyle: { color: '#059669', borderColor: '#064e3b', borderWidth: 0 } },
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

{#if abstractData && abstractData.length}
  <div
    bind:this={container}
    class="echarts-container"
    role="img"
    aria-label={title}
    style="width:100%; height: {height}"
  ></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-8">No rating data available</div>
{/if}
