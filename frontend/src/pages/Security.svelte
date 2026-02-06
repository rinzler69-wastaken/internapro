<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let loading = $state(true);
  let user = $state({});
  
  // Password State
  let passwordForm = $state({ current_password: '', password: '', password_confirmation: '' });
  let savingPassword = $state(false);
  let showPassword = $state(false);

  // 2FA State
  let qrCode = $state('');
  let secret = $state('');
  let verifyCode = $state('');
  let recoveryCodes = $state([]);
  let showRecovery = $state(false);
  
  // Loading States
  let setupLoading = $state(false);
  let verifyLoading = $state(false);
  let disableLoading = $state(false);

  onMount(async () => {
    try {
      const res = await api.getProfile();
      user = res.data.user;
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  });

  async function updatePassword() {
    savingPassword = true;
    try {
      await api.updatePassword(passwordForm);
      passwordForm = { current_password: '', password: '', password_confirmation: '' };
      alert('Password berhasil diperbarui');
    } catch (err) {
      alert(err.message || 'Gagal memperbarui password');
    } finally {
      savingPassword = false;
    }
  }

  async function enable2FA() {
    setupLoading = true;
    try {
      // Expect backend to return { qr_code: "data:image...", secret: "XYZ..." }
      const res = await api.setup2FA(); 
      qrCode = res.data.qr_code;
      secret = res.data.secret;
    } catch (err) {
      alert(err.message || 'Gagal memulai setup 2FA');
    } finally {
      setupLoading = false;
    }
  }

  async function confirm2FA() {
    verifyLoading = true;
    try {
      await api.verify2FA({ code: verifyCode });
      alert('2FA Berhasil diaktifkan!');
      // Refresh user state
      const res = await api.getProfile();
      user = res.data.user;
      auth.hydrate(user);
      qrCode = ''; // Clear setup screen
    } catch (err) {
      alert(err.message || 'Kode verifikasi salah');
    } finally {
      verifyLoading = false;
    }
  }

  async function disable2FA() {
    if (!confirm('Apakah Anda yakin ingin menonaktifkan 2FA? Akun Anda akan kurang aman.')) return;
    disableLoading = true;
    try {
      await api.disable2FA();
      const res = await api.getProfile();
      user = res.data.user;
      auth.hydrate(user);
      alert('2FA telah dinonaktifkan.');
    } catch (err) {
      alert(err.message || 'Gagal menonaktifkan 2FA');
    } finally {
      disableLoading = false;
    }
  }

  async function fetchRecoveryCodes() {
    try {
      // Ensure backend route exists for this
      const res = await api.getRecoveryCodes(); 
      recoveryCodes = res.data || [];
      showRecovery = true;
    } catch (err) {
      alert('Gagal mengambil kode recovery. Pastikan password Anda benar.');
    }
  }
</script>

<div class="slide-up max-w-4xl mx-auto space-y-6">
  <div class="d-flex align-center gap-4 mb-6">
    <a href="/profile" class="btn btn-icon btn-secondary h-10 w-10 flex items-center justify-center rounded-full">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
    </a>
    <div>
      <h2 class="text-2xl font-bold">Keamanan Akun</h2>
      <p class="text-muted">Kelola password dan Two-Factor Authentication (2FA)</p>
    </div>
  </div>

  {#if loading}
    <div class="card p-6">Memuat...</div>
  {:else}
    <div class="card mb-6">
      <h3 class="font-bold text-lg text-slate-800 mb-4">Ubah Password</h3>
      <div class="form-group form-group-relative mb-3">
          <label class="form-label" for="profile-current-password">Password Saat Ini</label>
          <input class="input" id="profile-current-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.current_password} />
      </div>
      <div class="form-group form-group-relative mb-3">
          <label class="form-label" for="profile-new-password">Password Baru</label>
          <input class="input" id="profile-new-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.password} />
      </div>
      <div class="form-group form-group-relative mb-4">
          <label class="form-label" for="profile-confirm-password">Konfirmasi Password</label>
          <input class="input" id="profile-confirm-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.password_confirmation} />
      </div>
      <button class="btn btn-primary" onclick={updatePassword} disabled={savingPassword}>
        {savingPassword ? 'Menyimpan...' : 'Perbarui Password'}
      </button>
    </div>

    <div class="card p-6">
      <div class="flex items-center gap-4 mb-6">
        {#if user.is_2fa_enabled}
          <div class="w-14 h-14 rounded-2xl flex items-center justify-center text-2xl text-white shadow-lg bg-gradient-to-br from-green-500 to-emerald-500">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M9 12l2 2 4-4"/></svg>
          </div>
          <div>
            <h3 class="font-bold text-lg text-slate-800">Two-Factor Authentication</h3>
            <p class="text-sm text-green-600 font-medium">Aktif dan Terkonfirmasi</p>
          </div>
        {:else}
          <div class="w-14 h-14 rounded-2xl flex items-center justify-center text-2xl text-white shadow-lg bg-gradient-to-br from-red-500 to-rose-500">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          </div>
          <div>
            <h3 class="font-bold text-lg text-slate-800">Two-Factor Authentication</h3>
            <p class="text-sm text-red-600 font-medium">Tidak Aktif</p>
          </div>
        {/if}
      </div>

      {#if user.is_2fa_enabled}
        <div class="bg-green-50 border border-green-200 rounded-xl p-4 mb-6 text-green-800 text-sm">
          Akun Anda dilindungi dengan aman. Setiap login memerlukan kode dari Google Authenticator.
        </div>

        <div class="mb-6">
          <h4 class="font-semibold text-slate-700 mb-2">Kode Recovery</h4>
          <p class="text-sm text-muted mb-3">Gunakan kode ini jika Anda kehilangan akses ke HP Anda.</p>
          
          {#if showRecovery}
            <div class="bg-slate-50 rounded-lg p-4 font-mono text-sm grid grid-cols-2 gap-2 border">
              {#each recoveryCodes as code}
                <div class="text-center py-1 bg-white rounded border">{code}</div>
              {/each}
            </div>
          {:else}
            <button class="btn btn-secondary w-full" onclick={fetchRecoveryCodes}>
              Lihat Kode Recovery
            </button>
          {/if}
        </div>

        <div class="border-t pt-4">
          <button class="btn btn-outline text-red-600 border-red-200 hover:bg-red-50" onclick={disable2FA} disabled={disableLoading}>
            {disableLoading ? 'Memproses...' : 'Nonaktifkan 2FA'}
          </button>
        </div>

      {:else if qrCode}
        <div class="bg-indigo-50 border border-indigo-200 rounded-xl p-4 mb-6 text-indigo-800 text-sm">
          Scan QR Code di bawah ini dengan aplikasi Authenticator Anda, lalu masukkan kode 6 angka.
        </div>

        <div class="text-center mb-6">
          <div class="inline-block bg-white p-2 rounded-xl border shadow-sm mb-4">
            <img src={qrCode} alt="QR Code" width="200" height="200" />
          </div>
          {#if secret}
             <p class="text-xs text-muted font-mono">Secret: {secret}</p>
          {/if}
        </div>

        <div class="max-w-xs mx-auto">
          <label class="form-label text-center mb-2">Kode Verifikasi</label>
          <input 
            type="text" 
            class="input text-center text-2xl tracking-widest font-mono mb-4" 
            placeholder="000000" 
            maxlength="6" 
            bind:value={verifyCode}
          />
          <button class="btn btn-primary w-full" onclick={confirm2FA} disabled={verifyLoading || verifyCode.length !== 6}>
            {verifyLoading ? 'Memverifikasi...' : 'Konfirmasi & Aktifkan'}
          </button>
          
          <button class="btn btn-ghost w-full mt-2 text-sm text-muted" onclick={() => qrCode = ''}>
            Batalkan Setup
          </button>
        </div>

      {:else}
        <div class="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-6 text-amber-800 text-sm">
          Tingkatkan keamanan akun Anda dengan mengaktifkan autentikasi dua langkah.
        </div>
        <button class="btn btn-primary" onclick={enable2FA} disabled={setupLoading}>
          {setupLoading ? 'Memproses...' : 'Mulai Setup 2FA'}
        </button>
      {/if}
    </div>
  {/if}
</div>
