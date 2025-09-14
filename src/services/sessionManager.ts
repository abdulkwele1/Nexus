import { useNexusStore } from '@/stores/nexus';

class SessionManager {
  private refreshInterval: number = 60 * 60 * 1000; // 1 hour
  private lastActivity: number = Date.now();
  private refreshTimer: NodeJS.Timeout | null = null;
  private store: any;

  constructor() {
    this.store = useNexusStore();
    this.setupActivityListeners();
    this.startRefreshTimer();
  }

  private setupActivityListeners() {
    // Listen for user activity events
    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click'];
    
    events.forEach(event => {
      document.addEventListener(event, () => {
        this.lastActivity = Date.now();
      }, true);
    });
  }

  private startRefreshTimer() {
    // Clear existing timer
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
    }

    // Set up new timer
    this.refreshTimer = setInterval(async () => {
      await this.checkAndRefreshSession();
    }, this.refreshInterval);
  }

  private async checkAndRefreshSession() {
    // Only refresh if user has been active in the last 2 hours
    const timeSinceActivity = Date.now() - this.lastActivity;
    const maxInactivity = 2 * 60 * 60 * 1000; // 2 hours

    if (timeSinceActivity > maxInactivity) {
      console.log('User inactive for 2+ hours, skipping session refresh');
      return;
    }

    try {
      console.log('Refreshing session...');
      const response = await this.store.user.refreshSession();
      
      if (response.ok) {
        const data = await response.json();
        console.log('Session refreshed successfully:', data);
      } else {
        console.warn('Session refresh failed:', response.status);
      }
    } catch (error) {
      console.error('Error refreshing session:', error);
    }
  }

  public destroy() {
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
      this.refreshTimer = null;
    }
  }
}

// Create a singleton instance
let sessionManager: SessionManager | null = null;

export function initializeSessionManager() {
  if (!sessionManager) {
    sessionManager = new SessionManager();
  }
  return sessionManager;
}

export function destroySessionManager() {
  if (sessionManager) {
    sessionManager.destroy();
    sessionManager = null;
  }
}
