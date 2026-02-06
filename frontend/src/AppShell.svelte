<script>
  import { onMount } from 'svelte';
  import { Router, goto, replace } from '@mateothegreat/svelte5-router';
  import Layout from './components/Layout.svelte';
  import { auth } from './lib/auth.svelte.js';
  import { api } from './lib/api.js';

  import Dashboard from './pages/Dashboard.svelte';
  import Tasks from './pages/Tasks.svelte';
  import TaskDetail from './pages/TaskDetail.svelte';
  import TaskCreate from './pages/TaskCreate.svelte';
  import TaskAssignments from './pages/TaskAssignments.svelte';
  import AssignmentDetail from './pages/AssignmentDetail.svelte';
  import Attendance from './pages/Attendance.svelte';
  import Assessments from './pages/Assessments.svelte';
  import Reports from './pages/Reports.svelte';
  import Notifications from './pages/Notifications.svelte';
  import Settings from './pages/Settings.svelte';
  import Interns from './pages/Interns.svelte';
  import Supervisors from './pages/Supervisors.svelte';
  import ExportImport from './pages/ExportImport.svelte';
  import Profile from './pages/Profile.svelte';
  import Calendar from './pages/Calendar.svelte';
  import NotFound from './pages/NotFound.svelte';
    import ProfileEdit from './pages/ProfileEdit.svelte';

  // NOTE: this router picks the *last* matching route,
  // so order from broad -> specific.
  const routes = [
    { path: '/dashboard', component: Dashboard },
    { path: '/tasks', component: Tasks },
    { path: /^\/tasks\/(?<id>[^/]+)$/, component: TaskDetail },
    { path: '/tasks/create', component: TaskCreate },
    { path: '/task-assignments', component: TaskAssignments },
    { path: /^\/task-assignments\/(?<id>[^/]+)$/, component: AssignmentDetail },
    { path: '/attendance', component: Attendance },
    { path: '/assessments', component: Assessments },
    { path: '/reports', component: Reports },
    { path: '/notifications', component: Notifications },
    { path: '/settings', component: Settings },
    { path: '/interns', component: Interns },
    { path: '/supervisors', component: Supervisors },
    { path: '/data-tools', component: ExportImport },
    { path: '/profile', component: Profile },
    { path: '/profile/edit', component: ProfileEdit },
    { path: '/calendar', component: Calendar },
  ];

  onMount(async () => {
    const params = new URLSearchParams(window.location.search);
    const oauthToken = params.get('token');
    if (oauthToken) {
      auth.token = oauthToken;
      localStorage.setItem('token', oauthToken);
      params.delete('token');
      params.delete('setup_required');
      const clean = window.location.pathname + (params.toString() ? `?${params.toString()}` : '');
      window.history.replaceState({}, '', clean);
    }

    if (auth.token && !auth.user) {
      try {
        const res = await api.getCurrentUser();
        auth.hydrate(res.data);
      } catch (err) {
        auth.logout();
        replace('/login');
      }
    } else {
      auth.isLoading = false;
      if (!auth.token) auth.logout();
    }
  });

  $effect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      const path = window.location.pathname;
      if (path !== '/login' && path !== '/forgot-password' && path !== '/reset-password') {
        replace('/login');
      }
      return;
    }

    if (!auth.isLoading && auth.isAuthenticated) {
      const path = window.location.pathname;
      if (!path || path === '/dashboard') return;
      if (path === '/') {
        replace('/dashboard');
        return;
      }
      const known = [
        '/dashboard',
        '/tasks',
        '/task-assignments',
        '/attendance',
        '/assessments',
        '/reports',
        '/notifications',
        '/settings',
        '/interns',
        '/supervisors',
        '/data-tools',
        '/profile',
        '/calendar',
      ];
      const isKnown =
        known.some((p) => path === p || path.startsWith(`${p}/`)) ||
        /^\/tasks\/[^/]+$/.test(path) ||
        /^\/task-assignments\/[^/]+$/.test(path);
      if (!isKnown) {
        replace('/');
      }
    }
  });
</script>

{#if auth.isLoading}
  <div class="main-content">
    <div class="card">Memuat...</div>
  </div>
{:else if !auth.isAuthenticated}
  <div class="main-content">
    <div class="card">Mengalihkan ke login...</div>
  </div>
{:else}
  <Layout>
    <Router {routes} notFoundComponent={NotFound} />
  </Layout>
{/if}
