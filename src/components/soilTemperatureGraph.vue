<template>
  <div class="soil-temperature-graph">
    <div class="graph-header">
      <button @click="toggleFullscreen" class="fullscreen-btn" title="Enlarge graph">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
        </svg>
      </button>
    </div>
    <div ref="chartContainer"></div>
    <!-- Vue tooltip only for drone indicators -->
    <div 
      v-if="activeTooltip" 
      class="tooltip-container"
      :style="tooltipStyle"
    >
      <div class="tooltip-content">
        <div class="tooltip-date">{{ tooltipData.date }}</div>
        <div class="tooltip-value">{{ tooltipData.value }}</div>
      </div>
    </div>
    
    <!-- Fullscreen modal -->
    <div v-if="isFullscreen" class="fullscreen-modal" @click.self="toggleFullscreen">
      <div class="fullscreen-content">
        <button @click="toggleFullscreen" class="close-fullscreen-btn" title="Close">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
        <div ref="fullscreenChartContainer" class="fullscreen-chart"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, defineExpose, computed, nextTick } from 'vue';
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

interface ClosestSensorData {
  name: string;
  value: number;
  time: Date;
  color: string;
  distance: number;
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
  sensorConfigs: Array<{ id: string; name: string }>;
  droneImages?: Array<{ date: string; count: number }>;
}

const props = defineProps<Props>();

const chartContainer = ref<HTMLElement | null>(null);
const fullscreenChartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const fullscreenSvg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);
const isFullscreen = ref(false);

// Add new refs for tooltip
const activeTooltip = ref(false);
const tooltipData = ref({ date: '', value: '' });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});

// Update colors to be more distinguishable
const colors = [
  '#4CAF50',  // Green (B3)
  '#2196F3',  // Blue (92)
  '#FF5722',  // Deep Orange (87)
  '#9C27B0',  // Purple (9D)
  '#FFC107',  // Amber (B9)
  '#00BCD4',  // Cyan (C6)
];

// Initialize sensors ref without internal visibility
const sensors = ref<Sensor[]>([]);

const nexusStore = useNexusStore();

// Add a flag to prevent concurrent fetches
let isFetching = false;
let fetchTimeout: ReturnType<typeof setTimeout> | null = null;

async function fetchAllSensorData() {
  // Debounce rapid calls
  if (fetchTimeout) {
    clearTimeout(fetchTimeout);
  }
  
  fetchTimeout = setTimeout(async () => {
    // Prevent concurrent fetches
    if (isFetching) {
      console.log('[TEMP_GRAPH] Fetch already in progress, skipping...');
      return;
    }
    
    isFetching = true;
    let fetchError = false;
    console.log(`[TEMP_GRAPH] fetchAllSensorData called. DataType: ${props.dataType}, Resolution: ${props.queryParams.resolution}`);
    
    try {
      // Fetch data for ALL sensors, not just visible ones
      // This ensures data is available when sensors become visible
      const allSensorNames = props.sensorConfigs.map(config => config.name);
      
      // Clear existing data
      sensors.value = allSensorNames.map(name => ({
        data: [],
        name: name
      }));

      const allSensorsDataPromises = sensors.value.map(async (sensor) => {
        let rawDataPoints;
        // Get sensor ID from parent component's sensorConfigs prop
        const sensorConfig = props.sensorConfigs.find(config => config.name === sensor.name);
        
        if (!sensorConfig) {
          console.warn(`[TEMP_GRAPH] No configuration found for sensor ${sensor.name}`);
          return;
        }

        // Ensure we only fetch temperature data here
        if (props.dataType === 'temperature') { 
          console.log(`[TEMP_GRAPH] Fetching temperature for sensor ${sensor.name} (ID: ${sensorConfig.id})`);
          try {
            rawDataPoints = await nexusStore.user.getSensorTemperatureData(sensorConfig.id);
            console.log(`[TEMP_GRAPH] Raw temperature data received for ${sensor.name}:`, rawDataPoints);
          } catch (error) {
            console.error(`[TEMP_GRAPH] Error fetching temperature data for sensor ${sensor.name}:`, error);
            return;
          }
        
          if (!rawDataPoints || !Array.isArray(rawDataPoints)) {
            console.error(`[TEMP_GRAPH] Invalid data received for sensor ${sensor.name}:`, rawDataPoints);
            return;
          }
        
          // Format temperature data
          const formattedTemperatureData: DataPoint[] = rawDataPoints
            .filter((point: any) => {
              const isValid = point && 
                            typeof point.date === 'string' && 
                            typeof point.soil_temperature === 'number' && 
                            !isNaN(point.soil_temperature);
              if (!isValid) {
                console.warn(`[TEMP_GRAPH] Invalid data point for ${sensor.name}:`, point);
              }
              return isValid;
            })
            .map((point: any) => ({
              time: new Date(point.date),
              temperature: Number(point.soil_temperature), 
            }))
            .sort((a, b) => a.time.getTime() - b.time.getTime());
          
          console.log(`[TEMP_GRAPH] Formatted ${formattedTemperatureData.length} valid points for sensor ${sensor.name}`);
          if (formattedTemperatureData.length > 0) {
            console.log(`[TEMP_GRAPH] First point:`, formattedTemperatureData[0]);
            console.log(`[TEMP_GRAPH] Last point:`, formattedTemperatureData[formattedTemperatureData.length - 1]);
          }

          sensor.data = formattedTemperatureData;
        } else {
          console.warn(`[TEMP_GRAPH] Incorrect dataType prop received: ${props.dataType}. Expected 'temperature'.`);
        }
      });
      
      await Promise.all(allSensorsDataPromises);
      
      // Log final state of sensors data
      sensors.value.forEach((sensor) => {
        console.log(`[TEMP_GRAPH] Final data state for sensor ${sensor.name}: ${sensor.data.length} points`);
      });
      
    } catch (error) {
      console.error("[TEMP_GRAPH] Error fetching sensor data:", error);
      fetchError = true;
      sensors.value.forEach(sensor => sensor.data = []);
    } finally {
      isFetching = false;
      console.log(`[TEMP_GRAPH] Fetch complete. fetchError: ${fetchError}`);
      if (!fetchError) {
        processAndDrawChart();
      } else {
        console.log("[TEMP_GRAPH] Calling createChart after fetch error to draw empty state.");
        createChart(); 
      }
    }
  }, 100); // 100ms debounce
}

