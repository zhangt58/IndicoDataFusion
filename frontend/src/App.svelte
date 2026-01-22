<script>
  import { onMount, onDestroy } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import {
    AlignJustifyOutline,
    BookOpenSolid,
    BookSolid,
    CogOutline,
    HomeSolid,
  } from 'flowbite-svelte-icons';
  import './app.css';
  import { goto } from './router/index.js';
  import LocalRouter from './router/Router.svelte';
  import AbstractPage from './pages/AbstractPage.svelte';
  import ContributionPage from './pages/ContributionPage.svelte';
  import EventInfoPage from './pages/EventInfoPage.svelte';
  import Settings from './components/Settings.svelte';
  import TitleBar from './components/TitleBar.svelte';
  import StatusBar from './components/StatusBar.svelte';
  import InitProblems from './components/InitProblems.svelte';

  let Router = LocalRouter;

  const routes = {
    '/': EventInfoPage,
    '/abstract': AbstractPage,
    '/contribution': ContributionPage,
  };

  let isSidebarOpen = $state(false);
  let settingsOpen = $state(false);
  let settingsTab = $state('about');

  // Keep onMount in case other init logic is needed
  onMount(() => {
    // no-op for now
  });

  // navigation helpers
  function navigate(path) {
    try {
      if (typeof goto === 'function') {
        goto(path);
      } else {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
      }
    } catch (e) {
      window.history.pushState({}, '', path);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }
    isSidebarOpen = false;
  }

  function quickGoto(path) {
    navigate(path);
  }

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  function toggleSettings() {
    settingsOpen = !settingsOpen;
  }

  function openSettingsHandler(e) {
    // If the event provides a tab, use it; otherwise default to 'about'
    try {
      settingsTab = e && e.detail && e.detail.tab ? e.detail.tab : 'about';
    } catch (err) {
      settingsTab = 'about';
    }
    settingsOpen = true;
  }

  onMount(() => {
    window.addEventListener('open:settings', openSettingsHandler);
  });

  onDestroy(() => {
    window.removeEventListener('open:settings', openSettingsHandler);
  });
</script>

<!-- Custom Title Bar for frameless window -->
<TitleBar />

<!-- Initialization problems (token/config issues) -->
<InitProblems />

<div class="flex min-h-screen pt-12 overflow-hidden h-full pb-10">
  {#if isSidebarOpen}
    <!-- horizontal quick buttons left of the toggle -->
    <div class="fixed top-12 right-4 z-40 flex items-center gap-2" in:fly={{ x: 20, duration: 300 }}>
      <!-- z-40 so titlebar (z-50) remains on top -->
      <button
        onclick={() => quickGoto('/')}
        class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
        aria-label="Go to Home"
        title="Home"
        type="button"
      >
        <HomeSolid class="text-gray-700 dark:text-gray-200" />
        <span class="hidden sm:inline">Home</span>
      </button>

      <button
        onclick={() => quickGoto('/abstract')}
        class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
        aria-label="Go to Abstracts"
        title="Abstracts"
        type="button"
      >
        <BookSolid class="text-gray-700 dark:text-gray-200" />
        <span class="hidden sm:inline">Abstract</span>
      </button>

      <button
        onclick={() => quickGoto('/contribution')}
        class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
        aria-label="Go to Contributions"
        title="Contributions"
        type="button"
      >
        <BookOpenSolid class="text-gray-700 dark:text-gray-200" />
        <span class="hidden sm:inline">Contrib</span>
      </button>

      <button
        onclick={toggleSettings}
        class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
        aria-label="Settings"
        title="Settings"
        type="button"
      >
        <CogOutline class="text-gray-700 dark:text-gray-200" />
        <span class="hidden sm:inline">Settings</span>
      </button>

      <button
        onclick={toggleSidebar}
        class="flex items-center gap-2 bg-white/60 dark:bg-gray-800/60 border border-gray-200/70 dark:border-gray-700/60 rounded px-3 py-2 shadow-md backdrop-blur-sm hover:bg-white/70 transition-all duration-200 hover:scale-105 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        aria-label="Toggle sidebar"
        title="Toggle sidebar"
        type="button"
      >
        <AlignJustifyOutline class="text-blue-500 dark:text-blue-400" />
        <span class="text-sm font-semibold text-blue-500 dark:text-blue-400 whitespace-nowrap"
          >IndicoDataFusion</span
        >
      </button>
    </div>
  {:else}
    <div class="fixed top-12 right-4 z-40 flex items-center gap-2" in:fade={{ duration: 200 }}>
      <button
        onclick={toggleSidebar}
        class="pointer-events-auto bg-white/60 dark:bg-gray-800/60 border border-gray-200/70 dark:border-gray-700/60 rounded px-3 py-2 shadow-md backdrop-blur-sm hover:bg-white/70 transition-all duration-200 hover:scale-110 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        aria-label="Toggle sidebar"
        title="Toggle sidebar"
        type="button"
      >
        <AlignJustifyOutline class="text-blue-500 dark:text-blue-400" />
      </button>
    </div>
  {/if}

  <div class="overflow-hidden mt-4">
    <div class="w-screen px-4 pb-4">
      <Router {routes} />
    </div>
  </div>
</div>

<!-- Settings Modal -->
<Settings bind:open={settingsOpen} bind:activeTab={settingsTab} />

<!-- Status Bar (bottom) -->
<StatusBar />