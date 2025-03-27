<script>
  export let data
</script>

{#key data.url}
  <div>
      <slot />
  </div>
{/key}
