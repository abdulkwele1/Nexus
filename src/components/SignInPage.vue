<template>
  <div class="flex-container">
    <input v-model="username" type="text" placeholder="Enter your username" />
    <input v-model="password" type="password" placeholder="Enter your password" />
    <button @click="handleLogin">Login</button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useNexusStore } from '@/stores/nexus';

const store = useNexusStore();

// Create a reactive reference for the username
const username = ref('');
const password = ref('');
const router = useRouter();

onMounted(() => {
  const loggedInUser = store.user.getCookie(); // Check session_id cookie
  if (loggedInUser) {
    router.push({ path: '/home' });
  }
});

// Handle the login action
async function handleLogin() {
  if (username.value.trim() && password.value.trim()) {
    try {
      const response = await store.user.login(username.value.trim(),password.value.trim() )

      const responseData = await response.json();

      if (response.ok && responseData.match) {
        // Set session cookie 
        document.cookie = `session_id=${responseData.cookie}; path=/; samesite=strict`; // Use session_id
        // update app state for current user
        store.user.userName = username
        store.user.loggedIn = true
        router.push({ path: '/home' });
      } else {
        alert('Invalid username or password');
      }
    } catch (error) {
      console.error('An error occurred during login:', error);
      alert('Login failed. Please try again.');
    }
  } else {
    alert('Please enter a valid username and password.');
  }
}
</script>

<style scoped>
.flex-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #ffffff;
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
