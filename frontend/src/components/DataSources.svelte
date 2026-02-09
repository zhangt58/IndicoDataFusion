<script>
  import Icon from '@iconify/svelte';
  import { Modal } from 'flowbite-svelte';
  import {
    addTagTo,
    removeTagFrom,
    collectAllTags,
    toggleFavoriteOn,
  } from '../utils/dataSourceUtils.js';
  import BaseUrlInput from './BaseUrlInput.svelte';
  import TagEditor from './TagEditor.svelte';
  import DataSourcesTableView from './DataSourcesTableView.svelte';
  let {
    configData = $bindable({ dataSources: [] }),
    expandedSources = {},
    nameErrors = {},
    committedBaseUrls = [],
    loading = false,
    applying = false,
    validateNames = () => {},
    currentActiveIndex = 0,
    apiTokens = [],
    onAddIndico = () => {},
    onDelete = (_index) => {},
    onToggle = (_index) => {},
    onActivate = (_index) => {},
    onToggleFavorite = null,
  } = $props();

  // Local dialog state for table view
  let showTableDialog = $state(false);

  let indicoDataSourcePlaceholders = $state({
    baseUrl: 'https://indico.jacow.org',
    eventId: '12345',
    timeout: '60s',
  });

  let testDataSourcePlaceholders = $state({
    dataDir: './testdata',
    eventInfo: 'info.json',
    abstracts: 'abstracts.json',
    contribs: 'contribs.json',
  });

  // Collect all existing tags across data sources (for suggestions)
  function getAllTags() {
    return collectAllTags(configData && configData.dataSources ? configData.dataSources : []);
  }

  // Collect all existing base URLs for suggestions
  function getAllBaseUrls() {
    const urls = new Set();
    (configData?.dataSources || []).forEach((ds) => {
      if (ds.indico?.baseUrl) urls.add(ds.indico.baseUrl);
    });
    return Array.from(urls);
  }

  // Get all existing names for validation
  function getAllNames() {
    return (configData?.dataSources || []).map((ds) => ds.name || '');
  }

  function openTableView() {
    showTableDialog = true;
  }

  function handleUpdate(index, updatedSource) {
    if (!configData || !Array.isArray(configData.dataSources)) return;
    if (index >= 0 && index < configData.dataSources.length) {
      configData.dataSources[index] = updatedSource;
      validateNames();
    }
  }
</script>

<div
  class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700"
