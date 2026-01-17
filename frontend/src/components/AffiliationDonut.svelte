<script>
  import { onMount, onDestroy } from 'svelte';

  // Props (Rune-style)
  // `height` accepts any CSS length (e.g. '40vh', '50%', '30rem'). Avoid using absolute pixels here.
  let {
    labels = [],
    series = [],
    colors = [],
    title = '',
    height = '40vh',
    // legendPosition: 'auto' (use heuristics), 'right', 'left', 'bottom', 'top', or 'none'
    legendPosition = 'auto',
  } = $props();

  let container = $state(null);
  let chart = null;
  let echarts = null;

  function buildOption() {
    const data = (labels || []).map((label, i) => ({
      name: label,
      value: series && series[i] ? series[i] : 0,
    }));

    const labelCount = (labels || []).length;

    // Configure legend and pie layout to avoid overlap
    let legendConfig;
    let seriesCenter = ['50%', '50%'];
    let seriesRadius = ['40%', '65%'];

    // If a consumer explicitly sets legendPosition, use that placement.
    // Otherwise fall back to the existing heuristic based on labelCount.
    if (legendPosition && legendPosition !== 'auto') {
      switch (legendPosition) {
        case 'right':
          legendConfig = {
            orient: 'vertical',
            right: '2%',
            top: 'center',
            type: 'scroll',
            textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 11 },
            itemWidth: 10,
            itemHeight: 10,
            itemGap: 6,
            align: 'left',
          };
          seriesCenter = ['30%', '50%'];
          seriesRadius = ['30%', '54%'];
          break;
        case 'left':
          legendConfig = {
            orient: 'vertical',
            left: '2%',
            top: 'center',
            type: 'scroll',
            textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 11 },
            itemWidth: 10,
            itemHeight: 10,
            itemGap: 6,
            align: 'right',
          };
          seriesCenter = ['65%', '50%'];
          seriesRadius = ['30%', '54%'];
          break;
        case 'bottom':
          legendConfig = {
            orient: 'horizontal',
            bottom: 0,
            textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 12 },
          };
          seriesCenter = ['50%', '45%'];
          seriesRadius = ['35%', '60%'];
          break;
        case 'top':
          legendConfig = {
            orient: 'horizontal',
            top: 0,
            textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 12 },
          };
          seriesCenter = ['50%', '55%'];
          seriesRadius = ['35%', '60%'];
          break;
        case 'none':
          legendConfig = { show: false };
          break;
        default:
          // unknown value: fall back to auto heuristics
          break;
      }
    } else {
      // Use more conservative thresholds and more space for the legend on the right
      if (labelCount >= 10) {
        // Very many legend items: vertical scrollable legend on the right, tighter gaps, shift pie further left
        legendConfig = {
          orient: 'vertical',
          right: '2%',
          top: 'center',
          type: 'scroll',
          textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 11 },
          itemWidth: 10,
          itemHeight: 10,
          itemGap: 4,
          align: 'left',
          pageIconSize: 10,
        };
        seriesCenter = ['30%', '50%'];
        seriesRadius = ['30%', '54%'];
      } else if (labelCount >= 8) {
        // Many legend items: vertical scrollable legend on the right and shift pie to the left
        legendConfig = {
          orient: 'vertical',
          right: '4%',
          top: 'center',
          type: 'scroll',
          textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 12 },
          itemWidth: 12,
          itemHeight: 12,
          itemGap: 6,
          align: 'left',
          pageIconSize: 10,
        };
        // shift pie further left and slightly reduce radius so legend has space
        seriesCenter = ['34%', '50%'];
        seriesRadius = ['32%', '56%'];
      } else if (labelCount > 6) {
        // Moderate number: still right-side vertical legend but with less shift
        legendConfig = {
          orient: 'vertical',
          right: 8,
          top: 'center',
          type: 'scroll',
          textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 12 },
          itemWidth: 12,
          itemHeight: 12,
          itemGap: 8,
          align: 'left',
        };
        seriesCenter = ['38%', '50%'];
        seriesRadius = ['35%', '60%'];
      } else {
        // Few items: horizontal legend at the bottom
        legendConfig = {
          orient: 'horizontal',
          bottom: 0,
          textStyle: { fontFamily: 'Inter, sans-serif', fontSize: 12 },
        };
        seriesCenter = ['50%', '50%'];
        seriesRadius = ['40%', '65%'];
      }
    }

    return {
      title: {
        text: title || '',
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600 },
      },
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)',
      },
      color: colors && colors.length ? colors : undefined,
      legend: legendConfig,
      series: [
        {
          name: title,
          type: 'pie',
          center: seriesCenter,
          radius: seriesRadius,
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 4, borderColor: '#fff', borderWidth: 1 },
          label: {
            show: true,
            position: 'outside',
            fontFamily: 'Inter, sans-serif',
            fontSize: 11,
            formatter: function (params) {
              return params.name;
            },
          },
          emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
          data,
        },
      ],
    };
  }

  function setOptions() {
    if (!chart) return;
    const opt = buildOption();
    chart.setOption(opt, { notMerge: true });
  }

  function resize() {
    chart?.resize();
  }

  onMount(async () => {
    // dynamic import to avoid SSR issues
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

  // update chart when props change
  $effect(() => {
    if (chart) setOptions();
  });
</script>

{#if series && series.length}
  <div bind:this={container} class="echarts-container" role="img" aria-label={title} style="width:100%; height: {height}"></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
{/if}

<style>
  :global(.echarts-container) {
    font-family: Inter, system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif;
    /* height is controlled via the `height` prop on the element */
  }
</style>
