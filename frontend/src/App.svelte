<script>
  import './app.css';
  import { GetAbstractData, GetContributionData } from '../wailsjs/go/backend/IndicoClient.js';
  // import { GetAbstractData, GetContributionData } from './wails-backend.js';

  import { Sidebar, SidebarGroup, SidebarItem, SidebarButton, uiHelpers } from "flowbite-svelte";
  import { BookSolid, BookOpenSolid } from "flowbite-svelte-icons";

  const sidebarUi = uiHelpers();
  let activeUrl = $state(window.location.pathname);
  let isSidebarOpen = $state(false);
  const closeSidebar = sidebarUi.close;

  // Data state
  let abstractData = $state([]);
  let contributionData = $state([]);
  let loading = $state(false);

  // sync sidebar state
  $effect(() => {
    isSidebarOpen = sidebarUi.isOpen;
  });

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
</script>

<SidebarButton onclick={sidebarUi.toggle} class="mb-2" />
<div class="relative">
  <Sidebar
   {activeUrl}
   backdrop={false}
   isOpen={isSidebarOpen}
   closeSidebar={closeSidebar}
   params={{ x: -50, duration: 50 }}
   class="z-50 h-full"
   position="absolute"
   classes={{ nonactive: "p-2", active: "p-2" }}
  >
    <SidebarGroup>
      <SidebarItem label="Abstract" href="/abstract" active={activeUrl === '/abstract'} onclick={handleNavigation}>
        {#snippet icon()}
          <BookSolid class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white" />
        {/snippet}
      </SidebarItem>
      <SidebarItem label="Contribution" href="/contribution" active={activeUrl === '/contribution'} onclick={handleNavigation}>
        {#snippet icon()}
          <BookOpenSolid class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white" />
        {/snippet}
      </SidebarItem>
    </SidebarGroup>
  </Sidebar>
</div>

<main class="container mx-auto p-4">
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
</main>

<style>
</style>
