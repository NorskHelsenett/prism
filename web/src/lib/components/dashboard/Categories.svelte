<script>
    import { dashboardStore } from '$lib/stores/dashboardStore';
    import { OWASPCategories } from '$lib/OWASP/OWASPCategories';

    $: owaspCounts = Object.entries(OWASPCategories).map(([key, value]) => {
        return {
            category: value,
            count: $dashboardStore.owasp[value] || 0
        };
    });

    $: uncategorizedCount = $dashboardStore.owasp['uncategorized'] || 0;
</script>

<div class="table-responsive">
    <table class="table table-vcenter card-table table-striped">
        <thead>
            <tr>
                <th>Category</th>
                <th>Count</th>
            </tr>
        </thead>
        <tbody>
            {#each owaspCounts as { category, count }}
                <tr>
                    <td>{category}</td>
                    <td class="text-secondary small">{count}</td>
                </tr>
            {/each}
            <tr>
                <td>Uncategorized</td>
                <td class="text-secondary small">{uncategorizedCount}</td>
            </tr>
        </tbody>
    </table>
</div>
