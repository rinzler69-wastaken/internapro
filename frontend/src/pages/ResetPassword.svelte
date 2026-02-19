<script>
  import { api } from "../lib/api.js";
  import { goto } from "@mateothegreat/svelte5-router";
  import { location } from "../lib/location.js";

  let email = $state("");
  let token = $state("");
  let password = $state("");
  let passwordConfirmation = $state("");
  let message = $state("");
  let loading = $state(false);
  let showPassword = $state(false);

  $effect(() => {
    const search = $location.search;
    if (!search) return;
    const params = new URLSearchParams(search);
    const paramEmail = params.get("email");
    const paramToken = params.get("token");
    if (paramEmail) email = paramEmail;
    if (paramToken) token = paramToken;
  });

  async function handleSubmit() {
    message = "";
    loading = true;
    try {
      await api.resetPassword({
        email,
        token,
        password,
        password_confirmation: passwordConfirmation,
      });
      message = "Password berhasil diperbarui. Silakan login.";
      setTimeout(() => goto("/login"), 1500);
    } catch (err) {
      message = err.message || "Gagal mereset password.";
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-container">
  <div class="card login-card fade-in">
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
      <h2 class="title">Buat Password Baru</h2>
      <p class="subtitle">Masukkan token dan password baru Anda.</p>
    </div>

    <!-- Message Alert -->
    {#if message}
      <div class="alert {message.includes('berhasil') ? 'success' : 'error'}">
        {#if !message.includes("berhasil")}
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
        {/if}
        <span>{message}</span>
      </div>
    {/if}

    <!-- Main Form -->
    <div class="form-body">
      {#if !email || !token}
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
          <span>Tautan reset tidak valid atau tidak lengkap.</span>
        </div>
      {/if}

      <div class="form-group">
        <label for="email-display">Email</label>
        <input
          class="input-field"
          type="email"
          id="email-display"
          value={email}
          disabled
          placeholder="nama@sekolah.com"
        />
        <p class="subtitle mt-1" style="font-size: 11px;">
          Email ini diambil secara otomatis dari tautan reset.
        </p>
      </div>

      <div class="form-group relative">
        <label for="password">Password Baru</label>
        <div class="password-wrapper">
          <input
            class="input-field"
            type={showPassword ? "text" : "password"}
            id="password"
            bind:value={password}
            placeholder="••••••••"
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
      </div>

      <div class="form-group relative">
        <label for="password-confirm">Konfirmasi Password</label>
        <div class="password-wrapper">
          <input
            class="input-field"
            type={showPassword ? "text" : "password"}
            id="password-confirm"
            bind:value={passwordConfirmation}
            placeholder="••••••••"
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
      </div>

      <button
        class="btn btn-primary mt-4"
        onclick={handleSubmit}
        disabled={loading}
      >
        {#if loading}
          <span class="spinner"></span> Menyimpan...
        {:else}
          Simpan Password
        {/if}
      </button>

      <div style="text-align: center; margin-top: 1.5rem;">
        <button class="btn-link" onclick={() => goto("/login")}>
          Kembali ke Login
        </button>
      </div>
    </div>
  </div>
</div>

<style>
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
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

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
    box-sizing: border-box;
  }

  .input-field:focus {
    outline: none;
    border-color: #2563eb;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
  }

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

  .btn-link {
    background: none;
    border: none;
    color: #2563eb;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
  }

  .btn-link:hover {
    text-decoration: underline;
  }

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

  .alert.success {
    background-color: #ecfdf5;
    color: #059669;
    border: 1px solid #a7f3d0;
  }

  .mt-4 {
    margin-top: 1rem;
  }

  .fade-in {
    animation: fadeIn 0.4s ease-out;
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
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
