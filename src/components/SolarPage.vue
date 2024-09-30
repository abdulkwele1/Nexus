<template>
  <div class="container">
    <button class="home-button" @click="goHome">Home</button>
    <div ref="chartContainer" class="chart-container">
      <button class="current-date-button" @click="openCalendar">
        Select Date Range &#9662;
      </button>
      <button class="export-button" @click="exportData">
        📄 Export
      </button>
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
        <!-- Error message if end date is before start date -->
        <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
        <button @click="closeCalendar">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import * as d3 from 'd3';

const router = useRouter();
const chartContainer = ref(null);
const showCalendar = ref(false);
const startDate = ref(null);
const endDate = ref(null);
const solarData = ref([]); // Store solar data
const errorMessage = ref(""); // Store error message

const goHome = () => {
  router.push('/home');
};

const openCalendar = () => {
  showCalendar.value = true;
};

const closeCalendar = () => {
  showCalendar.value = false;
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
      production = Math.floor(Math.random() * 100) + 150; // Random value between 150 and 250 for summer months
    } else {
      production = Math.floor(Math.random() * 50) + 50; // Random value between 50 and 100 for other months
    }

    data.push({
      date: new Date(currentDate),
      production
    });

    // Increment the date by 1 day
    currentDate.setDate(currentDate.getDate() + 1);
  }

  return data;
};

// Function to handle date validation
const validateDates = () => {
  errorMessage.value = ""; // Clear any previous error message
  if (startDate.value && endDate.value) {
    const start = new Date(startDate.value);
    const end = new Date(endDate.value);

    // Check if the end date is before the start date
    if (end < start) {
      errorMessage.value = "End point cannot be before starting point.";
    } else {
      solarData.value = generateSolarData(start, end);
      createChart(); // Update the chart with new data
    }
  }
};

// Watch for changes in the start and end dates and regenerate the chart automatically
watch([startDate, endDate], validateDates);

// Create the bar chart with daily data points
const createChart = () => {
  d3.select(chartContainer.value).select("svg").remove();

  const width = 960;
  const height = 500;
  const margin = { top: 30, right: 20, bottom: 50, left: 150 };

  const x = d3.scaleBand()
    .domain(solarData.value.map(d => d3.timeFormat("%Y-%m-%d")(d.date)))
    .range([margin.left, width - margin.right])
    .padding(0.1);

  const y = d3.scaleLinear()
    .domain([0, d3.max(solarData.value, d => d.production)]).nice()
    .range([height - margin.bottom, margin.top]);

  const svg = d3.select(chartContainer.value)
    .append("svg")
    .attr("width", width)
    .attr("height", height);

  svg.append("g")
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x).tickValues(x.domain().filter((d, i) => !(i % Math.floor(solarData.value.length / 10)))))
    .selectAll("text")
    .attr("transform", "rotate(-45)")
    .style("text-anchor", "end");

  svg.append("g")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y));

  svg.selectAll(".bar")
    .data(solarData.value)
    .enter()
    .append("rect")
    .attr("class", "bar")
    .attr("x", d => x(d3.timeFormat("%Y-%m-%d")(d.date)))
    .attr("y", d => y(d.production))
    .attr("width", x.bandwidth())
    .attr("height", d => y(0) - y(d.production))
    .attr("fill", "#69b3a2");

  svg.append("text")
    .attr("transform", "rotate(-90)")
    .attr("x", -height / 2)
    .attr("y", margin.left / 2)
    .attr("dy", "-1em")
    .attr("fill", "currentColor")
    .attr("text-anchor", "middle")
    .text("(kWh)");
};

// Function to export data to CSV
const exportData = () => {
  // Define the header for the CSV
  const header = "sensor_reading_date, daily_kw_generated\n";
  
  // Generate the CSV content
  const csvContent = "data:text/csv;charset=utf-8," 
    + header // Add the header
    + solarData.value.map(d => `${d3.timeFormat("%Y-%m-%d")(d.date)},${d.production}`).join("\n");

  // Create the link element and trigger download
  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  link.setAttribute("download", "solar_data.csv");
  document.body.appendChild(link); // Required for Firefox
  link.click();
  document.body.removeChild(link); // Clean up the link after download
};

onMounted(() => {
  solarData.value = generateSolarData(new Date("2023-01-01"), new Date("2023-01-31"));
  createChart();
});
</script>

<style scoped>
.container {
  position: relative;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}

.chart-container {
  position: relative;
}

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

.calendar-container {
  display: flex;
  justify-content: space-between;
  margin: 20px 0;
}

.calendar {
  width: 45%;
}

.error-message {
  color: red;
  font-size: 14px;
}

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
</style>
