<template>
  <div class="soil-moisture-graph">
    <div class="sensor-toggles">
      <label v-for="(sensor, index) in sensors" :key="sensor.name" class="toggle-label">
        <input 
          type="checkbox" 
          v-model="sensor.visible" 
          :style="{ '--sensor-color': colors[index] }"
        >
        {{ sensor.name }}
      </label>
    </div>
    <div ref="chartContainer"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, defineExpose } from 'vue';
import * as d3 from 'd3';
import { useNexusStore } from '@/stores/nexus';

// Capture D3's default multi-scale local time formatter
const localTimeScale = d3.scaleTime();
const localTimeFormatter = localTimeScale.tickFormat();

interface DataPoint {
  time: Date;
  moisture: number;
}

interface SensorData {
  [key: string]: DataPoint[];
}

interface Sensor {
  data: DataPoint[];
  name: string;
  visible: boolean;
}

interface Props {
  queryParams: {
    startDate: string;
    endDate: string;
    minMoisture: number;
    maxMoisture: number;
    resolution: 'raw' | 'hourly' | 'daily' | 'weekly' | 'monthly';
  }
}

const props = defineProps<Props>();

const chartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);

const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

// Define SENSOR_CONFIGS (could be props later if more dynamic)
// Adjusted to match the number of colors and typical sensor setup
const SENSOR_CONFIGS = [
  { id: 444574498032128, name: 'Sensor Alpha' },
  { id: 2, name: 'Sensor 2' },
  { id: 3, name: 'Sensor 3' },
  { id: 4, name: 'Sensor 4' },
];

//possibly where I would put the real data in due time
// const sensors = ref<Sensor[]>([
//   { data: mockData.sensor1, name: 'Sensor 1', visible: true },
//   { data: mockData.sensor2, name: 'Sensor 2', visible: true },
//   { data: mockData.sensor3, name: 'Sensor 3', visible: true },
//   { data: mockData.sensor4, name: 'Sensor 4', visible: true }
// ]);

const nexusStore = useNexusStore();

// Initialize sensors ref. It will be populated asynchronously.
const sensors = ref<Sensor[]>(SENSOR_CONFIGS.map(config => ({
  data: [], // Start with empty data
  name: config.name,
  visible: true, // Default visibility
})));

async function fetchAllSensorData() {
  try {
    const allSensorsDataPromises = SENSOR_CONFIGS.map(async (config, index) => {
      // Assuming getSensorMoistureData fetches all data if startDate/endDate are undefined.
      // The API response `responseData.sensor_moisture_data` is an array of raw data points.
      // Each point should have 'date' and 'soil_moisture'.
      const rawDataPoints = await nexusStore.user.getSensorMoistureData(config.id);

      // Log raw data points here
      console.log(`[soilMoistureGraph] Raw data for ${config.name} (ID: ${config.id}):`, JSON.parse(JSON.stringify(rawDataPoints)));

      if (rawDataPoints && Array.isArray(rawDataPoints)) {
        const formattedData: DataPoint[] = rawDataPoints.map((point: any) => ({
          // Assuming point.date from API is a string (e.g., ISO8601) and point.soil_moisture is a number
          time: new Date(point.date), 
          moisture: Number(point.soil_moisture), 
        })).sort((a, b) => a.time.getTime() - b.time.getTime()); // Sort data by time ascending
        sensors.value[index].data = formattedData;
      } else {
        console.warn(`No data returned for sensor ${config.name} (ID: ${config.id}) or data is not an array. API returned:`, rawDataPoints);
        sensors.value[index].data = []; // Ensure data is an empty array on error/no data
      }
    });
    await Promise.all(allSensorsDataPromises);
  } catch (error) {
    console.error("Error fetching sensor data:", error);
    sensors.value.forEach(sensor => sensor.data = []);
  } finally {
    createChart(); // Re-create chart once all data is fetched or after an error
  }
}

