<script>
  export let owaspData = {};
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
      <td>{category || 'Uncategorized'}</td>
      <td class="text-secondary text-center">{counts.low || '-'}</td>
      <td class="text-secondary text-center">{counts.medium || '-'}</td>
      <td class="text-secondary text-center">{counts.high || '-'}</td>
      <td class="text-secondary text-center">{counts.critical || '-'}</td>
    </tr>
    {/each}
  </tbody>
</table>