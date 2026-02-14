<script>
  import { onMount } from "svelte";
  import { replace, route } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import { needsProfileCompletion } from "../lib/auth-helpers.js";

  // --- STATE (Svelte 5 Runes) ---
  let name = $state("");
  let email = $state("");
  let password = $state("");
  let confirmPassword = $state("");

  // Role 'intern' or 'supervisor'
  let role = $state("intern");

  // Intern-specific fields
  let school = $state("");
  let department = $state("");
  let phone = $state("");
  let address = $state("");
  let startDate = $state("");
  let endDate = $state("");
  let nis = $state("");
  let supervisor_id = $state("");
  let googleId = $state("");
  let prefillAvatar = $state("");

  // Supervisor-specific fields
  let nip = $state("");
  let supervisor_phone = $state("");
  let institution = $state("");
  let supervisor_address = $state("");

  let error = $state("");
  let loading = $state(false);
  let showPassword = $state(false);
  let showConfirmPassword = $state(false);

  let isSubmitted = $state(false);
  let isGoogleRedirect = $state(false);
  let initialLoading = $state(true); // New: initial page load spinner

  let supervisors = $state([]);

  onMount(async () => {
    // Show spinner for 300ms on every page load
    const minLoadTime = new Promise((resolve) => setTimeout(resolve, 300));
    let prefillPromise = null;

    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get("token");
    if (urlToken) {
      localStorage.setItem("token", urlToken);
      auth.token = urlToken;
      // Clean URL
      params.delete("token");
      const clean =
        window.location.pathname +
        (params.toString() ? `?${params.toString()}` : "");
      window.history.replaceState({}, "", clean);
    }

    // Attempt to hydrate user if token exists but user is null
    if (auth.token && !auth.user) {
      try {
        const res = await api.getProfile();
        const profUser = res?.user || res?.data?.user || res?.data || res;
        auth.user = profUser;
      } catch (e) {
        console.error("Failed to fetch profile in register", e);
      }
    }

    // Prefill from existing session if available
    if (auth.user) {
      if (!name && auth.user.name) name = auth.user.name;
      if (!email && auth.user.email) email = auth.user.email;
    }

    if (auth.token && auth.user) {
      const isPending =
        auth.user.status === "pending" ||
        auth.user.intern?.status === "pending" ||
        auth.user.supervisor?.status === "pending";

      const incomplete = needsProfileCompletion(auth.user);

      if (incomplete) {
        name = auth.user.name || "";
        email = auth.user.email || "";
        isGoogleRedirect = true;
      } else if (isPending) {
        isSubmitted = true;
      } else {
        console.log(
          "[Register] User already fully registered, redirecting to dashboard",
        );
        initialLoading = false;
        await minLoadTime;
        replace("/dashboard");
        return;
      }
    }

    const oauthStatus = params.get("oauth");
    if (
      oauthStatus === "google_unregistered" ||
      oauthStatus === "google" ||
      params.get("status") === "unregistered"
    ) {
      isGoogleRedirect = true;
      // If we have token/user, we prefer that data. Otherwise use params if avail.
      if (!email) email = params.get("email") || "";
      if (!name) name = params.get("name") || "";
      googleId = params.get("google_id") || "";
      prefillAvatar = params.get("avatar") || "";

      // Best-effort prefill from Google account via profile API when data missing
      if ((!name || !email) && auth.token) {
        prefillPromise = api
          .getProfile()
          .then((prof) => {
            const profUser = prof?.data?.user || prof?.data;
            if (!name && profUser?.name) name = profUser.name;
            if (!email && profUser?.email) email = profUser.email;
          })
          .catch((err) => {
            console.warn("[Register] Failed to prefill from profile", err);
          });
      }
    } else if ((!name || !email) && auth.token && !prefillPromise) {
      // Non-Google path: still try to prefill via profile for usability
      prefillPromise = api
        .getProfile()
        .then((prof) => {
          const profUser = prof?.data?.user || prof?.data;
          if (!name && profUser?.name) name = profUser.name;
          if (!email && profUser?.email) email = profUser.email;
        })
        .catch((err) => {
          console.warn("[Register] Prefill via profile (generic) failed", err);
        });
    }

    api
      .get("/supervisors")
      .then((res) => {
        // Public supervisors endpoint returns { user_id, name }
        supervisors = (res.data || [])
          .map((s) => ({
            id: s.user_id ?? s.id ?? null,
            name: s.name || s.user?.name || "Pembimbing",
            institution: s.institution || s.institution_name || "",
          }))
          .filter((s) => s.id !== null);
      })
      .catch((err) => {
        console.error("Failed to fetch supervisors", err);
      })
      .finally(async () => {
        if (prefillPromise) {
          try {
            await prefillPromise;
          } catch (_) {
            /* ignore */
          }
        }
        await minLoadTime;
        initialLoading = false;
      });
  });

  async function handleRegister(e) {
    e?.preventDefault();
    if (password !== confirmPassword) {
      error = "Konfirmasi password tidak cocok";
      return;
    }

    error = "";
    loading = true;

    let payload = {
      role,
      name,
      email,
      password,
      confirm_password: confirmPassword,
      google_id: googleId || null,
      provider: googleId ? "google" : "local",
      avatar: prefillAvatar || null,
    };

    try {
      if (role === "intern") {
        // Intern registration
        const supervisorIdNumber = supervisor_id ? Number(supervisor_id) : null;
        if (supervisor_id && Number.isNaN(supervisorIdNumber)) {
          error = "Pembimbing tidak valid.";
          loading = false;
          return;
        }

        Object.assign(payload, {
          school,
          department,
          phone,
          address,
          start_date: startDate,
          end_date: endDate,
          nis,
          supervisor_id: supervisorIdNumber,
        });

        await api.post("/internship/register", payload);
      } else if (role === "supervisor") {
        // Supervisor registration
        Object.assign(payload, {
          nip,
          phone: supervisor_phone,
          institution,
          address: supervisor_address,
        });

        await api.post("/supervisor/register", payload);
      }

      // Only logout if user already had a session/token to clear; avoids 401 spam when anonymous.
      if (auth.token) {
        try {
          await api.logout();
        } catch (e) {
          /* ignore */
        }
      }

      loading = false;
      isSubmitted = true;
    } catch (err) {
      console.error(err);
      error =
        err.response?.data?.message ||
        err.message ||
        "Pendaftaran gagal. Silakan coba lagi.";
      loading = false;
    }
  }

  function handleGoogleRegister() {
    const redirectPath = "/dashboard";
    const backendUrl = import.meta.env.VITE_API_URL || "";
    window.location.href = `${backendUrl}/api/auth/google?redirect=1&redirect_path=${encodeURIComponent(redirectPath)}`;
  }
