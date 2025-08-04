<template>
  <div class="home-root">
    <!-- Map View -->
    <div class="map-container">
      <!-- Header Icon Tiles (overlay on map) -->
      <div class="header-tiles-bar">
        <RouterLink to="/sensors" class="header-tile">
          <i class="fas fa-thermometer-half"></i>
          <span>Sensors</span>
        </RouterLink>
        <RouterLink to="/battery" class="header-tile">
          <i class="fas fa-battery-full"></i>
          <span>Battery Status</span>
        </RouterLink>
        <RouterLink to="/solar" class="header-tile">
          <i class="fas fa-solar-panel"></i>
          <span>Solar Panels</span>
        </RouterLink>
        <RouterLink to="/drone" class="header-tile drone-tile">
          <div class="drone-icon">
            <div class="drone-body">
              <div class="drone-arm arm-1"></div>
              <div class="drone-arm arm-2"></div>
              <div class="drone-arm arm-3"></div>
              <div class="drone-arm arm-4"></div>
            </div>
          </div>
          <span>Drone</span>
        </RouterLink>
      </div>
      <img src="../assets/farmMap.png" alt="Farm Map" class="farm-map" />
      <!-- Sensor B3 point with tooltip -->
      <div
        class="map-point sensor sensor-b3"
        :style="pointStyle(59, 32)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, 'B3')"
        @mouseleave="handleMouseLeave"
        title="Sensor B3"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Sensor 92 point with tooltip -->
      <div
        class="map-point sensor sensor-92"
        :style="pointStyle(59, 40)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, '92')"
        @mouseleave="handleMouseLeave"
        title="Sensor 92"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Sensor 87 point with tooltip -->
      <div
        class="map-point sensor sensor-87"
        :style="pointStyle(45, 90)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, '87')"
        @mouseleave="handleMouseLeave"
        title="Sensor 87"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Sensor 9D point with tooltip -->
      <div
        class="map-point sensor sensor-9d"
        :style="pointStyle(30, 85)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, '9D')"
        @mouseleave="handleMouseLeave"
        title="Sensor 9D"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Sensor B9 point with tooltip -->
      <div
        class="map-point sensor sensor-b9"
        :style="pointStyle(35, 90)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, 'B9')"
        @mouseleave="handleMouseLeave"
        title="Sensor B9"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Sensor C6 point with tooltip -->
      <div
        class="map-point sensor sensor-c6"
        :style="pointStyle(18, 90)"
        @click="goTo('sensors')"
        @mouseenter="handleMouseEnter($event, 'C6')"
        @mouseleave="handleMouseLeave"
        title="Sensor C6"
      >
        <i class="fas fa-thermometer-half"></i>
      </div>

      <!-- Tooltip -->
      <div
        v-if="showTooltip"
        class="sensor-tooltip"
        :style="{
          left: `${tooltipPosition.x}px`,
          top: `${tooltipPosition.y - 120}px`,
          '--sensor-color': SENSORS[activeSensor]?.color || '#4CAF50'
        }"
      >
        <div class="tooltip-header">{{ sensorData.name }}</div>
        <div class="tooltip-content">
          <div class="tooltip-row">
            <i class="fas fa-microchip"></i>
            <span>ID: {{ sensorData.id }}</span>
          </div>
          <div class="tooltip-row">
            <i class="fas fa-tint"></i>
            <span>Moisture: {{ formatValue(sensorData.moisture, 'moisture') }}</span>
          </div>
          <div class="tooltip-row">
            <i class="fas fa-thermometer-half"></i>
            <span>Temperature: {{ formatValue(sensorData.temperature, 'temperature') }}</span>
          </div>
          <div class="tooltip-time">
            Last updated: {{ sensorData.lastUpdated ? formatTime(sensorData.lastUpdated) : 'Loading...' }}
          </div>
        </div>
      </div>
      <div
        class="map-point solar"
        :style="pointStyle(18, 50)"
        @click="goTo('solar')"
        title="Solar Panels"
      >
        <i class="fas fa-solar-panel"></i>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Importing Google Fonts */
@import url('https://fonts.googleapis.com/css2?family=Roboto:wght@500&display=swap');
/* Import Font Awesome */
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');

.home-root {
  min-height: 100vh;
  width: 100vw;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  padding: 0;
  margin: 0;
  overflow: hidden;
}

