<script>
  import { onMount } from 'svelte';
  import { Button } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import { GetStructuredConfigUI } from '../../wailsjs/go/main/App';
  import AffiliationSettings from './AffiliationSettings.svelte';

  let { active = false } = $props();

  /** @type {{ canonical: string; aliases: string[]; enabled: boolean }[]} */
  let mappings = $state([]);
  let loading = $state(false);
  let showAffiliationSettings = $state(false);

  /** @type {Record<string, string>} flat alias→canonical lookup */
  let aliasToCanonical = $state({});

  // Load affiliation mappings from config
  async function loadAffiliationMap() {
    loading = true;
    try {
      const config = await GetStructuredConfigUI();
      const cs = config?.chartSettings || {};

      if (cs.affiliationMap && Array.isArray(cs.affiliationMap)) {
        mappings = cs.affiliationMap.map((m) => ({
          canonical: String(m.canonical || '').trim(),
          aliases: Array.isArray(m.aliases)
            ? m.aliases.map((a) => String(a || '').trim()).filter(Boolean)
            : [],
          enabled: m.enabled !== false,
        }));
      } else {
        mappings = [];
      }
    } catch (err) {
      console.error('Failed to load affiliation mappings:', err);
      mappings = [];
    } finally {
      loading = false;
    }
  }

  // Reload when tab becomes active
  $effect(() => {
    if (active) {
      loadAffiliationMap();
    }
  });

  // Load on mount
  onMount(() => {
    loadAffiliationMap();
  });

  const enabledMappings = $derived(mappings.filter((m) => m.enabled));
  const disabledMappings = $derived(mappings.filter((m) => !m.enabled));

  function openEditor() {
    showAffiliationSettings = true;
  }

  function handleEditorClose() {
    showAffiliationSettings = false;
    loadAffiliationMap();
  }
</script>

<div class="space-y-4">
  <div class="flex items-center justify-between">
    <div>
      <h4 class="text-lg font-semibold text-gray-900 dark:text-white">
        Affiliation Deduplication Map
      </h4>
      <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
        View and manage affiliation name mappings for chart deduplication
      </p>
    </div>
    <Button color="blue" size="sm" onclick={openEditor}>
      <Icon icon="mdi:pencil" class="w-4 h-4 mr-1" />
      Edit Mappings
    </Button>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-8">
      <div class="text-center">
        <Icon icon="mdi:loading" class="w-8 h-8 animate-spin text-blue-500 mx-auto mb-2" />
        <p class="text-gray-600 dark:text-gray-400 text-sm">Loading mappings...</p>
      </div>
    </div>
  {:else if mappings.length === 0}
    <div class="text-center py-8 border border-gray-200 dark:border-gray-700 rounded-lg">
      <Icon icon="mdi:information-outline" class="w-12 h-12 text-gray-400 mx-auto mb-2" />
      <p class="text-gray-600 dark:text-gray-400 mb-3">No affiliation mappings configured yet.</p>
      <Button color="blue" size="sm" onclick={openEditor}>
        <Icon icon="mdi:plus" class="w-4 h-4 mr-1" />
        Add First Mapping
      </Button>
    </div>
  {:else}
    <div class="space-y-2">
      <!-- Enabled Mappings -->
      {#if enabledMappings.length > 0}
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-2">
          <div class="flex items-center gap-1 mb-2">
            <Icon icon="mdi:check-circle" class="w-4 h-4 text-green-600 dark:text-green-400" />
            <h5 class="font-semibold text-gray-900 dark:text-white">
              Active Mappings ({enabledMappings.length})
            </h5>
          </div>
          <div class="space-y-1">
            {#each enabledMappings as mapping}
              <div
                class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg px-2 py-1"
              >
                <div class="flex items-start gap-2">
                  <Icon
                    icon="mdi:arrow-right-circle"
                    class="w-4 h-4 text-blue-600 dark:text-blue-400 mt-1 shrink-0"
                  />
                  <div class="flex-1 min-w-0">
                    <div class="font-semibold text-blue-900 dark:text-blue-100 mb-0.5">
                      {mapping.canonical}
                    </div>
                    {#if mapping.aliases.length > 0}
                      <div class="flex flex-wrap gap-1">
                        {#each mapping.aliases as alias}
                          <span
                            class="inline-flex items-center px-1 py-0.5 rounded text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-blue-300 dark:border-blue-600"
                          >
                            {alias}
                          </span>
                        {/each}
                      </div>
                    {:else}
                      <p class="text-xs text-gray-500 dark:text-gray-400 italic">
                        No aliases defined
                      </p>
                    {/if}
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Disabled Mappings -->
      {#if disabledMappings.length > 0}
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-2">
          <div class="flex items-center gap-1 mb-2">
            <Icon icon="mdi:cancel" class="w-4 h-4 text-gray-400" />
            <h5 class="font-semibold text-gray-600 dark:text-gray-400">
              Disabled Mappings ({disabledMappings.length})
            </h5>
          </div>
          <div class="space-y-1">
            {#each disabledMappings as mapping}
              <div
                class="bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-lg px-2 py-1 opacity-60"
              >
                <div class="flex items-start gap-2">
                  <Icon icon="mdi:arrow-right-circle" class="w-4 h-4 text-gray-400 mt-1 shrink-0" />
                  <div class="flex-1 min-w-0">
                    <div class="font-semibold text-gray-600 dark:text-gray-400 mb-0.5">
                      {mapping.canonical}
                    </div>
                    {#if mapping.aliases.length > 0}
                      <div class="flex flex-wrap gap-1">
                        {#each mapping.aliases as alias}
                          <span
                            class="inline-flex items-center px-1 py-0.5 rounded text-xs bg-white dark:bg-gray-700 text-gray-500 dark:text-gray-500 border border-gray-300 dark:border-gray-600"
                          >
                            {alias}
                          </span>
                        {/each}
                      </div>
                    {:else}
                      <p class="text-xs text-gray-400 dark:text-gray-500 italic">
                        No aliases defined
                      </p>
                    {/if}
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Summary Stats -->
      <div
        class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 border border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-600 dark:text-gray-400">Total Mappings:</span>
          <span class="font-semibold text-gray-900 dark:text-white">{mappings.length}</span>
        </div>
        <div class="flex items-center justify-between text-sm mt-1">
          <span class="text-gray-600 dark:text-gray-400">Total Aliases:</span>
          <span class="font-semibold text-gray-900 dark:text-white">
            {mappings.reduce((sum, m) => sum + m.aliases.length, 0)}
          </span>
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Affiliation Settings Editor Modal -->
{#if showAffiliationSettings}
  <AffiliationSettings
    bind:open={showAffiliationSettings}
    bind:aliasToCanonical
    onSave={handleEditorClose}
  />
{/if}
