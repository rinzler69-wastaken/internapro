<script>
  import { Router, Route } from 'svelte-routing';
  import { onMount } from 'svelte';
  import { auth } from './lib/stores';

  // Pages
  import Login from './pages/Login.svelte';
  import Dashboard from './pages/Dashboard.svelte';
  import Attendance from './pages/Attendance.svelte';
  import Tasks from './pages/Tasks.svelte';
  import Analytics from './pages/Analytics.svelte';
  import Profile from './pages/Profile.svelte';

  // Components
  import Navbar from './components/Navbar.svelte';
  import Toast from './components/Toast.svelte';
  import LoadingSpinner from './components/LoadingSpinner.svelte';

  export let url = '';

  onMount(async () => {
    // Only check auth if there's a token
    const token = localStorage.getItem('token');
    if (token) {
      await auth.checkAuth();
    } else {
      // No token, set loading to false
      auth.setUser(null);
    }
  });
</script>

<main class="min-h-screen bg-gray-50">
  <Toast />

  {#if $auth.loading}
    <div class="flex items-center justify-center min-h-screen">
      <LoadingSpinner size="lg" />
    </div>
  {:else}
    <Router {url}>
      {#if $auth.isAuthenticated}
        <Navbar />
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Route path="/" component={Dashboard} />
          <Route path="/dashboard" component={Dashboard} />
          <Route path="/attendance" component={Attendance} />
          <Route path="/tasks" component={Tasks} />
          <Route path="/analytics" component={Analytics} />
          <Route path="/profile" component={Profile} />
          <Route path="/login" component={Dashboard} />
        </div>
      {:else}
        <Route path="/" component={Login} />
        <Route path="/login" component={Login} />
        <Route path="*" component={Login} />
      {/if}
    </Router>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
  
  .min-h-screen {
    min-height: 100vh;
  }
  
  .bg-gray-50 {
    background-color: #fafafa;
  }
  
  .flex {
    display: flex;
  }
  
  .items-center {
    align-items: center;
  }
  
  .justify-center {
    justify-content: center;
  }
  
  .max-w-7xl {
    max-width: 80rem;
  }
  
  .mx-auto {
    margin-left: auto;
    margin-right: auto;
  }
  
  .px-4 {
    padding-left: 1rem;
    padding-right: 1rem;
  }
  
  .py-8 {
    padding-top: 2rem;
    padding-bottom: 2rem;
  }
  
  @media (min-width: 640px) {
    .sm\:px-6 {
      padding-left: 1.5rem;
      padding-right: 1.5rem;
    }
  }
  
  @media (min-width: 1024px) {
    .lg\:px-8 {
      padding-left: 2rem;
      padding-right: 2rem;
    }
  }
</style>
