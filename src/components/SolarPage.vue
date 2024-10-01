<template>
  <div class="container">
    <button class="home-button" @click="goHome">Home</button>
    <div class="chart-container">
      <button class="current-date-button" @click="openCalendar">
        Select Date Range &#9662;
      </button>
      <div class="solar-panel-container">
        <button class="solar-panels-button" @click="toggleDropdown">
          Solar Panels
        </button>
        <div v-if="showDropdown" class="dropdown">
          <ul>
            <li @click="selectSolarPanel('Panel 1')">Panel 1</li>
            <li @click="selectSolarPanel('Panel 2')">Panel 2</li>
            <li @click="selectSolarPanel('Panel 3')">Panel 3</li>
          </ul>
        </div>
      </div>
      <label class="line-chart-label">
        <input type="checkbox" v-model="isLineChart" />
        Switch to Line Chart
      </label>
      <button class="export-button" @click="exportData">
        📄 Export
      </button>
      <Graph :solarData="solarData" :isLineChart="isLineChart" />
    </div>

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
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import Graph from './Graph.vue'; // Import the Graph component

const router = useRouter();
const showCalendar = ref(false);
const showDropdown = ref(false); // Controls the visibility of the dropdown
const startDate = ref(null);
const endDate = ref(null);
const solarData = ref([]); // Stores solar data
const errorMessage = ref(""); // Stores error message
const isLineChart = ref(false); // State for line chart toggle
const selectedPanel = ref("Panel 1"); // Store the selected solar panel, defaults to 'Panel 1'

const goHome = () => {
  router.push('/home');
};

const openCalendar = () => {
  showCalendar.value = true;
};

const closeCalendar = () => {
  showCalendar.value = false;
};

const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value; // Toggle dropdown visibility
};

const selectSolarPanel = (panel) => {
  selectedPanel.value = panel;
  console.log("Selected Solar Panel:", selectedPanel.value);
  showDropdown.value = false; // Close dropdown after selection
};

// Function to generate random solar production data for the given date range
const generateSolarData = (start, end) => {
  const data = [];
  const currentDate = new Date(start);
  const endDate = new Date(end);

  while (currentDate <= endDate) {
    const month = currentDate.getMonth();
    let production;
    if (month === 5 || month === 6 || month === 7) {
      production = Math.floor(Math.random() * 100) + 150; // Summer months
    } else {
      production = Math.floor(Math.random() * 50) + 50; // Other months
    }

    data.push({
      date: new Date(currentDate),
      production,
    });

    currentDate.setDate(currentDate.getDate() + 1); // Increment the date by 1 day
  }

  return data;
};

// Watch for changes in start and end dates
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

// Function to export data to CSV
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
/* Container for the overall layout */
.container {
  position: relative;
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
  top: -25px; /* Adjust this value to align with other buttons */
  left: 500px; /* Adjust this value to space from Select Date Range button */
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
  top: -30px; /* Adjust as necessary to align with other elements */
  left: 650px; /* Adjust this value to space from the Solar Panels button */
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
  top: 30px; /* Adjust this value to position below the button */
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