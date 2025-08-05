<template>
  <div class="battery-data-manager">
    <h2>Manage Battery Data</h2>
    <!-- Form for adding battery data -->
    <form @submit.prevent="addBatteryData">
      <!-- Sensor Selection -->
      <div class="form-group">
        <label for="sensorId">Sensor:</label>
        <select id="sensorId" v-model="formData.sensorId" required>
          <option value="">Select a sensor</option>
          <option v-for="(sensor, id) in sensors" :key="id" :value="sensor.id">
            {{ sensor.name }}
          </option>
        </select>
      </div>

      <!-- Date Field -->
      <div class="form-group">
        <label for="date">Date:</label>
        <input type="date" id="date" v-model="formData.date" required />
      </div>

      <!-- Battery Level Field -->
      <div class="form-group">
        <label for="batteryLevel">Battery Level (%):</label>
        <input
          type="number"
          id="batteryLevel"
          v-model.number="formData.batteryLevel"
          min="0"
          max="100"
          step="0.1"
          required
        />
      </div>

      <!-- Voltage Field -->
      <div class="form-group">
        <label for="voltage">Voltage (V):</label>
        <input
          type="number"
          id="voltage"
          v-model.number="formData.voltage"
          min="0"
          max="5"
          step="0.1"
          required
        />
      </div>

      <button type="submit" class="add-btn">
        Add Battery Data
      </button>
    </form>

    <!-- Button for removing battery data -->
    <button class="remove-btn" @click="removeBatteryData">
      Remove Battery Data
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useNexusStore } from '../stores/nexus';

// Define props
const props = defineProps({
  sensors: {
    type: Object,
    required: true,
  },
});

// Define emits
const emit = defineEmits(['dataAdded']);

const nexusStore = useNexusStore();

// Create a local reactive object for the form data
const formData = ref({
  sensorId: '',
  date: '',
  batteryLevel: 0,
  voltage: 0,
});

const resetForm = () => {
  formData.value = {
    sensorId: '',
    date: '',
    batteryLevel: 0,
    voltage: 0,
  };
};

const addBatteryData = async () => {
  try {
    // Set the time to noon UTC to avoid timezone issues
    const date = new Date(formData.value.date);
    date.setUTCHours(12, 0, 0, 0);
    const formattedDate = date.toISOString();

    const batteryData = [{
      date: formattedDate,
      battery_level: formData.value.batteryLevel,
      voltage: formData.value.voltage,
    }];

    const response = await nexusStore.user.setSensorBatteryData(
      formData.value.sensorId,
      batteryData
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    console.log('Added battery data:', data);

    // Emit event to notify parent component
    emit('dataAdded', formData.value.sensorId);
    
    resetForm();
  } catch (error) {
    console.error('Error adding battery data:', error);
  }
};

const removeBatteryData = () => {
  // Implement your removal logic here
  console.log('Remove battery data clicked');
};
</script>

<style scoped>
.battery-data-manager {
  padding: 1rem;
  background-color: #f5f5f5;
  border: 1px solid #ccc;
  margin-bottom: 1rem;
  border-radius: 10px;
}

.battery-data-manager h2 {
  font-weight: bold;
}

.form-group {
  margin-bottom: 0.8rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.3rem;
  font-weight: 500;
  color: #0b0c36;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
}

.remove-btn {
  margin-top: 1rem;
  background-color: #007bff;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  cursor: pointer;
  border-radius: 5px;
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
  border-radius: 5px;
}

.add-btn:hover {
  background-color: #0056b3;
}
</style>
  