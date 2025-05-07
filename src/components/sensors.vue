<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar">
      <button class="nav-button" @click="goTo('/home')">Home</button>
    </nav>
  </div>
  
  <div class="main-content">
    <!-- Query Side Panel -->
    <div class="query-panel">
      <div class="panel-section">
        <h3>Date Range</h3>
        <div class="date-inputs">
          <div class="input-group">
            <label>Start Date</label>
            <input 
              type="datetime-local" 
              v-model="queryParams.startDate"
              @change="updateGraphData"
            >
          </div>
          <div class="input-group">
            <label>End Date</label>
            <input 
              type="datetime-local" 
              v-model="queryParams.endDate"
              @change="updateGraphData"
            >
          </div>
        </div>
      </div>

      <div class="panel-section">
        <h3>Data Range</h3>
        <div class="range-inputs">
          <div class="input-group">
            <label>Min Moisture (%)</label>
            <input 
              type="number" 
              v-model="queryParams.minMoisture"
              min="0"
              max="100"
              @change="updateGraphData"
            >
          </div>
          <div class="input-group">
            <label>Max Moisture (%)</label>
            <input 
              type="number" 
              v-model="queryParams.maxMoisture"
              min="0"
              max="100"
              @change="updateGraphData"
            >
          </div>
        </div>
      </div>

      <div class="panel-section">
        <h3>Data Resolution</h3>
        <select v-model="queryParams.resolution" @change="updateGraphData">
          <option value="raw">Raw Data</option>
          <option value="hourly">Hourly Average</option>
          <option value="daily">Daily Average</option>
          <option value="weekly">Weekly Average</option>
        </select>
      </div>

      <div class="panel-section">
        <h3>Quick Filters</h3>
        <div class="quick-filters">
          <button @click="setTimeRange('1h')">Last Hour</button>
          <button @click="setTimeRange('24h')">Last 24 Hours</button>
          <button @click="setTimeRange('7d')">Last 7 Days</button>
          <button @click="setTimeRange('30d')">Last 30 Days</button>
        </div>
      </div>

      <button class="apply-btn" @click="updateGraphData">
        Apply Filters
      </button>
      
      <button class="export-btn" @click="exportToCSV">
        Export to CSV
      </button>
    </div>

    <!-- Main Content Area -->
    <div class="sensors-content">
      <SoilMoistureGraph 
        ref="graphComponent"
        :queryParams="queryParams" 
      />
      
      <!-- Real-time data display -->
      <div class="realtime-container">
        <div class="timer">{{ formattedTime }}</div>
        
        <div class="sensor-carousel">
          <div class="sensor-card" :style="{ '--sensor-color': colors[0] }">
            <h3>{{ currentRealtimeSensorData.name }}</h3>
            <div class="sensor-value">
              {{ currentRealtimeSensorData.moisture !== null ? currentRealtimeSensorData.moisture.toFixed(1) + '%' : 'Loading...' }}
            </div>
            <div class="sensor-time">
              Last updated: {{ currentRealtimeSensorData.lastUpdated }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import SoilMoistureGraph from './soilMoistureGraph.vue';
import { useNexusStore } from '@/stores/nexus';

const router = useRouter();
const goTo = (path: string) => {
  router.push(path);
};

// Real-time data setup
const currentTime = ref(new Date());
const currentSensorIndex = ref(0);
const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

const nexusStore = useNexusStore();
const REALTIME_SENSOR_ID = 444574498032128;

interface RealtimeSensorDisplay {
  name: string;
  moisture: number | null;
  lastUpdated: string;
}

const currentRealtimeSensorData = ref<RealtimeSensorDisplay>({
  name: 'Sensor Alpha',
  moisture: null,
  lastUpdated: 'N/A',
});

// Computed properties
const formattedTime = computed(() => {
  return currentTime.value.toLocaleTimeString();
});

// Timer and data update setup
let timeInterval: number;
let dataInterval: number;

onMounted(() => {
  // Update time every second
  timeInterval = setInterval(() => {
    currentTime.value = new Date();
  }, 1000);

  const fetchAndUpdateRealtimeSensor = async () => {
    try {
      // getSensorMoistureData returns: { id, sensor_id, date (string), soil_moisture (number) }[]
      const moistureDataArray = await nexusStore.user.getSensorMoistureData(REALTIME_SENSOR_ID);
      if (moistureDataArray && moistureDataArray.length > 0) {
        // Sort by date descending to get the latest
        // Ensure date strings are properly converted to Date objects for sorting
        const sortedData = [...moistureDataArray].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
        const latestReading = sortedData[0];
        currentRealtimeSensorData.value = {
          name: 'Sensor Alpha',
          moisture: Number(latestReading.soil_moisture),
          lastUpdated: new Date(latestReading.date).toLocaleTimeString(),
        };
      } else {
        currentRealtimeSensorData.value.moisture = null;
        currentRealtimeSensorData.value.lastUpdated = new Date().toLocaleTimeString() + ' (No data)';
      }
    } catch (error) {
      console.error('Error fetching real-time sensor data for ID', REALTIME_SENSOR_ID, ':', error);
      currentRealtimeSensorData.value.moisture = null;
      currentRealtimeSensorData.value.lastUpdated = new Date().toLocaleTimeString() + ' (Error)';
    }
  };

  fetchAndUpdateRealtimeSensor(); // Initial fetch
  dataInterval = setInterval(fetchAndUpdateRealtimeSensor, 5000); // Fetch every 5 seconds
});

onUnmounted(() => {
  clearInterval(timeInterval);
  clearInterval(dataInterval);
});

interface QueryParams {
  startDate: string;
  endDate: string;
  minMoisture: number;
  maxMoisture: number;
  resolution: 'raw' | 'hourly' | 'daily' | 'weekly';
}

const queryParams = ref<QueryParams>({
  startDate: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().slice(0, 16),
  endDate: new Date().toISOString().slice(0, 16),
  minMoisture: 0,
  maxMoisture: 100,
  resolution: 'raw'
});

const setTimeRange = (range: string) => {
  const now = new Date();
  let start = new Date();

  switch (range) {
    case '1h':
      start = new Date(now.getTime() - 60 * 60 * 1000);
      break;
    case '24h':
      start = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      break;
    case '7d':
      start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
      break;
    case '30d':
      start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
      break;
  }

  queryParams.value.startDate = start.toISOString().slice(0, 16);
  queryParams.value.endDate = now.toISOString().slice(0, 16);
  updateGraphData();
};

const updateGraphData = () => {
  // This will trigger the graph update through props
  queryParams.value = { ...queryParams.value };
};

// Add ref for the graph component
const graphComponent = ref<InstanceType<typeof SoilMoistureGraph> | null>(null);

const exportToCSV = () => {
  const data = graphComponent.value?.getFilteredData();
  
  if (!data || !data.length) {
    alert('No data available to export');
    return;
  }

  // Format the data for CSV
  const headers = ['Timestamp', 'Sensor', 'Moisture (%)'];
  const csvRows = [headers];

  data.forEach(sensor => {
    if (sensor.visible) { // Only export visible sensors
      sensor.data.forEach(point => {
        csvRows.push([
          new Date(point.time).toISOString(),
          sensor.name,
          point.moisture.toFixed(1)
        ]);
      });
    }
  });

  // Create CSV content
  const csvContent = csvRows
    .map(row => row.join(','))
    .join('\n');

  // Create and trigger download
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  
  link.setAttribute('href', url);
  link.setAttribute('download', `sensor_data_${new Date().toISOString().slice(0,19).replace(/[:]/g, '-')}.csv`);
  
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url); // Clean up the URL object
};
</script>

