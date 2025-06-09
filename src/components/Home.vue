<template>
  <div>
    <!-- Header Icon Tiles -->
    <div class="header-tiles">
      <RouterLink to="/sensors" class="header-tile">
        <i class="fas fa-thermometer-half"></i>
        <span>Sensors</span>
      </RouterLink>
      <RouterLink to="/solar" class="header-tile">
        <i class="fas fa-solar-panel"></i>
        <span>Solar Panels</span>
      </RouterLink>
      <RouterLink to="/drone" class="header-tile">
        <i class="fas fa-drone"></i>
        <span>Drone</span>
      </RouterLink>
    </div>
    <!-- Map View -->
    <div class="map-container">
      <img src="../assets/farmMap.png" alt="Farm Map" class="farm-map" />
      <!-- Example clickable points -->
      <div
        class="map-point sensor"
        :style="pointStyle(30, 70)"
        @click="goTo('sensors')"
        title="Sensors"
      >
        <i class="fas fa-thermometer-half"></i>
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

.header-tiles {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 40px;
  margin-top: 32px;
  margin-bottom: 24px;
}
.header-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.10);
  padding: 18px 28px 10px 28px;
  text-decoration: none;
  color: #222;
  font-family: 'Roboto', sans-serif;
  font-size: 16px;
  font-weight: 500;
  transition: transform 0.18s, box-shadow 0.18s, background 0.18s;
}
.header-tile i {
  font-size: 2rem;
  margin-bottom: 8px;
}
.header-tile:hover {
  background: #f5f5f5;
  transform: translateY(-4px) scale(1.05);
  box-shadow: 0 4px 16px rgba(25,118,210,0.13);
  color: #1976d2;
}

.map-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  margin: 40px auto 0 auto;
  display: flex;
  justify-content: center;
  align-items: center;
}
.farm-map {
  width: 100%;
  height: auto;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.15);
  display: block;
}
.map-point {
  position: absolute;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #fff;
  border: 3px solid #1976d2;
  box-shadow: 0 2px 8px rgba(0,0,0,0.18);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.map-point i {
  font-size: 1.1rem;
}
.map-point:hover {
  transform: scale(1.15);
  box-shadow: 0 4px 16px rgba(25,118,210,0.25);
}
.map-point.sensor {
  border-color: #43a047;
  background: #e8f5e9;
}
.map-point.solar {
  border-color: #fbc02d;
  background: #fffde7;
}
</style>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()

// Helper to position points by percentage (x, y)
function pointStyle(xPercent, yPercent) {
  return {
    left: `calc(${xPercent}% - 14px)`, // center the 28px point
    top: `calc(${yPercent}% - 14px)`
  }
}
function goTo(page) {
  router.push(`/${page}`)
}
</script>
