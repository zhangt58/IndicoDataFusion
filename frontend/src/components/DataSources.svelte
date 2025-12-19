<script>
  let {
    configData = { dataSources: [] },
    expandedSources = {},
    nameErrors = {},
    indicoDataSourcePlaceholders = {},
    testDataSourcePlaceholders = {},
    loading = false,
    applying = false,
    validateNames = () => {},
    currentActiveIndex = 0,
    apiTokens = [],
    onAddIndico = () => {},
    onDelete = (_index) => {},
    onToggle = (_index) => {},
  } = $props();
</script>

<div
  class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow p-4 border border-gray-200 dark:border-gray-700"
>
  <div class="flex items-center justify-between">
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Data Sources</h3>
    <!-- Add Indico Source button -->
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="w-7 h-7 flex items-center justify-center gap-2 rounded-full bg-gray-500 text-white text-sm font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-600 dark:hover:bg-gray-700 border border-transparent focus:outline-none focus:ring-2 focus:ring-gray-400 px-0 py-0"
        onclick={() => onAddIndico()}
        disabled={loading || applying}
        aria-label="Add Indico Source"
        title="Add a new Indico data source"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 5v14M5 12h14"
          />
        </svg>
      </button>
    </div>
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
              class="text-md font-semibold text-gray-900 dark:text-gray-100 bg-transparent border-b-2 border-transparent focus:border-indigo-500 px-1 py-0.5 hover:bg-gray-50 dark:hover:bg-gray-700 rounded-sm transition-colors placeholder-gray-400 cursor-text"
            />
            {#if nameErrors[i]}
              <span class="ml-2 text-red-500 text-xs font-medium" title={nameErrors[i]}
                >{nameErrors[i]}</span
              >
            {/if}
            <!-- pencil icon to indicate editability -->
            <svg
              aria-hidden="true"
              class="w-4 h-4 text-gray-400 ml-1"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 20h9" />
              <path d="M16.5 3.5a2.1 2.1 0 013 3L7 19l-4 1 1-4 12.5-12.5z" />
            </svg>
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
              <!-- reuse icon payload (parent uses flowbite variant) -->
              🗑
            </button>
            <svg
              class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {expandedSources[
                i
              ]
                ? 'rotate-180'
                : ''}"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 9l-7 7-7-7"
              ></path>
            </svg>
          </div>
        </div>

        <!-- Content -->
        {#if expandedSources[i]}
          <div class="px-4 pb-4 pt-2 border-t border-gray-200 dark:border-gray-700 space-y-2">
            {#if dataSource.type === 'indico' && dataSource.indico}
              <!-- Indico API Configuration -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                <div>
                  <label
                    for={`ds-${i}-baseUrl`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >Base URL</label
                  >
                  <input
                    id={`ds-${i}-baseUrl`}
                    type="text"
                    bind:value={dataSource.indico.baseUrl}
                    placeholder={indicoDataSourcePlaceholders.baseUrl}
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
                <div class="md:col-span-2">
                  <label
                    for={`ds-${i}-apiTokenName`}
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    >API Token</label
                  >
                  {#if apiTokens && apiTokens.length > 0}
                    <select
                      id={`ds-${i}-apiTokenName`}
                      bind:value={dataSource.indico.apiTokenName}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                      placeholder={indicoDataSourcePlaceholders.apiToken}
                      class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
                    class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              </div>
            {/if}
            <!-- Error message for data source name -->
            {#if nameErrors[i]}
              <div class="text-red-500 text-sm mt-1">{nameErrors[i]}</div>
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>