</script>

<div class="login-container">
  {#if initialLoading}
    <!-- Initial Loading Spinner -->
    <div class="initial-loader fade-in">
      <div class="brand-logo-large">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <line x1="22" x2="18" y1="8" y2="8" />
          <line x1="20" x2="20" y1="6" y2="10" />
        </svg>
      </div>
      <div class="spinner-large"></div>
    </div>
  {:else}
    <div class="card register-card fade-in">
      {#if isSubmitted}
        <!-- TAMPILAN SUKSES SUBMIT -->
        <div class="card-header fade-in">
          <div class="brand-logo success-logo">
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
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
              <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
          </div>
          <h2 class="title">Pendaftaran Terkirim</h2>
          <p class="subtitle">Akun Anda akan segera kami proses.</p>
        </div>

        <div class="approval-box-wrapper fade-in">
          <div class="approval-box">
            <p class="approval-text">
              Setelah mendaftar, akun Anda akan menunggu persetujuan dari admin.
              Anda akan menerima notifikasi setelah akun diaktifkan.
            </p>
          </div>
        </div>

        <button class="btn btn-outline" onclick={() => replace("/login")}>
          Kembali ke Login
        </button>
      {:else}
        <!-- FORM PENDAFTARAN -->
        <!-- Header -->
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
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <line x1="22" x2="18" y1="8" y2="8" />
              <line x1="20" x2="20" y1="6" y2="10" />
            </svg>
          </div>
          <h2 class="title">Pendaftaran</h2>
          <p class="subtitle">
            Pilih jenis akun dan isi data lengkap untuk mendaftar
          </p>
        </div>

        {#if isGoogleRedirect}
          <div class="alert warning">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            <span>
              Anda masuk dengan Google. Lengkapi data berikut untuk menyelesaikan pendaftaran.
            </span>
          </div>
        {/if}

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
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            <span>{error}</span>
          </div>
        {/if}

        <form onsubmit={handleRegister}>
          <!-- Role Selector -->
          <div class="form-group">
            <label for="role-selector">Daftar Sebagai</label>
            <div class="role-selector">
              <button
                type="button"
                class="role-option"
                class:active={role === "intern"}
                onclick={() => (role = "intern")}
              >
                <div class="role-icon">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    ><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" /><path
                      d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"
                    /></svg
                  >
                </div>
                <div class="role-text">
                  <span class="role-title">Peserta Magang</span>
                  <span class="role-desc">Siswa / Mahasiswa</span>
                </div>
              </button>
              <button
                type="button"
                class="role-option"
                class:active={role === "supervisor"}
                onclick={() => (role = "supervisor")}
              >
                <div class="role-icon">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    ><path d="M18 8h1a4 4 0 0 1 0 8h-1" /><path
                      d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z"
                    /><line x1="6" x2="6" y1="1" y2="4" /><line
                      x1="10"
                      x2="10"
                      y1="1"
                      y2="4"
                    /><line x1="14" x2="14" y1="1" y2="4" /></svg
                  >
                </div>
                <div class="role-text">
                  <span class="role-title">Pembimbing</span>
                  <span class="role-desc">Mentor / Guru</span>
                </div>
              </button>
            </div>
          </div>

          <div class="divider"><span>Informasi Akun</span></div>

          <!-- Account Info -->
          <div class="grid-2">
            <div class="form-group">
              <label for="fullName"
                >Nama Lengkap <span class="text-error">*</span></label
              >
              <input
                id="fullName"
                type="text"
                bind:value={name}
                class="input-field"
                placeholder="Nama lengkap"
                required
              />
            </div>
            <div class="form-group">
              <label for="email">Email <span class="text-error">*</span></label>
              <input
                id="email"
                type="email"
                bind:value={email}
                class="input-field"
                placeholder="nama@sekolah.com"
                required
                disabled={(isGoogleRedirect ||
                  auth.user?.role === "new_user") &&
                  email.length > 0}
              />
            </div>
          </div>

          <div class="grid-2">
            <div class="form-group relative">
              <label for="password"
                >Password <span class="text-error">*</span></label
              >
              <div class="password-wrapper">
                <input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  bind:value={password}
                  class="input-field"
                  placeholder="Min 8 karakter"
                  required
                />
                <button
                  type="button"
                  class="toggle-password"
                  onclick={() => (showPassword = !showPassword)}
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
            </div>
            <div class="form-group relative">
              <label for="confirmPassword"
                >Konfirmasi <span class="text-error">*</span></label
              >
              <div class="password-wrapper">
                <input
                  id="confirmPassword"
                  type={showConfirmPassword ? "text" : "password"}
                  bind:value={confirmPassword}
                  class="input-field"
                  placeholder="Ulangi password"
                  required
                />
                <button
                  type="button"
                  class="toggle-password"
                  onclick={() => (showConfirmPassword = !showConfirmPassword)}
                >
                  {#if showConfirmPassword}
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
            </div>
          </div>

          <!-- Role Specific Data -->
          {#if role === "intern"}
            <div class="divider"><span>Data Peserta</span></div>
            <div class="fade-in">
              <div class="grid-2">
                <div class="form-group">
                  <label for="nis">NISN / NIM</label>
                  <input
                    id="nis"
                    type="text"
                    bind:value={nis}
                    class="input-field"
                    placeholder="Nomor induk"
                  />
                </div>
                <div class="form-group">
                  <label for="phone">WhatsApp / Telepon</label>
                  <input
                    id="phone"
                    type="text"
                    bind:value={phone}
                    class="input-field"
                    placeholder="08xxxxxxxxxx"
                  />
                </div>
              </div>
              <div class="grid-2">
                <div class="form-group">
                  <label for="school"
                    >Sekolah / Kampus <span class="text-error">*</span></label
                  >
                  <input
                    id="school"
                    type="text"
                    bind:value={school}
                    class="input-field"
                    placeholder="Nama instansi"
                    required
                  />
                </div>
                <div class="form-group">
                  <label for="department"
                    >Jurusan <span class="text-error">*</span></label
                  >
                  <input
                    id="department"
                    type="text"
                    bind:value={department}
                    class="input-field"
                    placeholder="Contoh: RPL"
                    required
                  />
                </div>
              </div>
              <div class="form-group">
                <label for="address">Alamat</label>
                <textarea
                  id="address"
                  bind:value={address}
                  class="input-field"
                  rows="2"
                  placeholder="Alamat lengkap"
                ></textarea>
              </div>
              <div class="form-group">
                <label for="supervisor">Pembimbing</label>
                <select
                  id="supervisor"
                  class="input-field"
                  bind:value={supervisor_id}
                >
                  <option value="">-- Pilih Pembimbing (Opsional) --</option>
                  {#each supervisors as supervisor}
                    <option value={supervisor.id}
                      >{supervisor.name}{supervisor.institution
                        ? ` - ${supervisor.institution}`
                        : ""}</option
                    >
                  {/each}
                </select>
              </div>

              <div class="divider"><span>Periode Magang</span></div>
              <div class="grid-2">
                <div class="form-group">
                  <label for="startDate"
                    >Tanggal Mulai <span class="text-error">*</span></label
                  >
                  <input
                    id="startDate"
                    type="date"
                    bind:value={startDate}
                    class="input-field"
                    required
                  />
                </div>
                <div class="form-group">
                  <label for="endDate"
                    >Tanggal Selesai <span class="text-error">*</span></label
                  >
                  <input
                    id="endDate"
                    type="date"
                    bind:value={endDate}
                    class="input-field"
                    required
                  />
                </div>
              </div>
            </div>
          {/if}

          {#if role === "supervisor"}
            <div class="divider"><span>Data Pembimbing</span></div>
            <div class="fade-in">
              <div class="grid-2">
                <div class="form-group">
                  <label for="nip">NIP</label>
                  <input
                    id="nip"
                    type="text"
                    bind:value={nip}
                    class="input-field"
                    placeholder="Nomor Induk Pegawai"
                  />
                </div>
                <div class="form-group">
                  <label for="supervisorPhone">WhatsApp / Telepon</label>
                  <input
                    id="supervisorPhone"
                    type="text"
                    bind:value={supervisor_phone}
                    class="input-field"
                    placeholder="08xxxxxxxxxx"
                  />
                </div>
              </div>
              <div class="form-group">
                <label for="institution"
                  >Asal Instansi <span class="text-error">*</span></label
                >
                <input
                  id="institution"
                  type="text"
                  bind:value={institution}
                  class="input-field"
                  placeholder="Nama perusahaan / lembaga"
                  required
                />
              </div>
              <div class="form-group">
                <label for="supervisorAddress">Alamat</label>
                <textarea
                  id="supervisorAddress"
                  bind:value={supervisor_address}
                  class="input-field"
                  rows="2"
                  placeholder="Alamat lengkap"
                ></textarea>
              </div>
            </div>
          {/if}

          <div class="form-actions mt-6">
            <button type="submit" class="btn btn-primary" disabled={loading}>
              {#if loading}
                <span class="spinner"></span> Memproses...
              {:else}
                Kirim Pendaftaran
              {/if}
            </button>
          </div>

          <div style="text-align: center; margin-top: 1.5rem;">
            <a href="/login" use:route class="link-blue">
              Sudah punya akun? Masuk
            </a>
          </div>
        </form>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* Basic Reset & Container */
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background-color: #f3f4f6;
    padding: 1.5rem;
    font-family:
      "Inter",
      -apple-system,
      sans-serif;
  }

  /* Initial Loading Screen */
  .initial-loader {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
  }

  .brand-logo-large {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    background-color: #eff6ff;
    color: #111827;
    border-radius: 16px;
    animation: pulse 2s ease-in-out infinite;
  }

  .brand-logo-large svg {
    width: 36px;
    height: 36px;
  }

  .spinner-large {
    width: 32px;
    height: 32px;
    border: 3px solid #e5e7eb;
    border-top: 3px solid #111827;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.05);
      opacity: 0.9;
    }
  }

  /* Card Style */
  .card {
    background: white;
    width: 100%;
    max-width: 500px; /* Slightly wider for register */
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
    color: #111827;
    border-radius: 10px;
    margin-bottom: 1rem;
  }

  .success-logo {
    background-color: #ecfdf5;
    color: #059669;
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
    font-size: 0.85rem;
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
    box-sizing: border-box;
    font-family: inherit;
    background: #fff;
  }

  .input-field:focus {
    outline: none;
    border-color: #2563eb;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
  }
  .input-field:disabled {
    background: #f9fafb;
    cursor: not-allowed;
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
    font-size: 0.75rem;
    text-transform: uppercase;
    font-weight: 600;
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
    animation: fadeIn 0.3s ease-out forwards;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Grid helper */
  .grid-2 {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0;
  }
  @media (min-width: 640px) {
    .grid-2 {
      grid-template-columns: 1fr 1fr;
      gap: 1rem;
    }
  }

  .text-error {
    color: #dc2626;
  }
  .mt-6 {
    margin-top: 1.5rem;
  }

  /* Role Selector */
  .role-selector {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  .role-option {
    background: #f9fafb;
    border: 2px solid #e5e7eb;
    border-radius: 10px;
    padding: 1rem;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s;
  }
  .role-option:hover {
    border-color: #d1d5db;
    background: #f3f4f6;
  }
  .role-option.active {
    border-color: #2563eb;
    background: #eff6ff;
  }
  .role-icon {
    color: #6b7280;
    margin-bottom: 0.5rem;
    display: flex;
    justify-content: center;
  }
  .role-option.active .role-icon {
    color: #2563eb;
  }
  .role-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .role-title {
    font-weight: 600;
    font-size: 0.9rem;
    color: #374151;
  }
  .role-option.active .role-title {
    color: #1e40af;
  }
  .role-desc {
    font-size: 0.75rem;
    color: #9ca3af;
  }
</style>
