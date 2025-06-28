<template>
  <div class="soil-temperature-graph">
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
  sensorConfigs: Array<{ id: number; name: string }>;
}

const props = defineProps<Props>();

const chartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);

// Add new refs for tooltip
const activeTooltip = ref(false);
const tooltipData = ref({ date: '', value: '' });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});

const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

// Initialize sensors ref without internal visibility
const sensors = ref<Sensor[]>([]);

const nexusStore = useNexusStore();

async function fetchAllSensorData() {
  let fetchError = false;
  console.log(`[TEMP_GRAPH] fetchAllSensorData called. DataType: ${props.dataType}, Resolution: ${props.queryParams.resolution}`);
  
  try {
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
        
        if (rawDataPoints.length === 0) {
          console.warn(`[TEMP_GRAPH] No temperature data points received for sensor ${sensor.name}`);
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
    console.log(`[TEMP_GRAPH] Fetch complete. fetchError: ${fetchError}`);
    if (!fetchError) {
      processAndDrawChart();
    } else {
      console.log("[TEMP_GRAPH] Calling createChart after fetch error to draw empty state.");
      createChart(); 
    }
  }
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

// Add conversion functions
const celsiusToFahrenheit = (celsius: number): number => {
  return (celsius * 9/5) + 32;
};

const formatValue = (value: number) => {
  const fahrenheit = celsiusToFahrenheit(value);
  return `${fahrenheit.toFixed(1)}°F`;
};

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

    // Add hover effects for points
    points.on("mouseenter", function(event: MouseEvent, d: DataPoint) {
      const point = d3.select(this);
      point.attr("r", 5).attr("stroke-width", 2);

      // Update tooltip data
      tooltipData.value = {
        date: formatDate(d.time),
        value: formatValue(d.temperature)
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

  // Add legend
  const legend = svg.value!.append("g")
    .attr("font-family", "sans-serif")
    .attr("font-size", 10)
    .attr("text-anchor", "start")
    .selectAll("g")
    .data(sensors.value.filter(s => props.sensorVisibility[s.name]))
    .join("g")
    .attr("transform", (d, i) => `translate(0,${i * 20})`);

  // Add current filter information
  const filterInfo = svg.value!.append("g")
    .attr("transform", `translate(${width - 200}, 20)`);

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

  legend.append("rect")
    .attr("x", width - 19)
    .attr("width", 19)
    .attr("height", 19)
    .attr("fill", (d) => {
      const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === d.name);
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
        const sensorColorIndex = props.sensorConfigs.findIndex(sc => sc.name === sensor.name);
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
      console.log(`[TEMP_GRAPH] First aggregated point:`, sortedGroups[0]);
      console.log(`[TEMP_GRAPH] Last aggregated point:`, sortedGroups[sortedGroups.length - 1]);
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
    console.log(`[TEMP_GRAPH] Using manual date range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()}`);
  }

  const minTemperature = params.minValue;
  const maxTemperature = params.maxValue;

  const filtered = sensors.value.map(sensor => {
    const initialPoints = sensor.data.length;
    console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: Filtering ${initialPoints} points...`);

    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time;
      const temperature = point.temperature;
      
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

    console.log(`[TEMP_GRAPH] Sensor ${sensor.name}: ${initialPoints} points -> ${sensorFilteredData.length} points after filtering`);
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

// Add watch for dynamicTimeWindow prop
watch(() => props.dynamicTimeWindow, (newWindow) => {
  console.log(`[TEMP_GRAPH] dynamicTimeWindow changed to: ${newWindow}`);
  if (props.dataType === 'temperature') {
    console.log('[TEMP_GRAPH] Refetching data for new time window');
    fetchAllSensorData();
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
</style> 