<script>
  import { onMount } from 'svelte';
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';

  // State
  let loading = $state(true);
  let saving = $state(false);
  let savingPassword = $state(false);
  
  // Profile State
  let avatarFile = $state(null);
  let avatarPreview = $state(null);
  let fileInput;
  let form = $state({ name: '', email: '' });

  // Password State
  let showCurrentPassword = $state(false);
  let showNewPassword = $state(false);
  let passwordForm = $state({ current_password: '', password: '', password_confirmation: '' });

  async function fetchProfile() {
    loading = true;
    try {
      const res = await api.getProfile();
      // Handle struktur response yang mungkin berbeda
      const userData = res.data?.user || res.data || auth.user || {};
      if (auth.hydrate) auth.hydrate(userData);
      
      form = { 
        name: userData.name || '', 
        email: userData.email || '' 
      };

      if (userData.avatar) {
        avatarPreview = buildUploadUrl(userData.avatar);
      }
    } catch (err) {
      console.error(err);
      // Fallback ke auth store jika API gagal
      const fallback = auth.user || {};
      form = { name: fallback.name || '', email: fallback.email || '' };
    } finally {
      loading = false;
    }
  }

  function handleFileSelect(e) {
    const file = e.target.files?.[0];
    if (file) {
        if (file.size > 2 * 1024 * 1024) {
            alert("Ukuran file maksimal 2MB");
            return;
        }
        avatarFile = file;
        // Preview lokal
        const reader = new FileReader();
        reader.onload = (e) => avatarPreview = e.target.result;
        reader.readAsDataURL(file);
    }
  }

  async function saveProfile() {
    saving = true;
    try {
      // Gunakan FormData untuk upload file
      const formData = new FormData();
      formData.append('name', form.name);
      formData.append('email', form.email);
      
      if (avatarFile) {
        formData.append('avatar', avatarFile);
      }

      await api.updateProfile(formData);
      
      // Refresh data
      await fetchProfile();
      
      alert('Profil berhasil diperbarui!');
      avatarFile = null; // Reset input file
    } catch (err) {
      alert(err.message || 'Gagal menyimpan profil');
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
      passwordForm = { current_password: '', password: '', password_confirmation: '' };
      alert('Password berhasil diperbarui');
    } catch (err) {
      alert(err.message || 'Gagal memperbarui password');
    } finally {
      savingPassword = false;
    }
  }

  onMount(fetchProfile);

  function buildUploadUrl(path) {
    if (!path) return '';
    const clean = path.startsWith('/uploads/') ? path : `/uploads/${path}`;
    const params = [];
    if (auth.token) params.push(`token=${auth.token}`);
    params.push(`t=${Date.now()}`);
    return `${clean}?${params.join('&')}`;
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <div>
            <h2 class="title">Pengaturan Akun</h2>
            <p class="subtitle">Kelola informasi pribadi dan keamanan akun Anda.</p>
        </div>
    </div>

    {#if loading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Memuat profil...</p>
        </div>
    {:else}
        <div class="grid-layout animate-slide-up">
            
            <!-- PROFILE CARD -->
            <div class="card">
                <div class="card-header">
                    <div class="header-icon bg-emerald-soft">
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
                                    {form.name?.charAt(0) || 'U'}
                                </div>
                            {/if}
                            
                            <!-- Edit Button Overlay -->
                            <button class="avatar-edit-btn" onclick={() => fileInput.click()} title="Ganti Foto">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                            </button>
                            <input 
                                bind:this={fileInput}
                                type="file" 
                                accept="image/png, image/jpeg, image/jpg" 
                                hidden 
                                onchange={handleFileSelect} 
                            />
                        </div>
                        <div class="avatar-info">
                            <h4 class="font-bold text-slate-800">Foto Profil</h4>
                            <p class="text-xs text-slate-500">Format: JPG, PNG. Maks 2MB.</p>
                        </div>
                    </div>

                    <!-- Form Inputs -->
                    <div class="space-y-4">
                        <div class="form-group">
                            <label class="label" for="profile-name">Nama Lengkap</label>
                            <div class="input-wrapper">
                                <div class="input-icon">
                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                                </div>
                                <input class="input-field" id="profile-name" bind:value={form.name} placeholder="Nama Anda" />
                            </div>
                        </div>
                        <div class="form-group">
                            <label class="label" for="profile-email">Alamat Email</label>
                            <div class="input-wrapper">
                                <div class="input-icon">
                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
                                </div>
                                <input class="input-field" id="profile-email" type="email" bind:value={form.email} placeholder="email@contoh.com" />
                            </div>
                        </div>
                    </div>

                    <div class="card-footer">
                        <button class="btn-primary w-full" onclick={saveProfile} disabled={saving}>
                            {#if saving}
                                <div class="spinner-small"></div> Menyimpan...
                            {:else}
                                Simpan Perubahan
                            {/if}
                        </button>
                    </div>
                </div>
            </div>

            <!-- SECURITY CARD -->
            <div class="card">
                <div class="card-header">
                    <div class="header-icon bg-blue-soft">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                    </div>
                    <h3>Keamanan</h3>
                </div>

                <div class="card-body">
                    <div class="space-y-4">
                        <div class="form-group">
                            <label class="label" for="current-pass">Password Saat Ini</label>
                            <div class="input-wrapper">
                                <input 
                                    class="input-field pr-10" 
                                    id="current-pass" 
                                    type={showCurrentPassword ? 'text' : 'password'} 
                                    bind:value={passwordForm.current_password} 
                                    placeholder="••••••"
                                />
                                <button class="toggle-btn" onclick={() => showCurrentPassword = !showCurrentPassword}>
                                    {#if showCurrentPassword}
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>
                                    {:else}
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
                                    {/if}
                                </button>
                            </div>
                        </div>

                        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div class="form-group">
                                <label class="label" for="new-pass">Password Baru</label>
                                <div class="input-wrapper">
                                    <input 
                                        class="input-field pr-10" 
                                        id="new-pass" 
                                        type={showNewPassword ? 'text' : 'password'} 
                                        bind:value={passwordForm.password} 
                                        placeholder="••••••"
                                    />
                                    <button class="toggle-btn" onclick={() => showNewPassword = !showNewPassword}>
                                        {#if showNewPassword}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>
                                        {:else}
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
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

                    <div class="card-footer">
                        <button class="btn-outline w-full" onclick={updatePassword} disabled={savingPassword}>
                            {savingPassword ? 'Memproses...' : 'Perbarui Password'}
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
    font-family: 'Geist', 'Inter', sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 40px 24px;
  }

  .container { max-width: 1000px; margin: 0 auto; }

  /* HEADER */
  .header { margin-bottom: 32px; }
  .title { font-size: 28px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; }
  .subtitle { color: #64748b; font-size: 15px; margin: 0; }

  /* LAYOUT */
  .grid-layout {
    display: grid; grid-template-columns: 1fr; gap: 24px;
  }
  @media (min-width: 768px) { .grid-layout { grid-template-columns: 1fr 1fr; } }

  /* CARDS */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
    display: flex; flex-direction: column; height: 100%;
  }
  .card-header {
    padding: 20px 24px; border-bottom: 1px solid #f1f5f9; display: flex; align-items: center; gap: 12px;
    background: #ffffff;
  }
  .header-icon {
    width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center;
  }
  .bg-emerald-soft { background: #ecfdf5; color: #059669; }
  .bg-blue-soft { background: #eff6ff; color: #2563eb; }

  .card-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
  .card-body { padding: 24px; flex: 1; }
  .card-footer { margin-top: 24px; padding-top: 20px; border-top: 1px solid #f8fafc; }

  /* AVATAR UPLOADER */
  .avatar-section {
    display: flex; align-items: center; gap: 20px; margin-bottom: 32px;
  }
  .avatar-wrapper { position: relative; width: 80px; height: 80px; flex-shrink: 0; }
  .avatar-img {
    width: 100%; height: 100%; border-radius: 50%; object-fit: cover;
    border: 4px solid #f1f5f9; box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  }
  .avatar-placeholder {
    width: 100%; height: 100%; border-radius: 50%; background: #0f172a; color: white;
    display: flex; align-items: center; justify-content: center; font-size: 32px; font-weight: 600;
    border: 4px solid #f1f5f9;
  }
  
  .avatar-edit-btn {
    position: absolute; bottom: 0; right: 0;
    width: 28px; height: 28px; border-radius: 50%;
    background: #10b981; color: white; border: 2px solid white;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: transform 0.2s;
  }
  .avatar-edit-btn:hover { transform: scale(1.1); background: #059669; }

  /* FORM ELEMENTS */
  .space-y-4 > * + * { margin-top: 16px; }
  .form-group { margin-bottom: 0; }
  .label { display: block; font-size: 12px; font-weight: 600; color: #475569; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.02em; }
  
  .input-wrapper { position: relative; display: flex; align-items: center; }
  .input-icon {
    position: absolute; left: 12px; color: #94a3b8; pointer-events: none;
    display: flex; align-items: center;
  }
  
  .input-field {
    width: 100%; padding: 10px 12px 10px 40px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; transition: all 0.2s; background: #fff;
  }
  .input-field.pr-10 { padding-right: 40px; } /* Space for toggle button */
  
  .input-field:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }

  .toggle-btn {
    position: absolute; right: 10px; background: none; border: none;
    color: #94a3b8; cursor: pointer; display: flex; align-items: center;
    padding: 4px; border-radius: 6px; transition: color 0.2s;
  }
  .toggle-btn:hover { color: #475569; background: #f1f5f9; }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white;
    padding: 12px; border-radius: 10px; font-weight: 600; font-size: 14px; border: none;
    cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); transition: all 0.2s;
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  .btn-outline {
    background: white; border: 1px solid #cbd5e1; color: #475569;
    padding: 12px; border-radius: 10px; font-weight: 600; font-size: 14px;
    cursor: pointer; transition: all 0.2s;
  }
  .btn-outline:hover:not(:disabled) { border-color: #10b981; color: #059669; background: #ecfdf5; }
  .btn-outline:disabled { opacity: 0.6; cursor: not-allowed; }

  /* UTILS */
  .w-full { width: 100%; }
  .grid { display: grid; }
  .grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)); }
  .sm\:grid-cols-2 { @media (min-width: 640px) { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  .gap-4 { gap: 16px; }

  .spinner { width: 40px; height: 40px; border: 3px solid #e2e8f0; border-top-color: #10b981; border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite; }
  .spinner-small { width: 16px; height: 16px; border: 2px solid white; border-top-color: transparent; border-radius: 50%; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .loading-state { text-align: center; padding: 60px; color: #94a3b8; font-style: italic; }

  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s ease-out forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
