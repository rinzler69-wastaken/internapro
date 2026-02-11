<script>
  import { onMount } from "svelte";
  import { replace, route } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";

  // State Management (Svelte 5 Runes)
  let email = $state("");
  let password = $state("");
  let totp = $state("");
  let error = $state("");
  let loading = $state(false);
  let needs2FA = $state(false);
  let setupRequired = $state(false);
  let pendingApproval = $state(false);
  let showPassword = $state(false);

  onMount(async () => {
    // Check for OAuth callback params
    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    const newUser = params.get("new_user");
    const errorParam = params.get("error");

    if (errorParam) {
      error = decodeURIComponent(errorParam);
      // Clean URL
      window.history.replaceState({}, document.title, window.location.pathname);
    }

    if (token) {
      // Save token immediately
      localStorage.setItem("token", token);
      auth.token = token;

      if (newUser === "1") {
        // Redir to register for profile completion
        replace("/register?oauth=google_unregistered");
        return;
      }

      // Existing user: validate & fetch profile
      try {
        const res = await api.getProfile();
        auth.hydrate(res.data);
        replace("/dashboard");
        return;
      } catch (e) {
        console.error("Auto-login failed", e);
        error = "Sesi tidak valid. Silakan login ulang.";
        auth.logout();
      }
    }

    if (auth.token && auth.user) {
      replace("/dashboard");
    }
  });

  // Handler Login Biasa
  async function handleSubmit() {
    error = "";
    pendingApproval = false;
    loading = true;
    try {
      const res = await api.login(email, password, needs2FA ? totp : null);

      if (res.data.user?.status === "pending") {
        pendingApproval = true;
        loading = false;
        return;
      }

      if (res?.data?.require_2fa) {
        needs2FA = true;
        loading = false;
        return;
      }

      setupRequired = !!res?.data?.setup_required;

      // Always fetch fresh profile to know intern status
      try {
        const prof = await api.getProfile();
        const intern = prof.data?.intern;
        if (intern?.status === "pending") {
          pendingApproval = true;
          loading = false;
          replace("/waiting-approval");
          return;
        }
      } catch (e) {
        // ignore profile fetch errors, proceed
      }

      loading = false;
      replace("/dashboard");
    } catch (err) {
      if (err.response && err.response.status === 401) {
        error = "Email atau password Anda salah.";
        loading = false;
        return;
      }

      const msg = err.response?.data?.message || err.message || "Login gagal";
      if (
        msg.toLowerCase().includes("belum disetujui") ||
        msg.toLowerCase().includes("pending")
      ) {
        pendingApproval = true;
        error = "Pendaftaran Anda masih menunggu persetujuan admin.";
        loading = false;
        // replace('/register?status=pending');
      } else {
        error = msg;
        loading = false;
      }
    }
  }

  // legacy for debugging only
  // async function handleSubmit() {
  //   error = '';
  //   loading = true;
  //   try {
  //     const res = await api.login(email, password, needs2FA ? totp : null);
  //     if (res?.data?.require_2fa) {
  //       needs2FA = true;
  //       loading = false;
  //       return;
  //     }
  //     setupRequired = !!res?.data?.setup_required;
  //     loading = false;
  //     if (setupRequired) {
  //       replace('/dashboard');
  //     } else {
  //       replace('/dashboard');
  //     }
  //   } catch (err) {
  //     error = err.message || 'Login gagal';
  //     loading = false;
  //   }
  // }

  //   async function handleSubmit() {
  //   // Validation for empty fields
  //   if (!email || !password) {
  //     error = 'Email dan password wajib diisi';
  //     return;
  //   }

  //   error = '';
  //   pendingApproval = false;
  //   loading = true;

  //   try {
  //     const res = await api.login(email, password, needs2FA ? totp : null);

  //     // Check if user status is pending - BUT ONLY FOR INTERNS
  //     // Admin and supervisor accounts should bypass this check
  //     const userRole = res.data.user?.role?.toLowerCase();
  //     const userStatus = res.data.user?.status?.toLowerCase();

  //     if (userRole === 'intern' && userStatus === 'pending') {
  //       pendingApproval = true;
  //       loading = false;
  //       error = 'Pendaftaran Anda masih menunggu persetujuan admin.';
  //       return;
  //     }

  //     // Check if 2FA is required
  //     if (res?.data?.require_2fa) {
  //       needs2FA = true;
  //       loading = false;
  //       return;
  //     }

  //     // Check if setup is required
  //     setupRequired = !!res?.data?.setup_required;

  //     // Set auth state if token is present
  //     if (res.data.token) {
  //       auth.login(res.data.token, res.data.user);
  //     }

  //     loading = false;

  //     // Redirect to dashboard (or settings if setup required)
  //     if (setupRequired) {
  //       replace('/settings?setup=2fa');
  //     } else {
  //       replace('/dashboard');
  //     }
  //   } catch (err) {
  //     console.error('Login error:', err);
  //     loading = false;

  //     // Handle different error scenarios
  //     const msg = err.message || 'Login gagal';
  //     const msgLower = msg.toLowerCase();

  //     // Check for pending approval in error message - ONLY if the error explicitly mentions it
  //     // Don't assume all 401s are pending approval issues
  //     if (msgLower.includes('belum disetujui') ||
  //         (msgLower.includes('pending') && msgLower.includes('approval'))) {
  //       pendingApproval = true;
  //       error = 'Pendaftaran Anda masih menunggu persetujuan admin.';
  //       return;
  //     }

  //     // Enhanced error messages
  //     if (msgLower.includes('invalid') ||
  //         msgLower.includes('incorrect') ||
  //         msgLower.includes('wrong') ||
  //         msgLower.includes('unauthorized') ||
  //         msgLower.includes('401')) {
  //       error = 'Email atau password salah. Silakan coba lagi.';
  //     } else if (msgLower.includes('not found') || msgLower.includes('404')) {
  //       error = 'Akun tidak ditemukan. Periksa kembali email Anda.';
  //     } else if (msgLower.includes('network') || msgLower.includes('fetch')) {
  //       error = 'Gagal terhubung ke server. Periksa koneksi internet Anda.';
  //     } else if (msgLower.includes('totp') || msgLower.includes('2fa')) {
  //       error = 'Kode autentikasi 2FA salah atau kedaluwarsa.';
  //     } else {
  //       // Use the original error message if none of the patterns match
  //       error = msg;
  //     }
  //   }
  // }

  // Handler Login Google (Backend Redirect Flow)
  function handleGoogle() {
    // Pastikan path ini sesuai dengan route backend Anda yang mengarah ke passport/socialite/oauth
    // Backend akan melempar user ke accounts.google.com
    // After backend check, it will redirect to /register if profile is missing/pending.
    const redirectPath = "/dashboard";
    const backendUrl = import.meta.env.VITE_API_URL || ""; // Fallback jika pakai proxy
    window.location.href = `${backendUrl}/api/auth/google?redirect=1&redirect_path=${encodeURIComponent(redirectPath)}`;
  }

  async function handleCancelRegistration() {
    if (
      !confirm(
        "Apakah Anda yakin ingin membatalkan pendaftaran? Data Anda akan dihapus dan Anda dapat mendaftar ulang.",
      )
    )
      return;
    try {
      const userId = auth.user?.id;
      if (userId) {
        await api.delete(`/interns/${userId}`);
        auth.logout();
        pendingApproval = false;
        alert("Pendaftaran dibatalkan.");
      }
    } catch (err) {
      error = "Gagal membatalkan: " + (err.message || "Terjadi kesalahan");
    }
  }
