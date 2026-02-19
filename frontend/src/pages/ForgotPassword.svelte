<script>
  import { api } from "../lib/api.js";
  import { goto } from "@mateothegreat/svelte5-router";

  let email = $state("");
  let message = $state("");
  let resetUrl = $state("");
  let loading = $state(false);

  async function handleSubmit() {
    message = "";
    resetUrl = "";
    loading = true;
    try {
      const res = await api.requestPasswordReset(email);
      message =
        res.message || "Jika email terdaftar, tautan reset akan dikirim.";
      if (res.data?.reset_url) resetUrl = res.data.reset_url;
    } catch (err) {
      message = err.message || "Gagal mengirim permintaan reset.";
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
      <h2 class="title">Reset Password</h2>
      <p class="subtitle">Masukkan email Anda untuk menerima tautan reset.</p>
    </div>

    <!-- Message Alert -->
    {#if message}
      <div class="alert {message.includes('Gagal') ? 'error' : 'success'}">
        {#if message.includes("Gagal")}
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
      <div class="form-group">
        <label for="email">Email</label>
        <input
          class="input-field"
          type="email"
          id="email"
          bind:value={email}
          placeholder="nama@sekolah.com"
          onkeydown={(e) => e.key === "Enter" && handleSubmit()}
        />
      </div>

      <button
        class="btn btn-primary mt-4"
        onclick={handleSubmit}
        disabled={loading}
      >
        {#if loading}
          <span class="spinner"></span> Mengirim...
        {:else}
          Kirim Link Reset
        {/if}
      </button>

      {#if resetUrl}
        <div class="alert warning mt-4">
          <div style="font-size: 12px; word-break: break-all;">
            <strong>Dev link:</strong>
            <a href={resetUrl} style="color: inherit;">{resetUrl}</a>
          </div>
        </div>
      {/if}

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

  .alert.warning {
    background-color: #fffbeb;
    color: #d97706;
    border: 1px solid #fde68a;
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
