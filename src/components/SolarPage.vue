    <template>
    <div class="container">
        <button class="home-button" @click="goHome">Home</button>
        <div ref="chartContainer" class="chart-container">
        <button class="current-date-button" @click="openCalendar">
            Current Date &#9662; <!-- Down arrow icon -->
        </button>
        <button class="export-button" @click="exportData">
            📄 Export
        </button>
        </div> <!-- Chart container -->

        <div v-if="showCalendar" class="modal-overlay" @click="closeCalendar">
        <div class="modal" @click.stop>
            <h2>Select Dates</h2>
            <div class="calendar-container">
            <div class="calendar">
                <h3>Calendar 1</h3>
                <input type="date" v-model="selectedDate1" />
            </div>
            <div class="calendar">
                <h3>Calendar 2</h3>
                <input type="date" v-model="selectedDate2" />
            </div>
            </div>
            <button @click="closeCalendar">Close</button>
        </div>
        </div>
    </div>
    </template>


    <script setup lang="ts">
    import { useRouter } from 'vue-router';
    import * as d3 from 'd3'; // Import D3.js
    import { onMounted, ref } from 'vue';

    const router = useRouter(); // Create router instance
    const chartContainer = ref(null); // Reference for the chart container

    const showCalendar = ref(false); // State to control modal visibility
    const selectedDate1 = ref(null); // State for the first calendar date
    const selectedDate2 = ref(null); // State for the second calendar date

    const goHome = () => {
    router.push('/home'); // Adjust the path as necessary
    };

    const openCalendar = () => {
    showCalendar.value = true; // Show the calendar modal
    };

    const closeCalendar = () => {
    showCalendar.value = false; // Hide the calendar modal
    };

    const onCurrentDateClick = () => {
    console.log('Current Date button clicked');
    };

    const exportData = () => {
    console.log('Export Data button clicked');
    // Implement data export logic here
    };

    const createChart = () => {
    const width = 960;
    const height = 500;
    const marginTop = 30;
    const marginRight = 20;
    const marginBottom = 30;
    const marginLeft = 150;

    const x = d3.scaleUtc()
        .domain([new Date("2023-01-01"), new Date("2024-01-01")])
        .range([marginLeft, width - marginRight]);

    const y = d3.scaleLinear()
        .domain([0, 100])
        .range([height - marginBottom, marginTop]);

    const svg = d3.create("svg")
        .attr("width", width)
        .attr("height", height);

    svg.append("g")
        .attr("transform", `translate(0,${height - marginBottom})`)
        .call(d3.axisBottom(x));

    svg.append("g")
        .attr("transform", `translate(${marginLeft},0)`)
        .call(d3.axisLeft(y));

    svg.append("text")
        .attr("transform", "rotate(-90)")
        .attr("x", -height / 2)
        .attr("y", marginLeft / 2)
        .attr("dy", "-1em")
        .attr("fill", "currentColor")
        .attr("text-anchor", "middle")
        .text("(kWh)");

    chartContainer.value.appendChild(svg.node());
    };

    onMounted(() => {
    createChart(); // Create the chart when the component mounts
    });
    </script>



    <style scoped>
    .container {
    position: relative; /* Positioning context for absolute elements */
    height: 100vh; /* Full viewport height */
    display: flex;
    justify-content: center; /* Center the chart */
    align-items: center; /* Center the chart */
    }

    .chart-container {
    position: relative; /* Positioning context for the button */
    }

    /* Modal Styles */
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

    /* Style for the home button */
    .home-button {
    position: absolute; /* Keep button in top left */
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

    /* Style for the current date button */
    .current-date-button {
    position: absolute; /* Position relative to the chart */
    top: -25px; /* Adjust this value to position it above the graph */
    left: 125px; /* Adjust this value to align with the y-axis */
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

    /* Style for the export button */
    .export-button {
    position: absolute; /* Position relative to the chart */
    bottom: -40px; /* Adjust this value to position it above the x-axis */
    right: 0; /* Adjust this value to position it on the right */
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

