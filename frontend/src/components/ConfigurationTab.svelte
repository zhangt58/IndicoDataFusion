<script>
  import { onMount } from 'svelte';
  import { GetConfigPath, GetConfigYAML, ApplyConfigYAML } from '../../wailsjs/go/main/App';

  let pathInfo = null;
  let yamlText = '';
  let loading = true;
  let applying = false;
  let applyError = '';
  let applySuccess = '';

  async function loadConfig() {
    try {
      pathInfo = await GetConfigPath();
      yamlText = await GetConfigYAML();
      loading = false;
    } catch (e) {
      loading = false;
      applyError = `Failed to load config: ${e}`;
    }
  }

  onMount(loadConfig);

  async function apply() {
    applyError = '';
    applySuccess = '';
    applying = true;
    try {
      await ApplyConfigYAML(yamlText);
      applySuccess = 'Configuration applied successfully';
    } catch (e) {
      applyError = `Failed to apply configuration: ${e}`;
    } finally {
      applying = false;
    }
  }
</script>

<div class="p-4 space-y-4">
  {#if loading}
    <div class="flex items-center justify-center p-8">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading configuration...</p>
      </div>
    </div>
  {:else}
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
      <div class="flex flex-wrap items-center gap-2 text-sm">
        <span class="font-medium text-gray-700 dark:text-gray-300">Path:</span>
        <span class="text-gray-600 dark:text-gray-400">{pathInfo?.path}</span>
      </div>
      <div class="flex flex-wrap items-center gap-2 text-xs mt-1">
        <span class="text-gray-500 dark:text-gray-400">Source:</span>
        <span class="px-2 py-0.5 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">{pathInfo?.fromEnv ? 'Env (' + pathInfo?.envVarName + ')' : 'Default'}</span>
      </div>
    </div>

    <div>
      <label for="yaml-editor" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Configuration (YAML)</label>
      <textarea id="yaml-editor" bind:value={yamlText} rows="16" class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"></textarea>
    </div>

    {#if applyError}
      <div class="text-sm text-red-600 dark:text-red-400">{applyError}</div>
    {/if}
    {#if applySuccess}
      <div class="text-sm text-green-600 dark:text-green-400">{applySuccess}</div>
    {/if}

    <div class="flex justify-end">
      <button type="button" class="px-4 py-2 rounded bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400" on:click={apply} disabled={applying}>
        {applying ? 'Applying...' : 'Apply'}
      </button>
    </div>
  {/if}
</div>

