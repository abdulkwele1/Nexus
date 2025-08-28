<template>
  <div class="battery-graph">
    <canvas ref="batteryGraph"></canvas>

    <!-- Tooltip Container -->
    <div 
      v-if="activeTooltip" 
      class="tooltip-container"
      :style="tooltipStyle"
    >
      <div class="tooltip-content">
        <div class="tooltip-date">{{ tooltipData.date }}</div>
        <div v-for="(value, sensor) in tooltipData.values" :key="sensor" class="tooltip-value">
          {{ sensor }}: {{ value }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue';
import { Chart } from 'chart.js/auto';
import type { ChartConfiguration, ChartData, ChartType, ChartTypeRegistry } from 'chart.js';
import { useNexusStore } from '@/stores/nexus';

const store = useNexusStore();

interface BatteryData {
  date: string;
  battery_level: number;
}

interface SensorConfig {
  id: string;
  name: string;
  color: string;
  visible: boolean;
}

const props = defineProps<{
  startDate?: string;
  endDate?: string;
  sensors: Record<string, SensorConfig>;
}>();

// Refs
const batteryGraph = ref<HTMLCanvasElement | null>(null);
const chartInstance = ref<Chart<'bar', (number | null)[], string> | null>(null);
const activeTooltip = ref(false);
const tooltipData = ref<{
  date: string;
  values: Record<string, string>;
}>({ date: '', values: {} });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});

const sensorData = ref<Record<string, BatteryData[]>>({});
const refreshInterval = ref<number | null>(null);

const fetchBatteryData = async () => {
  try {
    const currentStartDate = props.startDate || new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    const currentEndDate = props.endDate || new Date().toISOString().split('T')[0];

    console.log('[batteryLevelsGraph] Fetching data from', currentStartDate, 'to', currentEndDate);
    console.log('[batteryLevelsGraph] Active sensors:', props.sensors);

    // Fetch data for all sensors in parallel
    const promises = Object.values(props.sensors).map(async (sensor) => {
      try {
        console.log(`[batteryLevelsGraph] Fetching data for sensor ${sensor.id} from ${currentStartDate} to ${currentEndDate}`);
        const response = await store.user.getSensorBatteryData(sensor.id, currentStartDate, currentEndDate);
        const data = response?.battery_level_data || [];
        console.log(`[batteryLevelsGraph] Got data for sensor ${sensor.id}:`, data);
        return { id: sensor.id, data };
      } catch (error) {
        console.error(`[batteryLevelsGraph] Error fetching data for sensor ${sensor.id}:`, error);
        return { id: sensor.id, data: [] };
      }
    });

    const results = await Promise.all(promises);
    
    // Update data for each sensor
    results.forEach(({ id, data }) => {
      if (data?.length) {
        const typedData = data as BatteryData[];
        sensorData.value[id] = typedData.sort((a, b) => 
          new Date(a.date).getTime() - new Date(b.date).getTime()
        );
        console.log(`[batteryLevelsGraph] Processed data for sensor ${id}:`, sensorData.value[id]);
      } else {
        console.log(`[batteryLevelsGraph] No data for sensor ${id}`);
      }
    });

    renderChart();
  } catch (error) {
    console.error("[batteryLevelsGraph] Error fetching data:", error);
  }
};

const renderChart = () => {
  if (!batteryGraph.value) return;

  const ctx = batteryGraph.value.getContext('2d');
  if (!ctx) return;

  // Destroy existing chart
  if (chartInstance.value) {
    chartInstance.value.destroy();
  }

  // Get all unique dates across all sensors
  const allDates = new Set<string>();
  Object.values(sensorData.value).forEach(data => {
    data.forEach(point => {
      allDates.add(point.date.split('T')[0]);
    });
  });

  const sortedDates = Array.from(allDates).sort();

  console.log('[batteryLevelsGraph] Creating chart with dates:', sortedDates);

  // Create datasets for visible sensors
  const datasets = Object.entries(props.sensors)
    .filter(([_, config]) => config.visible)
    .map(([key, config]) => {
      const sensorPoints = sensorData.value[config.id] || [];
      console.log(`[batteryLevelsGraph] Data points for ${config.name}:`, sensorPoints);
      
      return {
        label: config.name,
        data: sortedDates.map(date => {
          const point = sensorPoints.find(p => 
            p.date.split('T')[0] === date
          );
          const value = point?.battery_level || null;
          console.log(`[batteryLevelsGraph] ${config.name} value for ${date}:`, value);
          return value;
        }),
        backgroundColor: config.color,
        borderColor: config.color,
        borderWidth: 2,
        borderRadius: 4,
        hoverBackgroundColor: config.color,
        hoverBorderColor: config.color,
        hoverBorderWidth: 3
      };
    });

  const config: ChartConfiguration<'bar', (number | null)[], string> = {
    type: 'bar',
    data: {
      labels: sortedDates,
      datasets
    } as ChartData<'bar', (number | null)[], string>,
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          max: 100,
          grid: {
            color: 'rgba(144, 238, 144, 0.05)',
          },
          ticks: {
            color: 'rgba(255, 255, 255, 0.8)',
            font: { size: 12 }
          },
          title: {
            display: true,
            text: 'Battery Level (%)',
            color: 'rgba(255, 255, 255, 0.8)'
          }
        },
        x: {
          grid: { display: false },
          ticks: {
            color: 'rgba(255, 255, 255, 0.8)',
            font: { size: 12 }
          }
        }
      },
      plugins: {
        tooltip: {
          enabled: false,
          external: function(context) {
            const tooltip = context.tooltip;
            if (!tooltip || !tooltip.opacity) {
              activeTooltip.value = false;
              return;
            }

            const date = new Date(tooltip.dataPoints[0].label)
              .toLocaleDateString('en-US', {
                month: 'long',
                day: 'numeric',
                year: 'numeric'
              });

            const values: Record<string, string> = {};
            tooltip.dataPoints.forEach(point => {
              values[point.dataset.label || ''] = `${point.raw}%`;
            });

            tooltipData.value = { date, values };

            if (batteryGraph.value) {
              const rect = batteryGraph.value.getBoundingClientRect();
              tooltipStyle.value = {
                left: `${tooltip.caretX}px`,
                top: `${tooltip.caretY}px`
              };
            }

            activeTooltip.value = true;
          }
        },
        legend: {
          position: 'top',
          labels: {
            color: 'rgba(255, 255, 255, 0.8)',
            font: { size: 14 },
            padding: 20,
            usePointStyle: true
          }
        }
      }
    }
  };
  
  chartInstance.value = new Chart(ctx, config);
};

// Watch for changes in date range or sensor visibility
watch([
  () => props.startDate,
  () => props.endDate,
  () => props.sensors
], () => {
  fetchBatteryData();
}, { deep: true });

onMounted(() => {
  fetchBatteryData();
  refreshInterval.value = window.setInterval(fetchBatteryData, 10000);
});

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value);
  }
});
</script>

<style scoped>
.battery-graph {
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

@media (max-width: 1400px) {
  .battery-graph {
    width: calc(100% - 320px);
    min-width: 600px;
  }
}

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
  min-width: 160px;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transform: translateY(-4px);
}

.tooltip-content {
  font-size: 14px;
  line-height: 1.6;
  color: white;
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
  font-weight: 600;
  font-size: 14px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  margin-top: 4px;
}
</style>