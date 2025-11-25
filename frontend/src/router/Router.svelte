<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  export let routes = {};
  export let useHash = false; // allow hash routing for file:// or packaged apps
  const dispatch = createEventDispatcher();

  let current = '/';
  let ready = false;

  function checkRuntime() {
    // Check if Wails runtime is available
    if (typeof window === 'undefined') return false;
    return !!(window['backend'] || window['wailsRuntime'] || window['go'] || window['Wails']);
  }

  function getRouteComponent(path) {
    // simple direct match first, then try parameterized routes (e.g., /abstract/:id)
    if (routes[path]) return routes[path];
    // try matching parameterized routes
    for (const pattern in routes) {
      if (pattern.includes(':')) {
        const regex = new RegExp('^' + pattern.replace(/:[^/]+/g, '([^/]+)') + '$');
        if (regex.test(path)) return routes[pattern];
      }
    }
    return routes['*'] || null;
  }

  function updateCurrentFromLocation() {
    if (typeof window === 'undefined') return;
    current = useHash
      ? (window.location.hash ? window.location.hash.slice(1) : '/')
      : (window.location.pathname || '/');
  }

  function onPop() {
    updateCurrentFromLocation();
    dispatch('route', { path: current });
  }

  let _interval = null;
  let _timeout = null;

  onMount(() => {
    window.addEventListener('popstate', onPop);
    window.addEventListener('hashchange', onPop);

    // Initialize current route
    updateCurrentFromLocation();

    // Check if runtime is ready
    if (checkRuntime()) {
      ready = true;
    } else {
      // Listen for custom ready event
      const onWailsReady = () => {
        ready = true;
        clearInterval(_interval);
        clearTimeout(_timeout);
        window.removeEventListener('wails:ready', onWailsReady);
      };
      window.addEventListener('wails:ready', onWailsReady);

      // Poll for runtime availability
      _interval = setInterval(() => {
        if (checkRuntime()) {
          ready = true;
          clearInterval(_interval);
          clearTimeout(_timeout);
          _interval = null;
        }
      }, 50);

      // Fallback: assume ready after 2 seconds if runtime not detected
      _timeout = setTimeout(() => {
        ready = true;
        clearInterval(_interval);
        _interval = null;
        console.warn('Wails runtime not detected, proceeding anyway');
      }, 2000);
    }
  });

  onDestroy(() => {
    window.removeEventListener('popstate', onPop);
    window.removeEventListener('hashchange', onPop);
    if (_interval) clearInterval(_interval);
    if (_timeout) clearTimeout(_timeout);
  });
</script>

{#if ready}
  <svelte:component this={getRouteComponent(current)} />
{:else}
  <div>Loading application runtime…</div>
{/if}

<style>
/* minimal styles for loading fallback */
</style>
