<script>
  import { onMount, onDestroy } from 'svelte';
  import cloud from 'd3-cloud';
  import {
    GetWordFrequencies,
    GetStructuredConfigUI,
    ApplyStructuredConfigUI,
  } from '../../wailsjs/go/main/App';
  import { Select, Toggle, Button, Input, Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';

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

  // Custom excluded words management
  let useCustomExcluded = $state(false);
  let customExcludedWords = $state([]); // All words in config
  let selectedExcludedWords = $state([]); // Currently selected/active words
  let showExcludedWordsDialog = $state(false);
  let newExcludedWord = $state('');
  let wordsMgmtDlgPlacement = $state('top-center');

  // Check if input word exists in config
  const isExistingWord = $derived(() => {
    const word = newExcludedWord.trim().toLowerCase();
    return word && customExcludedWords.includes(word);
  });

  // Load custom excluded words from config on mount
  async function loadCustomExcludedWords() {
    try {
      const config = await GetStructuredConfigUI();
      if (config?.chartSettings?.excludedWords) {
        customExcludedWords = config.chartSettings.excludedWords;
        // Initialize all words as selected
        selectedExcludedWords = [...customExcludedWords];
      }
    } catch (err) {
      console.error('Failed to load custom excluded words:', err);
    }
  }

  // Save to config (only called by Add and Remove buttons)
  async function saveToConfig() {
    try {
      const config = await GetStructuredConfigUI();
      if (!config.chartSettings) {
        config.chartSettings = { excludedWords: [], affiliationMap: {} };
      }
      config.chartSettings.excludedWords = customExcludedWords;
      await ApplyStructuredConfigUI(config);
    } catch (err) {
      console.error('Failed to save excluded words:', err);
    }
  }

  // Add a word to the configured list (persists to config)
  async function addExcludedWord() {
    const word = newExcludedWord.trim().toLowerCase();
    if (word && !customExcludedWords.includes(word)) {
      customExcludedWords = [...customExcludedWords, word];
      selectedExcludedWords = [...selectedExcludedWords, word];
      await saveToConfig();
      // Refresh word cloud if custom filter is enabled
      if (useCustomExcluded) {
        fetchWordFrequencies();
      }
    }
    newExcludedWord = '';
  }

  // Remove a word from the configured list (persists to config)
  async function removeExcludedWord() {
    const word = newExcludedWord.trim().toLowerCase();
    if (word && customExcludedWords.includes(word)) {
      customExcludedWords = customExcludedWords.filter((w) => w !== word);
      selectedExcludedWords = selectedExcludedWords.filter((w) => w !== word);
      await saveToConfig();
      // Refresh word cloud if custom filter is enabled
      if (useCustomExcluded) {
        fetchWordFrequencies();
      }
    }
    newExcludedWord = '';
  }

  // Populate input with word for removal (doesn't persist)
  function prepareWordForRemoval(word) {
    newExcludedWord = word;
  }

  // Toggle word selection (doesn't persist to config)
  function toggleWordSelection(word) {
    if (selectedExcludedWords.includes(word)) {
      selectedExcludedWords = selectedExcludedWords.filter((w) => w !== word);
    } else {
      selectedExcludedWords = [...selectedExcludedWords, word];
    }
    // Refresh word cloud if custom filter is enabled
    if (useCustomExcluded) {
      fetchWordFrequencies();
    }
  }

  // Check if a word is selected
  function isWordSelected(word) {
    return selectedExcludedWords.includes(word);
  }

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
      // Use only selected words for filtering
      const excludedList = useCustomExcluded ? selectedExcludedWords : [];
      words = await GetWordFrequencies(
        inputText,
        minLength,
        Number(maxWords),
        enablePluralNorm,
        excludedList,
      );
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
    useCustomExcluded;
    selectedExcludedWords;
    fetchWordFrequencies();
  });

  // Regenerate word cloud when words or dimensions change
  $effect(() => {
    if (words && words.length > 0 && svgElement) {
      generateWordCloud();
    }
  });

  onMount(() => {
    loadCustomExcludedWords();
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
    class="flex flex-row items-center justify-center w-full gap-4 shadow-sm py-1 px-2 rounded-md"
  >
    <div class="flex-1">
      <Select size="sm" bind:value={maxWords} items={maxWordsOptions} />
    </div>
    <div>
      <Toggle size="small" bind:checked={enablePluralNorm}>Merge Plural</Toggle>
    </div>
    <div>
      <Toggle size="small" bind:checked={useCustomExcluded}>Use Custom Filter</Toggle>
    </div>
    <div>
      <Button
        size="xs"
        color="light"
        onclick={() => (showExcludedWordsDialog = true)}
        title="Manage excluded words"
      >
        <Icon icon="mdi:cog" class="w-4 h-4" />
      </Button>
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

<!-- Excluded Words Management Dialog -->
<Modal
  bind:open={showExcludedWordsDialog}
  size="lg"
  title="Manage Excluded Words"
  placement={wordsMgmtDlgPlacement}
>
  <div class="space-y-2">
    <p class="text-sm text-gray-600 dark:text-gray-400">
      Add/Remove words to/from your exclusion list.
    </p>

    <!-- Add/Remove word -->
    <div class="flex gap-2">
      <Input
        bind:value={newExcludedWord}
        placeholder="Enter a word to add or remove..."
        size="sm"
        onkeydown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            if (isExistingWord()) {
              removeExcludedWord();
            } else {
              addExcludedWord();
            }
          }
        }}
        class="flex-1"
      />
      <Button
        size="sm"
        color="blue"
        onclick={addExcludedWord}
        disabled={!newExcludedWord.trim() || isExistingWord()}
      >
        <Icon icon="mdi:plus" class="w-4 h-4 mr-1" />
        Add
      </Button>
      {#if isExistingWord()}
        <Button size="sm" color="red" onclick={removeExcludedWord}>
          <Icon icon="mdi:delete" class="w-4 h-4 mr-1" />
          Remove
        </Button>
      {/if}
    </div>

    <!-- All configured words (clickable to select/deselect) -->
    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
          All Configured Words (click to select/deselect)
        </h4>
        <span class="text-xs text-gray-500">
          {customExcludedWords.length} total
        </span>
      </div>
      <div
        class="border border-gray-200 dark:border-gray-700 rounded-lg p-2 max-h-60 overflow-y-auto bg-gray-50 dark:bg-gray-800"
      >
        {#if customExcludedWords.length === 0}
          <p class="text-sm text-gray-500 text-center py-4">
            No words configured yet. Add some above.
          </p>
        {:else}
          <div class="flex flex-wrap gap-2">
            {#each customExcludedWords as word}
              <div
                class="inline-flex items-center gap-1 rounded-lg px-2 py-1 text-sm transition-all cursor-pointer
                  {isWordSelected(word)
                  ? 'bg-blue-500 text-white hover:bg-blue-600'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'}"
                title={isWordSelected(word) ? 'Click to deselect' : 'Click to select'}
              >
                <button
                  type="button"
                  onclick={() => toggleWordSelection(word)}
                  class="inline-flex items-center gap-1.5"
                >
                  {#if isWordSelected(word)}
                    <Icon icon="mdi:check-circle" class="w-3.5 h-3.5" />
                  {/if}
                  <span>{word}</span>
                </button>
                <button
                  type="button"
                  onclick={(e) => {
                    e.stopPropagation();
                    prepareWordForRemoval(word);
                  }}
                  class="hover:text-red-500 dark:hover:text-red-400"
                  title="Click to populate input for removal"
                >
                  <Icon icon="mdi:close" class="w-3 h-3" />
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Currently selected words (active filters) -->
    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
          Currently Selected for Filtering (click to deselect)
        </h4>
        <span class="text-xs text-gray-500">
          {selectedExcludedWords.length} selected
        </span>
      </div>
      <div
        class="border border-blue-200 dark:border-blue-700 rounded-lg p-2 min-h-16 bg-blue-50 dark:bg-blue-900/20"
      >
        {#if selectedExcludedWords.length === 0}
          <p class="text-sm text-gray-500 text-center py-2">
            No words selected. Click words above to select them.
          </p>
        {:else}
          <div class="flex flex-wrap gap-2">
            {#each selectedExcludedWords as word}
              <button
                type="button"
                onclick={() => toggleWordSelection(word)}
                class="inline-flex items-center gap-1 bg-blue-500 text-white rounded-lg px-2 py-1 text-sm hover:bg-blue-600 transition-colors cursor-pointer"
                title="Click to deselect"
              >
                <Icon icon="mdi:check-circle" class="w-3.5 h-3.5" />
                <span>{word}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
</Modal>
