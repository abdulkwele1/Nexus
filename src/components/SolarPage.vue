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
      <button class="line-chart-toggle-button" @click="isLineChart = !isLineChart">
  {{ isLineChart ? "Bar chart" : "Line chart" }}
      </button>
      <!-- Export button -->
      <button class="export-button" @click="exportData">📄 Export</button>

      <!-- Show panels button -->
      <div class="solar-panel-container">
        <button class="solar-panels-button" @click="showPanelModal = true">Solar Panels</button>

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
    <!-- Solar Panel Selection Modal -->
      <div v-if="showPanelModal" class="modal-overlay" @click="showPanelModal = false">
        <div class="modal" @click.stop>
          <h2>Select Solar Panel</h2>
          <div class="solar-panel-options">
            <div class="solar-panel-card" @click="selectSolarPanel('Panel 1'); showPanelModal = false">
              <h3>Panel 1</h3>          
            </div>
            <div class="solar-panel-card" @click="selectSolarPanel('Panel 2'); showPanelModal = false">
              <h3>Panel 2</h3>
            </div>
            <div class="solar-panel-card" @click="selectSolarPanel('Panel 3'); showPanelModal = false">
              <h3>Panel 3</h3>
            </div>
          </div>
          <button @click="showPanelModal = false">Close</button>
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
import { useNexusStore } from '@/stores/nexus'

const store = useNexusStore()

const {
  VITE_NEXUS_API_URL,
} = import.meta.env;

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
const showPanelModal = ref(false);


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
<<<<<<< Updated upstream
    + solarData.value.map(d => `${d.date.toISOString().split('T')[0]},${d.kwh_yield}`).join("\n");
=======
    + solarData.value.map(d => `${d.date.toISOString().split('T')[0]},${d.production}`).join("\n");
>>>>>>> Stashed changes

  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  link.setAttribute("download", "solar_data.csv");
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

onMounted(async() => {
  const defaultPanelId = 1
  const startDate = '5/11/2024'
  const endDate = '5/12/2024'
<<<<<<< Updated upstream
  const response = await store.user.getPanelYieldData(defaultPanelId, startDate, endDate)
  const responseData= await response.json()
  const yieldData = responseData.yield_data
  // const mockData = generateSolarData(new Date("2023-01-01"), new Date("2023-01-31"));
  // console.log(mockData)
  // debugger
  solarData.value = yieldData.map(item => ({
      date: new Date(item.date),
      kwh_yield: parseFloat(item.kwh_yield) || 0,
}));
=======

  debugger
  const response = await store.user.getPanelYieldData(defaultPanelId, startDate, endDate)
  const responseData= await response.json()
  const yieldData = responseData.yield_data
  const mockData = generateSolarData(new Date("2023-01-01"), new Date("2023-01-31"));
  solarData.value = yieldData
>>>>>>> Stashed changes

});
</script>

<style scoped>
/* Navbar styling */
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
  top: -55px;
  left: 125px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff;
  color: white;
  border: 1px solid #ced4da;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.current-date-button:hover {
  background-color: #0056b3;
}

/* Export button */
.export-button {
  position: fixed;
  top: 70px;
  right: 30px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
  z-index: 1001;
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
  top: -55px;
  left: 500px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff;
  color: white;
  border: 1px solid #ced4da;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.solar-panels-button:hover {
  background-color: #0056b3;
}

/* Style for the line chart checkbox */
.line-chart-toggle-button {
  position: absolute;
  top: -55px;
  left: 650px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff; /* Light gray background */
  color: white;
  border: 1px solid #ced4da; /* Subtle border */
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.3s ease;
}

.line-chart-toggle-button:hover {
  background-color: #0056b3; /* Darker background on hover */
}

.line-chart-toggle-button:active {
  transform: translateY(1px); /* Slight push effect when clicked */
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

/* Solar Panel Selection Modal */
.solar-panel-options {
  display: flex;
  flex-direction: column; /* Align options in a column */
  gap: 15px; /* Space between options */
  margin-top: 20px; /* Space above options */
}

.solar-panel-card {
  background: #f8f9fa; /* Light background color */
  border: 1px solid #ced4da; /* Border for definition */
  border-radius: 8px; /* Rounded corners */
  padding: 15px; /* Padding for content */
  transition: transform 0.1s ease, box-shadow 0.1s ease; /* Smooth transition for hover effects */
  cursor: pointer; /* Pointer cursor to indicate it's clickable */
}

.solar-panel-card:hover {
  transform: translateY(-5px); /* Lift effect on hover */
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); /* Shadow effect on hover */
}

button {
  margin-top: 20px; /* Space above the button */
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  padding: 10px 15px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.3s; /* Smooth transition */
}

button:hover {
  background-color: #0056b3; /* Darker on hover */
  transform: scale(1.05); /* Slightly larger on hover */
}

</style>