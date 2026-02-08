<script>
  let {
    value = '',
    placeholder = 'https://',
    id = 'base-url-input',
    onChange = null,
    suggestions = []
  } = $props();

  const s = $state({
    default: ['https://indico.global'],
    merged: [] }
  );

  // compute merged suggestions (store-backed + provided) deduped and excluding current value
  $effect(() => {
    const fromDefault = s.default || [];
    const fromProp = Array.isArray(suggestions) ? suggestions : [];
    const merged = Array.from(new Set([...fromProp, ...fromDefault]));
    s.merged = merged.filter((h) => String(h || '').trim() !== String(value || '').trim());
  });

  function handleInput(e) {
    value = e.target.value;
    if (typeof onChange === 'function') onChange(value);
  }
</script>

<div>
  <input
    id={id}
    type="text"
    bind:value
    placeholder={placeholder}
    list={id + '-suggestions'}
    oninput={handleInput}
    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm subtle-placeholder"
  />
  <datalist id={id + '-suggestions'}>
    {#each s.merged as host}
      <option value={host}>{host}</option>
    {/each}
  </datalist>
</div>
