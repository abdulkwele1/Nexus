<template>
  <div class="soil-moisture-graph">
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
}

interface Props {
  queryParams: {
    startDate: string;
    endDate: string;
    minValue: number;
    maxValue: number;
    resolution: 'raw' | 'hourly' | 'daily' | 'weekly' | 'monthly';
  }
  sensorVisibility: { [key: string]: boolean };
  dynamicTimeWindow?: 'none' | 'lastHour' | 'last24Hours' | 'last7Days' | 'last30Days';
  dataType: 'moisture' | 'temperature';
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
  // { id: 2, name: 'Sensor 2' }, // Temporarily commented out
  // { id: 3, name: 'Sensor 3' }, // Temporarily commented out
  // { id: 4, name: 'Sensor 4' }, // Temporarily commented out
];

//possibly where I would put the real data in due time
// const sensors = ref<Sensor[]>([
//   { data: mockData.sensor1, name: 'Sensor 1', visible: true },
//   { data: mockData.sensor2, name: 'Sensor 2', visible: true },
//   { data: mockData.sensor3, name: 'Sensor 3', visible: true },
//   { data: mockData.sensor4, name: 'Sensor 4', visible: true }
// ]);

const nexusStore = useNexusStore();

// Initialize sensors ref without internal visibility
const sensors = ref<Sensor[]>(SENSOR_CONFIGS.map((config) => ({
  data: [], 
  name: config.name,
})));

async function fetchAllSensorData() {
  let fetchError = false;
  try {
    console.log(`[soilMoistureGraph] Fetching data with resolution: ${props.queryParams.resolution}`);
    const allSensorsDataPromises = SENSOR_CONFIGS.map(async (config, index) => {
      let rawDataPoints;
      if (props.dataType === 'moisture') {
        rawDataPoints = await nexusStore.user.getSensorMoistureData(config.id);
      } else {
        rawDataPoints = await nexusStore.user.getSensorTemperatureData(config.id);
      }
      
      if (rawDataPoints && Array.isArray(rawDataPoints)) {
        const formattedData: DataPoint[] = rawDataPoints.map((point: any) => ({
          time: new Date(point.date),
          moisture: props.dataType === 'moisture' ? Number(point.soil_moisture) : Number(point.soil_temperature),
        })).sort((a, b) => a.time.getTime() - b.time.getTime());
        
        sensors.value[index].data = formattedData;
        console.log(`[soilMoistureGraph] Fetched ${formattedData.length} points for sensor ${config.name}`);
      } else {
        console.warn(`No data for sensor ${config.name}`);
        sensors.value[index].data = [];
      }
    });
    await Promise.all(allSensorsDataPromises);
  } catch (error) {
    console.error("Error fetching sensor data:", error);
    fetchError = true;
    sensors.value.forEach(sensor => sensor.data = []);
  } finally {
    if (!fetchError) {
      processAndDrawChart();
    } else {
      createChart();
    }
  }
}

// Function to process data based on current state and draw
const processAndDrawChart = () => {
  console.log(`[soilMoistureGraph] Processing data with resolution: ${props.queryParams.resolution}`);
  const processedData = filterData(props.queryParams);
  console.log(`[soilMoistureGraph] Processed data points:`, 
    processedData.reduce((sum, s) => sum + s.data.length, 0));
  updateChart(processedData);
}

const formatDate = (date: Date) => {
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  }).replace(' ', ', ');
};

