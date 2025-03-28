<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar">
      <button class="nav-button" @click="goTo('/home')">Home</button>
    </nav>
  </div>
  <div class="sensors-content">
    <!-- SoilMoisture graph -->
    <SoilMoistureGraph />
    
    <!-- Real-time data display -->
    <div class="realtime-container">
      <div class="timer">{{ formattedTime }}</div>
      
      <div class="sensor-carousel">
        <button class="carousel-btn prev" @click="prevSensor">&lt;</button>
        
        <div class="sensor-card" :style="{ '--sensor-color': colors[currentSensorIndex] }">
          <h3>{{ currentSensor.name }}</h3>
          <div class="sensor-value">
            {{ currentSensor.moisture.toFixed(1) }}%
          </div>
          <div class="sensor-time">
            Last updated: {{ new Date().toLocaleTimeString() }}
          </div>
        </div>
        
        <button class="carousel-btn next" @click="nextSensor">&gt;</button>
      </div>
      
      <div class="sensor-dots">
        <span 
          v-for="(_, index) in sensors" 
          :key="index"
          :class="{ active: index === currentSensorIndex }"
          @click="currentSensorIndex = index"
        ></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import SoilMoistureGraph from './soilMoistureGraph.vue';

const router = useRouter();
const goTo = (path: string) => {
  router.push(path);
};

// Real-time data setup
const currentTime = ref(new Date());
const currentSensorIndex = ref(0);
const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

interface SensorReading {
  name: string;
  moisture: number;
}

const sensors = ref<SensorReading[]>([
  { name: 'Sensor 1', moisture: 30 },
  { name: 'Sensor 2', moisture: 35 },
  { name: 'Sensor 3', moisture: 25 },
  { name: 'Sensor 4', moisture: 40 }
]);

// Computed properties
const formattedTime = computed(() => {
  return currentTime.value.toLocaleTimeString();
});

const currentSensor = computed(() => {
  return sensors.value[currentSensorIndex.value];
});

// Navigation methods
const nextSensor = () => {
  currentSensorIndex.value = (currentSensorIndex.value + 1) % sensors.value.length;
};

const prevSensor = () => {
  currentSensorIndex.value = (currentSensorIndex.value - 1 + sensors.value.length) % sensors.value.length;
};

// Timer and data update setup
let timeInterval: number;
let dataInterval: number;

onMounted(() => {
  // Update time every second
  timeInterval = setInterval(() => {
    currentTime.value = new Date();
  }, 1000);

  // Update sensor data every 3 seconds
  dataInterval = setInterval(() => {
    sensors.value = sensors.value.map(sensor => ({
      ...sensor,
      moisture: Math.random() * 30 + 20 // Random value between 20-50
    }));
  }, 3000);
});

onUnmounted(() => {
  clearInterval(timeInterval);
  clearInterval(dataInterval);
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

.sensors-content {
  margin-top: 80px; /* Add space below the fixed navbar */
  padding: 20px;
  width: 100%;
  max-width: 1200px;
  margin-left: auto;
  margin-right: auto;
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

.carousel-btn {
  background: transparent;
  border: none;
  font-size: 24px;
  cursor: pointer;
  padding: 10px;
  color: #666;
  transition: color 0.3s;
}

.carousel-btn:hover {
  color: #333;
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

.sensor-dots {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.sensor-dots span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ddd;
  cursor: pointer;
  transition: background-color 0.3s;
}

.sensor-dots span.active {
  background: #666;
}
</style>
