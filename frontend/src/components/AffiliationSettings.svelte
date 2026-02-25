<script>
  import { onMount } from 'svelte';
  import { Modal, Button, Input, Label } from 'flowbite-svelte';
  import Icon from '@iconify/svelte';
  import { GetStructuredConfigUI, ApplyStructuredConfigUI } from '../../wailsjs/go/main/App';

  let {
    open = $bindable(false),
    /** @type {Record<string, string>} */
    aliasToCanonical = $bindable({}),
    onSave = null,
  } = $props();

  /** @type {{ canonical: string; aliases: string[]; enabled: boolean }[]} */
  let mappings = $state([]);

  // Keep aliasToCanonical in sync with the current (possibly unsaved) mappings state
  $effect(() => {
    /** @type {Record<string, string>} */
    const lookup = {};
    for (const m of mappings) {
      if (!m.enabled || !m.canonical) continue;
      for (const alias of m.aliases || []) {
        if (alias) lookup[alias] = m.canonical;
      }
    }
    aliasToCanonical = lookup;
  });
  let loading = $state(false);
  let saving = $state(false);

  /** @type {string | null} */
  let expandedCanonical = $state(null);

  let newCanonical = $state('');
  let newAlias = $state('');
  let newAliasInput = $state('');

  // Check if canonical already exists
  const isExistingCanonical = $derived(() => {
    const c = newCanonical.trim();
    return c ? mappings.some((m) => m.canonical === c) : false;
  });

  // Load affiliation mappings from config (backwards compatible)
  async function loadAffiliationMap() {
    loading = true;
    try {
      const config = await GetStructuredConfigUI();
      const cs = config?.chartSettings || {};

      if (cs.affiliationMap && typeof cs.affiliationMap === 'object') {
        // Legacy shape: { canonical: [aliases...] }
        const arr = [];
        for (const [canonical, aliases] of Object.entries(cs.affiliationMap)) {
          arr.push({
            canonical: String(canonical || '').trim(),
            aliases: Array.isArray(aliases)
              ? aliases.map((a) => String(a || '').trim()).filter(Boolean)
              : [],
            enabled: true,
          });
        }
        mappings = arr;
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

  // Save affiliation mappings to config
  async function saveAffiliationMap() {
    saving = true;
    try {
      const cleaned = mappings.map((m) => ({
        canonical: String(m.canonical || '').trim(),
        aliases: (m.aliases || []).map((a) => String(a || '').trim()).filter(Boolean),
        enabled: !!m.enabled,
      }));

      const withCanonical = cleaned.filter((m) => m.canonical !== '');

      // Check for alias conflicts across enabled mappings
      /** @type {Record<string, string>} */
      const seenAlias = {};
      for (const m of withCanonical) {
        if (!m.enabled) continue;
        for (const a of m.aliases) {
          if (seenAlias[a] && seenAlias[a] !== m.canonical) {
            alert(`Alias "${a}" is assigned to both "${seenAlias[a]}" and "${m.canonical}".`);
            saving = false;
            return;
          }
          seenAlias[a] = m.canonical;
        }
      }

      const config = await GetStructuredConfigUI();
      if (!config.chartSettings) {
        config.chartSettings = /** @type {any} */ ({});
      }
      /** @type {any} */
      const cs = config.chartSettings;
      cs.affiliationMappings = withCanonical;

      await ApplyStructuredConfigUI(config);
      open = false;
      if (onSave) onSave();
    } catch (err) {
      console.error('Failed to save affiliation mappings:', err);
      alert('Failed to save affiliation mappings: ' + (err?.message || err));
    } finally {
      saving = false;
    }
  }

  function toggleMapping(canonical) {
    mappings = mappings.map((m) => (m.canonical === canonical ? { ...m, enabled: !m.enabled } : m));
  }

  function toggleExpanded(canonical) {
    expandedCanonical = expandedCanonical === canonical ? null : canonical;
    newAliasInput = '';
  }

  function addMapping() {
    const canonical = newCanonical.trim();
    const alias = newAlias.trim();
    if (!canonical) return;
    const existing = mappings.find((m) => m.canonical === canonical);
    if (existing) {
      if (alias && !existing.aliases.includes(alias)) {
        mappings = mappings.map((m) =>
          m.canonical === canonical ? { ...m, aliases: [...m.aliases, alias] } : m,
        );
      }
    } else {
      mappings = [...mappings, { canonical, aliases: alias ? [alias] : [], enabled: true }];
    }
    newCanonical = '';
    newAlias = '';
  }

  function removeMapping(canonical) {
    mappings = mappings.filter((m) => m.canonical !== canonical);
    if (expandedCanonical === canonical) expandedCanonical = null;
  }

  function addAliasToExpanded() {
    const alias = newAliasInput.trim();
    if (!alias || !expandedCanonical) return;
    mappings = mappings.map((m) => {
      if (m.canonical !== expandedCanonical) return m;
      if (m.aliases.includes(alias)) return m;
      return { ...m, aliases: [...m.aliases, alias] };
    });
    newAliasInput = '';
  }

  function removeAlias(canonical, alias) {
    mappings = mappings.map((m) =>
      m.canonical === canonical ? { ...m, aliases: m.aliases.filter((a) => a !== alias) } : m,
    );
  }

  // Load on mount so aliasToCanonical is available immediately (before dialog is opened)
  onMount(() => {
    loadAffiliationMap();
  });

  // Reload and reset when dialog opens
  $effect(() => {
    if (open) {
      loadAffiliationMap();
      expandedCanonical = null;
    }
  });
</script>

<Modal bind:open size="lg" title="Affiliation Deduplication Settings" placement="top-center">
  <div class="space-y-4">
    <p class="text-sm text-gray-600 dark:text-gray-400">
      Map multiple affiliation names to a canonical name for deduplication. Click a canonical name
      chip to toggle whether the mapping is active.
    </p>

    {#if loading}
      <div class="text-center py-4">
        <p class="text-gray-500">Loading...</p>
      </div>
    {:else}
      <!-- Quick add mapping -->
      <div class="border border-gray-300 dark:border-gray-600 rounded-lg p-3 space-y-2">
        <h4 class="font-semibold text-sm text-gray-700 dark:text-gray-300">Add New Mapping</h4>
        <div class="flex gap-2 items-end">
          <div class="flex-1">
            <Label class="text-xs mb-1">Canonical Name</Label>
            <Input
              bind:value={newCanonical}
              placeholder="e.g., MIT"
              size="sm"
              onkeydown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addMapping();
                }
              }}
            />
          </div>
          <div class="flex-1">
            <Label class="text-xs mb-1">First Alias (optional)</Label>
            <Input
              bind:value={newAlias}
              placeholder="e.g., Massachusetts Institute of Technology"
              size="sm"
              onkeydown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addMapping();
                }
              }}
            />
          </div>
          <Button
            size="sm"
            color="blue"
            onclick={addMapping}
            disabled={!newCanonical.trim()}
            title={isExistingCanonical() ? 'Add alias to existing mapping' : 'Add new mapping'}
          >
            <Icon icon="mdi:plus" class="w-4 h-4 mr-1" />
            {isExistingCanonical() ? 'Add Alias' : 'Add'}
          </Button>
        </div>
      </div>

      <!-- All mappings as chips (click canonical = toggle enabled) -->
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
            All Mappings (click name to enable/disable, <Icon
              icon="mdi:chevron-down"
              class="inline w-3.5 h-3.5"
            /> to manage aliases)
          </h4>
          <span class="text-xs text-gray-500">
            {mappings.length} mappings,
            {mappings.filter((m) => m.enabled).length} active
          </span>
        </div>

        <div
          class="border border-gray-200 dark:border-gray-700 rounded-lg p-2 max-h-60 overflow-y-auto bg-gray-50 dark:bg-gray-800"
        >
          {#if mappings.length === 0}
            <p class="text-sm text-gray-500 text-center py-4">
              No mappings configured yet. Add some above.
            </p>
          {:else}
            <div class="flex flex-wrap gap-2">
              {#each mappings as mapping}
                <div
                  class="inline-flex items-center gap-1 rounded-lg text-sm transition-all
                    {mapping.enabled
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}"
                >
                  <!-- Toggle enabled by clicking canonical name -->
                  <button
                    type="button"
                    onclick={() => toggleMapping(mapping.canonical)}
                    class="inline-flex items-center gap-1.5 px-2 py-1 hover:opacity-80"
                    title={mapping.enabled ? 'Click to disable mapping' : 'Click to enable mapping'}
                  >
                    {#if mapping.enabled}
                      <Icon icon="mdi:check-circle" class="w-3.5 h-3.5 shrink-0" />
                    {/if}
                    <span class="font-medium">{mapping.canonical}</span>
                    {#if mapping.aliases.length > 0}
                      <span
                        class="text-xs opacity-75 {mapping.enabled
                          ? 'text-blue-100'
                          : 'text-gray-500 dark:text-gray-400'}"
                      >
                        ({mapping.aliases.length})
                      </span>
                    {/if}
                  </button>
                  <!-- Expand aliases editor -->
                  <button
                    type="button"
                    onclick={() => toggleExpanded(mapping.canonical)}
                    class="px-1 py-1 hover:opacity-80 border-l {mapping.enabled
                      ? 'border-blue-400'
                      : 'border-gray-300 dark:border-gray-600'}"
                    title="Manage aliases"
                    aria-label="Manage aliases for {mapping.canonical}"
                  >
                    <Icon
                      icon={expandedCanonical === mapping.canonical
                        ? 'mdi:chevron-up'
                        : 'mdi:chevron-down'}
                      class="w-3.5 h-3.5"
                    />
                  </button>
                  <!-- Remove mapping -->
                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      removeMapping(mapping.canonical);
                    }}
                    class="px-1 py-1 hover:text-red-300 {mapping.enabled
                      ? ''
                      : 'hover:text-red-500 dark:hover:text-red-400'}"
                    title="Remove mapping"
                    aria-label="Remove mapping {mapping.canonical}"
                  >
                    <Icon icon="mdi:close" class="w-3 h-3" />
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Active mappings summary panel (shows enabled aliases) -->
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
            Active Mappings (click canonical to disable)
          </h4>
          <span class="text-xs text-gray-500">
            {mappings.filter((m) => m.enabled).length} active
          </span>
        </div>
        <div
          class="border border-blue-200 dark:border-blue-700 rounded-lg p-2 min-h-12 bg-blue-50 dark:bg-blue-900/20"
        >
          {#if mappings.filter((m) => m.enabled).length === 0}
            <p class="text-sm text-gray-500 text-center py-2">
              No active mappings. Click a canonical chip above to enable it.
            </p>
          {:else}
            <div class="flex flex-wrap gap-2">
              {#each mappings.filter((m) => m.enabled) as mapping}
                <button
                  type="button"
                  onclick={() => toggleMapping(mapping.canonical)}
                  class="inline-flex items-center gap-1 bg-blue-500 text-white rounded-lg px-2 py-1 text-sm hover:bg-blue-600 transition-colors cursor-pointer"
                  title="Click to disable"
                >
                  <Icon icon="mdi:check-circle" class="w-3.5 h-3.5 shrink-0" />
                  <span class="font-medium">{mapping.canonical}</span>
                  {#if mapping.aliases.length > 0}
                    <span class="text-xs text-blue-100">→ {mapping.aliases.join(', ')}</span>
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Alias editor panel (shown when a chip is expanded) -->
      {#if expandedCanonical}
        {@const expandedMapping = mappings.find((m) => m.canonical === expandedCanonical)}
        {#if expandedMapping}
          <div
            class="border border-gray-300 dark:border-gray-600 rounded-lg p-3 space-y-2 bg-white dark:bg-gray-900"
          >
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
                Aliases for
                <span class="text-blue-600 dark:text-blue-400">{expandedCanonical}</span>
              </h4>
              <button
                type="button"
                onclick={() => (expandedCanonical = null)}
                class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                aria-label="Close alias editor"
              >
                <Icon icon="mdi:close" class="w-4 h-4" />
              </button>
            </div>

            <!-- Alias chips -->
            {#if expandedMapping.aliases.length === 0}
              <p class="text-xs text-gray-500">No aliases yet. Add one below.</p>
            {:else}
              <div class="flex flex-wrap gap-2">
                {#each expandedMapping.aliases as alias}
                  <div
                    class="inline-flex items-center gap-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg px-2 py-1 text-sm"
                  >
                    <span>{alias}</span>
                    <button
                      type="button"
                      onclick={() => removeAlias(expandedCanonical, alias)}
                      class="hover:text-red-500 dark:hover:text-red-400"
                      title="Remove alias"
                      aria-label="Remove alias {alias}"
                    >
                      <Icon icon="mdi:close" class="w-3 h-3" />
                    </button>
                  </div>
                {/each}
              </div>
            {/if}

            <!-- Add alias input -->
            <div class="flex gap-2">
              <Input
                bind:value={newAliasInput}
                placeholder="New alias name..."
                size="sm"
                class="flex-1"
                onkeydown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    addAliasToExpanded();
                  }
                }}
              />
              <Button
                size="sm"
                color="blue"
                onclick={addAliasToExpanded}
                disabled={!newAliasInput.trim()}
              >
                <Icon icon="mdi:plus" class="w-4 h-4 mr-1" />
                Add Alias
              </Button>
            </div>
          </div>
        {/if}
      {/if}

      <!-- Actions -->
      <div class="flex justify-end gap-2 pt-4 border-t border-gray-200 dark:border-gray-700">
        <Button color="light" onclick={() => (open = false)}>Cancel</Button>
        <Button color="blue" onclick={saveAffiliationMap} disabled={saving}>
          {#if saving}
            <Icon icon="mdi:loading" class="w-4 h-4 mr-1 animate-spin" />
            Saving...
          {:else}
            Save
          {/if}
        </Button>
      </div>
    {/if}
  </div>
</Modal>
