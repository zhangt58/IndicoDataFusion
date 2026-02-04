<script>
  import { onMount } from 'svelte';
  import {
    WindowMinimise,
    WindowToggleMaximise,
    Quit,
    WindowIsMaximised,
  } from '../../wailsjs/runtime/runtime.js';
  import Icon from '@iconify/svelte';
  import iconImage from '../assets/images/icon.png';

  let isMaximised = $state(false);

  onMount(async () => {
    // Check initial maximised state
    try {
      isMaximised = await WindowIsMaximised();
    } catch (e) {
      console.error('Failed to get window state:', e);
    }
  });

  async function handleMinimise() {
    try {
      WindowMinimise();
    } catch (e) {
      console.error('Failed to minimize window:', e);
    }
  }

  async function handleMaximise() {
    try {
      WindowToggleMaximise();
      // Update state after toggle
      setTimeout(async () => {
        isMaximised = await WindowIsMaximised();
        // Trigger resize event to update table size
        window.dispatchEvent(new Event('resize'));
      }, 100);
    } catch (e) {
      console.error('Failed to toggle maximize:', e);
    }
  }

  function handleClose() {
    try {
      Quit();
    } catch (e) {
      console.error('Failed to quit application:', e);
    }
  }
</script>

<div
  class="select-none fixed top-0 left-0 right-0 rounded-t-xl flex items-center justify-between h-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 shadow-md dark:shadow-black/40"
  style="--wails-draggable: drag; z-index: 9999;"
  role="button"
  tabindex="0"
  aria-label="Application title bar"
  ondblclick={handleMaximise}
  onkeydown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleMaximise();
    }
  }}
>
  <!-- Left: App title/icon -->
  <div class="flex items-center px-3 gap-2">
    <img src={iconImage} alt="App" class="w-6 h-6 rounded-sm" draggable="false" />
    <span class="text-sm font-semibold text-gray-700 dark:text-gray-200 rounded-sm">IDF</span>
  </div>

  <!-- Right: Window controls -->
  <div class="flex items-center h-full" style="--wails-draggable: no-drag">
    <!-- Minimize -->
    <button
      onclick={handleMinimise}
      ondblclick={(e) => {
        e.stopPropagation();
      }}
      class="h-full px-4 hover:bg-gray-200 dark:hover:bg-gray-800 transition-colors flex items-center justify-center group"
      aria-label="Minimize"
      title="Minimize"
      type="button"
    >
      <Icon
        icon="mdi:minus"
        class="w-4 h-4 text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100"
      />
    </button>

    <!-- Maximize/Restore -->
    <button
      onclick={handleMaximise}
      ondblclick={(e) => {
        e.stopPropagation();
      }}
      class="h-full px-4 hover:bg-gray-200 dark:hover:bg-gray-800 transition-colors flex items-center justify-center group"
      aria-label={isMaximised ? 'Restore' : 'Maximize'}
      title={isMaximised ? 'Restore' : 'Maximize'}
      type="button"
    >
      {#if isMaximised}
        <Icon
          icon="mdi:window-restore"
          class="w-4 h-4 text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100"
        />
      {:else}
        <Icon
          icon="mdi:window-maximize"
          class="w-4 h-4 text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100"
        />
      {/if}
    </button>

    <!-- Close -->
    <button
      onclick={handleClose}
      ondblclick={(e) => {
        e.stopPropagation();
      }}
      class="h-full px-4 hover:bg-red-500 dark:hover:bg-red-600 transition-colors flex items-center justify-center group"
      aria-label="Close"
      title="Close"
      type="button"
    >
      <Icon
        icon="mdi:close"
        class="w-4 h-4 text-gray-600 dark:text-gray-400 group-hover:text-white"
      />
    </button>
  </div>
</div>
