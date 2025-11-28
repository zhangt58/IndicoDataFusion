<script>
  import { onMount } from 'svelte';
  import { Modal } from 'flowbite-svelte';
  import { InfoCircleSolid, WindowSolid } from 'flowbite-svelte-icons';
  import { GetAppInfo } from '../../wailsjs/go/main/App';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import AboutTab from './AboutTab.svelte';
  import WindowTab from './WindowTab.svelte';
  import ConfigurationTab from './ConfigurationTab.svelte';

  /** @type {boolean} */
  export let open = false;

  let activeTab = 'about';
  let appInfo = null;
  let loading = true;

  function setTab(tab) {
    activeTab = tab;
  }

  function reportIssue() {
    if (!appInfo) return;
    const subject = encodeURIComponent(`${appInfo.name} ${appInfo.version} Issue Report`);
    const body = encodeURIComponent('Please describe the issue here...\n\n');
    const mailto = `mailto:${appInfo.authorEmail}?subject=${subject}&body=${body}`;
    BrowserOpenURL(mailto);
  }

  onMount(async () => {
    try {
      appInfo = await GetAppInfo();
      loading = false;
    } catch (error) {
      console.error('Failed to load app info:', error);
      loading = false;
    }
  });
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
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab === 'window' ? 'text-blue-600 dark:text-blue-500 border-b-2 border-blue-600 dark:border-blue-500' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      on:click={() => setTab('window')}
    >
      <WindowSolid class="w-4 h-4" />
      Window
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab === 'config' ? 'text-blue-600 dark:text-blue-500 border-b-2 border-blue-600 dark:border-blue-500' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      on:click={() => setTab('config')}
    >
      Configuration
    </button>
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
      {:else}
        <AboutTab {appInfo} reportIssue={reportIssue} />
      {/if}
    {:else if activeTab === 'window'}
      <WindowTab active={true} />
    {:else if activeTab === 'config'}
      <ConfigurationTab />
    {/if}
  </div>
</Modal>

<style>
  :global(.settings-dialog) {
    max-width: 600px;
  }
</style>
