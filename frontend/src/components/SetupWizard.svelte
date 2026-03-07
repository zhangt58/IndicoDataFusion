<script>
  /**
   * SetupWizard.svelte
   *
   * A guided wizard that appears when the app emits `app:initproblems` on startup.
   * Walks the user through:
   *   Step 1 — Understanding the problem (diagnose what went wrong)
   *   Step 2 — Create / verify an API token
   *   Step 3 — Configure the data source (Indico base URL + event ID + token reference)
   *   Step 4 — Done / confirm
   *
   * The wizard can also be opened manually via the `open:setup-wizard` window event
   * or by dispatching it from the Settings component.
   */
  import { onMount, onDestroy } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import Icon from '@iconify/svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import {
    GetInitProblems,
    AddAPIToken,
    GetStructuredConfigUI,
    ApplyStructuredConfigUI,
    OpenAbstractsFileDialog,
  } from '../../wailsjs/go/main/App.js';

  // ── State ────────────────────────────────────────────────────────────────────
  let open = $state(false);
  let step = $state(1); // 1-4
  let problems = $state([]);

  // Step 2 – API token form
  let tokenForm = $state({ name: '', baseUrl: 'https://', username: '', secret: '' });
  /** @type {Record<string, string>} */
  let tokenErrors = $state({});
  let tokenSaving = $state(false);
  let tokenSaved = $state(false);

  // Step 3 – Data source form; eventId kept as string so empty/partial input works
  let dsForm = $state({
    name: '',
    baseUrl: 'https://',
    eventId: '',
    apiTokenName: '',
    timeout: '60s',
    abstractsFile: '',
  });
  /** @type {Record<string, string>} */
  let dsErrors = $state({});
  let dsSaving = $state(false);
  let dsApplied = $state(false);

  // Existing API tokens loaded from config (for step 3 picker)
  let existingTokens = $state([]);

  // Toast
  let toast = $state({ show: false, msg: '', type: 'success' });
  let toastTimer = null;

  // ── Helpers ──────────────────────────────────────────────────────────────────

  function showToast(msg, type = 'success', ms = 3500) {
    if (toastTimer) clearTimeout(toastTimer);
    toast = { show: true, msg, type };
    toastTimer = setTimeout(() => {
      toast = { show: false, msg: '', type };
      toastTimer = null;
    }, ms);
  }

  /** Classify an error string to give a useful hint. */
  function classifyProblems(list) {
    return (list || []).map((p) => {
      const s = p.toLowerCase();
      let kind = 'warning';
      let hint = '';
      if (/token|api.?key|unauthorized|401|403/.test(s)) {
        kind = 'auth';
        hint =
          'Your API token is missing, expired, or does not have the right permissions. Go to Step 2 to add or update a token.';
      } else if (/not found|404|event.?id/.test(s)) {
        kind = 'config';
        hint =
          'The event ID may be wrong, or the event does not exist on this Indico instance. Verify the Event ID in Step 3.';
      } else if (/base.?url|host|no such host|dns/.test(s)) {
        kind = 'config';
        hint = 'The Indico Base URL could not be reached. Check the URL in Step 3.';
      } else if (/timeout|deadline/.test(s)) {
        kind = 'timeout';
        hint =
          'The request timed out. Check your network connection and increase the timeout value if needed.';
      } else if (/network|dial|connect/.test(s)) {
        kind = 'network';
        hint =
          'A network error occurred. Make sure you have internet access and the Indico instance is reachable.';
      } else if (/data source|no.*configured/.test(s)) {
        kind = 'config';
        hint = 'No data source has been configured yet. Use Step 3 to add one.';
      }
      return { text: p, kind, hint };
    });
  }

  const kindMeta = {
    auth: {
      icon: 'mdi:key-alert-outline',
      color: 'text-red-500 dark:text-red-400',
      bg: 'bg-red-50 dark:bg-red-950/40 border-red-200 dark:border-red-800',
    },
    config: {
      icon: 'mdi:cog-off-outline',
      color: 'text-orange-500 dark:text-orange-400',
      bg: 'bg-orange-50 dark:bg-orange-950/40 border-orange-200 dark:border-orange-800',
    },
    timeout: {
      icon: 'mdi:timer-alert-outline',
      color: 'text-yellow-500 dark:text-yellow-300',
      bg: 'bg-yellow-50 dark:bg-yellow-950/40 border-yellow-200 dark:border-yellow-800',
    },
    network: {
      icon: 'mdi:lan-disconnect',
      color: 'text-orange-500 dark:text-orange-400',
      bg: 'bg-orange-50 dark:bg-orange-950/40 border-orange-200 dark:border-orange-800',
    },
    warning: {
      icon: 'mdi:alert-outline',
      color: 'text-yellow-500 dark:text-yellow-300',
      bg: 'bg-yellow-50 dark:bg-yellow-950/40 border-yellow-200 dark:border-yellow-800',
    },
  };

  let classified = $derived(classifyProblems(problems));

  let needsToken = $derived(
    classified.some((c) => c.kind === 'auth') || existingTokens.length === 0,
  );
  let needsDataSource = $derived(
    classified.some((c) => c.kind === 'config' || c.kind === 'warning'),
  );

  // ── Load / open ──────────────────────────────────────────────────────────────

  /**
   * @param {string[]|null} incomingProblems - null means manually opened; always show wizard.
   */
  async function loadAndOpen(incomingProblems) {
    const manual = incomingProblems === null;
    try {
      const p = incomingProblems ?? (await GetInitProblems()) ?? [];
      problems = p;
      // Only bail-out for auto-trigger with zero problems; manual open always shows wizard
      if (!manual && p.length === 0) return;
    } catch (e) {
      console.error('SetupWizard: GetInitProblems failed', e);
      problems = [];
      if (!manual) return;
    }

    // Load existing config for pre-filling forms
    try {
      const cfg = await GetStructuredConfigUI();
      existingTokens = cfg?.apiTokens ?? [];

      // Pre-fill step-3 from the first indico data source
      if (Array.isArray(cfg?.dataSources) && cfg.dataSources.length > 0) {
        const ds = cfg.dataSources[0];
        if (ds?.type === 'indico' && ds.indico) {
          dsForm = {
            name: ds.name || '',
            baseUrl: ds.indico.baseUrl || 'https://',
            eventId: String(ds.indico.eventId ?? ''),
            apiTokenName: ds.indico.apiTokenName || '',
            timeout: ds.indico.timeout || '60s',
            abstractsFile: ds.indico.abstractsFile || '',
          };
        }
      }

      // Pre-fill step-2 token base URL from first existing token (only if still default)
      if (existingTokens.length > 0 && tokenForm.baseUrl === 'https://') {
        const t = existingTokens[0];
        tokenForm = {
          ...tokenForm,
          baseUrl: t.baseUrl || t.base_url || 'https://',
          username: t.username || '',
        };
      }
    } catch (e) {
      console.error('SetupWizard: GetStructuredConfigUI failed', e);
    }

    // Reset to step 1, clear saved flags
    step = 1;
    tokenSaved = false;
    dsApplied = false;
    tokenErrors = {};
    dsErrors = {};
    open = true;
  }

  function handleInitProblemsEvent(...args) {
    const payload = args?.[0];
    const p = Array.isArray(payload) ? payload : [];
    if (p.length > 0) loadAndOpen(p);
  }

  function handleManualOpen() {
    loadAndOpen(null);
  }

  let offEvent;
  onMount(() => {
    offEvent = EventsOn('app:initproblems', handleInitProblemsEvent);
    window.addEventListener('open:setup-wizard', handleManualOpen);
    // Eager check — startup may have fired before mount
    GetInitProblems()
      .then((p) => {
        if (p?.length > 0) loadAndOpen(p);
      })
      .catch(() => {});
  });
  onDestroy(() => {
    if (offEvent) EventsOff('app:initproblems');
    window.removeEventListener('open:setup-wizard', handleManualOpen);
    if (toastTimer) clearTimeout(toastTimer);
  });

  // ── Step 2 – API token ───────────────────────────────────────────────────────

  function validateTokenForm() {
    /** @type {Record<string, string>} */
    const errs = {};
    if (!tokenForm.name.trim()) errs.name = 'Token name is required (e.g. "my-indico-token").';
    if (!tokenForm.secret.trim()) errs.secret = 'Paste your Indico API token here.';
    try {
      const u = new URL(tokenForm.baseUrl.trim());
      if (u.protocol !== 'http:' && u.protocol !== 'https:')
        errs.baseUrl = 'Must start with http:// or https://';
    } catch {
      errs.baseUrl = 'Enter a valid URL (e.g. https://indico.jacow.org).';
    }
    tokenErrors = errs; // assign whole object for Svelte 5 reactivity
    return Object.keys(errs).length === 0;
  }

  async function saveToken() {
    if (!validateTokenForm()) return;
    tokenSaving = true;
    try {
      const entry = {
        name: tokenForm.name.trim(),
        baseUrl: tokenForm.baseUrl.trim(),
        username: tokenForm.username.trim(),
        token: '',
      };
      await AddAPIToken(entry, tokenForm.secret.trim());
      tokenSaved = true;
      // Update the in-memory token list (replace existing entry with same name or append)
      existingTokens = [...existingTokens.filter((t) => t.name !== entry.name), { ...entry }];
      // Auto-fill step-3 fields if still blank
      if (!dsForm.apiTokenName) dsForm = { ...dsForm, apiTokenName: entry.name };
      if (!dsForm.baseUrl || dsForm.baseUrl === 'https://')
        dsForm = { ...dsForm, baseUrl: entry.baseUrl };
      showToast('API token saved to system keyring ✓', 'success');
    } catch (e) {
      const msg = e?.message ?? String(e);
      tokenErrors = { ...tokenErrors, save: msg }; // whole-object assignment
      showToast('Failed to save token: ' + msg, 'error');
    } finally {
      tokenSaving = false;
    }
  }

  // ── Step 3 – Data source ─────────────────────────────────────────────────────

  function validateDsForm() {
    /** @type {Record<string, string>} */
    const errs = {};
    if (!dsForm.name.trim()) errs.name = 'Give this data source a short name (e.g. "IPAC27").';
    try {
      const u = new URL(dsForm.baseUrl.trim());
      if (u.protocol !== 'http:' && u.protocol !== 'https:')
        errs.baseUrl = 'Must start with http:// or https://';
    } catch {
      errs.baseUrl = 'Enter the Indico instance base URL (e.g. https://indico.jacow.org).';
    }
    // eventId is kept as a string so we parse manually
    const evIdStr = String(dsForm.eventId).trim();
    const evId = parseInt(evIdStr, 10);
    if (!evIdStr || isNaN(evId) || evId < 0 || !Number.isInteger(evId)) {
      errs.eventId = 'Enter a positive integer event ID (visible in the Indico event URL).';
    }
    if (!dsForm.apiTokenName.trim())
      errs.apiTokenName = 'Select or type the token name to use for this source.';
    if (!/^[0-9]+(ms|s|m|h)$/.test(dsForm.timeout.trim())) errs.timeout = 'e.g. 60s, 2m, 500ms';
    dsErrors = errs; // whole-object assignment for Svelte 5 reactivity
    return Object.keys(errs).length === 0;
  }

  async function applyDataSource() {
    if (!validateDsForm()) return;
    dsSaving = true;
    try {
      // Always fetch the latest config so we don't clobber other settings
      const cfg = await GetStructuredConfigUI();
      if (!Array.isArray(cfg.dataSources)) cfg.dataSources = [];

      const newDs = {
        name: dsForm.name.trim(),
        type: 'indico',
        indico: {
          baseUrl: dsForm.baseUrl.trim(),
          eventId: parseInt(String(dsForm.eventId).trim(), 10),
          apiTokenName: dsForm.apiTokenName.trim(),
          timeout: dsForm.timeout.trim(),
          abstractsFile: dsForm.abstractsFile ? dsForm.abstractsFile.trim() : '',
        },
      };

      // Replace existing entry with same name, or append
      const idx = cfg.dataSources.findIndex((d) => d?.name === newDs.name);
      if (idx >= 0) {
        cfg.dataSources[idx] = /** @type {any} */ (newDs);
      } else {
        cfg.dataSources.push(/** @type {any} */ (newDs));
      }

      // Activate the newly configured source
      cfg.activeDataSourceName = newDs.name;

      // Merge in-memory tokens that were added this session but not yet persisted to config
      // (AddAPIToken stores the secret in keychain but doesn't update the YAML apiTokens list —
      // ApplyStructuredConfigUI is the only path that writes the metadata list)
      if (!Array.isArray(cfg.apiTokens)) cfg.apiTokens = [];
      for (const t of existingTokens) {
        const exists = cfg.apiTokens.some((ct) => ct.name === t.name);
        if (!exists) {
          cfg.apiTokens.push({
            name: t.name,
            baseUrl: t.baseUrl || '',
            username: t.username || '',
            token: '',
          });
        }
      }

      await ApplyStructuredConfigUI(cfg);
      dsApplied = true;
      showToast('Data source saved and activated ✓', 'success');
    } catch (e) {
      const msg = e?.message ?? String(e);
      dsErrors = { ...dsErrors, save: msg }; // whole-object assignment
      showToast('Failed to apply config: ' + msg, 'error');
    } finally {
      dsSaving = false;
    }
  }

  // ── Navigation ───────────────────────────────────────────────────────────────

  function next() {
    if (step < 4) step++;
  }
  function back() {
    if (step > 1) step--;
  }
  function finish() {
    open = false;
  }
  function dismiss() {
    open = false;
  }

  // ── Constants ────────────────────────────────────────────────────────────────

  const STEPS = [
    { n: 1, label: 'Diagnose', icon: 'mdi:magnify' },
    { n: 2, label: 'API Token', icon: 'mdi:key-variant' },
    { n: 3, label: 'Data Source', icon: 'mdi:database-cog' },
    { n: 4, label: 'Done', icon: 'mdi:check-circle' },
  ];

  async function browseAbstractsFile() {
    try {
      const path = await OpenAbstractsFileDialog();
      if (path) dsForm = { ...dsForm, abstractsFile: path };
    } catch (e) {
      console.error('Failed to open file dialog:', e);
    }
  }
