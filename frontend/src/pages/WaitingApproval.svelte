<script>
  import { replace } from '@mateothegreat/svelte5-router';
  import { auth } from '../lib/auth.svelte.js';

  function handleBackToLogin() {
    auth.logout?.();
    try {
      localStorage.removeItem('token');
    } catch (_) {}
    replace('/login');
    // Fallback if router state is stale in shell-level overrides
    setTimeout(() => {
      if (window.location.pathname !== '/login') {
        window.location.assign('/login');
      }
    }, 30);
  }
</script>

<div class="login-container">
  <div class="card waiting-card fade-in">
    <div class="card-header fade-in">
      <div class="brand-logo warning-tint">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          width="24"
          height="24"
        >
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
      </div>
      <h2 class="title">Status Pendaftaran</h2>
      <p class="subtitle">Akun Anda akan segera kami proses.</p>
    </div>

    <div class="status-badge-container fade-in">
      <div class="status-badge-yellow">
        Menunggu Verifikasi
      </div>
    </div>

    <div class="approval-box-wrapper fade-in">
      <div class="approval-box">
        <p class="approval-text">
          Anda telah melakukan pendaftaran magang.
          <strong>Mohon tunggu approval dari admin.</strong>
        </p>
      </div>
    </div>

    <button class="btn btn-outline" onclick={handleBackToLogin}>
      Kembali ke Login
    </button>
  </div>
</div>

<style>
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 1rem;
    background-color: #f3f4f6;
  }

  .card {
    background: white;
    width: 100%;
    max-width: 400px;
    border-radius: 12px;
    padding: 2.5rem 2rem;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    text-align: center;
  }

  .card-header {
    margin-bottom: 1rem;
  }

  .title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #6b7280;
    font-size: 0.9rem;
    margin: 0;
  }

  .brand-logo.warning-tint {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    border-radius: 10px;
    margin-bottom: 1rem;
    background-color: #fffbeb;
    color: #d97706;
  }

  .status-badge-container {
    display: flex;
    justify-content: center;
    margin-bottom: 1.5rem;
  }

  .status-badge-yellow {
    background-color: #fef3c7;
    color: #92400e;
    font-weight: 600;
    font-size: 0.8rem;
    padding: 6px 16px;
    border-radius: 99px;
    border: 1px solid #fde68a;
  }

  .approval-box-wrapper {
    margin-bottom: 2rem;
  }

  .approval-box {
    background-color: #fefce8;
    border: 1px solid #fde047;
    color: #854d0e;
    padding: 1.25rem;
    border-radius: 8px;
    text-align: center;
  }

  .approval-text {
    margin: 0;
    font-size: 0.95rem;
    line-height: 1.5;
    font-weight: 500;
  }

  .btn {
    width: 100%;
    padding: 0.75rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-outline {
    background-color: white;
    border: 1px solid #d1d5db;
    color: #374151;
    transition: all 0.2s ease;
  }

  .btn-outline:hover {
    background-color: #f9fafb;
  }

  .fade-in {
    opacity: 0;
    animation: fadeIn 0.3s ease-out forwards;
  }

  .waiting-card.fade-in {
    animation-delay: 0.03s;
  }

  .card-header.fade-in {
    animation-delay: 0.08s;
  }

  .status-badge-container.fade-in {
    animation-delay: 0.13s;
  }

  .approval-box-wrapper.fade-in {
    animation-delay: 0.18s;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
