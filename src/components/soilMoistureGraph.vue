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
import { ref, onMounted, watch } from 'vue';
import * as d3 from 'd3';

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

const chartContainer = ref<HTMLElement | null>(null);
const svg = ref<d3.Selection<SVGSVGElement, unknown, null, undefined> | null>(null);
const tooltip = ref<d3.Selection<SVGGElement, unknown, null, undefined> | null>(null);

const colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728'];

const mockData: SensorData = {
  sensor1: Array.from({ length: 24 }, (_, i) => ({
    time: new Date(2024, 0, 1, i),
    moisture: Math.random() * 30 + 20 // Random values between 20-50%
  })),
  sensor2: Array.from({ length: 24 }, (_, i) => ({
    time: new Date(2024, 0, 1, i),
    moisture: Math.random() * 30 + 20
  })),
  sensor3: Array.from({ length: 24 }, (_, i) => ({
    time: new Date(2024, 0, 1, i),
    moisture: Math.random() * 30 + 20
  })),
  sensor4: Array.from({ length: 24 }, (_, i) => ({
    time: new Date(2024, 0, 1, i),
    moisture: Math.random() * 30 + 20
  }))
};

const sensors = ref<Sensor[]>([
  { data: mockData.sensor1, name: 'Sensor 1', visible: true },
  { data: mockData.sensor2, name: 'Sensor 2', visible: true },
  { data: mockData.sensor3, name: 'Sensor 3', visible: true },
  { data: mockData.sensor4, name: 'Sensor 4', visible: true }
]);

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

  // Add x-axis
  svg.value.append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x).ticks(width / 80).tickSizeOuter(0));

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
    if (!tooltip.value) return;

    const pointer = d3.pointer(event);
    const xPos = x.invert(pointer[0]);
    
    // Find the closest data point for each visible sensor
    const tooltipData = sensors.value
      .filter(sensor => sensor.visible)
      .map((sensor, i) => {
        const index = bisect(sensor.data, xPos);
        return {
          name: sensor.name,
          value: sensor.data[index].moisture,
          time: sensor.data[index].time,
          color: colors[i]
        };
      });

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

onMounted(() => {
  createChart();
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