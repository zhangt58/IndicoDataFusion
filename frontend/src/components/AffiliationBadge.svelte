<script>
  import { BuildingOutline } from 'flowbite-svelte-icons';

  let { affiliation, onclick, className = '', showCity = true } = $props();

  function handleClick(e) {
    e.stopPropagation();
    if (onclick && affiliation) {
      onclick(affiliation);
    }
  }

</script>

{#if affiliation}
  <button
    type="button"
    class="inline-flex items-center gap-1 text-xs hover:underline cursor-pointer transition-colors {className}"
    onclick={handleClick}
    title="Click to view affiliation details"
  >
    <BuildingOutline class="w-3 h-3" />
    <span class="block text-left">{affiliation.name}</span>
    {#if showCity && (affiliation.city || affiliation.country_code)}
      <span class="text-gray-500 dark:text-gray-400">
        ({affiliation.city || ''}{affiliation.city && affiliation.country_code ? ', ' : ''}{affiliation.country_code || ''})
      </span>
    {/if}
  </button>
{/if}

