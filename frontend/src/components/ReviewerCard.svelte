<script>
  import AffiliationBadge from './AffiliationBadge.svelte';
  import Icon from "@iconify/svelte";

  // Props
  let { reviewer = null, showEmail = true, onAffiliationClick = null } = $props();

  // Compute a display name (rune-style)
  const displayName = $derived(() =>
    reviewer
      ? reviewer.full_name || `${reviewer.first_name || ''} ${reviewer.last_name || ''}`.trim()
      : ''
  );

  // Avatar URL (nullable)
  const avatarUrl = $derived(() => (reviewer && reviewer.avatar_url ? reviewer.avatar_url : null));

  // Initials fallback
  function initials(name) {
    if (!name) return '';
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length === 0) return '';
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
</script>

{#if reviewer}
  <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-3 flex items-center gap-3">
    <!-- Avatar -->
    <div class="shrink-0">
      {#if avatarUrl()}
        <img
          src={avatarUrl()}
          alt={"Avatar of " + displayName()}
          class="w-12 h-12 rounded-full object-cover border border-gray-200 dark:border-gray-700"
        />
      {:else}
        <div
          class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-lg font-semibold text-gray-700 dark:text-gray-200 border border-gray-200 dark:border-gray-700"
          aria-hidden="true"
        >
          {initials(displayName())}
        </div>
      {/if}
    </div>

    <!-- Details -->
    <div class="flex-1 min-w-0">
      <div class="flex items-start justify-between gap-2">
        <div class="min-w-0">
          <div class="text-sm font-semibold text-gray-900 dark:text-white truncate">{displayName()}</div>
          {#if reviewer.title}
            <div class="text-xs text-gray-600 dark:text-gray-400 truncate">{reviewer.title}</div>
          {/if}
        </div>

        <div class="text-right text-xs text-gray-500 dark:text-gray-400">
          {#if reviewer.id}
            <div>ID: {reviewer.id}</div>
          {/if}
        </div>
      </div>

      <div class="mt-2 flex items-center gap-2 flex-wrap">
        {#if showEmail && reviewer.email}
          <Icon icon="mdi:email" class="w-4 h-4" />
          <a
            href={`mailto:${reviewer.email}`}
            class="text-xs truncate hover:text-blue-500 hover:dark:text-blue-100"
            title={reviewer.email}
          >
            {reviewer.email}
          </a>
        {/if}

        {#if reviewer.affiliation}
          <AffiliationBadge affiliation={reviewer.affiliation} onclick={onAffiliationClick} className="text-xs" />
        {/if}
      </div>
    </div>
  </div>
{:else}
  <div class="p-3 text-sm text-gray-500 dark:text-gray-400">No reviewer data available</div>
{/if}
