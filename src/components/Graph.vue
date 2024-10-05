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

// Create the tooltip once (outside of createChart to avoid recreating multiple tooltips)
const tooltip = d3.select("body").append("div")
  .attr("class", "tooltip")
  .style("opacity", 0)
  .style("z-index", 1000)  // Ensures the tooltip is above everything
  .style("position", "absolute");

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

  if (props.isLineChart) {
    const line = d3.line()
      .x(d => x(d3.timeFormat("%Y-%m-%d")(d.date)) + x.bandwidth() / 2)
      .y(d => y(d.production));

    // Draw the line
    svg.append("path")
      .datum(props.solarData)
      .attr("fill", "none")
      .attr("stroke", "#69b3a2")
      .attr("stroke-width", 2)
      .attr("d", line);

    // Create hover circles
    svg.selectAll(".hover-circle")
      .data(props.solarData)
      .enter()
      .append("circle")
      .attr("class", "hover-circle")
      .attr("cx", d => x(d3.timeFormat("%Y-%m-%d")(d.date)) + x.bandwidth() / 2)
      .attr("cy", d => y(d.production))
      .attr("r", 8)  // Easier hover target
      .attr("fill", "transparent")
      .on("mouseover", (event, d) => {
        tooltip.transition().duration(200).style("opacity", 1);
      })
      .on("mousemove", (event, d) => {
        tooltip.html(`Date: ${d3.timeFormat("%Y-%m-%d")(d.date)}<br>kWh: ${d.production}`)
          .style("left", (event.clientX + window.scrollX + 10) + "px")
          .style("top", (event.clientY + window.scrollY - 15) + "px");
      })
      .on("mouseout", () => {
        tooltip.transition().duration(200).style("opacity", 0);
      });
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
      .attr("fill", "#69b3a2")
      .on("mouseover", (event, d) => {
        tooltip.transition().duration(200).style("opacity", 1);
      })
      .on("mousemove", (event, d) => {
        tooltip.html(`Date: ${d3.timeFormat("%Y-%m-%d")(d.date)}<br>kWh: ${d.production}`)
          .style("left", (event.clientX + window.scrollX + 10) + "px")
          .style("top", (event.clientY + window.scrollY - 15) + "px");
      })
      .on("mouseout", () => {
        tooltip.transition().duration(200).style("opacity", 0);
      });
  }

  // Add y-axis label
  svg.append("text")
    .attr("transform", "rotate(-90)")
    .attr("x", -height / 2)
    .attr("y", margin.left / 2)
    .attr("dy", "-1em")
    .attr("fill", "currentColor")
    .attr("text-anchor", "middle")
    .text("(kWh)");
};

watch(() => [props.solarData, props.isLineChart], createChart);

onMounted(() => {
  createChart(); // Create the initial chart
});
</script>
