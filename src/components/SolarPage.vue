<template>
  <div>
    <nav class="top-nav">
      <button class="nav-button" @click="goTo('/home')">Home</button>
    </nav>
    <div class="solar-page-layout">
      <aside class="sidebar">
        <button
          :class="{ active: currentGraph === 'yield' }"
          @click="switchGraph('yield')"
        >Yield Graph</button>
        <button
          :class="{ active: currentGraph === 'consumption' }"
          @click="switchGraph('consumption')"
        >Consumption Graph</button>
        <hr />
        <button @click="showAddData = !showAddData">
          {{ showAddData ? 'Close Add Data' : 'Add Data' }}
      </button>
        <button @click="handleExport">Export Data</button>
        <button
          v-if="currentGraph === 'yield'"
          @click="isLineChart = !isLineChart"
        >
          {{ isLineChart ? 'Bar Chart' : 'Line Chart' }}
      </button>
        <div class="date-range">
          <label>Start:</label>
              <input type="date" v-model="startDate" />
          <label>End:</label>
              <input type="date" v-model="endDate" />
        </div>
        <div class="quick-filters">
          <button class="filter-btn" @click="handleQuickFilter('7days')">Last 7 Days</button>
          <button class="filter-btn" @click="handleQuickFilter('7months')">Last 7 Months</button>
        </div>
      </aside>
      <main class="main-content">
        <solarDataManagerUi
          v-if="showAddData"
          :graphType="currentGraph"
          :solarData="solarData"
          @dataAdded="fetchLatestData"
        />
        <Yield
          v-if="currentGraph === 'yield'"
          :solarData="solarData"
          :isLineChart="isLineChart"
        />
        <Consumption 
          v-if="currentGraph === 'consumption'"
          :startDate="startDate"
          :endDate="endDate"
          ref="consumptionRef"
        />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import Yield from './yieldGraph.vue';
import Consumption from './consumptionGraph.vue';
import { useNexusStore } from '@/stores/nexus'
import solarDataManagerUi from '@/components/solarDataManagerUi.vue';

interface SolarDataPoint {
  date: Date;
  kwh_yield: number;
}

const store = useNexusStore()
const router = useRouter();
const currentGraph = ref('yield');
const showAddData = ref(false);
const startDate = ref<string>('');
const endDate = ref<string>('');
const solarData = ref<SolarDataPoint[]>([]);
const isLineChart = ref(false);
const refreshInterval = ref<number | null>(null);
const consumptionRef = ref<InstanceType<typeof Consumption> | null>(null);
const isFetching = ref(false);

const switchGraph = (type: 'yield' | 'consumption') => {
  if (currentGraph.value !== type) {
    currentGraph.value = type;
    showAddData.value = false;
  }
};

const goTo = (path: string) => {
  router.push(path);
};

const handleExport = () => {
  if (currentGraph.value === 'yield') {
    const header = "sensor_reading_date,daily_kw_generated\n";
    const csvContent = "data:text/csv;charset=utf-8," 
      + header 
      + solarData.value.map((d: any) => `${d.date.toISOString().split('T')[0]},${d.kwh_yield}`).join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", "solar_data.csv");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  } else if (currentGraph.value === 'consumption' && consumptionRef.value) {
    consumptionRef.value.exportData();
  }
};

const fetchLatestData = async () => {
  if (isFetching.value) return;
  
  try {
    isFetching.value = true;
    const panelId = 1;
    const currentStartDate = startDate.value || '2024-12-20';
    const currentEndDate = endDate.value || '2024-12-24';
    
    console.log(`[SolarPage] Fetching data for panel ${panelId} from ${currentStartDate} to ${currentEndDate}`);
    
    const response = await store.user.getPanelYieldData(
      panelId,
      currentStartDate,
      currentEndDate
    );
    const responseData = await response.json();
    
    // Process and deduplicate data
    const yieldData = responseData.yield_data;
    const uniqueData = new Map<string, any>();
    
    // Keep only the latest entry for each date
    yieldData.forEach((item: any) => {
      const dateKey = new Date(item.date).toISOString().split('T')[0];
      uniqueData.set(dateKey, {
      date: new Date(item.date),
      kwh_yield: parseFloat(item.kwh_yield) || 0,
      });
    });
    
    // Convert back to array and sort by date
    const sortedData = Array.from(uniqueData.values())
      .sort((a, b) => a.date.getTime() - b.date.getTime());

    // Filter data for the selected date range
    const filteredData = sortedData.filter(item => {
      const itemDate = item.date.toISOString().split('T')[0];
      return itemDate >= currentStartDate && itemDate <= currentEndDate;
    });

    console.log('[SolarPage] Processed data points:', filteredData.length);
    solarData.value = filteredData;
      
  } catch (error) {
    console.error("[SolarPage] Error fetching updated data:", error);
  } finally {
    isFetching.value = false;
  }
};

// Debounce function to prevent too many updates
const debounce = (fn: Function, delay: number) => {
  let timeoutId: number;
  return (...args: any[]) => {
    clearTimeout(timeoutId);
    timeoutId = window.setTimeout(() => fn(...args), delay);
  };
};

// Debounced fetch function
const debouncedFetch = debounce(fetchLatestData, 500);
        
// Watch for changes in the date range
watch([startDate, endDate], async ([newStartDate, newEndDate]) => {
  if (newStartDate && newEndDate) {
    console.log(`[SolarPage] Date range changed to ${newStartDate} - ${newEndDate}`);
    if (currentGraph.value === 'yield') {
      await debouncedFetch();
    }
  }
}, { deep: true });

// Add quick filter handler
const handleQuickFilter = (filterType: '7days' | '7months') => {
  console.log(`[SolarPage] Quick filter selected: ${filterType}`);
  // Functionality will be implemented later
};

onMounted(async () => {
  // Set initial dates
  startDate.value = '2024-12-20';
  endDate.value = '2024-12-24';
  
  await fetchLatestData();
  currentGraph.value = 'yield';

  // Only set up refresh interval for yield graph
  refreshInterval.value = window.setInterval(() => {
    if (currentGraph.value === 'yield') {
      fetchLatestData();
    }
  }, 10000);
});

onUnmounted(() => {
  if (refreshInterval.value !== null) {
    clearInterval(refreshInterval.value);
  }
});
</script>

<style scoped>
.top-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: #000000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}
.nav-button {
  background: transparent;
  border: none;
  color: #ffffff;
  font-size: 16px;
  cursor: pointer;
  transition: color 0.2s;
}
.nav-button:hover {
  color: #007bff;
}
.solar-page-layout {
  display: flex;
  height: 100vh;
  margin-top: 60px;
}
.sidebar {
  width: 220px;
  background: #f5f7fa;
  padding: 24px 12px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  border-right: 1px solid #e0e0e0;
}
.sidebar button {
  width: 100%;
  padding: 10px;
  margin-bottom: 4px;
  border: none;
  background: #e9ecef;
  color: #333;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s;
}
.sidebar button.active,
.sidebar button:hover {
  background: #007bff;
  color: #fff;
}
.date-range {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 12px;
}
.main-content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
}
.quick-filters {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 16px;
}
.filter-btn {
  width: 100%;
  padding: 8px;
  background: #e9ecef;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}
.filter-btn:hover {
  background: #007bff;
  color: #fff;
}
</style>
