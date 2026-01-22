<script>
  import { Modal } from 'flowbite-svelte';
  import {
    CloseOutline,
    BuildingOutline,
    GlobeOutline,
    MapPinOutline,
  } from 'flowbite-svelte-icons';

  let { open = $bindable(false), affiliation = null } = $props();

  // Close dialog
  function closeDialog() {
    open = false;
  }

  // Format country with flag emoji (basic implementation)
  function getCountryFlag(countryCode) {
    if (!countryCode) return '';
    const codePoints = countryCode
      .toUpperCase()
      .split('')
      .map((char) => 127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
  }

  // Get full address string
  function getFullAddress() {
    if (!affiliation) return '';
    const parts = [];
    if (affiliation.street) parts.push(affiliation.street);
    if (affiliation.city) parts.push(affiliation.city);
    if (affiliation.postcode) parts.push(affiliation.postcode);
    if (affiliation.country_name) parts.push(affiliation.country_name);
    return parts.join(', ');
  }
</script>

<Modal bind:open size="md" dismissable={false} class="max-w-xl mx-auto">
  <div class="flex justify-between items-start mb-4">
    <div class="flex items-center gap-2">
      <BuildingOutline class="w-6 h-6 text-blue-600 dark:text-blue-400" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Affiliation Details</h3>
    </div>
    <button
      type="button"
      class="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-600 dark:hover:text-white"
      onclick={closeDialog}
    >
      <CloseOutline class="w-5 h-5" />
    </button>
  </div>

  {#if affiliation}
    <div class="space-y-4">
      <!-- Institution Name -->
      <div class="pb-3 border-b border-gray-200 dark:border-gray-700">
        <h4 class="text-xl font-bold text-gray-800 dark:text-white">
          {affiliation.name}
        </h4>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
          ID: {affiliation.id}
        </p>
      </div>

      <!-- Location Information -->
      {#if affiliation.city || affiliation.country_name}
        <div class="flex items-start gap-3">
          <MapPinOutline class="w-5 h-5 text-gray-500 dark:text-gray-400 mt-0.5 shrink-0" />
          <div class="flex-1">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1">Location</p>
            {#if affiliation.city}
              <p class="text-sm text-gray-600 dark:text-gray-400">
                {affiliation.city}{affiliation.postcode ? ` ${affiliation.postcode}` : ''}
              </p>
            {/if}
            {#if affiliation.country_name}
              <p class="text-sm text-gray-600 dark:text-gray-400 flex items-center gap-1">
                {#if affiliation.country_code}
                  <span class="text-lg">{getCountryFlag(affiliation.country_code)}</span>
                {/if}
                {affiliation.country_name}
                {#if affiliation.country_code}
                  <span class="text-xs text-gray-500">({affiliation.country_code})</span>
                {/if}
              </p>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Street Address -->
      {#if affiliation.street}
        <div class="flex items-start gap-3">
          <GlobeOutline class="w-5 h-5 text-gray-500 dark:text-gray-400 mt-0.5 shrink-0" />
          <div class="flex-1">
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1">Address</p>
            <p class="text-sm text-gray-600 dark:text-gray-400">{affiliation.street}</p>
          </div>
        </div>
      {/if}

      <!-- Full Address (if we have components) -->
      {#if getFullAddress()}
        <div class="mt-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
          <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Full Address</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">{getFullAddress()}</p>
        </div>
      {/if}
    </div>
  {:else}
    <div class="text-center py-8 text-gray-500 dark:text-gray-400">
      <p>No affiliation information available</p>
    </div>
  {/if}
</Modal>