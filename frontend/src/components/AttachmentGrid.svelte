<script>
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

  // Props: attachments (array)
  // dedupe: boolean - whether to deduplicate by title
  let { attachments = [], dedupe = false } = $props();

  // Deduplicate attachments by title, keeping the latest modified
  function deduplicateAttachments(arr) {
    if (!arr || arr.length === 0) return [];

    const titleMap = new Map();
    for (const attachment of arr) {
      const rawKey = attachment.title || attachment.filename || 'untitled';
      let key = String(rawKey).trim().toLowerCase();
      const existing = titleMap.get(key);

      if (!existing) {
        titleMap.set(key, attachment);
      } else {
        const existingDate = existing.modified_dt ? new Date(existing.modified_dt) : new Date(0);
        const currentDate = attachment.modified_dt ? new Date(attachment.modified_dt) : new Date(0);
        if (currentDate > existingDate) {
          titleMap.set(key, attachment);
        }
      }
    }

    return Array.from(titleMap.values());
  }

  // Get file type icon & classes
  function getFileIcon(attachment) {
    const contentType = (attachment.content_type || '') + '';
    const filename = attachment.filename || attachment.title || '';
    const ext = (filename.split('.').pop() || '').toLowerCase();

    if (contentType.includes('pdf') || ext === 'pdf') {
      return {
        icon: 'pdf',
        color: 'text-red-600 dark:text-red-400',
        bgColor: 'bg-red-50 dark:bg-red-900/20'
      };
    }
    if (contentType.includes('image') || ['jpg','jpeg','png','gif','svg','webp'].includes(ext)) {
      return {
        icon: 'image',
        color: 'text-green-600 dark:text-green-400',
        bgColor: 'bg-green-50 dark:bg-green-900/20'
      };
    }
    if (contentType.includes('word') || contentType.includes('msword') || ['doc','docx'].includes(ext)) {
      return {
        icon: 'document',
        color: 'text-blue-600 dark:text-blue-400',
        bgColor: 'bg-blue-50 dark:bg-blue-900/20'
      };
    }
    if (contentType.includes('excel') || contentType.includes('spreadsheet') || ['xls','xlsx','csv'].includes(ext)) {
      return {
        icon: 'table',
        color: 'text-emerald-600 dark:text-emerald-400',
        bgColor: 'bg-emerald-50 dark:bg-emerald-900/20'
      };
    }
    if (contentType.includes('presentation') || contentType.includes('powerpoint') || ['ppt','pptx'].includes(ext)) {
      return {
        icon: 'presentation',
        color: 'text-orange-600 dark:text-orange-400',
        bgColor: 'bg-orange-50 dark:bg-orange-900/20'
      };
    }
    if (['zip','rar','7z','tar','gz'].includes(ext)) {
      return {
        icon: 'archive',
        color: 'text-purple-600 dark:text-purple-400',
        bgColor: 'bg-purple-50 dark:bg-purple-900/20'
      };
    }
    if (contentType.includes('video') || ['mp4','avi','mov','wmv','flv','mkv'].includes(ext)) {
      return {
        icon: 'video',
        color: 'text-pink-600 dark:text-pink-400',
        bgColor: 'bg-pink-50 dark:bg-pink-900/20'
      };
    }
    return {
      icon: 'file',
      color: 'text-gray-600 dark:text-gray-400',
      bgColor: 'bg-gray-50 dark:bg-gray-700/50'
    };
  }

  async function openAttachmentLocal(url) {
    if (!url) return;
    try {
      await BrowserOpenURL(url);
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
      {@const fileInfo = getFileIcon(attachment)}
      <button
        onclick={() => openAttachmentLocal(attachment.download_url)}
        disabled={!attachment.download_url}
        class="group relative flex flex-col items-center justify-center w-20 h-20 rounded-md border-2 border-gray-200 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 transition-all {fileInfo.bgColor} {!attachment.download_url ? 'opacity-50 cursor-not-allowed' : 'hover:shadow-lg cursor-pointer'}"
        title={attachment.title || attachment.filename || 'Untitled'}
      >
        <span class="mb-1 block">
          {#if fileInfo.icon === 'pdf'}
            <svg class="w-8 h-8 {fileInfo.color}" fill="currentColor" viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z"/><path d="M14 2v6h6M9.5 12.5v5M11.5 12.5v5M9.5 15h2" stroke="currentColor" stroke-width="1" fill="none"/></svg>
          {:else}
            <svg class="w-8 h-8 {fileInfo.color}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
          {/if}
        </span>

        <span class="text-xs text-center {fileInfo.color} font-medium truncate w-full px-1">
          {(attachment.title || attachment.filename || 'Untitled').split('.')[0].substring(0, 12)}
          {(attachment.title || attachment.filename || '').length > 12 ? '...' : ''}
        </span>

        <!-- Tooltip (hover) -->
        <span class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-1 px-2 py-1 bg-gray-900 dark:bg-gray-700 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10 block">
          {attachment.title || attachment.filename || 'Untitled'}
          {#if attachment.content_type}
            <br /><span class="text-gray-300">{attachment.content_type}</span>
          {/if}
        </span>

      </button>
    {/each}
  </div>
{/if}
