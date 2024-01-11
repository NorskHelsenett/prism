<script>
  export let owaspData = {};

  function textColor(value){
    return value > 0 ? "text-info" : "text-secondary"
  }
</script>

<table class="table table-vcenter card-table">
  <thead>
    <tr>
      <th>OWASP</th>
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
      <td class="{textColor(counts.low)} text-center">{counts.low || '-'}</td>
      <td class="{textColor(counts.medium)} text-center">{counts.medium || '-'}</td>
      <td class="{textColor(counts.high)} text-center">{counts.high || '-'}</td>
      <td class="{textColor(counts.critical)} text-center">{counts.critical || '-'}</td>
    </tr>
    {/each}
  </tbody>
</table>