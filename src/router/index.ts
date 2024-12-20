import { createRouter, createWebHistory } from 'vue-router';
import HomeView from '../views/HomeView.vue';
import Home from '../components/Home.vue';
import userSettings from '../components/userSettings.vue';
import SolarPage from '../components/SolarPage.vue';
import Sensors from '../components/sensors.vue';
import Drone from '../components/drone.vue'
import { useNexusStore } from '@/stores/nexus'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    
  {
      path: '/drone',
      name: 'drone',
      component: Drone,
      meta: { requiresAuth: true }
    },

    {
      path: '/sensors',
      name: 'sensors',
      component: Sensors,
      meta: { requiresAuth: true }
    },

    {
      path: '/solar',
      name: 'solar',
      component: SolarPage, // Solar page
      meta: { requiresAuth: true } // Protect this route
    },

    {
      path: '/',
      name: 'home',
      component: HomeView // Login page
    },
    {
      path: '/home',
      name: 'homepage',
      component: Home,
      meta: { requiresAuth: true } // Protect this route
    },

    {
      path: '/settings',
      name: 'userSettings',
      component: userSettings,
      meta: { requiresAuth: true } // Protect this route
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue') // Lazy-loaded route
    }
  ]
});

// Navigation guard
router.beforeEach((to, from, next) => {
  const store = useNexusStore()

  const loggedInUser = store.user.getCookie(); // Check for session_id cookie
  if (to.matched.some(record => record.meta.requiresAuth) && !loggedInUser) {
    next('/'); // Redirect to login if not logged in
  } else {
    next(); // Allow access
  }
});

export default router;
