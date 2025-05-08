<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar" :class="{ 'navbar--hidden': !navbarVisible }">
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
              type="date"
              v-model="queryParams.startDate"
              @change="updateGraphData"
              :class="{ 'is-default-date': isStartDateInitialDefault }"
            >
          </div>
          <div class="input-group">
            <label>End Date</label>
            <input 
              type="date"
              v-model="queryParams.endDate"
              @change="updateGraphData"
              :class="{ 'is-default-date': isEndDateInitialDefault }"
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
        <h3>Selected Sensor</h3>
        <div class="sensor-slideshow">
          <button class="slideshow-btn prev" @click="prevSensor">&lt;</button>
          <span class="slideshow-label sensor-name">{{ currentSensorName }}</span>
          <button class="slideshow-btn next" @click="nextSensor">&gt;</button>
        </div>
      </div>

      <div class="panel-section">
        <h3>Data Resolution</h3>
        <div class="resolution-slideshow">
          <button class="slideshow-btn prev" @click="prevResolution">&lt;</button>
          <span class="slideshow-label">{{ currentResolutionLabel }}</span>
          <button class="slideshow-btn next" @click="nextResolution">&gt;</button>
        </div>
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

      <button class="export-btn" @click="exportToCSV">
        Export to CSV
      </button>
    </div>

    <!-- Main Content Area -->
    <div class="sensors-content">
    <div class="timer">{{ formattedTime }}</div>
      <SoilMoistureGraph 
        ref="graphComponent"
        :queryParams="queryParams" 
        :sensorVisibility="sensorVisibility" 
        :dynamicTimeWindow="dynamicTimeWindow"
      />
      
      <!-- Real-time data display -->
      <div class="realtime-container">
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
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import SoilMoistureGraph from './soilMoistureGraph.vue';
import { useNexusStore } from '@/stores/nexus';

const router = useRouter();
const goTo = (path: string) => {
  router.push(path);
};

// Helper function to format Date to YYYY-MM-DD string
const toLocalDateString = (date: Date): string => {
  const YYYY = date.getFullYear();
  const MM = String(date.getMonth() + 1).padStart(2, '0');
  const DD = String(date.getDate()).padStart(2, '0');
  return `${YYYY}-${MM}-${DD}`;
};

// Helper function to get day with ordinal suffix (st, nd, rd, th)
const getDayWithSuffix = (day: number): string => {
  if (day > 3 && day < 21) return `${day}th`; // Covers 11th, 12th, 13th
  switch (day % 10) {
    case 1: return `${day}st`;
    case 2: return `${day}nd`;
    case 3: return `${day}rd`;
    default: return `${day}th`;
  }
};

// --- New logic for default date check ---
const initialDefaultDateString = toLocalDateString(new Date()); // Store today's date string at component setup
// --- End new logic ---

// --- Add dynamicTimeWindow ---
const dynamicTimeWindow = ref<'none' | 'lastHour' | 'last24Hours' | 'last7Days' | 'last30Days'>('none');
// --- End Add dynamicTimeWindow ---

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
  const now = currentTime.value; // Use the reactive currentTime ref
  const month = now.toLocaleString('en-US', { month: 'long' }); // e.g., "May"
  const dayWithSuffix = getDayWithSuffix(now.getDate()); // e.g., "8th"
  const dayOfWeek = now.toLocaleString('en-US', { weekday: 'long' }); // e.g., "Tuesday"
  const time = now.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true }); // e.g., "1:51 PM"

  return `${month} ${dayWithSuffix} ${dayOfWeek}, ${time}`; // Combine them
});

// --- Navbar Scroll Behavior --- 
const lastScrollY = ref(0);
const navbarVisible = ref(true);
const scrollThreshold = 50; // Pixels to scroll before hiding starts

const handleScroll = () => {
  const currentScrollY = window.scrollY;
  
  if (currentScrollY < scrollThreshold) { // Always show near top
    navbarVisible.value = true;
  } else if (currentScrollY > lastScrollY.value) { // Scrolling Down
    navbarVisible.value = false;
  } else { // Scrolling Up
    navbarVisible.value = true;
  }
  
  // Update last scroll position, but only if difference is significant to avoid toggling on tiny scrolls
  if (Math.abs(currentScrollY - lastScrollY.value) > 10) {
     lastScrollY.value = currentScrollY;
  }
};
// --- End Navbar Scroll Behavior ---

