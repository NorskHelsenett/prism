<script>
  export let impact;

  function getSeverityClasses(localSeverity = "") {
      const baseClasses = {
          'low': ['bg-info'],
          'medium': ['bg-info', 'bg-custom-orange'],
          'high': ['bg-info', 'bg-custom-orange', 'bg-red'],
      };

      let classes = baseClasses[localSeverity.toLowerCase()] || [];
      while (classes.length < Object.keys(baseClasses).length) {
          classes.push('bg-muted opacity-01');
      }

      return classes;
  }

  let showTooltip = false

  let height = [12,16,20]
</script>

<div on:mouseover={() => showTooltip = true} on:mouseout={() => showTooltip = false} style="min-width: 9em !important;">
  {#each getSeverityClasses(impact) as severityClass, index}
    <span class={severityClass + ' text-white avatar ml-03'} style="height: {height[index]}px; width: 16px;">&nbsp;</span>
  {/each}
  {#if showTooltip}
    <div class="dropdown-menu dropdown-menu-demo dropdown-menu-arrow show">
        <div class="m-2 capitalize-first-letter">
          {impact}
      </div>
    </div>
  {/if}
</div>



<style>
  .opacity-01 {
    opacity: 0.1;
  }
  .ml-03 {
    margin-left: 3px;
  }
  .capitalize-first-letter::first-letter {
    text-transform: capitalize;
  }
  .bg-custom-orange {
      background-color: #ffa500; /* This is just an example orange color */
  }
</style>