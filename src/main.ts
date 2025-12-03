import './index.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue';
import router from './router'

// Add global error handler for unhandled errors
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error);
});

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason);
});

const app = createApp(App)

app.use(createPinia())
app.use(router)

// Add error handler to Vue app
app.config.errorHandler = (err, instance, info) => {
  console.error('Vue error:', err, info);
};

try {
  app.mount('#app')
} catch (error) {
  console.error('Failed to mount app:', error);
}
