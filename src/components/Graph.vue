<template>
  <div ref="chartContainer" class="chart-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, defineProps } from 'vue';
import * as d3 from 'd3';

const props = defineProps<{
  solarData: { date: Date; production: number }[];
  isLineChart: boolean;
}>();

const chartContainer = ref(null);

// Create the chart
const createChart = () => {
  d3.select(chartContainer.value).select("svg").remove(); // Clear previous SVG

  const width = 960;
  const height = 500;
  const margin = { top: 30, right: 20, bottom: 50, left: 150 };

  const x = d3.scaleBand()
    .domain(props.solarData.map(d => d3.timeFormat("%Y-%m-%d")(d.date)))
    .range([margin.left, width - margin.right])
    .padding(0.1);

  const y = d3.scaleLinear()
    .domain([0, d3.max(props.solarData, d => d.production)]).nice()
    .range([height - margin.bottom, margin.top]);

  const svg = d3.select(chartContainer.value)
    .append("svg")
    .attr("width", width)
    .attr("height", height);

  // Create the x-axis
  svg.append("g")
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x).tickValues(x.domain().filter((d, i) => !(i % Math.floor(props.solarData.length / 10)))))
    .selectAll("text")
    .attr("transform", "rotate(-45)")
    .style("text-anchor", "end");

  // Create the y-axis
  svg.append("g")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y));

  // Draw either bars or a line based on the checkbox state
  if (props.isLineChart) {
    const line = d3.line()
      .x(d => x(d3.timeFormat("%Y-%m-%d")(d.date)) + x.bandwidth() / 2)
      .y(d => y(d.production));

    svg.append("path")
      .datum(props.solarData)
      .attr("fill", "none")
      .attr("stroke", "#69b3a2")
      .attr("stroke-width", 2)
      .attr("d", line);
  } else {
    svg.selectAll(".bar")
      .data(props.solarData)
      .enter()
      .append("rect")
      .attr("class", "bar")
      .attr("x", d => x(d3.timeFormat("%Y-%m-%d")(d.date)))
      .attr("y", d => y(d.production))
      .attr("width", x.bandwidth())
      .attr("height", d => y(0) - y(d.production))
      .attr("fill", "#69b3a2");
  }

  svg.append("text")
    .attr("transform", "rotate(-90)")
    .attr("x", -height / 2)
    .attr("y", margin.left / 2)
    .attr("dy", "-1em")
    .attr("fill", "currentColor")
    .attr("text-anchor", "middle")
    .text("(kWh)");
};

// Watch for changes in the solar data and chart type
watch(() => [props.solarData, props.isLineChart], createChart);

onMounted(() => {
  createChart(); // Create the initial chart
});

// Example functions to calculate X and Y positions for the chart
const calculateX = (date) => {
  return d3.scaleBand().domain(props.solarData.map(d => d3.timeFormat("%Y-%m-%d")(d.date))).range([150, 960 - 20])(d3.timeFormat("%Y-%m-%d")(date)) + 150 / 2; // Add margin.left to center the circle
};

const calculateY = (production) => {
  return d3.scaleLinear()
    .domain([0, d3.max(props.solarData, d => d.production)]).nice()
    .range([500 - 50, 30])(production); // Add margin.bottom and margin.top
};
</script>

<style scoped>
.chart-container {
  width: 100%;
  height: 500px;
}

</style>
