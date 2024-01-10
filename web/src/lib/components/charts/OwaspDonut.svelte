<script>
  import { onMount } from 'svelte';
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

  export let severityData = { };
  let themeColors = { };

  onMount(() => {
    // Access the computed styles of the document's root element
    const style = getComputedStyle(document.documentElement);

    // Get the values of the CSS variables
    themeColors = {
      primary: style.getPropertyValue('--tblr-bg-surface').trim(),
      secondary: style.getPropertyValue('--tblr-body-color').trim(),
      success: style.getPropertyValue('--tblr-card-bg').trim(),
    }

    console.log(themeColors)

    // Now you can use `themeColors` array in your chart configuration
  });

  let topOwaspCategories = Object.entries(severityData).sort((a,b) => b[1] - a[1]).slice(0,3)

  const labels = topOwaspCategories.map(item => item[0]);
  const dataPoints = topOwaspCategories.map(item => item[1]);

  let data = {
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
      borderColor: themeColors.primary,
      hoverOffset: 8 // Distance the slice moves when hovered

    }],
  };

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
          weight: 'bold'
        },
        // Configure your line here
        connectors: {
          style: 'solid', // Line style
          length: 20 // Line length
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
    },
    elements: {
      arc: {
        borderWidth: 0 // No border for the arcs
      }
    }
  };

</script>

{#if topOwaspCategories.length && themeColors?.primary}
  <Doughnut {data} {options} />
{/if}