<style>
.navbar {
  position: fixed;
  left: 0;
  top: 0;
  width: 100%;
  display: flex;
  justify-content: space-around;
  background-color: rgba(255, 255, 255, 0.9); /* Slight transparency */
  backdrop-filter: blur(10px); /* Blur for smooth background */
  padding: 10px 20px;
  z-index: 1000; /* Ensures navbar stays on top of other content */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); /* Add subtle shadow */
  transition: background-color 0.3s ease, box-shadow 0.3s ease;
}

.navbar:hover {
  background-color: #fafafa; /* Slight color change on hover */
}

.nav-button {
  background-color: transparent;
  border: none;
  padding: 10px 20px;
  color: #333;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s, color 0.3s, transform 0.3s ease; /* Smooth transition for hover and active states */
}

.nav-button:hover {
  background-color: #f0f0f0; /* Slight background change */
  color: #0056b3; /* Hover color */
  transform: translateY(-2px); /* Subtle lift on hover */
}

.nav-button:active {
  transform: translateY(1px); /* Lower it when pressed */
  transition: transform 0.1s; /* Faster response on active */
}

/* Navbar adjustments on scroll for a smoother feel */
.navbar.scrolled {
  background-color: rgba(255, 255, 255, 1); /* Fully opaque when scrolled */
  box-shadow: 0 6px 15px rgba(0, 0, 0, 0.2); /* More shadow when scrolled */
}

.main-content {
  display: flex;
  margin-top: 60px;
  min-height: calc(100vh - 60px);
}

.query-panel {
  width: 300px;
  background: #f8f9fa;
  padding: 20px;
  border-right: 1px solid #dee2e6;
  height: calc(100vh - 60px);
  position: fixed;
  overflow-y: auto;
}

.panel-section {
  margin-bottom: 24px;
}

.panel-section h3 {
  margin: 0 0 12px 0;
  font-size: 1rem;
  color: #333;
}

.date-inputs, .range-inputs {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.input-group label {
  font-size: 0.875rem;
  color: #666;
}

.input-group input,
select {
  padding: 8px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  font-size: 0.875rem;
}

.quick-filters {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.quick-filters button {
  padding: 8px;
  background: white;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-filters button:hover {
  background: #f0f0f0;
  border-color: #adb5bd;
}

.apply-btn {
  width: 100%;
  padding: 12px;
  background: #0056b3;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s;
  margin-bottom: 0;
}

.apply-btn:hover {
  background: #004494;
}

.export-btn {
  width: 100%;
  padding: 12px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s;
  margin-top: 12px;
}

.export-btn:hover {
  background: #218838;
}

.sensors-content {
  margin-left: 300px; /* Match query-panel width */
  flex-grow: 1;
  padding: 20px;
}

.realtime-container {
  margin-top: 40px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.timer {
  text-align: center;
  font-size: 1.5rem;
  font-weight: bold;
  margin-bottom: 20px;
  color: #333;
}

.sensor-carousel {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
  margin-bottom: 20px;
}

.sensor-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid var(--sensor-color);
  min-width: 200px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.sensor-card h3 {
  margin: 0 0 10px 0;
  color: var(--sensor-color);
}

.sensor-value {
  font-size: 2rem;
  font-weight: bold;
  margin: 10px 0;
}

.sensor-time {
  font-size: 0.8rem;
  color: #666;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
  }

  .query-panel {
    width: 100%;
    height: auto;
    position: static;
    border-right: none;
    border-bottom: 1px solid #dee2e6;
  }

  .sensors-content {
    margin-left: 0;
  }
}
</style>