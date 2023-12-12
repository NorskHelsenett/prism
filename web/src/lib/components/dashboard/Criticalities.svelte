<script>
  import { dashboardStore } from '$lib/stores/dashboardStore';
  import { Pie } from 'svelte-chartjs';

  import {
    Chart as ChartJS,
    Title,
    Tooltip,
    Legend,
    ArcElement,
    CategoryScale,
  } from 'chart.js';

  ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale);

export let data = {
  labels: ['Information', 'Low', 'Medium', 'High', 'Critical'],
  datasets: [
    {
      data: [0, 0, 0, 0, 0],
      backgroundColor: [
        '#687980', // Muted Blue
        '#A9B7C0', // Light Grey-Blue
        '#7A9E9F', // Soft Teal
        '#90AFC5', // Greyish Blue
        '#336B87', // Dark Slate Blue
    ],
    hoverBackgroundColor: [
        '#789BA1', // Lighter Muted Blue
        '#BCC8D1', // Lighter Grey-Blue
        '#8CB0B2', // Lighter Soft Teal
        '#A0B9D0', // Lighter Greyish Blue
        '#497A9B', // Lighter Dark Slate Blue
    ],
    },
  ],
};



    $: if ($dashboardStore && $dashboardStore.criticality) {
        // Update chart data whenever the store data changes
        data.datasets[0].data = [
            $dashboardStore.criticality.information,
            $dashboardStore.criticality.low,
            $dashboardStore.criticality.medium,
            $dashboardStore.criticality.high,
            $dashboardStore.criticality.critical
        ];
    }

</script>

<Pie {data} options={{ responsive: true ,maintainAspectRatio: false, plugins: {legend:{position: "right"}}}} />

<style>
  [data-bs-theme='dark'] .chart-container {
    color: white;
  }
</style>