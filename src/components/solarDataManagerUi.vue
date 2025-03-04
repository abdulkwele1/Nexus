<template>
  <div class="solar-data-manager">
    <h2>Manage Solar Data</h2>
    <!-- Form for adding solar data -->
    <form @submit.prevent="addSolarData">
      <!-- Common Field: Date -->
      <div class="form-group">
        <label for="date">Date:</label>
        <input type="date" id="date" v-model="formData.date" required />
      </div>

      <!-- Fields for Yield Graph -->
      <div v-if="graphType === 'yield'">
        <div class="form-group">
          <label for="kwhProduced">kWh Produced:</label>
          <input
            type="number"
            id="kwhProduced"
            v-model.number="formData.kwhProduced"
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
            v-model.number="formData.totalStored"
            required
          />
        </div>
        <div class="form-group">
          <label for="kwhUsed">kWh Used:</label>
          <input
            type="number"
            id="kwhUsed"
            v-model.number="formData.kwhUsed"
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
import { ref } from 'vue';
import moment from 'moment';
import { useNexusStore } from '../stores/nexus';

// Add defineEmits import
const emit = defineEmits(['dataAdded']);

const nexusStore = useNexusStore();
const { VITE_NEXUS_API_URL } = import.meta.env;

// Define props the same way as in your first code.
const props = defineProps({
  solarData: {
    type: Array,
    required: true,
  },
  graphType: {
    type: String,
    required: true,
  },
});

// Create a local reactive object for the form data so it doesn't conflict with the passed-in solarData.
const formData = ref({
  date: '',
  kwhProduced: 0,
  totalStored: 0,
  kwhUsed: 0,
});

const resetForm = () => {
  formData.value = {
    date: '',
    kwhProduced: 0,
    totalStored: 0,
    kwhUsed: 0,
  };
};

const addSolarData = async () => {
  try {
    const panelId = 1;
    const formattedDate = moment(new Date(formData.value.date)).format("YYYY-MM-DDTHH:mm:ssZ");
    let response;

    if (props.graphType === 'yield') {
      const yield_data = [{
        date: formattedDate,
        kwh_yield: formData.value.kwhProduced,
      }];
      response = await nexusStore.user.setPanelYieldData(1, formattedDate, formattedDate, yield_data);
    } else if (props.graphType === 'consumption') {
      const consumption_data = [{
        date: formattedDate,
        capacity_kwh: formData.value.totalStored,
        consumed_kwh: formData.value.kwhUsed,
      }];
      response = await nexusStore.user.setPanelConsumptionData(panelId, formattedDate, formattedDate, consumption_data);
    }

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    console.log('Added solar data:', data);

    // Emit event with the type of data that was added
    emit('dataAdded', props.graphType);
    
    resetForm();
  } catch (error) {
    console.error('Error adding solar data:', error);
  }
};


const removeSolarData = () => {
  // Implement your removal logic here
  console.log('Remove solar data clicked');
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
  background-color: #0056b3;
}
</style>
