import { useNexusStore } from '@/stores/nexus';
import { useRouter } from 'vue-router';

interface FetchOptions extends RequestInit {
  skipAuthCheck?: boolean;
}

export async function apiFetch(url: string, options: FetchOptions = {}): Promise<Response> {
  try {
    const response = await fetch(url, {
      ...options,
      credentials: 'include', // Always include cookies
    });

    // Skip auth check if specified (e.g. for login/logout endpoints)
    if (options.skipAuthCheck) {
      return response;
    }

    // Handle 401 Unauthorized responses
    if (response.status === 401) {
      const store = useNexusStore();
      const router = useRouter();
      
      // Clear user state
      store.user.loggedIn = false;
      store.user.userName = '';
      
      // Clear the cookie
      document.cookie = 'session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
      
      // Redirect to login page
      router.push('/');
      
      throw new Error('Session expired. Please log in again.');
    }

    return response;
  } catch (error) {
    // Re-throw the error to be handled by the calling code
    throw error;
  }
}