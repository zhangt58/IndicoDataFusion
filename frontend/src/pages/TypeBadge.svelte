<script>
  let { text = '', className = '', ...rest } = $props();
</script>

<button
  class={'inline-flex items-start justify-start text-left px-2 py-1 rounded-sm text-xs ' +
    'bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200 hover:cursor-pointer ' +
    className}
  {...rest}
>
  {text}
</button>
