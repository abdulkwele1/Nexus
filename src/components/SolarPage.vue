<template>
  <div class="container">
    <!-- Navigation bar -->
    <nav class="navbar">
      <button class="nav-button" @click="goTo('/home')">Home</button>
      <button class="nav-button" @click="switchGraph('yield')">Solar Yield</button>
      <button class="nav-button" @click="switchGraph('consumption')">Solar Consumption</button>
    </nav>

    <!-- Solar Yield Section -->
    <!-- Solar Yield Section -->
<div class="chart-container" v-if="currentGraph === 'yield'">
  <!-- Graph Component -->
  <Graph :solarData="queriedData.length ? queriedData : panels[activePanel].data" :isLineChart="isLineChart" />

  <!-- Line chart switch button -->
  <button class="line-chart-toggle-button" @click="isLineChart = !isLineChart">
    {{ isLineChart ? "Bar chart" : "Line chart" }}
  </button>

  <!-- Export button -->
  <button class="export-button" @click="exportData">📄 Export</button>

  <!-- Solar Panels Button and Modal -->
  <div class="solar-panel-container">
    <button class="solar-panels-button" @click="togglePanelModal">Solar Panels</button>
    <p>Active Panel: {{ panels[activePanel].name }}</p>
  </div>

  <div v-if="showPanelModal" class="panel-modal">
    <p>Select a solar panel option:</p>
    <button v-for="(panel, index) in panels" :key="index" @click="selectPanel(index)">
      {{ panel.name }}
    </button>
    <button @click="togglePanelModal">Close</button>
  </div>

  <!-- Date Range Selection and Calendar Modal -->
  <button class="current-date-button" @click="openCalendar">
    Select Date Range &#9662;
  </button>

  <div v-if="showCalendar" class="modal-overlay" @click="closeCalendar">
    <div class="modal" @click.stop>
      <h2>Select Date Range</h2>
      <p>
        Allowed range:
        <strong>{{ panels[activePanel].data[0]?.date.toISOString().split('T')[0] }}</strong>
        to
        <strong>{{ panels[activePanel].data[panels[activePanel].data.length - 1]?.date.toISOString().split('T')[0] }}</strong>
      </p>
      <div class="calendar-container">
        <div class="calendar">
          <h3>Start Point</h3>
          <input
            type="date"
            v-model="startDate"
            :min="panels[activePanel].data[0]?.date.toISOString().split('T')[0]"
            :max="panels[activePanel].data[panels[activePanel].data.length - 1]?.date.toISOString().split('T')[0]"
          />
        </div>
        <div class="calendar">
          <h3>End Point</h3>
          <input
            type="date"
            v-model="endDate"
            :min="panels[activePanel].data[0]?.date.toISOString().split('T')[0]"
            :max="panels[activePanel].data[panels[activePanel].data.length - 1]?.date.toISOString().split('T')[0]"
          />
        </div>
      </div>
      <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
      <button @click="closeCalendar">Close</button>
    </div>
  </div>

  <!-- Data Entry Form -->
  <div class="data-entry-form">
    <h3>Add Data Point to {{ panels[activePanel].name }}</h3>
    <input
      type="text"
      v-model="newDate"
      placeholder="YYYY/MM/DD"
      @input="formatDateInput"
    />
    <input
      type="text"
      v-model="newProduction"
      placeholder="kWh Production"
      @keypress="allowOnlyNumbers"
    />
    <button @click="addData">Add Data</button>
    <button @click="clearData">Clear Data for {{ panels[activePanel].name }}</button>
  </div>

  <!-- List of Data Points with Remove Option -->
  <div class="data-list">
    <h3>Data Points for {{ panels[activePanel].name }} (Queried)</h3>
    <button @click="revertToAllData" class="toggle-data-button">
    Show All Data
    </button>

    <ul v-if="queriedData.length">
      <li v-for="(point, index) in queriedData" :key="index">
        {{ point.date.toISOString().split('T')[0] }} - {{ point.production }} kWh
        <button @click="removeData(index)">Remove</button>
      </li>
    </ul>

    <p v-if="!queriedData.length">No data points match the selected range.</p>
  </div>
</div>

    <!-- Solar Consumption Graph -->
    <div class="chart-container" v-if="currentGraph === 'consumption'">
      <BarGraph />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import Graph from './Graph.vue';
import BarGraph from './yieldGraph.vue';


import { useNexusStore } from '@/stores/nexus'

const store = useNexusStore()

const {
  VITE_NEXUS_API_URL,
} = import.meta.env;

