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
import { useNexusStore } from '@/stores/nexus';

const store = useNexusStore();

interface DataPoint {
  date: Date;
  kwh_yield: number;
}

const props = defineProps({
  solarData: {
    type: Array as () => DataPoint[],
    required: true,
  },
  isLineChart: {
    type: Boolean,
    default: false,
  },
  startDate: {
    type: String,
    default: null,
  },
  endDate: {
    type: String,
    default: null,
  },
});

const solarData = ref<DataPoint[]>([]);
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

const createChart = () => {
  if (!chartContainer.value) return;
  d3.select(chartContainer.value).select("svg").remove();

  const width = 900;
  const height = 600;
  const margin = { 
    top: 40, 
    right: 40, 
    bottom: 60, 
    left: 130 // Further increased left margin
  };

  // Sort data by date
  const sortedData = [...props.solarData].sort((a, b) => a.date.getTime() - b.date.getTime());

  const x = d3.scaleBand()
    .domain(sortedData.map(d => d3.timeFormat("%Y-%m-%d")(d.date)))
    .range([margin.left, width - margin.right])
    .padding(0.2);

  const y = d3.scaleLinear()
    .domain([0, d3.max(sortedData, d => d.kwh_yield) || 1])
    .range([height - margin.bottom, margin.top])
    .nice();

  const svg = d3.select(chartContainer.value)
    .append("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", `0 0 ${width} ${height}`);

  // Add gradient definition
  const gradient = svg.append("defs")
    .append("linearGradient")
    .attr("id", "yield-gradient")
    .attr("x1", "0%")
    .attr("y1", "0%")
    .attr("x2", "0%")
    .attr("y2", "100%");

  gradient.append("stop")
    .attr("offset", "0%")
    .attr("stop-color", "#90EE90")
    .attr("stop-opacity", 0.3);

  gradient.append("stop")
    .attr("offset", "100%")
    .attr("stop-color", "#90EE90")
    .attr("stop-opacity", 0.05);

  // Add grid lines
  svg.append("g")
    .attr("class", "grid-lines")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y)
      .tickSize(-width + margin.left + margin.right)
      .tickFormat(() => ""));

  // Add x-axis
  svg.append("g")
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x))
    .selectAll("text")
    .attr("y", 10) // Move text slightly down for better spacing
    .style("text-anchor", "middle") // Center the text under each tick
    .style("fill", "rgba(255, 255, 255, 0.8)")
    .text(d => {
      // Parse the date string that's in "YYYY-MM-DD" format
      const [year, month, day] = (d as string).split('-').map(Number);
      const date = new Date(year, month - 1, day);
      return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric'
      });
    });

  // Add y-axis with better spacing
  svg.append("g")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y)
      .ticks(5)
      .tickFormat((d: d3.NumberValue) => `${d.valueOf()} kWh`))
    .selectAll("text")
    .attr("x", -10); // Move tick labels further from axis

  // Add y-axis label with better positioning
  svg.append("text")
    .attr("transform", "rotate(-90)")
    .attr("y", margin.left/4) // Moved label even more to the left
    .attr("x", -(height/2))
    .attr("dy", "1em")
    .style("text-anchor", "middle")
    .style("fill", "rgba(255, 255, 255, 0.8)")
    .style("font-size", "14px")
    .text("Energy Yield (kWh)");

  if (props.isLineChart) {
    const timeScale = d3.scaleTime()
      .domain(d3.extent(sortedData, d => d.date) as [Date, Date])
      .range([margin.left, width - margin.right]);

    const line = d3.line<DataPoint>()
      .x(d => timeScale(d.date))
      .y(d => y(d.kwh_yield))
      .curve(d3.curveMonotoneX);

    // Add area under the line
    const area = d3.area<DataPoint>()
      .x(d => timeScale(d.date))
      .y0(height - margin.bottom)
      .y1(d => y(d.kwh_yield))
      .curve(d3.curveMonotoneX);

    // Add the area
    svg.append("path")
      .datum(sortedData)
      .attr("class", "area")
      .attr("d", area);

    // Add the line
    svg.append("path")
      .datum(sortedData)
      .attr("class", "line")
      .attr("fill", "none")
      .attr("d", line);

    // Add points
    svg.selectAll(".data-point")
      .data(sortedData)
      .enter()
      .append("circle")
      .attr("class", "data-point")
      .attr("cx", d => timeScale(d.date))
      .attr("cy", d => y(d.kwh_yield))
      .attr("r", 6)
      .attr("fill", "#90EE90")
      .on("mouseover", (event, d) => showTooltip(event, d))
      .on("mousemove", moveTooltip)
      .on("mouseout", hideTooltip);
  } else {
    // Bar chart
    svg.selectAll(".bar")
      .data(sortedData)
      .enter()
      .append("rect")
      .attr("class", "bar")
      .attr("x", d => x(d3.timeFormat("%Y-%m-%d")(d.date)) || 0)
      .attr("y", height - margin.bottom)
      .attr("width", x.bandwidth())
      .attr("height", 0)
      .attr("fill", "url(#yield-gradient)")
      .attr("rx", 4)
      .on("mouseover", (event, d) => showTooltip(event, d))
      .on("mousemove", moveTooltip)
      .on("mouseout", hideTooltip)
      .transition()
      .duration(800)
      .delay((d, i) => i * 50)
      .attr("y", d => y(d.kwh_yield))
      .attr("height", d => height - margin.bottom - y(d.kwh_yield));
  }
};

