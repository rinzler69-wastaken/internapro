<script>
  import { Router } from '@mateothegreat/svelte5-router';
  import Login from './pages/Login.svelte';
  import ForgotPassword from './pages/ForgotPassword.svelte';
  import ResetPassword from './pages/ResetPassword.svelte';
  import AppShell from './AppShell.svelte';
  import NotFound from './pages/NotFound.svelte';
  import RedirectDashboard from './pages/RedirectDashboard.svelte';

  // NOTE: this router picks the *last* matching route,
  // so place auth routes first and the app-shell catch-all last.
  const routes = [
    { path: '/', component: RedirectDashboard },
    { path: '/login', component: Login },
    { path: '/forgot-password', component: ForgotPassword },
    { path: '/reset-password', component: ResetPassword },
    { path: /^\/(?!login$|forgot-password$|reset-password$).*/, component: AppShell },
  ];
</script>

<Router {routes} notFoundComponent={NotFound} />
