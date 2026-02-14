<script>
  import { onMount } from "svelte";
  import { replace } from "@mateothegreat/svelte5-router";
  import { api } from "../lib/api.js";

  // State
  let name = $state("");
  let school = $state("");
  let major = $state("");
  let phone = $state("");
  let email = $state("");
  let password = $state("");
  let confirmPassword = $state("");
  let address = $state("");
  let startDate = $state("");
  let endDate = $state("");
  let nis = $state("");
  let supervisorId = $state("");

  let loading = $state(false);
  let error = $state("");

  // Deteksi apakah ini lemparan dari Google Login yang gagal
  let isGoogleError = $state(false);

  // Field visibility toggles
  let showPassword = $state(false);
  let showConfirmPassword = $state(false);

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    const oauthParam = params.get("oauth");
    if (
      params.get("error") === "google_unregistered" ||
      oauthParam === "google_unregistered" ||
      oauthParam === "google"
    ) {
      isGoogleError = true;
    }
    const emailParam = params.get("email");
    const nameParam = params.get("name");
    if (emailParam) email = emailParam;
    if (nameParam) name = nameParam;
  });

  async function handleRegister(e) {
    e?.preventDefault();
    if (password !== confirmPassword) {
      error = "Password tidak cocok";
      return;
    }

    loading = true;
    error = "";

    try {
      await api.post("/internship/register", {
        name,
        email,
        school,
        department: major,
        phone,
        password,
        confirm_password: confirmPassword,
        address,
        start_date: startDate,
        end_date: endDate,
        nis,
        supervisor_id: supervisorId ? Number(supervisorId) : null,
      });

      try {
        await api.logout();
      } catch (e) {}
      replace("/waiting-approval");
    } catch (err) {
      const msg = err.message || "Gagal mendaftar";
      const status = err.response?.status;

      if (status === 409 || msg.toLowerCase().includes("sudah terdaftar")) {
        try {
          const prof = await api.getProfile();
          const intern = prof.data?.intern;
          if (intern?.status === "pending" || intern?.id) {
            alert("Data sudah ada. Menunggu persetujuan admin.");
            replace("/waiting-approval");
            return;
          }
        } catch (_) {
          /* ignore */
        }
        alert("Email sudah terdaftar. Silakan login dengan akun tersebut.");
        replace("/login");
      } else {
        error = msg;
      }
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-container">
  <div class="card fade-in">
    <!-- HEADER -->
    <div class="card-header">
      {#if isGoogleError}
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
          <span
            >Akun Google ini belum terdaftar. Silakan lengkapi data magang untuk
            mendaftar.</span
          >
        </div>
      {/if}

      <div class="brand-logo">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          width="24"
          height="24"
        >
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
          ></path>
          <polyline points="14 2 14 8 20 8"></polyline>
        </svg>
      </div>
      <h2 class="title">Daftar Magang Baru</h2>
      <p class="subtitle">Lengkapi formulir di bawah untuk mendaftar.</p>
    </div>

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

    <form class="form-body" onsubmit={handleRegister}>
      <div class="grid-2">
        <div class="form-group">
          <label for="name">Nama Lengkap</label>
          <input
            class="input-field"
            type="text"
            id="name"
            bind:value={name}
            placeholder="Nama sesuai identitas"
            required
          />
        </div>
        <div class="form-group">
          <label for="email">Email</label>
          <input
            class="input-field"
            type="email"
            id="email"
            bind:value={email}
            placeholder="email@sekolah.com"
            required
          />
        </div>
      </div>

      <div class="grid-2">
        <div class="form-group">
          <label for="school">Sekolah / Kampus</label>
          <input
            class="input-field"
            type="text"
            id="school"
            bind:value={school}
            placeholder="Nama instansi"
            required
          />
        </div>
        <div class="form-group">
          <label for="major">Jurusan</label>
          <input
            class="input-field"
            type="text"
            id="major"
            bind:value={major}
            placeholder="Jurusan"
            required
          />
        </div>
      </div>

      <div class="grid-2">
        <div class="form-group">
          <label for="phone">No. WhatsApp</label>
          <input
            class="input-field"
            type="text"
            id="phone"
            bind:value={phone}
            placeholder="08xxxxxxxx"
          />
        </div>
        <div class="form-group">
          <label for="nis">NIS / NIM</label>
          <input
            class="input-field"
            type="text"
            id="nis"
            bind:value={nis}
            placeholder="Nomor induk"
          />
        </div>
      </div>

      <div class="form-group">
        <label for="address">Alamat</label>
        <textarea
          class="input-field"
          id="address"
          rows="2"
          bind:value={address}
          placeholder="Alamat tempat tinggal"
        ></textarea>
      </div>

      <div class="grid-2">
        <div class="form-group">
          <label for="start">Tanggal Mulai</label>
          <input
            class="input-field"
            type="date"
            id="start"
            bind:value={startDate}
            required
          />
        </div>
        <div class="form-group">
          <label for="end">Tanggal Selesai</label>
          <input
            class="input-field"
            type="date"
            id="end"
            bind:value={endDate}
            required
          />
        </div>
      </div>

      <div class="form-group">
        <label for="supervisor">ID Pembimbing (opsional)</label>
        <input
          class="input-field"
          type="number"
          id="supervisor"
          bind:value={supervisorId}
          placeholder="ID dari pembimbing jika ada"
        />
      </div>

      <div class="grid-2">
        <div class="form-group relative">
          <label for="password">Password</label>
          <div class="password-wrapper">
            <input
              class="input-field"
              type={showPassword ? "text" : "password"}
              id="password"
              bind:value={password}
              placeholder="Minimal 6 karakter"
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
          <label for="confirmPassword">Konfirmasi</label>
          <div class="password-wrapper">
            <input
              class="input-field"
              type={showConfirmPassword ? "text" : "password"}
              id="confirmPassword"
              bind:value={confirmPassword}
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

      <div class="form-actions mt-6">
        <button class="btn btn-primary" type="submit" disabled={loading}>
          {#if loading}
            <span class="spinner"></span> Mengirim...
          {:else}
            Kirim Pendaftaran
          {/if}
        </button>
      </div>

      <div style="text-align: center; margin-top: 1.5rem;">
        <a href="/login" class="link-blue">Kembali ke Login</a>
      </div>
    </form>
  </div>
</div>

<style>
  /* Reuse styling from Register/Login */
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 1.5rem;
    background-color: #f3f4f6;
    font-family: "Inter", sans-serif;
  }
  .card {
    background: white;
    width: 100%;
    max-width: 600px;
    border-radius: 12px;
    padding: 2.5rem 2rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }
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
  .title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin-bottom: 0.5rem;
  }
  .subtitle {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

  .form-group {
    margin-bottom: 1.25rem;
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
    box-sizing: border-box;
    background: #fff;
  }
  .input-field:focus {
    outline: none;
    border-color: #2563eb;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
  }

  .btn {
    width: 100%;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid #04070b;
    transition: all 0.2s;
    display: flex;
    justify-content: center;
    gap: 8px;
  }
  .btn-primary {
    background-color: #fff;
    color: #374151;
  }
  .btn-primary:hover {
    background-color: #111827;
    color: #fff;
  }

  .link-blue {
    color: #2563eb;
    font-size: 0.85rem;
    text-decoration: none;
    font-weight: 500;
  }
  .link-blue:hover {
    text-decoration: underline;
  }

  /* Grid & Helpers */
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
  .mt-6 {
    margin-top: 1.5rem;
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

  /* Password */
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

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #374151;
    border-top: 2px solid transparent;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  .btn-primary:hover .spinner {
    border-color: #fff;
    border-top-color: transparent;
  }
  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
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
