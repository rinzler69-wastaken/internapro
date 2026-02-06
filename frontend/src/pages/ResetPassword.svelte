<script>
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';
  import { location } from '../lib/location.js';

  let email = $state('');
  let token = $state('');
  let password = $state('');
  let passwordConfirmation = $state('');
  let message = $state('');
  let loading = $state(false);
  let showPassword = $state(false);

  $effect(() => {
    const search = $location.search;
    if (!search) return;
    const params = new URLSearchParams(search);
    const paramEmail = params.get('email');
    const paramToken = params.get('token');
    if (paramEmail) email = paramEmail;
    if (paramToken) token = paramToken;
  });

  async function handleSubmit() {
    message = '';
    loading = true;
    try {
      await api.resetPassword({
        email,
        token,
        password,
        password_confirmation: passwordConfirmation,
      });
      message = 'Password berhasil diperbarui. Silakan login.';
      setTimeout(() => goto('/login'), 800);
    } catch (err) {
      message = err.message || 'Gagal mereset password.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="main-content" style="margin-left:0; display:flex; align-items:center; justify-content:center; min-height:100vh;">
  <div class="card" style="max-width:420px; width:100%;">
    <div style="text-align:center; margin-bottom:16px;">
      <h2 class="font-geist" style="margin:12px 0 4px;">Buat Password Baru</h2>
      <p class="text-muted">Masukkan token dan password baru Anda.</p>
    </div>

    {#if message}
      <div class="badge" style="margin-bottom:12px;">{message}</div>
    {/if}

    <div class="form-group">
      <label class="form-label" for="reset-email">Email</label>
      <input class="input" id="reset-email" type="email" bind:value={email} />
    </div>
    <div class="form-group">
      <label class="form-label" for="reset-token">Token</label>
      <input class="input" id="reset-token" bind:value={token} />
    </div>
    <div class="form-group form-group-relative">
      <label class="form-label" for="reset-pass">Password Baru</label>
      <input class="input" id="reset-pass" type="{showPassword ? 'text' : 'password'}" bind:value={password} />
<button class="password-toggle" onclick={() => showPassword = !showPassword}>

      {#if showPassword}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
            <line x1="1" y1="1" x2="23" y2="23"></line>
          </svg>
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>
        {/if}
      </button>
    </div>
    <div class="form-group form-group-relative">
      <label class="form-label" for="reset-pass-confirm">Konfirmasi Password</label>
      <input class="input" id="reset-pass-confirm" type="{showPassword ? 'text' : 'password'}" bind:value={passwordConfirmation} />
      <button class="password-toggle" onclick={() => showPassword = !showPassword}>
        {#if showPassword}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
            <line x1="1" y1="1" x2="23" y2="23"></line>
          </svg>
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>
        {/if}
      </button>
    </div>

    <button class="btn btn-primary" style="width:100%;" onclick={handleSubmit} disabled={loading}>
      {loading ? 'Menyimpan...' : 'Simpan Password'}
    </button>

    <button class="btn btn-ghost" style="width:100%; margin-top:10px;" onclick={() => goto('/login')}>
      Kembali ke Login
    </button>
  </div>
</div>
