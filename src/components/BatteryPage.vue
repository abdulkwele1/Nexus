<template>
  <div class="battery-page">
    <nav class="top-nav">
      <button class="nav-button" @click="goTo('/home')">Home</button>
    </nav>
    <div class="page-layout">
      <!-- Sidebar -->
      <aside class="sidebar">
        <div class="sensor-toggles">
          <h3>Sensors</h3>
          <div 
            v-for="(sensor, key) in SENSORS" 
            :key="key" 
            class="sensor-toggle"
            :style="{ borderColor: sensor.color }"
          >
            <label :style="{ color: sensor.visible ? 'white' : 'rgba(255,255,255,0.5)' }">
              <input 
                type="checkbox" 
                v-model="sensor.visible"
                :style="{ accentColor: sensor.color }"
              >
              {{ sensor.name }}
            </label>
          </div>
        </div>

        <div class="date-range">
          <label>Start:</label>
          <input type="date" v-model="startDate" />
          <label>End:</label>
          <input type="date" v-model="endDate" />
        </div>

        <div class="quick-filters">
          <button class="filter-btn" @click="handleQuickFilter('7days')">Last 7 Days</button>
          <button class="filter-btn" @click="handleQuickFilter('30days')">Last 30 Days</button>
        </div>

        <!-- Battery Data Manager moved into sidebar -->
        <SensorBatteryUI
          :sensors="SENSORS"
          @dataAdded="handleDataAdded"
        />
      </aside>

      <!-- Main Content -->
      <main class="main-content">
        <!-- Battery Data Manager -->
        <!-- Battery Graph -->
        <BatteryLevels
          :sensors="SENSORS"
          :startDate="startDate"
          :endDate="endDate"
        />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BatteryLevels from './batteryLevelsGraph.vue';
import SensorBatteryUI from './sensorBatteryUI.vue';

const router = useRouter();
// Set initial dates to match our mock data
const today = new Date();
// Set date range to cover our mock data (last 7 days to be safe)
const startDate = ref(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]);
const endDate = ref(new Date().toISOString().split('T')[0]);
console.log('[BatteryPage] Initial date range:', startDate.value, 'to', endDate.value);

interface SensorConfig {
  id: string;
  name: string;
  color: string;
  visible: boolean;
}

import { useNexusStore } from '@/stores/nexus';

const store = useNexusStore();
const SENSORS = ref<Record<string, SensorConfig>>({});

// Available colors for sensors
const SENSOR_COLORS = [
  '#4CAF50', // green
  '#2196F3', // blue
  '#FF5722', // orange
  '#9C27B0', // purple
  '#FFC107', // amber
  '#00BCD4'  // cyan
];

// Fetch all sensors and set up their configs
const initializeSensors = async () => {
  try {
    const response = await store.user.getAllSensors();
    console.log('[BatteryPage] Got sensors response:', response);
    const sensors = response.sensors || [];
    
    const sensorConfigs: Record<string, SensorConfig> = {};
    console.log('[BatteryPage] Processing sensors:', sensors);
    sensors.forEach((sensor: any, index: number) => {
      // Make sure sensor has an ID
      if (!sensor.id) {
        console.warn('[BatteryPage] Sensor missing ID:', sensor);
        return;
      }
      
      // Use full ID as key to ensure uniqueness
      sensorConfigs[sensor.id] = {
        id: sensor.id,
        name: sensor.name || `Sensor ${sensor.id.slice(-2)}`, // Use provided name or generate one
        color: SENSOR_COLORS[index % SENSOR_COLORS.length],
        visible: true
      };
      console.log(`[BatteryPage] Added sensor config for ${sensor.id}:`, sensorConfigs[sensor.id]);
    });
    
    SENSORS.value = sensorConfigs;
  } catch (error) {
    console.error('Error fetching sensors:', error);
  }
};

const goTo = (path: string) => {
  router.push(path);
};

const handleQuickFilter = (filterType: '7days' | '30days') => {
  const end = new Date();
  const start = new Date();
  
  if (filterType === '7days') {
    start.setDate(end.getDate() - 7);
  } else if (filterType === '30days') {
    start.setDate(end.getDate() - 30);
  }
  
  startDate.value = start.toISOString().split('T')[0];
  endDate.value = end.toISOString().split('T')[0];
};

const handleDataAdded = (sensorId: string) => {
  console.log(`[BatteryPage] Battery data added for sensor: ${sensorId}`);
  // You can add additional logic here if needed
  // For example, refresh the graph data
};

onMounted(async () => {
  await initializeSensors();
});
</script>

<style scoped>
.battery-page {
  min-height: 100vh;
  background: #1a1a1a;
}

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
  color: #90EE90;
}

.page-layout {
  display: flex;
  padding-top: 60px;
  min-height: calc(100vh - 60px);
}

.sidebar {
  width: 250px;
  background: #1a1a1a;
  padding: 28px 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  border-right: 2px solid #90EE90;
  box-shadow: 4px 0 15px rgba(0, 0, 0, 0.2);
}

.sensor-toggles {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sensor-toggles h3 {
  color: white;
  margin: 0 0 12px 0;
  font-size: 1.2rem;
}

.sensor-toggle {
  padding: 10px;
  border-radius: 8px;
  border: 1px solid;
  background: rgba(255, 255, 255, 0.05);
  transition: all 0.3s ease;
}

.sensor-toggle:hover {
  background: rgba(255, 255, 255, 0.1);
}

.sensor-toggle label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.sensor-toggle input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.date-range {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  background: #242424;
  border-radius: 8px;
  border: 1px solid rgba(144, 238, 144, 0.1);
}

.date-range label {
  color: #90EE90;
  font-size: 14px;
}

.date-range input {
  padding: 8px 12px;
  border-radius: 4px;
  border: 1px solid rgba(144, 238, 144, 0.2);
  background: #1a1a1a;
  color: white;
  transition: all 0.3s ease;
}

.date-range input:focus {
  outline: none;
  border-color: #90EE90;
  box-shadow: 0 0 10px rgba(144, 238, 144, 0.1);
}

.quick-filters {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.filter-btn {
  padding: 10px 14px;
  background: #242424;
  border: 1px solid rgba(144, 238, 144, 0.1);
  border-radius: 6px;
  cursor: pointer;
  color: white;
  transition: all 0.3s ease;
}

.filter-btn:hover {
  background: #2f2f2f;
  border-color: #90EE90;
  transform: translateX(4px);
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow-x: hidden;
}

@media (max-width: 1400px) {
  .main-content {
    padding: 16px;
  }
}
</style>