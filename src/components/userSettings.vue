<template>
  <div class="settings-container">
    <button class="home-button" @click="goHome">Home</button>
    <div class="settings-content">
      <div class="divider"></div>
      <div class="option" @click="showChangePasswordModal = true">Change password</div>
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

// Computed property to check if passwords match
const passwordsMatch = computed(() => newPassword.value === confirmPassword.value);

// Clear the error message when input changes
const clearError = () => {
  if (errorMessage.value) {
    errorMessage.value = '';
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

// Handle user logout
const logout = async () => {
  await store.user.logout()
  document.cookie = 'session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';

  router.push('/');
};
</script>

<style scoped>
/* Container for settings page */
.settings-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh; /* Full viewport height */
  padding: 20px;
  box-sizing: border-box;
}

/* Style for the home button */
.home-button {
  position: absolute;
  top: 20px;
  left: 20px;
  padding: 10px 20px;
  font-size: 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.home-button:hover {
  background-color: #0056b3;
}

/* Container for settings options */
.settings-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 600px; /* Optional: limit the width */
}

/* Style for each option */
.option {
  padding: 20px;
  font-size: 18px;
  width: 100%;
  text-align: center;
  cursor: pointer;
  transition: background-color 0.3s;
  background-color: #D3D3D3;
  border-radius: 15px;
}

.option:hover {
  background-color: #f0f0f0;
}

/* Divider between options */
.divider {
  width: 100%;
  height: 1px;
  background-color: #ddd;
  margin: 10px 0;
}

/* Modal styling */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 300px;
}

/* Button container styling */
.button-container {
  display: flex;
  justify-content: space-between; /* Align buttons to opposite sides */
  margin-top: 20px;
}

/* Error message styling */
.error {
  color: red;
  font-size: 14px;
  margin-top: 10px;
}


</style>  
