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
        <h3>Selected Sensors</h3>
        <div class="sensor-selection">
          <div v-if="sensorsLoading" class="loading-state">
            Loading sensors...
          </div>
          <div v-else-if="sensorsError" class="error-state">
            {{ sensorsError }}
            <button @click="fetchSensors" class="retry-btn">Retry</button>
          </div>
          <div v-else-if="SENSOR_CONFIGS.length === 0" class="no-sensors-state">
            No sensors available
          </div>
          <div v-else>
            <div 
              v-for="(sensor, index) in SENSOR_CONFIGS" 
              :key="sensor.id"
              class="sensor-checkbox"
            >
              <label :style="{ '--sensor-color': colors[index % colors.length] }">
                <input 
                  type="checkbox" 
                  :checked="sensorVisibility[sensor.name]" 
                  @change="toggleSensor(sensor.name)"
                >
                <span class="sensor-name">{{ sensor.name }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      
      <div class="panel-section">
        <h3>Quick Filters</h3>
        <div class="quick-filters">
          <button @click="setTimeRange('1h')">Last Hour</button>
          <button @click="setTimeRange('24h')">Last 24 Hours</button>
          <button @click="setTimeRange('7d')">Last 7 Days</button>
          <button @click="setTimeRange('30d')">Last 30 Days</button>
          <button @click="resetToDefault" class="reset-btn">Reset to Default</button>
        </div>
      </div>

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
        <h3>Sensor Types</h3>
        <div class="sensor-type-slideshow">
          <button class="slideshow-btn prev" @click="prevSensorType">&lt;</button>
          <span class="slideshow-label">{{ currentSensorTypeLabel }}</span>
          <button class="slideshow-btn next" @click="nextSensorType">&gt;</button>
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
        <h3>Data Range</h3>
        <div class="range-inputs">
          <div class="input-group">
            <label>{{ currentSensorType === 'moisture' ? 'Min Moisture (%)' : 'Min Temperature (°C)' }}</label>
            <input 
              type="number" 
              v-model="queryParams.minValue"
              :min="currentSensorType === 'moisture' ? 0 : -10"
              :max="currentSensorType === 'moisture' ? 100 : 50"
              @change="updateGraphData"
            >
          </div>
          <div class="input-group">
            <label>{{ currentSensorType === 'moisture' ? 'Max Moisture (%)' : 'Max Temperature (°C)' }}</label>
            <input 
              type="number" 
              v-model="queryParams.maxValue"
              :min="currentSensorType === 'moisture' ? 0 : -10"
              :max="currentSensorType === 'moisture' ? 100 : 50"
              @change="updateGraphData"
            >
          </div>
        </div>
      </div>

      <button class="export-btn" @click="exportToCSV">
        Export to CSV
      </button>
    </div>

    <!-- Main Content Area -->
    <div class="sensors-content">
      <!-- Real-time data display -->
      <div class="realtime-container">
        <div class="sensor-carousel">
          <div class="sensor-card" :style="{ '--sensor-color': colors[0] }">
            <!-- Sensor Name and Trend Indicator -->
            <div class="sensor-card-header">
              <h3>{{ currentRealtimeSensorData.name }}</h3>
              <!-- Add the SVG trend indicator here -->
              <svg class="trend-indicator" width="40" height="20" viewBox="0 0 40 20">
                <path 
                  :d="trendPathData" 
                  fill="none" 
                  :stroke="colors[0]" 
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </div>
            <!-- Sensor Value -->
            <div class="sensor-value">
              {{ 
                currentRealtimeSensorData.value !== null 
                  ? (
                      currentSensorType === 'temperature' 
                      ? celsiusToFahrenheit(currentRealtimeSensorData.value)?.toFixed(1) + '°F' 
                      : currentRealtimeSensorData.value.toFixed(1) + '%'
                    )
                  : 'Loading...' 
              }}
            </div>
             <!-- Last Updated Time -->
            <div class="sensor-time">
              Last updated: {{ currentRealtimeSensorData.lastUpdated }}
            </div>
          </div>
        </div>
      </div>

      <SoilMoistureGraph 
        v-if="currentSensorType === 'moisture'"
        ref="moistureGraphComponent"
        :queryParams="moistureQueryParams" 
        :sensorVisibility="sensorVisibility" 
        :dynamicTimeWindow="dynamicTimeWindow"
        :dataType="currentSensorType"
        :sensorConfigs="SENSOR_CONFIGS"
      />
      <SoilTemperatureGraph
        v-else
        ref="temperatureGraphComponent"
        :queryParams="temperatureQueryParams"
        :sensorVisibility="sensorVisibility"
        :dynamicTimeWindow="dynamicTimeWindow"
        :dataType="currentSensorType"
        :sensorConfigs="SENSOR_CONFIGS"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import SoilMoistureGraph from './soilMoistureGraph.vue';
import SoilTemperatureGraph from './soilTemperatureGraph.vue';
import { useNexusStore } from '@/stores/nexus';

interface SensorConfig {
  id: string;
  name: string;
}

interface DataPoint {
  time: Date;
  moisture?: number;
  temperature?: number;
  sensorId?: string;
}

// Add conversion function
const celsiusToFahrenheit = (celsius: number | null): number | null => {
  if (celsius === null) return null;
  return (celsius * 9/5) + 32;
};

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
// Convert decimal 444574498032128 to hex string
const REALTIME_SENSOR_ID = "2CF7F1C0627000C4";

interface RealtimeSensorDisplay {
  name: string;
  value: number | null;
  lastUpdated: string;
}

const currentRealtimeSensorData = ref<RealtimeSensorDisplay>({
  name: 'Sensor Alpha',
  value: null,
  lastUpdated: 'N/A',
});

// --- Add ref to store the previous value --- 
const previousRealtimeSensorValue = ref<number | null>(null);

// --- Computed property for the trend --- 
const realtimeTrend = computed((): 'up' | 'down' | 'stable' => {
  if (currentRealtimeSensorData.value.value === null || previousRealtimeSensorValue.value === null) {
    return 'stable'; // Default to stable if no data or no previous data
  }
  if (currentRealtimeSensorData.value.value > previousRealtimeSensorValue.value) {
    return 'up';
  }
  if (currentRealtimeSensorData.value.value < previousRealtimeSensorValue.value) {
    return 'down';
  }
  return 'stable';
});

// --- Computed property for the SVG path data based on trend --- 
const trendPathData = computed(() => {
  switch (realtimeTrend.value) {
    case 'up':
      return 'M5,15 L20,5 L35,15'; // Simple upward arrow/line
    case 'down':
      return 'M5,5 L20,15 L35,5'; // Simple downward arrow/line
    default:
      return 'M5,10 L35,10'; // Simple horizontal line
  }
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

// --- Define Sensors for Slideshow (Dynamic from API) ---
const SENSOR_CONFIGS = ref<SensorConfig[]>([]);
const sensorsLoading = ref(true);
const sensorsError = ref<string | null>(null);

// Changed from activeSensorIndex to a direct sensorVisibility object
// Initialize with first sensor enabled
const sensorVisibility = ref<{ [key: string]: boolean }>({});

// Function to fetch sensors from the API
const fetchSensors = async () => {
  try {
    sensorsLoading.value = true;
    sensorsError.value = null;
    const sensors = await nexusStore.user.getAllSensors();
    console.log('[sensors.vue] Fetched sensors:', JSON.stringify(sensors, null, 2));
    
    if (Array.isArray(sensors) && sensors.length > 0) {
      SENSOR_CONFIGS.value = sensors.map((sensor: any) => ({
        id: sensor.id,
        name: sensor.name || `Sensor ${sensor.id}`,
      }));
      
      // Initialize visibility state for all sensors
      // By default, enable only the first sensor
      SENSOR_CONFIGS.value.forEach((sensor, index) => {
        sensorVisibility.value[sensor.name] = index === 0;
      });
    } else {
      sensorsError.value = 'No sensors found';
      SENSOR_CONFIGS.value = [];
    }
  } catch (error) {
    console.error('Error fetching sensors:', error);
    sensorsError.value = 'Failed to load sensors';
    SENSOR_CONFIGS.value = [];
  } finally {
    sensorsLoading.value = false;
  }
};

onMounted(async () => {
  await fetchSensors();
  
  // Update time every second
  timeInterval = setInterval(() => {
    currentTime.value = new Date();
  }, 1000);

  const fetchAndUpdateRealtimeSensor = async () => {
    try {
      let data;
      if (currentSensorType.value === 'moisture') {
        data = await nexusStore.user.getSensorMoistureData(REALTIME_SENSOR_ID);
      } else {
        data = await nexusStore.user.getSensorTemperatureData(REALTIME_SENSOR_ID);
      }
      
      if (data && data.length > 0) {
        const sortedData = [...data].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
        const latestReading = sortedData[0];
        const newValue = currentSensorType.value === 'moisture' 
            ? Number(latestReading.soil_moisture)
            : Number(latestReading.soil_temperature);

        // --- Store the current value as the previous one BEFORE updating --- 
        previousRealtimeSensorValue.value = currentRealtimeSensorData.value.value; 

        // --- Update the current data --- 
        currentRealtimeSensorData.value = {
          name: 'Sensor Alpha', 
          value: newValue,
          lastUpdated: new Date(latestReading.date).toLocaleTimeString(),
        };

        // --- Handle initial case where previous value was null --- 
        if (previousRealtimeSensorValue.value === null) {
          previousRealtimeSensorValue.value = newValue; // Set initial previous value
        }

      } else {
        // --- Store null as previous if no data --- 
        previousRealtimeSensorValue.value = currentRealtimeSensorData.value.value;
        currentRealtimeSensorData.value.value = null;
        currentRealtimeSensorData.value.lastUpdated = new Date().toLocaleTimeString() + ' (No data)';
      }
    } catch (error) {
      console.error('Error fetching real-time sensor data for ID', REALTIME_SENSOR_ID, ':', error);
       // --- Store null as previous on error --- 
      previousRealtimeSensorValue.value = currentRealtimeSensorData.value.value;
      currentRealtimeSensorData.value.value = null;
      currentRealtimeSensorData.value.lastUpdated = new Date().toLocaleTimeString() + ' (Error)';
    }
  };

  fetchAndUpdateRealtimeSensor(); // Initial fetch
  dataInterval = setInterval(fetchAndUpdateRealtimeSensor, 5000); // Fetch every 5 seconds

  // Periodically tell the graph to re-fetch its data
  const GRAPH_REFRESH_INTERVAL_MS = 1000;
  graphRefreshInterval = setInterval(() => {
    const currentGraph = currentSensorType.value === 'moisture' 
      ? moistureGraphComponent.value 
      : temperatureGraphComponent.value;
    
    if (currentGraph) {
      // Only refresh if we're in raw or hourly resolution
      if (queryParams.value.resolution === 'raw' || queryParams.value.resolution === 'hourly') {
        currentGraph.fetchAllSensorData();
      }
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
  minValue: number;
  maxValue: number;
  resolution: 'raw' | 'hourly' | 'daily' | 'weekly' | 'monthly';
}

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

// --- Define Sensor Types for Slider ---
const sensorTypeOptions = [
  { value: 'moisture', label: 'Soil Moisture' },
  { value: 'temperature', label: 'Soil Temperature' },
];
const defaultSensorTypeValue = 'moisture'; // Default to moisture graph
const initialSensorTypeIndex = sensorTypeOptions.findIndex(opt => opt.value === defaultSensorTypeValue);
const sensorTypeIndex = ref(initialSensorTypeIndex >= 0 ? initialSensorTypeIndex : 0);
// --- End Sensor Types for Slider ---

// Initialize query params for both types
const moistureQueryParams = ref<QueryParams>({
  startDate: initialDefaultDateString,
  endDate: initialDefaultDateString,
  minValue: 0,
  maxValue: 100,
  resolution: defaultResolutionValue as QueryParams['resolution']
});

const temperatureQueryParams = ref<QueryParams>({
  startDate: initialDefaultDateString,
  endDate: initialDefaultDateString,
  minValue: -10,
  maxValue: 50,
  resolution: defaultResolutionValue as QueryParams['resolution']
});

// Computed property to get the current query params based on sensor type
const queryParams = computed(() => {
  return currentSensorType.value === 'moisture' ? moistureQueryParams.value : temperatureQueryParams.value;
});

// Computed property to get the label of the currently selected resolution
const currentResolutionLabel = computed(() => {
  return resolutionOptions[resolutionIndex.value]?.label || 'N/A';
});

// Watch the slider index and update the queryParams string value
watch(resolutionIndex, (newIndex) => {
  const selectedOption = resolutionOptions[newIndex];
  if (selectedOption && queryParams.value.resolution !== selectedOption.value) {
    console.log(`[Sensors.vue] Resolution changed to: ${selectedOption.value}`);
    
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

    // For daily/weekly/monthly, ensure we have enough data
    if (newResolution === 'daily' || newResolution === 'weekly' || newResolution === 'monthly') {
      const startDate = new Date(newStartDate);
      const endDate = new Date(newEndDate);
      const daysDiff = Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));
      
      console.log(`[Sensors.vue] Date range spans ${daysDiff} days for ${newResolution} resolution`);
      
      // Ensure we have enough data points for the selected resolution
      if (newResolution === 'daily' && daysDiff < 2) {
        console.log('[Sensors.vue] Extending date range for daily resolution');
        newStartDate = toLocalDateString(new Date(endDate.getTime() - 7 * 24 * 60 * 60 * 1000));
      } else if (newResolution === 'weekly' && daysDiff < 7) {
        console.log('[Sensors.vue] Extending date range for weekly resolution');
        newStartDate = toLocalDateString(new Date(endDate.getTime() - 30 * 24 * 60 * 60 * 1000));
      } else if (newResolution === 'monthly' && daysDiff < 30) {
        console.log('[Sensors.vue] Extending date range for monthly resolution');
        newStartDate = toLocalDateString(new Date(endDate.getTime() - 90 * 24 * 60 * 60 * 1000));
      }
    }

    // Create new query params object
    const newQueryParams = {
      ...queryParams.value,
      resolution: newResolution,
      startDate: newStartDate,
      endDate: newEndDate
    };

    console.log('[Sensors.vue] Updated queryParams:', newQueryParams);

    // Update the appropriate query params based on sensor type
    if (currentSensorType.value === 'moisture') {
      moistureQueryParams.value = newQueryParams;
    } else {
      temperatureQueryParams.value = newQueryParams;
    }

    // Force a data refresh with the new resolution
    nextTick(() => {
      const currentGraph = currentSensorType.value === 'moisture' 
        ? moistureGraphComponent.value 
        : temperatureGraphComponent.value;
      
      if (currentGraph) {
        console.log('[Sensors.vue] Triggering data refresh with new resolution');
        currentGraph.fetchAllSensorData();
      }
    });
  }
});

// Watch sensor type index to handle changes
watch(sensorTypeIndex, (newIndex) => {
  const selectedType = sensorTypeOptions[newIndex];
  
  console.log(`[Sensors.vue] Sensor type changed to: ${selectedType.label} (${selectedType.value})`);
  
  // For now, this is just a placeholder for future implementation
  // When temperature graph is actually implemented, this is where
  // you would toggle between the different graph types
  
  // TODO: Implement temperature graph and switch between graph types
  // Example logic (commented out until implementation):
  // if (selectedType.value === 'temperature') {
  //   showTemperatureGraph.value = true;
  //   showMoistureGraph.value = false;
  // } else {
  //   showTemperatureGraph.value = false;
  //   showMoistureGraph.value = true;
  // }
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
  let start = new Date(); 
  let end = new Date();   
  let newResolution: QueryParams['resolution'] = queryParams.value.resolution; // Default to current

  switch (range) {
    case '1h':
      dynamicTimeWindow.value = 'lastHour';
      start = now; 
      end = now;   
      newResolution = 'hourly'; // Or 'raw' if preferred for last hour
      break;
    case '24h':
      dynamicTimeWindow.value = 'last24Hours';
      start = new Date(now.getTime() - 24 * 60 * 60 * 1000); 
      end = now; 
      newResolution = 'hourly';
      break;
    case '7d':
      dynamicTimeWindow.value = 'last7Days';
      start = new Date(now.getTime() - 6 * 24 * 60 * 60 * 1000); 
      end = now;
      newResolution = 'daily';
      break;
    case '30d':
      dynamicTimeWindow.value = 'last30Days';
      start = new Date(now.getTime() - 29 * 24 * 60 * 60 * 1000); 
      end = now;
      newResolution = 'daily';
      break;
  }

  // Update queryParams in a way that triggers reactivity
  const newQueryParams: QueryParams = {
    ...queryParams.value, // Spread existing values first
    startDate: toLocalDateString(start),
    endDate: toLocalDateString(end),
    resolution: newResolution,
  };

  // Update the correct reactive ref based on current sensor type
  if (currentSensorType.value === 'moisture') {
    moistureQueryParams.value = newQueryParams;
  } else {
    temperatureQueryParams.value = newQueryParams;
  }
  
  // Update the resolutionIndex ref to match the newResolution
  const newResolutionIndex = resolutionOptions.findIndex(opt => opt.value === newResolution);
  if (newResolutionIndex !== -1 && resolutionIndex.value !== newResolutionIndex) {
    resolutionIndex.value = newResolutionIndex;
  }

  // The watchers for queryParams (implicitly via moistureQueryParams/temperatureQueryParams) 
  // and dynamicTimeWindow in the graph components should handle refreshing the chart.
  console.log('[Sensors.vue] setTimeRange updated queryParams, dynamicTimeWindow, and resolution:', newQueryParams, dynamicTimeWindow.value);

  // Explicitly trigger graph data fetching if needed, though watchers should handle it.
  nextTick(() => {
    const currentGraph = currentSensorType.value === 'moisture' 
      ? moistureGraphComponent.value 
      : temperatureGraphComponent.value;
    if (currentGraph) {
      // currentGraph.fetchAllSensorData(); // Watchers in graph components should trigger this.
    }
  });
};

const updateGraphData = () => {
  dynamicTimeWindow.value = 'none'; // Reset dynamic window when manually changing dates/filters
  // Update the appropriate query params based on sensor type
  if (currentSensorType.value === 'moisture') {
    moistureQueryParams.value = { ...moistureQueryParams.value };
  } else {
    temperatureQueryParams.value = { ...temperatureQueryParams.value };
  }
  // Similar to setTimeRange, relying on prop reactivity in the child.
  nextTick(() => {
    if (currentSensorType.value === 'moisture' && moistureGraphComponent.value) {
       // moistureGraphComponent.value.processAndDrawChart();
       console.log('[Sensors.vue] updateGraphData updated queryParams and reset dynamicTimeWindow.');
    } else if (currentSensorType.value === 'temperature' && temperatureGraphComponent.value) {
      console.log('[Sensors.vue] updateGraphData updated queryParams and reset dynamicTimeWindow.');
    }
  });
};

// Add refs for both graph components
const moistureGraphComponent = ref<InstanceType<typeof SoilMoistureGraph> | null>(null);
const temperatureGraphComponent = ref<InstanceType<typeof SoilTemperatureGraph> | null>(null);

const exportToCSV = () => {
  const graphComponent = currentSensorType.value === 'moisture' 
    ? moistureGraphComponent.value 
    : temperatureGraphComponent.value;
  
  const data = graphComponent?.getFilteredData();
  
  if (!data || !data.length) {
    alert('No data available to export');
    return;
  }

  // Format the data for CSV
  const headers = ['Timestamp', 'Sensor', currentSensorType.value === 'moisture' ? 'Moisture (%)' : 'Temperature (°C)'];
  const csvRows = [headers];

  data.forEach(sensor => {
    if (sensorVisibility.value[sensor.name]) {
      sensor.data.forEach((point: DataPoint) => {
        csvRows.push([
          new Date(point.time).toISOString(),
          sensor.name,
          currentSensorType.value === 'moisture' 
            ? (point.moisture?.toFixed(1) ?? 'N/A')
            : (point.temperature?.toFixed(1) ?? 'N/A')
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
  URL.revokeObjectURL(url);
};

// --- Methods for Slideshow Navigation ---
const nextResolution = () => {
  resolutionIndex.value = (resolutionIndex.value + 1) % resolutionOptions.length;
};

const prevResolution = () => {
  resolutionIndex.value = (resolutionIndex.value - 1 + resolutionOptions.length) % resolutionOptions.length;
};

// --- Methods for Sensor Type Navigation ---
const nextSensorType = () => {
  sensorTypeIndex.value = (sensorTypeIndex.value + 1) % sensorTypeOptions.length;
};

const prevSensorType = () => {
  sensorTypeIndex.value = (sensorTypeIndex.value - 1 + sensorTypeOptions.length) % sensorTypeOptions.length;
};
// --- End Methods for Sensor Type Navigation ---

// Toggle sensor visibility
const toggleSensor = (sensorName: string) => {
  // Toggle the current state
  sensorVisibility.value[sensorName] = !sensorVisibility.value[sensorName];
  
  // Ensure at least one sensor is always selected
  const atLeastOneVisible = Object.values(sensorVisibility.value).some(visible => visible);
  if (!atLeastOneVisible) {
    // If no sensors are visible, re-enable the one that was just disabled
    sensorVisibility.value[sensorName] = true;
  }
  
  console.log(`[Sensors.vue] Toggled sensor ${sensorName}, now ${sensorVisibility.value[sensorName] ? 'visible' : 'hidden'}`);
};

// --- REMOVED Methods for Slideshow Navigation ---
// const nextSensor = () => { ... }
// const prevSensor = () => { ... }
// --- END REMOVED Methods for Slideshow Navigation ---

// --- REMOVED Computed Properties --- 
// const currentSensorName = computed(() => { ... });
// --- END REMOVED Computed Properties ---

// --- REMOVED Computed property to generate visibility object based on active index
// const sensorVisibility = computed(() => { ... });
// --- END REMOVED Computed property ---

// --- REMOVED Watcher for sensor changes
// watch(activeSensorIndex, (newIndex) => { ... });
// --- END REMOVED Watcher ---

// --- Computed Properties --- 
const currentSensorName = computed(() => {
  // Make sure SENSOR_CONFIGS is not empty
  return SENSOR_CONFIGS.value.length > 0 ? SENSOR_CONFIGS.value[0]?.name : 'No Sensors';
});

// Computed property for current sensor type label
const currentSensorTypeLabel = computed(() => {
  return sensorTypeOptions[sensorTypeIndex.value]?.label || 'N/A';
});

// Add computed property for current sensor type
const currentSensorType = computed(() => {
  return sensorTypeOptions[sensorTypeIndex.value].value as 'moisture' | 'temperature';
});

// Add the resetToDefault function in the script section
const resetToDefault = () => {
  const now = new Date();
  const todayStr = toLocalDateString(now);
  
  // Reset to default state
  dynamicTimeWindow.value = 'none';
  
  // Reset query params to default values
  const defaultQueryParams: QueryParams = {
    startDate: todayStr,
    endDate: todayStr,
    minValue: currentSensorType.value === 'moisture' ? 0 : -10,
    maxValue: currentSensorType.value === 'moisture' ? 100 : 50,
    resolution: 'hourly' as const // Default to hourly resolution
  };

  // Update the appropriate query params based on sensor type
  if (currentSensorType.value === 'moisture') {
    moistureQueryParams.value = defaultQueryParams;
  } else {
    temperatureQueryParams.value = defaultQueryParams;
  }

  // Reset resolution index to hourly
  const hourlyIndex = resolutionOptions.findIndex(opt => opt.value === 'hourly');
  if (hourlyIndex !== -1) {
    resolutionIndex.value = hourlyIndex;
  }

  // Trigger graph update
  nextTick(() => {
    const currentGraph = currentSensorType.value === 'moisture' 
      ? moistureGraphComponent.value 
      : temperatureGraphComponent.value;
    
    if (currentGraph) {
      currentGraph.fetchAllSensorData();
    }
  });
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
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.realtime-container {
  padding: 20px;
  background: #f8f9fa;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  max-width: 928px; /* Align with graph */
  margin-left: auto;   /* Center container */
  margin-right: auto;  /* Center container */
  width: 100%;
}

.timer {
  display: none;
}

.sensor-carousel {
  display: flex;
  align-items: center;
  justify-content: center; /* This can stay, card is 100% anyway */
  width: 100%; /* Make carousel fill realtime-container */
  gap: 20px;
  margin-bottom: 20px;
}

.sensor-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid var(--sensor-color);
  /* min-width: 200px; */ /* Replaced by width */
  width: 100%; /* Make card fill carousel */
  box-sizing: border-box; /* Include padding in width calculation */
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

/* Add styles for the sensor type slideshow control */
.sensor-type-slideshow {
  display: flex;
  align-items: center; 
  justify-content: space-between; 
  border: 1px solid #dee2e6; 
  border-radius: 4px;
  padding: 5px 10px; 
  background-color: white; 
}

/* Styles for the sensor selection checkboxes */
.sensor-selection {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sensor-checkbox label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  border-left: 4px solid var(--sensor-color);
  background: white;
  transition: background-color 0.2s;
}

.sensor-checkbox label:hover {
  background: #f0f0f0;
}

.sensor-checkbox input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.sensor-checkbox .sensor-name {
  font-weight: 500;
  color: var(--sensor-color);
}

.sensor-card-header {
  display: flex;
  justify-content: space-between; /* Pushes h3 and svg apart */
  align-items: center; /* Vertically aligns items */
  margin-bottom: 10px; /* Add space below header */
}

.sensor-card-header h3 {
  margin: 0; /* Remove default margin from h3 */
}

.trend-indicator path {
  transition: d 0.4s ease-in-out; /* Animate the path data change */
}

.reset-btn {
  grid-column: 1 / -1; /* Make reset button span full width */
  background: #f8f9fa !important;
  border-color: #dc3545 !important;
  color: #dc3545;
  font-weight: 500;
}

.reset-btn:hover {
  background: #dc3545 !important;
  color: white !important;
}

/* Sensor loading and error states */
.loading-state, .error-state, .no-sensors-state {
  padding: 12px;
  text-align: center;
  border-radius: 4px;
  margin: 8px 0;
}

.loading-state {
  background: #f8f9fa;
  color: #666;
  font-style: italic;
}

.error-state {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.no-sensors-state {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.retry-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 4px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 0.8rem;
  margin-left: 8px;
  transition: background-color 0.2s;
}

.retry-btn:hover {
  background: #c82333;
}
</style>