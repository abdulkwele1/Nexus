<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar">
      <button class="nav-button" @click="goTo('/home')">Home</button>
      <button class="nav-button" @click="switchGraph('yield')">Solar Yield</button>
      <button class="nav-button" @click="switchGraph('consumption')">Solar Consumption</button>
    </nav>

    <!-- Conditional rendering of graphs and controls for Solar Yield -->
    <div class="chart-container" v-if="currentGraph === 'yield'">
      <Graph :solarData="solarData" :isLineChart="isLineChart" />
      
      <!-- Line chart switch button -->
      <label class="line-chart-label">
        <input type="checkbox" v-model="isLineChart" />
        Switch to Line Chart
      </label>

      <!-- Export button -->
      <button class="export-button" @click="exportData">📄 Export</button>

      <!-- Show panels button -->
      <div class="solar-panel-container">
        <button class="solar-panels-button" @click="toggleDropdown">Solar Panels</button>
        <div v-if="showDropdown" class="dropdown">
          <ul>
            <li @click="selectSolarPanel('Panel 1')">Panel 1</li>
            <li @click="selectSolarPanel('Panel 2')">Panel 2</li>
            <li @click="selectSolarPanel('Panel 3')">Panel 3</li>
          </ul>
        </div>
      </div>

      <!-- Calendar -->
      <button class="current-date-button" @click="openCalendar">
        Select Date Range &#9662;
      </button>

      <!-- Calendar modal -->
      <div v-if="showCalendar" class="modal-overlay" @click="closeCalendar">
        <div class="modal" @click.stop>
          <h2>Select Date Range</h2>
          <div class="calendar-container">
            <div class="calendar">
              <h3>Start Point</h3>
              <input type="date" v-model="startDate" />
            </div>
            <div class="calendar">
              <h3>End Point</h3>
              <input type="date" v-model="endDate" />
            </div>
          </div>
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <button @click="closeCalendar">Close</button>
        </div>
      </div>
    </div>

    <!-- Solar Consumption graph -->
    <div class="chart-container" v-if="currentGraph === 'consumption'">
      <BarGraph />
    </div>
  </div>
</template>



<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import Graph from './Graph.vue'; // Import the Solar Yield Graph component
import BarGraph from './yieldGraph.vue';  // Bar graph for Solar Consumption

const router = useRouter();
const currentGraph = ref('yield');  // Default is Solar Yield graph
const showCalendar = ref(false);
const showDropdown = ref(false);
const startDate = ref(null);
const endDate = ref(null);
const solarData = ref([]);
const errorMessage = ref("");
const isLineChart = ref(false);
const selectedPanel = ref("Panel 1");

const switchGraph = (graphType) => {
  if (currentGraph.value !== graphType) {
    currentGraph.value = graphType;  // Switch graph if different from current
  }
};

const goTo = (path) => {
  router.push(path);
};

const openCalendar = () => {
  showCalendar.value = true;
};

const closeCalendar = () => {
  showCalendar.value = false;
};

const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value;
};

const selectSolarPanel = (panel) => {
  selectedPanel.value = panel;
  showDropdown.value = false;
};

const generateSolarData = (start, end) => {
  const data = [];
  const currentDate = new Date(start);
  const endDate = new Date(end);

  while (currentDate <= endDate) {
    const month = currentDate.getMonth();
    let production;
    if (month === 5 || month === 6 || month === 7) {
      production = Math.floor(Math.random() * 100) + 150;
    } else {
      production = Math.floor(Math.random() * 50) + 50;
    }

    data.push({
      date: new Date(currentDate),
      production,
    });

    currentDate.setDate(currentDate.getDate() + 1);
  }

  return data;
};

watch([startDate, endDate], () => {
  if (startDate.value && endDate.value) {
    const start = new Date(startDate.value);
    const end = new Date(endDate.value);
    
    if (end < start) {
      errorMessage.value = "End point cannot be before starting point.";
    } else {
      solarData.value = generateSolarData(start, end);
    }
  }
});

const exportData = () => {
  const header = "sensor_reading_date,daily_kw_generated\n";
  const csvContent = "data:text/csv;charset=utf-8," 
    + header 
    + solarData.value.map(d => `${d.date.toISOString().split('T')[0]},${d.production}`).join("\n");

  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  link.setAttribute("download", "solar_data.csv");
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

onMounted(() => {
  solarData.value = generateSolarData(new Date("2023-01-01"), new Date("2023-01-31"));
});
</script>

<style scoped>
/* Navbar styling */
.navbar {
  position: fixed;
  left: 50px;
  top: 0;
  width: 100%;
  display: flex;
  justify-content: space-around;
  background-color: #fff;
  padding: 10px;
  z-index: 1000; /* Ensures navbar stays on top of other content */
  border-bottom: 2px solid #ccc; /* Add a bottom border */
}

.nav-button {
  background-color: #fff;
  border: none;
  padding: 10px 20px;
  color: #333;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
}

.nav-button:hover {
  background-color: #ddd;
}

/* Container for the overall layout */
.container {
  position: relative;
  top: 60px; /* Adds space below the fixed navbar */
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}

.chart-container {
  position: relative;
  width: 1000px;
  height: 500px;
}

/* Home button style */
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

/* Date selection button */
.current-date-button {
  position: absolute;
  top: -25px;
  left: 125px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #f8f9fa;
  color: #343a40;
  border: 1px solid #ced4da;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.current-date-button:hover {
  background-color: #e2e6ea;
}

/* Export button */
.export-button {
  position: absolute;
  bottom: -40px;
  right: 0;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #28a745;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.export-button:hover {
  background-color: #218838;
}

/* Modal for calendar overlay */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
  text-align: center;
}

/* Calendar container inside the modal */
.calendar-container {
  display: flex;
  justify-content: space-between;
  margin: 20px 0;
}

.calendar {
  width: 45%;
}

/* Error message styling */
.error-message {
  color: red;
  font-size: 14px;
}

/* Button for switching between panels */
.solar-panels-button {
  position: absolute;
  top: -25px;
  left: 500px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #f8f9fa;
  color: #343a40;
  border: 1px solid #ced4da;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.solar-panels-button:hover {
  background-color: #e2e6ea;
}

/* Style for the line chart checkbox */
.line-chart-label {
  position: absolute;
  top: -30px;
  left: 650px;
  display: flex;
  align-items: center;
  font-size: 16px;
}

.line-chart-label input {
  margin-right: 5px;
}

/* Dropdown styling */
.dropdown {
  position: absolute;
  top: 30px;
  left: 0;
  background-color: white;
  border: 1px solid #ced4da;
  border-radius: 5px;
  box-shadow: 0px 8px 16px rgba(0, 0, 0, 0.1);
  z-index: 1;
}

.dropdown ul {
  list-style-type: none;
  padding: 0;
  margin: 0;
}

.dropdown li {
  padding: 10px 20px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.dropdown li:hover {
  background-color: #e2e6ea;
}
</style>