const formatDate = (date: Date) => {
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

const formatValue = (value: number) => {
  return `${value.toFixed(1)}%`;
};

const createChart = () => {
  if (!chartContainer.value) return;

  // Clear previous chart
  if (svg.value) {
    svg.value.remove();
  }

  // Chart dimensions and margins
  const width = 928;
  const height = 500;
  const marginTop = 20;
  const marginRight = 30;
  const marginBottom = 30;
  const marginLeft = 40;

  // Create SVG container
  svg.value = d3.create("svg")
    .attr("viewBox", [0, 0, width, height])
    .attr("width", width)
    .attr("height", height)
    .attr("style", "max-width: 100%; height: auto; height: intrinsic; font: 10px sans-serif;")
    .style("-webkit-tap-highlight-color", "transparent")
    .style("overflow", "visible");

  // Combine visible sensor data for domain calculation
  const allData = sensors.value
    .filter(sensor => sensor.visible)
    .flatMap(sensor => sensor.data);

  // Create scales
  const x = d3.scaleUtc()
    .domain(d3.extent(allData, d => d.time) as [Date, Date])
    .range([marginLeft, width - marginRight]);

  const y = d3.scaleLinear()
    .domain([0, d3.max(allData, d => d.moisture) as number])
    .range([height - marginBottom, marginTop]);

  // Create line generator
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(d.moisture));

  // Determine X-axis tick formatter based on resolution
  let xAxisTickFormat = localTimeFormatter; // Default for raw/hourly
  if (props.queryParams.resolution === 'daily' || 
      props.queryParams.resolution === 'weekly' || 
      props.queryParams.resolution === 'monthly') {
    // For daily, weekly, monthly, use a format that emphasizes date
    // D3's default scaleTime().tickFormat() is adaptive and good for dates too
    // but we can be more specific if needed, e.g., d3.timeFormat("%Y-%m-%d")
    // For now, let D3's adaptive localTimeFormatter handle it, it should prioritize date parts for wider spans.
    // If more specific formatting is needed, we can use d3.timeFormat here.
    // Example: xAxisTickFormat = d3.timeFormat("%x"); // Locale's date, e.g., 01/15/2024
  }

  // Add x-axis
  svg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x) 
        .ticks(width / 80)
        .tickSizeOuter(0)
        .tickFormat(xAxisTickFormat) 
    );

  // Add y-axis
  svg.value.append("g")
    .attr("transform", `translate(${marginLeft},0)`)
    .call(d3.axisLeft(y).ticks(height / 40))
    .call(g => g.select(".domain").remove())
    .call(g => g.selectAll(".tick line").clone()
      .attr("x2", width - marginLeft - marginRight)
      .attr("stroke-opacity", 0.1))
    .call(g => g.append("text")
      .attr("x", -marginLeft)
      .attr("y", 10)
      .attr("fill", "currentColor")
      .attr("text-anchor", "start")
      .text("↑ Soil Moisture (%)"));

  // Create tooltip container
  tooltip.value = svg.value.append("g")
    .attr("class", "tooltip")
    .style("display", "none");

  // Add lines for each visible sensor
  sensors.value.forEach((sensor, i) => {
    if (!sensor.visible) return;

    const path = svg.value!.append("path")
      .datum(sensor.data)
      .attr("fill", "none")
      .attr("stroke", colors[i])
      .attr("stroke-width", 1.5)
      .attr("d", line);

    // Add hover effect
    path.on("mouseover", function() {
      d3.select(this)
        .attr("stroke-width", 2.5);
    })
    .on("mouseout", function() {
      d3.select(this)
        .attr("stroke-width", 1.5);
    });
  });

  // Add legend
  const legend = svg.value.append("g")
    .attr("font-family", "sans-serif")
    .attr("font-size", 10)
    .attr("text-anchor", "start")
    .selectAll("g")
    .data(sensors.value.filter(s => s.visible))
    .join("g")
    .attr("transform", (d, i) => `translate(0,${i * 20})`);

  legend.append("rect")
    .attr("x", width - 19)
    .attr("width", 19)
    .attr("height", 19)
    .attr("fill", (d, i) => colors[i]);

  legend.append("text")
    .attr("x", width - 24)
    .attr("y", 9.5)
    .attr("dy", "0.32em")
    .text(d => d.name);

  // Add the chart to the container
  chartContainer.value.appendChild(svg.value.node());

  // Add mouse interaction for tooltip
  const bisect = d3.bisector<DataPoint, Date>(d => d.time).center;
  
  svg.value.on("pointermove", (event) => {
    if (!tooltip.value || !sensors.value.some(s => s.visible && s.data.length > 0)) return; // Ensure there's data

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    
    // Find the closest data point for each visible sensor
    const tooltipData = sensors.value
      .filter(sensor => sensor.visible && sensor.data.length > 0) // Ensure sensor has data
      .map((sensor, i) => {
        const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === sensor.name);
        const color = colors[sensorColorIndex % colors.length]; // Get color safely

        const index = bisect(sensor.data, xPos);
        // Ensure index is within bounds
        const dataPoint = sensor.data[Math.max(0, Math.min(index, sensor.data.length - 1))];
        if (!dataPoint) return null; // Should not happen if data.length > 0

        return {
          name: sensor.name,
          value: dataPoint.moisture,
          time: dataPoint.time,
          color: color
        };
      }).filter(item => item !== null); // Remove nulls if any

    if (tooltipData.length === 0) {
      tooltip.value.style("display", "none");
      return;
    }

    // Show tooltip
    tooltip.value.style("display", null);
    tooltip.value.attr("transform", `translate(${pointer[0]},${pointer[1]})`);

    // Update tooltip content
    const tooltipContent = tooltip.value.selectAll("g")
      .data(tooltipData)
      .join("g")
      .attr("transform", (d, i) => `translate(0,${i * 20})`);

    tooltipContent.selectAll("rect")
      .data(d => [d])
      .join("rect")
      .attr("x", -60)
      .attr("y", -15)
      .attr("width", 120)
      .attr("height", 20)
      .attr("fill", "white")
      .attr("stroke", "black")
      .attr("stroke-width", 0.5);

    tooltipContent.selectAll("text")
      .data(d => [d])
      .join("text")
      .attr("x", -55)
      .attr("y", 0)
      .attr("dy", "0.32em")
      .text(d => `${d.name}: ${formatValue(d.value)}`)
      .attr("fill", d => d.color);
  })
  .on("pointerleave", () => {
    if (tooltip.value) {
      tooltip.value.style("display", "none");
    }
  });
};