// Timer and data update setup
let timeInterval: number;
let dataInterval: number;
let graphRefreshInterval: number; // Interval for refreshing the main graph

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

  fetchAndUpdateRealtimeSensor(); // Initial fetch for the small real-time card
  dataInterval = setInterval(fetchAndUpdateRealtimeSensor, 5000); // Fetch every 5 seconds for the card

  // Periodically tell the graph to re-fetch its data
  const GRAPH_REFRESH_INTERVAL_MS = 1000; // Changed to refresh every 1 second
  graphRefreshInterval = setInterval(() => {
    if (graphComponent.value) {
      // console.log('[Sensors.vue] Triggering graph data refresh...'); // Optional: Comment out for less console noise
      graphComponent.value.fetchAllSensorData();
    }
  }, GRAPH_REFRESH_INTERVAL_MS);

  // Add scroll listener
  window.addEventListener('scroll', handleScroll, { passive: true });
  lastScrollY.value = window.scrollY;
});

onUnmounted(() => {
  clearInterval(timeInterval);
  clearInterval(dataInterval);
  clearInterval(graphRefreshInterval); // Clear the graph refresh interval

  // Remove scroll listener
  window.removeEventListener('scroll', handleScroll);
});

interface QueryParams {
  startDate: string;
  endDate: string;
  minMoisture: number;
  maxMoisture: number;
  resolution: 'raw' | 'hourly' | 'daily' | 'weekly' | 'monthly';
}

// --- Define Sensors for Slideshow (Matching soilMoistureGraph.vue) ---
// TODO: Centralize this configuration later
const SENSOR_CONFIGS = [
  { id: 444574498032128, name: 'Sensor Alpha' },
  // { id: 2, name: 'Sensor 2' }, // Keep commented out if still desired
  // { id: 3, name: 'Sensor 3' },
  // { id: 4, name: 'Sensor 4' },
];
const activeSensorIndex = ref(0); // Start with the first sensor
// --- End Sensor Definitions ---

// --- Define Resolutions for Slider ---
const resolutionOptions = [
  { value: 'raw', label: 'Raw Data' },
  { value: 'hourly', label: 'Hourly Average' },
  { value: 'daily', label: 'Daily Average' },
  { value: 'weekly', label: 'Weekly Average' },
  { value: 'monthly', label: 'Monthly Average' },
];
const defaultResolutionValue = 'hourly'; // Keep default consistent
// Find the index of the default resolution
const initialResolutionIndex = resolutionOptions.findIndex(opt => opt.value === defaultResolutionValue);
const resolutionIndex = ref(initialResolutionIndex >= 0 ? initialResolutionIndex : 0); // Slider v-model (0-4)
// --- End Resolutions for Slider ---

const queryParams = ref<QueryParams>({
  startDate: initialDefaultDateString,
  endDate: initialDefaultDateString,
  minMoisture: 0,
  maxMoisture: 100,
  // Initialize resolution based on the default string value
  resolution: defaultResolutionValue as QueryParams['resolution'] 
});

// Computed property to get the label of the currently selected resolution
const currentResolutionLabel = computed(() => {
  return resolutionOptions[resolutionIndex.value]?.label || 'N/A';
});

