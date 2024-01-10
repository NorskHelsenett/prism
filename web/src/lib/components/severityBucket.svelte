<script>
  export let severity;

  function getSeverityClasses(localSeverity = "") {
      const baseClasses = {
          'information': ['bg-info'],
          'low': ['bg-info', 'bg-primary'],
          'medium': ['bg-info', 'bg-primary', 'bg-custom-orange'],
          'high': ['bg-info', 'bg-primary', 'bg-custom-orange', 'bg-warning'],
          'critical': ['bg-info', 'bg-primary', 'bg-custom-orange', 'bg-warning', 'bg-red']
      };

      let classes = baseClasses[localSeverity.toLowerCase()] || [];
      while (classes.length < 5) {
          classes.push('bg-muted opacity-01');
      }

      return classes;
  }

  let showTooltip = false
</script>

<div on:mouseover={() => showTooltip = true} on:mouseout={() => showTooltip = false}>
  {#each getSeverityClasses(severity) as severityClass}
    <span class={severityClass + ' text-white avatar ml-03'} style="height: 16px; width: 16px;">&nbsp;</span>
  {/each}
  {#if showTooltip}
    <div class="dropdown-menu dropdown-menu-demo dropdown-menu-arrow show">
        <div class="m-2 capitalize-first-letter">
          {severity}
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
</style>