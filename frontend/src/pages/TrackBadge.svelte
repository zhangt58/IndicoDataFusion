<script>
  let {
    text = '',
    type = '',
    className = '',
    ...rest
  } = $props();

  // Compute classes as a derived value (Svelte 5 Rune API) so it's reactive and allowed in runes mode
  const classes = $derived(() =>
    'inline-flex items-start justify-start text-left px-2 py-1 rounded-sm text-xs hover:cursor-pointer ' +
    (type === 'accepted'
      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-100'
      : type === 'reviewed'
      ? 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-200'
      : type === 'proposed'
      ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'
      : 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200') +
    (className ? ' ' + className : '')
  );

</script>

<button
  type="button"
  {...rest}
  class={classes()}
>
  {text}
</button>