const fetchLatestData = async () => {
  try {
    const panelId = 1;
    const currentStartDate = props.startDate || '2024-12-20';
    const currentEndDate = props.endDate || '2024-12-24';
    
    console.log(`[yieldGraph] Fetching data for panel ${panelId} from ${currentStartDate} to ${currentEndDate}`);
    
    const response = await store.user.getPanelYieldData(
      panelId,
      currentStartDate,
      currentEndDate
    );
    const responseData = await response.json();
    
    // Process and deduplicate data
    const yieldData = responseData.yield_data;
    const uniqueData = new Map<string, any>();
    
    // Keep only the latest entry for each date
    yieldData.forEach((item: any) => {
      const dateKey = new Date(item.date).toISOString().split('T')[0];
      uniqueData.set(dateKey, {
        date: new Date(item.date),
        kwh_yield: parseFloat(item.kwh_yield) || 0,
      });
    });
    
    // Convert back to array and sort by date
    const sortedData = Array.from(uniqueData.values())
      .sort((a, b) => a.date.getTime() - b.date.getTime());

    // Filter data for the selected date range
    const filteredData = sortedData.filter(item => {
      const itemDate = item.date.toISOString().split('T')[0];
      return itemDate >= currentStartDate && itemDate <= currentEndDate;
    });

    console.log('[yieldGraph] Processed data points:', filteredData.length);
    
    // Update the data
    solarData.value = filteredData;
    
    // Update the chart
    createChart();
  } catch (error) {
    console.error("[yieldGraph] Error fetching updated data:", error);
  }
};

// Watch for changes in the date range props
watch(() => [props.startDate, props.endDate], ([newStartDate, newEndDate]) => {
  if (newStartDate && newEndDate) {
    console.log(`[yieldGraph] Date range changed to ${newStartDate} - ${newEndDate}`);
    fetchLatestData();
  }
}, { immediate: true });

// Watch for changes in solarData or isLineChart to re-render chart
watch([() => props.solarData, () => props.isLineChart], () => {
  console.log('[yieldGraph] Data or chart type changed, updating chart');
  createChart();
}, { deep: true });

onMounted(() => {
  console.log('[yieldGraph] Component mounted');
  createChart();
});
</script>

<style scoped>
.chart-container {
  width: 900px;
  height: 600px;
  position: relative;
  padding: 32px;
  background: #1a1a1a;
  border-radius: 16px;
  border: 1px solid rgba(144, 238, 144, 0.1);
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.2),
    0 0 0 1px rgba(144, 238, 144, 0.05);
  margin: 100px auto;
  margin-left: 180px;
  transition: all 0.3s ease;
}

/* Make sure the graph container is visible on smaller screens */
@media (max-width: 1400px) {
  .chart-container {
    width: calc(100% - 320px);
    min-width: 600px;
  }
}

.tooltip-container {
  position: absolute;
  background: rgba(26, 26, 26, 0.98);
  border: 1px solid rgba(144, 238, 144, 0.2);
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(144, 238, 144, 0.1),
    inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  pointer-events: none;
  z-index: 1000;
  min-width: 140px;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transform: translateY(-4px);
  transition: all 0.2s ease;
}

.tooltip-content {
  font-size: 14px;
  line-height: 1.6;
  letter-spacing: 0.3px;
}

.tooltip-date {
  color: rgba(144, 238, 144, 0.9);
  margin-bottom: 8px;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tooltip-value {
  color: white;
  font-weight: 600;
  font-size: 16px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* Style the axis lines and text */
:deep(.domain),
:deep(.tick line) {
  stroke: rgba(144, 238, 144, 0.2);
  stroke-width: 1px;
}

:deep(.tick text) {
  fill: rgba(255, 255, 255, 0.8);
  font-size: 12px;
  font-weight: 500;
}

/* Add grid lines */
:deep(.grid-lines line) {
  stroke: rgba(144, 238, 144, 0.05);
  stroke-dasharray: 4,4;
}

/* Style the data points */
:deep(circle) {
  transition: all 0.2s ease;
  stroke: #1a1a1a;
  stroke-width: 2;
}

:deep(circle):hover {
  stroke: rgba(144, 238, 144, 1);
  stroke-width: 3;
  filter: drop-shadow(0 0 8px rgba(144, 238, 144, 0.4));
}

/* Style the line/area */
:deep(path.line) {
  stroke: #90EE90;
  stroke-width: 2.5;
  filter: drop-shadow(0 2px 4px rgba(144, 238, 144, 0.2));
}

:deep(path.area) {
  fill: url(#yield-gradient);
  opacity: 0.2;
}
</style>
