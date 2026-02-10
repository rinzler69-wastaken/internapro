<script>
  import { onMount } from 'svelte';
  import { replace } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let { id } = $props();
  let loading = $state(true);
  let error = $state('');
  let data = $state(null);

  function formatDate(value) {
    if (!value) return '-';
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return value;
    return d.toLocaleDateString('id-ID', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' });
  }

  function formatTime(value) {
    if (!value) return '-';
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return '-';
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
  }

  function withBase(url) {
    if (!url) return '';
    const needsToken = url.startsWith('/uploads') || url.startsWith('uploads/');
    const token = auth.token;
    const tokenQS = needsToken && token ? `${url.includes('?') ? '&' : '?'}token=${token}` : '';
    if (url.startsWith('http')) return `${url}${tokenQS}`;
    const base = import.meta.env.VITE_API_URL || '';
    const normalized = url.startsWith('/') ? url : `/${url}`;
    return `${base}${normalized}${tokenQS}`;
  }

  function statusBadge(status) {
    switch (status) {
      case 'present': return { text: 'Hadir', cls: 'bg-emerald-100 text-emerald-700 border-emerald-200' };
      case 'late': return { text: 'Terlambat', cls: 'bg-amber-100 text-amber-700 border-amber-200' };
      case 'permission': return { text: 'Izin', cls: 'bg-purple-100 text-purple-700 border-purple-200' };
      case 'sick': return { text: 'Sakit', cls: 'bg-blue-100 text-blue-700 border-blue-200' };
      case 'absent': return { text: 'Tidak Hadir', cls: 'bg-rose-100 text-rose-700 border-rose-200' };
      default: return { text: status || '-', cls: 'bg-slate-100 text-slate-600 border-slate-200' };
    }
  }

  async function load() {
    loading = true;
    error = '';
    try {
      const res = await api.getAttendanceById(id);
      data = res.data;
    } catch (err) {
      error = err.message || 'Gagal memuat presensi';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (!auth.token) {
      replace('/login');
      return;
    }
    load();
  });
</script>

<div class="page-shell">
  <div class="page-header">
    <button class="pill-btn" onclick={() => window.history.back()}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
      Kembali
    </button>
    <div>
      <h2 class="title">Detail Presensi</h2>
      {#if data}
        <p class="subtitle">{data.intern_name || auth.user?.name} · {formatDate(data.date || data.check_in_time)}</p>
      {/if}
    </div>
  </div>

  {#if loading}
    <div class="card">Memuat...</div>
  {:else if error}
    <div class="card error">{error}</div>
  {:else if data}
    <div class="grid lg:grid-cols-2 gap-4">
      <div class="card">
        <div class="card-head">
          <h3>Informasi Presensi</h3>
          {#if data.status}
            {@const badge = statusBadge(data.status)}
            <span class={`badge ${badge.cls}`}>{badge.text}</span>
          {/if}
        </div>
        <div class="info-grid">
          <div>
            <p class="label">Tanggal</p>
            <p class="value">{formatDate(data.date || data.check_in_time)}</p>
          </div>
          <div>
            <p class="label">Siswa</p>
            <p class="value">{data.intern_name || auth.user?.name || '-'}</p>
          </div>
          <div>
            <p class="label">Presensi Masuk</p>
            <p class="value">{formatTime(data.check_in_time)}</p>
          </div>
          <div>
            <p class="label">Presensi Keluar</p>
            <p class="value">{formatTime(data.check_out_time)}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-head">
          <h3>Detail Tambahan</h3>
        </div>
        <div class="info-grid">
          <div>
            <p class="label">Alasan Terlambat</p>
            <p class="value muted">{data.late_reason || 'Tidak ada'}</p>
          </div>
          <div>
            <p class="label">Catatan / Izin</p>
            <p class="value muted">{data.notes || 'Tidak ada catatan'}</p>
          </div>
          <div>
            <p class="label">Bukti Izin</p>
            {#if data.proof_file}
              <a class="link" href={withBase(data.proof_file)} target="_blank" rel="noreferrer">Lihat Dokumen</a>
            {:else}
              <p class="value muted">Tidak ada file</p>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .page-shell { max-width: 1400px; margin: 0 auto; padding: 1rem; }
  .page-header { display:flex; align-items:center; gap:12px; margin-bottom:1rem; }
  .pill-btn { display:inline-flex; align-items:center; gap:6px; padding:8px 14px; border:1px solid #e5e7eb; border-radius:999px; background:#fff; cursor:pointer; font-weight:600; color:#0f172a; }
  .pill-btn:hover { border-color:#cbd5e1; background:#f8fafc; }
  .title { font-size:1.25rem; font-weight:700; color:#0f172a; margin:0; }
  .subtitle { color:#64748b; margin:2px 0 0 0; font-size:0.95rem; }
  .card { background:#fff; border:1px solid #e5e7eb; border-radius:16px; padding:16px; box-shadow:0 10px 30px -18px rgba(15,23,42,0.2); }
  .card.error { color:#b91c1c; border-color:#fecdd3; background:#fff1f2; }
  .card-head { display:flex; align-items:center; justify-content:space-between; margin-bottom:12px; }
  .info-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(180px,1fr)); gap:12px; }
  .label { font-size:0.8rem; color:#94a3b8; margin:0; text-transform:uppercase; letter-spacing:0.02em; }
  .value { font-size:1rem; font-weight:600; color:#0f172a; margin:4px 0 0 0; }
  .value.muted { color:#94a3b8; font-weight:500; }
  .badge { padding:6px 10px; border-radius:999px; font-size:0.8rem; font-weight:700; border:1px solid transparent; }
  .link { color:#2563eb; font-weight:600; text-decoration:none; }
  .link:hover { text-decoration:underline; }
  @media (max-width: 768px) {
    .page-shell { padding: 0.75rem; }
    .card { padding:12px; }
  }
</style>
