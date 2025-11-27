<script>
  import { onMount } from 'svelte';
  import { Modal } from 'flowbite-svelte';
  import { InfoCircleSolid } from 'flowbite-svelte-icons';
  import { GetAppInfo } from '../../wailsjs/go/main/App';

  /** @type {boolean} */
  export let open = false;

  let activeTab = 'about';
  let appInfo = null;
  let loading = true;

  onMount(async () => {
    try {
      appInfo = await GetAppInfo();
      loading = false;
    } catch (error) {
      console.error('Failed to load app info:', error);
      loading = false;
    }
  });

  function setTab(tab) {
    activeTab = tab;
  }
</script>

<Modal bind:open={open} size="lg" dismissable={true} class="settings-dialog">
  <div class="flex justify-between items-center mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white">Settings</h3>
  </div>

  <!-- Tabs Navigation -->
  <div class="flex border-b border-gray-200 dark:border-gray-700 mb-4">
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab === 'about' ? 'text-blue-600 dark:text-blue-500 border-b-2 border-blue-600 dark:border-blue-500' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      on:click={() => setTab('about')}
    >
      <InfoCircleSolid class="w-4 h-4" />
      About
    </button>
    <!-- Add more tabs here as needed -->
  </div>

  <!-- Tab Content -->
  <div class="min-h-[300px]">
    {#if activeTab === 'about'}
      {#if loading}
        <div class="flex items-center justify-center p-8">
          <div class="text-center">
            <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500 mx-auto mb-4"></div>
            <p class="text-gray-600 dark:text-gray-400">Loading app info...</p>
          </div>
        </div>
      {:else if appInfo}
        <div class="space-y-6 p-4">
          <!-- App Logo/Header -->
          <div class="text-center">
            <div class="inline-flex items-center justify-center mb-4">
              <img
                src="/src/assets/images/icon.png"
                alt="{appInfo.name} Icon"
                class="w-32 h-32 object-contain"
              />
            </div>
            <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">{appInfo.name}</h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">Aggregating Indico data into one desktop app</p>
          </div>

          <!-- Version Information -->
          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-1">
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
          <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4 space-y-1">
            <div class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Author</span>
              <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.author}</span>
            </div>
            <div class="flex justify-between items-center py-2 border-b border-blue-200 dark:border-blue-800">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Organization</span>
              <span class="text-sm text-gray-600 dark:text-gray-400">{appInfo.company}</span>
            </div>
            <div class="flex justify-between items-center py-2">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Email</span>
              <a href="mailto:{appInfo.authorEmail}" class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
                {appInfo.authorEmail}
              </a>
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
    {/if}
    <!-- Add more tab contents here as needed -->
  </div>
</Modal>

<style>
  :global(.settings-dialog) {
    max-width: 600px;
  }
</style>


