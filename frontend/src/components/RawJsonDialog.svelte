<script>
  let {
    open = $bindable(false),
    data = $bindable({}),
    title = $bindable('View JSON')
  } = $props();

  let showCopied = $state(false);

  async function copyJson() {
    try {
      const text = JSON.stringify(data || {}, null, 2);
      if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        showCopied = true;
        setTimeout(() => (showCopied = false), 2000);
      } else {
        // fallback: create a temporary textarea
        const ta = document.createElement('textarea');
        ta.value = text;
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        showCopied = true;
        setTimeout(() => (showCopied = false), 2000);
      }
    } catch (err) {
      alert('Copy failed: ' + (err && err.message ? err.message : String(err)));
    }
  }

  function close() {
    open = false;
  }

  function downloadJson() {
    try {
      const text = JSON.stringify(data || {}, null, 2);
      const blob = new Blob([text], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = (typeof title === 'string' ? title.replace(/\s+/g, '_') : 'data') + '.json';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      alert('Download failed: ' + (err && err.message ? err.message : String(err)));
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-40 flex items-center justify-center" aria-hidden={!open}>
    <div
      class="fixed inset-0 bg-black/50 transition-opacity"
      role="presentation"
      onclick={close}
    ></div>

    <div
      class="relative z-50 max-w-3xl w-full mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-100">{title}</h2>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="text-xs px-2 py-1 rounded bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600"
            onclick={copyJson}
            aria-label={showCopied ? 'Copied' : 'Copy JSON'}
            title={showCopied ? 'Copied' : 'Copy JSON'}
          >
            {#if showCopied}
              ✓ Copied
            {:else}
              Copy
            {/if}
          </button>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600"
            onclick={downloadJson}
            title="Download JSON"
          >
            Download
          </button>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 hover:bg-red-200 dark:hover:bg-red-800"
            onclick={close}
            title="Close dialog"
          >
            Close
          </button>
        </div>
      </div>

      <div class="p-4 max-h-[60vh] overflow-auto bg-gray-50 dark:bg-gray-900">
        <pre class="text-xs text-gray-800 dark:text-gray-100 whitespace-pre-wrap">
{JSON.stringify(data || {}, null, 2)}
        </pre>
      </div>
    </div>
  </div>
{/if}

