<template>
  <div class="consumption-graph">
    <!-- Canvas for the graph -->
    <canvas ref="consumptionGraph"></canvas>

    <!-- Tooltip Container -->
    <div 
      v-if="activeTooltip" 
      class="tooltip-container"
      :style="tooltipStyle"
    >
      <div class="tooltip-content">
        <div class="tooltip-date">{{ tooltipData.date }}</div>
        <div class="tooltip-value">{{ tooltipData.value }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue';
import { Chart } from 'chart.js/auto';
import type { ChartConfiguration, ChartData, ChartType, TooltipItem } from 'chart.js';
import { useNexusStore } from '@/stores/nexus'
const store = useNexusStore()

interface ConsumptionData {
  date: string;
  capacity_kwh: number;
  consumed_kwh: number;
}

// Refs for data and labels
const consumptionGraph = ref<HTMLCanvasElement | null>(null);
const electricityUsageData = ref<number[]>([]);
const directUsageData = ref<number[]>([]);
const labels = ref<string[]>([]);
const refreshInterval = ref<number | null>(null);

// Props for date range
const props = defineProps<{
  startDate?: string;
  endDate?: string;
  isPieChart?: boolean;
}>();

// Add proper type for chart instance
let chartInstance: Chart | null = null;

// Function to render the chart
const renderChart = () => {
  const ctx = consumptionGraph.value?.getContext('2d');
  if (!ctx) {
    console.error('[consumptionGraph] Could not get canvas context');
    return;
  }

  // Destroy previous chart if it exists
  if (chartInstance) {
    chartInstance.destroy();
    chartInstance = null;
  }

  // Make sure we have data to display
  if (!electricityUsageData.value.length || !directUsageData.value.length) {
    console.log('[consumptionGraph] No data available to render chart');
    return;
  }

  console.log('[consumptionGraph] Rendering chart as:', props.isPieChart ? 'pie' : 'bar');

  if (props.isPieChart) {
    // Pie chart configuration
    const totalData = electricityUsageData.value.map((electricity, index) => ({
      electricity,
      direct: directUsageData.value[index],
      date: labels.value[index]
    }));

    // Calculate averages for the selected period
    const avgElectricity = totalData.reduce((sum, point) => sum + point.electricity, 0) / totalData.length;
    const avgDirect = totalData.reduce((sum, point) => sum + point.direct, 0) / totalData.length;
    
    // Update the chart configuration
    const config: ChartConfiguration = {
      type: 'pie',
      data: {
        labels: ['Avg. Electricity Stored', 'Avg. Direct Usage'],
        datasets: [{
          data: [avgElectricity || 0, avgDirect || 0],
          backgroundColor: ['rgba(144, 238, 144, 0.8)', 'rgba(255, 193, 7, 0.8)'],
          borderColor: ['rgba(144, 238, 144, 0.2)', 'rgba(255, 193, 7, 0.2)'],
          borderWidth: 2,
          hoverBackgroundColor: ['rgba(144, 238, 144, 1)', 'rgba(255, 193, 7, 1)'],
          hoverBorderColor: ['rgba(144, 238, 144, 0.4)', 'rgba(255, 193, 7, 0.4)'],
          hoverBorderWidth: 3
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        animation: {
          duration: 750,
          easing: 'easeOutQuart'
        },
        plugins: {
          title: {
            display: true,
            text: `Average Energy Distribution (${totalData.length} days)`,
            color: 'rgba(255, 255, 255, 0.9)',
            font: {
              size: 16,
              weight: 500
            },
            padding: 20
          },
          tooltip: {
            callbacks: {
              label: function(context) {
                const value = context.raw as number;
                return `${context.label}: ${value.toFixed(1)}kWh`;
              }
            },
            backgroundColor: 'rgba(26, 26, 26, 0.98)',
            titleColor: 'rgba(144, 238, 144, 0.9)',
            bodyColor: 'white',
            borderColor: 'rgba(144, 238, 144, 0.2)',
            borderWidth: 1,
            padding: 12,
            boxPadding: 6,
            usePointStyle: true
          },
          legend: {
            position: 'bottom',
            labels: {
              color: 'rgba(255, 255, 255, 0.8)',
              padding: 20,
              font: {
                size: 14
              },
              usePointStyle: true
            }
          }
        }
      }
    };

    chartInstance = new Chart(ctx, config);
  } else {
    // Bar chart configuration
    const maxElectricity = Math.max(...electricityUsageData.value);
    const maxDirectUsage = Math.max(...directUsageData.value);
    const maxValue = Math.max(maxElectricity, maxDirectUsage);
    const yAxisMax = Math.ceil(maxValue / 100) * 100;

    const config: ChartConfiguration = {
      type: 'bar',
      data: {
        labels: labels.value,
        datasets: [
          {
            label: 'Electricity Stored (kWh)',
            data: electricityUsageData.value,
            backgroundColor: 'rgba(144, 238, 144, 0.8)',
            borderColor: 'rgba(144, 238, 144, 0.2)',
            borderWidth: 2,
            borderRadius: 4,
            hoverBackgroundColor: 'rgba(144, 238, 144, 1)',
            hoverBorderColor: 'rgba(144, 238, 144, 0.4)',
            hoverBorderWidth: 3
          },
          {
            label: 'Direct Usage (kWh)',
            data: directUsageData.value,
            backgroundColor: 'rgba(255, 193, 7, 0.8)',
            borderColor: 'rgba(255, 193, 7, 0.2)',
            borderWidth: 2,
            borderRadius: 4,
            hoverBackgroundColor: 'rgba(255, 193, 7, 1)',
            hoverBorderColor: 'rgba(255, 193, 7, 0.4)',
            hoverBorderWidth: 3
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true,
            max: yAxisMax,
            grid: {
              color: 'rgba(144, 238, 144, 0.05)',
              display: true
            },
            border: {
              display: false
            },
            ticks: {
              stepSize: Math.max(10, Math.floor(yAxisMax / 10)),
              color: 'rgba(255, 255, 255, 0.8)',
              font: {
                size: 12
              }
            }
          },
          x: {
            grid: {
              display: false
            },
            border: {
              display: false
            },
            ticks: {
              color: 'rgba(255, 255, 255, 0.8)',
              font: {
                size: 12
              }
            }
          }
        },
        plugins: {
          tooltip: {
            enabled: false,
            external: function(context) {
              const model = context.tooltip;
              if (!model || !model.opacity) {
                activeTooltip.value = false;
                return;
              }

              // Set tooltip content
              const rawDate = model.dataPoints[0].label;
              const formattedDate = new Date(rawDate).toLocaleDateString('en-US', {
                month: 'long',
                day: 'numeric',
                year: 'numeric'
              });
              const electricityValue = model.dataPoints[0].raw as number;
              const directValue = model.dataPoints[1]?.raw as number;

              tooltipData.value = {
                date: formattedDate,
                value: directValue !== undefined 
                  ? `Electricity: ${electricityValue.toFixed(1)}kWh\nDirect: ${directValue.toFixed(1)}kWh`
                  : `Electricity: ${electricityValue.toFixed(1)}kWh`
              };

              // Position tooltip at mouse position
              if (consumptionGraph.value) {
                const rect = consumptionGraph.value.getBoundingClientRect();
                const mouseX = model.caretX;
                const mouseY = model.caretY;
                
                tooltipStyle.value = {
                  left: `${mouseX}px`,
                  top: `${mouseY}px`
                };
              }

              activeTooltip.value = true;
            }
          },
          legend: {
            labels: {
              color: 'rgba(255, 255, 255, 0.8)',
              font: {
                size: 14
              },
              padding: 20,
              usePointStyle: true
            }
          }
        }
      },
    };

    chartInstance = new Chart(ctx, config);
  }
};

const fetchLatestData = async () => {
  try {
    const panelId = 1;
    const currentStartDate = props.startDate || '2024-12-20';
    const currentEndDate = props.endDate || '2024-12-24';
    
    console.log(`[consumptionGraph] Fetching data for panel ${panelId} from ${currentStartDate} to ${currentEndDate}`);
    
    const response = await store.user.getPanelConsumptionData(
      panelId,
      currentStartDate,
      currentEndDate
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const responseData = await response.json();
    console.log('[consumptionGraph] Response data:', responseData);

    if (!responseData.consumption_data) {
      throw new Error('No consumption data in response');
    }

    // Process and deduplicate data
    const consumptionSolarData: ConsumptionData[] = responseData.consumption_data;
    const uniqueData = new Map<string, ConsumptionData>();
    
    // Keep only the latest entry for each date
    consumptionSolarData.forEach(item => {
      const dateKey = item.date.split('T')[0];
      uniqueData.set(dateKey, item);
    });

    // Convert back to array and sort by date
    const sortedData = Array.from(uniqueData.values())
      .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());

    // Filter data for the selected date range
    const filteredData = sortedData.filter(item => {
      const itemDate = item.date.split('T')[0];
      return itemDate >= currentStartDate && itemDate <= currentEndDate;
    });

    console.log('[consumptionGraph] Processed data points:', filteredData.length);

    electricityUsageData.value = filteredData.map(item => item.capacity_kwh);
    directUsageData.value = filteredData.map(item => item.consumed_kwh);
    labels.value = filteredData.map(item => item.date.split('T')[0]);

    console.log('[consumptionGraph] Updated chart data:', {
      electricityUsage: electricityUsageData.value,
      directUsage: directUsageData.value,
      labels: labels.value
    });

    renderChart();
  } catch (error) {
    console.error("[consumptionGraph] Error fetching updated data:", error);
  }
};

// Watch for changes in the date range props
watch(() => [props.startDate, props.endDate], ([newStartDate, newEndDate]) => {
  if (newStartDate && newEndDate) {
    console.log(`[consumptionGraph] Date range changed to ${newStartDate} - ${newEndDate}`);
    fetchLatestData();
  }
}, { immediate: true });

// Add watch for isPieChart changes
watch(() => props.isPieChart, (newValue) => {
  console.log('[consumptionGraph] Chart type changed to:', newValue ? 'pie' : 'bar');
  renderChart();
});

// Expose the export function to the parent component
const exportData = () => {
  const csvContent = "data:text/csv;charset=utf-8,Panel,Electricity Used (kWh),Direct Usage (kWh)\n"
    + labels.value.map((label, index) => `${label},${electricityUsageData.value[index]},${directUsageData.value[index]}`).join("\n");
  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  link.setAttribute("download", "Consumption_data.csv");
  document.body.appendChild(link); 
  link.click();
  document.body.removeChild(link);
};

// Expose the export function to the parent
defineExpose({
  exportData
});

onMounted(async () => {
  console.log('[consumptionGraph] Component mounted');
  await fetchLatestData();
  
  // Set up refresh interval with a longer delay
  refreshInterval.value = setInterval(fetchLatestData, 10000); // Changed to 10 seconds
  console.log('[consumptionGraph] Refresh interval set up');
});

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value);
    console.log('[consumptionGraph] Refresh interval cleared');
  }
});

