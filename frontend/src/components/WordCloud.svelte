<script>
  import { onMount, onDestroy } from 'svelte';
  import cloud from 'd3-cloud';
  import { GetWordFrequencies } from '../../wailsjs/go/main/App';
  import { Select, Toggle } from 'flowbite-svelte';

  let {
    text = '',
    abstracts = [],
    minLength = 3,
    maxWords = $bindable(100),
    width = 800,
    height = '70vh',
    title = 'Word Cloud',
  } = $props();

  let container = $state(null);
  let svgElement = $state(null);
  let loading = $state(false);
  let error = $state(null);
  let words = $state([]);
  let resizeObserver = null;
  let actualWidth = $state(800);
  let actualHeight = $state(400);

  // plural normalization toggle
  let enablePluralNorm = $state(false);

  let colorScheme = [
    '#1f77b4',
    '#ff7f0e',
    '#2ca02c',
    '#d62728',
    '#9467bd',
    '#8c564b',
    '#e377c2',
    '#7f7f7f',
    '#bcbd22',
    '#17becf',
  ];

  let maxWordsOptions = [
    { value: 100, name: 'Show top 100 Words' },
    { value: 200, name: 'Show top 200 Words' },
    { value: 300, name: 'Show top 300 Words' },
    { value: 400, name: 'Show top 400 Words' },
    { value: 500, name: 'Show top 500 Words' },
  ];

  // Extract text from abstracts array
  const extractedText = $derived(() => {
    if (text && text.trim().length > 0) {
      return text;
    }
    if (abstracts && abstracts.length > 0) {
      return abstracts
        .map((a) => {
          const title = a.title || '';
          const content = a.content || '';
          return `${title} ${content}`;
        })
        .join(' ');
    }
    return '';
  });

  // Fetch word frequencies from backend
  async function fetchWordFrequencies() {
    const inputText = extractedText();
    if (!inputText || inputText.trim().length === 0) {
      words = [];
      return;
    }

    loading = true;
    error = null;
    try {
      // Ensure maxWords is a number (Select emits string values)
      words = await GetWordFrequencies(inputText, minLength, Number(maxWords), enablePluralNorm);
    } catch (err) {
      console.error('Failed to fetch word frequencies:', err);
      error = err.message || 'Failed to fetch word frequencies';
      words = [];
    } finally {
      loading = false;
    }
  }

  // Generate word cloud layout
  function generateWordCloud() {
    if (!svgElement || !words || words.length === 0) return;

    // Clear previous content
    while (svgElement.firstChild) {
      svgElement.removeChild(svgElement.firstChild);
    }

    // Find max count for scaling
    const maxCount = Math.max(...words.map((w) => w.count), 1);

    // Prepare data for d3-cloud
    const cloudWords = words.map((w) => ({
      text: w.word,
      size: Math.max(10, Math.min(80, (w.count / maxCount) * 80)),
      count: w.count,
    }));

    // Create word cloud layout
    const layout = cloud()
      .size([actualWidth, actualHeight])
      .words(cloudWords)
      .padding((d) => Math.max(2, Math.round(d.size * 0.08)))
      .rotate(() => (Math.random() > 0.85 ? Math.round((Math.random() - 0.5) * 90) : 0))
      .spiral('archimedean')
      .font('Inter, sans-serif')
      .fontSize((d) => d.size)
      .on('end', draw);

    layout.start();
  }

  // Draw words on SVG
  function draw(cloudWords) {
    if (!svgElement) return;

    // Create SVG group
    const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
    g.setAttribute('transform', `translate(${actualWidth / 2},${actualHeight / 2})`);

    // Add words
    cloudWords.forEach((word, i) => {
      const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      text.setAttribute('text-anchor', 'middle');
      text.setAttribute('transform', `translate(${word.x},${word.y})rotate(${word.rotate})`);
      text.setAttribute('font-size', word.size);
      text.setAttribute('font-family', 'Inter, sans-serif');
      text.setAttribute('fill', colorScheme[i % colorScheme.length]);
      text.setAttribute('style', 'cursor: default; transition: opacity 0.2s;');
      text.classList.add(
        'select-none',
        'transition',
        'duration-200',
        'hover:scale-110',
        'hover:cursor-pointer',
      );
      text.textContent = word.text;

      // Add hover effect
      text.addEventListener('mouseenter', () => {
        text.setAttribute('opacity', '0.7');
      });
      text.addEventListener('mouseleave', () => {
        text.setAttribute('opacity', '1');
      });

      // Add tooltip
      const titleEl = document.createElementNS('http://www.w3.org/2000/svg', 'title');
      titleEl.textContent = `${word.text}: ${word.count}`;
      text.appendChild(titleEl);

      g.appendChild(text);
    });

    svgElement.appendChild(g);
  }

  // Update dimensions when container resizes
  function updateDimensions() {
    if (container) {
      const rect = container.getBoundingClientRect();
      actualWidth = rect.width || width;
      actualHeight = rect.height || height;
    }
  }

  // Watch for text changes
  $effect(() => {
    // reference reactive inputs so the effect re-runs when they change
    extractedText();
    minLength;
    maxWords;
    enablePluralNorm;
    fetchWordFrequencies();
  });

  // Regenerate word cloud when words or dimensions change
  $effect(() => {
    if (words && words.length > 0 && svgElement) {
      generateWordCloud();
    }
  });

  onMount(() => {
    updateDimensions();

    // Watch for container resize
    if (typeof ResizeObserver !== 'undefined' && container) {
      resizeObserver = new ResizeObserver(() => {
        updateDimensions();
      });
      resizeObserver.observe(container);
    }
  });

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
  });
</script>

<div
  bind:this={container}
  class="flex flex-col items-center justify-center"
  style="width: 100%; height: {height}; min-height: {height};"
>
  {#if title}
    <h3 class="text-center text-lg font-semibold text-gray-900 dark:text-white mb-2">
      {title}
    </h3>
  {/if}

  <div
    class="flex flex-row items-center justify-center w-full gap-10 shadow-sm py-1 px-2 rounded-md"
  >
    <div class="flex-1">
      <Select size="sm" bind:value={maxWords} items={maxWordsOptions} />
    </div>
    <div class="ml-auto">
      <Toggle size="small" bind:checked={enablePluralNorm}>Merge Plural?</Toggle>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center h-full">
      <div class="text-gray-500 dark:text-gray-400">Loading word cloud...</div>
    </div>
  {:else if error}
    <div class="flex items-center justify-center h-full">
      <div class="text-red-500 dark:text-red-400">Error: {error}</div>
    </div>
  {:else if !extractedText() || extractedText().trim().length === 0}
    <div class="flex items-center justify-center h-full">
      <div class="text-gray-500 dark:text-gray-400">No text provided</div>
    </div>
  {:else if words.length === 0}
    <div class="flex items-center justify-center h-full">
      <div class="text-gray-500 dark:text-gray-400">No words to display</div>
    </div>
  {:else}
    <svg bind:this={svgElement} width={actualWidth} height={actualHeight} class="block"> </svg>
  {/if}
</div>