</script>

<div class="login-container">
  <div class="card login-card">
    {#if pendingApproval}
      <div class="card-header fade-in">
        <div class="brand-logo warning-logo">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="10"></circle><line
              x1="12"
              y1="8"
              x2="12"
              y2="12"
            ></line><line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
        </div>
        <h2 class="title">Menunggu Persetujuan</h2>
      </div>

      <div class="approval-box-wrapper fade-in">
        <div class="approval-box">
          <p class="approval-text">
            Akun Anda sedang menunggu persetujuan admin.
          </p>
        </div>
      </div>

      <button
        class="btn btn-outline"
        onclick={() => {
          pendingApproval = false;
          auth.logout();
        }}
      >
        Kembali ke Login
      </button>
    {:else}
      <!-- Logo & Header -->
      <div class="card-header">
        <div class="brand-logo">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M3 7.5l9-4 9 4-9 4-9-4z" />
            <path d="M6 10.5v4.5c0 2 3 3.5 6 3.5s6-1.5 6-3.5v-4.5" />
          </svg>
        </div>
        <h2 class="title">Masuk InternaPro</h2>
        <p class="subtitle">Kelola magang, tugas, dan presensi.</p>
      </div>

      <!-- Error Alert -->
      {#if error}
        <div class="alert error">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            ><circle cx="12" cy="12" r="10"></circle><line
              x1="12"
              y1="8"
              x2="12"
              y2="12"
            ></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg
          >
          <span>{error}</span>
        </div>
      {/if}

      {#if setupRequired}
        <div class="alert warning">
          <span>Akun Anda membutuhkan setup keamanan.</span>
        </div>
      {/if}

      <!-- Main Form -->
      <div class="form-body">
        <div class="form-group">
          <label for="email">Email</label>
          <input
            class="input-field"
            type="email"
            id="email"
            bind:value={email}
            placeholder="nama@sekolah.com"
            autocomplete="email"
          />
        </div>

        <div class="form-group relative">
          <div class="label-row">
            <label for="password">Password</label>
          </div>
          <div class="password-wrapper">
            <input
              class="input-field"
              type={showPassword ? "text" : "password"}
              id="password"
              bind:value={password}
              placeholder="••••••••"
              autocomplete="current-password"
              onkeydown={(e) => e.key === "Enter" && handleSubmit()}
            />
            <button
              class="toggle-password"
              onclick={() => (showPassword = !showPassword)}
              type="button"
              aria-label="Toggle password visibility"
            >
              {#if showPassword}
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  ><path
                    d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                  ></path><line x1="1" y1="1" x2="23" y2="23"></line></svg
                >
              {:else}
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  ><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                  ></path><circle cx="12" cy="12" r="3"></circle></svg
                >
              {/if}
            </button>
          </div>
          <div style="text-align: right; margin-top: 6px;">
            <a href="/forgot-password" use:route class="link-muted small"
              >Lupa password?</a
            >
          </div>
        </div>

        {#if needs2FA}
          <div class="form-group fade-in">
            <label for="totp">Kode Autentikasi (2FA)</label>
            <input
              class="input-field"
              type="text"
              id="totp"
              bind:value={totp}
              placeholder="123 456"
              maxlength="6"
              onkeydown={(e) => e.key === "Enter" && handleSubmit()}
            />
          </div>
        {/if}

        <button
          class="btn btn-primary"
          onclick={handleSubmit}
          disabled={loading}
        >
          {#if loading}
            <span class="spinner"></span> Memproses...
          {:else}
            Masuk
          {/if}
        </button>

        <div class="divider">
          <span>atau masuk dengan</span>
        </div>

        <button class="btn btn-google" onclick={handleGoogle} type="button">
          <svg
            class="google-icon"
            viewBox="0 0 24 24"
            width="18"
            height="18"
            xmlns="http://www.w3.org/2000/svg"
          >
            <g transform="matrix(1, 0, 0, 1, 27.009001, -39.238998)">
              <path
                fill="#4285F4"
                d="M -3.264 51.509 C -3.264 50.719 -3.334 49.969 -3.454 49.239 L -14.754 49.239 L -14.754 53.749 L -8.284 53.749 C -8.574 55.229 -9.424 56.479 -10.684 57.329 L -10.684 60.329 L -6.824 60.329 C -4.564 58.239 -3.264 55.159 -3.264 51.509 Z"
              />
              <path
                fill="#34A853"
                d="M -14.754 63.239 C -11.514 63.239 -8.804 62.159 -6.824 60.329 L -10.684 57.329 C -11.764 58.049 -13.134 58.489 -14.754 58.489 C -17.884 58.489 -20.534 56.379 -21.484 53.529 L -25.464 53.529 L -25.464 56.619 C -23.494 60.539 -19.444 63.239 -14.754 63.239 Z"
              />
              <path
                fill="#FBBC05"
                d="M -21.484 53.529 C -21.734 52.809 -21.864 52.039 -21.864 51.239 C -21.864 50.439 -21.734 49.669 -21.484 48.949 L -21.484 45.859 L -25.464 45.859 C -26.284 47.479 -26.754 49.299 -26.754 51.239 C -26.754 53.179 -26.284 54.999 -25.464 56.619 L -21.484 53.529 Z"
              />
              <path
                fill="#EA4335"
                d="M -14.754 43.989 C -12.984 43.989 -11.404 44.599 -10.154 45.789 L -6.734 42.369 C -8.804 40.429 -11.514 39.239 -14.754 39.239 C -19.444 39.239 -23.494 41.939 -25.464 45.859 L -21.484 48.949 C -20.534 46.099 -17.884 43.989 -14.754 43.989 Z"
              />
            </g>
          </svg>
          <span>Masuk dengan Google</span>
        </button>

        <!-- Link Daftar Magang Baru -->
        <div style="text-align: center; margin-top: 1rem;">
          <a href="/register" use:route class="link-blue">
            Belum punya akun? Daftar Magang
          </a>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Basic Reset & Container */
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background-color: #f3f4f6;
    padding: 1rem;
    font-family:
      "Inter",
      -apple-system,
      sans-serif;
  }

  /* Card Style */
  .card {
    background: white;
    width: 100%;
    max-width: 400px;
    border-radius: 12px;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    padding: 2.5rem 2rem;
  }

  /* Header */
  .card-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .brand-logo {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    background-color: #eff6ff;
    color: #111827; /*warna logo brand*/
    border-radius: 10px;
    margin-bottom: 1rem;
  }

  .warning-logo {
    background-color: #fffbeb;
    color: #d97706;
  }

  .title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

  /* Form Elements */
  .form-group {
    margin-bottom: 1.25rem;
  }

  .form-group.relative {
    position: relative;
  }

  label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    margin-bottom: 0.5rem;
  }

  .input-field {
    width: 100%;
    padding: 0.75rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 0.95rem;
    transition: all 0.2s;
    box-sizing: border-box; /* Important for padding */
  }

  .input-field:focus {
    outline: none;
    border-color: #2563eb;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
  }

  /* Password Toggle */
  .password-wrapper {
    position: relative;
  }

  .toggle-password {
    position: absolute;
    right: 0;
    top: 0;
    height: 100%;
    padding: 0 12px;
    background: none;
    border: none;
    color: #6b7280;
    cursor: pointer;
    display: flex;
    align-items: center;
  }

  .toggle-password:hover {
    color: #374151;
  }

  /* Buttons */
  .btn {
    width: 100%;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.95rem;
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid transparent;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
  }

  .btn-primary {
    background-color: #ffffff;
    color: #374151;
    border: 1px solid #04070b;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #111827;
    color: #ffffff;
  }

  .btn-primary:disabled {
    background-color: rgb(134, 134, 135);
    cursor: not-allowed;
  }

  .btn-outline {
    background-color: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-outline:hover {
    background-color: #f9fafb;
    border-color: #9ca3af;
  }

  /* Google Button */
  .btn-google {
    background-color: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-google:hover {
    background-color: #111827;
    border-color: #c0c4cc;
    color: #ffffff;
  }

  /* Divider */
  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    margin: 1.5rem 0;
  }

  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    border-bottom: 1px solid #e5e7eb;
  }

  .divider span {
    padding: 0 0.75rem;
    color: #9ca3af;
    font-size: 0.8rem;
    text-transform: uppercase;
    font-weight: 500;
    letter-spacing: 0.05em;
  }

  /* Alerts */
  .alert {
    padding: 0.75rem;
    border-radius: 6px;
    font-size: 0.875rem;
    margin-bottom: 1.5rem;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .alert.error {
    background-color: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
  }

  .alert.warning {
    background-color: #fffbeb;
    color: #d97706;
    border: 1px solid #fde68a;
  }

  /* Utilities */
  .link-muted {
    color: #6b7280;
    text-decoration: none;
    transition: color 0.2s;
  }

  .link-muted:hover {
    color: #2563eb;
    text-decoration: underline;
  }

  /* Styles baru untuk link daftar */
  .link-blue {
    color: #2563eb;
    font-size: 0.85rem;
    text-decoration: none;
    font-weight: 500;
    cursor: pointer;
  }

  .link-blue:hover {
    text-decoration: underline;
  }

  .small {
    font-size: 0.85rem;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #ffffff;
    border-top: 2px solid transparent;
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

  /* Approval Box */
  .approval-box-wrapper {
    margin-bottom: 2rem;
  }

  .approval-box {
    background-color: #fefce8;
    border: 1px solid #fde047;
    color: #854d0e;
    padding: 1.25rem;
    border-radius: 8px;
    text-align: center;
  }

  .approval-text {
    margin: 0;
    font-size: 0.95rem;
    line-height: 1.5;
    font-weight: 500;
  }

  .fade-in {
    animation: fadeIn 0.3s ease-in;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(-5px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
