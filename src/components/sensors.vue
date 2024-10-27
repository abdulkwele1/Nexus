<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar">
      <button class="nav-button" @click="goTo('/home')">Home</button>
    </nav>
  </div>
  <div class="center-container">
    <div class="button-container">
      <div v-for="(button, index) in buttons" :key="index" class="dropdown-button">
        <button @click="toggleDropdown(index)">
          {{ button.name }}
        </button>
        <transition name="fade-slide">
          <div v-if="button.showSensors" class="sensor-dropdown">
            <div class="sensor" v-if="index < 4" v-for="sensor in button.sensors" :key="sensor.id">
              <span class="sensor-icon">🌱</span>
              <div class="sensor-info">
                <p class="sensor-details">
                  {{ sensor.label }} - Battery: {{ sensor.battery }}%
                </p>
                <p class="sensor-coordinates">
                  Coordinates: ({{ sensor.coordinates.x }}, {{ sensor.coordinates.y }})
                </p>
              </div>
            </div>
            <div v-if="index === 4" class="sensor-info">
              <p>Coordinates for Solar Panels:</p>
              <div v-for="(coordinate, coordIndex) in button.coordinates" :key="coordIndex" class="sensor-coordinates">
                Solar Panel {{ coordIndex + 1 }}: Coordinates: ({{ coordinate.x }}, {{ coordinate.y }})
              </div>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script>
import { useRouter } from 'vue-router';

export default {
  setup() {
    const router = useRouter(); // Initialize the router

    const goTo = (path) => {
      router.push(path);
    };

    return { goTo };
  },
  data() {
    return {
      buttons: [
        { name: "North field", showSensors: false, sensors: [this.generateSensor("Sensor 1"), this.generateSensor("Sensor 2")] },
        { name: "South field", showSensors: false, sensors: [this.generateSensor("Sensor 1"), this.generateSensor("Sensor 2")] },
        { name: "West field", showSensors: false, sensors: [this.generateSensor("Sensor 1"), this.generateSensor("Sensor 2")] },
        { name: "East field", showSensors: false, sensors: [this.generateSensor("Sensor 1"), this.generateSensor("Sensor 2")] },
        { name: "Solar Panels", showSensors: false, coordinates: this.generateMultipleCoordinates(3) }, // 3 random coordinates
      ],
    };
  },
  methods: {
    toggleDropdown(index) {
      this.buttons[index].showSensors = !this.buttons[index].showSensors;
    },
    generateSensor(label) {
      return {
        id: Math.random(),
        label: label,
        battery: Math.floor(Math.random() * 101),
        coordinates: {
          x: (Math.random() * 100).toFixed(2),
          y: (Math.random() * 100).toFixed(2),
        },
      };
    },
    generateMultipleCoordinates(count) {
      return Array.from({ length: count }, () => this.generateCoordinates());
    },
    generateCoordinates() {
      return {
        x: (Math.random() * 100).toFixed(2),
        y: (Math.random() * 100).toFixed(2),
      };
    },
  },
};
</script>

<style>
.center-container {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  height: 100vh;
  padding-top: 30vh;
}

.button-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 800px; /* Increased width for button container */
}

.dropdown-button button {
  width: 100%; /* Full width button */
  padding: 0.5rem;
  background-color: #007bff;
  color: white;
  border: none;
  cursor: pointer;
  border-radius: 5px;
  transition: background-color 0.3s, transform 0.3s, box-shadow 0.3s; /* Include box-shadow for hover */
}

.dropdown-button button:hover {
  background-color: #0056b3;
  transform: scale(1.05);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2); /* Add shadow on hover */
}

.dropdown-button button:focus {
  outline: none; /* Remove default outline */
  box-shadow: 0 0 0 4px rgba(0, 123, 255, 0.5); /* Custom focus outline */
}

.sensor-dropdown {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem;
  background-color: #f9f9f9; /* Slightly lighter background */
  border: 1px solid #ddd;
  border-radius: 8px; /* Slightly larger radius for dropdown */
  margin-top: 0.5rem;
  width: 100%;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1); /* Add shadow to dropdown */
}

.sensor {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background: #e8f5e9;
  border-radius: 5px;
  transition: transform 0.2s, box-shadow 0.2s; /* Include transition */
}

.sensor:hover {
  transform: translateY(-2px); /* Lift effect on hover */
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2); /* Shadow effect on hover */
}

.sensor-icon {
  font-size: 24px; /* Adjusts size of the emoji */
}

.sensor-info {
  display: flex;
  flex-direction: column;
}

.sensor-details {
  font-size: 1.1rem; /* Slightly larger text */
  color: #333; /* Darker color for better readability */
}

.sensor-coordinates {
  font-size: 0.9rem; /* Smaller font size for coordinates */
  color: #666; /* Lighter color for less emphasis */
}

.fade-slide-enter-active, .fade-slide-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.fade-slide-enter, .fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* Responsive Design */
@media (max-width: 768px) {
  .button-container {
    width: 100%; /* Full width on smaller screens */
    padding: 1rem; /* Add some padding */
  }
}

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
</style>
