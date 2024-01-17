<script>
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

  export let severityData = {"A01:Broken Access Control":0,"A02:Cryptographic Failures":0,"A03:Injection":0,"A05:Security Misconfiguration":0,"A08:Software and Data Integrity Failures":0,"uncategorized":0};
  let data = {}

  let options = {
    responsive: true,
    maintainAspectRatio: false,
    cutoutPercentage: 80, // This makes it a donut chart
    plugins: {
      datalabels: {
        display: true,
        formatter: (value, context) => {
          // Return the label you want to display
          return context.chart.data.labels[context.dataIndex];
        },
        align: 'end',
        anchor: 'end',
        clamp: true,
        font: {
          weight: 'regular',
        },
        // Configure your line here
        connectors: {
          style: 'solid', // Line style
          length: 50 // Line length
        },
        padding: {
          top: 0,
          right: 0,
          bottom: 0,
          left: 0
        }
      },
      legend: {
        display: false // Hide legend
      },
      title: {
        display: false,
        text: 'NHN TOP 3 (Basert på OWASP TOP 10)',
        position: 'bottom',
        font: {
          size: 16
        }
      },
      tooltip: {
        enabled: true // Enable tooltips
      }
    }
  };

  $: if (severityData) {
    let topOwaspCategories = Object.entries(severityData).sort((a,b) => b[1] - a[1]).slice(0,3)

    const labels = topOwaspCategories.map(item => item[0]);
    const dataPoints = topOwaspCategories.map(item => item[1]);
    data = {
      labels: labels, // Adjust the slice as needed based on your data
      datasets: [{
        data: dataPoints, // Map severity data to dataset
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
        // Add border color and width here if needed
        borderWidth: 10,
        borderColor: 'transparent',
        hoverOffset: 15 // Distance the slice moves when hovered

      }],
    };
    }
</script>

{#if data.datasets[0].data.length}
  <Doughnut {data} {options} />
{/if}
