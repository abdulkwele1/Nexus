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
    minMoisture: number;
    maxMoisture: number;
    resolution: 'raw' | 'hourly' | 'daily' | 'weekly' | 'monthly';
  }
  sensorVisibility: { [key: string]: boolean };
  dynamicTimeWindow?: 'none' | 'lastHour' | 'last24Hours' | 'last7Days' | 'last30Days';
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
    const allSensorsDataPromises = SENSOR_CONFIGS.map(async (config, index) => {
      const rawDataPoints = await nexusStore.user.getSensorMoistureData(config.id);
      console.log(`[soilMoistureGraph] Raw data for ${config.name} (ID: ${config.id}):`, JSON.parse(JSON.stringify(rawDataPoints)));
      if (rawDataPoints && Array.isArray(rawDataPoints)) {
        const formattedData: DataPoint[] = rawDataPoints.map((point: any) => ({
          time: new Date(point.date),
          moisture: Number(point.soil_moisture),
        })).sort((a, b) => a.time.getTime() - b.time.getTime());
        // Temporarily update sensors.value with the raw data. 
        // The subsequent call to updateChart will process it.
        sensors.value[index].data = formattedData; 
      } else {
        console.warn(`No data returned for sensor ${config.name} (ID: ${config.id}) or data is not an array. API returned:`, rawDataPoints);
        sensors.value[index].data = [];
      }
    });
    await Promise.all(allSensorsDataPromises);
  } catch (error) {
    console.error("Error fetching sensor data:", error);
    fetchError = true;
    sensors.value.forEach(sensor => sensor.data = []); // Clear data on error
  } finally {
    if (!fetchError) {
      // After fetching ALL raw data, trigger the update process which will filter/aggregate
      console.log('[soilMoistureGraph] fetchAllSensorData completed. Triggering chart update/processing...');
      // Call the function that handles filtering/aggregation based on current props
      processAndDrawChart(); 
    } else {
      // On error, clear the chart by drawing with the already cleared sensors.value
      createChart(); 
    }
    // REMOVED direct call to createChart() or updateChart() without processing
  }
}

// Function to process data based on current state and draw
const processAndDrawChart = () => {
  console.log(`[soilMoistureGraph] processAndDrawChart: Processing data with current queryParams:`, JSON.stringify(props.queryParams));
  // Pass the current sensors.value (which holds raw data after fetch, or processed data after filter change) 
  // to filterData. filterData needs to know it might receive raw data.
  const processedData = filterData(props.queryParams); // Applies filter/aggregation based on props
  updateChart(processedData); // Update chart with the processed data
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
  return `${value.toFixed(1)}%`;
};

