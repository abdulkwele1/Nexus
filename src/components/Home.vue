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
        <!-- Admin controls -->
        <div v-if="isAdmin" class="admin-controls">
          <button 
            @click="toggleEditMode" 
            class="admin-btn"
            :class="{ active: editMode }"
            title="Toggle Edit Mode"
          >
            <i class="fas" :class="editMode ? 'fa-check' : 'fa-edit'"></i>
            <span>{{ editMode ? 'Done' : 'Edit' }}</span>
          </button>
          <button 
            v-if="editMode"
            @click="showAddSensorModal = true" 
            class="admin-btn add-sensor-btn"
            title="Add Sensor Icon"
          >
            <i class="fas fa-plus"></i>
            <span>Add Icon</span>
          </button>
        </div>
      </div>
      <img :src="plotImageUrl" alt="Plot 1" class="plot-1" />
      <!-- Dynamic sensor points loaded from API -->
      <div
        v-for="(sensor, index) in sensors"
        :key="`${sensor.id}-${sensor.mapIconId || 0}`"
        class="map-point sensor"
        :class="{
          [`sensor-${index}`]: true,
          'draggable': editMode && isAdmin,
          'dragging': draggingSensor === sensor.mapIconId
        }"
        :style="{
          ...pointStyle(sensor.xPercent || getDefaultX(index), sensor.yPercent || getDefaultY(index)),
          '--sensor-color': sensor.color || getDefaultColor(index),
          'border-color': sensor.color || getDefaultColor(index),
          'background': getColorBackground(sensor.color || getDefaultColor(index)),
          cursor: editMode && isAdmin ? 'move' : 'pointer'
        }"
        @click="editMode && isAdmin ? null : goTo('sensors')"
        @mousedown="editMode && isAdmin ? startDrag($event, sensor, index) : null"
        @mouseenter="!editMode ? handleMouseEnter($event, sensor) : null"
        @mouseleave="!editMode ? handleMouseLeave() : null"
        :title="editMode && isAdmin ? `Drag to move - ${sensor.name}` : sensor.name"
      >
        <i class="fas fa-thermometer-half"></i>
        <button 
          v-if="editMode && isAdmin && sensor.mapIconId" 
          @click.stop="deleteSensorIcon(sensor.mapIconId)"
          class="delete-icon-btn"
          title="Delete this icon"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- Tooltip -->
      <div
        v-if="showTooltip && !editMode"
        class="sensor-tooltip"
        :style="{
          left: `${tooltipPosition.x}px`,
          top: `${tooltipPosition.y - 120}px`,
          '--sensor-color': activeSensor?.color || '#4CAF50'
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

    <!-- Add Sensor Icon Modal (Admin Only) -->
    <div v-if="showAddSensorModal && isAdmin" class="modal-overlay" @click.self="showAddSensorModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Add Sensor Icon</h3>
          <button @click="showAddSensorModal = false" class="close-modal-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <p>Select a sensor to add an icon for:</p>
          <div class="sensor-list">
            <div
              v-for="sensor in availableSensors"
              :key="sensor.id"
              class="sensor-option"
              @click="addSensorIcon(sensor)"
            >
              <i class="fas fa-thermometer-half"></i>
              <span>{{ sensor.name }}</span>
            </div>
          </div>
        </div>
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
.plot-1 {
  width: 100%;
  height: 100%;
  object-fit: cover;
  position: absolute;
  top: 0;
  left: 0;
  z-index: 0;
  filter: brightness(0.98) saturate(1.05);
  transition: filter 0.3s;
}
.map-container:hover .plot-1 {
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

/* Dynamic sensor colors are now applied via inline styles */

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

/* Admin Controls */
.admin-controls {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-left: 20px;
  padding-left: 20px;
  border-left: 1px solid rgba(255, 255, 255, 0.2);
}

.admin-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  padding: 8px 16px;
  color: #222;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
}

.admin-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.admin-btn.active {
  background: #4CAF50;
  color: white;
  border-color: #4CAF50;
}

.add-sensor-btn {
  background: #2196F3;
  color: white;
  border-color: #2196F3;
}

.add-sensor-btn:hover {
  background: #1976d2;
}

/* Draggable sensor styles */
.map-point.draggable {
  cursor: move;
}

.map-point.dragging {
  opacity: 0.7;
  z-index: 1000;
  transform: scale(1.1);
}

.delete-icon-btn {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ff4444;
  color: white;
  border: 2px solid white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  z-index: 10;
  transition: all 0.2s;
}

