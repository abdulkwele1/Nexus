<template>
  <div class="flex-container">
    <input v-model="username" type="text" placeholder="Enter your username" />
    <input v-model="password" type="text" placeholder="Enter your password" />
    <button @click="handleLogin">Login</button>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'

const {
  VITE_NEXUS_API_URL,
} = import.meta.env;

// Create a reactive reference for the username
const username = ref('');
const password = ref('');

const router = useRouter(); // Initialize router

// run every time page (component) loads
// https://learnvue.co/articles/vue-lifecycle-hooks-guide
onMounted(() => {
    console.log('mounted in the composition api!')
    // If user is on login page and document.cookie == “” is false (is not empty), redirect them to the home page
})

// Handle the login action
async function handleLogin () {
  // TODO refactor
  // Check if the username input is not empty after trimming whitespace
  if (username.value.trim()) {
    try {
      // Construct the URL with the username
      const url = `${VITE_NEXUS_API_URL}/login`;

      let data = {
        "password":  password.value.trim(),
        "username":  username.value.trim()
      }

      let response = await fetch(url, {
        headers: {
                'Content-Type': 'application/json'
            },
        method: 'POST',
        body: JSON.stringify(data)})

    let responseData = await response.json()
   
  if (response.ok && responseData.match) {
    // set user cookie
     // default expire time: 1 day
    document.cookie = (username.value.trim() + ":" + responseData.cookie)
    // Redirect to /home on successful login
    router.push({path: '/home'});
  } else {
    // Handle invalid login
    alert('Invalid username or password');
  }

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
  background-color: ##FFFFFF; /* Optional: Background color */
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
