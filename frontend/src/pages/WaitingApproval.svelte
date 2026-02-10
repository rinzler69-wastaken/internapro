<script>
  import { replace } from '@mateothegreat/svelte5-router';
  import { auth } from '../lib/auth.svelte.js';

  function handleBackToLogin() {
    auth.logout(); 
    replace('/login');
  }
</script>

<div class="login-container">
  <div class="card" style="text-align: center;">
    
    <div class="card-header">
       <!-- Icon Kuning -->
       <div class="brand-logo warning-tint">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="24" height="24">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
      </div>
      <h2 class="title">Status Pendaftaran</h2>
    </div>

    <!-- KOTAK TINTED KUNING -->
    <div class="status-badge-container">
      <div class="status-badge-yellow">
        Menunggu Verifikasi
      </div>
    </div>

    <div class="message-content">
      <p class="subtitle" style="margin-bottom: 2rem;">
        Anda telah melakukan pendaftaran magang. <br>
        <strong>Mohon tunggu approval dari admin.</strong>
      </p>
    </div>

    <button class="btn btn-outline" onclick={handleBackToLogin}>
        Kembali ke Login
    </button>
  </div>
</div>

<style>
  /* CSS Dasar (Sama dengan Login) */
  .login-container { display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 1rem; background-color: #f3f4f6; }
  .card { background: white; width: 100%; max-width: 400px; border-radius: 12px; padding: 2.5rem 2rem; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
  .card-header { margin-bottom: 1rem; }
  .title { font-size: 1.5rem; font-weight: 600; color: #111827; }
  .subtitle { color: #4b5563; font-size: 0.95rem; line-height: 1.5; }

  /* CSS Khusus Kuning */
  .brand-logo.warning-tint {
      display: inline-flex; align-items: center; justify-content: center; width: 48px; height: 48px; border-radius: 10px; margin-bottom: 1rem;
      background-color: #fffbeb;
      color: #d97706;
  }
  .status-badge-container { display: flex; justify-content: center; margin-bottom: 1.5rem; }
  .status-badge-yellow {
      background-color: #FEF3C7; color: #92400E;
      font-weight: 600; font-size: 0.8rem;
      padding: 6px 16px; border-radius: 99px; border: 1px solid #FDE68A;
  }
  .btn { width: 100%; padding: 0.75rem; border-radius: 8px; font-weight: 600; cursor: pointer; }
  .btn-outline { background-color: white; border: 1px solid #d1d5db; color: #374151; }
  .btn-outline:hover { background-color: #f9fafb; }
</style>