const formatValue = (value: number) => {
  return props.dataType === 'moisture' ? `${value.toFixed(1)}%` : `${value.toFixed(1)}°C`;
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
    .filter(sensor => props.sensorVisibility[sensor.name])
    .flatMap(sensor => sensor.data);

  let yMinActual = d3.min(allData, d => d.moisture) as number;
  let yMaxActual = d3.max(allData, d => d.moisture) as number;

  // Handle cases where there's no data or min/max are undefined
  if (allData.length === 0 || yMinActual === undefined || yMaxActual === undefined) {
    yMinActual = 0;
    yMaxActual = 100;
  }

  const dataRange = yMaxActual - yMinActual;
  let yDomainMinWithPadding = Math.max(0, yMinActual - (dataRange * 0.1));
  let yDomainMaxWithPadding = Math.min(100, yMaxActual + (dataRange * 0.1));

  // Create scales
  const x = d3.scaleUtc()
    .range([marginLeft, width - marginRight]);

  let currentXDomain = d3.extent(allData, d => d.time) as [Date, Date];
  if (!currentXDomain[0] || !currentXDomain[1]) { // No data after filtering
    const today = new Date();
    if (props.queryParams.resolution === 'weekly') {
      currentXDomain = [d3.timeWeek.floor(today), d3.timeWeek.offset(d3.timeWeek.floor(today), 1)];
    } else if (props.queryParams.resolution === 'monthly') {
      currentXDomain = [d3.timeMonth.floor(today), d3.timeMonth.offset(d3.timeMonth.floor(today), 1)];
    } else {
      currentXDomain = [today, d3.timeDay.offset(today, 1)];
    }
  } else if (currentXDomain[0].getTime() === currentXDomain[1].getTime()) { // Single data point
    const singleDate = currentXDomain[0];
    if (props.queryParams.resolution === 'weekly') {
      currentXDomain = [d3.timeWeek.floor(singleDate), d3.timeWeek.offset(d3.timeWeek.floor(singleDate), 1)];
    } else if (props.queryParams.resolution === 'monthly') {
      currentXDomain = [d3.timeMonth.floor(singleDate), d3.timeMonth.offset(d3.timeMonth.floor(singleDate), 1)];
    } else if (props.queryParams.resolution === 'daily') {
      currentXDomain = [d3.timeDay.floor(singleDate), d3.timeDay.offset(d3.timeDay.floor(singleDate), 1)];
    } else { // hourly or raw, expand by a few hours
      currentXDomain = [d3.timeHour.offset(singleDate, -1), d3.timeHour.offset(singleDate, 1)];
    }
  }
  x.domain(currentXDomain);

  const y = d3.scaleLinear()
    .domain([yDomainMinWithPadding, yDomainMaxWithPadding])
    .range([height - marginBottom, marginTop]);

  // Create line generator
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(d.moisture));

  // Determine if we should show time based on data range and resolution
  const timeRange = x.domain()[1].getTime() - x.domain()[0].getTime();
  const isWithin24Hours = timeRange <= 24 * 60 * 60 * 1000;

  // Add x-axis
  const xAxisGroup = svg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`);
  
  let xAxis = d3.axisBottom(x);
  let tickFormat: (date: Date) => string;

  // --- Specific handling for last24Hours --- 
  if (props.dynamicTimeWindow === 'last24Hours') {
    xAxis.ticks(d3.timeHour.every(3)); // Ticks every 3 hours
    tickFormat = d3.timeFormat("%I %p"); // Format like "02 PM"
    xAxis.tickFormat(tickFormat as any);
    console.log("[soilMoistureGraph] Applying specific x-axis formatting for last24Hours.");
  } else {
  // --- Fallback to resolution-based formatting --- 
    switch (props.queryParams.resolution) {
      case 'monthly':
        xAxis.ticks(d3.timeMonth.every(1));
        tickFormat = d3.timeFormat("%b %Y"); // e.g., "Jan 2023"
        xAxis.tickFormat(tickFormat as any);
        break;
      case 'weekly':
        xAxis.ticks(d3.timeWeek.every(1));
        tickFormat = d3.timeFormat("%b %d"); // e.g., "Jan 15" (start of week)
        xAxis.tickFormat(tickFormat as any);
        break;
      case 'daily':
        const daysInRange = Math.ceil(timeRange / (24 * 60 * 60 * 1000));
        xAxis.ticks(Math.min(7, daysInRange > 0 ? daysInRange : 1)); // Limit to 7 ticks for daily
        tickFormat = d3.timeFormat("%b %d");
        xAxis.tickFormat(tickFormat as any);
        break;
      case 'hourly':
        // Apply 3-hour ticks for general hourly if not last24h? Or keep more granular?
        // Let's try keeping it more granular for the general 'hourly' setting
        const hoursInRange = Math.ceil(timeRange / (60 * 60 * 1000));
        xAxis.ticks(Math.min(12, hoursInRange > 0 ? hoursInRange : 1)); // Adjust tick count for hourly
        tickFormat = d3.timeFormat("%I:%M %p");
        xAxis.tickFormat(tickFormat as any);
        break;
      default: // 'raw'
        xAxis.ticks(width / 80); // Default tick count for raw data
        tickFormat = isWithin24Hours ? 
          d3.timeFormat("%I:%M %p") : 
          d3.timeFormat("%b %d");
        xAxis.tickFormat(tickFormat as any);
    }
  }
  xAxisGroup.call(xAxis);

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
    if (!props.sensorVisibility[sensor.name]) return;

    const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === sensor.name);
    const color = colors[sensorColorIndex % colors.length];

    const path = svg.value!.append("path")
      .datum(sensor.data)
      .attr("fill", "none")
      .attr("stroke", color)
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
    .data(sensors.value.filter(s => props.sensorVisibility[s.name]))
    .join("g")
    .attr("transform", (d, i) => `translate(0,${i * 20})`);

  legend.append("rect")
    .attr("x", width - 19)
    .attr("width", 19)
    .attr("height", 19)
    .attr("fill", (d) => {
      const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === d.name);
      return colors[sensorColorIndex % colors.length];
    });

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
    const visibleSensorsWithData = sensors.value.filter(s => props.sensorVisibility[s.name] && s.data.length > 0);
    if (!tooltip.value || visibleSensorsWithData.length === 0) {
      if (tooltip.value) tooltip.value.style("display", "none");
      return;
    }

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    
    const tooltipData = visibleSensorsWithData
      .map((sensor) => {
        const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === sensor.name);
        const color = colors[sensorColorIndex % colors.length];

        const index = bisect(sensor.data, xPos);
        const dataPoint = sensor.data[Math.max(0, Math.min(index, sensor.data.length - 1))];
        if (!dataPoint) return null;

        return {
          name: sensor.name,
          value: dataPoint.moisture,
          time: dataPoint.time,
          color: color
        };
      }).filter(item => item !== null);

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
      .attr("fill", d => d.color)
      .selectAll("tspan")
      .data(d_text => [
        `${d_text.name}: ${formatValue(d_text.value)}`,
        formatDate(d_text.time)
      ])
      .join("tspan")
      .attr("x", -55)
      .attr("dy", (d_tspan, i_tspan) => i_tspan === 0 ? "-0.1em" : "1.2em")
      .text(d_tspan => d_tspan);
  })
  .on("pointerleave", () => {
    if (tooltip.value) {
      tooltip.value.style("display", "none");
    }
  });
};

// Add watch for queryParams
watch(() => props.queryParams, (newParams) => {
  console.log(`[soilMoistureGraph] Query params changed. New resolution: ${newParams.resolution}`);
  processAndDrawChart();
}, { deep: true });

// Add watch for sensorVisibility prop (NEW)
watch(() => props.sensorVisibility, () => {
  console.log('[soilMoistureGraph] sensorVisibility prop changed. Triggering createChart...');
  // Just need to redraw, filtering happens within createChart based on the new prop
  createChart(); 
}, { deep: true });

// --- Watch for dynamicTimeWindow prop --- 
watch(() => props.dynamicTimeWindow, () => {
  console.log(`[soilMoistureGraph] dynamicTimeWindow prop changed to: ${props.dynamicTimeWindow}. Triggering chart processing...`);
  processAndDrawChart(); 
});
// --- End watch for dynamicTimeWindow prop --- 

// Add data aggregation function
const aggregateData = (sensorsToAggregate: Sensor[], resolution: 'hourly' | 'daily' | 'weekly' | 'monthly'): Sensor[] => {
  console.log(`[soilMoistureGraph] Aggregating data with resolution: ${resolution}`);
  
  return sensorsToAggregate.map(sensor => {
    if (!sensor.data || sensor.data.length === 0) {
      console.log(`[soilMoistureGraph] No data to aggregate for sensor ${sensor.name}`);
      return { ...sensor, data: [] };
    }

    console.log(`[soilMoistureGraph] Processing ${sensor.data.length} points for sensor ${sensor.name}`);
    const aggregatedData: DataPoint[] = [];
    const groups = new Map<string, { sum: number, count: number, time: Date }>();

    sensor.data.forEach(point => {
      const date = new Date(point.time);
      let key: string;
      let groupTime: Date;

      switch (resolution) {
        case 'hourly':
          groupTime = d3.timeHour.floor(date);
          key = groupTime.toISOString();
          break;
        case 'daily':
          groupTime = d3.timeDay.floor(date);
          key = groupTime.toISOString();
          break;
        case 'weekly':
          // Use d3's timeWeek to get the start of the week
          groupTime = d3.timeWeek.floor(date);
          key = groupTime.toISOString();
          break;
        case 'monthly':
          groupTime = new Date(date.getFullYear(), date.getMonth(), 1);
          key = `${groupTime.getFullYear()}-${groupTime.getMonth()}`;
          break;
      }

      if (!groups.has(key)) {
        groups.set(key, { sum: 0, count: 0, time: groupTime });
      }
      
      const group = groups.get(key)!;
      group.sum += point.moisture;
      group.count += 1;
    });

    // Convert groups to array and sort by time
    const sortedGroups = Array.from(groups.entries())
      .map(([_, group]) => ({
        time: group.time,
        moisture: group.sum / group.count
      }))
      .sort((a, b) => a.time.getTime() - b.time.getTime());

    console.log(`[soilMoistureGraph] Aggregated ${sensor.data.length} points into ${sortedGroups.length} points for sensor ${sensor.name}`);
    if (sortedGroups.length > 0) {
      console.log(`[soilMoistureGraph] First aggregated point:`, sortedGroups[0]);
      console.log(`[soilMoistureGraph] Last aggregated point:`, sortedGroups[sortedGroups.length - 1]);
    }

    return { ...sensor, data: sortedGroups };
  });
};

// Add data filtering function
const filterData = (params: Props['queryParams']) => {
  const now = new Date();
  let filterRangeStart: Date;
  let filterRangeEnd: Date;
  let useInclusiveEnd = false;

  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
    filterRangeEnd = now;
    useInclusiveEnd = true;

    switch (props.dynamicTimeWindow) {
      case 'lastHour':
        filterRangeStart = new Date(now.getTime() - 1 * 60 * 60 * 1000);
        break;
      case 'last24Hours':
        filterRangeStart = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        break;
      case 'last7Days':
        filterRangeStart = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case 'last30Days':
        filterRangeStart = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      default:
        filterRangeStart = new Date(params.startDate + 'T00:00:00');
        const tempEndDate = new Date(params.endDate + 'T00:00:00');
        tempEndDate.setDate(tempEndDate.getDate() + 1);
        filterRangeEnd = tempEndDate;
        useInclusiveEnd = false;
        break;
    }
  } else {
    filterRangeStart = new Date(params.startDate + 'T00:00:00');
    const tempEndDate = new Date(params.endDate + 'T00:00:00');
    tempEndDate.setDate(tempEndDate.getDate() + 1);
    filterRangeEnd = tempEndDate;
    useInclusiveEnd = false;
  }

  console.log(`[soilMoistureGraph] Filtering data from ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
  console.log(`[soilMoistureGraph] Using resolution: ${params.resolution}`);

  const minValue = params.minValue;
  const maxValue = params.maxValue;

  const filtered = sensors.value.map(sensor => {
    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time;
      const value = point.moisture;
      
      let inTimeRange;
      if (useInclusiveEnd) {
        inTimeRange = date >= filterRangeStart && date <= filterRangeEnd;
      } else {
        inTimeRange = date >= filterRangeStart && date < filterRangeEnd;
      }
      
      const isAboveMin = value >= minValue;
      const isBelowMax = value <= maxValue;
      
      return inTimeRange && isAboveMin && isBelowMax;
    });

    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  if (params.resolution !== 'raw') {
    const aggregated = aggregateData(filtered, params.resolution);
    console.log(`[soilMoistureGraph] After aggregation, total points:`, 
      aggregated.reduce((sum, s) => sum + s.data.length, 0));
    return aggregated;
  }

  return filtered;
};