>
  <div class="flex items-center justify-between mb-2">
    <div class="flex items-center gap-2">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 h-8 flex items-center">
        Data Sources
      </h3>
      <button
        type="button"
        class="h-8 px-2 rounded bg-gray-100 dark:bg-gray-900 hover:bg-gray-200 dark:hover:bg-gray-800 text-sm flex items-center justify-center"
        onclick={openTableView}
        title="Open table view"
        aria-label="Open table view"
      >
        <Icon icon="mdi:grid" class="w-4 h-4" />
      </button>
    </div>

    <button
      type="button"
      class="w-6 h-6 flex items-center justify-center gap-2 rounded-full bg-gray-500 text-white text-sm font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-600 dark:hover:bg-gray-700 border border-transparent focus:outline-none focus:ring-2 focus:ring-gray-400 px-0 py-0"
      onclick={() => onAddIndico()}
      disabled={loading || applying}
      aria-label="Add Indico Source"
      title="Add a new Indico data source"
    >
      <Icon icon="mdi:plus" class="w-5 h-5" aria-hidden="true" />
    </button>
  </div>

  <div class="max-h-80 overflow-y-auto space-y-1 pr-3">
    {#each configData.dataSources as dataSource, i (i)}
      <div
        class="bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <!-- Header -->
        <div
          role="button"
          tabindex="0"
          onclick={() => onToggle(i)}
          onkeydown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              onToggle(i);
            }
          }}
          class="w-full flex items-center justify-between p-1 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        >
          <span class="flex items-center gap-2">
            <!-- Make the name editable -->
            <input
              id={`ds-name-${i}`}
              type="text"
              bind:value={dataSource.name}
              oninput={() => validateNames()}
              placeholder="Data source name"
              title="Edit data source name"
              aria-label={`Data source name ${i}`}
              class="text-md font-semibold text-gray-900 dark:text-gray-100 bg-transparent border-b-2 border-transparent focus:border-indigo-500 px-1 py-0.5 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-sm transition-colors cursor-text subtle-placeholder"
            />
            {#if nameErrors[i]}
              <span class="ml-2 text-red-500 text-xs font-medium" title={nameErrors[i]}
                >{nameErrors[i]}</span
              >
            {/if}
            <!-- pencil icon to indicate editability -->
            <Icon icon="mdi:pencil" class="w-4 h-4 text-gray-400 ml-1" aria-hidden="true" />
            <span
              class="px-2 py-0.5 text-xs rounded-full {dataSource.type === 'indico'
                ? 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200'
                : 'bg-gray-100 dark:bg-gray-900 text-gray-800 dark:text-gray-200'}"
            >
              {dataSource.type === 'indico' ? 'API' : 'Test Data'}
            </span>
            {#if currentActiveIndex === i}
              <span
                class="px-2 py-0.5 text-xs rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200"
                >Active</span
              >
            {/if}
          </span>
          <div class="flex items-center gap-2">
            <!-- Activate button: sets this source as the selected active source -->
            <button
              type="button"
              class="p-1 rounded focus:outline-none focus:ring-2 focus:ring-indigo-300"
              onclick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                if (typeof onActivate === 'function') onActivate(i);
              }}
              aria-pressed={currentActiveIndex === i}
              title={currentActiveIndex === i ? 'Active data source' : 'Set as active'}
              aria-label={`Activate data source ${dataSource.name || i}`}
            >
              <Icon
                icon={currentActiveIndex === i
                  ? 'mdi:checkbox-marked-circle'
                  : 'mdi:checkbox-blank-circle-outline'}
                class="w-5 h-5 text-indigo-500"
                aria-hidden="true"
              />
            </button>

            <!-- Favorite toggle button -->
            <button
              type="button"
              class="p-1 rounded focus:outline-none focus:ring-2 focus:ring-indigo-300"
              onclick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                const ds = configData.dataSources[i];
                if (!ds) return;
                // if parent provided an onToggleFavorite handler, call it so the parent can persist changes
                if (onToggleFavorite && typeof onToggleFavorite === 'function') {
                  onToggleFavorite(ds, i);
                } else {
                  // fallback to local util for backwards compatibility
                  toggleFavoriteOn(ds);
                }
              }}
              aria-pressed={dataSource.favorite}
              title={dataSource.favorite ? 'Unmark favorite' : 'Mark favorite'}
            >
              <Icon
                icon={dataSource.favorite ? 'mdi:star' : 'mdi:star-outline'}
                class="w-5 h-5 text-yellow-500"
                aria-hidden="true"
              />
            </button>

            <!-- Delete button: stop propagation so header click doesn't toggle -->
            <button
              type="button"
              class="text-red-500 hover:text-red-700 p-1 rounded focus:outline-none focus:ring-2 focus:ring-red-300"
              onclick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onDelete(i);
              }}
              aria-label={`Delete data source ${dataSource.name || i}`}
              title="Delete data source"
            >
              <Icon icon="mdi:delete" class="w-5 h-5" aria-hidden="true" />
            </button>
            <Icon
              icon="mdi:chevron-down"
              class={`w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform ${expandedSources[i] ? 'rotate-180' : ''}`}
            />
          </div>
        </div>

        <!-- Content -->
        {#if expandedSources[i]}
          <div class="px-4 pb-4 pt-2 border-t border-gray-200 dark:border-gray-700 space-y-2">
            {#if dataSource.type === 'indico' && dataSource.indico}
              <!-- Indico API Configuration -->
              <div class="grid grid-cols-1 md:grid-cols-[2fr_1fr] gap-2">
                <div>
                  <label
                    for={`ds-${i}-baseUrl`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Base URL</label
                  >
                  <BaseUrlInput
                    id={`ds-${i}-baseUrl`}
                    value={dataSource.indico.baseUrl}
                    onChange={(v) => {
                      dataSource.indico.baseUrl = v;
                    }}
                    placeholder={indicoDataSourcePlaceholders.baseUrl}
                    suggestions={committedBaseUrls}
                  />
                </div>
                <div>
                  <label
                    for={`ds-${i}-eventId`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Event ID</label
                  >
                  <input
                    id={`ds-${i}-eventId`}
                    type="number"
                    bind:value={dataSource.indico.eventId}
                    placeholder={indicoDataSourcePlaceholders.eventId}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                </div>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-[2fr_1fr] gap-2">
                <div>
                  <label
                    for={`ds-${i}-apiTokenName`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >API Token</label
                  >
                  {#if apiTokens && apiTokens.length > 0}
                    <select
                      id={`ds-${i}-apiTokenName`}
                      bind:value={dataSource.indico.apiTokenName}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                    >
                      {#each apiTokens as t}
                        <option value={t.name}
                          >{t.name}{t.username ? ` — ${t.username}` : ''}</option
                        >
                      {/each}
                    </select>
                  {:else}
                    <input
                      id={`ds-${i}-apiTokenName`}
                      type="text"
                      bind:value={dataSource.indico.apiTokenName}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                    />
                  {/if}
                </div>
                <div>
                  <label
                    for={`ds-${i}-timeout`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Timeout</label
                  >
                  <input
                    id={`ds-${i}-timeout`}
                    type="text"
                    bind:value={dataSource.indico.timeout}
                    placeholder={indicoDataSourcePlaceholders.timeout}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">e.g., 15s, 1m, 500ms</p>
                </div>
              </div>
            {:else if dataSource.type === 'test' && dataSource.test}
              <!-- Test Data Configuration -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                <div class="md:col-span-2">
                  <label
                    for={`ds-${i}-test-dataDir`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Data Directory</label
                  >
                  <input
                    id={`ds-${i}-test-dataDir`}
                    type="text"
                    bind:value={dataSource.test.dataDir}
                    placeholder={testDataSourcePlaceholders.dataDir}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                </div>
                <div>
                  <label
                    for={`ds-${i}-test-eventInfo`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Event Info File</label
                  >
                  <input
                    id={`ds-${i}-test-eventInfo`}
                    type="text"
                    bind:value={dataSource.test.eventInfo}
                    placeholder={testDataSourcePlaceholders.eventInfo}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                </div>
                <div>
                  <label
                    for={`ds-${i}-test-abstracts`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Abstracts File</label
                  >
                  <input
                    id={`ds-${i}-test-abstracts`}
                    type="text"
                    bind:value={dataSource.test.abstracts}
                    placeholder={testDataSourcePlaceholders.abstracts}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                </div>
                <div>
                  <label
                    for={`ds-${i}-test-contribs`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Contributions File</label
                  >
                  <input
                    id={`ds-${i}-test-contribs`}
                    type="text"
                    bind:value={dataSource.test.contribs}
                    placeholder={testDataSourcePlaceholders.contribs}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                  />
                </div>
              </div>
            {/if}

            <!-- New: Favorite / Description / Tags -->
            <div class="pt-2 border-t border-gray-100 dark:border-gray-800">
              <div class="grid grid-cols-1 gap-2">
                <div>
                  <label
                    for={`ds-${i}-description`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Description</label
                  >
                  <input
                    id={`ds-${i}-description`}
                    type="text"
                    bind:value={dataSource.description}
                    placeholder="Optional note about this data source"
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 subtle-placeholder"
                    onclick={(e) => {
                      e.stopPropagation();
                    }}
                  />
                </div>

                <div>
                  <label
                    for={`ds-${i}-tags`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Tags</label
                  >
                  <TagEditor
                    tags={dataSource.tags}
                    onAdd={(t) => {
                      const ds = configData.dataSources[i];
                      if (ds) addTagTo(ds, t);
                    }}
                    onRemove={(idx) => {
                      const ds = configData.dataSources[i];
                      if (ds) removeTagFrom(ds, idx);
                    }}
                    suggestions={getAllTags().filter((t) => !(dataSource.tags || []).includes(t))}
                    placeholder="Add tag and press Enter"
                  />
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/each}
  </div>

  <Modal
    bind:open={showTableDialog}
    size="xl"
    title="Data Sources — Table View"
    dismissable
    outsideclose
  >
    <DataSourcesTableView
      dataSources={configData.dataSources}
      {currentActiveIndex}
      {apiTokens}
      existingNames={getAllNames()}
      existingTags={getAllTags()}
      existingBaseUrls={getAllBaseUrls()}
      {onActivate}
      {onDelete}
      onToggleFavorite={onToggleFavorite || toggleFavoriteOn}
      onUpdate={handleUpdate}
    />
  </Modal>
</div>
