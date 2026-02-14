<script>
  import { onMount } from "svelte";
  import { Router, replace } from "@mateothegreat/svelte5-router";
  import Layout from "./components/Layout.svelte";
  import { auth } from "./lib/auth.svelte.js";
  import { api } from "./lib/api.js";
  import { needsProfileCompletion } from "./lib/auth-helpers.js";
  import { location } from "./lib/location.js";

  import Dashboard from "./pages/Dashboard.svelte";
  import RegisterIntern from "./pages/RegisterIntern.svelte";
  import Tasks from "./pages/Tasks.svelte";
  import TaskEdit from "./pages/TaskEdit.svelte";
  import TaskDetail from "./pages/TaskDetail.svelte";
  import TaskAssignments from "./pages/TaskAssignments.svelte";
  import AssignmentDetail from "./pages/AssignmentDetail.svelte";
  import Attendance from "./pages/Attendance.svelte";
  import AttendanceDetail from "./pages/AttendanceDetail.svelte";
  import Assessments from "./pages/Assessments.svelte";
  import AssessmentDetail from "./pages/AssessmentDetail.svelte";
  import Reports from "./pages/Reports.svelte";
  import ReportDetail from "./pages/ReportDetail.svelte";
  import Notifications from "./pages/Notifications.svelte";
  import Settings from "./pages/Settings.svelte";
  import Interns from "./pages/Interns.svelte";
  import Supervisors from "./pages/Supervisors.svelte";
  import ExportImport from "./pages/ExportImport.svelte";
  import Profile from "./pages/Profile.svelte";
  import Calendar from "./pages/Calendar.svelte";
  import NotFound from "./pages/NotFound.svelte";
  import ProfileEdit from "./pages/ProfileEdit.svelte";
  import WaitingApproval from "./pages/WaitingApproval.svelte";
  import Forbidden from "./pages/Forbidden.svelte";
  import InternDetails from "./pages/InternDetails.svelte";

  let profileReady = $state(false);

  // NOTE: this router picks the *last* matching route,
  // so order from broad -> specific.
  const routes = [
    { path: "/dashboard", component: Dashboard },
    { path: "/register-intern", component: RegisterIntern },
    { path: "/tasks", component: Tasks },
    { path: /^\/tasks\/(?<id>[^/]+)$/, component: TaskDetail },
    // { path: "/task-assignments", component: TaskAssignments },
    // { path: "/task-assignments", component: TaskAssignments },
    // { path: /^\/task-assignments\/(?<id>[^/]+)$/, component: AssignmentDetail },
    { path: "/attendance", component: Attendance },
    { path: /^\/attendance\/(?<id>[^/]+)$/, component: AttendanceDetail },
    { path: "/assessments", component: Assessments },
    { path: /^\/assessments\/(?<id>[^/]+)$/, component: AssessmentDetail },
    { path: "/reports", component: Reports },
    { path: /^\/reports\/(?<id>[^/]+)$/, component: ReportDetail },
    { path: "/notifications", component: Notifications },
    { path: "/settings", component: Settings },
    { path: "/interns", component: Interns },
    { path: /^\/interns\/(?<id>[^/]+)\/details$/, component: InternDetails },
    { path: "/supervisors", component: Supervisors },
    { path: "/data-tools", component: ExportImport },
    { path: "/profile", component: Profile },
    { path: "/profile/edit", component: ProfileEdit },
    { path: "/calendar", component: Calendar },
    { path: "/forbidden", component: Forbidden },
    { path: /^\/tasks\/edit\/(?<id>[^/]+)$/, component: TaskEdit },
  ];

  onMount(async () => {
    const params = new URLSearchParams(window.location.search);
    const oauthToken = params.get("token");
    if (oauthToken) {
      auth.token = oauthToken;
      localStorage.setItem("token", oauthToken);
      params.delete("token");
      params.delete("setup_required");
      const clean =
        window.location.pathname +
        (params.toString() ? `?${params.toString()}` : "");
      window.history.replaceState({}, "", clean);
    }

    if (auth.token && !auth.user) {
      try {
        const res = await api.getCurrentUser();
        let userData = res?.user || res?.data?.user || res?.data || res;

        // Always fetch full profile to attach intern/supervisor status for routing
        try {
          const prof = await api.getProfile();
          const profUser = prof?.user || prof?.data?.user || prof?.data || prof;
          userData = {
            ...(profUser || userData),
            intern: prof?.intern || prof?.data?.intern || userData?.intern,
            supervisor: prof?.supervisor || prof?.data?.supervisor || userData?.supervisor,
          };
        } catch (profileErr) {
          console.warn("[AppShell] Failed to load full profile", profileErr);
        }

        auth.hydrate(userData);
      } catch (err) {
        auth.logout();
        replace("/login");
      } finally {
        auth.isLoading = false;
        profileReady = true;
      }
    } else {
      auth.isLoading = false;
      if (!auth.token) auth.logout();
      profileReady = true;
    }
  });

  const isPending = $derived(
    auth.user?.status === "pending" ||
      auth.user?.intern?.status === "pending" ||
      auth.user?.supervisor?.status === "pending" ||
      (auth.user?.InternID && auth.user?.intern?.status === undefined) // legacy payloads
  );
  const internDataReady = $derived(
    auth.user?.role !== "intern" ||
      !!auth.user?.intern ||
      !!auth.user?.intern_id ||
      !!auth.user?.InternID
  );
  const dataReady = $derived(profileReady && internDataReady);
  const needsCompletion = $derived(needsProfileCompletion(auth.user));
  const currentPath = $derived($location.path || window.location.pathname);

  $effect(() => {
    if (auth.isLoading) return;
    if (!dataReady) return;

    // Force redirect to waiting approval immediately if pending to avoid layout flash
    if (isPending && currentPath !== "/waiting-approval") {
      replace("/waiting-approval");
      return;
    }

    // 0. Handle OAuth Token in URL (Priority)
    // This prevents premature redirection to /login when returning from OAuth
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get("token");
    if (urlToken && !auth.token) {
      console.log("[AppShell] Detected token in URL, hydrating auth...");
      auth.token = urlToken;
      localStorage.setItem("token", urlToken);
      // Determine if we should trigger a profile fetch here or let components handle it
      // For now, setting the token is enough to pass the isAuthenticated check
    }

    const path = currentPath;

    // 1. Not authenticated -> Redirect to /login
    if (!auth.isAuthenticated) {
      // If we have a token, allow hydration to finish instead of bouncing
      if (!auth.token) {
        if (
          path !== "/login" &&
          path !== "/forgot-password" &&
          path !== "/reset-password"
        ) {
          console.log(
            "[AppShell] Not authenticated, redirecting to /login from",
            path,
          );
          replace("/login");
        }
        return;
      }
      // token present, stay put
      return;
    }

    // 2. Authenticated
    const role = auth.user?.role || "";
    console.log("[AppShell] Current state:", {
      path,
      role,
      isPending,
      needsCompletion,
    });

    // 2a. Pending users are forced to waiting page
    if (isPending) {
      const allowed = ["/waiting-approval", "/login"];
      if (!allowed.includes(path)) {
        console.log("[AppShell] Pending user, redirecting to /waiting-approval");
        replace("/waiting-approval");
      }
      return;
    }

    // 2b. Users missing required profile data must complete register (legacy safety)
    if (needsCompletion) {
      const allowed = ["/register", "/register-intern", "/login"];
      if (!allowed.includes(path)) {
        console.log("[AppShell] Incomplete profile, redirecting to /register");
        replace("/register");
      }
      return;
    }

    // Restriction: admin only pages
    if (path.startsWith("/settings") && role !== "admin") {
      replace("/forbidden");
      return;
    }

    // Default redirect for root or generic dashboard landing
    if (path === "/" || path === "/dashboard") {
      const target = isPending
        ? "/waiting-approval"
        : needsCompletion
          ? "/register"
          : "/dashboard";
      if (path !== target) {
        console.log("[AppShell] Default redirect to", target);
        replace(target);
      }
    }
  });
</script>

{#if auth.isLoading || !dataReady}
  <div class="auth-loading-overlay">
    <div class="auth-loading-card">
      <div class="auth-spinner"></div>
      <p>Memuat Akun...</p>
    </div>
  </div>
{:else if isPending}
  <!-- Pending users: show waiting page without dashboard chrome -->
  <WaitingApproval />
{:else}
  <Layout>
    <Router {routes} notFoundComponent={NotFound} />
  </Layout>
{/if}

<style>
  .auth-loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #f3f4f6;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    font-family: "Inter", sans-serif;
  }

  .auth-loading-card {
    background: white;
    padding: 2.5rem;
    border-radius: 12px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
    min-width: 300px;
  }

  .auth-spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top: 4px solid #111827;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .auth-loading-card p {
    color: #6b7280;
    font-size: 0.95rem;
    font-weight: 500;
    margin: 0;
  }
</style>
