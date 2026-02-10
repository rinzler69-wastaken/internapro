<script>
  import { Router } from '@mateothegreat/svelte5-router';
  
  // Import halaman Auth
  import Login from './pages/Login.svelte';
  import Register from './pages/Register.svelte'; // <-- 1. Import halaman Register
  import ForgotPassword from './pages/ForgotPassword.svelte';
  import ResetPassword from './pages/ResetPassword.svelte';
  
  // Import Layout & Lainnya
  import AppShell from './AppShell.svelte';
  import NotFound from './pages/NotFound.svelte';
  import RedirectDashboard from './pages/RedirectDashboard.svelte';

  // NOTE: Router ini memilih route yang cocok, jadi urutan penting.
  // Auth routes ditaruh di atas, dan catch-all (AppShell) ditaruh paling bawah dengan pengecualian regex.
  const routes = [
    { path: '/', component: RedirectDashboard },
    { path: '/login', component: Login },
    { path: '/register', component: Register }, // <-- 2. Daftarkan route /register
    { path: '/forgot-password', component: ForgotPassword },
    { path: '/reset-password', component: ResetPassword },
    
    // 3. Update Regex: Tambahkan '|register$' agar tidak dianggap masuk ke AppShell (Dashboard)
    { path: /^\/(?!login$|register$|forgot-password$|reset-password$).*/, component: AppShell },
  ];
</script>

<Router {routes} notFoundComponent={NotFound} />