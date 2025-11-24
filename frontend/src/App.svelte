<script>
  import { TextAlignJustifyIcon } from '@lucide/svelte';
import './app.css';

  // Runtime wrappers for the generated Wails client (avoids compile-time module resolution issues)
  function GetAbstractData() {
    if (typeof window !== 'undefined' && window.go && window.go.backend && window.go.backend.IndicoClient && window.go.backend.IndicoClient.GetAbstractData) {
      return window.go.backend.IndicoClient.GetAbstractData();
    }
    return Promise.resolve([]);
  }

  function GetContributionData() {
    if (typeof window !== 'undefined' && window.go && window.go.backend && window.go.backend.IndicoClient && window.go.backend.IndicoClient.GetContributionData) {
      return window.go.backend.IndicoClient.GetContributionData();
    }
    return Promise.resolve([]);
  }

  import { BookSolid, BookOpenSolid } from "flowbite-svelte-icons";

  let activeUrl = $state(window.location.pathname);
  let isSidebarOpen = $state(false);

  // Data state
  let abstractData = $state([]);
  let contributionData = $state([]);
  let loading = $state(false);

  // track navigation changes
  $effect(() => {
     const handleNavigation = () => {
         activeUrl = window.location.pathname;
     }
     window.addEventListener('popstate', handleNavigation);
     return () => window.removeEventListener('popstate', handleNavigation);
  });

  // Load data based on current route
  $effect(() => {
    if (activeUrl === '/abstract') {
      loadAbstractData();
    } else if (activeUrl === '/contribution') {
      loadContributionData();
    }
  });

  async function loadAbstractData() {
    loading = true;
    try {
      abstractData = await GetAbstractData();
    } catch (error) {
      console.error('Failed to load abstract data:', error);
    } finally {
      loading = false;
    }
  }

  async function loadContributionData() {
    loading = true;
    try {
      contributionData = await GetContributionData();
    } catch (error) {
      console.error('Failed to load contribution data:', error);
    } finally {
      loading = false;
    }
  }

  function handleNavigation(event) {
    event.preventDefault();
    const href = event.currentTarget.getAttribute('href');
    window.history.pushState({}, '', href);
    activeUrl = href;
  }

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  // Programmatic navigation helper used by the top-right quick buttons
  function navigate(path) {
    window.history.pushState({}, '', path);
    activeUrl = path;
    isSidebarOpen = false;
  }

  // close on Escape
  $effect(() => {
    const onKey = (e) => {
      if (e.key === 'Escape') {
        isSidebarOpen = false;
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  });
</script>

<div class="flex min-h-screen">

  {#if isSidebarOpen}
  <!-- horizontal quick buttons left of the toggle -->
  <div class="fixed top-4 right-18 z-[99998] flex items-center gap-2">
    <button
      onclick={() => navigate('/abstract')}
      class="flex items-center gap-2 bg-white/50 dark:bg-gray-800/50 border border-gray-200/60 dark:border-gray-700/60 rounded px-3 py-2 text-sm shadow-sm hover:bg-white/70 transition"
      aria-label="Go to Abstracts"
      title="Abstracts"
      type="button"
    >
      <BookSolid class="h-4 w-4 text-gray-700 dark:text-gray-200" />
      <span class="hidden sm:inline">Abstract</span>
    </button>

    <button
      onclick={() => navigate('/contribution')}
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
    onclick={toggleSidebar}
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

  {#if loading}
    <div class="flex items-center justify-center p-8">
      <div class="text-lg text-gray-600 dark:text-gray-400">Loading...</div>
    </div>
  {:else if activeUrl === '/abstract'}
    <section class="space-y-4">
      <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-200 mb-4">Abstracts</h2>
      {#each abstractData as abstract}
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
          <h3 class="text-xl font-bold text-gray-800 dark:text-white mb-2">{abstract.title}</h3>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">By: {abstract.author}</p>
          <p class="text-gray-700 dark:text-gray-300 mb-3">{abstract.description}</p>
          <div class="flex gap-2 flex-wrap">
            {#each abstract.keywords as keyword}
              <span class="px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full text-sm">
                {keyword}
              </span>
            {/each}
          </div>
        </div>
      {/each}
    </section>
  {:else if activeUrl === '/contribution'}
    <section class="space-y-4">
      <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-200 mb-4">Contributions</h2>
      {#each contributionData as contribution}
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
          <div class="flex items-start justify-between mb-3">
            <h3 class="text-xl font-bold text-gray-800 dark:text-white">{contribution.title}</h3>
            <span class="px-3 py-1 rounded-full text-sm font-medium {contribution.status === 'Accepted' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : contribution.status === 'Under Review' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
              {contribution.status}
            </span>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">Contributor: {contribution.contributor}</p>
          <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <span>Type: {contribution.type}</span>
            <span>Submitted: {contribution.submittedAt}</span>
          </div>
        </div>
      {/each}
    </section>
  {:else}
    <section class="text-center p-12">
      <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-200 mb-4">Welcome to Indico Data Fusion</h2>
      <p class="text-gray-600 dark:text-gray-400">Select a section from the sidebar to get started.</p>
    </section>
  {/if}
 </div>
</div>

<style>
/* Force the sidebar area and its children to be transparent. Use !important to override utility classes from Flowbite/Tailwind. */
.custom-sidebar,
.custom-sidebar *,
.custom-sidebar *::before,
.custom-sidebar *::after {
  background: transparent !important;
  background-image: none !important;
  border-color: transparent !important;
  box-shadow: none !important;
  -webkit-backdrop-filter: none !important;
  backdrop-filter: none !important;
  outline: none !important;
}

/* Ensure nav/container itself is transparent */
.custom-sidebar {
  background: transparent !important;
}

/* Ensure anchors, list items, and their hover/focus states remain transparent */
.custom-sidebar a,
.custom-sidebar a:visited,
.custom-sidebar a:hover,
.custom-sidebar a:focus,
.custom-sidebar li,
.custom-sidebar li:hover,
.custom-sidebar li:focus {
  background: transparent !important;
  box-shadow: none !important;
}

/* Target any element inside the sidebar that uses Tailwind background utilities like bg-white, bg-opacity, etc. */
.custom-sidebar [class*="bg-"] {
  background: transparent !important;
  background-image: none !important;
}
.custom-sidebar [class*="bg-"]::before,
.custom-sidebar [class*="bg-"]::after {
  background: transparent !important;
  background-image: none !important;
}
.custom-sidebar nav,
.custom-sidebar nav *,
.custom-sidebar ul,
.custom-sidebar li,
.custom-sidebar a,
.custom-sidebar a * {
  background: transparent !important;
  background-image: none !important;
  border: none !important;
  box-shadow: none !important;
  -webkit-backdrop-filter: none !important;
  backdrop-filter: none !important;
}
</style>
