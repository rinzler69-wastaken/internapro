<script>
  import { onMount } from 'svelte';
  import { replace } from '@mateothegreat/svelte5-router'; // Sesuaikan router kamu
  import { api } from '../lib/api.js';

  // State
  let name = $state('');
  let school = $state('');
  let major = $state('');
  let phone = $state('');
  let email = $state(''); // Ditambah email untuk pendaftaran manual
  let password = $state('');
  let confirmPassword = $state('');
  let address = $state('');
  let startDate = $state('');
  let endDate = $state('');
  let nis = $state('');
  let supervisorId = $state('');
  let loading = $state(false);
  
  // Deteksi apakah ini lemparan dari Google Login yang gagal
  // Contoh URL: /register?error=google_unregistered
  let isGoogleError = $state(false);

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('error') === 'google_unregistered') {
        isGoogleError = true;
        // Opsional: Ambil nama/email dari URL jika backend mengirimnya
        // name = params.get('name') || '';
    }
    const emailParam = params.get('email');
    const nameParam = params.get('name');
    if (emailParam) email = emailParam;
    if (nameParam) name = nameParam;
  });

  async function handleRegister() {
    loading = true;
    try {
      // Panggil API register kamu
      await api.post('/internship/register', { name, email, school, department: major, phone, password, confirm_password: confirmPassword, address, start_date: startDate, end_date: endDate, nis, supervisor_id: supervisorId || null });
      // logout so pending user must re-login after approval
      try { await api.logout(); } catch (e) {}
      replace('/waiting-approval');
    } catch (err) {
      const msg = err.message || 'Gagal mendaftar';
      const status = err.response?.status;

      if (status === 409 || msg.toLowerCase().includes('sudah terdaftar')) {
        // Coba deteksi apakah profil magang sudah pending/ada
        try {
          const prof = await api.getProfile();
          const intern = prof.data?.intern;
          if (intern?.status === 'pending' || intern?.id) {
            alert('Data sudah ada. Menunggu persetujuan admin.');
            replace('/waiting-approval');
            return;
          }
        } catch (_) {
          /* ignore and fall through */
        }
        alert('Email sudah terdaftar. Silakan login dengan akun tersebut.');
        replace('/login');
      } else {
        alert('Gagal mendaftar: ' + msg);
      }
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-container">
  <div class="card fade-in">
    
    <!-- HEADER LOGIC -->
    {#if isGoogleError}
        <!-- TAMPILAN JIKA DARI GOOGLE LOGIN (ERROR) -->
        <div style="text-align: center; margin-bottom: 2rem;">
            <div class="alert-box">
                <span class="alert-title">Anda tidak terdaftar di website kami</span>
                <span class="alert-subtitle">Silahkan mendaftar terlebih dahulu</span>
            </div>
        </div>
    {:else}
        <!-- TAMPILAN NORMAL (MANUAL) -->
        <div class="card-header">
            <div class="brand-logo">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="24" height="24">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                    <polyline points="14 2 14 8 20 8"></polyline>
                </svg>
            </div>
            <h2 class="title">Daftar Magang Baru</h2>
            <p class="subtitle">Lengkapi formulir di bawah untuk mendaftar.</p>
        </div>
    {/if}

    <div class="form-body">
      <div class="form-group">
        <label for="name">Nama Lengkap</label>
        <input class="input-field" type="text" id="name" bind:value={name} placeholder="Nama sesuai identitas" />
      </div>

      <div class="form-group">
        <label for="email">Email</label>
        <input class="input-field" type="email" id="email" bind:value={email} placeholder="email@sekolah.com" />
      </div>

      <div class="form-group">
        <label for="school">Asal Sekolah / Kampus</label>
        <input class="input-field" type="text" id="school" bind:value={school} placeholder="Contoh: SMK Negeri 1" />
      </div>

      <div class="form-group">
        <label for="major">Jurusan</label>
        <input class="input-field" type="text" id="major" bind:value={major} placeholder="Contoh: Teknik Komputer Jaringan" />
      </div>

      <div class="form-group">
        <label for="phone">No. WhatsApp</label>
        <input class="input-field" type="text" id="phone" bind:value={phone} placeholder="0812..." />
      </div>

      <div class="form-group">
        <label for="nis">NIS / NIM</label>
        <input class="input-field" type="text" id="nis" bind:value={nis} placeholder="Nomor induk siswa/mahasiswa" />
      </div>

      <div class="form-group">
        <label for="address">Alamat</label>
        <textarea class="input-field" id="address" rows="2" bind:value={address} placeholder="Alamat tempat tinggal"></textarea>
      </div>

      <div class="form-group">
        <label for="start">Tanggal Mulai</label>
        <input class="input-field" type="date" id="start" bind:value={startDate} />
      </div>

      <div class="form-group">
        <label for="end">Tanggal Selesai</label>
        <input class="input-field" type="date" id="end" bind:value={endDate} />
      </div>

      <div class="form-group">
        <label for="supervisor">Pembimbing (opsional)</label>
        <input class="input-field" type="number" id="supervisor" bind:value={supervisorId} placeholder="ID pembimbing (jika ada)" />
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <input class="input-field" type="password" id="password" bind:value={password} placeholder="Minimal 6 karakter" />
      </div>

      <div class="form-group">
        <label for="confirmPassword">Konfirmasi Password</label>
        <input class="input-field" type="password" id="confirmPassword" bind:value={confirmPassword} placeholder="Ulangi password" />
      </div>

      <button class="btn btn-primary" onclick={handleRegister} disabled={loading}>
        {#if loading} Mengirim... {:else} Kirim Pendaftaran {/if}
      </button>

      <div style="text-align: center;">
         <a href="/login" class="link-muted">Kembali ke Login</a>
      </div>
    </div>
  </div>
</div>

<style>
  /* Copy CSS dari .login-container sampai ke bawah dari file HTML sebelumnya */
  /* Pastikan sertakan class .alert-box, .alert-title, .alert-subtitle */
  .login-container { display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 1rem; background-color: #f3f4f6; }
  .card { background: white; width: 100%; max-width: 400px; border-radius: 12px; padding: 2.5rem 2rem; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
  .card-header { text-align: center; margin-bottom: 2rem; }
  .brand-logo { display: inline-flex; align-items: center; justify-content: center; width: 48px; height: 48px; background-color: #eff6ff; color: #111827; border-radius: 10px; margin-bottom: 1rem; }
  .title { font-size: 1.5rem; font-weight: 600; margin-bottom: 0.5rem; }
  .subtitle { color: #6b7280; font-size: 0.875rem; margin: 0; }
  .form-group { margin-bottom: 1.25rem; }
  label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.5rem; }
  .input-field { width: 100%; padding: 0.75rem 1rem; border: 1px solid #d1d5db; border-radius: 8px; }
  .btn { width: 100%; padding: 0.75rem; border-radius: 8px; font-weight: 600; cursor: pointer; border: none; margin-bottom: 1rem; }
  .btn-primary { background-color: #111827; color: white; }
  .link-muted { color: #6b7280; text-decoration: none; font-size: 0.85rem; }
  
  /* Alert Styles Khusus Page Ini */
  .alert-box { background-color: #fef2f2; border: 1px solid #fee2e2; padding: 1rem; border-radius: 8px; margin-bottom: 1.5rem; text-align: center; }
  .alert-title { color: #dc2626; font-weight: 600; font-size: 0.95rem; display: block; }
  .alert-subtitle { color: #ef4444; font-size: 0.85rem; display: block; }
</style>
