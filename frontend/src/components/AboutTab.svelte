<script>
  /** @type {{ name: string, version: string, author: string, company: string, buildDate: string } | null } */
  export let appInfo = null;
  /** @type {(e?: any) => void} */
  export let reportIssue;
</script>

{#if appInfo}
  <div class="space-y-4 p-2">
    <!-- App Logo/Header -->
    <div class="text-center">
      <div class="inline-flex items-center justify-center mb-1">
        <img
          src="/src/assets/images/icon.png"
          alt="{appInfo.name} Icon"
          class="w-30 h-30 object-contain"
        />
      </div>
      <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-1">{appInfo.name}</h2>
      <p class="text-sm text-gray-500 dark:text-gray-400">Aggregating Indico data into one desktop app</p>
    </div>

    <!-- Version Information -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 space-y-1">
      <div class="flex justify-between items-center py-2 border-b border-gray-200 dark:border-gray-700">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Version</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.version}</span>
      </div>
      <div class="flex justify-between items-center py-2 border-b border-gray-200 dark:border-gray-700">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Build Date</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.buildDate}</span>
      </div>
    </div>

    <!-- Author Information -->
    <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-2 space-y-1">
      <div class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Author</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.author}</span>
      </div>
      <div class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Organization</span>
        <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.company}</span>
      </div>
      <div class="flex justify-between items-center py-2">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">See an Issue?</span>
        <button
          type="button"
          on:click={reportIssue}
          class="text-sm px-3 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700 dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400"
        >
          Report
        </button>
      </div>
    </div>

    <!-- Copyright & Links -->
    <div class="pt-2 border-t border-gray-200 dark:border-gray-700 text-center space-y-1">
      <p class="text-xs text-gray-500 dark:text-gray-400">
        © {new Date().getFullYear()} {appInfo.author}. All rights reserved.
      </p>
    </div>
  </div>
{:else}
  <div class="flex items-center justify-center p-8">
    <p class="text-gray-600 dark:text-gray-400">Failed to load app information</p>
  </div>
{/if}
