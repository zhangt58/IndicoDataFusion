<script>
  import { Tooltip } from 'flowbite-svelte';
  export let abstract = null;

  function getTitle(a) {
    return a?.title || '(No title)';
  }

  function getSpeakerName(a) {
    if (!a) return '';
    if (a._speakerName) return a._speakerName;
    const persons = a.persons ?? [];
    if (!Array.isArray(persons) || persons.length === 0) return '';
    const sp = persons.find((p) => p?.is_speaker === true || p?.is_speaker === 'true') || persons[0];
    if (!sp) return '';
    const first = sp.first_name ?? '';
    const last = sp.last_name ?? '';
    if (first || last) return `${first}${first && last ? ' ' : ''}${last}`.trim();
    return sp.name || sp.Name || '';
  }

  function getSnippet(a, max = 200) {
    if (!a) return '';
    if (a._snippet) return a._snippet;
    const raw = a.content ?? '';
    const s = String(raw || '').replace(/\s+/g, ' ').trim().slice(0, max);
    return s + (String(raw || '').length > max ? '…' : '');
  }
</script>

{#if abstract}
  <!-- Use explicit dark tooltip type and light text classes for readability -->
  <Tooltip trigger="hover" placement="top" arrow={true} type="dark">
    <div class="max-w-xs whitespace-normal">
      <div class="font-semibold text-white">{getTitle(abstract)}</div>
      <div class="text-sm text-gray-200 mt-1">{getSpeakerName(abstract)}</div>
      <div class="text-xs text-gray-300 mt-2">{getSnippet(abstract)}</div>
    </div>
  </Tooltip>
{/if}
