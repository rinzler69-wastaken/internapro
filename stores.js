import { writable, derived } from 'svelte/store';
import { api } from './api';

// Auth store
function createAuthStore() {
  const { subscribe, set, update } = writable({
    user: null,
    loading: true,
    isAuthenticated: false,
  });

  return {
    subscribe,
    login: async (email, password, totpCode = null) => {
      const response = await api.login(email, password, totpCode);
      if (response.data.user) {
        set({
          user: response.data.user,
          loading: false,
          isAuthenticated: true,
        });
      }
      return response;
    },
    logout: async () => {
      await api.logout();
      set({
        user: null,
        loading: false,
        isAuthenticated: false,
      });
    },
    checkAuth: async () => {
      try {
        const response = await api.getCurrentUser();
        set({
          user: response.data,
          loading: false,
          isAuthenticated: true,
        });
      } catch (error) {
        // Don't throw error, just set unauthenticated state
        console.log('Not authenticated');
        set({
          user: null,
          loading: false,
          isAuthenticated: false,
        });
      }
    },
    setUser: (user) => {
      set({
        user,
        loading: false,
        isAuthenticated: !!user,
      });
    },
  };
}

export const auth = createAuthStore();

// Attendance store
export const todayAttendance = writable(null);
export const attendanceLoading = writable(false);

// Toast notifications store
function createToastStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,
    add: (message, type = 'info', duration = 3000) => {
      const id = Date.now();
      update(toasts => [...toasts, { id, message, type }]);
      
      if (duration > 0) {
        setTimeout(() => {
          update(toasts => toasts.filter(t => t.id !== id));
        }, duration);
      }
      return id;
    },
    remove: (id) => {
      update(toasts => toasts.filter(t => t.id !== id));
    },
  };
}

export const toast = createToastStore();

// Convenience methods for toast
toast.success = (message, duration) => toast.add(message, 'success', duration);
toast.error = (message, duration) => toast.add(message, 'error', duration);
toast.info = (message, duration) => toast.add(message, 'info', duration);

// Location store for geolocation
export const location = writable({
  latitude: null,
  longitude: null,
  error: null,
  loading: false,
});

export function getCurrentLocation() {
  location.update(loc => ({ ...loc, loading: true, error: null }));

  if (!navigator.geolocation) {
    location.update(loc => ({ 
      ...loc, 
      loading: false, 
      error: 'Geolocation is not supported by your browser' 
    }));
    return;
  }

  navigator.geolocation.getCurrentPosition(
    (position) => {
      location.set({
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
        error: null,
        loading: false,
      });
    },
    (error) => {
      let errorMessage = 'Unable to get location';
      switch(error.code) {
        case error.PERMISSION_DENIED:
          errorMessage = 'Location permission denied';
          break;
        case error.POSITION_UNAVAILABLE:
          errorMessage = 'Location information unavailable';
          break;
        case error.TIMEOUT:
          errorMessage = 'Location request timed out';
          break;
      }
      location.update(loc => ({ 
        ...loc, 
        loading: false, 
        error: errorMessage 
      }));
    },
    {
      enableHighAccuracy: true,
      timeout: 10000,
      maximumAge: 0
    }
  );
}
