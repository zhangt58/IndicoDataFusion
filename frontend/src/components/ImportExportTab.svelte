<script>
  import { tick } from 'svelte';
  import { ExportConfig, ImportConfig, GetStructuredConfigUI } from '../../wailsjs/go/main/App';
  import PasswordDialog from './PasswordDialog.svelte';
  import Icon from '@iconify/svelte';

  // Props
  let { configData = $bindable(null), onConfigReload = null } = $props();

  // Load config if not provided
  async function loadConfig() {
    try {
      configData = await GetStructuredConfigUI();
    } catch (e) {
      console.error('Failed to load config:', e);
      showToastMsg('Failed to load configuration', 'error');
    }
  }

  // Load config on mount if not already loaded
  $effect(() => {
    if (!configData) {
      loadConfig();
    }
  });

  // Toast state
  let showToast = $state(false);
  let toastMessage = $state('');
  let toastType = $state('success'); // 'success' | 'error' | 'info'
  let toastTimeoutId = null;

  async function showToastMsg(msg, type = 'success', duration = 3500) {
    // clear previous timeout
    if (toastTimeoutId) {
      clearTimeout(toastTimeoutId);
      toastTimeoutId = null;
    }

    toastMessage = msg || '';
    toastType = type || 'success';

    // Restart the toast animation reliably by toggling the visibility and yielding a tick.
    showToast = false;
    await tick();
    showToast = true;

    // auto-hide after duration
    toastTimeoutId = setTimeout(() => {
      showToast = false;
      toastTimeoutId = null;
    }, duration);
  }

  // Export/Import state
  let showExportPasswordDialog = $state(false);
  let showImportPasswordDialog = $state(false);
  let importFileData = $state('');
  let fileInputRef = $state(null);
  let exportingConfig = $state(false);
  let importingConfig = $state(false);
  let showConfigFile = $state(false);

  // Export configuration
  async function handleExport(password) {
    exportingConfig = true;
    try {
      const encryptedData = await ExportConfig(password);

      // Download the file
      const blob = new Blob([encryptedData], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `idf-config-export-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      showToastMsg('Configuration exported successfully', 'success');
      showExportPasswordDialog = false;
    } catch (e) {
      throw new Error(`Export failed: ${e}`);
    } finally {
      exportingConfig = false;
    }
  }

  // Import configuration
  async function handleImport(password) {
    if (!importFileData) {
      throw new Error('No file selected');
    }

    importingConfig = true;
    try {
      await ImportConfig(importFileData, password);

      // Reload the configuration from backend
      if (onConfigReload) {
        await onConfigReload();
      } else {
        await loadConfig();
      }

      showToastMsg('Configuration imported successfully', 'success');
      showImportPasswordDialog = false;
      importFileData = '';
    } catch (e) {
      throw new Error(`Import failed: ${e}`);
    } finally {
      importingConfig = false;
    }
  }

  // Handle file selection for import
  function handleFileSelect(event) {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const result = e.target?.result;
      if (typeof result === 'string') {
        importFileData = result;
        showImportPasswordDialog = true;
      }
    };
    reader.onerror = () => {
      showToastMsg('Failed to read file', 'error');
    };
    reader.readAsText(file);

    // Reset file input so the same file can be selected again
    if (event.target) {
      event.target.value = '';
    }
  }

  function openExportDialog() {
    showExportPasswordDialog = true;
  }

  function openImportDialog() {
    if (fileInputRef) {
      fileInputRef.click();
    }
  }
</script>

<div class="p-2 space-y-2 max-w-5xl mx-auto">
  <!-- Export/Import Section -->
  <div
    class="bg-gray-50 dark:bg-gray-800 rounded-lg shadow px-4 py-3 border border-gray-200 dark:border-gray-700"
  >
    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
      Export & Import Configuration
    </h3>
    <p class="text-sm text-gray-600 dark:text-gray-400 mb-0">
      Export your configuration to a password-protected file, or import a previously exported
      configuration.
    </p>
    <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">
    Includes data sources, API tokens, cache settings, affiliation map, and word
    cloud exclusions.
    </p>

    <div class="flex items-center gap-3">
      <button
        type="button"
        class="px-4 py-2 rounded-lg bg-indigo-600 text-white font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-indigo-500 dark:hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:ring-offset-2 transition-colors"
        onclick={openExportDialog}
        disabled={exportingConfig || importingConfig}
        title="Export configuration with encrypted API tokens"
      >
        <Icon icon="mdi:export" class="w-5 h-5 inline-block mr-2" />
        Export Configuration
      </button>

      <button
        type="button"
        class="px-4 py-2 rounded-lg bg-gray-600 text-white font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-500 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2 transition-colors"
        onclick={openImportDialog}
        disabled={exportingConfig || importingConfig}
        title="Import configuration from encrypted file"
      >
        <Icon icon="mdi:import" class="w-5 h-5 inline-block mr-2" />
        Import Configuration
      </button>

      <!-- Hidden file input for import -->
      <input
        type="file"
        bind:this={fileInputRef}
        onchange={handleFileSelect}
        accept=".json"
        class="hidden"
      />
    </div>

    <!-- Toast notification -->
    {#if showToast}
      <div
        class="mt-3 transform transition-all duration-300 ease-out flex items-start gap-3 rounded-lg shadow-lg overflow-hidden px-3 py-2 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600"
        role="status"
        aria-live={toastType === 'error' ? 'assertive' : 'polite'}
      >
        <div class="mt-0.5">
          {#if toastType === 'success'}
            <Icon icon="mdi:check-circle" class="w-5 h-5 text-green-600" />
          {:else if toastType === 'error'}
            <Icon icon="mdi:close-circle" class="w-5 h-5 text-red-600" />
          {:else}
            <Icon icon="mdi:information" class="w-5 h-5 text-blue-600" />
          {/if}
        </div>
        <div class="flex-1 text-sm leading-tight">
          <div class="font-medium text-gray-900 dark:text-gray-100">{toastMessage}</div>
        </div>
        <button
          class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          onclick={() => {
            showToast = false;
            if (toastTimeoutId) {
              clearTimeout(toastTimeoutId);
              toastTimeoutId = null;
            }
          }}
          aria-label="Dismiss toast"
        >
          <Icon icon="mdi:close" class="w-5 h-5" />
        </button>
      </div>
    {/if}
  </div>

  <!-- Configuration File Path Info (Collapsible) -->
  <div
    class="bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
  >
    <div
      role="button"
      tabindex="0"
      onclick={() => (showConfigFile = !showConfigFile)}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          showConfigFile = !showConfigFile;
        }
      }}
      class="w-full flex items-center justify-between p-3 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors cursor-pointer"
    >
      <h4 class="text-base font-medium text-gray-900 dark:text-gray-100">Configuration File</h4>
      <Icon
        icon="mdi:chevron-down"
        class={`w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform ${showConfigFile ? 'rotate-180' : ''}`}
      />
    </div>
    {#if showConfigFile}
      <div class="px-3 pb-3 border-t border-gray-200 dark:border-gray-700 space-y-2 pt-2">
        <div class="flex flex-wrap items-center gap-2 text-sm">
          <span class="text-gray-600 dark:text-gray-400 font-medium">Path:</span>
          <span class="text-gray-800 dark:text-gray-200 font-mono text-xs break-all"
            >{configData?.pathInfo?.path || configData?.path || 'Not set'}</span
          >
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Export Password Dialog -->
<PasswordDialog
  bind:open={showExportPasswordDialog}
  title="Export Configuration"
  message="Enter a password to encrypt your configuration export. This password will be required to import the file. Export includes data sources, API tokens, cache settings, affiliation map, and word cloud exclusions."
  confirmLabel="Export"
  working={exportingConfig}
  onConfirm={handleExport}
  onCancel={() => {
    showExportPasswordDialog = false;
    exportingConfig = false;
  }}
/>

<!-- Import Password Dialog -->
<PasswordDialog
  bind:open={showImportPasswordDialog}
  title="Import Configuration"
  message="Enter the password used to encrypt this configuration file. Import will restore data sources, API tokens, cache settings, affiliation map, and word cloud exclusions."
  confirmLabel="Import"
  working={importingConfig}
  onConfirm={handleImport}
  onCancel={() => {
    showImportPasswordDialog = false;
    importingConfig = false;
    importFileData = '';
  }}
/>
