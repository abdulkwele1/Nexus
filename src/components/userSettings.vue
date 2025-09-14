<template>
  <div class="settings-container">
    <!-- Navbar -->
    <nav class="navbar">
      <div class="max-w-screen-xl flex flex-wrap items-center justify-center mx-auto py-2 px-4">
        <div class="flex items-center justify-center w-full">
          <a @click="goHome" class="flex items-center space-x-3 rtl:space-x-reverse cursor-pointer">
            <img alt="Vue logo" class="logo" src="@/assets/nexus.png" width="80" height="32" />
          </a>
          <div class="ml-8">
            <ul class="font-medium flex flex-row space-x-8">
              <li>
                <a @click="goHome" class="text-white hover:text-[#90EE90] cursor-pointer">Home</a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </nav>

    <!-- Settings Content -->
    <div class="settings-content">
      <div class="option" @click="showChangePasswordModal = true">Change password</div>
      <div class="divider"></div>
      <div class="option" @click="addNewSensor">Add New Sensor</div>
      <div class="divider"></div>
      <div class="option" @click="logout">Log Out</div>
    </div>

    <!-- Change Password Modal -->
    <div v-if="showChangePasswordModal" class="modal">
      <div class="modal-content">
        <h2>Change Password</h2>
        <input type="password" v-model="currentPassword" placeholder="Current Password" />
        <input type="password" v-model="newPassword" placeholder="New Password" @input="clearError" />
        <input type="password" v-model="confirmPassword" placeholder="Confirm New Password" @input="clearError" />

        <div class="button-container">
          <button class="submit-button" @click="changePassword" :disabled="!passwordsMatch">Submit</button>
          <button class="cancel-button" @click="showChangePasswordModal = false">Cancel</button>
        </div>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </div>
    </div>

    <!-- Add Sensor Modal -->
    <div v-if="showAddSensorModal" class="modal">
      <div class="modal-content">
        <h2>Add New Sensor</h2>
        <div class="info-message">
          <span class="info-icon">ℹ️</span>
          <span class="info-text">Only sensors through SenseCAP can be added with no optimization</span>
        </div>
        <input type="text" v-model="sensorEUI" placeholder="Device EUI (required)" @input="clearSensorError" />
        <input type="text" v-model="sensorName" placeholder="Sensor Name (optional)" @input="clearSensorError" />
        <input type="text" v-model="sensorLocation" placeholder="Location (optional)" @input="clearSensorError" />

        <div class="button-container">
          <button class="submit-button" @click="addSensor" :disabled="!sensorEUI.trim()">Add Sensor</button>
          <button class="cancel-button" @click="closeAddSensorModal">Cancel</button>
        </div>

        <p v-if="sensorErrorMessage" class="error">{{ sensorErrorMessage }}</p>
        <p v-if="sensorSuccessMessage" class="success">{{ sensorSuccessMessage }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';

import { useNexusStore } from '@/stores/nexus'

const store = useNexusStore()

const {
  VITE_NEXUS_API_URL,
} = import.meta.env;

const router = useRouter();
const showChangePasswordModal = ref(false);
const currentPassword = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const errorMessage = ref('');

// Add sensor modal state
const showAddSensorModal = ref(false);
const sensorEUI = ref('');
const sensorName = ref('');
const sensorLocation = ref('');
const sensorErrorMessage = ref('');
const sensorSuccessMessage = ref('');

// Computed property to check if passwords match
const passwordsMatch = computed(() => newPassword.value === confirmPassword.value);

// Clear the error message when input changes
const clearError = () => {
  if (errorMessage.value) {
    errorMessage.value = '';
  }
};

// Clear sensor error/success messages when input changes
const clearSensorError = () => {
  if (sensorErrorMessage.value) {
    sensorErrorMessage.value = '';
  }
  if (sensorSuccessMessage.value) {
    sensorSuccessMessage.value = '';
  }
};

// Change password functionality
const changePassword = async () => {
  if (!passwordsMatch.value) {
    errorMessage.value = 'Passwords do not match';
    return;
  }

  // Check if new password is the same as current password
  if (newPassword.value === currentPassword.value) {
    errorMessage.value = 'New password cannot be the same as the current password.';
    return;
  }

  // Create and log the request body
  const requestBody = JSON.stringify({
    current_password: currentPassword.value,
    new_password: newPassword.value,
  });
  console.log('Request Body:', requestBody); // Log the request body

  try {
    const response = await fetch(`${VITE_NEXUS_API_URL}/change-password`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: requestBody,
    });

    if (!response.ok) {
      const errorResponse = await response.json();
      console.error('Error response:', errorResponse); // Log the error for better visibility
      errorMessage.value = errorResponse.error || 'Failed to change password';
      return;
    }

    alert('Password changed successfully');
    showChangePasswordModal.value = false;
    currentPassword.value = '';
    newPassword.value = '';
    confirmPassword.value = '';
  } catch (error) {
    console.error('Error:', error);
    alert('An error occurred while changing the password');
  }
};

// Navigate to the home page
const goHome = () => {
  router.push('/home');
};