// Watch the slider index and update the queryParams string value
watch(resolutionIndex, (newIndex) => {
  const selectedOption = resolutionOptions[newIndex];
  if (selectedOption && queryParams.value.resolution !== selectedOption.value) {
    console.log(`[Sensors.vue] Slideshow changed. Index: ${newIndex}, New Resolution: ${selectedOption.value}`);
    
    const newResolution = selectedOption.value as QueryParams['resolution'];
    let newStartDate = queryParams.value.startDate;
    let newEndDate = queryParams.value.endDate;

    // Reset date range if switching TO Raw or Hourly
    if (newResolution === 'raw' || newResolution === 'hourly') {
      const todayStr = toLocalDateString(new Date());
      if (newStartDate !== todayStr || newEndDate !== todayStr) {
         console.log('[Sensors.vue] Resetting date range to Today for Raw/Hourly resolution.');
         newStartDate = todayStr;
         newEndDate = todayStr;
      }
    }

    // Assign a completely new object to ensure reactivity
    queryParams.value = {
      ...queryParams.value,
      resolution: newResolution,
      startDate: newStartDate,
      endDate: newEndDate
    };

    // --- ADDED: Explicitly trigger child update ---
    nextTick(() => { // Wait for DOM update cycle / state propagation
      console.log('[Sensors.vue] Inside nextTick. Attempting to call processAndDrawChart...');
      if (graphComponent.value) {
        console.log('[Sensors.vue] Directly triggering processAndDrawChart on graph component...');
        graphComponent.value.processAndDrawChart(); 
      } else {
         console.warn('[Sensors.vue] Could not find graphComponent ref inside nextTick to trigger update.');
      }
    });
    // --- END ADDED ---
  }
});

// --- New logic for default date check ---
// Computed properties to check if the current value matches the INITIAL default
const isStartDateInitialDefault = computed(() => {
  return queryParams.value.startDate === initialDefaultDateString;
});

const isEndDateInitialDefault = computed(() => {
  return queryParams.value.endDate === initialDefaultDateString;
});
// --- End new logic ---

const setTimeRange = (range: string) => {
  const now = new Date();
  let start = new Date(); // Will be used to set queryParams.startDate
  let end = new Date();   // Will be used to set queryParams.endDate

  switch (range) {
    case '1h':
      dynamicTimeWindow.value = 'lastHour';
      start = now; // Date picker shows today
      end = now;   // Date picker shows today
      break;
    case '24h':
      dynamicTimeWindow.value = 'last24Hours';
      start = new Date(now.getTime() - 24 * 60 * 60 * 1000); // e.g. yesterday 10am if now is today 10am
      end = now; // Date picker end is today
      break;
    case '7d':
      dynamicTimeWindow.value = 'last7Days';
      start = new Date(now.getTime() - 6 * 24 * 60 * 60 * 1000); // 6 days ago to include today in picker
      end = now;
      break;
    case '30d':
      dynamicTimeWindow.value = 'last30Days';
      start = new Date(now.getTime() - 29 * 24 * 60 * 60 * 1000); // 29 days ago to include today in picker
      end = now;
      break;
  }

  queryParams.value.startDate = toLocalDateString(start);
  queryParams.value.endDate = toLocalDateString(end);
  // No need to call updateGraphData() here as queryParams watcher in graph + dynamicTimeWindow watcher will handle it.
  // However, to ensure reactivity if only queryParams changes without dynamicTimeWindow changing *from a non-'none'* state,
  // we might still need to ensure queryParams object itself is new if we weren't relying on dynamicTimeWindow.
  // Let's explicitly assign a new object for queryParams to be safe and consistent.
  queryParams.value = { ...queryParams.value };

  // If the graph component needs to be explicitly told to process after these changes:
  nextTick(() => {
    if (graphComponent.value) {
      // The graph will react to queryParams and dynamicTimeWindow prop changes.
      // A direct call to processAndDrawChart might be redundant if watchers are correctly set up in child.
      // For now, relying on prop reactivity.
      // graphComponent.value.processAndDrawChart(); 
      console.log('[Sensors.vue] setTimeRange updated queryParams and dynamicTimeWindow.');
    }
  });
};