// Modify updateChart to accept the processed data to draw
const updateChart = (dataToDraw: Sensor[]) => {
  console.log('[soilMoistureGraph] updateChart called with processed data. Points count:', dataToDraw.reduce((sum, s) => sum + s.data.length, 0));
  console.log('[soilMoistureGraph] Current resolution:', props.queryParams.resolution);
  // Update the reactive ref that createChart uses with the fully processed data
  sensors.value = dataToDraw; 
  createChart(); // Draw the chart with the processed data
};

// Update the getFilteredData function to return only currently visible data - USE PROP
const getFilteredData = () => {
  const filteredData = filterData(props.queryParams);
  // Filter based on the prop
  return filteredData.filter(sensor => props.sensorVisibility[sensor.name]); 
};

// Add watch for dataType changes
watch(() => props.dataType, () => {
  console.log(`[soilMoistureGraph] Data type changed to: ${props.dataType}. Fetching new data...`);
  fetchAllSensorData();
});

defineExpose({
  getFilteredData,
  fetchAllSensorData, // Expose the data fetching function
  processAndDrawChart // Expose the processing function
});

onMounted(async () => {
  await fetchAllSensorData();
  // createChart(); // createChart is now called within fetchAllSensorData -> processAndDrawChart -> updateChart
});
</script>

<style scoped>
.soil-moisture-graph {
  width: 100%;
  max-width: 928px;
  margin: 0 auto;
  padding: 20px;
}

.tooltip {
  pointer-events: none;
}
</style>