.header-tiles-bar {
  position: absolute;
  top: 90px; /* Adjusted to create more space below navbar */
  left: 50%;
  transform: translateX(-50%);
  width: auto;
  z-index: 10;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 32px;
  background: rgba(255,255,255,0.18);
  backdrop-filter: blur(12px);
  box-shadow: 0 4px 24px rgba(0,0,0,0.10);
  padding: 14px 36px 10px 36px;
  min-height: 54px;
  border: none;
  border-radius: 18px;
}
.header-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  color: #222;
  font-family: 'Roboto', sans-serif;
  font-size: 16px;
  font-weight: 500;
  transition: transform 0.18s, box-shadow 0.18s, background 0.18s, color 0.18s;
  border-radius: 10px;
  padding: 6px 18px 2px 18px;
}
.header-tile i {
  font-size: 2rem;
  margin-bottom: 6px;
}
.header-tile:hover {
  background: #e3f2fd;
  transform: translateY(-2px) scale(1.08);
  box-shadow: 0 4px 16px rgba(25,118,210,0.13);
  color: #1976d2;
}

.map-container {
  position: relative;
  width: 100%;
  max-width: 100vw;
  height: 100vh;
  min-height: 0;
  margin: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 0;
  box-shadow: none;
  overflow: hidden;
  background: #fff;
}
.farm-map {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 0;
  display: block;
  filter: brightness(0.98) saturate(1.05);
  transition: filter 0.3s;
}
.map-container:hover .farm-map {
  filter: brightness(1) saturate(1.1);
}
.map-point {
  position: absolute;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #fff;
  border: 3px solid #1976d2;
  box-shadow: 0 2px 12px rgba(0,0,0,0.18);
  cursor: pointer;
  transition: transform 0.22s, box-shadow 0.22s, border-color 0.22s;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
  font-size: 1.3rem;
  animation: pointPulse 2.2s infinite cubic-bezier(0.4,0,0.2,1);
}
@keyframes pointPulse {
  0% { box-shadow: 0 2px 12px rgba(25,118,210,0.18); }
  50% { box-shadow: 0 6px 24px rgba(25,118,210,0.22); }
  100% { box-shadow: 0 2px 12px rgba(25,118,210,0.18); }
}
.map-point i {
  font-size: 1.3rem;
}
.map-point:hover {
  transform: scale(1.18);
  box-shadow: 0 8px 32px rgba(25,118,210,0.22);
  border-color: #1565c0;
}
.map-point.sensor {
  border-color: #43a047;
  background: #e8f5e9;
}
.map-point.solar {
  border-color: #fbc02d;
  background: #fffde7;
}

/* Responsive adjustments */
@media (max-width: 900px) {
  .header-tiles-bar {
    gap: 18px;
    padding: 10px 12px 6px 12px;
    border-radius: 12px;
    min-height: 44px;
    top: 80px; /* Adjusted for smaller screens */
  }
  .map-container {
    max-width: 100vw;
    height: calc(100vh - 44px);
    min-height: 0;
    border-radius: 0;
    margin-top: 0;
  }
  .farm-map {
    border-radius: 0;
  }
}
@media (max-width: 600px) {
  .header-tiles-bar {
    flex-direction: row;
    gap: 8px;
    padding: 6px 4px 2px 4px;
    border-radius: 8px;
    top: 70px; /* Adjusted for mobile screens */
    min-height: 32px;
  }
  .map-container {
    max-width: 100vw;
    height: calc(100vh - 44px);
    min-height: 0;
    border-radius: 0;
    margin-top: 0;
  }
  .farm-map {
    border-radius: 0;
  }
  .header-tile {
    font-size: 13px;
    padding: 2px 6px 0 6px;
  }
  .header-tile i {
    font-size: 1.2rem;
    margin-bottom: 2px;
  }
}

/* Add new tooltip styles */
.sensor-tooltip {
  position: fixed;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 16px;
  padding: 16px;
  min-width: 200px;
  box-shadow: 
    0 4px 24px -1px rgba(0, 0, 0, 0.08),
    0 0 1px 0 rgba(0, 0, 0, 0.06),
    inset 0 0 0 1px rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  z-index: 1000;
  pointer-events: none;
  transform: translateX(-50%);
}

.tooltip-header {
  font-weight: 600;
  color: var(--sensor-color);
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.tooltip-content {
  font-size: 0.9rem;
  color: rgba(0, 0, 0, 0.8);
}

.tooltip-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  padding: 4px 0;
}

.tooltip-row i {
  width: 16px;
  color: rgba(102, 102, 102, 0.8);
}

.tooltip-time {
  font-size: 0.8rem;
  color: rgba(102, 102, 102, 0.8);
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}

.map-point.sensor-b3 {
  border-color: #4CAF50;
  background: #E8F5E9;
}

.map-point.sensor-92 {
  border-color: #2196F3;
  background: #E3F2FD;
}

.map-point.sensor-87 {
  border-color: #FF5722;
  background: #FBE9E7;
}

