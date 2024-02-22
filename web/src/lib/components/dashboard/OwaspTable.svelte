<script>
	import { onMount } from "svelte";

  export let owaspData = {};
  export let vulnerabilities = []

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
      const criticality = item.Vulnerability.criticality.toLowerCase();

      if (!(category in results)) {
        results[category] = { information: 0, low: 0, medium: 0, high: 0, critical: 0 };
      }

      if (criticality in results[category]) {
        results[category][criticality]++;
      }
    });

    return results;
  }

  $: {
    if (vulnerabilities.length > 0){
      owaspData = categorizeData(vulnerabilities);
    }
  }
</script>

<table class="table table-vcenter card-table">
  <thead>
    <tr>
      <th>OWASP</th>
      <th>Information</th>
      <th>Low</th>
      <th>Medium</th>
      <th>High</th>
      <th>Critical</th>
    </tr>
  </thead>
  <tbody>
    {#each Object.entries(owaspData) as [category, counts]}
    <tr>
      <td><strong>{category || 'Uncategorized'}</strong></td>
      <td class="{textColor(counts.information)} text-center">{counts.information || '-'}</td>
      <td class="{textColor(counts.low)} text-center">{counts.low || '-'}</td>
      <td class="{textColor(counts.medium)} text-center">{counts.medium || '-'}</td>
      <td class="{textColor(counts.high)} text-center">{counts.high || '-'}</td>
      <td class="{textColor(counts.critical)} text-center">{counts.critical || '-'}</td>
    </tr>
    {/each}
  </tbody>
</table>