<script>
  import { onMount } from 'svelte';
  import Icon from '@iconify/svelte';

  let { error = null, message = null, title = 'Failed to load information' } = $props();

  // Parse a brief, human-readable reason from the raw error
  let reason = $derived.by(() => {
    const raw = (error instanceof Error ? error.message : String(error ?? '')) || message || '';
    if (!raw) return null;
    if (/status\s+401/i.test(raw))
      return { kind: 'auth', text: 'Authentication failed (HTTP 401).' };
    if (/status\s+403/i.test(raw))
      return { kind: 'auth', text: 'Access denied (HTTP 403) — API token may lack permissions.' };
    if (/status\s+404/i.test(raw))
      return {
        kind: 'config',
        text: 'Resource not found (HTTP 404) — check the event ID or base URL.',
      };
    if (/status\s+5\d\d/i.test(raw)) {
      const m = raw.match(/status\s+(\d+)/i);
      return {
        kind: 'server',
        text: `Indico server error (HTTP ${m ? m[1] : '5xx'}) — the server may be unavailable.`,
      };
    }
    if (/timeout|deadline/i.test(raw))
      return { kind: 'timeout', text: 'Request timed out — the server took too long to respond.' };
    if (/no such host|dial|connect|network|ECONNREFUSED/i.test(raw))
      return {
        kind: 'network',
        text: 'Network error — could not reach the server. Check the base URL.',
      };
    if (/token|api.?key|unauthorized/i.test(raw))
      return { kind: 'auth', text: 'API token missing or invalid.' };
    // Trim long raw errors
    const trimmed = raw.length > 120 ? raw.slice(0, 120) + '…' : raw;
    return { kind: 'generic', text: trimmed, full_text: raw };
  });

  // Per-kind fix guidance
  const guidance = {
    auth: 'Open Settings → Configuration → Advanced → API Tokens and verify the token is correct and has the required scope. If you cannot fix authentication immediately, you can still review abstracts by setting an "Additional abstracts file" on the data source (Settings → Data Sources) or by using the Open Setup Wizard button to configure it quickly — note that review assignments still require API access.',
    config:
      'Open Settings → Configuration and double-check the Base URL and Event ID for the active data source.',
    server:
      'The remote Indico server returned an error. Try refreshing in a moment, or contact the event administrator.',
    timeout:
      'Open Settings → Configuration and increase the request timeout, or verify the server is reachable.',
    network:
      'Open Settings → Configuration and verify the Base URL is reachable from this machine.',
    generic:
      'Open Settings → Configuration and verify the Base URL, Event ID, and API Token are correct.',
  };

  const kindIcon = {
    auth: 'mdi:key-alert',
    config: 'mdi:cog-off',
    server: 'mdi:server-network-off',
    timeout: 'mdi:timer-alert',
    network: 'mdi:lan-disconnect',
    generic: 'mdi:alert-circle',
  };

  function openSettings() {
    try {
      window.dispatchEvent(new CustomEvent('open:settings', { detail: { tab: 'config' } }));
    } catch (e) {
      console.warn('Could not dispatch open:settings event', e);
    }
  }

  function openSetupWizard() {
    try {
      window.dispatchEvent(new CustomEvent('open:setup-wizard'));
    } catch (e) {
      console.warn('Could not dispatch open:setup-wizard event', e);
    }
  }

  onMount(() => {
    if (error) console.debug('LoadErrorHint error:', error);
  });
</script>

<div class="flex items-start justify-center p-10">
  <div
    class="w-full max-w-lg rounded-xl border border-red-200 dark:border-red-800 bg-white dark:bg-gray-900 shadow-md overflow-hidden"
  >
    <!-- Header bar -->
    <div
      class="flex items-center gap-3 bg-red-50 dark:bg-red-950 px-5 py-4 border-b border-red-200 dark:border-red-800"
    >
      <Icon icon="mdi:alert-circle" class="w-6 h-6 text-red-500 dark:text-red-400 shrink-0" />
      <h2 class="text-base font-semibold text-red-700 dark:text-red-300">{title}</h2>
    </div>

    <div class="px-5 py-4 space-y-4">
      <!-- Fail reason -->
      {#if reason}
        <div
          class="flex items-start gap-3 rounded-lg bg-red-50 dark:bg-red-950/60 border border-red-100 dark:border-red-900 px-4 py-3"
        >
          <Icon
            icon={kindIcon[reason.kind] ?? 'mdi:alert-circle'}
            class="w-5 h-5 mt-0.5 text-red-500 dark:text-red-400 shrink-0"
          />
          <p class="text-sm text-red-700 dark:text-red-300 leading-snug" title={reason.full_text}>
            {reason.text}
          </p>
        </div>
      {/if}

      <!-- Fix guidance -->
      <div
        class="flex items-start gap-3 rounded-lg bg-amber-50 dark:bg-amber-950/50 border border-amber-200 dark:border-amber-800 px-4 py-3"
      >
        <Icon
          icon="mdi:lightbulb-on"
          class="w-5 h-5 mt-0.5 text-amber-500 dark:text-amber-400 shrink-0"
        />
        <p class="text-sm text-amber-800 dark:text-amber-200 leading-snug">
          {guidance[reason?.kind ?? 'generic']}
        </p>
      </div>

      <!-- Action -->
      <div class="flex justify-end space-x-2">
        <button
          onclick={openSettings}
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 active:bg-indigo-800 text-white text-sm font-medium transition-colors"
        >
          <Icon icon="mdi:cog" class="w-4 h-4" />
          Open Settings
        </button>
        <button
          onclick={openSetupWizard}
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white text-sm font-medium transition-colors"
          title="Open the Setup Wizard to guide issue resolution"
          aria-label="Open Setup Wizard"
        >
          <Icon icon="mdi:auto-fix" class="w-4 h-4" />
          Open Setup Wizard
        </button>
      </div>
    </div>
  </div>
</div>
