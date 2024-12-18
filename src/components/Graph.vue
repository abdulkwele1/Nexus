<template>
  <div ref="chartContainer" class="chart-container"></div>
</template>

<script setup lang="ts">
import { onMounted, watch, defineProps, ref } from 'vue';
import * as d3 from 'd3';

const props = defineProps({
  solarData: {
    type: Array,
    required: true,
  },

  isLineChart:{
    type: Boolean,
    required: true,
  },
});

const chartContainer = ref(null);

const createChart = () => {
  d3.select(chartContainer.value).select("svg").remove();

  const width = 960;
  const height = 500;
  const margin = { top: 30, right: 20, bottom: 50, left: 150 };

  const x = d3.scaleBand()
    .domain(props.solarData.length ? props.solarData.map(d => d3.timeFormat("%Y-%m-%d")(d.date)) : [" "])
    .range([margin.left, width - margin.right])
    .padding(0.1);

  const y = d3.scaleLinear()
    .domain([0, d3.max(props.solarData, d => d.production) || 1])
    .range([height - margin.bottom, margin.top]);

  const svg = d3.select(chartContainer.value)
    .append("svg")
    .attr("width", width)
    .attr("height", height);

  svg.append("g")
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x).tickValues(props.solarData.length ? x.domain() : [" "]))
    .selectAll("text")
    .attr("transform", "rotate(-45)")
    .style("text-anchor", "end");

  svg.append("g")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y));

  // Tooltip setup
  const tooltip = d3.select(chartContainer.value)
    .append("div")
    .style("position", "absolute")
    .style("background", "#f9f9f9")
    .style("padding", "5px")
    .style("border", "1px solid #d3d3d3")
    .style("border-radius", "5px")
    .style("visibility", "hidden")
    .style("pointer-events", "none");

  const showTooltip = (event, d) => {
    tooltip
      .html(`Date: ${d3.timeFormat("%Y/%m/%d")(d.date)}<br>kWh: ${d.production}`)
      .style("visibility", "visible");
  };

  const moveTooltip = (event) => {
    tooltip
      .style("top", (event.pageY - 10) + "px")
      .style("left", (event.pageX + 10) + "px");
  };

  const hideTooltip = () => {
    tooltip.style("visibility", "hidden");
  };

  // Render chart based on type
  if (props.solarData.length) {
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

      svg.selectAll(".hover-circle")
        .data(props.solarData)
        .enter()
        .append("circle")
        .attr("class", "hover-circle")
        .attr("cx", d => x(d3.timeFormat("%Y-%m-%d")(d.date)) + x.bandwidth() / 2)
        .attr("cy", d => y(d.production))
        .attr("r", 4)
        .attr("fill", "#69b3a2")
        .on("mouseover", showTooltip)
        .on("mousemove", moveTooltip)
        .on("mouseout", hideTooltip);
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
        .on("mouseover", showTooltip)
        .on("mousemove", moveTooltip)
        .on("mouseout", hideTooltip);
    }
  }
};

// Watch for changes in solarData or isLineChart to re-render chart
watch(() => props.solarData, createChart, { immediate: true });
watch(() => props.isLineChart, createChart, { immediate: true });

onMounted(createChart);
</script>

<style scoped>
.chart-container {
  width: 100%;
  height: 100%;
}
</style>