// Tooltip refs
const activeTooltip = ref(false);
const tooltipData = ref({ date: '', value: '' });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});
</script>

<style scoped>
.consumption-graph {
  width: 900px;
  height: 600px;
  position: relative;
  padding: 32px;
  background: #1a1a1a;
  border-radius: 16px;
  border: 1px solid rgba(144, 238, 144, 0.1);
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.2),
    0 0 0 1px rgba(144, 238, 144, 0.05);
  margin: 100px auto;
  margin-left: 180px;
  transition: all 0.3s ease;
}

/* Make sure the graph container is visible on smaller screens */
@media (max-width: 1400px) {
  .consumption-graph {
    width: calc(100% - 320px);
    min-width: 600px;
  }
}

/* Style the chart elements */
:deep(.chart-js-render-monitor) {
  border-radius: 12px;
}

:deep(.chartjs-tooltip) {
  background: rgba(26, 26, 26, 0.98) !important;
  border: 1px solid rgba(144, 238, 144, 0.2) !important;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(144, 238, 144, 0.1),
    inset 0 0 0 1px rgba(255, 255, 255, 0.05) !important;
}

:deep(.chartjs-render-monitor) {
  filter: drop-shadow(0 4px 12px rgba(144, 238, 144, 0.1));
}

/* Style the axes and grid */
:deep(.chartjs-render-monitor) {
  & > canvas {
    border-radius: 12px;
  }
}

/* Custom tooltip styling */
.tooltip-container {
  position: absolute;
  background: rgba(26, 26, 26, 0.98);
  border: 1px solid rgba(144, 238, 144, 0.2);
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(144, 238, 144, 0.1),
    inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  pointer-events: none;
  z-index: 1000;
  min-width: 140px;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transform: translateY(-4px);
  transition: all 0.2s ease;
}

.tooltip-content {
  font-size: 14px;
  line-height: 1.6;
  letter-spacing: 0.3px;
}

.tooltip-date {
  color: rgba(144, 238, 144, 0.9);
  margin-bottom: 8px;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tooltip-value {
  color: white;
  font-weight: 600;
  font-size: 16px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  white-space: pre-line;
}
</style>