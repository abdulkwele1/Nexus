<script setup>
import SignInPage from '../components/SignInPage.vue'
import { onMounted } from 'vue'

// run every time page (component) loads
// https://learnvue.co/articles/vue-lifecycle-hooks-guide
onMounted(() => {
    console.log('hommie')
    // If user is on login page and document.cookie == “” is false (is not empty), redirect them to the home page
})

</script>

<template>
  <main>
    <SignInPage />
  </main>
</template>
