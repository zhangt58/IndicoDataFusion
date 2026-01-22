<script>
  let { text = '', className = '', ...rest } = $props();
</script>

<button
  class={'inline-flex items-start justify-start text-left text-xs px-2 py-1 hover:cursor-pointer ' +
   'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 ' +
   'rounded-sm ' + className
  }
  {...rest}
>
  {text}
</button>