const processAndDrawChart = () => {
  console.log(`[TEMP_GRAPH] processAndDrawChart called. QueryParams:`, JSON.parse(JSON.stringify(props.queryParams)));
  
  // Log initial state of sensors data
  sensors.value.forEach((sensor, index) => {
    console.log(`[TEMP_GRAPH] Initial data for sensor ${sensor.name}: ${sensor.data.length} points`);
    if (sensor.data.length > 0) {
      console.log(`[TEMP_GRAPH] First point:`, sensor.data[0]);
      console.log(`[TEMP_GRAPH] Last point:`, sensor.data[sensor.data.length - 1]);
    }
  });
  
  const processedData = filterData(props.queryParams);
  const totalPoints = processedData.reduce((sum, s) => sum + s.data.length, 0);
  console.log(`[TEMP_GRAPH] After filter/aggregation, total points to draw: ${totalPoints}`);
  
  if (totalPoints > 0) {
      console.log('[TEMP_GRAPH] First sensor data after processing:', processedData[0]?.data[0]);
    console.log('[TEMP_GRAPH] Last sensor data after processing:', processedData[0]?.data[processedData[0].data.length - 1]);
  } else {
    console.warn('[TEMP_GRAPH] No data points to draw after processing');
  }
  
  updateChart(processedData);
};

const formatDate = (date: Date) => {
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  }).replace(' ', ', ');
};

// Add conversion function
const celsiusToFahrenheit = (celsius: number): number => {
  return (celsius * 9/5) + 32;
};

// Format functions
function formatValue(value: number) {
  if (value === null) return 'Loading...';
  return `${celsiusToFahrenheit(value).toFixed(1)}°F`;
}

