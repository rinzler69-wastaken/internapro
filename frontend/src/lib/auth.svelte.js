// src/lib/auth.svelte.js
export const auth = $state({
  user: null,
  token: localStorage.getItem('token') || null,
  isAuthenticated: false,
  isLoading: true,

  // Methods to mutate state
  login(userData, token) {
    this.user = userData;
    this.token = token;
    this.isAuthenticated = true;
    this.isLoading = false;
    localStorage.setItem('token', token);
  },

  hydrate(userData) {
    this.user = userData;
    this.isAuthenticated = !!userData;
    this.isLoading = false;
  },

  logout() {
    this.user = null;
    this.token = null;
    this.isAuthenticated = false;
    this.isLoading = false;
    localStorage.removeItem('token');
  }
});
