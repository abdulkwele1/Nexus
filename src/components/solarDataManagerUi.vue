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

      <button type="submit" :class="{'add-btn': !isEditing, 'edit-btn': isEditing}">Add Solar Data</button>
    </form>

    <!-- Button for removing solar data -->
    <button class="remove-btn" @click="removeSolarData">
      Remove Solar Data
    </button>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: "SolarDataManagerUi",
  props: {
    // Accepts a graphType prop which must be either 'yield' or 'consumption'
    graphType: {
      type: String,
      required: true,
      validator: (value) => ['yield', 'consumption'].includes(value)
    }
  },
  data() {
    return {
      // Initialize with fields for both graph types.
      solarData: {
        date: '',
        // For yield graph:
        kwhProduced: null,
        // For consumption graph:
        totalStored: null,
        kwhUsed: null,
      },
    };
  },
  methods: {
    async addSolarData() {
      try {
        let endpoint = '';
        let payload = {};

        if (this.graphType === 'yield') {
          // Define endpoint and payload for yield data.
          endpoint = '/api/solarData/yield';
          payload = {
            date: this.solarData.date,
            kwhProduced: this.solarData.kwhProduced,
          };
        } else if (this.graphType === 'consumption') {
          // Define endpoint and payload for consumption data.
          endpoint = '/api/solarData/consumption';
          payload = {
            date: this.solarData.date,
            totalStored: this.solarData.totalStored,
            kwhUsed: this.solarData.kwhUsed,
          };
        }

        // Adjust the endpoint and payload as needed for your backend API.
        const response = await axios.post(endpoint, payload);
        console.log('Added solar data:', response.data);
        this.resetForm();
      } catch (error) {
        console.error('Error adding solar data:', error);
      }
    },
    async removeSolarData() {
      try {
        let endpoint = '';
        if (this.graphType === 'yield') {
          endpoint = '/api/solarData/yield/last';
        } else if (this.graphType === 'consumption') {
          endpoint = '/api/solarData/consumption/last';
        }

        // Adjust the logic and endpoint as needed.
        const response = await axios.delete(endpoint);
        console.log('Removed solar data:', response.data);
      } catch (error) {
        console.error('Error removing solar data:', error);
      }
    },
    resetForm() {
      // Reset the common field.
      this.solarData.date = '';
      if (this.graphType === 'yield') {
        this.solarData.kwhProduced = null;
      } else if (this.graphType === 'consumption') {
        this.solarData.totalStored = null;
        this.solarData.kwhUsed = null;
      }
    },
  },
};
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
  background-color: #0056b3
}
</style>