const router = useRouter();
const currentGraph = ref('yield');
const showCalendar = ref(false);
const showPanelModal = ref(false);
const isLineChart = ref(false);
const activePanel = ref(0); // Index of the currently active panel
const newDate = ref("");
const newProduction = ref("");
const startDate = ref(null);
const endDate = ref(null);
const errorMessage = ref("");

const exportData = () => {
  const activeData = panels.value[activePanel.value].data;

  if (!activeData.length) {
    alert("No data to export for this panel.");
    return;
  }

  const csvContent = [
    "Date,Production (kWh)", // CSV header
    ...activeData.map(
      (point) => `${point.date.toISOString().split('T')[0]},${point.production}`
    ),
  ].join("\n");

  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
  link.href = url;
  link.setAttribute(
    "download",
    `${panels.value[activePanel.value].name}_data.csv`
  );
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// Panels data
const panels = ref([
  { name: "Panel 1", data: [], id: 1},
  { name: "Panel 2", data: [], id: 2},
  { name: "Panel 3", data: [], id: 3}
]);

watch(
  panels,
  (newPanels) => {
    localStorage.setItem("panels", JSON.stringify(newPanels));
  },
  { deep: true }
);

// Load panel data from local storage
 onMounted(async () => {

  const defaultPanelId = 1
  const startDate = '5/11/2024'
  const endDate = '5/12/2024'

  const response = await store.user.getPanelYieldData(defaultPanelId, startDate, endDate)
  const yieldData = response.yield_data
  panels.value[0].data = yieldData
});

// Watch for changes in panels and sync with local storage
watch(
  panels,
  (newPanels) => {
    localStorage.setItem('panels', JSON.stringify(newPanels));
  },
  { deep: true }
);

const graphKey = ref(0);

const refreshGraph = () => {
  graphKey.value += 1; // Changes the key to force a re-render
};

// Panel switching
const togglePanelModal = () => {
  showPanelModal.value = !showPanelModal.value;
};
const selectPanel = (index) => {
  activePanel.value = index;

  // Reset the date range and queriedData when switching panels
  startDate.value = null;
  endDate.value = null;
  queriedData.value = panels.value[index].data; // Show all data for the new panel initially
  showPanelModal.value = false;
};

// query data
const queriedData = ref([]); // Store the filtered data

watch(
  [startDate, endDate],
  ([newStartDate, newEndDate]) => {
    if (newStartDate && newEndDate) {
      const start = new Date(newStartDate).getTime();
      const end = new Date(newEndDate).getTime();

      // Debug: Log the start, end, and data points for verification
      console.log("Start Date:", newStartDate);
      console.log("End Date:", newEndDate);
      console.log("Original Data Points:", panels.value[activePanel.value].data);

      if (start > end) {
        errorMessage.value = "Start date must be earlier than end date.";
        queriedData.value = [];
      } else {
        errorMessage.value = "";
        queriedData.value = panels.value[activePanel.value].data.filter((point) => {
          const pointDate = new Date(point.date).getTime();
          return pointDate >= start && pointDate <= end;
        });

        // Debug: Log the filtered data points
        console.log("Filtered Data Points:", queriedData.value);
      }
    }
  },
  { immediate: true } // Ensure it runs initially on mount
);

// Calendar modal
const openCalendar = () => {
  const activePanelData = panels.value[activePanel.value].data;

  if (activePanelData.length === 0) {
    alert("No data available for this panel to query.");
    return;
  }

  // Determine the minimum and maximum dates from the active panel data
  const dates = activePanelData.map((point) => point.date);
  const minDate = new Date(Math.min(...dates));
  const maxDate = new Date(Math.max(...dates));

  // Set default start and end dates
  startDate.value = minDate.toISOString().split('T')[0];
  endDate.value = maxDate.toISOString().split('T')[0];

  showCalendar.value = true;
};

const closeCalendar = () => {
  // If both dates are selected, validate them
  if (startDate.value && endDate.value) {
    const selectedStartDate = new Date(startDate.value).getTime();
    const selectedEndDate = new Date(endDate.value).getTime();
    const minDate = new Date(panels.value[activePanel.value].data[0].date).getTime();
    const maxDate = new Date(panels.value[activePanel.value].data[panels.value[activePanel.value].data.length - 1].date).getTime();

    if (selectedStartDate < minDate || selectedEndDate > maxDate) {
      errorMessage.value = `Dates must be between ${new Date(minDate).toISOString().split('T')[0]} and ${new Date(maxDate).toISOString().split('T')[0]}.`;
      return;
    }

    errorMessage.value = ""; // Clear the error if dates are valid
  } else {
    // If either startDate or endDate is missing, just reset the error and close the modal
    errorMessage.value = "";
  }

  showCalendar.value = false; // Close the modal
};

// Format the date input as YYYY/MM/DD
const formatDateInput = () => {
  let date = newDate.value.replace(/\D/g, "");
  if (date.length >= 5 && parseInt(date.substring(0, 4), 10) > 2025) {
    alert("Year cannot exceed 2025.");
    newDate.value = "";
    return;
  }
  if (date.length > 4) date = date.slice(0, 4) + '/' + date.slice(4);
  if (date.length > 7) {
    const month = parseInt(date.slice(5, 7), 10);
    if (month > 12) {
      alert("Month cannot exceed 12.");
      newDate.value = date.slice(0, 5);
      return;
    }
    date = date.slice(0, 7) + '/' + date.slice(7);
  }
  if (date.length === 10) {
    const day = parseInt(date.slice(8, 10), 10);
    if (day > 31) {
      alert("Day cannot exceed 31.");
      newDate.value = date.slice(0, 8);
      return;
    }
  }
  newDate.value = date.slice(0, 10);
};

const revertToAllData = () => {
  queriedData.value = panels.value[activePanel.value].data;
};

const showAllData = ref(true); // State to toggle between all data and queried data
const toggleDataView = () => {
  showAllData.value = !showAllData.value;
};

// Remove data from the active panel
const removeData = async (dataID) => {
  try {
    const response = await fetch(`/api/panels/data/${dataID}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const error = await response.text();
      console.error("Failed to remove data:", error);
      return;
    }

    alert("Data removed successfully!");
    fetchPanelData(); // Refresh the data on the graph
  } catch (err) {
    console.error("Error removing data:", err);
  }
};

// Clear data for the active panel
const clearData = () => {
  panels.value[activePanel.value].data = [];
};

// Other existing methods
const switchGraph = (graphType) => {
  currentGraph.value = graphType;
};
const goTo = (path) => {
  router.push(path);
};
const allowOnlyNumbers = (event) => {
  const char = String.fromCharCode(event.which);
  if (!/[0-9.]/.test(char) || (char === '.' && newProduction.value.includes('.'))) {
    event.preventDefault();
  }
};

const addData = async () => {
  if (!newDate.value || !newProduction.value) {
    alert("Please provide both date and production values.");
    return;
  }

  try {
    const response = await fetch(`/api/panels/${activePanel.value}/data`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        date: newDate.value,
        production: parseFloat(newProduction.value),
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      console.error("Failed to add data:", error);
      return;
    }

    alert("Data added successfully!");
    fetchPanelData(); // Refresh the data on the graph
  } catch (err) {
    console.error("Error adding data:", err);
  }
};

const fetchPanelData = async () => {
  try {
    const response = await store.user.getPanelData(activePanel.value)
    if (!response.ok) {
      const error = await response.text();
      console.error("Failed to fetch data:", error);
      return;
    }

    const data = await response.json();
    panels[activePanel.value].data = data.map((d) => ({
      id: d.id,
      date: new Date(d.date),
      production: d.production,
    }));
    refreshGraph(); // Update the graph with the new data
  } catch (err) {
    console.error("Error fetching data:", err);
  }
};

</script>

<style scoped>
.toggle-data-button {
  margin-top: 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  padding: 10px 15px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.3s;
}

.toggle-data-button:hover {
  background-color: #0056b3;
  transform: scale(1.05);
}

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
.solar-panel-container {
  margin-top: 20px;
}

.solar-panels-button {
  padding: 10px 20px;
  font-size: 16px;
  cursor: pointer;
  background-color: #5e60ce;
  color: #ffffff;
  border: none;
  border-radius: 5px;
  transition: background-color 0.3s ease;
}

.solar-panels-button:hover {
  background-color: #4b4fb3;
}

.panel-modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  padding: 20px;
  background-color: #ffffff;
  border: 1px solid #d3d3d3;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  width: 300px;
  text-align: center;
  z-index: 1000; /* Ensure it’s above other elements */
}

.panel-modal h3 {
  margin-bottom: 15px;
}

.panel-modal button {
  display: block;
  width: 100%;
  padding: 10px;
  margin: 5px 0;
  background-color: #5e60ce;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.panel-modal button:hover {
  background-color: #4b4fb3;
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