// Update the createChart function to handle empty data better
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

  // Create SVG container with extra width for legend
  svg.value = d3.create("svg")
    .attr("viewBox", [0, 0, width + 120, height]) // Add 120px for legend
    .attr("width", width + 120)
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

  if (allData.length === 0) {
    console.warn("[TEMP_GRAPH] No data points available for chart creation");
    return;
  }

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

  // Update x-axis formatting based on resolution and time window
  if (props.dynamicTimeWindow === 'last24Hours') {
    if (props.queryParams.resolution === 'weekly') {
      // For weekly view of last 24 hours, show day names
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%a"); // Shows day names (Mon, Tue, etc.)
    } else if (props.queryParams.resolution === 'monthly') {
      // For monthly view of last 24 hours, show dates
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15" format
    } else {
      // For hourly/raw data in last 24 hours
      xAxis.ticks(d3.timeHour.every(3));
      tickFormat = d3.timeFormat("%I %p"); // Shows "2 PM" format
    }
  } else if (props.dynamicTimeWindow === 'last7Days') {
    if (props.queryParams.resolution === 'hourly') {
      // For hourly data in last 7 days, show fewer ticks
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%a %I %p"); // Shows "Mon 2 PM" format
    } else if (props.queryParams.resolution === 'weekly') {
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%a"); // Shows day names
    } else if (props.queryParams.resolution === 'monthly') {
      // For monthly data in last 7 days, show one tick per month
      xAxis.ticks(d3.timeMonth.every(1));
      tickFormat = d3.timeFormat("%b %Y"); // Shows "Jan 2024" format
    } else {
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%a"); // Shows day names
    }
  } else if (props.dynamicTimeWindow === 'last30Days') {
    if (props.queryParams.resolution === 'hourly') {
      // For hourly data in last 30 days, show one tick per day
      xAxis.ticks(d3.timeDay.every(2)); // Show every other day to reduce clutter
      tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15" format
    } else if (props.queryParams.resolution === 'weekly') {
      xAxis.ticks(d3.timeWeek.every(1));
      tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15" format
    } else if (props.queryParams.resolution === 'monthly') {
      xAxis.ticks(d3.timeMonth.every(1));
      tickFormat = d3.timeFormat("%b %Y"); // Shows "Jan 2024" format
    } else {
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15" format
    }
  } else {
    switch (props.queryParams.resolution) {
      case 'monthly':
        xAxis.ticks(d3.timeMonth.every(1));
        tickFormat = d3.timeFormat("%b %Y"); // Shows "Jan 2024"
        break;
      case 'weekly':
        xAxis.ticks(d3.timeWeek.every(1));
        tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15"
        break;
      case 'daily':
        xAxis.ticks(d3.timeDay.every(1));
        tickFormat = d3.timeFormat("%b %d"); // Shows "Jan 15"
        break;
      case 'hourly':
        xAxis.ticks(d3.timeDay.every(1));
        tickFormat = d3.timeFormat("%a %I %p"); // Shows "Mon 2 PM"
        break;
      default: // 'raw'
        xAxis.ticks(width / 80);
        tickFormat = isWithin24Hours ? 
          d3.timeFormat("%I:%M %p") : // Shows "2:30 PM"
          d3.timeFormat("%b %d"); // Shows "Jan 15"
    }
  }
  xAxis.tickFormat(tickFormat as any);
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

    // Get the correct color index based on the sensor's position in SENSOR_CONFIGS
    const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
    const color = colors[sensorColorIndex % colors.length];

    const path = svg.value!.append("path")
      .datum(sensor.data)
      .attr("fill", "none")
      .attr("stroke", color)
      .attr("stroke-width", 1.5)
      .attr("d", line);

    // Add points for each data point
    const points = svg.value!.append("g")
      .attr("class", "points")
      .selectAll("circle")
      .data(sensor.data)
      .join("circle")
      .attr("cx", d => x(d.time))
      .attr("cy", d => y(celsiusToFahrenheit(d.temperature)))
      .attr("r", 3)
      .attr("fill", color)
      .attr("stroke", "white")
      .attr("stroke-width", 1)
      .style("cursor", "pointer")
      .style("pointer-events", "all");

    // Add hover effects for points - just visual feedback, no Vue tooltip
    points.on("mouseenter", function(event: MouseEvent, d: DataPoint) {
      const point = d3.select(this);
      point.attr("r", 5).attr("stroke-width", 2);
      // Don't show Vue tooltip - let SVG tooltip handle it
    })
    .on("mouseleave", function() {
      const point = d3.select(this);
      point.attr("r", 3).attr("stroke-width", 1);
    });

    // Add hover effect for the line
    path.on("mouseover", function(this: SVGPathElement) {
      d3.select(this)
        .attr("stroke-width", 2.5);
    })
    .on("mouseout", function(this: SVGPathElement) {
      d3.select(this)
        .attr("stroke-width", 1.5);
    });
  });

  // Add legend outside the graph area
  const legend = svg.value.append("g")
    .attr("font-family", "sans-serif")
    .attr("font-size", 10)
    .attr("text-anchor", "start")
    .selectAll("g")
    .data(sensors.value.filter(s => props.sensorVisibility[s.name]))
    .join("g")
    .attr("transform", (d: Sensor, i: number) => `translate(${width + 10},${marginTop + (i * 25)})`);

  // Add colored rectangles
  legend.append("rect")
    .attr("width", 15)
    .attr("height", 15)
    .attr("fill", (d: Sensor) => {
      const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === d.name);
      return colors[sensorColorIndex % colors.length];
    })
    .attr("rx", 2); // Slightly rounded corners

  // Add sensor names next to rectangles
  legend.append("text")
    .attr("x", 25) // Increased spacing between rectangle and text
    .attr("y", 12)
    .style("font-size", "12px")
    .style("fill", "black") // Changed to black
    .text((d: Sensor) => d.name.replace(/^Sensor\s+/i, ""));

  // Move filter info below legend
  const filterInfo = svg.value.append("g")
    .attr("transform", `translate(${width + 10}, ${marginTop + (sensors.value.filter(s => props.sensorVisibility[s.name]).length * 25) + 20})`);

  // Add time window info
  filterInfo.append("text")
    .attr("x", 0)
    .attr("y", 0)
    .attr("fill", "#666")
    .text("Time Window:")
    .append("tspan")
    .attr("x", 0)
    .attr("y", 15)
    .attr("fill", "#333")
    .attr("font-weight", "bold")
    .text(props.dynamicTimeWindow === 'none' ? 'Custom Range' : 
          props.dynamicTimeWindow === 'lastHour' ? 'Last Hour' :
          props.dynamicTimeWindow === 'last24Hours' ? 'Last 24 Hours' :
          props.dynamicTimeWindow === 'last7Days' ? 'Last 7 Days' :
          props.dynamicTimeWindow === 'last30Days' ? 'Last 30 Days' : 'Custom Range');

  // Add resolution info
  filterInfo.append("text")
    .attr("x", 0)
    .attr("y", 35)
    .attr("fill", "#666")
    .text("Resolution:")
    .append("tspan")
    .attr("x", 0)
    .attr("y", 50)
    .attr("fill", "#333")
    .attr("font-weight", "bold")
    .text(props.queryParams.resolution === 'raw' ? 'Raw Data' :
          props.queryParams.resolution === 'hourly' ? 'Hourly Average' :
          props.queryParams.resolution === 'daily' ? 'Daily Average' :
          props.queryParams.resolution === 'weekly' ? 'Weekly Average' :
          'Monthly Average');

  // Add drone image indicators if we have them
  if (props.droneImages && props.droneImages.length > 0) {
    console.log('[soilTemperatureGraph] Adding drone image indicators');
    
    // Create a group for drone indicators
    const droneGroup = svg.value.append("g")
      .attr("class", "drone-indicators");

    props.droneImages.forEach(imageData => {
      const imageDate = new Date(imageData.date);
      const xPos = x(imageDate);
      
      // Only draw if the date falls within our visible range
      if (xPos >= marginLeft && xPos <= width - marginRight) {
        // Add drone icon/indicator
        const indicator = droneGroup.append("g")
          .attr("transform", `translate(${xPos}, ${marginTop})`)
          .style("cursor", "pointer");

        // Add drone icon (using a simple triangle for now)
        indicator.append("path")
          .attr("d", "M-6,-6 L6,-6 L0,6 Z")
          .attr("fill", "#FF5722")
          .attr("stroke", "white")
          .attr("stroke-width", 1);

        // Add count badge if more than 1 image
        if (imageData.count > 1) {
          indicator.append("circle")
            .attr("cx", 6)
            .attr("cy", -6)
            .attr("r", 6)
            .attr("fill", "#FF5722");

          indicator.append("text")
            .attr("x", 6)
            .attr("y", -4)
            .attr("text-anchor", "middle")
            .attr("fill", "white")
            .attr("font-size", "8px")
            .text(imageData.count);
        }

        // Add hover effect
        indicator
          .on("mouseenter", function(event: MouseEvent) {
            // Stop event propagation to prevent SVG tooltip from showing
            event.stopPropagation();
            
            d3.select(this)
              .transition()
              .duration(200)
              .attr("transform", `translate(${xPos}, ${marginTop}) scale(1.2)`);

            // Hide D3 SVG tooltip when Vue tooltip is active
            if (tooltip.value) {
              tooltip.value.style("display", "none");
            }

            // Update tooltip
            tooltipData.value = {
              date: formatDate(imageDate),
              value: `${imageData.count} drone image${imageData.count > 1 ? 's' : ''}`
            };

            const rect = chartContainer.value!.getBoundingClientRect();
            tooltipStyle.value = {
              left: `${event.clientX - rect.left + 10}px`,
              top: `${event.clientY - rect.top - 10}px`
            };

            activeTooltip.value = true;
          })
          .on("mouseleave", function() {
            d3.select(this)
              .transition()
              .duration(200)
              .attr("transform", `translate(${xPos}, ${marginTop}) scale(1)`);

            activeTooltip.value = false;
          })
          .on("pointermove", function(event: PointerEvent) {
            // Stop event propagation to prevent SVG tooltip from showing when moving over drone indicators
            event.stopPropagation();
          })
          .on("click", () => {
            // Emit event to parent to show drone images for this date
            emit('showDroneImages', imageData.date);
          });
      }
    });
  }

  // Add the chart to the container
  chartContainer.value.appendChild(svg.value!.node()!);

  // Add mouse interaction for tooltip
  const bisect = d3.bisector<DataPoint, Date>(d => d.time).center;
  
  svg.value!.on("pointermove", (event) => {
    // Hide Vue tooltip when D3 SVG tooltip is active
    if (activeTooltip.value) {
      activeTooltip.value = false;
    }

    const visibleSensorsWithData = sensors.value.filter(s => props.sensorVisibility[s.name] && s.data.length > 0);
    if (!tooltip.value || visibleSensorsWithData.length === 0) {
      // Keep tooltip visible if it was already shown
      return;
    }

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    const yPos = pointer[1];
    
    // Find the closest sensor line to the mouse pointer
    let closestSensorData: ClosestSensorData | null = null;
    let minDistance = Infinity;
    
    for (const sensor of visibleSensorsWithData) {
        const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
        const color = colors[sensorColorIndex % colors.length];

        const index = bisect(sensor.data, xPos);
        const dataPoint = sensor.data[Math.max(0, Math.min(index, sensor.data.length - 1))];
      if (!dataPoint) continue;

      // Calculate distance from mouse to this sensor's data point
      const pointX = x(dataPoint.time);
      const pointY = y(celsiusToFahrenheit(dataPoint.temperature));
      const distance = Math.sqrt(Math.pow(pointer[0] - pointX, 2) + Math.pow(yPos - pointY, 2));
      
      // Track the closest sensor
      if (distance < minDistance) {
        minDistance = distance;
        closestSensorData = {
          name: sensor.name,
          value: dataPoint.temperature,
          time: dataPoint.time,
          color: color,
          distance: distance
        };
      }
    }

    // Always show tooltip when mouse is over the graph, even if no closest sensor found
    // (this keeps the last tooltip visible until cursor moves away)
    if (!closestSensorData) {
      // Keep tooltip visible but don't update it
      return;
    }

    // Show tooltip with only the closest sensor
    tooltip.value.style("display", null);
    tooltip.value.attr("transform", `translate(${pointer[0]},${pointer[1]})`);

    // Clear previous content
    tooltip.value.selectAll("*").remove();

    // Create tooltip content group
    const tooltipGroup = tooltip.value.append("g");

    // Use closestSensorData directly (already null-checked above)
    const data = closestSensorData;

    tooltipGroup.append("rect")
      .attr("x", -70)
      .attr("y", -25)
      .attr("width", 140)
      .attr("height", 45)
      .attr("fill", "white")
      .attr("stroke", data.color)
      .attr("stroke-width", 2)
      .attr("rx", 4);

    // Add colored circle icon
    tooltipGroup.append("circle")
      .attr("cx", -60)
      .attr("cy", -5)
      .attr("r", 6)
      .attr("fill", data.color);

    // Add sensor name and value
    tooltipGroup.append("text")
      .attr("x", -50)
      .attr("y", -2)
      .attr("fill", "black")
      .attr("font-size", "12px")
      .attr("font-weight", "bold")
      .text(`${data.name}: ${formatValue(data.value)}`);

    // Add date
    tooltipGroup.append("text")
      .attr("x", -60)
      .attr("y", 12)
      .attr("fill", "#666")
      .attr("font-size", "10px")
      .text(formatDate(data.time));
  })
  .on("pointerleave", () => {
    // Only hide tooltip when cursor actually leaves the graph area
    if (tooltip.value) {
      tooltip.value.style("display", "none");
    }
  });
};

