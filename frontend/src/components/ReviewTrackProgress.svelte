<script>
  import { onMount, onDestroy } from 'svelte';

  let {
    labels = [],
    reviewedSeries = [],
    remainingSeries = [],
    colors = ['#10B981', '#D1D5DB'],
    title = 'Progress by Track',
    height = '20vh',
  } = $props();

  let container = $state(null);
  let chart = null;
  let echarts = null;

  function buildOption() {
    const hasTitle = typeof title === 'string' && title.trim() !== '';
    const gridTop = hasTitle ? 44 : 20;

    const opt = {
      grid: { left: '12%', right: '6%', top: gridTop, bottom: 40 },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: function (params) {
          // params will be array of series items for that category
          const cat = params && params[0] ? params[0].name : '';
          const rev = params.find((p) => p.seriesName === 'Reviewed')?.value || 0;
          const rem = params.find((p) => p.seriesName === 'Remaining')?.value || 0;
          const tot = rev + rem;
          const pct = tot ? ((rev / tot) * 100).toFixed(1) : '0.0';
          return `${cat}<br/>Reviewed: ${rev}<br/>Remaining: ${rem}<br/>Progress: ${pct}%`;
        },
      },
      xAxis: { type: 'value', axisLabel: { fontFamily: 'Inter, sans-serif' } },
      yAxis: {
        type: 'category',
        data: labels,
        inverse: true,
        axisLabel: { fontFamily: 'Inter, sans-serif' },
      },
      series: [
        {
          name: 'Remaining',
          type: 'bar',
          stack: 'total',
          // make bars thinner: fixed width with a sensible max
          barWidth: 12,
          barMaxWidth: 20,
          data: remainingSeries,
          itemStyle: { color: colors[1] || '#D1D5DB', borderRadius: [6, 6, 6, 6] },
          barCategoryGap: '30%',
          label: { show: false },
        },
        {
          name: 'Reviewed',
          type: 'bar',
          stack: 'total',
          // match width for the stacked reviewed segment
          barWidth: 12,
          barMaxWidth: 20,
          data: reviewedSeries,
          itemStyle: { color: colors[0] || '#10B981', borderRadius: [6, 6, 6, 6] },
          barCategoryGap: '30%',
          label: { show: true, position: 'insideRight', formatter: '{c}' },
        },
      ],
      animation: true,
    };

    if (hasTitle) {
      opt.title = {
        text: title,
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600 },
      };
    }

    return opt;
  }

  function setOptions() {
    if (!chart) return;
    chart.setOption(buildOption(), { notMerge: true });
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
    if (chart) setOptions();
  });
</script>

{#if labels && labels.length}
  <div
    bind:this={container}
    role="img"
    aria-label={title}
    style="height: {height}; width:100%"
  ></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-6">No track data available</div>
{/if}

