<script>
  import { onMount } from 'svelte';
  import { BookSolid, BookOpenSolid, AlignJustifyOutline } from 'flowbite-svelte-icons';
  import './app.css';

  import { DarkMode } from 'flowbite-svelte';

  import LocalRouter from './router/Router.svelte';
  import { goto } from './router/index.js';
  let Router = LocalRouter;

  import AbstractPage from './pages/AbstractPage.svelte';
  import ContributionPage from './pages/ContributionPage.svelte';

  const routes = {
    '/': AbstractPage,
    '/abstract': AbstractPage,
    '/contribution': ContributionPage
  };

  let isSidebarOpen = false;

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
</script>

<div class="flex min-h-screen">

  {#if isSidebarOpen}
  <!-- horizontal quick buttons left of the toggle -->
  <div class="fixed top-2 right-4 z-50 flex items-center gap-2 animate-slide-in">
    <button
      on:click={() => quickGoto('/abstract')}
      class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
      aria-label="Go to Abstracts"
      title="Abstracts"
      type="button"
    >
      <BookSolid class="text-gray-700 dark:text-gray-200" />
      <span class="hidden sm:inline">Abstract</span>
    </button>

    <button
      on:click={() => quickGoto('/contribution')}
      class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition-all duration-200 hover:scale-105"
      aria-label="Go to Contributions"
      title="Contributions"
      type="button"
    >
      <BookOpenSolid class="text-gray-700 dark:text-gray-200" />
      <span class="hidden sm:inline">Contrib</span>
    </button>
    <!-- Dark mode toggle provided by Flowbite (placed with contrib buttons) -->
    <DarkMode class="flex items-center px-3 py-2 rounded bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 shadow-md backdrop-blur-sm" />

    <button
      on:click={toggleSidebar}
      class="flex items-center gap-2 bg-white/60 dark:bg-gray-800/60 border border-gray-200/70 dark:border-gray-700/60 rounded px-3 py-2 shadow-md backdrop-blur-sm hover:bg-white/70 transition-all duration-200 hover:scale-105 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      aria-label="Toggle sidebar"
      title="Toggle sidebar"
      type="button"
    >
      <AlignJustifyOutline class="text-blue-500 dark:bg-gray-200"/>
      <span class="text-sm font-semibold text-blue-500 dark:text-gray-200 whitespace-nowrap">IndicoDataFusion</span>
    </button>
  </div>
  {:else}
  <div class="fixed top-2 right-4 z-50 flex items-center gap-2 animate-fade-in">
    <button
      on:click={toggleSidebar}
      class="pointer-events-auto bg-white/60 dark:bg-gray-800/60 border border-gray-200/70 dark:border-gray-700/60 rounded px-3 py-2 shadow-md backdrop-blur-sm hover:bg-white/70 transition-all duration-200 hover:scale-110 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      aria-label="Toggle sidebar"
      title="Toggle sidebar"
      type="button"
    >
      <AlignJustifyOutline class="text-blue-500 dark:bg-gray-200"/>
    </button>
  </div>
  {/if}



<div class="flex-1 p-6">
  <svelte:component this={Router} {routes} />
</div>
</div>

<style>
  /* Animation for sliding in from right */
  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateX(20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  /* Animation for fading in */
  @keyframes fade-in {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }

  .animate-fade-in {
    animation: fade-in 0.2s ease-out;
  }
</style>
