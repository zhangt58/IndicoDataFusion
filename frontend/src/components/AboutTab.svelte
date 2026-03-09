<script>
  import iconImage from '../assets/images/icon.png';
  import CheckForUpdates from './CheckForUpdates.svelte';

  let { appInfo = null, reportIssue } = $props();
</script>

{#if appInfo}
  <div class="space-y-4 p-2">
    <!-- App Logo/Header -->
    <div class="text-center">
      <div class="inline-flex items-center justify-center mb-1">
        <img src={iconImage} alt="{appInfo.name} Icon" class="w-30 h-30 object-contain" />
      </div>
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-1">{appInfo.name}</h2>
      <p class="text-sm text-gray-500 dark:text-gray-400">
        Aggregating Indico data into one desktop app
      </p>
    </div>

    <!-- Version Information -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 space-y-1">
      <div
        class="flex justify-between items-center py-2 border-b border-gray-200 dark:border-gray-700"
      >
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Version</span>
        <div class="flex items-center gap-2">
          <CheckForUpdates />
          <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.version}</span>
        </div>
      </div>
      <div
        class="flex justify-between items-center py-2 border-b border-gray-200 dark:border-gray-700"
      >
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Build Date</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.buildDate}</span>
      </div>
    </div>

    <!-- Author Information -->
    <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-2 space-y-1">
      <div
        class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800"
      >
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Author</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.author}</span>
      </div>
      <div
        class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800"
      >
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Organization</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.organization}</span>
      </div>
      <div class="flex justify-between items-center py-2">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">See an Issue?</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            onclick={reportIssue}
            class="text-sm px-3 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700 dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400"
          >
            Report
          </button>
          <button
            type="button"
            onclick={() => window.dispatchEvent(new CustomEvent('open:setup-wizard'))}
            class="text-sm px-3 py-1 rounded bg-amber-500 text-white hover:bg-amber-600 dark:bg-amber-500 dark:hover:bg-amber-600 focus:outline-none focus:ring-2 focus:ring-amber-400"
            aria-label="Open Setup Wizard"
            title="Open Setup Wizard"
          >
            Wizard
          </button>
        </div>
      </div>
    </div>

    <!-- Copyright & Links -->
    <div class="pt-2 border-t border-gray-200 dark:border-gray-700 text-center space-y-1">
      <p class="text-xs text-gray-500 dark:text-gray-400">
        © {new Date().getFullYear()}
        {appInfo.author}. All rights reserved.
      </p>
    </div>
  </div>
{:else}
  <div class="text-center p-4 text-gray-500">No application information available.</div>
{/if}
