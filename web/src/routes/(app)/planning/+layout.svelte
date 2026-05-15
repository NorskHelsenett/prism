<script>
  let { data, children } = $props();
</script>

{#key data.url}
  <div>
      {@render children?.()}
  </div>
{/key}
