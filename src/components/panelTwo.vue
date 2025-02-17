<template>
  <div class="chart-container">
    <h3>Solar Panel 2 Data</h3>
    <Graph
    :key="graphKey"
    :"solarData"
    :isLineChart="isLineChart">
    </Graph>

    <!-- Line chart switch button -->
    <button class="line-chart-toggle-button" @click="isLineChart = !isLineChart">
      {{ isLineChart ? "Bar chart" : "Line chart" }}
    </button>

    <!-- Export button -->
    <button class="export-button" @click="exportData">📄 Export</button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import Graph from './Graph.vue';

const isLineChart = ref(false);
const solarData = ref([]);

const graphKey = ref(0);

const refreshGraph = () => {
  graphKey.value += 1;
}


const addData = () => {
  // Validate the inputs
  if (!newDate.value.trim() || !newProduction.value.trim()) {
    alert("Please enter both a valid date and production value.");
    return;
  }

  // Parse the date
  const parsedDate = new Date(newDate.value);
  if (isNaN(parsedDate.getTime())) {
    alert("Invalid date format. Please use YYYY/MM/DD.");
    return;
  }

  // Parse the production value
  const productionValue = parseFloat(newProduction.value);
  if (isNaN(productionValue) || productionValue <= 0) {
    alert("Production value must be a positive number.");
    return;
  }

  // Add the data point to the current panel
  panels.value[activePanel.value].data.push({
    date: parsedDate,
    production: productionValue,
  });

  // Clear input fields
  newDate.value = "";
  newProduction.value = "";

  // Debugging log
  console.log("New Data Added:", panels.value[activePanel.value].data);

  // Refresh the graph
  refreshGraph();
};

onMounted(() => {
  // Load data specific to Solar Panel 2 from local storage
  const savedData = JSON.parse(localStorage.getItem("solarDataPanel2") || "[]");
  solarData.value = savedData.map((data) => ({
    date: new Date(data.date),
    production: data.production,
  }));
});

const exportData = () => {
  // Export Solar Panel 2 data
  const data = JSON.stringify(solarData.value, null, 2);
  const blob = new Blob([data], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "solar_data_panel2.json";
  link.click();
};
</script>

<style scoped>
/* Add custom styling for Panel Two here */
.chart-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.line-chart-toggle-button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.line-chart-toggle-button:hover {
  background-color: #0056b3;
}

.export-button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #28a745;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.export-button:hover {
  background-color: #218838;
}
</style>
