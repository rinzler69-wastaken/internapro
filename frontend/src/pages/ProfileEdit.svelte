<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import { goto } from '@mateothegreat/svelte5-router'; // Assuming you use this router

  let user = $state({ name: '', email: '' });
  let avatarPreview = $state(null);
  let avatarFile = $state(null);
  let loading = $state(false);

  // Password State
  let passForm = $state({ current_password: '', password: '', password_confirmation: '' });

  onMount(async () => {
    const res = await api.getProfile();
    user = res.data.user;
    if (user.avatar) avatarPreview = `/uploads/${user.avatar}`;
  });

  function handleFileSelect(e) {
    const file = e.target.files[0];
    if (file) {
      if (file.size > 2 * 1024 * 1024) {
        alert('File terlalu besar (Max 2MB)');
        return;
      }
      avatarFile = file;
      const reader = new FileReader();
      reader.onload = (e) => avatarPreview = e.target.result;
      reader.readAsDataURL(file);
    }
  }

  async function updateProfile(e) {
    e.preventDefault();
    loading = true;
    
    // CRITICAL: Use FormData for file uploads
    const formData = new FormData();
    formData.append('name', user.name);
    formData.append('email', user.email);
    if (avatarFile) {
        formData.append('avatar', avatarFile);
    }

    try {
        // You'll need to update api.js to support FormData or use fetch directly here
        // Assuming api.updateProfile handles FormData if passed
        await api.updateProfile(formData); 
        
        // Refresh Auth Context
        const res = await api.getProfile();
        auth.hydrate(res.data.user);
        
        alert('Profil berhasil diperbarui');
    } catch (err) {
        alert(err.message || 'Gagal update profil');
    } finally {
        loading = false;
    }
  }

  async function updatePassword(e) {
      e.preventDefault();
      // ... same logic as your previous Settings.svelte for password ...
  }
</script>

<div class="slide-up">
    <div class="d-flex align-center gap-4 mb-6">
        <a href="/profile" class="btn btn-secondary btn-icon" style="width:40px; height:40px; display:flex; align-items:center; justify-content:center;">
            ←
        </a>
        <div>
            <h2 style="margin:0;">Edit Profil</h2>
            <p class="text-muted" style="margin:0;">Perbarui informasi akun Anda</p>
        </div>
    </div>

    <div class="grid-2" style="display:grid; grid-template-columns: 1fr 1fr; gap:20px;">
        <div class="card">
            <div class="card-header mb-4"><h3 class="card-title">Informasi Profil</h3></div>
            
            <form onsubmit={updateProfile}>
                <div class="form-group mb-4">
                    <label class="form-label">Foto Profil</label>
                    <div class="d-flex align-center gap-4">
                        <div style="position:relative; width:100px; height:100px;">
                            {#if avatarPreview}
                                <img src={avatarPreview} alt="Preview" style="width:100%; height:100%; border-radius:50%; object-fit:cover; border:3px solid var(--border);">
                            {:else}
                                <div style="width:100%; height:100%; border-radius:50%; background:#e0e7ff; display:flex; align-items:center; justify-content:center; font-weight:bold; font-size:24px;">
                                    {user.name?.charAt(0) || 'U'}
                                </div>
                            {/if}
                        </div>
                        <div>
                            <label class="btn btn-secondary cursor-pointer">
                                Pilih Foto
                                <input type="file" accept="image/*" onchange={handleFileSelect} hidden>
                            </label>
                            <p class="text-muted mt-2" style="font-size:12px;">Max 2MB. JPG, PNG.</p>
                        </div>
                    </div>
                </div>

                <div class="form-group mb-3">
                    <label class="form-label">Nama Lengkap</label>
                    <input class="input" bind:value={user.name} required>
                </div>
                <div class="form-group mb-4">
                    <label class="form-label">Email</label>
                    <input class="input" type="email" bind:value={user.email} required>
                </div>
                <button type="submit" class="btn btn-primary" disabled={loading}>Simpan Perubahan</button>
            </form>
        </div>
    </div>
</div>