// Watch for changes in sensor visibility
watch(() => sensors.value.map(s => s.visible), () => {
  createChart();
}, { deep: true });

// Add watch for queryParams
watch(() => props.queryParams, () => {
  // Filter and process data based on query params
  const filteredData = filterData(props.queryParams);
  updateChart(filteredData);
}, { deep: true });

// Add data filtering function
const filterData = (params: Props['queryParams']) => {
  const startDate = new Date(params.startDate);
  const endDate = new Date(params.endDate);
  const minMoisture = params.minMoisture;
  const maxMoisture = params.maxMoisture;

  // Log the filter boundaries
  console.log('[soilMoistureGraph] Filtering with Start:', startDate, ' (ISO:', startDate.toISOString(), ') End:', endDate, ' (ISO:', endDate.toISOString(), ') MinM:', minMoisture, 'MaxM:', maxMoisture);
  console.log('[soilMoistureGraph] Query Params received:', JSON.stringify(params));

  const filtered = sensors.value.map(sensor => {
    // Log a few original data points for this sensor before filtering
    if (sensor.data.length > 0) {
      console.log(`[soilMoistureGraph] Original data for ${sensor.name} (first 3 points):`, 
        JSON.stringify(sensor.data.slice(0, 3).map(p => ({ time: p.time.toISOString(), moisture: p.moisture })))
      );
    }

    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time; // point.time is already a Date object
      const moisture = point.moisture;
      
      const isAfterStartDate = date >= startDate;
      const isBeforeEndDate = date <= endDate;
      const isAboveMinMoisture = moisture >= minMoisture;
      const isBelowMaxMoisture = moisture <= maxMoisture;
      
      const includePoint = isAfterStartDate && isBeforeEndDate && isAboveMinMoisture && isBelowMaxMoisture;
      
      // Log decision for a sample of points
      // if (Math.random() < 0.01) { // Adjust frequency as needed
      //   console.log(`[soilMoistureGraph] Point: ${date.toISOString()} (${moisture}), Start: ${startDate.toISOString()}, End: ${endDate.toISOString()}, MinM: ${minMoisture}, MaxM: ${maxMoisture}, Included: ${includePoint}`);
      // }
      return includePoint;
    });

    console.log(`[soilMoistureGraph] Sensor ${sensor.name} - Original points: ${sensor.data.length}, Filtered points: ${sensorFilteredData.length}`);

    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  // Apply resolution aggregation if needed
  if (params.resolution !== 'raw') {
    return aggregateData(filtered, params.resolution);
  }

  return filtered;
};

