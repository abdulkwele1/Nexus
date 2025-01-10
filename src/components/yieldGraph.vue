<template>
  <div class="yield-graph">
    <!-- Canvas for the graph -->
    <canvas ref="yieldGraph"></canvas>

    <!-- Export Button -->
    <button class="export-button" @click="exportData">📄 Export</button>

    <!-- Dropdown Calendar Button -->
    <button class="calendar-button" @click="toggleCalendar">Select Date Range &#9662;</button>

    <!-- Calendar Modal for selecting date range -->
    <div v-if="showCalendar" class="modal-overlay" @click="toggleCalendar">
      <div class="modal" @click.stop>
        <h2>Select Date Range</h2>
        <div class="calendar-container">
          <div class="calendar">
            <label>Start point:</label>
            <input type="date" v-model="startDate" />
          </div>
          <div class="calendar">
            <label>End point:</label>
            <input type="date" v-model="endDate" />
          </div>
        </div>
        <button @click="toggleCalendar">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { Chart } from 'chart.js/auto';



// Refs for data and labels
const yieldGraph = ref<HTMLCanvasElement | null>(null);
const electricityUsageData = ref<number[]>([70, 50, 90, 60]); // Example percentage data for electricity usage
const directUsageData = ref<number[]>([60, 40, 80, 50]); // Example percentage data for direct usage
const labels = ref<string[]>([]); // Updated to be dynamic with dates

// Calendar date selection
const startDate = ref<string>('');
const endDate = ref<string>('');
const showCalendar = ref(false); // Controls calendar modal visibility

// Function to toggle the calendar modal
const toggleCalendar = () => {
  showCalendar.value = !showCalendar.value;
};

// Function to generate labels with panel names and dates
const generateLabels = () => {
  // Example logic: Display panels with selected date range
  labels.value = [
    `Panel 1 (${startDate.value} - ${endDate.value})`,
    `Panel 2 (${startDate.value} - ${endDate.value})`,
    `Panel 3 (${startDate.value} - ${endDate.value})`,
    `Panel 4 (${startDate.value} - ${endDate.value})`,
  ];
};

// Function to render the chart
let chartInstance: Chart | null = null; // Variable to store chart instance
const renderChart = () => {
  const ctx = yieldGraph.value?.getContext('2d');
  if (!ctx) return;

  // Destroy previous chart if it exists
  if (chartInstance) {
    chartInstance.destroy();
  }

  chartInstance = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels.value,
      datasets: [
        {
          label: 'Electricity Used (%)',
          data: electricityUsageData.value,
          backgroundColor: '#007bff', // Blue for electricity usage
        },
        {
          label: 'Direct Usage (%)',
          data: directUsageData.value,
          backgroundColor: '#ffc107', // Yellow for direct usage
        },
      ],
    },
    options: {
      responsive: true,
      scales: {
        y: {
          beginAtZero: true,
          max: 100,
        },
      },
    },
  });
};

// Function to generate random data based on the date range
const generateRandomData = () => {
  // Generate random electricity usage data
  electricityUsageData.value = myApiResponse.map(item => item.capacityKw);

  // Ensure direct usage does not exceed electricity usage
  directUsageData.value = apiResponse.map(item => item.consumedKw);
};

// Watch for changes in the startDate and endDate
watch([startDate, endDate], ([newStartDate, newEndDate]) => {
  if (newStartDate && newEndDate) {
    console.log(`Fetching data between ${newStartDate} and ${newEndDate}`);
    generateLabels();  // Generate labels with date range
    generateRandomData();
    updateGraphData(); // Automatically update the graph when dates are selected
  }
});

// Function to update the graph with new data
const updateGraphData = () => {
  console.log("Updating graph with new data from", startDate.value, "to", endDate.value);
  renderChart(); // Re-render chart with updated data
};

// Function to export data as CSV
const exportData = () => {
  const csvContent = "data:text/csv;charset=utf-8,Panel,Electricity Used (%),Direct Usage (%)\n"
    + labels.value.map((label, index) => `${label},${electricityUsageData.value[index]},${directUsageData.value[index]}`).join("\n");

  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  link.setAttribute("download", "Consumption_data.csv");
  document.body.appendChild(link); 
  link.click();
  document.body.removeChild(link);
};

// Mount the chart on component load
onMounted(async() => {
  renderChart();
});
</script>

<style scoped>
.yield-graph {
  width: 900px;
  height: 600px;
  margin: auto;
  position: relative;
}

/* Export Button Styling */
.export-button {
  position: fixed;
  top: 90px;
  right: 30px;
  z-index: 1001;
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.export-button:hover {
  background-color: #218838;
}

/* Calendar Button Styling */
.calendar-button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.calendar-button:hover {
  background-color: #0056b3;
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

.calendar label {
  font-weight: bold;
}
</style>
