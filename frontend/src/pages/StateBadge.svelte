<script>
  let {
    state = '',
    className = ''
  } = $props();

  const classes = $derived(() =>
      'inline-block text-xs font-medium rounded-sm px-2 py-1 ' +
      (state === 'accepted'
        ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 '
        : state === 'rejected'
        ? 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200 '
        : state === 'submitted'
        ? 'bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-200 '
        : state === 'withdrawn'
        ? 'bg-gray-100 dark:bg-gray-900 text-gray-800 dark:text-gray-200 '
        : 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 ') +
      (className ? ' ' + className : '')
  );
</script>

<span
  role="status"
  aria-label={state}
  title={state}
  class={classes()}
>
  {state}
</span>