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
  try {
    const allSensorsDataPromises = SENSOR_CONFIGS.map(async (config, index) => {
      const rawDataPoints = await nexusStore.user.getSensorTemperatureData(config.id);
      
      if (rawDataPoints && Array.isArray(rawDataPoints)) {
        const formattedData: DataPoint[] = rawDataPoints.map((point: any) => ({
          time: new Date(point.date),
          temperature: Number(point.soil_temperature),
        })).sort((a, b) => a.time.getTime() - b.time.getTime());
        
        sensors.value[index].data = formattedData;
      } else {
        console.warn(`No temperature data for sensor ${config.name}`);
        sensors.value[index].data = [];
      }
    });
    await Promise.all(allSensorsDataPromises);
  } catch (error) {
    console.error("Error fetching temperature data:", error);
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

const processAndDrawChart = () => {
  console.log(`[soilTemperatureGraph] processAndDrawChart: Processing data with current queryParams:`, JSON.stringify(props.queryParams));
  const processedData = filterData(props.queryParams);
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
    .domain(d3.extent(allData, d => d.time) as [Date, Date])
    .range([marginLeft, width - marginRight]);

  const y = d3.scaleLinear()
    .domain([yDomainMinWithPadding, yDomainMaxWithPadding])
    .range([height - marginBottom, marginTop]);

  // Create line generator with Fahrenheit conversion
  const line = d3.line<DataPoint>()
    .x(d => x(d.time))
    .y(d => y(celsiusToFahrenheit(d.temperature)));

  // Determine if we should show time based on data range
  const timeRange = x.domain()[1].getTime() - x.domain()[0].getTime();
  const isWithin24Hours = timeRange <= 24 * 60 * 60 * 1000;

  // Add x-axis with conditional formatting
  svg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x)
      .ticks(width / 80)
      .tickSizeOuter(0)
      .tickFormat(isWithin24Hours ? 
        d3.timeFormat("%H:%M") as any : 
        d3.timeFormat("%b %d") as any
      )
    );

  // Add y-axis with Fahrenheit values
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
      .text("↑ Soil Temperature (°F)"));

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
watch(() => props.queryParams, () => {
  console.log('[soilTemperatureGraph] Query Params changed. Triggering chart processing...');
  processAndDrawChart();
}, { deep: true });

// Add watch for sensorVisibility prop
watch(() => props.sensorVisibility, () => {
  console.log('[soilTemperatureGraph] sensorVisibility prop changed. Triggering createChart...');
  createChart();
}, { deep: true });

// Watch for dynamicTimeWindow prop
watch(() => props.dynamicTimeWindow, () => {
  console.log(`[soilTemperatureGraph] dynamicTimeWindow prop changed to: ${props.dynamicTimeWindow}. Triggering chart processing...`);
  processAndDrawChart();
});

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

  const minTemperature = params.minValue;
  const maxTemperature = params.maxValue;

  const filtered = sensors.value.map(sensor => {
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

    return {
      ...sensor,
      data: sensorFilteredData
    };
  });

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
      const date = new Date(point.time);

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
        let groupTime = d3.timeHour.floor(date);
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