<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // --- STATE UMUM ---
  let loading = $state(true);
  let user = $state({
    name: '',
    email: '',
    avatar: '',
    is_2fa_enabled: false,
  });

  // --- STATE PROFIL ---
  let avatarFile = $state(null);
  let avatarPreview = $state(null);
  let fileInput = $state(null);
  let profileForm = $state({ name: '', email: '' });
  let savingProfile = $state(false);

  // --- STATE PASSWORD ---
  let passwordForm = $state({ current_password: '', password: '', password_confirmation: '' });
  let savingPassword = $state(false);
  let showCurrentPass = $state(false);
  let showNewPass = $state(false);

  // --- STATE 2FA ---
  let qrCode = $state('');
  let secret = $state('');
  let verifyCode = $state('');
  let recoveryCodes = $state([]);
  let showRecovery = $state(false);
  
  let setupLoading = $state(false);
  let verifyLoading = $state(false);
  let disableLoading = $state(false);

  // --- INITIAL LOAD ---
  onMount(async () => {
    await fetchUserData();
  });

  async function fetchUserData() {
    loading = true;
    try {
      const res = await api.getProfile();
      // Fail-safe: Pastikan user selalu objek dengan properti dasar
      const userData = res.data?.user || res.data || auth.user || {};
      user = {
        name: '',
        email: '',
        avatar: '',
        is_2fa_enabled: false,
        ...userData,
      };
      if (auth.hydrate) auth.hydrate(userData);

      // Isi form profil
      profileForm = { 
        name: user.name || '', 
        email: user.email || '' 
      };

      // Preview Avatar (Cache busting)
      if (user.avatar) {
        avatarPreview = buildUploadUrl(user.avatar);
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function buildUploadUrl(path) {
    if (!path) return '';
    const clean = path.startsWith('/uploads/') ? path : `/uploads/${path}`;
    const params = [];
    if (auth.token) params.push(`token=${auth.token}`);
    params.push(`t=${Date.now()}`);
    return `${clean}?${params.join('&')}`;
  }

  // ==========================================
  // LOGIC PROFIL
  // ==========================================
  function handleFileSelect(e) {
    const file = e.target.files?.[0];
    if (file) {
        if (file.size > 2 * 1024 * 1024) {
            alert("File terlalu besar (Max 2MB)");
            return;
        }
        avatarFile = file;
        const reader = new FileReader();
        reader.onload = (e) => avatarPreview = e.target.result;
        reader.readAsDataURL(file);
    }
  }

  async function saveProfile() {
    savingProfile = true;
    try {
      const formData = new FormData();
      formData.append('name', profileForm.name);
      formData.append('email', profileForm.email);
      if (avatarFile) formData.append('avatar', avatarFile);

      await api.updateProfile(formData);
      await fetchUserData(); // Refresh data

      alert('Profil berhasil diperbarui!');
      avatarFile = null;
      if (fileInput) fileInput.value = '';
    } catch (err) {
      alert(err.message || 'Gagal update profil');
    } finally {
      savingProfile = false;
    }
  }

  // ==========================================
  // LOGIC PASSWORD
  // ==========================================
  async function updatePassword() {
    if (passwordForm.password !== passwordForm.password_confirmation) {
        alert("Konfirmasi password tidak cocok");
        return;
    }
    savingPassword = true;
    try {
      await api.updatePassword(passwordForm);
      passwordForm = { current_password: '', password: '', password_confirmation: '' };
      alert('Password berhasil diperbarui');
    } catch (err) {
      alert(err.message || 'Gagal ganti password');
    } finally {
      savingPassword = false;
    }
  }

  // ==========================================
  // LOGIC 2FA
  // ==========================================
  async function enable2FA() {
    setupLoading = true;
    try {
      const res = await api.setup2FA(); 
      // Pastikan backend mengembalikan struktur ini
      qrCode = res.data.qr_code || res.data.qrCodeUrl; 
      secret = res.data.secret;
    } catch (err) {
      alert(err.message || 'Gagal setup 2FA');
    } finally {
      setupLoading = false;
    }
  }

  async function confirm2FA() {
    verifyLoading = true;
    try {
      await api.verify2FA({ code: verifyCode });
      alert('2FA Berhasil diaktifkan!');
      
      await fetchUserData(); // Refresh user state (is_2fa_enabled: true)
      
      qrCode = ''; 
      verifyCode = '';
    } catch (err) {
      alert(err.message || 'Kode verifikasi salah');
    } finally {
      verifyLoading = false;
    }
  }

  async function disable2FA() {
    if (!confirm('Nonaktifkan 2FA? Akun Anda akan menjadi kurang aman.')) return;
    disableLoading = true;
    try {
      await api.disable2FA();
      await fetchUserData(); // Refresh user state (is_2fa_enabled: false)
      alert('2FA telah dinonaktifkan.');
      showRecovery = false;
    } catch (err) {
      alert(err.message || 'Gagal menonaktifkan 2FA');
    } finally {
      disableLoading = false;
    }
  }

  async function fetchRecoveryCodes() {
    try {
      const res = await api.getRecoveryCodes(); // Pastikan endpoint ini ada di api.js
      recoveryCodes = res.data || [];
      showRecovery = true;
    } catch (err) {
      alert('Gagal mengambil kode recovery. Coba refresh halaman.');
    }
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <a href="/dashboard" class="btn-back">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5"/><path d="M12 19l-7-7 7-7"/></svg>
            Kembali ke Dashboard
        </a>
        <div class="mt-4">
            <h2 class="title">Pengaturan Akun</h2>
            <p class="subtitle">Kelola profil pribadi dan keamanan akun Anda.</p>
        </div>
    </div>

    {#if loading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Memuat data...</p>
        </div>
    {:else}
        <div class="grid-layout animate-slide-up">
            
            <!-- KOLOM KIRI: EDIT PROFIL -->
            <div class="card">
                <div class="card-header">
                    <div class="icon-circle bg-emerald">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                    </div>
                    <h3>Profil Saya</h3>
                </div>
                
                <div class="card-body">
                    <!-- Avatar Upload -->
                    <div class="avatar-section">
                        <div class="avatar-wrapper">
                            {#if avatarPreview}
                                <img src={avatarPreview} alt="Avatar" class="avatar-img" />
                            {:else}
                                <div class="avatar-placeholder">
                                    {profileForm.name?.charAt(0) || 'U'}
                                </div>
                            {/if}
                            <button class="avatar-btn" onclick={() => fileInput.click()} title="Ganti Foto">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>
                            </button>
                            <input bind:this={fileInput} type="file" accept="image/*" hidden onchange={handleFileSelect} />
                        </div>
                        <div class="avatar-text">
                            <h4 class="font-bold text-slate-800">Foto Profil</h4>
                            <p class="text-xs text-slate-500">JPG/PNG, Maks 2MB.</p>
                        </div>
                    </div>

                    <div class="divider"></div>

                    <!-- Form Inputs -->
                    <div class="space-y-4">
                        <div class="form-group">
                            <label class="label" for="profile-name">Nama Lengkap</label>
                            <div class="input-wrapper">
                                <input class="input-field" id="profile-name" bind:value={profileForm.name} placeholder="Nama Anda" />
                            </div>
                        </div>
                        <div class="form-group">
                            <label class="label" for="profile-email">Alamat Email</label>
                            <div class="input-wrapper">
                                <input class="input-field" id="profile-email" type="email" bind:value={profileForm.email} placeholder="email@contoh.com" />
                            </div>
                        </div>
                    </div>

                    <div class="action-area">
                        <button class="btn-primary w-full" onclick={saveProfile} disabled={savingProfile}>
                            {#if savingProfile}
                                <div class="spinner-small"></div> Menyimpan...
                            {:else}
                                Simpan Perubahan
                            {/if}
                        </button>
                    </div>
                </div>
            </div>

            <!-- KOLOM KANAN: KEAMANAN (PASSWORD & 2FA) -->
            <div class="security-column space-y-6">
                
                <!-- CARD 2FA -->
                <div class="card">
                    <div class="card-header">
                        <div class="icon-circle bg-purple">
                           <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><path d="M9 12l2 2 4-4"></path></svg>
                        </div>
                        <h3>Autentikasi 2 Langkah</h3>
                    </div>

                    <div class="card-body">
                        {#if user?.is_2fa_enabled}
                             <!-- STATE: 2FA AKTIF -->
                             <div class="status-active">
                                <div class="badge-success">
                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
                                    Aktif & Terlindungi
                                </div>
                                <p class="text-sm text-slate-500 mt-2">
                                    Akun Anda dilindungi dengan kode autentikator setiap kali login.
                                </p>
                             </div>

                             <div class="recovery-section">
                                <h4 class="text-sm font-semibold text-slate-700 mb-2">Kode Pemulihan</h4>
                                {#if showRecovery}
                                    <div class="recovery-grid">
                                        {#each recoveryCodes as code}
                                            <div class="code-box">{code}</div>
                                        {/each}
                                    </div>
                                    <button class="btn-text-sm" onclick={() => showRecovery = false}>Sembunyikan</button>
                                {:else}
                                    <button class="btn-outline-sm w-full" onclick={fetchRecoveryCodes}>Lihat Kode Pemulihan</button>
                                {/if}
                             </div>

                             <div class="mt-6 pt-4 border-t border-slate-100">
                                <button class="btn-danger-text w-full" onclick={disable2FA} disabled={disableLoading}>
                                    {disableLoading ? 'Memproses...' : 'Nonaktifkan 2FA'}
                                </button>
                             </div>

                        {:else if qrCode}
                             <!-- STATE: PROSES SETUP (SCAN QR) -->
                             <div class="setup-2fa">
                                <p class="text-sm text-slate-600 mb-4 text-center">
                                    Scan QR code ini dengan aplikasi <strong>Google Authenticator</strong> atau Authy.
                                </p>
                                <div class="qr-box">
                                    <img src={qrCode} alt="QR Code" />
                                </div>
                                {#if secret}
                                    <p class="secret-text">Secret: {secret}</p>
                                {/if}

                                <div class="verify-form">
                                    <label class="label text-center" for="verify-code">Masukkan Kode 6 Digit</label>
                                    <input 
                                        id="verify-code"
                                        type="text" 
                                        class="input-field text-center tracking-widest font-mono text-lg" 
                                        placeholder="000000" 
                                        maxlength="6"
                                        bind:value={verifyCode} 
                                    />
                                    <button class="btn-primary w-full mt-3" onclick={confirm2FA} disabled={verifyLoading || verifyCode.length < 6}>
                                        {verifyLoading ? 'Memverifikasi...' : 'Aktifkan 2FA'}
                                    </button>
                                </div>
                                <button class="btn-text w-full mt-2" onclick={() => qrCode = ''}>Batalkan</button>
                             </div>

                        {:else}
                             <!-- STATE: 2FA BELUM AKTIF -->
                             <div class="status-inactive">
                                <p class="text-sm text-slate-500 mb-4">
                                    Tambahkan lapisan keamanan ekstra. Anda akan diminta memasukkan kode dari HP saat login.
                                </p>
                                <button class="btn-outline w-full" onclick={enable2FA} disabled={setupLoading}>
                                    {setupLoading ? 'Memproses...' : 'Aktifkan Sekarang'}
                                </button>
                             </div>
                        {/if}
                    </div>
                </div>

                <!-- CARD PASSWORD -->
                <div class="card">
                    <div class="card-header">
                        <div class="icon-circle bg-blue">
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                        </div>
                        <h3>Ganti Password</h3>
                    </div>

                    <div class="card-body">
                        <div class="space-y-4">
                            <div class="form-group">
                                <label class="label" for="curr-pass">Password Saat Ini</label>
                                <div class="input-wrapper">
                                    <input class="input-field pr-10" id="curr-pass" type={showCurrentPass ? 'text' : 'password'} bind:value={passwordForm.current_password} />
                                    <button class="toggle-btn" onclick={() => showCurrentPass = !showCurrentPass}>
                                        {#if showCurrentPass}👁️{:else}👁️‍🗨️{/if}
                                    </button>
                                </div>
                            </div>
                            <div class="grid grid-cols-2 gap-3">
                                <div class="form-group">
                                    <label class="label" for="new-pass">Password Baru</label>
                                    <input class="input-field" id="new-pass" type="password" bind:value={passwordForm.password} />
                                </div>
                                <div class="form-group">
                                    <label class="label" for="confirm-pass">Konfirmasi</label>
                                    <input class="input-field" id="confirm-pass" type="password" bind:value={passwordForm.password_confirmation} />
                                </div>
                            </div>
                        </div>

                        <div class="action-area mt-6">
                            <button class="btn-secondary w-full" onclick={updatePassword} disabled={savingPassword}>
                                {savingPassword ? 'Memproses...' : 'Perbarui Password'}
                            </button>
                        </div>
                    </div>
                </div>

            </div>
        </div>
    {/if}
  </div>
</div>

<style>
  :global(body) { font-family: 'Geist', 'Inter', sans-serif; color: #0f172a; }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 40px 24px;
  }
  .container { max-width: 1100px; margin: 0 auto; }

  /* HEADER */
  .header { margin-bottom: 32px; }
  .title { font-size: 28px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; }
  .subtitle { color: #64748b; font-size: 15px; margin: 0; }
  .btn-back {
      display: inline-flex; align-items: center; gap: 8px; color: #64748b; 
      font-weight: 600; font-size: 13px; text-decoration: none; transition: all 0.2s;
      background: white; padding: 8px 14px; border-radius: 99px; border: 1px solid #e2e8f0;
  }
  .btn-back:hover { color: #0f172a; border-color: #cbd5e1; transform: translateX(-2px); }

  /* LAYOUT */
  .grid-layout { display: grid; grid-template-columns: 1fr; gap: 24px; }
  @media (min-width: 900px) { .grid-layout { grid-template-columns: 1fr 1fr; } }

  /* CARDS */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
  }
  .card-header {
    padding: 20px 24px; border-bottom: 1px solid #f1f5f9; display: flex; align-items: center; gap: 12px;
  }
  .card-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
  .icon-circle { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
  .bg-emerald { background: #ecfdf5; color: #059669; }
  .bg-blue { background: #eff6ff; color: #2563eb; }
  .bg-purple { background: #faf5ff; color: #7e22ce; }

  .card-body { padding: 24px; }
  .divider { height: 1px; background: #f1f5f9; margin: 24px 0; }

  /* AVATAR */
  .avatar-section { display: flex; align-items: center; gap: 20px; }
  .avatar-wrapper { position: relative; width: 80px; height: 80px; flex-shrink: 0; }
  .avatar-img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; border: 3px solid #fff; box-shadow: 0 0 0 2px #10b981; }
  .avatar-placeholder { 
    width: 100%; height: 100%; border-radius: 50%; background: #0f172a; color: white;
    display: flex; align-items: center; justify-content: center; font-size: 32px; font-weight: 600;
  }
  .avatar-btn {
    position: absolute; bottom: 0; right: 0; width: 28px; height: 28px; border-radius: 50%;
    background: #10b981; color: white; border: 2px solid white; cursor: pointer;
    display: flex; align-items: center; justify-content: center; transition: transform 0.2s;
  }
  .avatar-btn:hover { transform: scale(1.1); background: #059669; }

  /* FORM */
  .space-y-4 > * + * { margin-top: 16px; }
  .label { display: block; font-size: 12px; font-weight: 600; color: #475569; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.02em; }
  .input-wrapper { position: relative; }
  .input-field {
    width: 100%; padding: 10px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; transition: all 0.2s; background: #fff; box-sizing: border-box;
  }
  .input-field.pr-10 { padding-right: 40px; }
  .input-field:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }
  .toggle-btn { position: absolute; right: 10px; top: 10px; background: none; border: none; cursor: pointer; font-size: 14px; }

  /* 2FA STYLES */
  .badge-success {
    display: inline-flex; align-items: center; gap: 6px; padding: 6px 12px;
    background: #d1fae5; color: #065f46; border-radius: 99px; font-size: 13px; font-weight: 600;
  }
  .qr-box { background: white; padding: 10px; border: 1px solid #e2e8f0; border-radius: 12px; display: inline-block; margin: 0 auto 10px; }
  .qr-box img { width: 160px; height: 160px; }
  .secret-text { font-family: monospace; font-size: 12px; color: #64748b; background: #f1f5f9; padding: 4px 8px; border-radius: 4px; display: inline-block; margin-bottom: 16px; }
  .setup-2fa { text-align: center; }
  
  .recovery-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 8px; }
  .code-box { background: #f8fafc; border: 1px solid #e2e8f0; padding: 6px; text-align: center; font-family: monospace; font-size: 13px; border-radius: 6px; color: #334155; }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white;
    padding: 12px; border-radius: 10px; font-weight: 600; font-size: 14px; border: none;
    cursor: pointer; width: 100%; transition: all 0.2s;
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  .btn-secondary {
    background: #0f172a; color: white; border: none; padding: 12px; border-radius: 10px; font-weight: 600; font-size: 14px; cursor: pointer; width: 100%; transition: all 0.2s;
  }
  .btn-secondary:hover:not(:disabled) { background: #1e293b; }

  .btn-outline, .btn-outline-sm {
    background: white; border: 1px solid #cbd5e1; color: #475569;
    padding: 10px; border-radius: 10px; font-weight: 600; font-size: 14px;
    cursor: pointer; transition: all 0.2s;
  }
  .btn-outline:hover:not(:disabled) { border-color: #10b981; color: #059669; background: #ecfdf5; }
  .btn-outline-sm { padding: 8px; font-size: 13px; }

  .btn-text, .btn-text-sm, .btn-danger-text { background: none; border: none; cursor: pointer; font-weight: 500; }
  .btn-text { color: #64748b; font-size: 14px; } .btn-text:hover { color: #334155; }
  .btn-text-sm { color: #3b82f6; font-size: 13px; } .btn-text-sm:hover { text-decoration: underline; }
  .btn-danger-text { color: #ef4444; font-size: 13px; font-weight: 600; } .btn-danger-text:hover { text-decoration: underline; }

  .spinner, .spinner-small { border: 2px solid white; border-top-color: transparent; border-radius: 50%; animation: spin 1s linear infinite; display: inline-block; }
  .spinner { width: 40px; height: 40px; border: 3px solid #e2e8f0; border-top-color: #10b981; }
  .spinner-small { width: 16px; height: 16px; vertical-align: middle; margin-right: 6px; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .loading-state { text-align: center; padding: 60px; color: #94a3b8; }
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s ease-out forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
