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
                ? `Electricity: ${electricityValue}%\nDirect: ${directValue}%`
                : `Electricity: ${electricityValue}%`
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
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 
    0 4px 24px -1px rgba(0, 0, 0, 0.08),
    0 0 1px 0 rgba(0, 0, 0, 0.06),
    inset 0 0 0 1px rgba(255, 255, 255, 0.15);
  pointer-events: none;
  z-index: 1000;
  min-width: 120px;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

.tooltip-content {
  font-size: 12px;
  line-height: 1.5;
  color: rgba(0, 0, 0, 0.8);
}

.tooltip-date {
  color: rgba(0, 0, 0, 0.6);
  margin-bottom: 6px;
  font-weight: 500;
}

.tooltip-value {
  color: rgba(0, 0, 0, 0.9);
  font-weight: 600;
}
</style>