const createChart = () => {
  // ****** ADDED LOG to check resolution ******
  console.log('[soilMoistureGraph] createChart called. Current resolution from props:', props.queryParams.resolution);
  // ****** END ADDED LOG ******

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

  // Combine visible sensor data for domain calculation - USE PROP
  const allData = sensors.value
    .filter(sensor => props.sensorVisibility[sensor.name]) // USE PROP
    .flatMap(sensor => sensor.data);

  let yMinActual = d3.min(allData, d => d.moisture) as number;
  let yMaxActual = d3.max(allData, d => d.moisture) as number;

  // Handle cases where there's no data or min/max are undefined
  if (allData.length === 0 || yMinActual === undefined || yMaxActual === undefined) {
      yMinActual = 0;  // Default min for empty chart
      yMaxActual = 50; // Default max for empty chart
      if (allData.length > 0 && yMinActual === undefined && yMaxActual === undefined) { // Only one data point
        const singleValue = allData[0].moisture;
        yMinActual = singleValue;
        yMaxActual = singleValue;
      }
  }

  console.log(`[soilMoistureGraph] Actual Min: ${yMinActual}, Actual Max: ${yMaxActual} (from sensors visible via prop)`);

  const dataRange = yMaxActual - yMinActual;

  let yDomainMinWithPadding;
  let yDomainMaxWithPadding;

  if (dataRange === 0) { // If all data points are the same value (or only one point)
      yDomainMinWithPadding = Math.max(0, yMinActual - 10);       // Give 10 units below
      yDomainMaxWithPadding = Math.min(100, yMaxActual + 10);   // Give 10 units above
      // Ensure there's a visible range if min and max ended up the same (e.g., data is 0 or 100)
      if (yDomainMinWithPadding === yDomainMaxWithPadding) {
          if (yDomainMinWithPadding <= 10) { // If it's near 0
            yDomainMaxWithPadding = yDomainMinWithPadding + 20; 
          } else if (yDomainMaxWithPadding >= 90) { // If it's near 100
            yDomainMinWithPadding = yDomainMaxWithPadding - 20;
          } else { // Otherwise, center it
            yDomainMinWithPadding -= 10;
            yDomainMaxWithPadding += 10;
          }
      }
  } else {
      const paddingPercentage = 0.25; // Add 25% padding to both top and bottom of the data range
      let rangePadding = dataRange * paddingPercentage;

      // If data range is very small, ensure padding is at least a few units
      if (dataRange > 0 && rangePadding < 2.5) {
          rangePadding = 2.5; // Minimum 2.5 units padding either side for small ranges
      }

      yDomainMinWithPadding = yMinActual - rangePadding;
      yDomainMaxWithPadding = yMaxActual + rangePadding;
  }

  // Final checks to ensure min/max are within 0-100 bounds and sensible
  yDomainMinWithPadding = Math.max(0, yDomainMinWithPadding);
  yDomainMaxWithPadding = Math.min(100, yDomainMaxWithPadding);

  // Ensure min is not >= max, and there's a minimum visible range (e.g. 10 units)
  if (yDomainMinWithPadding >= yDomainMaxWithPadding) {
      yDomainMinWithPadding = Math.max(0, yDomainMaxWithPadding - 10); 
      // If still equal (e.g. max was 0 or min was 100), adjust one side
      if (yDomainMinWithPadding === yDomainMaxWithPadding) {
          if (yDomainMaxWithPadding <= 10) yDomainMaxWithPadding = yDomainMinWithPadding + 10; // eg. 0, make it 0-10
          else yDomainMinWithPadding = yDomainMaxWithPadding - 10; // e.g. 100, make it 90-100
      }
  }
  if (yDomainMaxWithPadding - yDomainMinWithPadding < 5) { // Ensure at least 5 units of visible range if possible
    const currentMidpoint = (yDomainMaxWithPadding + yDomainMinWithPadding) / 2;
    yDomainMinWithPadding = Math.max(0, currentMidpoint - 2.5);
    yDomainMaxWithPadding = Math.min(100, currentMidpoint + 2.5);
    if (yDomainMinWithPadding === yDomainMaxWithPadding) { // Last resort for flat lines at 0 or 100
        yDomainMaxWithPadding = yDomainMinWithPadding + (yDomainMinWithPadding < 50 ? 5: -5) // Add 5 or subtract 5
        yDomainMinWithPadding = Math.max(0, yDomainMinWithPadding);
        yDomainMaxWithPadding = Math.min(100, yDomainMaxWithPadding);
    }
  }


  console.log(`[soilMoistureGraph] Calculated Y-Domain Min: ${yDomainMinWithPadding}, Max: ${yDomainMaxWithPadding} WITH DYNAMIC PADDING`);

  // Create scales
  const x = d3.scaleUtc()
    .domain(d3.extent(allData, d => d.time) as [Date, Date])
    .range([marginLeft, width - marginRight]);

  const y = d3.scaleLinear()
    .domain([yDomainMinWithPadding, yDomainMaxWithPadding]) 
    .range([height - marginBottom, marginTop]);

  // Create line generator
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(d.moisture));

  // Determine X-axis tick formatter based on resolution
  let xAxisTickFormat: (date: Date) => string;
  
  if (props.queryParams.resolution === 'raw' || props.queryParams.resolution === 'hourly') {
    // For raw and hourly data, show time in addition to date
    xAxisTickFormat = localTimeFormatter;
  } else {
    // For daily, weekly, monthly, only show the date without time
    xAxisTickFormat = d3.timeFormat("%b %d"); // Format like "Jan 15"
  }

  // Add x-axis
  svg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x) 
        .ticks(width / 80)
        .tickSizeOuter(0)
        .tickFormat(xAxisTickFormat as any) 
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

  // Add lines for each visible sensor - USE PROP
  sensors.value.forEach((sensor, i) => {
    if (!props.sensorVisibility[sensor.name]) return; // USE PROP

    const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === sensor.name);
    const color = colors[sensorColorIndex % colors.length];

    const path = svg.value!.append("path")
      .datum(sensor.data)
      .attr("fill", "none")
      .attr("stroke", color) // Use mapped color
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

  // Add legend - USE PROP
  const legend = svg.value.append("g")
    .attr("font-family", "sans-serif")
    .attr("font-size", 10)
    .attr("text-anchor", "start")
    .selectAll("g")
    .data(sensors.value.filter(s => props.sensorVisibility[s.name])) // USE PROP
    .join("g")
      .attr("transform", (d, i) => `translate(0,${i * 20})`);

  legend.append("rect")
    .attr("x", width - 19)
    .attr("width", 19)
    .attr("height", 19)
    .attr("fill", (d) => { // Find color based on filtered data name
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

  // Add mouse interaction for tooltip - USE PROP
  const bisect = d3.bisector<DataPoint, Date>(d => d.time).center;
  
  svg.value.on("pointermove", (event) => {
    const visibleSensorsWithData = sensors.value.filter(s => props.sensorVisibility[s.name] && s.data.length > 0);
    if (!tooltip.value || visibleSensorsWithData.length === 0) { 
      if (tooltip.value) tooltip.value.style("display", "none"); 
      return;
    }

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    
    // Find the closest data point for each visible sensor - USE PROP
    const tooltipData = visibleSensorsWithData // Use pre-filtered list
      .map((sensor) => { // Removed index 'i' as it's not reliable here
        const sensorColorIndex = SENSOR_CONFIGS.findIndex(sc => sc.name === sensor.name);
        const color = colors[sensorColorIndex % colors.length]; 

        const index = bisect(sensor.data, xPos);
        const dataPoint = sensor.data[Math.max(0, Math.min(index, sensor.data.length - 1))];
        if (!dataPoint) return null; 

        // ****** ADDED LOG for Step 2: Tooltip Raw Value ******
        if (sensor.name === 'Sensor Alpha') {
          console.log('[soilMoistureGraph] Tooltip raw value for Sensor Alpha:', dataPoint.moisture, 'at time:', dataPoint.time.toISOString());
        }
        // ****** END ADDED LOG ******

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
watch(() => props.queryParams, () => {
  // Query params changed, process the current raw data and redraw
  console.log('[soilMoistureGraph] Query Params changed. Triggering chart processing...');
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

// Add data filtering function
const filterData = (params: Props['queryParams']) => {
  const now = new Date();
  let filterRangeStart: Date;
  let filterRangeEnd: Date;
  let useInclusiveEnd = false; // For ranges ending 'now' or being precise single points

  if (props.dynamicTimeWindow && props.dynamicTimeWindow !== 'none') {
    console.log(`[soilMoistureGraph] filterData: Using dynamicTimeWindow = ${props.dynamicTimeWindow}`);
    filterRangeEnd = now; // Default end for dynamic windows is 'now'
    useInclusiveEnd = true; // Dynamic windows are typically inclusive of their end time (now)

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
        // This case should ideally not be reached if props.dynamicTimeWindow is valid and not 'none'
        console.warn(`[soilMoistureGraph] filterData: Unexpected dynamicTimeWindow value: ${props.dynamicTimeWindow}. Falling back to queryParams.`);
        filterRangeStart = new Date(params.startDate + 'T00:00:00');
        const tempEndDate = new Date(params.endDate + 'T00:00:00');
        tempEndDate.setDate(tempEndDate.getDate() + 1);
        filterRangeEnd = tempEndDate;
        useInclusiveEnd = false; // Day ranges are exclusive of the end time (start of next day)
        break;
    }
  } else {
    console.log('[soilMoistureGraph] filterData: Using queryParams for date range.');
    filterRangeStart = new Date(params.startDate + 'T00:00:00');
    const tempEndDate = new Date(params.endDate + 'T00:00:00');
    tempEndDate.setDate(tempEndDate.getDate() + 1); // Start of the next day
    filterRangeEnd = tempEndDate;
    useInclusiveEnd = false; // Day ranges are exclusive of the end time (start of next day)
  }

  console.log(`[soilMoistureGraph] filterData: Effective time range: ${filterRangeStart.toISOString()} to ${filterRangeEnd.toISOString()} (inclusiveEnd: ${useInclusiveEnd})`);

  const minMoisture = params.minMoisture;
  const maxMoisture = params.maxMoisture;
  
  // Log the filter boundaries from params (original date strings)
  // console.log('[soilMoistureGraph] Filtering with Start Day (Local 00:00): ', startDate, ' (ISO:', startDate.toISOString(), ') End Day (Local 23:59 implicitly by < next day): ', endDateRaw, ' (< ', endOfRangeDate.toISOString(), ') MinM:', minMoisture, 'MaxM:', maxMoisture);
  // console.log('[soilMoistureGraph] Query Params received (Date strings):', JSON.stringify(params));

  const filtered = sensors.value.map(sensor => {
    // Log a few original data points for this sensor before filtering
    // if (sensor.data.length > 0) {
    //   console.log(`[soilMoistureGraph] Original data for ${sensor.name} (first 3 points):`, 
    //     JSON.stringify(sensor.data.slice(0, 3).map(p => ({ time: p.time.toISOString(), moisture: p.moisture })))
    //   );
    // }

    const sensorFilteredData = sensor.data.filter(point => {
      const date = point.time; // point.time is already a Date object
      const moisture = point.moisture;
      
      let inTimeRange;
      if (useInclusiveEnd) {
        inTimeRange = date >= filterRangeStart && date <= filterRangeEnd;
      } else {
        inTimeRange = date >= filterRangeStart && date < filterRangeEnd;
      }
      
      const isAboveMinMoisture = moisture >= minMoisture;
      const isBelowMaxMoisture = moisture <= maxMoisture;
      
      const includePoint = inTimeRange && isAboveMinMoisture && isBelowMaxMoisture;
      
      // Log decision for a sample of points
      // if (Math.random() < 0.01) { // Adjust frequency as needed
      //   console.log(`[soilMoistureGraph] Point: ${date.toISOString()} (${moisture}), Start: ${filterRangeStart.toISOString()}, End: ${filterRangeEnd.toISOString()}, MinM: ${minMoisture}, MaxM: ${maxMoisture}, Included: ${includePoint}`);
      // }
      return includePoint;
    });

    console.log(`[soilMoistureGraph] Sensor ${sensor.name} - Original points: ${sensor.data.length}, After time/moisture filter: ${sensorFilteredData.length}`);

    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

  // ****** ADDED LOG for Aggregation Decision ******
  console.log(`[soilMoistureGraph] filterData: Checking resolution for aggregation. Current resolution: "${params.resolution}"`);
  // ****** END ADDED LOG ******

  // Apply resolution aggregation if needed
  if (params.resolution !== 'raw') {
    // ****** ADDED LOG for Calling Aggregation ******
    console.log(`[soilMoistureGraph] filterData: Resolution is "${params.resolution}", calling aggregateData.`);
    // ****** END ADDED LOG ******
    return aggregateData(filtered, params.resolution);
  }

  return filtered;
};

// Add data aggregation function
const aggregateData = (sensorsToAggregate: Sensor[], resolution: 'hourly' | 'daily' | 'weekly' | 'monthly'): Sensor[] => {
  return sensorsToAggregate.map(sensor => {
    // ****** ADDED LOG for Starting Aggregation ******
    console.log(`[soilMoistureGraph] aggregateData: Starting aggregation for sensor "${sensor.name}" with resolution "${resolution}". Input points: ${sensor.data?.length || 0}`);
    // ****** END ADDED LOG ******

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
      const averageMoisture = group.sum / group.count;
      // ****** ADDED LOG for Step 3 (Conditional): Aggregation Results ******
      if (sensor.name === 'Sensor Alpha') { // Log only for Sensor Alpha
        console.log(`[soilMoistureGraph] AGGREGATION for Sensor Alpha - Group Time: ${group.time.toISOString()}, Avg Moisture: ${averageMoisture}, Sum: ${group.sum}, Count: ${group.count}`);
      }
      // ****** END ADDED LOG ******
      aggregatedData.push({
        time: group.time,
        moisture: averageMoisture,
      });
    });

    // Sort by time, as grouping might change order
    aggregatedData.sort((a, b) => a.time.getTime() - b.time.getTime());

    // ****** ADDED LOG for Aggregation Result ******
    console.log(`[soilMoistureGraph] aggregateData: Finished aggregation for sensor "${sensor.name}". Output points: ${aggregatedData.length}`);
    // ****** END ADDED LOG ******

    return { ...sensor, data: aggregatedData };
  });
};

// Modify updateChart to accept the processed data to draw
const updateChart = (dataToDraw: Sensor[]) => {
  console.log('[soilMoistureGraph] updateChart called with processed data. Points count:', dataToDraw.reduce((sum, s) => sum + s.data.length, 0));
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