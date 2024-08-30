<template>
  <div class="flex-container">
    <input v-model="username" type="text" placeholder="Enter your username" />
    <input v-model="password" type="text" placeholder="Enter your password" />
    <button @click="handleLogin">Login</button>
  </div>
</template>

<script setup>
import { ref } from 'vue';

// Create a reactive reference for the username
const username = ref('');
const password = ref('');

// Handle the login action
async function handleLogin () {
  // Check if the username input is not empty after trimming whitespace
  if (username.value.trim()) {
    try {
      // Construct the URL with the username
      const url = `http://localhost:8080/login`;

      let data = {
        "password":  username.value.trim(),
        "username":  password.value.trim()
      }

      let response = await fetch(url, {
        headers: {
                'Content-Type': 'application/json'
            },
        method: 'POST',
        body: JSON.stringify(data)})

    let responseData = await response.json()
      alert(`api says ${JSON.stringify(responseData)}`)

    } catch (error) {
      console.error('An error occurred during redirection:', error);
      alert('Failed to redirect.');
    }
  } else {
    alert('Please enter a valid username.');
  }
};
</script>

<style scoped>
.flex-container {
  display: flex;
  flex-direction: column; /* Align items vertically */
  justify-content: center; /* Center vertically */
  align-items: center; /* Center horizontally */
  height: 100vh; /* Full viewport height */
  background-color: #f0f0f0; /* Optional: Background color */
}

input {
  margin-bottom: 10px;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  background-color: #007bff;
  color: white;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}
</style>
