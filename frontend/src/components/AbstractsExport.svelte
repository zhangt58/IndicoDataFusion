<script>
  import {
    ExportAbstractsToFile,
    SaveAbstractsFileDialog,
    IsTestMode,
    ReviewMode,
    GetRedactionConfig,
    SaveRedactionConfig,
  } from '../../wailsjs/go/main/App';

  // Props (Svelte 5 Rune API)
  let { showToastMsg = (msg, type = 'info') => console.log(`${type}: ${msg}`) } = $props();

  // The export UI state
  let canExportAbstracts = $state(false);
  let exportingAbstracts = $state(false);

  // Redaction config: null until loaded
  let redactionConfig = $state(null);
  let savingRedaction = $state(false);

  // Label/key pairs for the redaction checkboxes
  const REDACTION_FIELDS = [
    { key: 'redactScore', label: 'Score' },
    { key: 'redactJudge', label: 'Judge' },
    { key: 'redactJudgmentComment', label: 'Judgment comment' },
    { key: 'redactJudgmentDT', label: 'Judgment date/time' },
    { key: 'redactSubmitter', label: 'Submitter' },
    { key: 'redactReviews', label: 'Reviews' },
    { key: 'redactComments', label: 'Comments' },
    { key: 'redactCustomFields', label: 'Custom fields' },
    { key: 'redactModifiedBy', label: 'Modified-by person' },
    { key: 'redactFiles', label: 'Files' },
  ];

  async function loadAbstractsExportState() {
    try {
      const [testMode, reviewMode] = await Promise.all([IsTestMode(), ReviewMode()]);
      canExportAbstracts = !testMode && !reviewMode;
    } catch (e) {
      console.error('Failed to determine export eligibility:', e);
      canExportAbstracts = false;
    }

    if (canExportAbstracts) {
      try {
        redactionConfig = await GetRedactionConfig();
      } catch (e) {
        console.error('Failed to load redaction config:', e);
      }
    }
  }

  $effect(() => {
    loadAbstractsExportState();
  });

  async function handleRedactionToggle(key) {
    if (!redactionConfig) return;
    redactionConfig = { ...redactionConfig, [key]: !redactionConfig[key] };
    savingRedaction = true;
    try {
      await SaveRedactionConfig(redactionConfig);
      showToastMsg('Redaction settings saved', 'success');
    } catch (e) {
      console.error('Failed to save redaction config:', e);
      showToastMsg('Failed to save redaction settings', 'error');
    } finally {
      savingRedaction = false;
    }
  }

  async function handleExportAbstracts() {
    exportingAbstracts = true;
    try {
      const path = await SaveAbstractsFileDialog();
      if (!path) {
        // user cancelled
        return;
      }
      await ExportAbstractsToFile(path);
      showToastMsg('Abstracts exported successfully', 'success');
    } catch (e) {
      showToastMsg(`Export failed: ${e}`, 'error');
    } finally {
      exportingAbstracts = false;
    }
  }
</script>

{#if canExportAbstracts}
  <div
    class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow px-4 py-3 border border-gray-200 dark:border-gray-700"
  >
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">
      Export Abstracts Data
    </h3>
    <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">
      Export all abstracts to a JSON file. Use the checkboxes below to control which sensitive
      fields are redacted (removed) before export. This is useful for sharing data with reviewers
      without exposing scores, judgments, or reviewer identities.
    </p>

    <!-- Redaction settings -->
    <div class="mb-3">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
        Fields to redact
        {#if savingRedaction}
          <span class="w-4 h-4 inline-block ml-1 animate-spin text-indigo-500">...</span>
        {/if}
      </h4>
      {#if redactionConfig}
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-x-4 gap-y-1.5">
          {#each REDACTION_FIELDS as field (field.key)}
            <label
              class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer select-none"
            >
              <input
                type="checkbox"
                class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700"
                checked={redactionConfig[field.key]}
                onclick={() => handleRedactionToggle(field.key)}
                disabled={savingRedaction}
              />
              {field.label}
            </label>
          {/each}
        </div>
      {:else}
        <p class="text-sm text-gray-400 dark:text-gray-500 italic">Loading redaction settings…</p>
      {/if}
    </div>

    <button
      type="button"
      class="px-4 py-2 rounded-lg bg-emerald-600 text-white font-medium hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-emerald-500 dark:hover:bg-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-400 focus:ring-offset-2 transition-colors"
      onclick={handleExportAbstracts}
      disabled={exportingAbstracts || !redactionConfig}
      title="Export abstracts data to a JSON file (with redaction applied)"
    >
      {#if exportingAbstracts}
        <span class="w-5 h-5 inline-block mr-2 animate-spin">...</span>
        Exporting…
      {:else}
        <span class="w-5 h-5 inline-block mr-2">⤓</span>
        Export Abstracts
      {/if}
    </button>
  </div>
{/if}