// Add data aggregation function
const aggregateData = (sensorsToAggregate: Sensor[], resolution: 'hourly' | 'daily' | 'weekly' | 'monthly'): Sensor[] => {
  return sensorsToAggregate.map(sensor => {
    if (!sensor.data || sensor.data.length === 0) {
      return { ...sensor, data: [] };
    }

    const aggregatedData: DataPoint[] = [];
    const groups = new Map<string, { sum: number, count: number, time: Date }>();

    sensor.data.forEach(point => {
      let key = '';
      const date = new Date(point.time); // Ensure it's a Date object

      switch (resolution) {
        case 'hourly':
          key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}-${date.getHours()}`;
          break;
        case 'daily':
          key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
          break;
        case 'weekly':
          const weekStartDate = d3.timeWeek.floor(date);
          key = `${weekStartDate.getFullYear()}-${d3.timeWeek.count(d3.timeYear.floor(weekStartDate), weekStartDate)}`;
          break;
        case 'monthly':
          key = `${date.getFullYear()}-${date.getMonth()}`;
          break;
      }

      if (!groups.has(key)) {
        let groupTime = d3.timeHour.floor(date); // Default
        if (resolution === 'daily') {
          groupTime = d3.timeDay.floor(date);
        } else if (resolution === 'weekly') {
          groupTime = d3.timeWeek.floor(date);
        } else if (resolution === 'monthly') {
          groupTime = d3.timeMonth.floor(date);
        }
        groups.set(key, { sum: 0, count: 0, time: groupTime });
      }
      
      const group = groups.get(key)!;
      group.sum += point.moisture;
      group.count += 1;
    });

    groups.forEach(group => {
      aggregatedData.push({
        time: group.time,
        moisture: group.sum / group.count,
      });
    });

    // Sort by time, as grouping might change order
    aggregatedData.sort((a, b) => a.time.getTime() - b.time.getTime());

    return { ...sensor, data: aggregatedData };
  });
};

// Modify createChart to accept filtered data
const updateChart = (filteredData: Sensor[]) => {
  sensors.value = filteredData;
  createChart();
};

// Update the getFilteredData function to return only currently visible data
const getFilteredData = () => {
  const filteredData = filterData(props.queryParams);
  // Only return sensors that are currently visible
  return filteredData.filter(sensor => sensor.visible);
};

defineExpose({
  getFilteredData
});

onMounted(async () => {
  await fetchAllSensorData();
  // createChart(); // createChart is now called within fetchAllSensorData
});
</script>

<style scoped>
.soil-moisture-graph {
  width: 100%;
  max-width: 928px;
  margin: 0 auto;
  padding: 20px;
}

.sensor-toggles {
  margin-bottom: 20px;
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.toggle-label input[type="checkbox"] {
  appearance: none;
  width: 16px;
  height: 16px;
  border: 2px solid var(--sensor-color);
  border-radius: 4px;
  cursor: pointer;
  position: relative;
}

.toggle-label input[type="checkbox"]:checked {
  background-color: var(--sensor-color);
}

.toggle-label input[type="checkbox"]:checked::after {
  content: "✓";
  position: absolute;
  color: white;
  font-size: 12px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.tooltip {
  pointer-events: none;
}
</style>