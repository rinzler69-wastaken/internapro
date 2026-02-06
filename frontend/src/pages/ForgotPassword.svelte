<script>
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';

  let email = $state('');
  let message = $state('');
  let resetUrl = $state('');
  let loading = $state(false);

  async function handleSubmit() {
    message = '';
    resetUrl = '';
    loading = true;
    try {
      const res = await api.requestPasswordReset(email);
      message = res.message || 'Jika email terdaftar, tautan reset akan dikirim.';
      if (res.data?.reset_url) resetUrl = res.data.reset_url;
    } catch (err) {
      message = err.message || 'Gagal mengirim permintaan reset.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="main-content" style="margin-left:0; display:flex; align-items:center; justify-content:center; min-height:100vh;">
  <div class="card" style="max-width:420px; width:100%;">
    <div style="text-align:center; margin-bottom:16px;">
      <h2 class="font-geist" style="margin:12px 0 4px;">Reset Password</h2>
      <p class="text-muted">Masukkan email Anda untuk menerima tautan reset.</p>
    </div>

    {#if message}
      <div class="badge" style="margin-bottom:12px;">{message}</div>
    {/if}

    <div class="form-group">
      <label class="form-label" for="reset-email">Email</label>
      <input class="input" id="reset-email" type="email" bind:value={email} placeholder="name@email.com" />
    </div>

    <button class="btn btn-primary" style="width:100%;" onclick={handleSubmit} disabled={loading}>
      {loading ? 'Mengirim...' : 'Kirim Link Reset'}
    </button>

    {#if resetUrl}
      <div class="text-muted" style="margin-top:12px; font-size:12px;">
        Dev link: <a href={resetUrl}>{resetUrl}</a>
      </div>
    {/if}

    <button class="btn btn-ghost" style="width:100%; margin-top:10px;" onclick={() => goto('/login')}>
      Kembali ke Login
    </button>
  </div>
</div>