// Handle adding new sensor
const addNewSensor = () => {
  showAddSensorModal.value = true;
  // Clear any previous messages
  sensorErrorMessage.value = '';
  sensorSuccessMessage.value = '';
};

// Close add sensor modal and reset form
const closeAddSensorModal = () => {
  showAddSensorModal.value = false;
  sensorEUI.value = '';
  sensorName.value = '';
  sensorLocation.value = '';
  sensorErrorMessage.value = '';
  sensorSuccessMessage.value = '';
};

// Add sensor functionality
const addSensor = async () => {
  if (!sensorEUI.value.trim()) {
    sensorErrorMessage.value = 'Device EUI is required';
    return;
  }

  try {
    const result = await store.user.addSensor(
      sensorEUI.value.trim(),
      sensorName.value.trim() || undefined,
      sensorLocation.value.trim() || undefined
    );

    if (result.success) {
      sensorSuccessMessage.value = result.message || 'Sensor added successfully!';
      // Clear form after successful addition
      setTimeout(() => {
        closeAddSensorModal();
        // Optionally refresh the sensors list or navigate to sensors page
        // You could emit an event here to notify parent components
      }, 2000);
    } else {
      sensorErrorMessage.value = result.error || 'Failed to add sensor';
    }
  } catch (error) {
    console.error('Error adding sensor:', error);
    sensorErrorMessage.value = 'An unexpected error occurred';
  }
};

// Handle user logout
const logout = async () => {
  await store.user.logout()
  // Clear the store state
  store.user.loggedIn = false;
  store.user.userName = '';
  
  router.push('/');
};
</script>

<style scoped>
/* Navbar styles */
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1000;
  margin-top: -1px;
  height: 60px;
  background-color: #000000; /* Pure black for navbar */
  border-bottom: 2px solid #90EE90;
  box-shadow: 0 0 10px rgba(144, 238, 144, 0.3);
}

.max-w-screen-xl {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
}

.logo {
  width: 80px;
  height: 32px;
  object-fit: contain;
}

/* Settings container adjustments */
.settings-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 80px 20px 20px;
  box-sizing: border-box;
  background-color: #121212; /* Slightly lighter than black for main background */
}

/* Remove the home button styles since it's now in the navbar */
.home-button {
  display: none;
}

/* Settings content adjustments */
.settings-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 600px;
  background: #1a1a1a; /* Even lighter for the content area */
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
}

.option {
  padding: 20px;
  font-size: 18px;
  width: 100%;
  text-align: center;
  cursor: pointer;
  transition: background-color 0.3s;
  border-radius: 4px;
  color: white;
}

.option:hover {
  background-color: #242424; /* Slightly lighter on hover */
}

.divider {
  width: 100%;
  height: 2px;
  background-color: #90EE90; /* Light green color */
  margin: 10px 0;
  box-shadow: 0 0 8px rgba(144, 238, 144, 0.5); /* Subtle glow effect */
}

/* Modal styling */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.modal-content {
  background: #1a1a1a;
  padding: 30px;
  border-radius: 8px;
  width: 300px;
  box-shadow: 0 4px 20px rgba(144, 238, 144, 0.1);
  border: 1px solid rgba(144, 238, 144, 0.2);
}

.modal-content h2 {
  margin-bottom: 20px;
  color: white;
  font-size: 24px;
}

.modal-content input {
  width: 100%;
  padding: 10px;
  margin-bottom: 15px;
  border: 1px solid #333;
  border-radius: 4px;
  font-size: 16px;
  background: #2a2a2a;
  color: white;
}

.modal-content input:focus {
  outline: none;
  border-color: #90EE90;
  box-shadow: 0 0 5px rgba(144, 238, 144, 0.3);
}

.button-container {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
}

.submit-button, .cancel-button {
  padding: 10px 20px;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.submit-button {
  background-color: #90EE90;
  color: black;
  border: none;
}

.submit-button:hover:not(:disabled) {
  background-color: #7CCD7C;
  box-shadow: 0 0 10px rgba(144, 238, 144, 0.3);
}

.submit-button:disabled {
  background-color: #444;
  color: #666;
  cursor: not-allowed;
}

.cancel-button {
  background-color: transparent;
  border: 1px solid #90EE90;
  color: #90EE90;
}

.cancel-button:hover {
  background-color: rgba(144, 238, 144, 0.1);
  box-shadow: 0 0 10px rgba(144, 238, 144, 0.2);
}

.error {
  color: #ff6b6b;
  font-size: 14px;
  margin-top: 15px;
  text-align: center;
}

.success {
  color: #90EE90;
  font-size: 14px;
  margin-top: 15px;
  text-align: center;
}

.info-message {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(144, 238, 144, 0.1);
  border: 1px solid rgba(144, 238, 144, 0.3);
  border-radius: 6px;
  padding: 12px;
  margin: 15px 0;
  font-size: 14px;
}

.info-icon {
  font-size: 16px;
  color: #90EE90;
  flex-shrink: 0;
}

.info-text {
  color: #e0e0e0;
  line-height: 1.4;
}

/* Navigation text color */
.navbar a {
  color: white;
}

.navbar a:hover {
  color: #90EE90;
}
</style>
