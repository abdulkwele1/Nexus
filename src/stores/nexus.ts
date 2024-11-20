import { defineStore } from 'pinia'
const {
    VITE_NEXUS_API_URL,
  } = import.meta.env;

export const useNexusStore = defineStore('nexus', {
  state: () => {
    return {
        user: new User(VITE_NEXUS_API_URL),
    }
  }
})



class User {
    baseURL: string;
    userName: string;
    loggedIn: boolean;

    constructor(baseURL: any) {
        this.baseURL = baseURL
        this.userName = ""
        this.loggedIn = false
    }

    async login(userName: any, password: any): Promise<any> {
        const loginAPIUrl = `${VITE_NEXUS_API_URL}/login`
        const data = {
            username: userName,
            password: password,
          };

          const response = await fetch(loginAPIUrl, {
            headers: {
              'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify(data),
          });

          return response
    }

    async logout() {
        const response = await fetch(`${this.baseURL}/logout`, {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            }
          });

          if (!response.ok) {
            const errorResponse = await response.json();
            console.error('Error response:', errorResponse); // Log the error for better visibility

            return false;
          }

          this.loggedIn = false

          return true
    }

    // Function to get a cookie by name
    getCookie(): any {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; session_id=`);
        if (parts.length === 2) {
            return parts?.pop()?.split(';')?.shift();
        }
        return false
  }
}