// Consolidated watcher for queryParams and dynamicTimeWindow to avoid race conditions
watch([() => props.queryParams, () => props.dynamicTimeWindow], ([newParams, newWindow], [oldParams, oldWindow]) => {
  console.log(`[TEMP_GRAPH] Query params or time window changed. Resolution: ${newParams.resolution}, Window: ${newWindow}`);
  
  // Determine if we need to re-fetch data
  const dateRangeChanged = oldParams && (newParams.startDate !== oldParams.startDate || newParams.endDate !== oldParams.endDate);
  const timeWindowChanged = newWindow !== oldWindow;
  const resolutionChanged = oldParams && newParams.resolution !== oldParams.resolution;
  
  // Re-fetch if date range or time window changed, or if resolution changed significantly
  if (dateRangeChanged || timeWindowChanged || (resolutionChanged && (newParams.resolution === 'raw' || oldParams.resolution === 'raw'))) {
    console.log(`[TEMP_GRAPH] Significant change detected, re-fetching all sensor data...`);
    fetchAllSensorData();
  } else if (resolutionChanged) {
    // Just re-process existing data if only resolution changed (and it's not raw)
    console.log(`[TEMP_GRAPH] Resolution changed, re-processing existing data...`);
    processAndDrawChart();
  } else {
    // Other param changes, just re-process
    processAndDrawChart();
  }
}, { deep: true });

