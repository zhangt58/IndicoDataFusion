<script>
  let { className = '', children, ...rest } = $props();
</script>

<button
  class={'inline-flex items-start justify-start text-left p-0 bg-transparent ' +
    'border-0 text-sm font-normal text-sky-600 hover:text-sky-700 ' +
    'dark:text-sky-300 dark:hover:text-sky-100 hover:cursor-pointer ' +
    className}
  {...rest}
>
  {@render children()}
</button>
