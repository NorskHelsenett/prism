<script>
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

  export let severityData = { critical: 0, high: 0, medium: 0, low: 0, information: 0 };

  let data = {
  labels: [
    'Critical',
    'High',
    'Medium',
    'Low',
    'Information',
  ],
  datasets: [
    {
      data: [1, 4, 5, 1, 6],
      backgroundColor: [
        '#D63939', // Greyish Blue
        '#F76706', // Soft Teal
        '#F59F01', // Light Grey-Blue
        '#0054A6', // Dark Slate Blue
        '#4399E1', // Muted Blue
    ],
    hoverBackgroundColor: [
        '#BF3030', // More saturated Greyish Blue
        '#E65C00', // More saturated Soft Teal
        '#E08C00', // More saturated Light Grey-Blue
        '#00458C', // More saturated Dark Slate Blue
        '#3A87D1', // More saturated Muted Blue
    ],
    borderColor: 'transparent',
    borderWidth: 0,
    hoverOffset: 15 // Distance the slice moves when hovered
    },
  ],
};

  $: if (severityData) {
    data.datasets[0].data = [
      severityData.critical,
      severityData.high,
      severityData.medium,
      severityData.low,
      severityData.information,
    ];
  }

  let options = {
    responsive: true ,
    maintainAspectRatio: false,
    plugins: {
      datalabels: {
        display: false
      },
      legend:{
        position: "right",
        display: false
      }
    }
  }

</script>

{#if severityData}
	<Pie {data} {options} />
{/if}