// Add watch for sensorVisibility prop
watch(() => props.sensorVisibility, () => {
  console.log('[TEMP_GRAPH] sensorVisibility prop changed. Triggering createChart...');
  createChart();
}, { deep: true });

// Add data aggregation function
const aggregateData = (sensorsToAggregate: Sensor[], resolution: 'hourly' | 'daily' | 'weekly' | 'monthly'): Sensor[] => {
  console.log(`[TEMP_GRAPH] Aggregating data with resolution: ${resolution}`);
  
  return sensorsToAggregate.map(sensor => {
    if (!sensor.data || sensor.data.length === 0) {
      console.log(`[TEMP_GRAPH] No data to aggregate for sensor ${sensor.name}`);
      return { ...sensor, data: [] };
    }

    console.log(`[TEMP_GRAPH] Processing ${sensor.data.length} points for sensor ${sensor.name}`);
    const groups = new Map<string, { sum: number, count: number, time: Date }>();

    sensor.data.forEach(point => {
      const date = new Date(point.time);
      let groupTime: Date;
      let key: string;

      switch (resolution) {
        case 'hourly':
          // For hourly, use the exact hour
          groupTime = new Date(date);
          groupTime.setMinutes(0, 0, 0);
          key = groupTime.toISOString();
          break;
        case 'daily':
          groupTime = d3.timeDay.floor(date);
          key = groupTime.toISOString();
          break;
        case 'weekly':
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
      group.sum += point.temperature;
      group.count += 1;
    });

    // Convert groups to array and sort by time
    const sortedGroups = Array.from(groups.entries())
      .map(([_, group]) => ({
        time: group.time,
        temperature: group.sum / group.count
      }))
      .sort((a, b) => a.time.getTime() - b.time.getTime());

    console.log(`[TEMP_GRAPH] Aggregated ${sensor.data.length} points into ${sortedGroups.length} points for sensor ${sensor.name}`);
    if (sortedGroups.length > 0) {
      console.log(`[TEMP_GRAPH] First aggregated point: ${sortedGroups[0].time.toISOString()} = ${sortedGroups[0].temperature}`);
      console.log(`[TEMP_GRAPH] Last aggregated point: ${sortedGroups[sortedGroups.length - 1].time.toISOString()} = ${sortedGroups[sortedGroups.length - 1].temperature}`);
      // Log missing days for daily aggregation to help debug
      if (resolution === 'daily' && sortedGroups.length > 0) {
        const firstDay = d3.timeDay.floor(sortedGroups[0].time);
        const lastDay = d3.timeDay.floor(sortedGroups[sortedGroups.length - 1].time);
        const expectedDays = Math.ceil((lastDay.getTime() - firstDay.getTime()) / (24 * 60 * 60 * 1000)) + 1;
        if (sortedGroups.length < expectedDays) {
          console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: Expected ${expectedDays} days but only have ${sortedGroups.length} days with data (missing ${expectedDays - sortedGroups.length} days)`);
        }
      }
    }

    return { ...sensor, data: sortedGroups };
  });
};

// Update the filterData function to handle time windows better
const filterData = (params: Props['queryParams']) => {
  const now = new Date();
  let filterRangeStart: Date;
  let filterRangeEnd: Date;
  let useInclusiveEnd = false;

  console.log(`[TEMP_GRAPH] filterData called. Dynamic Window: ${props.dynamicTimeWindow}, Resolution: ${params.resolution}`);

  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
    filterRangeEnd = now;
    useInclusiveEnd = true;
    console.log(`[TEMP_GRAPH] Applying dynamic time window: ${props.dynamicTimeWindow}. Current Time (browser): ${now.toISOString()}`);

    switch (props.dynamicTimeWindow) {
      case 'lastHour':
        filterRangeStart = new Date(now.getTime() - 65 * 60 * 1000); // Add 5-minute buffer
        break;
      case 'last24Hours':
        // If using weekly/monthly resolution, extend the range
        if (params.resolution === 'weekly') {
          filterRangeStart = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        } else if (params.resolution === 'monthly') {
          filterRangeStart = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        } else {
          filterRangeStart = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        }
        break;
      case 'last7Days':
        const start7Days = new Date(now);
        // If using monthly resolution, extend the range to 3 months
        if (params.resolution === 'monthly') {
          start7Days.setMonth(now.getMonth() - 2);
          start7Days.setDate(1); // Start from the beginning of the month
        } else {
          // Go back 7 days (not 6) to include today + 6 previous days = 7 days total
          start7Days.setTime(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        }
        start7Days.setHours(0, 0, 0, 0);
        filterRangeStart = start7Days;
        break;
      case 'last30Days':
        const start30Days = new Date(now);
        // If using monthly resolution, extend the range to 3 months
        if (params.resolution === 'monthly') {
          start30Days.setMonth(now.getMonth() - 2);
          start30Days.setDate(1); // Start from the beginning of the month
        } else {
          // Go back 30 days (not 29) to include today + 29 previous days = 30 days total
          start30Days.setTime(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        }
        start30Days.setHours(0, 0, 0, 0);
        filterRangeStart = start30Days;
        break;
      default:
        filterRangeStart = new Date(params.startDate + 'T00:00:00');
        const tempEndDateDef = new Date(params.endDate + 'T00:00:00');
        tempEndDateDef.setDate(tempEndDateDef.getDate() + 1);
        filterRangeEnd = tempEndDateDef;
        useInclusiveEnd = false;
        break;
    }
  } else {
    filterRangeStart = new Date(params.startDate + 'T00:00:00');
    const tempEndDate = new Date(params.endDate + 'T00:00:00');
    tempEndDate.setDate(tempEndDate.getDate() + 1);
    filterRangeEnd = tempEndDate;
    useInclusiveEnd = false;
    console.log(`[TEMP_GRAPH] Using manual date range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
  }

  const minTemperature = params.minValue;
  const maxTemperature = params.maxValue;

  const filtered = sensors.value.map(sensor => {
    const initialPoints = sensor.data.length;
    console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: Filtering ${initialPoints} points...`);

    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time; // This is already a Date object from UTC timestamp
      const temperature = point.temperature;
      
      // Compare using getTime() for precise UTC timestamp comparison
      const pointTime = date.getTime();
      const startTime = filterRangeStart.getTime();
      const endTime = filterRangeEnd.getTime();
      
      let inTimeRange;
      if (useInclusiveEnd) {
        inTimeRange = pointTime >= startTime && pointTime <= endTime;
      } else {
        inTimeRange = pointTime >= startTime && pointTime < endTime;
      }
      
      const isAboveMinTemp = temperature >= minTemperature;
      const isBelowMaxTemp = temperature <= maxTemperature;
      
      if (props.dynamicTimeWindow === 'lastHour') {
        console.log(`[TEMP_GRAPH] Point time: ${date.toISOString()} (${date.toLocaleString()} local), Filter: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}, inRange: ${inTimeRange}, value: ${temperature}`);
      }
      
      return inTimeRange && isAboveMinTemp && isBelowMaxTemp;
    });

    console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: ${initialPoints} points -> ${sensorFilteredData.length} points after filtering`);
    console.log(`[TEMP_GRAPH] Filter range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
    if (sensorFilteredData.length > 0) {
      console.log(`[TEMP_GRAPH] First filtered point: ${sensorFilteredData[0].time.toISOString()}, Last filtered point: ${sensorFilteredData[sensorFilteredData.length - 1].time.toISOString()}`);
    } else if (initialPoints > 0) {
      console.warn(`[TEMP_GRAPH] Sensor ${sensor.name}: Had ${initialPoints} points but none matched the filter range!`);
      if (initialPoints > 0) {
        const firstPoint = sensor.data[0].time;
        const lastPoint = sensor.data[initialPoints - 1].time;
        console.warn(`[TEMP_GRAPH] Data range: ${firstPoint.toISOString()} to ${lastPoint.toISOString()}`);
        console.warn(`[TEMP_GRAPH] Filter range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
      }
    }
    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  // For raw data, return filtered data without aggregation
  if (params.resolution === 'raw') {
    console.log(`[TEMP_GRAPH] Returning raw data without aggregation`);
    return filtered;
  }

  // Apply aggregation for other resolutions
    console.log(`[TEMP_GRAPH] Aggregating filtered data with resolution: ${params.resolution}`);
    const aggregated = aggregateData(filtered, params.resolution);
  aggregated.forEach(s => console.log(`[TEMP_GRAPH] Sensor ${s.name}: ${s.data.length} points after aggregation`));
    return aggregated;
};

const updateChart = (dataToDraw: Sensor[]) => {
  sensors.value = dataToDraw;
  createChart();
};

const getFilteredData = () => {
  const filteredData = filterData(props.queryParams);
  return filteredData.filter(sensor => props.sensorVisibility[sensor.name]);
};

// Add watch for dataType prop
watch(() => props.dataType, (newDataType) => {
  console.log(`[TEMP_GRAPH] dataType changed to: ${newDataType}`);
  if (newDataType === 'temperature') {
    console.log('[TEMP_GRAPH] Refetching data for temperature view');
    fetchAllSensorData();
  }
}, { immediate: true });

// Fullscreen functionality
const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value;
  if (isFullscreen.value) {
    // Wait for DOM to update, then create fullscreen chart
    nextTick(() => {
      createFullscreenChart();
    });
  }
};

// Create chart in fullscreen mode
const createFullscreenChart = () => {
  if (!fullscreenChartContainer.value) {
    console.error("[TEMP_GRAPH] fullscreenChartContainer ref is not available.");
    return;
  }

  // Clear previous chart
  if (fullscreenSvg.value) {
    fullscreenSvg.value.remove();
  }

  // Use larger dimensions for fullscreen
  const width = Math.min(window.innerWidth - 100, 1400);
  const height = Math.min(window.innerHeight - 150, 800);
  const marginTop = 20;
  const marginRight = 30;
  const marginBottom = 30;
  const marginLeft = 40;

  // Create SVG container
  fullscreenSvg.value = d3.create("svg")
    .attr("viewBox", [0, 0, width + 120, height])
    .attr("width", width + 120)
    .attr("height", height)
    .attr("style", "max-width: 100%; height: auto; font: 10px sans-serif;")
    .style("-webkit-tap-highlight-color", "transparent")
    .style("overflow", "visible") as d3.Selection<SVGSVGElement, unknown, null, undefined>;

  if (!fullscreenSvg.value) return;

  // Use the same chart creation logic but with fullscreen container
  // We'll reuse the createChart logic but adapt it for fullscreen
  const allData = sensors.value
    .filter(sensor => props.sensorVisibility[sensor.name])
    .flatMap(sensor => sensor.data);
    
  if (allData.length === 0) {
    console.warn("[TEMP_GRAPH] No data points available for fullscreen chart");
    return;
  }

  // Convert Celsius to Fahrenheit for display
  const fahrenheitData = allData.map(d => ({
    ...d,
    temperature: celsiusToFahrenheit(d.temperature)
  }));

  let yMinActual = d3.min(fahrenheitData, d => d.temperature) as number;
  let yMaxActual = d3.max(fahrenheitData, d => d.temperature) as number;

  if (fahrenheitData.length === 0 || yMinActual === undefined || yMaxActual === undefined) {
    yMinActual = 32;
    yMaxActual = 122;
  }

  const dataRange = yMaxActual - yMinActual;
  let yDomainMinWithPadding = Math.max(14, yMinActual - (dataRange * 0.1));
  let yDomainMaxWithPadding = Math.min(122, yMaxActual + (dataRange * 0.1));

  // Create scales
  const x = d3.scaleUtc()
    .range([marginLeft, width - marginRight]);

  let currentXDomain = d3.extent(allData, d => d.time) as [Date, Date];
  if (!currentXDomain[0] || !currentXDomain[1]) {
    const today = new Date();
    currentXDomain = [today, d3.timeDay.offset(today, 1)];
  } else if (currentXDomain[0].getTime() === currentXDomain[1].getTime()) {
    const singleDate = currentXDomain[0];
    currentXDomain = [d3.timeHour.offset(singleDate, -1), d3.timeHour.offset(singleDate, 1)];
  }
  x.domain(currentXDomain);

  const y = d3.scaleLinear()
    .domain([yDomainMinWithPadding, yDomainMaxWithPadding])
    .range([height - marginBottom, marginTop]);

  // Create line generator
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(celsiusToFahrenheit(d.temperature)));

  // Add x-axis
  const xAxisGroup = fullscreenSvg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`);

  let xAxis = d3.axisBottom(x);
  let tickFormat: (date: Date) => string;

  // Use same axis formatting logic as createChart
  const timeRange = x.domain()[1].getTime() - x.domain()[0].getTime();
  const isWithin24Hours = timeRange <= 24 * 60 * 60 * 1000;

  if (props.dynamicTimeWindow === 'last24Hours') {
    if (props.queryParams.resolution === 'weekly') {
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%a");
    } else if (props.queryParams.resolution === 'monthly') {
      xAxis.ticks(d3.timeDay.every(1));
      tickFormat = d3.timeFormat("%b %d");
    } else {
      xAxis.ticks(d3.timeHour.every(3));
      tickFormat = d3.timeFormat("%I %p");
    }
  } else {
    switch (props.queryParams.resolution) {
      case 'monthly':
        xAxis.ticks(d3.timeMonth.every(1));
        tickFormat = d3.timeFormat("%b %Y");
        break;
      case 'weekly':
        xAxis.ticks(d3.timeWeek.every(1));
        tickFormat = d3.timeFormat("%b %d");
        break;
      case 'daily':
        xAxis.ticks(d3.timeDay.every(1));
        tickFormat = d3.timeFormat("%b %d");
        break;
      case 'hourly':
        xAxis.ticks(d3.timeDay.every(1));
        tickFormat = d3.timeFormat("%a %I %p");
        break;
      default:
        xAxis.ticks(width / 80);
        tickFormat = isWithin24Hours ? 
          d3.timeFormat("%I:%M %p") : 
          d3.timeFormat("%b %d");
    }
  }
  xAxis.tickFormat(tickFormat as any);
  xAxisGroup.call(xAxis);

  // Add y-axis
  fullscreenSvg.value.append("g")
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

  // Create tooltip container for fullscreen
  const fullscreenTooltip = fullscreenSvg.value.append("g")
    .attr("class", "tooltip")
    .style("display", "none");

  // Add lines for each visible sensor
  sensors.value.forEach((sensor, i) => {
    if (!props.sensorVisibility[sensor.name] || sensor.data.length === 0) {
      return;
    }

    const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
    const color = colors[sensorColorIndex % colors.length];

    const path = fullscreenSvg.value!.append("path")
      .datum(sensor.data)
      .attr("fill", "none")
      .attr("stroke", color)
      .attr("stroke-width", 2)
      .attr("d", line);

    // Add points
    const points = fullscreenSvg.value!.append("g")
      .attr("class", "points")
      .selectAll("circle")
      .data(sensor.data)
      .join("circle")
      .attr("cx", d => x(d.time))
      .attr("cy", d => y(celsiusToFahrenheit(d.temperature)))
      .attr("r", 4)
      .attr("fill", color)
      .attr("stroke", "white")
      .attr("stroke-width", 1.5)
      .style("cursor", "pointer");

    // Add hover effects
    points.on("mouseenter", function() {
      d3.select(this).attr("r", 6).attr("stroke-width", 2);
    })
    .on("mouseleave", function() {
      d3.select(this).attr("r", 4).attr("stroke-width", 1.5);
    });
  });

  // Add legend
  const legend = fullscreenSvg.value.append("g")
    .attr("font-family", "sans-serif")
    .attr("font-size", 12)
    .attr("text-anchor", "start")
    .selectAll("g")
    .data(sensors.value.filter(s => props.sensorVisibility[s.name]))
    .join("g")
    .attr("transform", (d: Sensor, i: number) => `translate(${width + 10},${marginTop + (i * 30)})`);

  legend.append("rect")
    .attr("width", 18)
    .attr("height", 18)
    .attr("fill", (d: Sensor) => {
      const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === d.name);
      return colors[sensorColorIndex % colors.length];
    })
    .attr("rx", 2);

  legend.append("text")
    .attr("x", 28)
    .attr("y", 14)
    .style("font-size", "14px")
    .style("fill", "black")
    .text((d: Sensor) => d.name.replace(/^Sensor\s+/i, ""));

  // Add mouse interaction for tooltip (similar to main chart)
  const bisect = d3.bisector<DataPoint, Date>(d => d.time).center;
  
  fullscreenSvg.value.on("pointermove", (event) => {
    const visibleSensorsWithData = sensors.value.filter(s => props.sensorVisibility[s.name] && s.data.length > 0);
    if (!fullscreenTooltip || visibleSensorsWithData.length === 0) {
      return;
    }

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    const yPos = pointer[1];
    
    let closestSensorData: ClosestSensorData | null = null;
    let minDistance = Infinity;
    
    for (const sensor of visibleSensorsWithData) {
      const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
      const color = colors[sensorColorIndex % colors.length];

      const index = bisect(sensor.data, xPos);
      const dataPoint = sensor.data[Math.max(0, Math.min(index, sensor.data.length - 1))];
      if (!dataPoint) continue;

      const pointX = x(dataPoint.time);
      const pointY = y(celsiusToFahrenheit(dataPoint.temperature));
      const distance = Math.sqrt(Math.pow(pointer[0] - pointX, 2) + Math.pow(yPos - pointY, 2));
      
      if (distance < minDistance) {
        minDistance = distance;
        closestSensorData = {
          name: sensor.name,
          value: dataPoint.temperature,
          time: dataPoint.time,
          color: color,
          distance: distance
        };
      }
    }

    if (!closestSensorData) {
      return;
    }

    fullscreenTooltip.style("display", null);
    fullscreenTooltip.attr("transform", `translate(${pointer[0]},${pointer[1]})`);
    fullscreenTooltip.selectAll("*").remove();

    const tooltipGroup = fullscreenTooltip.append("g");
    const data = closestSensorData;

    tooltipGroup.append("rect")
      .attr("x", -70)
      .attr("y", -25)
      .attr("width", 140)
      .attr("height", 45)
      .attr("fill", "white")
      .attr("stroke", data.color)
      .attr("stroke-width", 2)
      .attr("rx", 4);

    tooltipGroup.append("circle")
      .attr("cx", -60)
      .attr("cy", -5)
      .attr("r", 6)
      .attr("fill", data.color);

    tooltipGroup.append("text")
      .attr("x", -50)
      .attr("y", -2)
      .attr("fill", "black")
      .attr("font-size", "12px")
      .attr("font-weight", "bold")
      .text(`${data.name}: ${formatValue(data.value)}`);

    tooltipGroup.append("text")
      .attr("x", -60)
      .attr("y", 12)
      .attr("fill", "#666")
      .attr("font-size", "10px")
      .text(formatDate(data.time));
  })
  .on("pointerleave", () => {
    if (fullscreenTooltip) {
      fullscreenTooltip.style("display", "none");
    }
  });

  // Add the chart to the fullscreen container
  fullscreenChartContainer.value.appendChild(fullscreenSvg.value!.node()!);
};

// Watch for fullscreen changes to recreate chart
watch(isFullscreen, (newValue) => {
  if (newValue && fullscreenChartContainer.value) {
    nextTick(() => {
      createFullscreenChart();
    });
  }
});

// Update onMounted to ensure data is fetched
onMounted(async () => {
  console.log('[TEMP_GRAPH] Component mounted, fetching initial data');
  if (props.dataType === 'temperature') {
    await fetchAllSensorData();
  }
});

defineExpose({
  getFilteredData,
  fetchAllSensorData,
  processAndDrawChart
});

// Add emits
const emit = defineEmits(['showDroneImages']);
</script>

<style scoped>
.soil-temperature-graph {
  width: 100%;
  max-width: 928px;
  margin: 0 auto;
  padding: 20px;
  position: relative;
}

.tooltip-container {
  position: absolute;
  background: rgba(255, 255, 255, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 
    0 4px 24px -1px rgba(0, 0, 0, 0.08),
    0 0 1px 0 rgba(0, 0, 0, 0.06),
    inset 0 0 0 1px rgba(255, 255, 255, 0.15);
  pointer-events: none;
  z-index: 1000;
  min-width: 120px;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

.tooltip-content {
  font-size: 12px;
  line-height: 1.5;
  color: rgba(0, 0, 0, 0.8);
}

.tooltip-date {
  color: rgba(0, 0, 0, 0.6);
  margin-bottom: 6px;
  font-weight: 500;
}

.tooltip-value {
  color: rgba(0, 0, 0, 0.9);
  font-weight: 600;
}

.points circle {
  transition: r 0.2s, stroke-width 0.2s;
}

.graph-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 10px;
}

.fullscreen-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 8px;
  cursor: pointer;
  color: #666;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fullscreen-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: #333;
  border-color: rgba(255, 255, 255, 0.3);
}

.fullscreen-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 20px;
}

.fullscreen-content {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-radius: 8px;
  padding: 60px 40px 40px;
}

.close-fullscreen-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(0, 0, 0, 0.1);
  border: none;
  border-radius: 6px;
  padding: 10px;
  cursor: pointer;
  color: #333;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2001;
}

.close-fullscreen-btn:hover {
  background: rgba(0, 0, 0, 0.2);
  color: #000;
}

.fullscreen-chart {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: auto;
}
</style> 