.delete-icon-btn:hover {
  background: #ff0000;
  transform: scale(1.1);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 0;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h3 {
  margin: 0;
  color: #333;
}

.close-modal-btn {
  background: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #666;
  padding: 5px 10px;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-modal-btn:hover {
  background: #f5f5f5;
  color: #333;
}

.modal-body {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.modal-body p {
  margin: 0 0 15px 0;
  color: #666;
}

.sensor-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sensor-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.sensor-option:hover {
  background: #f5f5f5;
  border-color: #2196F3;
  transform: translateX(5px);
}

.sensor-option i {
  color: #4CAF50;
  font-size: 18px;
}

.sensor-option span {
  font-weight: 500;
  color: #333;
}
</style>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useNexusStore } from '@/stores/nexus';
import plotImageUrl from '@/assets/Plot1.JPG?url';

const router = useRouter();
const nexusStore = useNexusStore();

// Admin and edit mode
const isAdmin = ref(false);
const editMode = ref(false);
const showAddSensorModal = ref(false);

// Drag and drop state
const draggingSensor = ref(null);
const dragOffset = ref({ x: 0, y: 0 });
let nextMapIconId = 1;

// Dynamic sensors loaded from API
const sensors = ref([]);
const sensorData = ref({
  id: '',
  name: '',
  moisture: null,
  temperature: null,
  lastUpdated: null
});

const activeSensor = ref(null);

// Default colors for sensors (cycling through if more than 6)
const defaultColors = [
  '#4CAF50',  // Green
  '#2196F3',  // Blue
  '#FF5722',  // Deep Orange
  '#9C27B0',  // Purple
  '#FFC107',  // Amber
  '#00BCD4',  // Cyan
];

// Default positions (distributed across map)
const defaultPositions = [
  { x: 59, y: 32 },
  { x: 59, y: 40 },
  { x: 45, y: 90 },
  { x: 30, y: 85 },
  { x: 35, y: 90 },
  { x: 18, y: 90 },
  { x: 50, y: 50 },
  { x: 70, y: 60 },
  { x: 25, y: 50 },
  { x: 80, y: 30 },
];

function getDefaultColor(index) {
  return defaultColors[index % defaultColors.length];
}

function getDefaultX(index) {
  return defaultPositions[index % defaultPositions.length]?.x || 50;
}

function getDefaultY(index) {
  return defaultPositions[index % defaultPositions.length]?.y || 50;
}

function getColorBackground(color) {
  // Convert hex to rgba with low opacity for background
  const hex = color.replace('#', '');
  const r = parseInt(hex.substr(0, 2), 16);
  const g = parseInt(hex.substr(2, 2), 16);
  const b = parseInt(hex.substr(4, 2), 16);
  return `rgba(${r}, ${g}, ${b}, 0.15)`;
}

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
function handleMouseEnter(event, sensor) {
  const rect = event.target.getBoundingClientRect();
  tooltipPosition.value = {
    x: rect.left + window.scrollX,
    y: rect.top + window.scrollY
  };
  activeSensor.value = sensor;
  showTooltip.value = true;
  fetchSensorData(sensor);
}

// Update fetch sensor data
async function fetchSensorData(sensor) {
  try {
    if (!sensor || !sensor.id) return;

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

// Load sensor positions from localStorage
function loadSensorPositions() {
  try {
    const saved = localStorage.getItem('sensorMapPositions');
    if (saved) {
      return JSON.parse(saved);
    }
  } catch (error) {
    console.error('Error loading sensor positions:', error);
  }
  return {};
}

// Save sensor positions to localStorage
function saveSensorPositions() {
  try {
    const positions = {};
    sensors.value.forEach(sensor => {
      if (sensor.mapIconId && sensor.xPercent !== null && sensor.yPercent !== null) {
        positions[sensor.mapIconId] = {
          xPercent: sensor.xPercent,
          yPercent: sensor.yPercent,
          sensorId: sensor.id
        };
      }
    });
    localStorage.setItem('sensorMapPositions', JSON.stringify(positions));
  } catch (error) {
    console.error('Error saving sensor positions:', error);
  }
}

// Fetch sensors from API
async function fetchSensors() {
  try {
    const result = await nexusStore.user.getAllSensors();
    const list = Array.isArray(result)
      ? result
      : (Array.isArray(result?.sensors) ? result.sensors : []);

    const savedPositions = loadSensorPositions();
    
    // Create initial sensor icons (one per sensor)
    const sensorIcons = [];
    list.forEach((sensor, index) => {
      // Check if there are saved positions for this sensor
      const savedIcons = Object.entries(savedPositions)
        .filter(([_, pos]) => pos.sensorId === sensor.id)
        .map(([iconId, pos]) => ({
          id: sensor.id,
          name: sensor.name || `Sensor ${sensor.id}`,
          color: getDefaultColor(index),
          xPercent: pos.xPercent,
          yPercent: pos.yPercent,
          latitude: sensor.latitude,
          longitude: sensor.longitude,
          mapIconId: iconId
        }));

      if (savedIcons.length > 0) {
        sensorIcons.push(...savedIcons);
      } else {
        // Create default icon if no saved positions
        sensorIcons.push({
          id: sensor.id,
          name: sensor.name || `Sensor ${sensor.id}`,
          color: getDefaultColor(index),
          xPercent: sensor.latitude ? null : getDefaultX(index),
          yPercent: sensor.longitude ? null : getDefaultY(index),
          latitude: sensor.latitude,
          longitude: sensor.longitude,
          mapIconId: `icon-${nextMapIconId++}`
        });
      }
    });

    // Update nextMapIconId based on existing icons
    const maxIconId = Math.max(...sensorIcons.map(s => {
      const match = s.mapIconId?.match(/icon-(\d+)/);
      return match ? parseInt(match[1]) : 0;
    }), 0);
    nextMapIconId = maxIconId + 1;

    sensors.value = sensorIcons;
    saveSensorPositions();
  } catch (error) {
    console.error('Error fetching sensors:', error);
    sensors.value = [];
  }
}

// Get available sensors for adding icons
const availableSensors = ref([]);

async function loadAvailableSensors() {
  try {
    const result = await nexusStore.user.getAllSensors();
    const list = Array.isArray(result)
      ? result
      : (Array.isArray(result?.sensors) ? result.sensors : []);
    
    availableSensors.value = list.map(sensor => ({
      id: sensor.id,
      name: sensor.name || `Sensor ${sensor.id}`
    }));
  } catch (error) {
    console.error('Error loading available sensors:', error);
    availableSensors.value = [];
  }
}

// Toggle edit mode
function toggleEditMode() {
  editMode.value = !editMode.value;
  if (editMode.value) {
    // Hide tooltip when entering edit mode
    showTooltip.value = false;
    // Load available sensors when entering edit mode
    loadAvailableSensors();
  } else {
    // Save positions when exiting edit mode
    saveSensorPositions();
  }
}

// Add sensor icon
function addSensorIcon(sensor) {
  const newIcon = {
    id: sensor.id,
    name: sensor.name,
    color: getDefaultColor(sensors.value.length),
    xPercent: 50, // Center of map
    yPercent: 50,
    mapIconId: `icon-${nextMapIconId++}`
  };
  sensors.value.push(newIcon);
  saveSensorPositions();
  showAddSensorModal.value = false;
}

// Delete sensor icon
function deleteSensorIcon(mapIconId) {
  if (confirm('Are you sure you want to delete this sensor icon?')) {
    sensors.value = sensors.value.filter(s => s.mapIconId !== mapIconId);
    saveSensorPositions();
  }
}

// Drag and drop functions
function startDrag(event, sensor, index) {
  if (!editMode.value || !isAdmin.value) return;
  
  event.preventDefault();
  draggingSensor.value = sensor.mapIconId;
  
  const rect = event.currentTarget.getBoundingClientRect();
  const mapContainer = event.currentTarget.closest('.map-container');
  const mapRect = mapContainer.getBoundingClientRect();
  
  // Calculate offset from mouse to center of sensor icon
  dragOffset.value = {
    x: event.clientX - rect.left - rect.width / 2,
    y: event.clientY - rect.top - rect.height / 2
  };
  
  document.addEventListener('mousemove', handleDrag);
  document.addEventListener('mouseup', stopDrag);
}

function handleDrag(event) {
  if (!draggingSensor.value) return;
  
  const mapContainer = document.querySelector('.map-container');
  if (!mapContainer) return;
  
  const mapRect = mapContainer.getBoundingClientRect();
  const x = event.clientX - mapRect.left - dragOffset.value.x;
  const y = event.clientY - mapRect.top - dragOffset.value.y;
  
  // Convert to percentages
  const xPercent = Math.max(0, Math.min(100, (x / mapRect.width) * 100));
  const yPercent = Math.max(0, Math.min(100, (y / mapRect.height) * 100));
  
  // Update sensor position
  const sensor = sensors.value.find(s => s.mapIconId === draggingSensor.value);
  if (sensor) {
    sensor.xPercent = xPercent;
    sensor.yPercent = yPercent;
  }
}

function stopDrag() {
  draggingSensor.value = null;
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
  saveSensorPositions();
}

// Update data periodically
let updateInterval;

onMounted(async () => {
  // Check if user is admin
  try {
    await nexusStore.user.getUserSettings();
    isAdmin.value = nexusStore.user.isAdmin || nexusStore.user.userRole === 'admin' || nexusStore.user.userRole === 'root_admin';
  } catch (error) {
    console.error('Error checking admin status:', error);
    isAdmin.value = false;
  }
  
  await fetchSensors();
  
  // Listen for sensor updates
  window.addEventListener('sensorsUpdated', fetchSensors);
});

onUnmounted(() => {
  window.removeEventListener('sensorsUpdated', fetchSensors);
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
});

// Tooltip handlers
function handleMouseLeave() {
  showTooltip.value = false;
}
</script>
