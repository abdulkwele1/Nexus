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
  </div>
</template>

<script setup lang="ts">
//imports
import { useRouter } from 'vue-router';
import SoilMoistureGraph from './soilMoistureGraph.vue';

//const
const router = useRouter();
const goTo = (path: string) => {
  router.push(path);
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

.sensors-content {
  margin-top: 80px; /* Add space below the fixed navbar */
  padding: 20px;
  width: 100%;
  max-width: 1200px;
  margin-left: auto;
  margin-right: auto;
}
</style>