const updateGraphData = () => {
  dynamicTimeWindow.value = 'none'; // Reset dynamic window when manually changing dates/filters
  // This will trigger the graph update through props
  queryParams.value = { ...queryParams.value };
  // Similar to setTimeRange, relying on prop reactivity in the child.
  nextTick(() => {
    if (graphComponent.value) {
       // graphComponent.value.processAndDrawChart();
       console.log('[Sensors.vue] updateGraphData updated queryParams and reset dynamicTimeWindow.');
    }
  });
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

// --- Methods for Slideshow Navigation ---
const nextResolution = () => {
  resolutionIndex.value = (resolutionIndex.value + 1) % resolutionOptions.length;
};

const prevResolution = () => {
  resolutionIndex.value = (resolutionIndex.value - 1 + resolutionOptions.length) % resolutionOptions.length;
};

const nextSensor = () => {
  if (SENSOR_CONFIGS.length > 0) { // Prevent error if array is empty
    activeSensorIndex.value = (activeSensorIndex.value + 1) % SENSOR_CONFIGS.length;
  }
};
const prevSensor = () => {
  if (SENSOR_CONFIGS.length > 0) { // Prevent error if array is empty
    activeSensorIndex.value = (activeSensorIndex.value - 1 + SENSOR_CONFIGS.length) % SENSOR_CONFIGS.length;
  }
};

// --- Computed Properties --- 
const currentSensorName = computed(() => {
  // Make sure SENSOR_CONFIGS is not empty
  return SENSOR_CONFIGS.length > 0 ? SENSOR_CONFIGS[activeSensorIndex.value]?.name : 'No Sensors';
});

// Computed property to generate visibility object based on active index
const sensorVisibility = computed(() => {
  const visibility: { [key: string]: boolean } = {};
  SENSOR_CONFIGS.forEach((config, index) => {
    visibility[config.name] = (index === activeSensorIndex.value);
  });
  console.log("[Sensors.vue] Computed sensorVisibility:", visibility);
  return visibility;
});

// --- End Computed Properties ---

// Watcher for sensor changes (might be needed later to update graph prop)
watch(activeSensorIndex, (newIndex) => {
  console.log(`[Sensors.vue] Sensor changed. Index: ${newIndex}, Name: ${currentSensorName.value}`);
  // TODO: Update graph visibility prop based on this index
});
</script>

<style>
.navbar {
  position: fixed;
  left: 0;
  top: 0;
  width: 100%;
  display: flex;
  justify-content: space-around;
  background-color: black;
  padding: 10px 20px;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease-in-out;
  transform: translateY(0);
}

.navbar.navbar--hidden {
  transform: translateY(-100%);
}

.nav-button {
  background-color: transparent;
  border: none;
  padding: 10px 20px;
  color: white;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s, color 0.3s, transform 0.3s ease;
}

.nav-button:hover {
  background-color: #333;
  color: #eee;
  transform: translateY(-2px);
}

.nav-button:active {
  transform: translateY(1px);
  transition: transform 0.1s;
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

/* Add this rule */
.input-group input[type="date"].is-default-date {
  color: grey;
  /* You could also use opacity: 0.7; or font-style: italic; */
}

/* Optional: Remove grey style when user focuses the input */
.input-group input[type="date"].is-default-date:focus {
  color: initial; /* Reverts to browser default or inherited color */
}

/* Add styles for the slideshow control */
.resolution-slideshow {
  display: flex;
  align-items: center; /* Vertically align buttons and label */
  justify-content: space-between; /* Space out buttons and label */
  border: 1px solid #dee2e6; /* Optional border */
  border-radius: 4px;
  padding: 5px 10px; /* Add some padding */
  background-color: white; /* Optional background */
}

.slideshow-btn {
  background: transparent;
  border: none;
  font-size: 1.5em; /* Make arrows bigger */
  cursor: pointer;
  color: #555;
  padding: 0 5px; /* Give arrows some space */
  line-height: 1; /* Adjust line height for vertical centering */
  transition: color 0.2s;
}

.slideshow-btn:hover {
  color: #0056b3;
}

.slideshow-label {
  font-size: 0.9em;
  color: #333; /* Make label text darker */
  font-weight: 500; /* Slightly bolder */
  text-align: center;
  flex-grow: 1; /* Allow label to take up space */
  margin: 0 5px; /* Add margin around label */
}

/* Add styles for the sensor slideshow control */
.sensor-slideshow {
  display: flex;
  align-items: center; 
  justify-content: space-between; 
  border: 1px solid #dee2e6; 
  border-radius: 4px;
  padding: 5px 10px; 
  background-color: white; 
}

.sensor-slideshow .slideshow-label.sensor-name {
  font-size: 0.95em; /* Slightly larger for sensor name maybe */
  color: #1f77b4; /* Match first sensor color? Or keep #333 */
  font-weight: 600; 
  text-align: center;
  flex-grow: 1; 
  margin: 0 5px; 
}
</style>