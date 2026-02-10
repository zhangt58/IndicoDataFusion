<script>
  let {
    open = $bindable(false),
    title = 'Enter Password',
    message = '',
    confirmLabel = 'Confirm',
    cancelLabel = 'Cancel',
    onConfirm = () => {},
    onCancel = () => {},
    working = false,
  } = $props();

  let password = $state('');
  let showPassword = $state(false);
  let error = $state('');

  // Reset state when dialog opens
  $effect(() => {
    if (open) {
      password = '';
      showPassword = false;
      error = '';
    }
  });

  async function doConfirm() {
    if (!password.trim()) {
      error = 'Password cannot be empty';
      return;
    }

    error = '';
    try {
      await onConfirm(password);
      // Only close if onConfirm doesn't throw
      open = false;
    } catch (e) {
      error = e.message || String(e);
      console.error('PasswordDialog onConfirm error', e);
    }
  }

  function doCancel() {
    open = false;
    try {
      onCancel();
    } catch (e) {
      console.error('PasswordDialog onCancel callback error', e);
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && !working) {
      e.preventDefault();
      doConfirm();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      doCancel();
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-40 flex items-center justify-center">
    <div
      class="absolute inset-0 bg-black/40"
      role="button"
      tabindex="0"
      aria-label="Close dialog"
      onclick={doCancel}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'Spacebar') {
          e.preventDefault();
          doCancel();
        }
      }}
    ></div>

    <div
      role="dialog"
      aria-modal="true"
      tabindex="0"
      onkeydown={handleKeydown}
      class="relative z-50 w-full max-w-md mx-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 pointer-events-auto"
    >
      <h4 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">{title}</h4>
      {#if message}
        <p class="text-sm text-gray-700 dark:text-gray-300 mb-3">{message}</p>
      {/if}

      <div class="space-y-3">
        <div class="relative">
          <label
            for="password-input"
            class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
          >
            Password
          </label>
          <div class="relative">
            <input
              id="password-input"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              disabled={working}
              placeholder="Enter password"
              class="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
              autocomplete="off"
            />
            <button
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
              onclick={() => (showPassword = !showPassword)}
              disabled={working}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {#if showPassword}
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                  />
                </svg>
              {:else}
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                  />
                </svg>
              {/if}
            </button>
          </div>
        </div>

        {#if error}
          <div class="text-sm text-red-600 dark:text-red-400">
            {error}
          </div>
        {/if}
      </div>

      <div class="mt-4 flex justify-end gap-2">
        <button
          type="button"
          class="px-3 py-1 rounded bg-gray-200 dark:bg-gray-700 text-sm disabled:opacity-50 disabled:cursor-not-allowed"
          onclick={doCancel}
          disabled={working}
        >
          {cancelLabel}
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 text-white bg-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
          onclick={doConfirm}
          disabled={working}
        >
          {working ? 'Working...' : confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
