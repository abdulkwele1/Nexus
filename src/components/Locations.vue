<template>
  <div class="container">
    <!-- Home Button -->
    <button class="home-button" @click="goHome">Home</button>

    <div class="content">
      <!-- Locations with Coordinates -->
      <div class="locations-container">
        <div class="location-item" v-for="(location, index) in locations" :key="index">
          <span class="coordinates">{{ location.coordinates }}</span>
          <button class="location-button" @click="locationClicked(location.name)">
            {{ location.name }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();
const goHome = () => {
  router.push('/home');
};

const locations = [
  { name: 'South Field', coordinates: '33.5467° N, 101.8793° W' },
  { name: 'North Field', coordinates: '34.0522° N, 118.2437° W' },
  { name: 'West Field', coordinates: '35.4676° N, 97.5164° W' },
  { name: 'East Field', coordinates: '40.7128° N, 74.0060° W' },
  { name: 'Solar Panels', coordinates: '34.0522° N, 118.2437° W' }
];

const locationClicked = (location) => {
  alert(`You clicked on ${location}`);
};
</script>

<style scoped>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

/* Style for the home button */
.home-button {
  position: absolute;
  top: 20px;
  left: 20px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.home-button:hover {
  background-color: #0056b3;
}

/* Container for centering content */
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

/* Flex container for locations */
.content {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  width: 90%;
  max-width: 1200px;
  gap: 20px;
}

/* Locations flex container */
.locations-container {
  display: flex;
  flex-direction: column;
}

/* Style for each location item */
.location-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

/* Style for the location buttons */
.location-button {
  padding: 10px 15px;
  font-size: 14px;
  background-color: #ffffff;
  color: #333;
  border: 2px solid #007bff;
  border-radius: 0;
  cursor: pointer;
  transition: background-color 0.3s, color 0.3s;
  margin-left: 10px;
}

/* Style for coordinates */
.coordinates {
  font-size: 14px;
  color: #555;
}
</style>
