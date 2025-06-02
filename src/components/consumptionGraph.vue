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
}>();

// Function to render the chart
let chartInstance: Chart | null = null; // Variable to store chart instance
const renderChart = () => {
  const ctx = consumptionGraph.value?.getContext('2d');
  if (!ctx) return;

  // Destroy previous chart if it exists
  if (chartInstance) {
    chartInstance.destroy();
  }

  // Set a fixed maximum or calculate based on data
  const maxElectricity = Math.max(...electricityUsageData.value);
  const maxDirectUsage = Math.max(...directUsageData.value);
  const maxValue = Math.max(maxElectricity, maxDirectUsage);
  const yAxisMax = Math.ceil(maxValue / 100) * 100;

  chartInstance = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels.value,
      datasets: [
        {
          label: 'Electricity Stored (%)',
          data: electricityUsageData.value,
          backgroundColor: '#007bff',
        },
        {
          label: 'Direct Usage (%)',
          data: directUsageData.value,
          backgroundColor: '#ffc107',
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
          ticks: {
            stepSize: Math.max(10, Math.floor(yAxisMax / 10))
          }
        }
      },
      animation: {
        duration: 0
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
            const date = model.dataPoints[0].label;
            const electricityValue = model.dataPoints[0].raw as number;
            const directValue = model.dataPoints[1]?.raw as number;

            tooltipData.value = {
              date: date,
              value: directValue !== undefined 
                ? `Electricity: ${electricityValue}%\nDirect: ${directValue}%`
                : `Electricity: ${electricityValue}%`
            };

            // Position tooltip relative to the chart container
            if (consumptionGraph.value) {
              const rect = consumptionGraph.value.getBoundingClientRect();
              const tooltipX = model.caretX;
              const tooltipY = model.caretY;
              
              // Calculate position relative to the chart container
              const left = tooltipX + rect.left;
              const top = tooltipY + rect.top;
              
              // Ensure tooltip stays within chart bounds
              const tooltipWidth = 120; // Approximate tooltip width
              const tooltipHeight = 60; // Approximate tooltip height
              
              let finalLeft = left;
              let finalTop = top;
              
              // Adjust if tooltip would go off the right edge
              if (left + tooltipWidth > rect.right) {
                finalLeft = left - tooltipWidth;
              }
              
              // Adjust if tooltip would go off the bottom
              if (top + tooltipHeight > rect.bottom) {
                finalTop = top - tooltipHeight;
              }
              
              tooltipStyle.value = {
                left: `${finalLeft}px`,
                top: `${finalTop}px`
              };
            }

            activeTooltip.value = true;
          }
        }
      }
    },
  });
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

    console.log('[consumptionGraph] Processed data points:', sortedData.length);

    electricityUsageData.value = sortedData.map(item => item.capacity_kwh);
    directUsageData.value = sortedData.map(item => item.consumed_kwh);
    labels.value = sortedData.map(item => item.date.split('T')[0]);

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
}, { deep: true });

// Expose the export function to the parent component
const exportData = () => {
  const csvContent = "data:text/csv;charset=utf-8,Panel,Electricity Used (%),Direct Usage (%)\n"
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
  margin: auto;
  position: relative;
}

.tooltip-container {
  position: absolute;
  background: white;
  border: 1px solid #ccc;
  border-radius: 4px;
  padding: 8px 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  pointer-events: none;
  z-index: 1000;
  min-width: 120px;
}

.tooltip-content {
  font-size: 12px;
  line-height: 1.4;
}

.tooltip-date {
  color: #666;
  margin-bottom: 4px;
}

.tooltip-value {
  color: #333;
  font-weight: 500;
  white-space: pre-line;
}
</style>