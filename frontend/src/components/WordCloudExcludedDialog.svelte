<script>
  import { Modal } from 'flowbite-svelte';
  import WordCloudExclusionTab from './WordCloudExclusionTab.svelte';

  /**
   * Props:
   * - open: boolean (bindable) – controls dialog visibility
   * - customExcludedWords: string[] (bindable) – all configured excluded words (persisted)
   * - selectedExcludedWords: string[] (bindable) – currently active/selected words (not persisted)
   *
   * Note: This component now reuses WordCloudExclusionTab to eliminate code duplication.
   * The tab component manages all state internally and loads data when active.
   */
  let {
    open = $bindable(false),
    customExcludedWords = $bindable([]),
    selectedExcludedWords = $bindable([]),
  } = $props();
</script>

<Modal bind:open size="lg" title="Manage Excluded Words" placement="top-center">
  <WordCloudExclusionTab active={open} />
</Modal>