</script>

{#if open}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-200 bg-black/60 backdrop-blur-sm"
    role="presentation"
    transition:fade={{ duration: 200 }}
  ></div>

  <!-- Wizard panel -->
  <div
    class="fixed inset-0 z-201 flex items-center justify-center p-4"
    role="dialog"
    aria-modal="true"
    aria-label="Setup Wizard"
    transition:fly={{ y: 40, duration: 300 }}
  >
    <div
      class="relative w-full max-w-2xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl flex flex-col overflow-hidden max-h-[92vh]"
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between gap-4 px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-linear-to-r from-indigo-50 to-purple-50 dark:from-indigo-950/60 dark:to-purple-950/60 shrink-0"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-indigo-600 flex items-center justify-center shadow">
            <Icon icon="mdi:auto-fix" class="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-white">Setup Wizard</h2>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Let's get you connected to Indico
            </p>
          </div>
        </div>
        <button
          onclick={dismiss}
          aria-label="Dismiss wizard"
          class="p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
        >
          <Icon icon="mdi:close" class="w-5 h-5" />
        </button>
      </div>

      <!-- Step progress bar -->
      <div
        class="flex items-center px-6 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 shrink-0"
      >
        {#each STEPS as s, i}
          <div class="flex items-center {i < STEPS.length - 1 ? 'flex-1' : ''}">
            <div class="flex flex-col items-center gap-0.5">
              <div
                class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold border-2 transition-all duration-200
                {step === s.n
                  ? 'bg-indigo-600 border-indigo-600 text-white shadow-md scale-110'
                  : step > s.n
                    ? 'bg-green-500 border-green-500 text-white'
                    : 'bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-400 dark:text-gray-500'}"
              >
                {#if step > s.n}
                  <Icon icon="mdi:check" class="w-4 h-4" />
                {:else}
                  <Icon icon={s.icon} class="w-3.5 h-3.5" />
                {/if}
              </div>
              <span
                class="text-[10px] font-medium leading-none mt-0.5 whitespace-nowrap
                {step === s.n
                  ? 'text-indigo-600 dark:text-indigo-400'
                  : step > s.n
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-gray-400 dark:text-gray-500'}"
              >
                {s.label}
              </span>
            </div>
            {#if i < STEPS.length - 1}
              <div
                class="flex-1 h-0.5 mx-1 rounded-full transition-all duration-300
                {step > s.n ? 'bg-green-400' : 'bg-gray-200 dark:bg-gray-700'}"
              ></div>
            {/if}
          </div>
        {/each}
      </div>

      <!-- Scrollable step content -->
      <div class="flex-1 overflow-y-auto px-6 py-5 space-y-4">
        <!-- ── Step 1: Diagnose ─────────────────────────────────────────────── -->
        {#if step === 1}
          <div in:fade={{ duration: 150 }}>
            <h3
              class="text-sm font-bold text-gray-900 dark:text-white mb-1 flex items-center gap-2"
            >
              <Icon icon="mdi:stethoscope" class="w-4 h-4 text-indigo-500" />
              What went wrong?
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
              The app encountered one or more problems while trying to connect to Indico. Review the
              details below to understand what needs to be fixed.
            </p>

            {#if classified.length === 0}
              <div
                class="flex items-center gap-2 p-3 rounded-lg bg-green-50 dark:bg-green-950/40 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 text-sm"
              >
                <Icon icon="mdi:check-circle" class="w-5 h-5 shrink-0" />
                <span
                  >No problems detected. You can still use this wizard to review or update your
                  configuration.</span
                >
              </div>
            {:else}
              <ul class="space-y-2">
                {#each classified as c}
                  {@const meta = kindMeta[c.kind]}
                  <li class="rounded-lg border p-3 {meta.bg}">
                    <div class="flex items-start gap-2">
                      <Icon icon={meta.icon} class="w-4 h-4 mt-0.5 shrink-0 {meta.color}" />
                      <div class="min-w-0">
                        <p
                          class="text-xs font-mono text-gray-700 dark:text-gray-200 wrap-break-word leading-snug"
                        >
                          {c.text}
                        </p>
                        {#if c.hint}
                          <p class="mt-1 text-xs text-gray-600 dark:text-gray-300 leading-snug">
                            <span class="font-semibold">Hint:</span>
                            {c.hint}
                          </p>
                        {/if}
                      </div>
                    </div>
                  </li>
                {/each}
              </ul>
            {/if}

            <!-- Recommended path callout -->
            <div
              class="mt-4 rounded-lg bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800 p-3 text-xs text-indigo-800 dark:text-indigo-200"
            >
              <p class="font-semibold mb-1 flex items-center gap-1.5">
                <Icon icon="mdi:lightbulb-on-outline" class="w-3.5 h-3.5" />
                Recommended fix
              </p>
              {#if needsToken && needsDataSource}
                <p>
                  You likely need to <strong>add an API token</strong> (Step 2) and then
                  <strong>configure a data source</strong> (Step 3).
                </p>
              {:else if needsToken}
                <p>Go to <strong>Step 2</strong> to add or update your Indico API token.</p>
              {:else if needsDataSource}
                <p>
                  Go to <strong>Step 3</strong> to fix the data source configuration (URL, Event ID, etc.).
                </p>
              {:else}
                <p>
                  Review your network connection, then check the data source configuration in Step
                  3.
                </p>
              {/if}
            </div>
          </div>

          <!-- ── Step 2: API Token ────────────────────────────────────────────── -->
        {:else if step === 2}
          <div in:fade={{ duration: 150 }}>
            <h3
              class="text-sm font-bold text-gray-900 dark:text-white mb-1 flex items-center gap-2"
            >
              <Icon icon="mdi:key-variant" class="w-4 h-4 text-indigo-500" />
              Add or update your API token
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
              An Indico API token is required to authenticate requests. Tokens are stored securely
              in your OS keychain — never in plain text files.
            </p>

            <!-- How-to callout -->
            <details
              class="mb-3 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/40 text-xs text-blue-800 dark:text-blue-200"
            >
              <summary
                class="px-3 py-2 cursor-pointer font-semibold flex items-center gap-1.5 select-none"
              >
                <Icon icon="mdi:help-circle-outline" class="w-3.5 h-3.5" />
                How to get an Indico API token
              </summary>
              <div class="px-3 pb-3 pt-1 leading-relaxed">
                <ol class="list-decimal ml-4 space-y-1">
                  <li>
                    Log in to your Indico instance (e.g. <code
                      class="font-mono bg-blue-100 dark:bg-blue-900 px-0.5 rounded"
                      >https://indico.jacow.org</code
                    >).
                  </li>
                  <li>
                    Click your avatar / name in the top-right corner → <strong>My Profile</strong>.
                  </li>
                  <li>Navigate to <strong>Settings → API tokens</strong> tab.</li>
                  <li>
                    Click <strong>Create new token</strong>, give it a name, and enable all required
                    scopes.
                  </li>
                  <li>
                    Copy the token value and paste it below. <strong
                      >You can only see it once!</strong
                    >
                  </li>
                </ol>
              </div>
            </details>

            <div class="space-y-3">
              <div>
                <label
                  for="wiz-token-name"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Token name <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400"
                    >(a short label, used to reference this token in data sources)</span
                  >
                </label>
                <input
                  id="wiz-token-name"
                  type="text"
                  placeholder="e.g. my-indico-token"
                  bind:value={tokenForm.name}
                  class="w-full rounded-lg border px-3 py-2 text-sm
                    {tokenErrors.name
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                {#if tokenErrors.name}<p class="mt-0.5 text-xs text-red-500">
                    {tokenErrors.name}
                  </p>{/if}
              </div>

              <div>
                <label
                  for="wiz-token-baseurl"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Indico instance URL <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400"
                    >(the server this token belongs to)</span
                  >
                </label>
                <input
                  id="wiz-token-baseurl"
                  type="text"
                  placeholder="https://indico.jacow.org"
                  bind:value={tokenForm.baseUrl}
                  list="wiz-baseurl-suggestions"
                  class="w-full rounded-lg border px-3 py-2 text-sm
                    {tokenErrors.baseUrl
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                <datalist id="wiz-baseurl-suggestions">
                  <option value="https://indico.jacow.org">https://indico.jacow.org</option>
                  <option value="https://indico.global">https://indico.global</option>
                  <option value="https://indico.cern.ch">https://indico.cern.ch</option>
                </datalist>
                {#if tokenErrors.baseUrl}<p class="mt-0.5 text-xs text-red-500">
                    {tokenErrors.baseUrl}
                  </p>{/if}
              </div>

              <div>
                <label
                  for="wiz-token-username"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Username <span class="ml-1 font-normal text-gray-400"
                    >(optional, for display only)</span
                  >
                </label>
                <input
                  id="wiz-token-username"
                  type="text"
                  placeholder="your.name@institution.org"
                  bind:value={tokenForm.username}
                  class="w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
              </div>

              <div>
                <label
                  for="wiz-token-secret"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  API token value <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400"
                    >(stored in OS keychain, never in config files)</span
                  >
                </label>
                <input
                  id="wiz-token-secret"
                  type="password"
                  placeholder="Paste your token here"
                  bind:value={tokenForm.secret}
                  autocomplete="off"
                  class="w-full rounded-lg border px-3 py-2 text-sm font-mono
                    {tokenErrors.secret
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                {#if tokenErrors.secret}<p class="mt-0.5 text-xs text-red-500">
                    {tokenErrors.secret}
                  </p>{/if}
                {#if tokenErrors.save}<p class="mt-0.5 text-xs text-red-500">
                    {tokenErrors.save}
                  </p>{/if}
              </div>

              <button
                type="button"
                onclick={saveToken}
                disabled={tokenSaving}
                class="w-full flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition-colors
                  {tokenSaved
                  ? 'bg-green-600 hover:bg-green-700'
                  : 'bg-indigo-600 hover:bg-indigo-700'} text-white disabled:opacity-60"
              >
                {#if tokenSaving}
                  <Icon icon="mdi:loading" class="w-4 h-4 animate-spin" />
                  Saving…
                {:else if tokenSaved}
                  <Icon icon="mdi:check-circle" class="w-4 h-4" />
                  Token saved — click again to update
                {:else}
                  <Icon icon="mdi:content-save-outline" class="w-4 h-4" />
                  Save token to keychain
                {/if}
              </button>
            </div>

            {#if existingTokens.length > 0}
              <div
                class="mt-3 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 px-3 py-2 text-xs text-gray-600 dark:text-gray-400"
              >
                <p
                  class="font-semibold mb-1 text-gray-700 dark:text-gray-300 flex items-center gap-1"
                >
                  <Icon icon="mdi:information-outline" class="w-3.5 h-3.5" />
                  Existing tokens ({existingTokens.length})
                </p>
                <ul class="space-y-0.5">
                  {#each existingTokens as t}
                    <li class="flex items-center gap-1.5">
                      <Icon icon="mdi:key" class="w-3 h-3 text-indigo-400 shrink-0" />
                      <span class="font-mono">{t.name}</span>
                      {#if t.baseUrl || t.base_url}
                        <span class="text-gray-400">— {t.baseUrl || t.base_url}</span>
                      {/if}
                    </li>
                  {/each}
                </ul>
                <p class="mt-1 text-gray-400">
                  You can skip this step if an existing token is still valid.
                </p>
              </div>
            {/if}
          </div>

          <!-- ── Step 3: Data Source ──────────────────────────────────────────── -->
        {:else if step === 3}
          <div in:fade={{ duration: 150 }}>
            <h3
              class="text-sm font-bold text-gray-900 dark:text-white mb-1 flex items-center gap-2"
            >
              <Icon icon="mdi:database-cog-outline" class="w-4 h-4 text-indigo-500" />
              Configure your data source
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
              A data source tells the app <em>which</em> Indico event to load and <em>how</em> to
              authenticate. Fill in the details below, then click
              <strong>Save &amp; activate</strong>.
            </p>

            <details
              class="mb-3 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/40 text-xs text-blue-800 dark:text-blue-200"
            >
              <summary
                class="px-3 py-2 cursor-pointer font-semibold flex items-center gap-1.5 select-none"
              >
                <Icon icon="mdi:help-circle-outline" class="w-3.5 h-3.5" />
                Where do I find the Base URL and Event ID?
              </summary>
              <div class="px-3 pb-3 pt-1 leading-relaxed">
                <p>Open the Indico event in your browser. The URL looks like:</p>
                <code
                  class="block bg-blue-100 dark:bg-blue-900 rounded px-2 py-1 font-mono text-[11px] break-all mt-1"
                >
                  https://indico.jacow.org/event/<strong>14439</strong>/
                </code>
                <ul class="list-disc ml-4 mt-2 space-y-0.5">
                  <li>
                    <strong>Base URL</strong> — everything before
                    <code class="font-mono bg-blue-100 dark:bg-blue-900 px-0.5 rounded"
                      >/event/</code
                    >, e.g.
                    <code class="font-mono bg-blue-100 dark:bg-blue-900 px-0.5 rounded"
                      >https://indico.jacow.org</code
                    >
                  </li>
                  <li>
                    <strong>Event ID</strong> — the number after
                    <code class="font-mono bg-blue-100 dark:bg-blue-900 px-0.5 rounded"
                      >/event/</code
                    >, e.g.
                    <code class="font-mono bg-blue-100 dark:bg-blue-900 px-0.5 rounded">14439</code>
                  </li>
                </ul>
              </div>
            </details>

            <div class="space-y-3">
              <!-- Name -->
              <div>
                <label
                  for="wiz-ds-name"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Data source name <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400">(short label, e.g. "IPAC25")</span>
                </label>
                <input
                  id="wiz-ds-name"
                  type="text"
                  placeholder="IPAC25"
                  bind:value={dsForm.name}
                  class="w-full rounded-lg border px-3 py-2 text-sm
                    {dsErrors.name
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                {#if dsErrors.name}<p class="mt-0.5 text-xs text-red-500">{dsErrors.name}</p>{/if}
              </div>

              <!-- Base URL -->
              <div>
                <label
                  for="wiz-ds-baseurl"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Indico Base URL <span class="text-red-500">*</span>
                </label>
                <input
                  id="wiz-ds-baseurl"
                  type="text"
                  placeholder="https://indico.jacow.org"
                  bind:value={dsForm.baseUrl}
                  list="wiz-ds-baseurl-suggestions"
                  class="w-full rounded-lg border px-3 py-2 text-sm
                    {dsErrors.baseUrl
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                <datalist id="wiz-ds-baseurl-suggestions">
                  <option value="https://indico.jacow.org">https://indico.jacow.org</option>
                  <option value="https://indico.global">https://indico.global</option>
                  <option value="https://indico.cern.ch">https://indico.cern.ch</option>
                  {#each existingTokens as t}
                    {#if t.baseUrl || t.base_url}
                      <option value={t.baseUrl || t.base_url}>{t.baseUrl || t.base_url}</option>
                    {/if}
                  {/each}
                </datalist>
                {#if dsErrors.baseUrl}<p class="mt-0.5 text-xs text-red-500">
                    {dsErrors.baseUrl}
                  </p>{/if}
              </div>

              <!-- Event ID — kept as text to avoid type=number coercion quirks -->
              <div>
                <label
                  for="wiz-ds-eventid"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Event ID <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400">(integer from the event URL)</span>
                </label>
                <input
                  id="wiz-ds-eventid"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  placeholder="14439"
                  bind:value={dsForm.eventId}
                  class="w-full rounded-lg border px-3 py-2 text-sm font-mono
                    {dsErrors.eventId
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                {#if dsErrors.eventId}<p class="mt-0.5 text-xs text-red-500">
                    {dsErrors.eventId}
                  </p>{/if}
              </div>

              <!-- API token picker -->
              <div>
                <label
                  for="wiz-ds-tokenname"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  API token to use <span class="text-red-500">*</span>
                  <span class="ml-1 font-normal text-gray-400">(select from your saved tokens)</span
                  >
                </label>
                {#if existingTokens.length > 0}
                  <select
                    id="wiz-ds-tokenname"
                    bind:value={dsForm.apiTokenName}
                    class="w-full rounded-lg border px-3 py-2 text-sm
                      {dsErrors.apiTokenName
                      ? 'border-red-400 dark:border-red-600'
                      : 'border-gray-300 dark:border-gray-600'}
                      bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                  >
                    <option value="">— select a token —</option>
                    {#each existingTokens as t}
                      <option value={t.name}>{t.name}{t.username ? ` (${t.username})` : ''}</option>
                    {/each}
                  </select>
                {:else}
                  <input
                    id="wiz-ds-tokenname"
                    type="text"
                    placeholder="token name from Step 2"
                    bind:value={dsForm.apiTokenName}
                    class="w-full rounded-lg border px-3 py-2 text-sm font-mono
                      {dsErrors.apiTokenName
                      ? 'border-red-400 dark:border-red-600'
                      : 'border-gray-300 dark:border-gray-600'}
                      bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                  />
                  <p
                    class="mt-0.5 text-xs text-amber-500 dark:text-amber-400 flex items-center gap-1"
                  >
                    <Icon icon="mdi:alert-outline" class="w-3.5 h-3.5 shrink-0" />
                    No tokens saved yet — go back to Step 2 to add one first.
                  </p>
                {/if}
                {#if dsErrors.apiTokenName}<p class="mt-0.5 text-xs text-red-500">
                    {dsErrors.apiTokenName}
                  </p>{/if}
              </div>

              <!-- Timeout -->
              <div>
                <label
                  for="wiz-ds-timeout"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Request timeout
                  <span class="ml-1 font-normal text-gray-400">(default: 60s)</span>
                </label>
                <input
                  id="wiz-ds-timeout"
                  type="text"
                  placeholder="60s"
                  bind:value={dsForm.timeout}
                  class="w-32 rounded-lg border px-3 py-2 text-sm font-mono
                    {dsErrors.timeout
                    ? 'border-red-400 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'}
                    bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                />
                {#if dsErrors.timeout}<p class="mt-0.5 text-xs text-red-500">
                    {dsErrors.timeout}
                  </p>{/if}
              </div>

              <!-- Additional abstracts file (review mode) -->
              <div>
                <label
                  for="wiz-ds-abstractsfile"
                  class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1"
                >
                  Additional abstracts file (review mode)
                  <span class="ml-1 font-normal text-gray-400">(optional)</span>
                </label>
                <div class="flex gap-2 items-center">
                  <input
                    id="wiz-ds-abstractsfile"
                    type="text"
                    placeholder="Leave empty to use live Indico API"
                    bind:value={dsForm.abstractsFile}
                    class="flex-1 rounded-lg border px-3 py-2 text-sm font-mono
                      {dsErrors.abstractsFile
                      ? 'border-red-400 dark:border-red-600'
                      : 'border-gray-300 dark:border-gray-600'}
                      bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-400"
                  />
                  <button
                    type="button"
                    class="shrink-0 px-2 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    onclick={(e) => {
                      e.stopPropagation();
                      browseAbstractsFile();
                    }}
                    title="Browse for abstracts JSON file"
                    aria-label="Browse for abstracts file"
                  >
                    <Icon icon="mdi:folder-open-outline" class="w-4 h-4" aria-hidden="true" />
                  </button>
                  {#if dsForm.abstractsFile}
                    <button
                      type="button"
                      class="shrink-0 px-2 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-red-50 dark:hover:bg-red-900 text-red-500 focus:outline-none focus:ring-2 focus:ring-red-300"
                      onclick={() => (dsForm = { ...dsForm, abstractsFile: '' })}
                      title="Clear abstracts file (use live API)"
                      aria-label="Clear abstracts file"
                    >
                      <Icon icon="mdi:close" class="w-4 h-4" aria-hidden="true" />
                    </button>
                  {/if}
                </div>
                <p class="mt-0.5 text-xs text-gray-400">
                  Optional — when set, review mode will load this JSON file of abstracts in addition
                  to the event data. Clear to use the live Indico API.
                </p>
              </div>

              <!-- Save & activate button -->
              <button
                type="button"
                onclick={applyDataSource}
                disabled={dsSaving}
                class="w-full flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition-colors
                  {dsApplied
                  ? 'bg-green-600 hover:bg-green-700'
                  : 'bg-indigo-600 hover:bg-indigo-700'} text-white disabled:opacity-60"
              >
                {#if dsSaving}
                  <Icon icon="mdi:loading" class="w-4 h-4 animate-spin" />
                  Applying…
                {:else if dsApplied}
                  <Icon icon="mdi:check-circle" class="w-4 h-4" />
                  Data source applied — click again to re-apply
                {:else}
                  <Icon icon="mdi:content-save-outline" class="w-4 h-4" />
                  Save &amp; activate data source
                {/if}
              </button>
              {#if dsErrors.save}
                <p class="text-xs text-red-500 flex items-center gap-1">
                  <Icon icon="mdi:alert-circle-outline" class="w-3.5 h-3.5 shrink-0" />
                  {dsErrors.save}
                </p>
              {/if}
            </div>
          </div>

          <!-- ── Step 4: Done ─────────────────────────────────────────────────── -->
        {:else if step === 4}
          <div
            in:fade={{ duration: 150 }}
            class="flex flex-col items-center text-center py-4 gap-4"
          >
            {#if dsApplied || tokenSaved}
              <div
                class="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/50 flex items-center justify-center"
              >
                <Icon icon="mdi:check-circle" class="w-8 h-8 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <h3 class="text-base font-bold text-gray-900 dark:text-white mb-1">
                  Setup complete!
                </h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
                  Your configuration has been saved. The app will reload data from the configured
                  source automatically.
                </p>
              </div>
              <div
                class="w-full rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950/40 px-4 py-3 text-sm text-green-800 dark:text-green-200 text-left space-y-1"
              >
                {#if tokenSaved}
                  <p class="flex items-center gap-1.5">
                    <Icon icon="mdi:check" class="w-4 h-4 shrink-0" /> API token saved to keychain
                  </p>
                {/if}
                {#if dsApplied}
                  <p class="flex items-center gap-1.5">
                    <Icon icon="mdi:check" class="w-4 h-4 shrink-0" />
                    Data source <strong>{dsForm.name}</strong> configured and activated
                  </p>
                {/if}
              </div>
              <p class="text-xs text-gray-400 dark:text-gray-500">
                If you still see errors, open <strong>Settings → Data Sources</strong> for advanced options,
                or re-run this wizard from the notification banner.
              </p>
            {:else}
              <div
                class="w-16 h-16 rounded-full bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center"
              >
                <Icon
                  icon="mdi:alert-circle-outline"
                  class="w-8 h-8 text-amber-500 dark:text-amber-400"
                />
              </div>
              <div>
                <h3 class="text-base font-bold text-gray-900 dark:text-white mb-1">
                  Nothing was saved yet
                </h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
                  Complete at least one step, or open <strong>Settings → Data Sources</strong> to edit
                  manually.
                </p>
              </div>
              <button
                type="button"
                onclick={() => {
                  step = 1;
                }}
                class="inline-flex items-center gap-2 rounded-lg px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors"
              >
                <Icon icon="mdi:arrow-left" class="w-4 h-4" />
                Start over
              </button>
            {/if}
          </div>
        {/if}
      </div>
      <!-- /scrollable content -->

      <!-- Footer navigation -->
      <div
        class="flex items-center justify-between gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/80 shrink-0"
      >
        <div class="flex items-center gap-2">
          {#if step > 1}
            <button
              type="button"
              onclick={back}
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              <Icon icon="mdi:arrow-left" class="w-4 h-4" />
              Back
            </button>
          {/if}
          <button
            type="button"
            onclick={dismiss}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            Skip wizard
          </button>
        </div>

        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400 dark:text-gray-500">Step {step} of {STEPS.length}</span
          >
          {#if step < 4}
            <button
              type="button"
              onclick={next}
              class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors shadow-sm"
            >
              Next
              <Icon icon="mdi:arrow-right" class="w-4 h-4" />
            </button>
          {:else}
            <button
              type="button"
              onclick={finish}
              class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-green-600 hover:bg-green-700 text-white text-sm font-semibold transition-colors shadow-sm"
            >
              <Icon icon="mdi:check" class="w-4 h-4" />
              Finish
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>

  <!-- Toast notification -->
  {#if toast.show}
    <div
      class="fixed bottom-6 left-1/2 -translate-x-1/2 z-210 px-4 py-2 rounded-xl shadow-lg text-sm font-medium flex items-center gap-2
        {toast.type === 'error'
        ? 'bg-red-600 text-white'
        : toast.type === 'info'
          ? 'bg-blue-600 text-white'
          : 'bg-green-600 text-white'}"
      role="status"
      aria-live="polite"
      in:fly={{ y: 12, duration: 200 }}
      out:fade={{ duration: 150 }}
    >
      <Icon
        icon={toast.type === 'error'
          ? 'mdi:close-circle'
          : toast.type === 'info'
            ? 'mdi:information'
            : 'mdi:check-circle'}
        class="w-4 h-4 shrink-0"
      />
      {toast.msg}
    </div>
  {/if}
{/if}
