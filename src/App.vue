<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import LandingPage from './components/LandingPage.vue'; // This includes your Navbar

// Get the current route
const route = useRoute();

// Determine if the current route is /settings
const isSettingsPage = computed(() => route.path === '/settings');
</script>

<template>
  <header>
    <div class="wrapper">
      <!-- Conditionally render LandingPage based on the route path -->
      <LandingPage v-if="!isSettingsPage" msg="Second time!" />
    </div>
  </header>

  <RouterView />
</template>

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
  width: 1000%;
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