.map-point.sensor-9d {
  border-color: #9C27B0;
  background: #F3E5F5;
}

.map-point.sensor-b9 {
  border-color: #FFC107;
  background: #FFF8E1;
}

.map-point.sensor-c6 {
  border-color: #00BCD4;
  background: #E0F7FA;
}

/* Static Drone Icon Styles */
.drone-tile {
  position: relative;
}

.drone-icon {
  position: relative;
  width: 40px;
  height: 40px;
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drone-body {
  position: relative;
  width: 16px;
  height: 16px;
  background: #1976d2;
  border-radius: 4px;
}

.drone-arm {
  position: absolute;
  width: 6px;
  height: 6px;
  background: #64b5f6;
  border-radius: 50%;
}

.arm-1 {
  top: -10px;
  left: -10px;
}

.arm-2 {
  top: -10px;
  right: -10px;
}

.arm-3 {
  bottom: -10px;
  left: -10px;
}

.arm-4 {
  bottom: -10px;
  right: -10px;
}
</style>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useNexusStore } from '@/stores/nexus';

const router = useRouter();
const nexusStore = useNexusStore();

const SENSORS = {
  B3: {
    id: '2CF7F1C0649007B3',
    name: 'Sensor B3',
    color: '#4CAF50'
  },
  '92': {
    id: '2CF7F1C064900792',
    name: 'Sensor 92',
    color: '#2196F3'
  },
  '87': {
    id: '2CF7F1C064900787',
    name: 'Sensor 87',
    color: '#FF5722'
  },
  '9D': {
    id: '2CF7F1C06490079D',
    name: 'Sensor 9D',
    color: '#9C27B0'
  },
  'B9': {
    id: '2CF7F1C064900792', // Changed to match the physical sensor's ID
    name: 'Sensor B9',
    color: '#FFC107'
  },
  'C6': {
    id: '2CF7F1C0649007C6',
    name: 'Sensor C6',
    color: '#00BCD4'
  }
};

// Sensor data state
const sensorData = ref({
  id: '',
  name: '',
  moisture: null,
  temperature: null,
  lastUpdated: null
});

const activeSensor = ref(null);

// Tooltip state
const showTooltip = ref(false);
const tooltipPosition = ref({ x: 0, y: 0 });

// Add conversion function
const celsiusToFahrenheit = (celsius) => {
  return (celsius * 9/5) + 32;
};

// Helper to position points by percentage (x, y)
function pointStyle(xPercent, yPercent) {
  return {
    left: `calc(${xPercent}% - 19px)`, // center the 38px point
    top: `calc(${yPercent}% - 19px)`
  }
}

function goTo(page) {
  router.push(`/${page}`)
}

// Format functions
function formatValue(value, type) {
  if (value === null) return 'Loading...';
  return type === 'temperature' 
    ? `${celsiusToFahrenheit(value).toFixed(1)}°F`
    : `${value.toFixed(1)}%`;
}

function formatTime(time) {
  if (!time) return '';
  return new Date(time).toLocaleTimeString();
}

// Update tooltip handlers
function handleMouseEnter(event, sensorKey) {
  const rect = event.target.getBoundingClientRect();
  tooltipPosition.value = {
    x: rect.left + window.scrollX,
    y: rect.top + window.scrollY
  };
  activeSensor.value = sensorKey;
  showTooltip.value = true;
  fetchSensorData(sensorKey);
}

// Update fetch sensor data
async function fetchSensorData(sensorKey) {
  try {
    const sensor = SENSORS[sensorKey];
    if (!sensor) return;

    const [moistureData, temperatureData] = await Promise.all([
      nexusStore.user.getSensorMoistureData(sensor.id),
      nexusStore.user.getSensorTemperatureData(sensor.id)
    ]);

    if (moistureData?.length && temperatureData?.length) {
      const latestMoisture = moistureData[moistureData.length - 1];
      const latestTemperature = temperatureData[temperatureData.length - 1];

      sensorData.value = {
        id: sensor.id,
        name: sensor.name,
        moisture: Number(latestMoisture.soil_moisture),
        temperature: Number(latestTemperature.soil_temperature),
        lastUpdated: new Date(Math.max(
          new Date(latestMoisture.date),
          new Date(latestTemperature.date)
        ))
      };
    }
  } catch (error) {
    console.error('Error fetching sensor data:', error);
  }
}

// Update data periodically
let updateInterval;

onMounted(() => {
  // No need for initial fetch or interval anymore
});

onUnmounted(() => {
  // No need to clear interval anymore
});

// Tooltip handlers
function handleMouseLeave() {
  showTooltip.value = false;
}
</script>
