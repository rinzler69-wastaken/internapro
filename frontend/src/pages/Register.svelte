<script>
  import { onMount } from "svelte";
  import { replace, route } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";

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

  let supervisors = $state([]);

  onMount(async () => {
    // Attempt to hydrate user if token exists but user is null
    if (auth.token && !auth.user) {
      try {
        const res = await api.getProfile();
        auth.user = res.data;
      } catch (e) {
        console.error("Failed to fetch profile in register", e);
        // Don't logout immediately, maybe token is valid but profile fetch failed?
        // But likely invalid.
      }
    }

    if (auth.token && auth.user) {
      if (auth.user.role === "new_user") {
        // Allow stay, pre-fill
        name = auth.user.name || "";
        email = auth.user.email || "";
        isGoogleRedirect = true;
      } else {
        replace("/dashboard");
        return;
      }
    }

    const params = new URLSearchParams(window.location.search);
    const oauthStatus = params.get("oauth");
    if (
      oauthStatus === "google_unregistered" ||
      params.get("status") === "unregistered"
    ) {
      isGoogleRedirect = true;
      // If we have token/user, we prefer that data. Otherwise use params if avail.
      if (!email) email = params.get("email") || "";
      if (!name) name = params.get("name") || "";
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

<div class="register-container">
  <div class="register-card">
    {#if isSubmitted}
      <!-- TAMPILAN SUKSES SUBMIT -->
      <div class="card p-0 overflow-hidden text-center fade-in">
        <div class="p-6 md:p-8">
          <div
            class="w-16 h-16 mx-auto mb-4 rounded-2xl flex items-center justify-center text-2xl bg-green-100 text-green-600"
          >
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
          <h1 class="text-xl font-bold text-slate-800 mb-2">
            Pendaftaran Terkirim
          </h1>
          <p class="text-slate-500 text-sm">
            Akun Anda akan segera kami proses.
          </p>
        </div>

        <div class="p-6 md:p-8 bg-slate-50 border-t border-slate-200">
          <div class="p-4 rounded-xl bg-yellow-50 border border-yellow-200">
            <div class="flex gap-3">
              <svg
                class="text-yellow-500 mt-0.5 flex-shrink-0"
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
              <p class="text-sm text-yellow-800">
                Setelah mendaftar, akun Anda akan menunggu persetujuan dari
                admin. Anda akan menerima notifikasi setelah akun diaktifkan.
              </p>
            </div>
          </div>
          <button
            class="btn btn-secondary w-full mt-6"
            onclick={() => replace("/login")}
          >
            Kembali ke Login
          </button>
        </div>
      </div>
    {:else}
      <!-- FORM PENDAFTARAN -->
      <form class="card p-0 overflow-hidden" onsubmit={handleRegister}>
        <!-- Header -->
        <div class="form-section text-center bg-gray-50">
          <div
            class="w-16 h-16 mx-auto mb-4 rounded-2xl flex items-center justify-center text-2xl text-white bg-gradient-to-br from-purple-500 to-indigo-500 shadow-lg shadow-indigo-500/30"
          >
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
              ><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle
                cx="9"
                cy="7"
                r="4"
              /><line x1="22" x2="18" y1="8" y2="8" /><line
                x1="20"
                x2="20"
                y1="6"
                y2="10"
              /></svg
            >
          </div>
          <h1 class="text-xl font-extrabold text-slate-800 mb-1 tracking-tight">
            Pendaftaran
          </h1>
          <p class="text-slate-500 text-sm">
            Pilih jenis akun dan isi data lengkap untuk mendaftar
          </p>
        </div>

        {#if error}
          <div class="alert error mx-6 mt-4">
            <span>{error}</span>
          </div>
        {/if}

        <!-- Role Selector -->
        <div class="form-section">
          <div class="section-header">
            <div class="section-icon bg-purple-100 text-purple-600">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                ><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle
                  cx="9"
                  cy="7"
                  r="4"
                /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path
                  d="M16 3.13a4 4 0 0 1 0 7.75"
                /></svg
              >
            </div>
            <div>
              <h4 class="font-bold text-slate-800 text-base">Daftar Sebagai</h4>
              <p class="text-sm text-slate-500">
                Pilih jenis akun yang ingin Anda buat
              </p>
            </div>
          </div>

          <div class="role-selector">
            <button
              type="button"
              class="role-option"
              class:active={role === "intern"}
              onclick={() => (role = "intern")}
            >
              <div class="icon-wrapper intern" class:active={role === "intern"}>
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
                  ><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" /><path
                    d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"
                  /></svg
                >
              </div>
              <div class="font-semibold text-slate-700 text-sm">
                Peserta Magang
              </div>
              <div class="text-xs text-slate-500 mt-1">Siswa / Mahasiswa</div>
            </button>
            <button
              type="button"
              class="role-option"
              class:active={role === "supervisor"}
              onclick={() => (role = "supervisor")}
            >
              <div
                class="icon-wrapper supervisor"
                class:active={role === "supervisor"}
              >
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
              <div class="font-semibold text-slate-700 text-sm">Pembimbing</div>
              <div class="text-xs text-slate-500 mt-1">Mentor / Guru</div>
            </button>
          </div>
        </div>

        <div class="section-divider"></div>

        <!-- Section: Informasi Akun -->
        <div class="form-section">
          <div class="section-header">
            <div class="section-icon bg-purple-100 text-purple-600">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
            </div>
            <div>
              <h4 class="font-bold text-slate-800 text-base">Informasi Akun</h4>
              <p class="text-sm text-slate-500">
                Kredensial untuk login ke sistem
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="form-group mb-0 sm:col-span-2">
              <label class="form-label" for="fullName"
                >Nama Lengkap <span class="text-rose-500">*</span></label
              >
              <input
                id="fullName"
                type="text"
                bind:value={name}
                class="input-field"
                placeholder="Masukkan nama lengkap"
                required
              />
            </div>

            <div class="form-group mb-0 sm:col-span-2">
              <label class="form-label" for="email"
                >Email <span class="text-rose-500">*</span></label
              >
              <input
                id="email"
                type="email"
                bind:value={email}
                class="input-field"
                placeholder="email@example.com"
                required
                disabled={isGoogleRedirect && email.length > 0}
              />
            </div>

            <div class="form-group mb-0">
              <label class="form-label" for="password"
                >Password <span class="text-rose-500">*</span></label
              >
              <div class="password-wrapper">
                <input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  bind:value={password}
                  class="input-field"
                  placeholder="Minimal 8 karakter"
                  required
                />
                <button
                  type="button"
                  class="toggle-password"
                  aria-label={showPassword
                    ? "Sembunyikan password"
                    : "Tampilkan password"}
                  title={showPassword
                    ? "Sembunyikan password"
                    : "Tampilkan password"}
                  onclick={() => (showPassword = !showPassword)}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><path
                      d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                    /><circle cx="12" cy="12" r="3" /></svg
                  >
                </button>
              </div>
            </div>

            <div class="form-group mb-0">
              <label class="form-label" for="confirmPassword"
                >Konfirmasi Password <span class="text-rose-500">*</span></label
              >
              <input
                id="confirmPassword"
                type="password"
                bind:value={confirmPassword}
                class="input-field"
                placeholder="Ulangi password"
                required
              />
            </div>
          </div>
        </div>

        <div class="section-divider"></div>

        <!-- Section: Data Sesuai Role -->
        {#if role === "intern"}
          <div class="form-section intern-fields fade-in">
            <div class="section-header">
              <div class="section-icon bg-green-100 text-green-600">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
              </div>
              <div>
                <h4 class="font-bold text-slate-800 text-base">
                  Data Peserta Magang
                </h4>
                <p class="text-sm text-slate-500">
                  Informasi tentang diri Anda
                </p>
              </div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="form-group mb-0">
                <label class="form-label" for="nis">NISN / NIM</label>
                <input
                  id="nis"
                  type="text"
                  bind:value={nis}
                  class="input-field"
                  placeholder="Nomor induk"
                />
              </div>
              <div class="form-group mb-0">
                <label class="form-label" for="phone">WhatsApp / Telepon</label>
                <input
                  id="phone"
                  type="text"
                  bind:value={phone}
                  class="input-field"
                  placeholder="08xxxxxxxxxx"
                />
              </div>
              <div class="form-group mb-0">
                <label class="form-label" for="school"
                  >Asal Sekolah / Kampus <span class="text-rose-500">*</span
                  ></label
                >
                <input
                  id="school"
                  type="text"
                  bind:value={school}
                  class="input-field"
                  placeholder="Nama instansi pendidikan"
                  required
                />
              </div>
              <div class="form-group mb-0">
                <label class="form-label" for="department"
                  >Jurusan / Bidang Studi <span class="text-rose-500">*</span
                  ></label
                >
                <input
                  id="department"
                  type="text"
                  bind:value={department}
                  class="input-field"
                  placeholder="Contoh: RPL, TKJ"
                  required
                />
              </div>
              <div class="form-group mb-0 sm:col-span-2">
                <label class="form-label" for="address">Alamat</label>
                <textarea
                  id="address"
                  bind:value={address}
                  class="input-field"
                  rows="2"
                  placeholder="Alamat tempat tinggal saat ini"
                ></textarea>
              </div>
              <div class="form-group mb-0 sm:col-span-2">
                <label class="form-label" for="supervisor">Pembimbing</label>
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
            </div>

            <div class="section-header mt-6">
              <div class="section-icon bg-green-100 text-green-600">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  ><rect
                    x="3"
                    y="4"
                    width="18"
                    height="18"
                    rx="2"
                    ry="2"
                  /><line x1="16" x2="16" y1="2" y2="6" /><line
                    x1="8"
                    x2="8"
                    y1="2"
                    y2="6"
                  /><line x1="3" x2="21" y1="10" y2="10" /></svg
                >
              </div>
              <div>
                <h4 class="font-bold text-slate-800 text-base">
                  Periode Magang
                </h4>
                <p class="text-sm text-slate-500">Rencana durasi magang Anda</p>
              </div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="form-group mb-0">
                <label class="form-label" for="startDate"
                  >Tanggal Mulai <span class="text-rose-500">*</span></label
                >
                <input
                  id="startDate"
                  type="date"
                  bind:value={startDate}
                  class="input-field"
                  required
                />
              </div>
              <div class="form-group mb-0">
                <label class="form-label" for="endDate"
                  >Tanggal Selesai <span class="text-rose-500">*</span></label
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
          <div class="form-section supervisor-fields fade-in">
            <div class="section-header">
              <div class="section-icon bg-blue-100 text-blue-600">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  ><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle
                    cx="8.5"
                    cy="7"
                    r="4"
                  /><path d="M20 8v6M23 11h-6" /></svg
                >
              </div>
              <div>
                <h4 class="font-bold text-slate-800 text-base">
                  Data Pembimbing
                </h4>
                <p class="text-sm text-slate-500">
                  Informasi tentang diri Anda sebagai pembimbing
                </p>
              </div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="form-group mb-0">
                <label class="form-label" for="nip">NIP</label>
                <input
                  id="nip"
                  type="text"
                  bind:value={nip}
                  class="input-field"
                  placeholder="Nomor Induk Pegawai"
                />
              </div>
              <div class="form-group mb-0">
                <label class="form-label" for="supervisorPhone"
                  >WhatsApp / Telepon</label
                >
                <input
                  id="supervisorPhone"
                  type="text"
                  bind:value={supervisor_phone}
                  class="input-field"
                  placeholder="08xxxxxxxxxx"
                />
              </div>
              <div class="form-group mb-0 sm:col-span-2">
                <label class="form-label" for="institution"
                  >Asal Instansi <span class="text-rose-500">*</span></label
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
              <div class="form-group mb-0 sm:col-span-2">
                <label class="form-label" for="supervisorAddress">Alamat</label>
                <textarea
                  id="supervisorAddress"
                  bind:value={supervisor_address}
                  class="input-field"
                  rows="2"
                  placeholder="Alamat tempat tinggal saat ini"
                ></textarea>
              </div>
            </div>
          </div>
        {/if}

        <!-- Footer -->
        <div class="form-section pt-0">
          <div
            class="p-4 rounded-xl mb-5 bg-purple-50 border border-purple-200/80"
          >
            <div class="flex gap-3">
              <svg
                class="text-purple-500 mt-0.5 flex-shrink-0"
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
              <p class="text-sm text-slate-700">
                Setelah mendaftar, akun Anda akan menunggu persetujuan dari
                admin. Anda akan menerima notifikasi setelah akun diaktifkan.
              </p>
            </div>
          </div>
          <button
            type="submit"
            class="btn btn-primary w-full py-3 text-sm"
            disabled={loading}
          >
            {#if loading}
              <span class="spinner"></span> Memproses...
            {:else}
              Kirim Pendaftaran
            {/if}
          </button>
          <p class="text-center text-sm text-slate-500 mt-5">
            Sudah punya akun?
            <a
              href="/login"
              use:route
              class="text-purple-600 hover:text-purple-700 font-medium"
            >
              Masuk di sini
            </a>
          </p>
        </div>
      </form>
    {/if}
  </div>
</div>

<style>
  .register-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
    background-color: #f1f5f9;
  }
  .register-card {
    width: 100%;
    max-width: 640px;
  }
  .card {
    background-color: white;
    border-radius: 1.5rem; /* 24px */
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  }
  .form-section {
    padding: 1.5rem;
  }
  @media (min-width: 640px) {
    .form-section {
      padding: 2rem;
    }
  }
  .section-divider {
    height: 8px;
    background-color: #f1f5f9;
  }
  .section-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding-bottom: 1rem;
    margin-bottom: 1.25rem;
    border-bottom: 1px solid #e2e8f0;
  }
  .section-icon {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
  }

  /* Role Selector */
  .role-selector {
    display: flex;
    gap: 1rem;
  }
  .role-option {
    flex: 1;
    border: 2px solid #e2e8f0;
    border-radius: 12px;
    padding: 1rem;
    cursor: pointer;
    transition: all 0.2s;
    text-align: center;
    background-color: transparent;
    font-family: inherit;
  }
  .role-option:hover {
    border-color: #c7d2fe;
    background-color: #fafbff;
  }
  .role-option.active {
    border-color: #818cf8;
    background: #eef2ff;
    box-shadow: 0 0 0 3px #c7d2fe;
  }
  .icon-wrapper {
    width: 48px;
    height: 48px;
    margin: 0 auto 0.75rem;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    transition: all 0.2s;
  }
  .icon-wrapper.intern {
    background-color: #dcfce7;
    color: #16a34a;
  }
  .icon-wrapper.supervisor {
    background-color: #dbeafe;
    color: #2563eb;
  }
  .icon-wrapper.active.intern {
    background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
    color: white;
  }
  .icon-wrapper.active.supervisor {
    background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    color: white;
  }

  /* Forms */
  .form-group {
    margin-bottom: 1.25rem;
  }
  .form-label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #334155;
    margin-bottom: 0.5rem;
  }
  .input-field {
    width: 100%;
    padding: 0.625rem 1rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    font-size: 0.95rem;
    transition: all 0.2s;
    box-sizing: border-box;
  }
  .input-field:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
  }
  .password-wrapper {
    position: relative;
  }
  .toggle-password {
    position: absolute;
    right: 1px;
    top: 1px;
    bottom: 1px;
    padding: 0 12px;
    background: none;
    border: none;
    color: #64748b;
    cursor: pointer;
    display: flex;
    align-items: center;
    border-radius: 0 7px 7px 0;
  }

  /* Buttons */
  .btn {
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    border: 1px solid transparent;
  }
  .btn-primary {
    background-color: #4f46e5;
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    background-color: #4338ca;
  }
  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  .btn-secondary {
    background-color: #e2e8f0;
    color: #1e293b;
  }
  .btn-secondary:hover {
    background-color: #cbd5e1;
  }

  .alert.error {
    background-color: #fef2f2;
    color: #dc2626;
    padding: 0.75rem 1.25rem;
    border-radius: 0.5rem;
    font-size: 0.9rem;
  }

  /* Utils */
  .fade-in {
    animation: fadeIn 0.5s ease-out;
  }
  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
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
</style>
