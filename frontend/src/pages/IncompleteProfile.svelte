<script>
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';
  import { replace } from '@mateothegreat/svelte5-router';

  let { intern = null } = $props();

  async function handleLogout() {
    await api.logout();
    auth.logout?.();
    replace('/login');
  }

  function goToRegistration() {
    const params = new URLSearchParams({
      email: auth.user?.email || '',
      name: auth.user?.name || '',
      oauth: 'google',
      status: 'unregistered'
    }).toString();
    replace(`/register-intern?${params}`);
  }
</script>

<div class="slide-up max-w-2xl mx-auto mt-20">
  <div class="card text-center p-8">
    <div class="w-20 h-20 rounded-full bg-amber-100 text-amber-600 flex items-center justify-center text-4xl mx-auto mb-6 shadow-sm">
      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
        <line x1="12" y1="9" x2="12" y2="13"/>
        <line x1="12" y1="17" x2="12.01" y2="17"/>
      </svg>
    </div>
    
    <h2 class="text-2xl font-bold text-slate-800 mb-3">Profil Belum Lengkap</h2>
    
    <p class="text-slate-600 mb-8 max-w-md mx-auto">
      Akun Anda belum terdaftar sebagai siswa magang. Mohon untuk daftar terlebih dahulu atau minta bantuan ke pembimbing magang.
    </p>
    
    <div class="flex flex-col sm:flex-row justify-center gap-4">
      <button onclick={goToRegistration} class="btn btn-primary px-6 py-3">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
          <path d="M12 5v14M5 12h14"/>
        </svg>
        Lengkapi Pendaftaran
      </button>
      
      <a href="/profile" class="btn btn-outline px-6 py-3">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
          <circle cx="12" cy="7" r="4"/>
        </svg>
        Lihat Profil
      </a>
      
      <button onclick={handleLogout} class="btn btn-outline px-6 py-3">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="inline mr-2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
          <polyline points="16 17 21 12 16 7"/>
          <line x1="21" y1="12" x2="9" y2="12"/>
        </svg>
        Logout
      </button>
    </div>
  </div>
</div>

<style>
  .slide-up {
    animation: slideUp 0.4s ease-out;
  }
  
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
