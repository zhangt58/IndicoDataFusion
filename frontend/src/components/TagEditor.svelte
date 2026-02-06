<script>
  // Reusable tag editor: displays tag chips, remove buttons, input to add tags,
  // and optional suggestion buttons.
  let { tags = [], onAdd = (_t) => {}, onRemove = (_i) => {}, suggestions = [], placeholder = 'Add tag and press Enter' } = $props();

  let input = $state('');

  function add() {
    const t = String(input || '').trim();
    if (!t) return;
    onAdd(t);
    input = '';
  }

  function onKeydown(e) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      add();
    }
  }

  function clickSuggestion(e, s) {
    e.preventDefault();
    e.stopPropagation();
    onAdd(s);
  }

  function remove(idx) {
    onRemove(idx);
  }
</script>

<div class="tag-editor">
  <div class="flex flex-wrap gap-2 mb-2">
    {#if tags}
      {#each tags as tag, ti}
        <span class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-100 dark:bg-gray-800 text-sm rounded-full">
          <span>{tag}</span>
          <button
            type="button"
            class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 ml-1"
            onclick={(e) => { e.preventDefault(); e.stopPropagation(); remove(ti); }}
            aria-label={`Remove tag ${tag}`}
          >
            ✕
          </button>
        </span>
      {/each}
    {/if}
  </div>

  <div class="flex gap-2">
    <input
      type="text"
      bind:value={input}
      onkeydown={(e) => onKeydown(e)}
      placeholder={placeholder}
      class="flex-1 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
      onclick={(e) => { e.stopPropagation(); }}
    />
    <button
      type="button"
      class="px-3 py-2 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-300"
      onclick={(e) => { e.preventDefault(); e.stopPropagation(); add(); }}
    >
      Add
    </button>
  </div>

  {#if suggestions && suggestions.length}
    <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
      Suggested:
      {#each suggestions as s}
        <button
          type="button"
          class="underline ml-1"
          onclick={(e) => clickSuggestion(e, s)}
        >{s}</button>
      {/each}
    </div>
  {/if}
</div>
