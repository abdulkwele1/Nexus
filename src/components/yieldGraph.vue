<template>
  <div ref="chartContainer" class="chart-container">
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
import { onMounted, watch, defineProps, ref } from 'vue';
import * as d3 from 'd3';

interface DataPoint {
  date: Date;
  kwh_yield: number;
}

const props = defineProps({
  solarData: {
    type: Array as () => DataPoint[],
    required: true,
  },

  isLineChart:{
    type: Boolean,
    required: true,
  },
});

const chartContainer = ref<HTMLElement | null>(null);
const activeTooltip = ref(false);
const tooltipData = ref({ date: '', value: '' });
const tooltipStyle = ref({
  left: '0px',
  top: '0px'
});

const formatDate = (date: Date) => {
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric'
  });
};

const formatValue = (value: number) => {
  return `${value.toFixed(1)} kWh`;
};

const createChart = () => {
  if (!chartContainer.value) return;
  d3.select(chartContainer.value).select("svg").remove();

  const width = 960;
  const height = 500;
  const margin = { top: 30, right: 20, bottom: 50, left: 150 };

  // Sort data by date to ensure proper line connection
  const sortedData = [...props.solarData].sort((a, b) => a.date.getTime() - b.date.getTime());

  const x = d3.scaleBand()
    .domain(sortedData.length ? sortedData.map(d => d3.timeFormat("%Y-%m-%d")(d.date)) : [" "])
    .range([margin.left, width - margin.right])
    .padding(0.1);

  const y = d3.scaleLinear()
    .domain([0, d3.max(sortedData, d => d.kwh_yield) || 1])
    .range([height - margin.bottom, margin.top]);

  const svg = d3.select(chartContainer.value)
    .append("svg")
    .attr("width", width)
    .attr("height", height);

  svg.append("g")
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x).tickValues(sortedData.length ? x.domain() : [" "]))
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

  const showTooltip = (event: MouseEvent, d: DataPoint) => {
    if (!chartContainer.value) return;
    const rect = chartContainer.value.getBoundingClientRect();
    tooltipData.value = {
      date: formatDate(d.date),
      value: formatValue(d.kwh_yield)
    };
    tooltipStyle.value = {
      left: `${event.clientX - rect.left + 10}px`,
      top: `${event.clientY - rect.top - 10}px`
    };
    activeTooltip.value = true;
  };

  const moveTooltip = (event: MouseEvent) => {
    if (!chartContainer.value) return;
    const rect = chartContainer.value.getBoundingClientRect();
    tooltipStyle.value = {
      left: `${event.clientX - rect.left + 10}px`,
      top: `${event.clientY - rect.top - 10}px`
    };
  };

  const hideTooltip = () => {
    activeTooltip.value = false;
  };

  // Render chart based on type
  if (sortedData.length) {
    if (props.isLineChart) {
      // Create a time scale for the line chart
      const timeScale = d3.scaleTime()
        .domain(d3.extent(sortedData, d => d.date) as [Date, Date])
        .range([margin.left, width - margin.right]);

      // Create the line generator with linear interpolation
      const line = d3.line<DataPoint>()
        .x(d => timeScale(d.date))
        .y(d => y(d.kwh_yield))
        .curve(d3.curveLinear); // Use linear interpolation

      // Add the line path
      svg.append("path")
        .datum(sortedData)
        .attr("fill", "none")
        .attr("stroke", "#69b3a2")
        .attr("stroke-width", 2)
        .attr("d", line);

      // Add points
      svg.selectAll(".hover-circle")
        .data(sortedData)
        .enter()
        .append("circle")
        .attr("class", "hover-circle")
        .attr("cx", d => timeScale(d.date))
        .attr("cy", d => y(d.kwh_yield))
        .attr("r", 4)
        .attr("fill", "#69b3a2")
        .on("mouseover", showTooltip)
        .on("mousemove", moveTooltip)
        .on("mouseout", hideTooltip);

      // Update x-axis for line chart
      const xAxis = d3.axisBottom(timeScale)
        .ticks(d3.timeDay.every(1))
        .tickFormat(d3.timeFormat("%Y-%m-%d") as any);

      svg.select("g")
        .attr("transform", `translate(0,${height - margin.bottom})`)
        .call(xAxis as any)
        .selectAll("text")
        .attr("transform", "rotate(-45)")
        .style("text-anchor", "end");
    } else {
      svg.selectAll(".bar")
        .data(sortedData)
        .enter()
        .append("rect")
        .attr("class", "bar")
        .attr("x", d => x(d3.timeFormat("%Y-%m-%d")(d.date)) || 0)
        .attr("y", d => y(d.kwh_yield))
        .attr("width", x.bandwidth())
        .attr("height", d => y(0) - y(d.kwh_yield))
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
  position: relative;
}

.tooltip-container {
  position: absolute;
  background: white;
  border: 1px solid #ccc;
  border-radius: 4px;
  padding: 8px 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  pointer-events: none;
  z-index: 1000;
  min-width: 120px;
}

.tooltip-content {
  font-size: 12px;
  line-height: 1.4;
}

.tooltip-date {
  color: #666;
  margin-bottom: 4px;
}

.tooltip-value {
  color: #333;
  font-weight: 500;
}

.points circle {
  transition: r 0.2s, stroke-width 0.2s;
}
</style>
