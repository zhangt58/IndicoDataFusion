<script>
  import { Button, Input } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';

  /**
   * Props:
   * - active: boolean - Controls when to load data
   * - customExcludedWords: string[] (optional bindable) - All configured words, if parent wants to control
   * - selectedExcludedWords: string[] (optional bindable) - Currently selected words, if parent wants to control
   *
   * If bindable props are provided, component syncs with parent.
   * If not provided, component manages its own state (standalone mode for Settings tab).
   */
  let {
    active = false,
    customExcludedWords = $bindable([]),
    selectedExcludedWords = $bindable([]),
  } = $props();

  // Internal state when not controlled by parent
  let internalCustomWords = $state([]);
  let internalSelectedWords = $state([]);

  let newExcludedWord = $state('');
  let loading = $state(false);

  // Use parent state if provided, otherwise use internal state
  const words = $derived(
    customExcludedWords.length > 0 || internalCustomWords.length > 0
      ? customExcludedWords.length > 0
        ? customExcludedWords
        : internalCustomWords
      : [],
  );
  const selectedWords = $derived(
    selectedExcludedWords.length > 0 || internalSelectedWords.length > 0
      ? selectedExcludedWords.length > 0
        ? selectedExcludedWords
        : internalSelectedWords
      : [],
  );

  // Check if input word exists in config
  const isExistingWord = $derived(() => {
    const word = newExcludedWord.trim().toLowerCase();
    return word && words.includes(word);
  });

  // Load excluded words from config
  async function loadExcludedWords() {
    loading = true;
    try {
      const config = await GetStructuredConfigUI();
      if (config?.chartSettings?.excludedWords) {
        const loaded = [...config.chartSettings.excludedWords];
        // Update both parent (if bound) and internal state
        customExcludedWords = loaded;
        internalCustomWords = loaded;

        // Initialize selected words
        selectedExcludedWords = [...loaded];
        internalSelectedWords = [...loaded];
      }
    } catch (err) {
      console.error('Failed to load excluded words:', err);
    } finally {
      loading = false;
    }
  }

  // Save to config (only called by Add and Remove buttons)
  async function saveToConfig() {
    try {
      const config = await GetStructuredConfigUI();
      if (!config.chartSettings) {
        // Initialize with empty arrays, affiliationMap will be preserved if it exists
        config.chartSettings = /** @type {any} */ ({ excludedWords: [], affiliationMap: [] });
      }
      config.chartSettings.excludedWords = words;
      await ApplyStructuredConfigUI(config);
    } catch (err) {
      console.error('Failed to save excluded words:', err);
    }
  }

  // Add a word to the configured list (persists to config)
  async function addExcludedWord() {
    const word = newExcludedWord.trim().toLowerCase();
    if (word && !words.includes(word)) {
      const newWords = [...words, word];
      const newSelected = [...selectedWords, word];

      // Update both parent (if bound) and internal state
      customExcludedWords = newWords;
      internalCustomWords = newWords;
      selectedExcludedWords = newSelected;
      internalSelectedWords = newSelected;

      await saveToConfig();
    }
    newExcludedWord = '';
  }

  // Remove a word from the configured list (persists to config)
  async function removeExcludedWord() {
    const word = newExcludedWord.trim().toLowerCase();
    if (word && words.includes(word)) {
      const newWords = words.filter((w) => w !== word);
      const newSelected = selectedWords.filter((w) => w !== word);

      // Update both parent (if bound) and internal state
      customExcludedWords = newWords;
      internalCustomWords = newWords;
      selectedExcludedWords = newSelected;
      internalSelectedWords = newSelected;

      await saveToConfig();
    }
    newExcludedWord = '';
  }

  // Populate input with word (for removal workflow)
  function prepareWordForRemoval(word) {
    newExcludedWord = word;
  }

  // Toggle word selection (does not persist to config)
  function toggleWordSelection(word) {
    const newSelected = selectedWords.includes(word)
      ? selectedWords.filter((w) => w !== word)
      : [...selectedWords, word];

    // Update both parent (if bound) and internal state
    selectedExcludedWords = newSelected;
    internalSelectedWords = newSelected;
  }

  function isWordSelected(word) {
    return selectedWords.includes(word);
  }

  // Load when tab becomes active
  $effect(() => {
    if (active) {
      loadExcludedWords();
    }
  });
</script>

<div class="p-2 space-y-2">
  {#if loading}
    <div class="flex items-center justify-center p-8">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading excluded words...</p>
      </div>
    </div>
  {:else}
    <p class="text-sm text-gray-600 dark:text-gray-400">
      Manage words to exclude from word cloud visualizations.
    </p>

    <!-- Add/Remove word input row -->
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

    <!-- All configured words (click to select/deselect) -->
    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
          All Configured Words (click to select/deselect)
        </h4>
        <span class="text-xs text-gray-500">{words.length} total</span>
      </div>
      <div
        class="border border-gray-200 dark:border-gray-700 rounded-lg p-2 max-h-60 overflow-y-auto bg-gray-50 dark:bg-gray-800"
      >
        {#if words.length === 0}
          <p class="text-sm text-gray-500 text-center py-4">
            No words configured yet. Add some above.
          </p>
        {:else}
          <div class="flex flex-wrap gap-2">
            {#each words as word}
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
        <span class="text-xs text-gray-500">{selectedWords.length} selected</span>
      </div>
      <div
        class="border border-blue-200 dark:border-blue-700 rounded-lg p-2 min-h-16 bg-blue-50 dark:bg-blue-900/20"
      >
        {#if selectedWords.length === 0}
          <p class="text-sm text-gray-500 text-center py-2">
            No words selected. Click words above to select them.
          </p>
        {:else}
          <div class="flex flex-wrap gap-2">
            {#each selectedWords as word}
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
  {/if}
</div>
