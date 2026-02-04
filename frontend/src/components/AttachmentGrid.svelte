<script>
  import { OpenSafeURL } from '../../wailsjs/go/main/App';
  import Icon from '@iconify/svelte';
  import { getAttachmentIcon } from '../utils/attachmentIcons.js';
  import { deduplicateAttachments } from '../utils/attachmentUtils.js';

  // Props: attachments (array)
  // dedupe: boolean - whether to deduplicate by title
  let { attachments = [], dedupe = false } = $props();

  async function openAttachmentLocal(url) {
    console.log('Opening attachment URL:', url);
    if (!url) return;
    try {
      await OpenSafeURL(url);
    } catch (e) {
      console.error('Failed to open attachment URL:', e);
    }
  }

  // Derived items: use provided attachments array
  let items = $derived(attachments || []);
  let displayItems = $derived(dedupe ? deduplicateAttachments(items) : items);
</script>

{#if displayItems && displayItems.length > 0}
  <!-- Always render wrap-style tile layout (horizontal wrapping) -->
  <div class="flex flex-wrap gap-2">
    {#each displayItems as attachment}
      {@const fileInfo = getAttachmentIcon(attachment)}
      <button
        onclick={() => openAttachmentLocal(attachment.download_url)}
        disabled={!attachment.download_url}
        class="group relative flex flex-col items-center justify-center w-20 h-20 rounded-md border-2 border-gray-200 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 transition-all {fileInfo.bgColor} {!attachment.download_url
          ? 'opacity-50 cursor-not-allowed'
          : 'hover:shadow-lg cursor-pointer'}"
        title={attachment.title || attachment.filename || 'Untitled'}
      >
        <span class="mb-1 block">
          <Icon icon={fileInfo.icon} class={`w-8 h-8 ${fileInfo.color}`} />
        </span>

        <span class="text-xs text-center {fileInfo.color} font-medium truncate w-full px-1">
          {(attachment.title || attachment.filename || 'Untitled').split('.')[0].substring(0, 12)}
          {(attachment.title || attachment.filename || '').length > 12 ? '...' : ''}
        </span>

        <!-- Tooltip (hover) -->
        <span
          class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-1 px-2 py-1 bg-gray-900 dark:bg-gray-700 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10 block"
        >
          {attachment.title || attachment.filename || 'Untitled'}
          {#if attachment.content_type}
            <br /><span class="text-gray-300">{attachment.content_type}</span>
          {/if}
        </span>
      </button>
    {/each}
  </div>
{/if}
