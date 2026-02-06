<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let settings = $state([]);
  let values = $state({});
  let loading = $state(false);

  let twoFaQr = $state('');
  let twoFaSecret = $state('');
  let twoFaCode = $state('');
  let twoFaLoading = $state(false);
  let twoFaVerifyLoading = $state(false);
  let twoFaDisableLoading = $state(false);
  let twoFaError = $state('');
  let twoFaSuccess = $state('');
  let highlight2FA = $state(false);

  const is2FAEnabled = $derived(!!auth.user?.is_2fa_enabled);

  async function fetchSettings() {
    loading = true;
    try {
      const res = await api.getSettings();
      settings = res.data || [];
      values = {};
      settings.forEach((s) => (values[s.key] = s.value || ''));
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function saveSettings() {
    try {
      await api.updateSettings(values);
      await fetchSettings();
    } catch (err) {
      alert(err.message || 'Gagal menyimpan');
    }
  }

  function reset2FAState() {
    twoFaQr = '';
    twoFaSecret = '';
    twoFaCode = '';
    twoFaError = '';
    twoFaSuccess = '';
  }

  async function handleSetup2FA() {
    twoFaLoading = true;
    twoFaError = '';
    twoFaSuccess = '';
    try {
      const res = await api.setup2FA();
      twoFaQr = res.data?.qr_code || '';
      twoFaSecret = res.data?.secret || '';
      if (!twoFaQr) {
        twoFaError = 'QR code tidak tersedia. Coba ulangi.';
      }
    } catch (err) {
      twoFaError = err.message || 'Gagal memulai setup 2FA';
    } finally {
      twoFaLoading = false;
    }
  }

  async function handleVerify2FA() {
    if (!twoFaCode) return;
    twoFaVerifyLoading = true;
    twoFaError = '';
    twoFaSuccess = '';
    try {
      await api.verify2FA(twoFaCode);
      const res = await api.getCurrentUser();
      auth.hydrate(res.data);
      twoFaSuccess = '2FA berhasil diaktifkan.';
      reset2FAState();
    } catch (err) {
      twoFaError = err.message || 'Verifikasi 2FA gagal';
    } finally {
      twoFaVerifyLoading = false;
    }
  }

  async function handleDisable2FA() {
    if (!confirm('Nonaktifkan 2FA untuk akun ini?')) return;
    twoFaDisableLoading = true;
    twoFaError = '';
    twoFaSuccess = '';
    try {
      await api.disable2FA();
      const res = await api.getCurrentUser();
      auth.hydrate(res.data);
      twoFaSuccess = '2FA berhasil dinonaktifkan.';
      reset2FAState();
    } catch (err) {
      twoFaError = err.message || 'Gagal menonaktifkan 2FA';
    } finally {
      twoFaDisableLoading = false;
    }
  }

  onMount(fetchSettings);
  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('setup') === '2fa') {
      highlight2FA = true;
    }
  });
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Pengaturan Sistem</h3>
  <p class="text-muted">Kelola konfigurasi kantor, jam kerja, dan lainnya.</p>
</div>

<div class="card" style={`margin-bottom:16px; ${highlight2FA ? 'border-color:#f59e0b; background:#fffbeb;' : ''}`}>
  <h3>Keamanan 2FA</h3>
  <p class="text-muted">Aktifkan autentikasi dua langkah untuk keamanan akun.</p>

  {#if twoFaError}
    <div class="badge danger" style="margin-bottom:12px;">{twoFaError}</div>
  {/if}
  {#if twoFaSuccess}
    <div class="badge success" style="margin-bottom:12px;">{twoFaSuccess}</div>
  {/if}

  {#if is2FAEnabled}
    <div class="badge success" style="margin-bottom:12px;">2FA aktif</div>
    <button class="btn btn-outline" onclick={handleDisable2FA} disabled={twoFaDisableLoading}>
      {twoFaDisableLoading ? 'Memproses...' : 'Nonaktifkan 2FA'}
    </button>
  {:else}
    <div class="badge warning" style="margin-bottom:12px;">2FA belum aktif</div>
    <button class="btn btn-primary" onclick={handleSetup2FA} disabled={twoFaLoading}>
      {twoFaLoading ? 'Membuat QR...' : 'Generate QR Code'}
    </button>

    {#if twoFaQr}
      <div style="margin-top:16px; display:grid; gap:12px; max-width:320px;">
        <img src={twoFaQr} alt="QR Code 2FA" style="width:200px; height:200px; border:1px solid #e5e7eb; border-radius:8px;" />
        {#if twoFaSecret}
          <div class="text-muted" style="font-size:12px;">
            Secret: <span style="font-family:ui-monospace, SFMono-Regular, Menlo, monospace;">{twoFaSecret}</span>
          </div>
        {/if}
        <div class="form-group">
          <label class="form-label" for="twofa-code">Kode 2FA</label>
          <input class="input" id="twofa-code" type="text" inputmode="numeric" placeholder="123456" bind:value={twoFaCode} />
        </div>
        <button class="btn btn-primary" onclick={handleVerify2FA} disabled={twoFaVerifyLoading || !twoFaCode}>
          {twoFaVerifyLoading ? 'Memverifikasi...' : 'Verifikasi & Aktifkan'}
        </button>
      </div>
    {/if}
  {/if}
</div>

<div class="card">
  {#if loading}
    <div>Memuat...</div>
  {:else if settings.length === 0}
    <div class="empty-state">Belum ada pengaturan.</div>
  {:else}
    <div style="display:grid; gap:12px;">
      {#each settings as s}
        <div class="form-group">
          <label class="form-label" for={s.key}>{s.key}</label>
          <input class="input" bind:value={values[s.key]} id={s.key} />
          <small class="text-muted">{s.description || ''}</small>
        </div>
      {/each}
    </div>
    <button class="btn btn-primary" onclick={saveSettings}>Simpan</button>
  {/if}
</div>
