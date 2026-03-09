<script>
  import { onMount, onDestroy } from 'svelte';
  import { Modal } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import { GetAppInfo, OpenSafeURL, GetOSInfo } from '../../wailsjs/go/main/App';
  import AboutTab from './AboutTab.svelte';
  import WindowTab from './WindowTab.svelte';
  import DataSourceTab from './DataSourceTab.svelte';
  import ImportExportTab from './ImportExportTab.svelte';
  import CacheTab from './CacheTab.svelte';
  import AffiliationMapTab from './AffiliationMapTab.svelte';

  let { open = $bindable(false), activeTab = $bindable('about') } = $props();

  let appInfo = $state(null);
  let loading = $state(true);

  // Handle global events requesting the settings modal to open and switch tabs
  function handleOpenSettingsEvent(e) {
    try {
      const tab = e && e.detail && e.detail.tab ? e.detail.tab : 'about';
      activeTab = tab || 'about';
      open = true;
    } catch (err) {
      console.error('open:settings handler error', err);
      open = true;
      activeTab = 'about';
    }
  }

  // Close settings when the setup wizard is opened so the modal doesn't overlap the wizard
  function handleOpenSetupWizardCloseSettings() {
    // If settings modal is open, close it — setup wizard will open via the same event
    open = false;
  }

  onMount(() => {
    window.addEventListener('open:settings', handleOpenSettingsEvent);
    window.addEventListener('open:setup-wizard', handleOpenSetupWizardCloseSettings);
  });

  onDestroy(() => {
    window.removeEventListener('open:settings', handleOpenSettingsEvent);
    window.removeEventListener('open:setup-wizard', handleOpenSetupWizardCloseSettings);
  });

  function setTab(tab) {
    activeTab = tab;
  }

  // Refactored: open the repository's "new issue" page with an empty title so the reporter
  // must provide a meaningful title, and a body that contains only the IDF version, OS,
  // and an issue description placeholder (no other metadata).
  async function reportIssue() {
    if (!appInfo) return;
    try {
      const repo = appInfo.repoURL || appInfo.RepoURL || null;
      if (!repo) {
        console.error('No repository URL available to report an issue');
        return;
      }

      // Normalize repo URL: remove trailing .git and trailing slash
      let repoURL = String(repo)
        .replace(/\.git$/i, '')
        .replace(/\/$/, '');

      // Leave title empty so the reporter provides a meaningful title
      const title = ''; // intentionally empty

      // Ask backend for OS info (reliable host info)
      let os = 'Unknown OS';
      try {
        os = await GetOSInfo();
      } catch (e) {
        console.debug('GetOSInfo failed, falling back to Unknown OS', e);
      }

      // Body: only include IDF version, OS info, and a spot for the reporter to describe the issue
      const bodyLines = [
        `Version: ${appInfo.version || ''}`,
        `OS: ${os}`,
        '',
        'Describe the issue here:',
      ];
      const body = encodeURIComponent(bodyLines.join('\n'));

      let issueURL = repoURL;

      // GitHub: /issues/new?title=...&body=...
      if (/github\.com/i.test(repoURL)) {
        issueURL = `${repoURL}/issues/new?title=${title}&body=${body}`;
      } else if (/gitlab\.com/i.test(repoURL)) {
        // GitLab: /-/issues/new?issue[title]=...&issue[description]=...
        // leave title empty
        issueURL = `${repoURL}/-/issues/new?issue[title]=${title}&issue[description]=${body}`;
      } else {
        // Fallback: open the repo page so user can find issue tracker
        issueURL = repoURL;
      }

      OpenSafeURL(issueURL);
    } catch (err) {
      console.error('reportIssue error', err);
    }
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

<Modal bind:open size="lg" dismissable={true} class="max-w-fit">
  <div
    class="flex justify-between items-center mb-4 border-b border-gray-200 dark:border-gray-700 pb-4"
  >
    <h3 class="text-xl font-semibold text-gray-900 dark:text-white">Settings</h3>
  </div>

  <!-- Tabs Navigation -->
  <div class="flex border-b border-gray-200 dark:border-gray-700 mb-4">
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'about'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('about')}
    >
      <Icon icon="mdi:information" class="w-4 h-4" />
      About
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'window'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('window')}
    >
      <Icon icon="mdi:window-maximize" class="w-4 h-4" />
      Window
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'config'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('config')}
    >
      <Icon icon="mdi:database-cog" class="w-4 h-4" />
      Data Sources
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'importexport'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('importexport')}
    >
      <Icon icon="mdi:swap-horizontal" class="w-4 h-4" />
      Import/Export
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'cache'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('cache')}
    >
      <Icon icon="mdi:database" class="w-4 h-4" />
      Cache
    </button>
    <button
      type="button"
      class="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors {activeTab ===
      'affiliation'
        ? 'text-blue-600 dark:text-blue-500'
        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
      onclick={() => setTab('affiliation')}
    >
      <Icon icon="mdi:office-building" class="w-4 h-4" />
      Affiliation Map
    </button>
  </div>

  <!-- Tab Content -->
  <div class="min-h-75">
    {#if activeTab === 'about'}
      {#if loading}
        <div class="flex items-center justify-center p-8">
          <div class="text-center">
            <div class="animate-spin rounded-full h-12 w-12 border-indigo-500 mx-auto mb-4"></div>
            <p class="text-gray-600 dark:text-gray-400">Loading app info...</p>
          </div>
        </div>
      {:else}
        <AboutTab {appInfo} {reportIssue} />
      {/if}
    {:else if activeTab === 'window'}
      <WindowTab active={true} />
    {:else if activeTab === 'config'}
      <DataSourceTab />
    {:else if activeTab === 'importexport'}
      <ImportExportTab />
    {:else if activeTab === 'cache'}
      <CacheTab />
    {:else if activeTab === 'affiliation'}
      <AffiliationMapTab active={true} />
    {/if}
  </div>
</Modal>
