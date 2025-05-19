<template>
  <div class="soil-temperature-graph">
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
  temperature: number;
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
  dataType: 'temperature' | 'moisture';
}

const props = defineProps<Props>();

const chartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);

const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

// Define SENSOR_CONFIGS (could be props later if more dynamic)
const SENSOR_CONFIGS = [
  { id: 444574498032128, name: 'Sensor Alpha' },
];

const nexusStore = useNexusStore();

// Initialize sensors ref without internal visibility
const sensors = ref<Sensor[]>(SENSOR_CONFIGS.map((config) => ({
  data: [], 
  name: config.name,
})));

async function fetchAllSensorData() {
  let fetchError = false;
  sensors.value.forEach(sensor => sensor.data = []); // Clear data before fetch
  console.log(`[TEMP_GRAPH] fetchAllSensorData called. DataType: ${props.dataType}, Resolution: ${props.queryParams.resolution}`);
  try {
    const allSensorsDataPromises = SENSOR_CONFIGS.map(async (config, index) => {
      let rawDataPoints;
      // Ensure we only fetch temperature data here
      if (props.dataType === 'temperature') { 
        console.log(`[TEMP_GRAPH] Fetching temperature for sensor ${config.name} (ID: ${config.id})`);
        rawDataPoints = await nexusStore.user.getSensorTemperatureData(config.id);
        console.log(`[TEMP_GRAPH] Raw data received for ${config.name}:`, rawDataPoints);
      } else {
        console.warn(`[TEMP_GRAPH] Incorrect dataType prop received: ${props.dataType}. Expected 'temperature'.`);
        rawDataPoints = []; // Avoid errors later
      }
      
      if (rawDataPoints && Array.isArray(rawDataPoints) && rawDataPoints.length > 0) {
        // Renamed variable for clarity
        const formattedTemperatureData: DataPoint[] = rawDataPoints.map((point: any) => ({
          time: new Date(point.date),
          // Ensure we are using the correct field 'soil_temperature'
          temperature: Number(point.soil_temperature), 
        })).sort((a, b) => a.time.getTime() - b.time.getTime());
        
        // Check for invalid data points (NaN temperature)
        const validDataPoints = formattedTemperatureData.filter(p => !isNaN(p.temperature));
        if (validDataPoints.length !== formattedTemperatureData.length) {
            console.warn(`[TEMP_GRAPH] Filtered out ${formattedTemperatureData.length - validDataPoints.length} invalid data points for ${config.name}`);
        }

        sensors.value[index].data = validDataPoints;
        console.log(`[TEMP_GRAPH] Formatted ${validDataPoints.length} valid points for sensor ${config.name}. First point:`, validDataPoints[0]);
      } else {
        console.warn(`[TEMP_GRAPH] No valid array data received for sensor ${config.name}`);
        sensors.value[index].data = [];
      }
    });
    await Promise.all(allSensorsDataPromises);
  } catch (error) {
    console.error("[TEMP_GRAPH] Error fetching sensor data:", error);
    fetchError = true;
    sensors.value.forEach(sensor => sensor.data = []);
  } finally {
    console.log(`[TEMP_GRAPH] Fetch complete. fetchError: ${fetchError}`);
    if (!fetchError) {
      processAndDrawChart();
    } else {
      // Still call createChart to render empty axes if fetch failed
      console.log("[TEMP_GRAPH] Calling createChart after fetch error to draw empty state.");
      createChart(); 
    }
  }
}

