<script>
  import { run } from 'svelte/legacy';

  import { Doughnut } from 'svelte-chartjs';
  import ChartDataLabels from 'chartjs-plugin-datalabels';
  import {
    Chart as ChartJS,
    Title,
    Tooltip,
    Legend,
    ArcElement,
    CategoryScale,
  } from 'chart.js';

  ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale, ChartDataLabels);

  let { severityData = $bindable({ critical: 0, high: 0, medium: 0, low: 0, information: 0 }), vulnerabilities = [] } = $props();

  let data = $state({
  labels: [
    'Critical',
    'High',
    'Medium',
    'Low',
    'Information',
  ],
  datasets: [
    {
      data: [0, 0, 0, 0, 0],
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
    spacing: 0,
    hoverOffset: 15 // Distance the slice moves when hovered
    },
  ],
});

  run(() => {
    if (severityData) {
      data.datasets[0].data = [
        severityData.critical,
        severityData.high,
        severityData.medium,
        severityData.low,
        severityData.information,
      ];
    }
  });

  // Non-reactive options to prevent $state.snapshot from trying to clone formatter functions
  const options = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: 0,
    plugins: {
      datalabels: {
        display: true,
        align: 'end',
        anchor: 'end',
        clamp: true,
        formatter: (value, context) => {
          return value > 0 ? context.chart.data.labels[context.dataIndex] : "";
        },
        color: '#4399E1',
        font: {
          weight: 'regular',
        },
        offset: 20,
      },
      legend:{
        position: "right",
        display: false
      }
    },
    layout: {
      padding: {
        top: 40,
        right: 30,
        bottom: 50,
        left: 30
      }
    }
  };

  run(() => {
    if (vulnerabilities.length > 0) {
      // Reset counts
      severityData = { critical: 0, high: 0, medium: 0, low: 0, information: 0 };

      vulnerabilities.forEach(vulnerability => {
        const criticality = vulnerability.Vulnerability.criticality.toLowerCase();

        if (criticality === 'critical') {
          severityData.critical += 1;
        } else if (criticality === 'high') {
          severityData.high += 1;
        } else if (criticality === 'medium') {
          severityData.medium += 1;
        } else if (criticality === 'low') {
          severityData.low += 1;
        } else if (criticality === 'information') {
          severityData.information += 1;
        }
      });
    }
  });

</script>

{#if severityData}
  <Doughnut {data} {options} />
{/if}
