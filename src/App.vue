<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import navBar from './components/navBar.vue'; // This includes your Navbar
import { initializeSessionManager, destroySessionManager } from './services/sessionManager';

// Get the current route
const route = useRoute();

// Determine if the current route is /settings
const isSettingsPage = computed(() => route.path === '/settings');

// Initialize session manager when app mounts
onMounted(() => {
  initializeSessionManager();
});

// Clean up session manager when app unmounts
onUnmounted(() => {
  destroySessionManager();
});
</script>

<template>
  <div class="app-container">
    <!-- Conditionally render navBar based on the route path -->
    <navBar v-if="!isSettingsPage" msg="Second time!" />
    <RouterView />
  </div>
</template>

<style>
html, body {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  min-height: 100vh;
  width: 100%;
  overflow-x: hidden;
}

#app, .app-container {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  width: 100%;
}

header {
  margin: 0;
  padding: 0;
  line-height: 1.5;
}

.wrapper {
  margin: 0;
  padding: 0;
}
</style>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

nav {
  width: 100%;
  font-size: 18px;
  text-align: center;
  margin-top: 2rem;
}

nav a.router-link-exact-active {
  color: var(--color-text);
}

nav a.router-link-exact-active:hover {
  background-color: transparent;
}

nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 1px solid var(--color-border);
}

nav a:first-of-type {
  border: 0;
}

@media (min-width: 1024px) {
  header {
    display: center;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
    display: flex;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }

  body {
    background-color: #f0f0f0; /* Set the entire page background color to light grey */
  }
}
</style>
