<script>
  import { onMount } from "svelte";
  import { auth } from "../lib/auth.svelte.js";
  import { api } from "../lib/api.js";

  // --- STATE ---
  let loading = $state(true);
  let saving = $state(false);
  let savingPassword = $state(false);

  // Profile State
  let avatarFile = $state(null);
  let avatarPreview = $state(null);
  let fileInput = $state(null); // Referensi elemen input
  let form = $state({ name: "", email: "" });

  // Password State
  let showCurrentPass = $state(false);
  let showNewPass = $state(false);
  let passwordForm = $state({
    current_password: "",
    password: "",
    password_confirmation: "",
  });

  onMount(async () => {
    await fetchProfile();
  });

  async function fetchProfile() {
    loading = true;
    try {
      const res = await api.getProfile();
      // Handle berbagai kemungkinan struktur response backend
      const userData = res.data?.user || res.data || auth.user || {};
      if (auth.hydrate) auth.hydrate(userData);

      form = {
        name: userData.name || "",
        email: userData.email || "",
      };

      if (userData.avatar) {
        avatarPreview = buildUploadUrl(userData.avatar);
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function handleFileSelect(e) {
    const file = e.target.files?.[0];
    if (file) {
      // Validasi Ukuran (2MB)
      if (file.size > 2 * 1024 * 1024) {
        alert("Ukuran file terlalu besar. Maksimal 2MB.");
        return;
      }

      avatarFile = file;

      // Preview Lokal (Langsung muncul sebelum upload ke server)
      const reader = new FileReader();
      reader.onload = (e) => (avatarPreview = e.target.result);
      reader.readAsDataURL(file);
    }
  }

  function triggerFileInput() {
    fileInput.click();
  }

  async function saveProfile() {
    saving = true;
    try {
      const formData = new FormData();
      formData.append("name", form.name);
      formData.append("email", form.email);

      if (avatarFile) {
        formData.append("avatar", avatarFile);
      }

      await api.updateProfile(formData);

      // Refresh data dari server untuk memastikan
      await fetchProfile();

      alert("Profil berhasil diperbarui!");

      // Reset input file
      avatarFile = null;
      if (fileInput) fileInput.value = "";
    } catch (err) {
      console.error(err);
      alert(err.message || "Gagal menyimpan profil");
    } finally {
      saving = false;
    }
  }

  async function updatePassword() {
    if (passwordForm.password !== passwordForm.password_confirmation) {
      alert("Konfirmasi password tidak cocok");
      return;
    }

    savingPassword = true;
    try {
      await api.updatePassword(passwordForm);
      passwordForm = {
        current_password: "",
        password: "",
        password_confirmation: "",
      };
      alert(
        "Password berhasil diperbarui. Silakan login ulang jika diperlukan.",
      );
    } catch (err) {
      alert(err.message || "Gagal memperbarui password");
    } finally {
      savingPassword = false;
    }
  }

  function buildUploadUrl(path) {
    if (!path) return "";
    const clean = path.startsWith("/uploads/") ? path : `/uploads/${path}`;
    const params = [];
    if (auth.token) params.push(`token=${auth.token}`);
    params.push(`t=${Date.now()}`);
    return `${clean}?${params.join("&")}`;
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link
    rel="preconnect"
    href="https://fonts.gstatic.com"
    crossorigin="anonymous"
  />
  <link
    href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"
    rel="stylesheet"
  />
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0"
  />
</svelte:head>

<div class="page-bg">
  <div class="container animate-fade-in">
    <!-- Header -->
    <div class="header">
      <div class="title-stack">
        <h2 class="title">Edit Profil</h2>
        <!-- <p class="subtitle">Perbarui informasi akun & keamanan Anda.</p> -->
      </div>
      <div class="actions">
        <a href="/profile" class="btn ghost">
          <span class="material-symbols-outlined">arrow_left_alt</span>
          Kembali
        </a>

        <a href="/dashboard" class="btn ghost">
          <span class="material-symbols-outlined">dashboard</span>
          Dashboard
        </a>
      </div>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat profil...</p>
      </div>
    {:else}
      <div class="grid-layout animate-slide-up">
        <!-- KOLOM KIRI: EDIT PROFIL -->
        <div class="card profile-card">
          <div class="card-header">
            <div class="icon-circle bg-emerald">
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                ><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle
                  cx="12"
                  cy="7"
                  r="4"
                /></svg
              >
            </div>
            <h3>Profil Saya</h3>
          </div>

          <div class="card-body">
            <!-- Avatar Section -->
            <div class="avatar-section">
              <div class="avatar-container">
                {#if avatarPreview}
                  <img src={avatarPreview} alt="Avatar" class="avatar-img" />
                {:else}
                  <div class="avatar-placeholder">
                    {form.name?.charAt(0) || "U"}
                  </div>
                {/if}

                <!-- Edit Overlay Button -->
                <button
                  class="avatar-btn"
                  onclick={triggerFileInput}
                  title="Ganti Foto"
                >
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    ><path
                      d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"
                    /><circle cx="12" cy="13" r="4" /></svg
                  >
                </button>

                <!-- Hidden Input -->
                <input
                  bind:this={fileInput}
                  type="file"
                  accept="image/png, image/jpeg, image/jpg"
                  hidden
                  onchange={handleFileSelect}
                />
              </div>
              <div class="avatar-text">
                <h4 class="font-bold text-slate-800">Foto Profil</h4>
                <p class="text-sm text-slate-500">
                  Klik ikon kamera untuk mengubah.<br />Maksimal 2MB (JPG/PNG).
                </p>
              </div>
            </div>

            <div class="divider"></div>

            <!-- Form Inputs -->
            <div class="form-stack pb-8">
              <div class="form-group">
                <label class="label" for="profile-name">Nama Lengkap</label>
                <div class="input-wrapper">
                  <div class="input-icon">
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      ><path
                        d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"
                      /><circle cx="12" cy="7" r="4" /></svg
                    >
                  </div>
                  <input
                    class="input-field"
                    id="profile-name"
                    bind:value={form.name}
                    placeholder="Nama Anda"
                  />
                </div>
              </div>
              <div class="form-group">
                <label class="label" for="profile-email">Alamat Email</label>
                <div class="input-wrapper">
                  <div class="input-icon">
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      ><path
                        d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"
                      /><polyline points="22,6 12,13 2,6" /></svg
                    >
                  </div>
                  <input
                    class="input-field"
                    id="profile-email"
                    type="email"
                    bind:value={form.email}
                    placeholder="email@contoh.com"
                  />
                </div>
              </div>
            </div>

            <div class="action-area">
              <button
                class="btn-primary w-full"
                onclick={saveProfile}
                disabled={saving}
              >
                <span class="material-symbols-outlined">save</span>
                {#if saving}
                  <div class="spinner-small"></div>
                   Menyimpan...
                {:else}
                  Simpan Perubahan
                {/if}
              </button>
            </div>
          </div>
        </div>

        <!-- KOLOM KANAN: KEAMANAN -->
        <div class="card security-card">
          <div class="card-header">
            <div class="icon-circle bg-blue">
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                ><rect x="3" y="11" width="18" height="11" rx="2" ry="2" /><path
                  d="M7 11V7a5 5 0 0 1 10 0v4"
                /></svg
              >
            </div>
            <h3>Ganti Password</h3>
          </div>

          <div class="card-body">
            <p class="text-sm text-slate-500 mb-6">
              Pastikan password Anda kuat dan aman. Minimal 8 karakter.
            </p>

            <div class="form-stack pb-8">
              <div class="form-group">
                <label class="label" for="curr-pass">Password Saat Ini</label>
                <div class="input-wrapper">
                  <input
                    class="input-field pr-10"
                    id="curr-pass"
                    type={showCurrentPass ? "text" : "password"}
                    bind:value={passwordForm.current_password}
                    placeholder="••••••"
                  />
                  <button
                    class="toggle-btn"
                    onclick={() => (showCurrentPass = !showCurrentPass)}
                    tabindex="-1"
                  >
                    {#if showCurrentPass}
                      <svg
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

              <div class="form-stack">
                <div class="form-group">
                  <label class="label" for="new-pass">Password Baru</label>
                  <div class="input-wrapper">
                    <input
                      class="input-field pr-10"
                      id="new-pass"
                      type={showNewPass ? "text" : "password"}
                      bind:value={passwordForm.password}
                      placeholder="••••••"
                    />
                    <button
                      class="toggle-btn"
                      onclick={() => (showNewPass = !showNewPass)}
                      tabindex="-1"
                    >
                      {#if showNewPass}
                        <svg
                          width="16"
                          height="16"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          ><path
                            d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                          ></path><line x1="1" y1="1" x2="23" y2="23"
                          ></line></svg
                        >
                      {:else}
                        <svg
                          width="16"
                          height="16"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          ><path
                            d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                          ></path><circle cx="12" cy="12" r="3"></circle></svg
                        >
                      {/if}
                    </button>
                  </div>
                </div>
                <div class="form-group">
                  <label class="label" for="confirm-pass">Konfirmasi</label>
                  <div class="input-wrapper">
                    <input
                      class="input-field"
                      id="confirm-pass"
                      type="password"
                      bind:value={passwordForm.password_confirmation}
                      placeholder="••••••"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div class="action-area mt-8">
              <button
                class="btn-blue w-full"
                onclick={updatePassword}
                disabled={savingPassword}
              >
                <span class="material-symbols-outlined">key</span>
                {#if savingPassword}
                  <div class="spinner-small dark"></div>
                   Memproses...
                {:else}
                  Perbarui Password
                {/if}
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(body) {
    font-family:
      "Plus Jakarta Sans",
      "Inter",
      system-ui,
      -apple-system,
      sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    /* background:
      radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.06) 0, transparent 45%),
      radial-gradient(at 100% 20%, rgba(16, 185, 129, 0.07) 0, transparent 35%),
      #f8fafc; */
    padding: 0px;
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* HEADER */
  .header {
    margin-bottom: 0px;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 8px;
  }

  .title-stack h2.title {
    font-size: 20px;
    font-weight: 600;
    color: #0f172a;
    margin: 0;
    letter-spacing: -0.02em;
  }

  .title-stack {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  /* LAYOUT */
  .grid-layout {
    display: grid;
    grid-template-columns: 1fr;
    gap: 16px;
  }

  @media (min-width: 850px) {
    .grid-layout {
      grid-template-columns: 1fr 1fr;
    }
  }

  /* CARDS */
  .card {
    background: rgba(255, 255, 255, 0.9);
    border-radius: 20px;
    border: 1px solid oklch(92.9% 0.013 255.508);
    box-shadow: 0 10px 30px rgba(15, 23, 42, 0.04);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    height: 100%;
    transition:
      transform 0.3s ease,
      box-shadow 0.3s ease;
  }

  .card:hover {
    /* transform: translateY(-4px);  */
    box-shadow: 0 10px 20px -5px rgba(0, 0, 0, 0.05);
    transition: box-shadow 0.3s ease;
  }

  .card-header {
    padding: 20px 22px;
    border-bottom: 1px solid #f1f5f9;
    display: flex;
    align-items: center;
    gap: 12px;
    background: #ffffff;
  }

  .card-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
  }

  .icon-circle {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .bg-emerald {
    background: #ecfdf5;
    color: #059669;
  }

  .bg-blue {
    background: #eff6ff;
    color: #2563eb;
  }

  .card-body {
    padding: 22px;
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .divider {
    height: 1px;
    background: #f1f5f9;
    margin: 20px 0;
  }

  /* AVATAR SECTION */
  .avatar-section {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .avatar-container {
    position: relative;
    width: 78px;
    height: 78px;
    flex-shrink: 0;
  }

  .avatar-img {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    border: 4px solid #ffffff;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    background: #0f172a;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    font-weight: 800;
    border: 4px solid #ffffff;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
  }

  .avatar-btn {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: #10b981;
    color: white;
    border: 3px solid white;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: transform 0.2s;
  }

  .avatar-btn:hover {
    transform: scale(1.1);
    background: #059669;
  }

  .avatar-text h4 {
    margin: 0 0 4px 0;
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .avatar-text p {
    font-size: 12px;
    color: #64748b;
    margin: 0;
  }

  /* FORM ELEMENTS */
  .form-stack {
    display: flex;
    flex-direction: column;
    gap: 14px;
    flex: 1;
  }

  .form-group {
    margin-bottom: 0;
  }

  .label {
    display: block;
    font-size: 12px;
    font-weight: 700;
    color: #64748b;
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .input-icon {
    position: absolute;
    left: 14px;
    color: #94a3b8;
    pointer-events: none;
    display: flex;
    align-items: center;
  }

  .input-field {
    width: 100%;
    padding: 12px 14px 12px 44px;
    border: 1px solid #cbd5e1;
    border-radius: 12px;
    font-size: 14px;
    color: #0f172a;
    transition: all 0.2s;
    background: #fff;
  }

  /* Jika tidak ada icon, padding kiri normal */
  .security-card .input-field:not(.pr-10) {
    padding-left: 14px;
  }

  /* Jika ada toggle button di kanan */
  .input-field.pr-10 {
    padding-right: 40px;
    padding-left: 14px;
  }

  .input-field:focus {
    outline: none;
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  }

  .toggle-btn {
    position: absolute;
    right: 10px;
    background: none;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    display: flex;
    align-items: center;
    padding: 6px;
    border-radius: 8px;
    transition: color 0.2s;
  }

  .toggle-btn:hover {
    color: #475569;
    background: #f1f5f9;
  }

  /* BUTTONS */
  .action-area {
    margin-top: auto;
  }

  .btn {
    padding: 10px 18px;
    border-radius: 999px;
    font-weight: 700;
    font-size: 14px;
    border: 1px solid transparent;
    cursor: pointer;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition:
      transform 0.12s ease,
      box-shadow 0.12s ease,
      background 0.12s ease;
  }

  .btn.ghost {
    background: #fff;
    border: 2px solid #e2e8f0;
    color: #0f172a;
    font-weight: 600;
  }

  .btn:hover:not(:disabled) {
    transform: translateY(-1px);
  }

  .btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: #fff;
    border: 1px solid #10b981;
    padding: 12px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    align-self: center;
    gap: 8px;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    transition: all 0.2s;
  }

  .btn-primary:hover:not(:disabled),
  .btn-blue:hover:not(:disabled) {
    transform:
      box-shadow 0.2s,
      background 0.2s;
    box-shadow: 0 6px 20px rgba(16, 185, 129, 0.2);
  }

  .btn-blue {
    background: linear-gradient(135deg, #2563eb, #1d4ed8);
    color: #fff;
    border: 1px solid #10b981;
    padding: 12px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
  }

  /* UTILS */
  .w-full {
    width: 100%;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e2e8f0;
    border-top-color: #6366f1;
    border-radius: 50%;
    margin: 0 auto 12px;
    animation: spin 1s linear infinite;
  }

  .spinner-small {
    width: 16px;
    height: 16px;
    border: 2px solid white;
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .spinner-small.dark {
    border-color: #cbd5e1;
    border-top-color: #0f172a;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .loading-state {
    text-align: center;
    padding: 60px;
    color: #94a3b8;
    font-style: italic;
  }

  .animate-fade-in {
    opacity: 0;
    animation: fadeIn 0.6s ease-out forwards;
  }

  .animate-slide-up {
    opacity: 0;
    animation: slideUp 0.6s ease-out forwards;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* MOBILE RESPONSIVE */
  @media (max-width: 900px) {
    .header {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }

    .title-stack {
      text-align: left;
    }

    .actions {
      width: 100%;
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
      justify-content: stretch;
    }

    .actions .btn {
      width: 100%;
      justify-content: center;
    }
  }
</style>
