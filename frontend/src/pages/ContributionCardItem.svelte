<script>
  import { BrowserOpenURL } from '../../wailsjs/runtime';
  import { formatDate } from '../utils/dateUtils.js';
  import TypeBadge from './TypeBadge.svelte';
  import SessionBadge from './SessionBadge.svelte';
  import AttachmentGrid from '../components/AttachmentGrid.svelte';

  let { contribution = {} } = $props();

</script>

<div
  class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700"
>
  <!-- Title and Code -->
  <div class="flex justify-between items-start mb-3">
    <div class="flex-1">
      <h3 class="text-xl font-bold text-gray-800 dark:text-white">{contribution.title}</h3>
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        ID: {contribution.id}
        {#if contribution.friendly_id}(#{contribution.friendly_id}){/if}
        {#if contribution.code}
          <span class="ml-2 px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded"
            >Code: {contribution.code}</span
          >
        {/if}
      </p>
    </div>
    {#if contribution.type}
      <TypeBadge text={contribution.type} />
    {/if}
  </div>

  <!-- Description -->
  {#if contribution.description}
    <div class="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded">
      <p class="text-sm text-gray-700 dark:text-gray-300">{contribution.description}</p>
    </div>
  {/if}

  <!-- Session and Track -->
  <div class="mb-3 flex gap-4 flex-wrap">
    {#if contribution.session}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Session:</p>
        <SessionBadge text={contribution.session} />
      </div>
    {/if}

    {#if contribution.track}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Track:</p>
        <span
          class="text-sm px-2 py-1 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded"
        >
          {contribution.track}
        </span>
      </div>
    {/if}
  </div>

  <!-- Date, Time, and Location -->
  <div class="mb-3 grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
    {#if contribution.startDate}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Start:</p>
        <p class="text-gray-700 dark:text-gray-300">{formatDate(contribution.startDate)}</p>
      </div>
    {/if}

    {#if contribution.endDate}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">End:</p>
        <p class="text-gray-700 dark:text-gray-300">{formatDate(contribution.endDate)}</p>
      </div>
    {/if}

    {#if contribution.duration}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Duration:</p>
        <p class="text-gray-700 dark:text-gray-300">{contribution.duration} minutes</p>
      </div>
    {/if}

    {#if contribution.location}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Location:</p>
        <p class="text-gray-700 dark:text-gray-300">{contribution.location}</p>
      </div>
    {/if}

    {#if contribution.room || contribution.roomFullname}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Room:</p>
        <p class="text-gray-700 dark:text-gray-300">
          {contribution.roomFullname || contribution.room}
        </p>
      </div>
    {/if}

    {#if contribution.board_number}
      <div>
        <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">Board Number:</p>
        <p class="text-gray-700 dark:text-gray-300">{contribution.board_number}</p>
      </div>
    {/if}
  </div>

  <!-- Speakers -->
  {#if contribution.speakers && contribution.speakers.length > 0}
    <div class="mb-3">
      <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Speakers:</p>
      <div class="flex gap-2 flex-wrap">
        {#each contribution.speakers as speaker}
          <div class="px-3 py-1 bg-blue-50 dark:bg-blue-900 rounded text-sm">
            <span class="font-medium text-blue-800 dark:text-blue-200">
              🎤 {speaker.fullName || `${speaker.first_name} ${speaker.last_name}`}
            </span>
            {#if speaker.affiliation}
              <div class="text-xs text-blue-600 dark:text-blue-400">{speaker.affiliation}</div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Primary Authors -->
  {#if contribution.primaryauthors && contribution.primaryauthors.length > 0}
    <div class="mb-3">
      <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Primary Authors:</p>
      <div class="flex gap-2 flex-wrap">
        {#each contribution.primaryauthors as author}
          <div class="px-3 py-1 bg-amber-50 dark:bg-amber-900 rounded text-sm">
            <span class="font-medium text-amber-800 dark:text-amber-200">
              {author.fullName || `${author.first_name} ${author.last_name}`}
            </span>
            {#if author.affiliation}
              <div class="text-xs text-amber-600 dark:text-amber-400">{author.affiliation}</div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Co-Authors -->
  {#if contribution.coauthors && contribution.coauthors.length > 0}
    <div class="mb-3">
      <p class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
        Co-Authors ({contribution.coauthors.length}):
      </p>
      <div class="flex gap-2 flex-wrap">
        {#each contribution.coauthors as coauthor}
          <div class="px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">
            <span class="text-gray-800 dark:text-gray-200">
              {coauthor.fullName || `${coauthor.first_name} ${coauthor.last_name}`}
            </span>
            {#if coauthor.affiliation}
              <div class="text-xs text-gray-600 dark:text-gray-400">{coauthor.affiliation}</div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Keywords and References -->
  {#if (contribution.keywords && contribution.keywords.length > 0) || (contribution.references && contribution.references.length > 0)}
    <div class="mb-3 flex gap-4">
      {#if contribution.keywords && contribution.keywords.length > 0}
        <div>
          <p class="text-xs font-semibold text-gray-600 dark:text-gray-400 mb-1">Keywords:</p>
          <div class="flex gap-1 flex-wrap">
            {#each contribution.keywords as keyword}
              <span
                class="px-2 py-0.5 bg-teal-100 dark:bg-teal-900 text-teal-800 dark:text-teal-200 rounded text-xs"
              >
                {keyword}
              </span>
            {/each}
          </div>
        </div>
      {/if}

      {#if contribution.references && contribution.references.length > 0}
        <div>
          <p class="text-xs font-semibold text-gray-600 dark:text-gray-400">
            References: {contribution.references.length}
          </p>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Materials and Folders -->
  {#if contribution.folders && contribution.folders.length > 0}
    <div class="mb-2">
      {#each contribution.folders as folder}
        <div class="mb-1 last:mb-0">
          <div class="flex items-center gap-2 mb-2">
            <svg
              class="w-5 h-5 text-amber-600 dark:text-amber-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
              />
            </svg>
            <h3 class="text-base font-medium text-gray-700 dark:text-gray-300">
              {folder.title || 'Attachments'}
            </h3>
            {#if folder.default_folder}
              <span class="px-2 py-0.5 text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded">
                Default
              </span>
            {/if}
            {#if folder.is_protected}
              <span title="Protected">
                <svg class="w-4 h-4 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </span>
            {/if}
          </div>

          {#if folder.description}
            <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">{folder.description}</p>
          {/if}

          <AttachmentGrid attachments={folder.attachments} dedupe={true} />
        </div>
      {/each}
    </div>
  {/if}

  <!-- Link to Indico -->
  {#if contribution.url}
    <div class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
      <a
        href={contribution.url}
        onclick={async (e) => {
          e.preventDefault();
          try {
            await BrowserOpenURL(contribution.url);
          } catch (e) {
            console.error('BrowserOpenURL failed', e);
          }
        }}
        class="text-sm text-blue-600 dark:text-blue-400 hover:underline"
        title="Open contribution link in web-browser"
      >
        View on Indico →
      </a>
    </div>
  {/if}
</div>
