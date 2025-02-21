<!-- src/components/solarDataManagerUi.vue -->
<template>
  <div class="solar-data-manager">
    <h2>Manage Solar Data</h2>
    <!-- Form for adding solar data -->
    <form @submit.prevent="addSolarData">
      <!-- Common Field: Date -->
      <div class="form-group">
        <label for="date">Date:</label>
        <input type="date" id="date" v-model="solarData.date" required />
      </div>

      <!-- Fields for Yield Graph -->
      <div v-if="graphType === 'yield'">
        <div class="form-group">
          <label for="kwhProduced">kWh Produced:</label>
          <input
            type="number"
            id="kwhProduced"
            v-model.number="solarData.kwhProduced"
            required
          />
        </div>
      </div>

      <!-- Fields for Consumption Graph -->
      <div v-if="graphType === 'consumption'">
        <div class="form-group">
          <label for="totalStored">Total kWh Stored:</label>
          <input
            type="number"
            id="totalStored"
            v-model.number="solarData.totalStored"
            required
          />
        </div>
        <div class="form-group">
          <label for="kwhUsed">kWh Used:</label>
          <input
            type="number"
            id="kwhUsed"
            v-model.number="solarData.kwhUsed"
            required
          />
        </div>
      </div>

      <button type="submit" class="add-btn">
        Add Solar Data
      </button>
    </form>

    <!-- Button for removing solar data -->
    <button class="remove-btn" @click="removeSolarData">
      Remove Solar Data
    </button>
  </div>
</template>

<script setup lang="ts">
const { VITE_NEXUS_API_URL } = import.meta.env;


import { defineProps, ref } from 'vue';
import moment from 'moment';

const props = defineProps({
  solarData: {
    type: Array,
    required: true,
  },

  graphType:{
    type: String,
    required: true,
  },
});

const addSolarData = (async() =>  {
    try {
      let endpoint = '';
      let payload = {};

      if (props.graphType === 'yield') {
        // Adjust the URL to match your backend endpoint.  
        // Here the panel_id is hard-coded as 1, change as needed.
        endpoint = `${VITE_NEXUS_API_URL}/panels/1/yield_data`;
        payload = {
          yield_data: [{
            date: moment(new Date(props.solarData.date)).format("YYYY-MM-DDTHH:mm:ssZ"),
            kwh_yield: props.solarData.kwhProduced,
          }],
        };
      } else if (this.graphType === 'consumption') {
        endpoint = `${VITE_NEXUS_API_URL}/panels/1/consumption_data`;
        payload = {
          consumption_data: [{
            date: props.solarData.date,
            capacity_kwh: props.solarData.totalStored,
            consumed_kwh: props.solarData.kwhUsed,
          }],
        };
      }

      const response = await fetch(endpoint, {
        method: 'POST',
          credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log('Added solar data:', data);
      this.resetForm();
    } catch (error) {
      console.error('Error adding solar data:', error);
    }
  });

</script>

<style scoped>
.solar-data-manager {
  padding: 1rem;
  background-color: #f5f5f5;
  border: 1px solid #ccc;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 0.8rem;
}

.remove-btn {
  margin-top: 1rem;
  background-color: #007bff;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  cursor: pointer;
}

.remove-btn:hover {
  background-color: #0056b3;
}

.add-btn {
  margin-top: 1rem;
  background-color: #007bff;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  cursor: pointer;
}

.add-btn:hover {
  background-color: #0056b3;
}
</style>
