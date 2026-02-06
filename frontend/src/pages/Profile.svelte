<script>
  import { onMount } from 'svelte';
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';

  let profile = $state(null);
  let loading = $state(false);
  let saving = $state(false);
  let savingPassword = $state(false);
  let avatarFile = $state(null);
  let showPassword = $state(false);

  let form = $state({ name: '', email: '' });
  let passwordForm = $state({ current_password: '', password: '', password_confirmation: '' });

  async function fetchProfile() {
    loading = true;
    try {
      const res = await api.getProfile();
      profile = res.data || {};
      const user = profile.user || auth.user || {};
      form = { name: user.name || '', email: user.email || '' };
    } catch (err) {
      console.error(err);
      const fallback = auth.user || {};
      form = { name: fallback.name || '', email: fallback.email || '' };
    } finally {
      loading = false;
    }
  }

  async function saveProfile() {
    saving = true;
    try {
await api.updateProfile({ form, avatarFile });      await fetchProfile();
    } catch (err) {
      alert(err.message || 'Gagal menyimpan profil');
    } finally {
      saving = false;
    }
  }

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

  onMount(fetchProfile);
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Profil</h3>
  <p class="text-muted">Perbarui informasi akun dan keamanan.</p>
</div>

{#if loading}
  <div class="card">Memuat...</div>
{:else}
  <div class="card" style="margin-bottom:16px;">
    <h4>Informasi Akun</h4>
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="profile-name">Nama</label>
        <input class="input" id="profile-name" bind:value={form.name} />
      </div>
      <div class="form-group">
        <label class="form-label" for="profile-email">Email</label>
        <input class="input" id="profile-email" type="email" bind:value={form.email} />
      </div>
      <div class="form-group">
        <label class="form-label" for="profile-avatar">Avatar</label>
        <input
          class="input"
          id="profile-avatar"
          type="file"
          accept=".png,.jpg,.jpeg"
          onchange={(e) => {
            const target = e.currentTarget;
            if (target instanceof HTMLInputElement) {
              avatarFile = target.files?.[0] || null;
            }
          }}
        />
      </div>
    </div>
    <button class="btn btn-primary" style="margin-top:12px;" onclick={saveProfile} disabled={saving}>
      {saving ? 'Menyimpan...' : 'Simpan Profil'}
    </button>
  </div>

  <div class="card">
    <h4>Keamanan</h4>
    <div class="form-group form-group-relative">
        <label class="form-label" for="profile-current-password">Password Saat Ini</label>
        <input class="input" id="profile-current-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.current_password} />
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
        <label class="form-label" for="profile-new-password">Password Baru</label>
        <input class="input" id="profile-new-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.password} />
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
        <label class="form-label" for="profile-confirm-password">Konfirmasi Password</label>
        <input class="input" id="profile-confirm-password" type="{showPassword ? 'text' : 'password'}" bind:value={passwordForm.password_confirmation} />
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
    <button class="btn btn-outline" style="margin-top:12px;" onclick={updatePassword} disabled={savingPassword}>
      {savingPassword ? 'Menyimpan...' : 'Perbarui Password'}
    </button>
  </div>
{/if}
