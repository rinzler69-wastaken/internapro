<script>
  import { Router } from "@mateothegreat/svelte5-router";

  // Import halaman Auth
  import Login from "./pages/Login.svelte";
  import Register from "./pages/Register.svelte";
  import ForgotPassword from "./pages/ForgotPassword.svelte";
  import ResetPassword from "./pages/ResetPassword.svelte";
  import WaitingApproval from "./pages/WaitingApproval.svelte";

  // Import Layout & Lainnya
  import AppShell from "./AppShell.svelte";
  import NotFound from "./pages/NotFound.svelte";
  import RedirectDashboard from "./pages/RedirectDashboard.svelte";

  // NOTE: Router ini memilih route yang cocok, jadi urutan penting.
  // Kita list route secara eksplisit untuk menghindari konflik catch-all yang menyebabkan blank screen.
  const routes = [
    { path: "/", component: RedirectDashboard },
    { path: "/login", component: Login },
    { path: "/register", component: Register },
    { path: "/forgot-password", component: ForgotPassword },
    { path: "/reset-password", component: ResetPassword },
    { path: "/waiting-approval", component: WaitingApproval },

    // Dashboard & Protected Routes (Handled by AppShell)
    { path: "/dashboard", component: AppShell },
    { path: "/tasks", component: AppShell },
    { path: "/tasks/:id", component: AppShell },
    { path: "/task-assignments", component: AppShell },
    { path: "/task-assignments/:id", component: AppShell },
    { path: "/attendance", component: AppShell },
    { path: "/assessments", component: AppShell },
    { path: "/reports", component: AppShell },
    { path: "/notifications", component: AppShell },
    { path: "/settings", component: AppShell },
    { path: "/settings/offices", component: AppShell },
    { path: "/interns", component: AppShell },
    { path: "/interns/:id/details", component: AppShell },
    { path: "/supervisors", component: AppShell },
    { path: "/data-tools", component: AppShell },
    { path: "/profile", component: AppShell },
    { path: "/profile/edit", component: AppShell },
    { path: "/calendar", component: AppShell },
    { path: "/forbidden", component: AppShell },
    { path: "/register-intern", component: AppShell },
  ];
</script>

<Router {routes} notFoundComponent={NotFound} />
