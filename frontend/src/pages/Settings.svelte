<script>
  import { onMount } from 'svelte';
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';

  // --- STATE ---
  let activeTab = $state('profile'); // 'profile', 'security', 'notifications', 'appearance'
  let loading = $state(true);
  let saving = $state(false);
  
  // Profile State
  let avatarFile = $state(null);
  let avatarPreview = $state(null);
  let fileInput;
  let profileForm = $state({ name: '', email: '', bio: '', phone: '' });

  // Security State
  let passwordForm = $state({ current_password: '', password: '', password_confirmation: '' });
  let showCurrentPass = $state(false);
  let showNewPass = $state(false);

  // Mock Settings State (Untuk UI)
  let notifSettings = $state({ email_task: true, email_login: true, push_task: false });
  let appSettings = $state({ theme: 'light', language: 'id' });

  onMount(async () => {
    await fetchUserData();
  });

  async function fetchUserData() {
    loading = true;
    try {
      const res = await api.getProfile();
      const userData = res.data?.user || res.data || auth.user || {};
      if (auth.hydrate) auth.hydrate(userData);
      
      profileForm = { 
        name: userData.name || '', 
        email: userData.email || '',
        phone: userData.phone || '', // Jika ada di backend
        bio: userData.bio || ''      // Jika ada di backend
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

  function buildUploadUrl(path) {
    if (!path) return '';
    const clean = path.startsWith('/uploads/') ? path : `/uploads/${path}`;
    const params = [];
    if (auth.token) params.push(`token=${auth.token}`);
    params.push(`t=${Date.now()}`);
    return `${clean}?${params.join('&')}`;
  }

  function handleFileSelect(e) {
    const file = e.target.files?.[0];
    if (file) {
        if (file.size > 2 * 1024 * 1024) {
            alert("Maksimal 2MB"); return;
        }
        avatarFile = file;
        const reader = new FileReader();
        reader.onload = (e) => avatarPreview = e.target.result;
        reader.readAsDataURL(file);
    }
  }

  async function saveProfile() {
    saving = true;
    try {
      const formData = new FormData();
      formData.append('name', profileForm.name);
      formData.append('email', profileForm.email);
      if (avatarFile) formData.append('avatar', avatarFile);

      await api.updateProfile(formData);
      await fetchUserData();

      alert('Profil berhasil diperbarui!');
      avatarFile = null;
    } catch (err) {
      alert(err.message || 'Gagal update profil');
    } finally {
      saving = false;
    }
  }

  async function updatePassword() {
    if (passwordForm.password !== passwordForm.password_confirmation) {
        alert("Konfirmasi password tidak cocok"); return;
    }
    saving = true;
    try {
      await api.updatePassword(passwordForm);
      passwordForm = { current_password: '', password: '', password_confirmation: '' };
      alert('Password berhasil diperbarui');
    } catch (err) {
      alert(err.message || 'Gagal ganti password');
    } finally {
      saving = false;
    }
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <div class="settings-layout">
        <!-- SIDEBAR NAVIGATION -->
        <aside class="settings-sidebar">
            <div class="sidebar-header">
                <h2>Pengaturan</h2>
                <p>Kelola akun dan preferensi</p>
            </div>
            
            <nav class="nav-menu">
                <button 
                    class="nav-item {activeTab === 'profile' ? 'active' : ''}" 
                    onclick={() => activeTab = 'profile'}
                >
                    <div class="icon-box"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg></div>
                    <span>Profil Saya</span>
                </button>

                <button 
                    class="nav-item {activeTab === 'security' ? 'active' : ''}" 
                    onclick={() => activeTab = 'security'}
                >
                    <div class="icon-box"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg></div>
                    <span>Login & Keamanan</span>
                </button>

                <button 
                    class="nav-item {activeTab === 'notifications' ? 'active' : ''}" 
                    onclick={() => activeTab = 'notifications'}
                >
                    <div class="icon-box"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg></div>
                    <span>Notifikasi</span>
                </button>

                <button 
                    class="nav-item {activeTab === 'appearance' ? 'active' : ''}" 
                    onclick={() => activeTab = 'appearance'}
                >
                    <div class="icon-box"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg></div>
                    <span>Tampilan</span>
                </button>
            </nav>
        </aside>

        <!-- CONTENT AREA -->
        <main class="settings-content">
            
            {#if activeTab === 'profile'}
                <div class="content-card animate-slide-up">
                    <div class="card-header">
                        <h3>Informasi Profil</h3>
                        <p>Perbarui foto dan detail pribadi Anda di sini.</p>
                    </div>

                    <div class="card-body">
                        <!-- Avatar -->
                        <div class="avatar-row">
                            <div class="avatar-preview">
                                {#if avatarPreview}
                                    <img src={avatarPreview} alt="Avatar" />
                                {:else}
                                    <div class="avatar-placeholder">{profileForm.name?.charAt(0) || 'U'}</div>
                                {/if}
                            </div>
                            <div class="avatar-actions">
                                <button class="btn-secondary" onclick={() => fileInput.click()}>Ganti Foto</button>
                                <button class="btn-text text-red" onclick={() => { avatarFile=null; avatarPreview=null; }}>Hapus</button>
                                <p class="help-text">JPG, GIF atau PNG. Maksimal 2MB.</p>
                                <input bind:this={fileInput} type="file" accept="image/*" hidden onchange={handleFileSelect} />
                            </div>
                        </div>

                        <div class="form-divider"></div>

                        <!-- Form -->
                        <div class="form-grid">
                            <div class="form-group">
                                <label class="label">Nama Lengkap</label>
                                <div class="input-wrapper">
                                    <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                                    <input class="input" bind:value={profileForm.name} placeholder="Nama Lengkap" />
                                </div>
                            </div>

                            <div class="form-group">
                                <label class="label">Email</label>
                                <div class="input-wrapper">
                                    <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
                                    <input class="input" bind:value={profileForm.email} placeholder="email@contoh.com" />
                                </div>
                            </div>
                        </div>

                        <div class="form-actions">
                            <button class="btn-primary" onclick={saveProfile} disabled={saving}>
                                {saving ? 'Menyimpan...' : 'Simpan Perubahan'}
                            </button>
                        </div>
                    </div>
                </div>
            {/if}

            {#if activeTab === 'security'}
                <div class="content-card animate-slide-up">
                    <div class="card-header">
                        <h3>Ganti Password</h3>
                        <p>Pastikan akun Anda aman dengan password yang kuat.</p>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="label">Password Saat Ini</label>
                            <input class="input" type="password" bind:value={passwordForm.current_password} placeholder="••••••" />
                        </div>
                        <div class="form-grid">
                            <div class="form-group">
                                <label class="label">Password Baru</label>
                                <input class="input" type="password" bind:value={passwordForm.password} placeholder="Min. 8 karakter" />
                            </div>
                            <div class="form-group">
                                <label class="label">Konfirmasi Password</label>
                                <input class="input" type="password" bind:value={passwordForm.password_confirmation} placeholder="Ulangi password" />
                            </div>
                        </div>
                        <div class="form-actions">
                            <button class="btn-primary" onclick={updatePassword} disabled={saving}>
                                {saving ? 'Memproses...' : 'Update Password'}
                            </button>
                        </div>
                    </div>
                </div>

                <div class="content-card mt-6 animate-slide-up">
                    <div class="card-header border-none pb-0">
                        <h3>Autentikasi Dua Faktor (2FA)</h3>
                        <p>Tambahkan lapisan keamanan ekstra.</p>
                    </div>
                    <div class="card-body flex justify-between items-center">
                        <div class="flex items-center gap-3">
                            <div class="icon-circle bg-emerald-light text-emerald">
                                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
                            </div>
                            <div>
                                <h4 class="font-bold text-sm">Status: Nonaktif</h4>
                                <p class="text-xs text-slate-500">Aktifkan untuk melindungi akun Anda.</p>
                            </div>
                        </div>
                        <button class="btn-outline">Setup 2FA</button>
                    </div>
                </div>
            {/if}

            {#if activeTab === 'notifications'}
                <div class="content-card animate-slide-up">
                    <div class="card-header">
                        <h3>Preferensi Notifikasi</h3>
                        <p>Pilih bagaimana Anda ingin dihubungi.</p>
                    </div>
                    <div class="card-body space-y-4">
                        <label class="toggle-row">
                            <div>
                                <span class="toggle-title">Notifikasi Tugas Baru</span>
                                <span class="toggle-desc">Terima email saat ada tugas baru diberikan.</span>
                            </div>
                            <input type="checkbox" class="toggle" bind:checked={notifSettings.email_task} />
                        </label>
                        <div class="form-divider"></div>
                        <label class="toggle-row">
                            <div>
                                <span class="toggle-title">Notifikasi Login</span>
                                <span class="toggle-desc">Beritahu saya jika ada login dari perangkat baru.</span>
                            </div>
                            <input type="checkbox" class="toggle" bind:checked={notifSettings.email_login} />
                        </label>
                    </div>
                    <div class="card-footer">
                        <button class="btn-primary">Simpan Preferensi</button>
                    </div>
                </div>
            {/if}

            {#if activeTab === 'appearance'}
                <div class="content-card animate-slide-up">
                    <div class="card-header">
                        <h3>Tampilan Aplikasi</h3>
                        <p>Sesuaikan pengalaman visual Anda.</p>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="label">Bahasa</label>
                            <select class="select" bind:value={appSettings.language}>
                                <option value="id">Bahasa Indonesia</option>
                                <option value="en">English</option>
                            </select>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label">Tema</label>
                            <div class="theme-grid">
                                <button class="theme-opt active">
                                    <div class="theme-preview light"></div>
                                    <span>Terang</span>
                                </button>
                                <button class="theme-opt">
                                    <div class="theme-preview dark"></div>
                                    <span>Gelap</span>
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            {/if}

        </main>
    </div>
  </div>
</div>

<style>
  :global(body) { font-family: 'Geist', 'Inter', sans-serif; color: #0f172a; background: #f8fafc; }
  
  .page-bg {
    min-height: 100vh;
    padding: 40px 24px;
    background: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%);
  }
  .container { max-width: 1100px; margin: 0 auto; }

  /* LAYOUT GRID */
  .settings-layout {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 32px;
    align-items: start;
  }
  @media (max-width: 768px) { .settings-layout { grid-template-columns: 1fr; } }

  /* SIDEBAR */
  .settings-sidebar {
    background: white; border-radius: 16px; padding: 24px;
    border: 1px solid #e2e8f0; position: sticky; top: 24px;
  }
  .sidebar-header h2 { font-size: 20px; font-weight: 600; margin: 0; color: #0f172a; }
  .sidebar-header p { font-size: 13px; color: #64748b; margin: 4px 0 24px; }
  
  .nav-menu { display: flex; flex-direction: column; gap: 4px; }
  .nav-item {
    display: flex; align-items: center; gap: 12px;
    padding: 10px 12px; border-radius: 10px;
    background: transparent; border: none; width: 100%; text-align: left;
    color: #64748b; font-weight: 500; font-size: 14px; cursor: pointer;
    transition: all 0.2s;
  }
  .nav-item:hover { background: #f1f5f9; color: #334155; }
  .nav-item.active { background: #ecfdf5; color: #059669; font-weight: 600; }
  .nav-item.active .icon-box { color: #059669; }

  .icon-box { width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; }

  /* CONTENT CARDS */
  .content-card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
  }
  .card-header { padding: 24px; border-bottom: 1px solid #f1f5f9; }
  .card-header h3 { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; }
  .card-header p { margin: 4px 0 0; font-size: 14px; color: #64748b; }
  
  .card-body { padding: 24px; }
  .card-footer { padding: 20px 24px; border-top: 1px solid #f1f5f9; text-align: right; background: #fcfcfc; }

  /* AVATAR */
  .avatar-row { display: flex; align-items: center; gap: 24px; }
  .avatar-preview { width: 80px; height: 80px; flex-shrink: 0; }
  .avatar-preview img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; border: 3px solid #fff; box-shadow: 0 0 0 2px #10b981; }
  .avatar-placeholder { 
    width: 100%; height: 100%; border-radius: 50%; background: #0f172a; color: white;
    display: flex; align-items: center; justify-content: center; font-size: 32px; font-weight: 600;
  }
  .avatar-actions { display: flex; flex-direction: column; align-items: flex-start; gap: 6px; }
  .help-text { font-size: 12px; color: #94a3b8; margin: 0; }
  .text-red { color: #ef4444; } .text-red:hover { text-decoration: underline; }

  /* FORMS */
  .form-divider { height: 1px; background: #f1f5f9; margin: 24px 0; }
  .form-grid { display: grid; grid-template-columns: 1fr; gap: 20px; }
  @media (min-width: 640px) { .form-grid { grid-template-columns: 1fr 1fr; } }
  
  .label { display: block; font-size: 13px; font-weight: 600; color: #334155; margin-bottom: 6px; }
  .input-wrapper { position: relative; }
  .input-icon { position: absolute; left: 12px; top: 12px; color: #94a3b8; pointer-events: none; }
  
  .input, .select {
    width: 100%; padding: 10px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; background: #fff; transition: all 0.2s; box-sizing: border-box;
  }
  .input-wrapper .input { padding-left: 40px; }
  .input:focus, .select:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }

  /* TOGGLES & SWITCHES */
  .toggle-row { display: flex; justify-content: space-between; align-items: center; cursor: pointer; }
  .toggle-title { display: block; font-weight: 600; font-size: 14px; color: #1e293b; }
  .toggle-desc { display: block; font-size: 13px; color: #64748b; }
  
  .toggle {
    appearance: none; width: 44px; height: 24px; background: #cbd5e1; border-radius: 99px; position: relative; cursor: pointer; transition: 0.3s;
  }
  .toggle::after {
    content: ''; position: absolute; top: 2px; left: 2px; width: 20px; height: 20px; background: white; border-radius: 50%; transition: 0.3s;
  }
  .toggle:checked { background: #10b981; }
  .toggle:checked::after { transform: translateX(20px); }

  /* THEME SELECTOR */
  .theme-grid { display: flex; gap: 16px; }
  .theme-opt {
    border: 2px solid #e2e8f0; background: white; border-radius: 12px; padding: 8px; cursor: pointer; flex: 1; text-align: center;
  }
  .theme-opt.active { border-color: #10b981; background: #ecfdf5; }
  .theme-preview { height: 60px; border-radius: 8px; margin-bottom: 8px; border: 1px solid #e2e8f0; }
  .theme-preview.light { background: #f8fafc; }
  .theme-preview.dark { background: #1e293b; }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white; border: none;
    padding: 10px 20px; border-radius: 10px; font-weight: 600; font-size: 14px; cursor: pointer;
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2); transition: 0.2s;
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 12px rgba(16, 185, 129, 0.3); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  .btn-secondary {
    background: white; border: 1px solid #cbd5e1; color: #334155; padding: 8px 16px;
    border-radius: 8px; font-weight: 600; font-size: 13px; cursor: pointer; transition: 0.2s;
  }
  .btn-secondary:hover { background: #f8fafc; border-color: #94a3b8; }
  
  .btn-outline {
    background: transparent; border: 1px solid #10b981; color: #10b981; padding: 8px 16px;
    border-radius: 8px; font-weight: 600; font-size: 13px; cursor: pointer;
  }
  .btn-text { background: none; border: none; cursor: pointer; font-size: 13px; font-weight: 500; }

  /* UTILS */
  .mt-4 { margin-top: 16px; } .mt-6 { margin-top: 24px; }
  .icon-circle { width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
  .bg-emerald-light { background: #ecfdf5; } .text-emerald { color: #059669; }

  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
</style>
