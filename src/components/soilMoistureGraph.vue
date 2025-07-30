<template>
  <div class="soil-moisture-graph">
    <div ref="chartContainer"></div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, defineExpose, computed } from 'vue';
import * as d3 from 'd3';
import type { Selection } from 'd3';
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
  sensorConfigs: Array<{ id: string; name: string }>;
  droneImages?: Array<{ date: string; count: number }>; // Add this prop
}

const props = defineProps<Props>();

const chartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);

// Update colors to be more distinguishable
const colors = [
  '#4CAF50',  // Green (B3)
  '#2196F3',  // Blue (92)
  '#FF5722',  // Deep Orange (87)
  '#9C27B0',  // Purple (9D)
  '#FFC107',  // Amber (B9)
  '#00BCD4',  // Cyan (C6)
];

// Remove hardcoded SENSOR_CONFIGS and use props
const nexusStore = useNexusStore();

// Initialize sensors ref without internal visibility
const sensors = ref<Sensor[]>([]);

// Add new refs for tooltip
const activeTooltip = ref(false);
const tooltipData = ref({ date: '', value: '' });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});

async function fetchAllSensorData() {
  let fetchError = false;
  try {
    console.log(`[soilMoistureGraph] Fetching data with resolution: ${props.queryParams.resolution}`);
    
    // Get visible sensors from props
    const visibleSensors = Object.entries(props.sensorVisibility)
      .filter(([_, isVisible]) => isVisible)
      .map(([name]) => name);
    
    // Clear existing data
    sensors.value = visibleSensors.map(name => ({
      data: [],
      name: name
    }));

    const allSensorsDataPromises = sensors.value.map(async (sensor) => {
      let rawDataPoints;
      // Get sensor ID from parent component's sensorConfigs prop
      const sensorConfig = props.sensorConfigs.find(config => config.name === sensor.name);
      
      if (!sensorConfig) {
        console.warn(`[soilMoistureGraph] No configuration found for sensor ${sensor.name}`);
        return;
      }
      console.log(`[soilMoistureGraph] Requesting moisture data for sensor ID: ${sensorConfig.id}`);

      if (props.dataType === 'moisture') {
        rawDataPoints = await nexusStore.user.getSensorMoistureData(sensorConfig.id);
      } else {
        rawDataPoints = await nexusStore.user.getSensorTemperatureData(sensorConfig.id);
      }
      
      if (rawDataPoints && Array.isArray(rawDataPoints)) {
        const formattedData: DataPoint[] = rawDataPoints.map((point: any) => ({
          time: new Date(point.date),
          moisture: props.dataType === 'moisture' ? Number(point.soil_moisture) : Number(point.soil_temperature),
        })).sort((a, b) => a.time.getTime() - b.time.getTime());
        
        sensor.data = formattedData;
        console.log(`[soilMoistureGraph] Fetched ${formattedData.length} points for sensor ${sensor.name}`);
      } else {
        console.warn(`No data for sensor ${sensor.name}`);
        sensor.data = [];
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

// Add conversion function
const celsiusToFahrenheit = (celsius: number) => {
  return (celsius * 9/5) + 32;
};

const formatDate = (date: Date) => {
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });
};

// Format functions
function formatValue(value: number, type: 'moisture' | 'temperature') {
  if (value === null) return 'Loading...';
  return type === 'temperature' 
    ? `${celsiusToFahrenheit(value).toFixed(1)}°F`
    : `${value.toFixed(1)}%`;
}

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

  // Create SVG container with extra width for legend
  svg.value = d3.create("svg")
    .attr("viewBox", [0, 0, width + 120, height]) // Add 120px for legend
    .attr("width", width + 120)
    .attr("height", height)
    .attr("style", "max-width: 100%; height: auto; height: intrinsic; font: 10px sans-serif;")
    .style("-webkit-tap-highlight-color", "transparent")
    .style("overflow", "visible") as Selection<SVGSVGElement, unknown, null, undefined>;

  if (!svg.value) return;

  // Combine visible sensor data for domain calculation
  const allData = sensors.value
    .filter(sensor => props.sensorVisibility[sensor.name])
    .flatMap(sensor => sensor.data);

  if (allData.length === 0) {
    console.warn("[soilMoistureGraph] No data points available for chart creation");
    return;
  }

  let yMinActual = d3.min(allData, (d: DataPoint) => d.moisture) as number;
  let yMaxActual = d3.max(allData, (d: DataPoint) => d.moisture) as number;

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

  let currentXDomain = d3.extent(allData, (d: DataPoint) => d.time) as [Date, Date];
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
    .x((d: DataPoint) => x(d.time))
    .y((d: DataPoint) => y(d.moisture));

  // Determine if we should show time based on data range and resolution
  const timeRange = x.domain()[1].getTime() - x.domain()[0].getTime();
  const isWithin24Hours = timeRange <= 24 * 60 * 60 * 1000;

  // Add x-axis
  const xAxisGroup = svg.value.append("g")
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

  // Add y-axis
  svg.value.append("g")
    .attr("transform", `translate(${marginLeft},0)`)
    .call(d3.axisLeft(y).ticks(height / 40))
    .call((g: Selection<SVGGElement, unknown, null, undefined>) => g.select(".domain").remove())
    .call((g: Selection<SVGGElement, unknown, null, undefined>) => g.selectAll(".tick line").clone()
      .attr("x2", width - marginLeft - marginRight)
      .attr("stroke-opacity", 0.1))
    .call((g: Selection<SVGGElement, unknown, null, undefined>) => g.append("text")
      .attr("x", -marginLeft)
      .attr("y", 10)
      .attr("fill", "currentColor")
      .attr("text-anchor", "start")
      .text("↑ Soil Moisture (%)"));

  // Create tooltip container
  tooltip.value = svg.value.append("g")
    .attr("class", "tooltip")
    .style("display", "none")
    .style("pointer-events", "none"); // Prevent tooltip from interfering with mouse events

  // Add lines for each visible sensor
  sensors.value.forEach((sensor, index) => {
    if (!props.sensorVisibility[sensor.name]) return;

    // Get the correct color index based on the sensor's position in SENSOR_CONFIGS
    const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
    const color = colors[sensorColorIndex % colors.length];

    // Add the line
    const path = svg.value!.append("path")
      .attr("fill", "none")
      .attr("stroke", color)
      .attr("stroke-width", 1.5)
      .attr("d", line(sensor.data));

    // Add points
    const points = svg.value!.append("g")
      .attr("class", "points")
      .selectAll("circle")
      .data(sensor.data)
      .join("circle")
      .attr("cx", (d: DataPoint) => x(d.time))
      .attr("cy", (d: DataPoint) => y(d.moisture))
      .attr("r", 3)
      .attr("fill", color)
      .attr("stroke", "white")
      .attr("stroke-width", 1)
      .style("cursor", "pointer")
      .style("pointer-events", "all");

    // Add hover effects for points
    points.on("mouseenter", function(event: MouseEvent, d: DataPoint) {
      const point = d3.select(this);
      point.attr("r", 5).attr("stroke-width", 2);

      // Update tooltip data
      tooltipData.value = {
        date: formatDate(d.time),
        value: formatValue(d.moisture, props.dataType)
      };

      // Position tooltip
      const rect = chartContainer.value!.getBoundingClientRect();
      tooltipStyle.value = {
        left: `${event.clientX - rect.left + 10}px`,
        top: `${event.clientY - rect.top - 10}px`
      };

      activeTooltip.value = true;
    })
    .on("mouseleave", function() {
      const point = d3.select(this);
      point.attr("r", 3).attr("stroke-width", 1);
      activeTooltip.value = false;
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
    console.log('[soilMoistureGraph] Adding drone image indicators');
    
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
          .on("mouseenter", function(event) {
            d3.select(this)
              .transition()
              .duration(200)
              .attr("transform", `translate(${xPos}, ${marginTop}) scale(1.2)`);

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
          .on("click", () => {
            // Emit event to parent to show drone images for this date
            emit('showDroneImages', imageData.date);
          });
      }
    });
  }

  // Add the chart to the container
  if (chartContainer.value && svg.value.node()) {
    chartContainer.value.appendChild(svg.value.node()!);
    }
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

  console.log(`[soilMoistureGraph] filterData called. Dynamic Window: ${props.dynamicTimeWindow}, Resolution: ${params.resolution}`);

  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
    filterRangeEnd = now;
    useInclusiveEnd = true;
    console.log(`[soilMoistureGraph] Applying dynamic time window: ${props.dynamicTimeWindow}. Current Time (browser): ${now.toISOString()}`);

    switch (props.dynamicTimeWindow) {
      case 'lastHour':
        filterRangeStart = new Date(now.getTime() - 60 * 60 * 1000);
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
          start7Days.setDate(now.getDate() - 6);
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
          start30Days.setDate(now.getDate() - 29);
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
    console.log(`[soilMoistureGraph] Using manual date range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
  }

  const minValue = params.minValue;
  const maxValue = params.maxValue;

  const filtered = sensors.value.map(sensor => {
    const initialPoints = sensor.data.length;
    console.log(`[soilMoistureGraph] Sensor ${sensor.name}: Filtering ${initialPoints} points...`);

    // Log first/last points BEFORE filtering
    if (initialPoints > 0) {
      console.log(`[soilMoistureGraph] Sensor ${sensor.name}: Raw first point time: ${sensor.data[0]?.time.toISOString()}, Raw last point time: ${sensor.data[initialPoints - 1]?.time.toISOString()}`);
    }

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
      
      if (props.dynamicTimeWindow === 'lastHour') {
        console.log(`[soilMoistureGraph] Point time: ${date.toISOString()}, inRange: ${inTimeRange}, value: ${value}, aboveMin: ${isAboveMin}, belowMax: ${isBelowMax}`);
      }
      
      return inTimeRange && isAboveMin && isBelowMax;
    });

    console.log(`[soilMoistureGraph] Sensor ${sensor.name}: ${initialPoints} points -> ${sensorFilteredData.length} points after filtering`);
    if (sensorFilteredData.length > 0) {
      console.log(`[soilMoistureGraph] First filtered point: ${sensorFilteredData[0].time.toISOString()}, Last filtered point: ${sensorFilteredData[sensorFilteredData.length - 1].time.toISOString()}`);
    }

    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  // For last hour, we want to show data with the selected resolution
  if (props.dynamicTimeWindow === 'lastHour') {
    console.log(`[soilMoistureGraph] Last hour view - applying selected resolution: ${params.resolution}`);
    if (params.resolution !== 'raw') {
      return aggregateData(filtered, params.resolution);
    }
    return filtered;
  }

  // Apply aggregation for other cases
  if (params.resolution !== 'raw') {
    console.log(`[soilMoistureGraph] Aggregating data with resolution: ${params.resolution}`);
    const aggregated = aggregateData(filtered, params.resolution);
    aggregated.forEach(s => console.log(`[soilMoistureGraph] Sensor ${s.name}: ${s.data.length} points after aggregation`));
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

// Add emits
const emit = defineEmits(['showDroneImages']);
</script>

<style scoped>
.soil-moisture-graph {
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
</style>