<script>
  import { Modal, Button, Input } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';

  /**
   * Props:
   * - open: boolean (bindable) – controls dialog visibility
   * - customExcludedWords: string[] (bindable) – all configured excluded words (persisted)
   * - selectedExcludedWords: string[] (bindable) – currently active/selected words (not persisted)
   */
  let {
    open = $bindable(false),
    customExcludedWords = $bindable([]),
    selectedExcludedWords = $bindable([]),
  } = $props();

  let newExcludedWord = $state('');

  // Check if input word exists in config
  const isExistingWord = $derived(() => {
    const word = newExcludedWord.trim().toLowerCase();
    return word && customExcludedWords.includes(word);
  });

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
    }
    newExcludedWord = '';
  }

  // Populate input with word (for removal workflow)
  function prepareWordForRemoval(word) {
    newExcludedWord = word;
  }

  // Toggle word selection (does not persist to config)
  function toggleWordSelection(word) {
    if (selectedExcludedWords.includes(word)) {
      selectedExcludedWords = selectedExcludedWords.filter((w) => w !== word);
    } else {
      selectedExcludedWords = [...selectedExcludedWords, word];
    }
  }

  function isWordSelected(word) {
    return selectedExcludedWords.includes(word);
  }
</script>

<Modal bind:open size="lg" title="Manage Excluded Words" placement="top-center">
  <div class="space-y-2">
    <p class="text-sm text-gray-600 dark:text-gray-400">
      Add/Remove words to/from your exclusion list.
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
        <span class="text-xs text-gray-500">{customExcludedWords.length} total</span>
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
        <span class="text-xs text-gray-500">{selectedExcludedWords.length} selected</span>
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
