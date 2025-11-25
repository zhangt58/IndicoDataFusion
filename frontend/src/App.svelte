<script>
  import { onMount } from 'svelte';
  import { TextAlignJustifyIcon } from '@lucide/svelte';
  import { BookSolid, BookOpenSolid } from 'flowbite-svelte-icons';
  import './app.css';

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

  // perform any async setup if needed (kept for parity)
  onMount(() => {
    // Nothing required here; the local goto helper is ready synchronously
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
  <div class="fixed top-4 right-18 z-[99998] flex items-center gap-2">
    <button
      on:click={() => quickGoto('/abstract')}
      class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition"
      aria-label="Go to Abstracts"
      title="Abstracts"
      type="button"
    >
      <BookSolid class="h-4 w-4 text-gray-700 dark:text-gray-200" />
      <span class="hidden sm:inline">Abstract</span>
    </button>

    <button
      on:click={() => quickGoto('/contribution')}
      class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition"
      aria-label="Go to Contributions"
      title="Contributions"
      type="button"
    >
      <BookOpenSolid class="h-4 w-4 text-gray-700 dark:text-gray-200" />
      <span class="hidden sm:inline">Contrib</span>
    </button>
  </div>
  {/if}

  <button
    on:click={toggleSidebar}
    class="fixed top-3 right-4 z-[999] pointer-events-auto bg-white/60 dark:bg-gray-800/60 border border-gray-200/70 dark:border-gray-700/60 rounded p-3 shadow-md backdrop-blur-sm hover:bg-white/70 transition focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
    aria-label="Toggle sidebar"
    title="Toggle sidebar"
    type="button"
  >
      <TextAlignJustifyIcon class="h-4 w-4"/>
  </button>



<div class="flex-1 p-6">
  <div class="flex items-center gap-4 mb-6">
    <h1 class="text-3xl font-bold text-gray-800 dark:text-white">Indico Data Fusion</h1>
  </div>

  <!-- router outlet -->
  <svelte:component this={Router} {routes} />

 </div>
</div>

<style>
/* Small local styles kept (if needed add more here) */
</style>
