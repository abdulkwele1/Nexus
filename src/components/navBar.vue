<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

const isHomePage = computed(() => route.path === '/home');
const isSolarPage = computed(() => route.path === '/solar');
const issensors = computed(() => route.path === '/sensors');

const navStyle = computed(() => ({
  backgroundColor: isHomePage.value ? '#000000' : 'transparent',
  borderColor: isHomePage.value ? '#000000' : 'transparent'
}));
</script>

<template>
  <nav v-if="!isSolarPage && !issensors" :style="navStyle" class="navbar">
    <div class="max-w-screen-xl flex flex-wrap items-center justify-between mx-auto py-2 px-4" :class="{ 'text-white': isHomePage }">
      <a href="/" class="flex items-center space-x-3 rtl:space-x-reverse">
        <img alt="Vue logo" class="logo" src="@/assets/nexus.png" width="80" height="32" />
      </a>
      <button
        data-collapse-toggle="navbar-default"
        type="button"
        class="inline-flex items-center p-2 w-10 h-10 justify-center text-sm rounded-lg md:hidden hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200"
        :class="{ 'text-white hover:bg-gray-700 focus:ring-gray-600': isHomePage }"
        aria-controls="navbar-default"
        aria-expanded="false"
      >
        <span class="sr-only">Open main menu</span>
        <svg
          class="w-5 h-5"
          aria-hidden="true"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 17 14"
        >
          <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 1h15M1 7h15M1 13h15"/>
        </svg>
      </button>
      <div class="hidden w-full md:block md:w-auto" id="navbar-default">
        <ul class="font-medium flex flex-col p-4 md:p-0 mt-4 border md:flex-row md:space-x-8 rtl:space-x-reverse md:mt-0 md:border-0">
          <li>
            <a href="#" class="block py-1 px-2 rounded hover:bg-gray-100 md:hover:bg-transparent md:border-0 md:hover:text-blue-200 md:p-0" :class="{ 'text-white': isHomePage }">About</a>
          </li>
          <li>
            <a href="#" class="block py-1 px-2 rounded hover:bg-gray-100 md:hover:bg-transparent md:border-0 md:hover:text-blue-200 md:p-0" :class="{ 'text-white': isHomePage }">Learn more</a>
          </li>
          <li v-if="isHomePage">
            <RouterLink to="/settings" class="flex items-center">
              <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                <path d="M24 13.616v-3.232c-1.651-.587-2.694-.752-3.219-2.019v-.001c-.527-1.271.1-2.134.847-3.707l-2.285-2.285c-1.561.742-2.433 1.375-3.707.847h-.001c-1.269-.526-1.435-1.576-2.019-3.219h-3.232c-.582 1.635-.749 2.692-2.019 3.219h-.001c-1.271.528-2.132-.098-3.707-.847l-2.285 2.285c.745 1.568 1.375 2.434.847 3.707-.527 1.271-1.584 1.438-3.219 2.02v3.232c1.632.58 2.692.749 3.219 2.019.53 1.282-.114 2.166-.847 3.707l2.285 2.286c1.562-.743 2.434-1.375 3.707-.847h.001c1.27.526 1.436 1.579 2.019 3.219h3.232c.582-1.636.75-2.69 2.027-3.222h.001c1.262-.524 2.12.101 3.698.851l2.285-2.286c-.744-1.563-1.375-2.433-.848-3.706.527-1.271 1.588-1.44 3.221-2.021zm-12 2.384c-2.209 0-4-1.791-4-4s1.791-4 4-4 4 1.791 4 4-1.791 4-4 4z"/>
              </svg>
            </RouterLink>
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1000;
  margin-top: -1px;
  height: 60px; /* Fixed height for navbar */
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

.body {
  margin: 0;
  padding: 0;
}
</style>