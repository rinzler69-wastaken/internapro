<script>
  import { onMount } from 'svelte';
  import { goto, replace, route } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let email = $state('');
  let password = $state('');
  let totp = $state('');
  let error = $state('');
  let loading = $state(false);
  let needs2FA = $state(false);
  let setupRequired = $state(false);
  let showPassword = $state(false);

  onMount(async () => {
    if (!auth.token) return;
    try {
      if (!auth.user) {
        const res = await api.getCurrentUser();
        auth.hydrate(res.data);
      }
      replace('/');
    } catch (err) {
      auth.logout();
    }
  });

  function handleGoogle() {
    const redirectPath = '/dashboard';
    window.location.href = `/api/auth/google?redirect=1&redirect_path=${encodeURIComponent(redirectPath)}`;
  }

  async function handleSubmit() {
    error = '';
    loading = true;
    try {
      const res = await api.login(email, password, needs2FA ? totp : null);
      if (res?.data?.require_2fa) {
        needs2FA = true;
        loading = false;
        return;
      }
      setupRequired = !!res?.data?.setup_required;
      loading = false;
      if (setupRequired) {
        replace('/settings?setup=2fa');
      } else {
        replace('/dashboard');
      }
    } catch (err) {
      error = err.message || 'Login gagal';
      loading = false;
    }
  }
</script>

<div class="main-content" style="margin-left:0; display:flex; align-items:center; justify-content:center; min-height:100vh;">
  <div class="card" style="max-width:420px; width:100%;">
    <div style="text-align:center; margin-bottom:16px;">
      <div class="sidebar-brand-icon login-logo" style="margin:0 auto;">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
          <path d="M3 7.5l9-4 9 4-9 4-9-4z" />
          <path d="M6 10.5v4.5c0 2 3 3.5 6 3.5s6-1.5 6-3.5v-4.5" />
        </svg>
      </div>
      <h2 class="font-geist font-bold" style="margin:12px 0 4px;">Masuk InternaPro</h2>
      <p class="text-muted">Kelola magang, tugas, dan presensi secara praktis.</p>
    </div>

    {#if error}
      <div class="badge danger" style="margin-bottom:12px;">{error}</div>
    {/if}

    <div class="form-group">
      <label class="form-label" for="email">Email</label>
      <input class="input" type="email" id="email" bind:value={email} placeholder="name@email.com" />
    </div>

    <div class="form-group form-group-relative">
      <label class="form-label" for="password">Password</label>
      <input class="input" type={showPassword ? 'text' : 'password'} id="password" bind:value={password} placeholder="********" />
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

    {#if needs2FA}
      <div class="form-group">
        <label class="form-label" for="totp">Kode 2FA</label>
        <input class="input" type="text" id="totp" bind:value={totp} placeholder="123456" />
      </div>
    {/if}

    {#if setupRequired}
      <div class="badge warning" style="margin-bottom:12px;">Akun Anda membutuhkan setup 2FA. Silakan aktifkan setelah login.</div>
    {/if}

    <button class="btn btn-primary" style="width:100%;" onclick={handleSubmit} disabled={loading}>
      {loading ? 'Memproses...' : 'Masuk'}
    </button>

    <div style="display:flex; justify-content:space-between; align-items:center; margin-top:12px; font-size:12px;">
      <a href="/forgot-password" use:route class="text-muted">Lupa password?</a>
      <button class="btn btn-outline" type="button" onclick={handleGoogle}>Masuk dengan Google</button>
    </div>
  </div>
</div>