const processAndDrawChart = () => {
  console.log(`[TEMP_GRAPH] processAndDrawChart called. QueryParams:`, JSON.parse(JSON.stringify(props.queryParams)));
  const processedData = filterData(props.queryParams);
  const totalPoints = processedData.reduce((sum, s) => sum + s.data.length, 0);
  console.log(`[TEMP_GRAPH] After filter/aggregation, total points to draw: ${totalPoints}`);
  if (totalPoints > 0) {
      console.log('[TEMP_GRAPH] First sensor data after processing:', processedData[0]?.data[0]);
  }
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

// Add conversion functions
const celsiusToFahrenheit = (celsius: number): number => {
  return (celsius * 9/5) + 32;
};

const formatValue = (value: number) => {
  const fahrenheit = celsiusToFahrenheit(value);
  return `${fahrenheit.toFixed(1)}°F`;
};

const createChart = () => {
  if (!chartContainer.value) {
      console.error("[TEMP_GRAPH] chartContainer ref is not available.");
      return;
  }

  console.log("[TEMP_GRAPH] createChart called.");
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
    .style("overflow", "visible") as d3.Selection<SVGSVGElement, unknown, null, undefined>;

  if (!svg.value) return;

  // Combine visible sensor data for domain calculation
  const allData = sensors.value
    .filter(sensor => props.sensorVisibility[sensor.name])
    .flatMap(sensor => sensor.data);
    
  console.log(`[TEMP_GRAPH] createChart: Processing ${allData.length} total visible data points for domains.`);

  // Convert Celsius to Fahrenheit for display
  const fahrenheitData = allData.map(d => ({
    ...d,
    temperature: celsiusToFahrenheit(d.temperature)
  }));

  let yMinActual = d3.min(fahrenheitData, d => d.temperature) as number;
  let yMaxActual = d3.max(fahrenheitData, d => d.temperature) as number;

  // Handle cases where there's no data or min/max are undefined
  if (fahrenheitData.length === 0 || yMinActual === undefined || yMaxActual === undefined) {
    yMinActual = 32; // 0°C in Fahrenheit
    yMaxActual = 122; // 50°C in Fahrenheit
  }

  const dataRange = yMaxActual - yMinActual;
  let yDomainMinWithPadding = Math.max(14, yMinActual - (dataRange * 0.1)); // 14°F is -10°C
  let yDomainMaxWithPadding = Math.min(122, yMaxActual + (dataRange * 0.1)); // 122°F is 50°C

  // Create scales
  const x = d3.scaleUtc()
    .range([marginLeft, width - marginRight]);

  // Adjust domain based on aggregated data and handle single point case
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

  // Create line generator with Fahrenheit conversion
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(celsiusToFahrenheit(d.temperature)));

  // Determine if we should show time based on data range and resolution
  const timeRange = x.domain()[1].getTime() - x.domain()[0].getTime();
  const isWithin24Hours = timeRange <= 24 * 60 * 60 * 1000;

  // Add x-axis
  const xAxisGroup = svg.value!.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`);

  let xAxis = d3.axisBottom(x);
  let tickFormat: (date: Date) => string;

  // --- Specific handling for last24Hours --- 
  if (props.dynamicTimeWindow === 'last24Hours') {
    xAxis.ticks(d3.timeHour.every(3)); // Ticks every 3 hours
    tickFormat = d3.timeFormat("%I %p"); // Format like "02 PM"
    xAxis.tickFormat(tickFormat as any);
    console.log("[TEMP_GRAPH] Applying specific x-axis formatting for last24Hours.");
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

  // Add y-axis with Fahrenheit values
  svg.value!.append("g")
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
      .text("↑ Soil Temperature (°F)"));

  // Create tooltip container
  tooltip.value = svg.value!.append("g")
    .attr("class", "tooltip")
    .style("display", "none");

  // Add lines for each visible sensor
  sensors.value.forEach((sensor, i) => {
    if (!props.sensorVisibility[sensor.name]) {
        console.log(`[TEMP_GRAPH] Skipping line for hidden sensor: ${sensor.name}`);
        return; 
    }
    if (sensor.data.length === 0) {
        console.log(`[TEMP_GRAPH] Skipping line for sensor with no data: ${sensor.name}`);
        return;
    }
    console.log(`[TEMP_GRAPH] Drawing line for visible sensor: ${sensor.name} with ${sensor.data.length} points.`);

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
  const legend = svg.value!.append("g")
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
  chartContainer.value.appendChild(svg.value!.node()!);

  // Add mouse interaction for tooltip
  const bisect = d3.bisector<DataPoint, Date>(d => d.time).center;
  
  svg.value!.on("pointermove", (event) => {
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
          value: dataPoint.temperature,
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
  console.log(`[TEMP_GRAPH] Query params changed. New resolution: ${newParams.resolution}`);
  processAndDrawChart();
}, { deep: true });

// Add watch for sensorVisibility prop
watch(() => props.sensorVisibility, () => {
  console.log('[TEMP_GRAPH] sensorVisibility prop changed. Triggering createChart...');
  createChart();
}, { deep: true });

// Watch for dynamicTimeWindow prop
watch(() => props.dynamicTimeWindow, () => {
  console.log(`[TEMP_GRAPH] dynamicTimeWindow prop changed to: ${props.dynamicTimeWindow}. Triggering chart processing...`);
  processAndDrawChart();
});

// Add data filtering function
const filterData = (params: Props['queryParams']) => {
  console.log(`[TEMP_GRAPH] filterData called. Date range: ${params.startDate} to ${params.endDate}. Value range: ${params.minValue}°C to ${params.maxValue}°C.`);
  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
      console.log(`[TEMP_GRAPH] Applying dynamic time window: ${props.dynamicTimeWindow}`);
  }

  const now = new Date();
  let filterRangeStart: Date;
  let filterRangeEnd: Date;
  let useInclusiveEnd = false;

  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
    filterRangeEnd = now;
    useInclusiveEnd = true;

    switch (props.dynamicTimeWindow) {
      case 'lastHour':
        // Add a 5-minute buffer to account for potential data delay
        filterRangeStart = new Date(now.getTime() - 65 * 60 * 1000);
        console.log(`[TEMP_GRAPH] >> lastHour: Calculated filter range (with buffer): ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
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
   console.log(`[TEMP_GRAPH] Filtering time from ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}. Inclusive end: ${useInclusiveEnd}`);

  const minTemperature = params.minValue;
  const maxTemperature = params.maxValue;

  const filtered = sensors.value.map(sensor => {
    const initialPoints = sensor.data.length;
    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time;
      const temperature = point.temperature; // Already in Celsius here
      
      let inTimeRange;
      if (useInclusiveEnd) {
        inTimeRange = date >= filterRangeStart && date <= filterRangeEnd;
      } else {
        inTimeRange = date >= filterRangeStart && date < filterRangeEnd;
      }
      
      const isAboveMinTemp = temperature >= minTemperature;
      const isBelowMaxTemp = temperature <= maxTemperature;
      
      return inTimeRange && isAboveMinTemp && isBelowMaxTemp;
    });
    console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: ${initialPoints} points -> ${sensorFilteredData.length} points after time/value filtering.`);
    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  if (params.resolution !== 'raw') {
    console.log(`[TEMP_GRAPH] Aggregating filtered data with resolution: ${params.resolution}`);
    const aggregated = aggregateData(filtered, params.resolution);
    aggregated.forEach(s => console.log(`[TEMP_GRAPH] Sensor ${s.name}: ${s.data.length} points after aggregation.`));
    return aggregated;
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
      const date = new Date(point.time);

      switch (resolution) {
        case 'hourly':
          key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}-${date.getHours()}`;
          break;
        case 'daily':
          key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
          break;
        case 'weekly':
          // Use d3's timeWeek to get the start of the week
          const weekStartDate = d3.timeWeek.floor(date);
          key = weekStartDate.toISOString();
          break;
        case 'monthly':
          const monthStartDate = new Date(date.getFullYear(), date.getMonth(), 1);
          key = `${monthStartDate.getFullYear()}-${monthStartDate.getMonth()}`;
          break;
      }

      if (!groups.has(key)) {
        let groupTime = d3.timeHour.floor(date);
        if (resolution === 'daily') {
          groupTime = d3.timeDay.floor(date);
        } else if (resolution === 'weekly') {
          groupTime = d3.timeWeek.floor(date);
        } else if (resolution === 'monthly') {
          groupTime = new Date(date.getFullYear(), date.getMonth(), 1);
        }
        groups.set(key, { sum: 0, count: 0, time: groupTime });
      }
      
      const group = groups.get(key)!;
      group.sum += point.temperature;
      group.count += 1;
    });

    groups.forEach(group => {
      const averageTemperature = group.sum / group.count;
      aggregatedData.push({
        time: group.time,
        temperature: averageTemperature,
      });
    });

    aggregatedData.sort((a, b) => a.time.getTime() - b.time.getTime());

    return { ...sensor, data: aggregatedData };
  });
};

const updateChart = (dataToDraw: Sensor[]) => {
  sensors.value = dataToDraw;
  createChart();
};

const getFilteredData = () => {
  const filteredData = filterData(props.queryParams);
  return filteredData.filter(sensor => props.sensorVisibility[sensor.name]);
};

defineExpose({
  getFilteredData,
  fetchAllSensorData,
  processAndDrawChart
});

onMounted(async () => {
  await fetchAllSensorData();
});
</script>

<style scoped>
.soil-temperature-graph {
  width: 100%;
  max-width: 928px;
  margin: 0 auto;
  padding: 20px;
}

.tooltip {
  pointer-events: none;
}
</style> 