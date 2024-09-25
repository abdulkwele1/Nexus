<template>
  <div class="settings-container">
    <button class="home-button" @click="goHome">Home</button>
    <div class="settings-content">
      <div class="option" @click="optionOne">
        Change email
      </div>
      <div class="divider"></div>
      <div class="option" @click="optionTwo">
        Change password
      </div>
      <div class="divider"></div>
      <div class="option" @click="logout">
        Log Out
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

// Function to delete a cookie by name
const deleteCookie = (name: string) => {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
};

// Navigate to the home page
const goHome = () => {
  router.push('/home'); // Adjust the path as necessary
};

// Placeholder for changing email
const optionOne = () => {
  // Implement change email functionality
};

// Placeholder for changing password
const optionTwo = () => {
  // Implement change password functionality
};

// Handle user logout
const logout = () => {
  // Clear the session cookie
  deleteCookie('session_id'); // Call deleteCookie to remove the cookie
  
  // Redirect to the login page
  router.push('/'); // This redirects to the login page
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
</style>
