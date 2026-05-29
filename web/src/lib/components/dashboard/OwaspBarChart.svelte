<script>
  import { run } from 'svelte/legacy';

	import { onMount } from "svelte";

  let { owaspData = $bindable({}), vulnerabilities = [] } = $props();

  function textColor(value){
    return value > 0 ? "text-info" : "text-secondary"
  }

  onMount(() => {
    if (vulnerabilities.length > 0){
      owaspData = categorizeData(vulnerabilities);
    }
  })

  function categorizeData(vulns) {
    const results = {};

    vulns.forEach(item => {
      const category = item.Vulnerability.category || '';
      const criticality = (item.Vulnerability?.criticality || '').toLowerCase();

      if (!(category in results)) {
        results[category] = { information: 0, low: 0, medium: 0, high: 0, critical: 0 };
      }

      if (criticality in results[category]) {
        results[category][criticality]++;
      }
    });

    return results;
  }


  let totalFinding = $state(0);
  let categoryTotals = $state();

  function calculateWidth(value) {
    const total = categoryTotals.filter(item => item.category === value)[0].total;
    return `${Math.round((total / totalFinding) * 100)}%`;
  }

  run(() => {
    if (vulnerabilities.length > 0){
      owaspData = categorizeData(vulnerabilities);
    }

    totalFinding = 0

    categoryTotals = Object.entries(owaspData).map(([category, counts]) => {
      const total = Object.values(counts).reduce((acc, count) => acc + count, 0);
      totalFinding += total; // Find the max for scaling the bars
      return { category, total };
    });
  });
</script>

<table class="table table-vcenter card-table table-fixed">
  <thead class="th-lg" hidden>
    <tr>
      <th>OWASP</th>
      <th>Percentage</th>
      <th>Bar</th>
    </tr>
  </thead>
  <tbody>
    {#each Object.entries(owaspData) as [category, counts]}
    <tr>
      <td style="min-width: 30% !important"><strong>{category || 'Uncategorized'}</strong></td>
      <td class="text-azure">{calculateWidth(category)}</td>
      <td class="bar-td">
        <div class="bar-container">
          <span class="bar" style="width: {calculateWidth(category)}" title="{calculateWidth(category)}"></span>
          <span class="bar-background"></span>
        </div>
      </td>
    </tr>
    {/each}
  </tbody>
</table>

<style>
  :root {
    --bar-height: 20px; /* Define the variable for bar height */
    --bar-radius: 3px; /* Define the variable for bar border radius */
  }

  .bar-container {
    position: relative;
    height: var(--bar-height); /* Use the variable for height */
  }

  .bar {
    display: block;
    position: absolute;
    height: var(--bar-height); /* Use the variable for height */
    border-radius: var(--bar-radius); /* Use the variable for border radius */
    background-color: #4399E2;
    z-index: 1;
  }

  .bar-background {
    display: block;
    position: absolute;
    height: var(--bar-height); /* Use the variable for height */
    border-radius: var(--bar-radius); /* Use the variable for border radius */
    background-color: #4398e223;
    left: 0;
    right: 0;
    z-index: 0;
  }

  .bar-td {
    width: 70% !important;
    padding: 0;
  }
</style>