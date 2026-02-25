<script>
  import { onMount, onDestroy } from 'svelte';

  // Props
  let {
    labels = [],
    series = [],
    colors = [],
    title = 'Institutes',
    height = '40vh',
    onItemClick = null,
  } = $props();

  let container = $state(null);
  let chart = null;
  let echarts = null;
  let _clickHandler = null;

  function buildOption() {
    const total = (series || []).reduce((a, b) => a + (b || 0), 0);
    const labelCount = (labels || []).length;

    const showDataZoom = labelCount > 15;
    const dataZoom = showDataZoom
      ? [
          {
            type: 'slider',
            yAxisIndex: 0,
            start: 0,
            end: Math.min(100, Math.round((15 / labelCount) * 100)),
          },
        ]
      : [];

    return {
      title: {
        text: title || '',
        left: 'center',
        top: 6,
        textStyle: { fontFamily: 'Inter, sans-serif', fontWeight: 600 },
      },
      grid: { left: '12%', right: '6%', top: 44, bottom: 40 },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: function (params) {
          const p = params && params[0] ? params[0] : null;
          if (!p) return '';
          const name = p.name;
          const value = p.value;
          const pct = total ? ((value / total) * 100).toFixed(1) : '0.0';
          return `${name} — ${value} (${pct}%)`;
        },
      },
      xAxis: {
        type: 'value',
        axisLabel: { fontFamily: 'Inter, sans-serif' },
      },
      yAxis: {
        type: 'category',
        data: labels,
        inverse: true,
        axisLabel: {
          fontFamily: 'Inter, sans-serif',
          formatter: function (v) {
            if (!v) return '';
            return v.length > 50 ? v.slice(0, 47) + '...' : v;
          },
        },
        triggerEvent: !!onItemClick,
      },
      dataZoom,
      series: [
        {
          name: title,
          type: 'bar',
          data: series,
          barCategoryGap: '30%',
          itemStyle: {
            borderRadius: 6,
            color: function () {
              // Use a single color for all bars: prefer the first color from the palette
              if (!colors || colors.length === 0) return undefined;
              return colors[0];
            },
          },
          label: {
            show: false,
            position: 'insideRight',
            formatter: '{c}',
            fontFamily: 'Inter, sans-serif',
          },
        },
      ],
      animation: true,
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
    echarts = await import('echarts');
    if (!container) return;
    chart = echarts.init(container);
    setOptions();
    // register click handler if provided
    if (onItemClick && chart && typeof chart.on === 'function') {
      _clickHandler = function (params) {
        try {
          onItemClick(params);
        } catch (err) {
          console.error('Error in onItemClick handler:', err);
        }
      };
      chart.on('click', _clickHandler);
    }
    window.addEventListener('resize', resize);
  });

  onDestroy(() => {
    window.removeEventListener('resize', resize);
    if (chart) {
      if (_clickHandler && typeof chart.off === 'function') {
        chart.off('click', _clickHandler);
        _clickHandler = null;
      }
    }
    if (chart) {
      chart.dispose();
      chart = null;
    }
  });

  $effect(() => {
    // Explicitly track all props to ensure reactivity
    labels;
    series;
    colors;
    title;
    height;

    if (chart) setOptions();
  });
</script>

{#if series && series.length}
  <div
    bind:this={container}
    role="img"
    aria-label={title}
    style="height: {height}; width: 100%"
  ></div>
{:else}
  <div class="text-sm text-gray-500 text-center py-8">No data available</div>
{/if}
