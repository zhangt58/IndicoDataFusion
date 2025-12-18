<script>
  import { onMount, onDestroy } from 'svelte';
  import { WindowGetSize, WindowSetSize } from '../../wailsjs/runtime/runtime.js';

  let { active = false } = $props();

  let winWidth = $state(0);
  let winHeight = $state(0);
  let resizePending = false;

  const widthId = 'win-width-input';
  const heightId = 'win-height-input';

  async function refreshWindowSize() {
    try {
      const winSize = await WindowGetSize();
      winWidth = winSize.w;
      winHeight = winSize.h;
    } catch (e) {
      // ignore failures silently
    }
  }
  function handleResize() {
    if (!active) return;
    if (resizePending) return;
    resizePending = true;
    requestAnimationFrame(async () => {
      await refreshWindowSize();
      resizePending = false;
    });
  }

  onMount(async () => {
    await refreshWindowSize();
    window.addEventListener('resize', handleResize);
  });

  onDestroy(() => {
    window.removeEventListener('resize', handleResize);
  });

  function applyWindowSize() {
    const w = Math.max(400, Number(winWidth) || 0);
    const h = Math.max(300, Number(winHeight) || 0);
    WindowSetSize(w, h);
    winWidth = w;
    winHeight = h;
  }

  function setPreset(w, h) {
    winWidth = w;
    winHeight = h;
    applyWindowSize();
  }
</script>

<div class="space-y-4 p-4">
  <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 items-end">
    <div>
      <label for={widthId} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >Width (px)</label
      >
      <input
        id={widthId}
        type="number"
        bind:value={winWidth}
        min="400"
        class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
    </div>
    <div>
      <label for={heightId} class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >Height (px)</label
      >
      <input
        id={heightId}
        type="number"
        bind:value={winHeight}
        min="300"
        class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
    </div>
    <div class="flex sm:justify-end">
      <button
        type="button"
        class="px-4 py-2 h-[42px] sm:h-[42px] rounded bg-indigo-600 text-white hover:bg-indigo-700 dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400"
        onclick={applyWindowSize}>Apply</button
      >
    </div>
  </div>

  <div class="flex flex-wrap gap-2">
    <button
      type="button"
      class="px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-600"
      onclick={() => setPreset(1024, 768)}>1024×768</button
    >
    <button
      type="button"
      class="px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-600"
      onclick={() => setPreset(1280, 800)}>1280×800</button
    >
    <button
      type="button"
      class="px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-600"
      onclick={() => setPreset(1440, 900)}>1440×900</button
    >
    <button
      type="button"
      class="px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-100 hover:bg-gray-200 dark:hover:bg-gray-600"
      onclick={() => setPreset(1920, 1080)}>1920×1080</button
    >
  </div>
</div>
