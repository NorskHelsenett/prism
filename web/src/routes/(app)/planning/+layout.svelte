<script>
  import { fly } from 'svelte/transition'

  export let data
  export let reverse = false;
</script>

{#key data.url}
  <div
    in:fly={{ x: reverse ? -200 : 200, duration: 300, delay: 300 }}
    out:fly={{ x: reverse ? 200 : -200, duration: 300 }}>
      <slot />
  </div>
{/key}
