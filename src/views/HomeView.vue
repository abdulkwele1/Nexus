<script setup>
import SignInPage from '../components/SignInPage.vue'
import { onMounted } from 'vue'

// run every time page (component) loads
// https://learnvue.co/articles/vue-lifecycle-hooks-guide
onMounted(() => {
    console.log('hommie')
    // If user is on login page and document.cookie == "” is false (is not empty), redirect them to the home page
})

</script>

<template>
  <nav class="login-navbar">
    <div class="nav-content">
      <img src="../assets/nexus.png" alt="Nexus Logo" class="nav-logo" />
    </div>
  </nav>
  <main>
    <SignInPage />
  </main>
</template>

<style scoped>
.login-navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background-color: black;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.nav-content {
  width: 100%;
  max-width: 1200px;
  padding: 0 20px;
  display: flex;
  align-items: center;
}

.nav-logo {
  height: 32px;
  width: auto;
  object-fit: contain;
}

main {
  padding-top: 60px; /* Add padding to account for fixed navbar */
}
</style>
