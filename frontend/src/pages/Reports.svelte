<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let reports = $state([]);
  let interns = $state([]);
  let loading = $state(true);
  let submitting = $state(false);

  // Form State
  let form = $state({
    intern_id: '',
    title: '',
    content: '',
    type: 'weekly',
    period_start: '',
    period_end: '',
  });

  async function fetchReports() {
    loading = true;
    try {
      const res = await api.getReports({ page: 1, limit: 50 });
      reports = res.data || [];
      
      // Fetch data intern hanya jika user bukan intern (untuk dropdown)
      if (auth.user?.role !== 'intern') {
        const internRes = await api.getInterns({ page: 1, limit: 100, status: 'active' });
        interns = internRes.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function handleCreate() {
    // Validasi sederhana
    if (!form.title || !form.content || !form.period_start || !form.period_end) {
        alert("Mohon lengkapi semua field wajib.");
        return;
    }

    submitting = true;
    try {
      await api.createReport({
        ...form,
        intern_id: form.intern_id ? Number(form.intern_id) : null,
      });
      
      alert('Laporan berhasil dibuat!');
      
      // Reset form (kecuali intern_id jika admin ingin input lagi untuk org yg sama)
      form.title = '';
      form.content = '';
      form.period_start = '';
      form.period_end = '';
      
      await fetchReports();
    } catch (err) {
      alert(err.message || 'Gagal membuat laporan');
    } finally {
      submitting = false;
    }
  }

  // Helpers UI
  function getTypeColor(type) {
      switch(type) {
          case 'weekly': return 'bg-blue-100 text-blue-700 border-blue-200';
          case 'monthly': return 'bg-purple-100 text-purple-700 border-purple-200';
          case 'final': return 'bg-rose-100 text-rose-700 border-rose-200';
          default: return 'bg-slate-100 text-slate-600 border-slate-200';
      }
  }

  function getStatusColor(status) {
      switch(status) {
          case 'approved': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
          case 'rejected': return 'bg-red-100 text-red-700 border-red-200';
          case 'pending': return 'bg-amber-100 text-amber-700 border-amber-200';
          default: return 'bg-slate-100 text-slate-600 border-slate-200';
      }
  }

  function formatDate(dateStr) {
      if(!dateStr) return '-';
      return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
  }

  onMount(fetchReports);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <div>
            <h2 class="title">Laporan & Jurnal</h2>
            <p class="subtitle">Dokumentasi aktivitas mingguan dan bulanan peserta magang.</p>
        </div>
        <div class="header-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                <polyline points="14 2 14 8 20 8"></polyline>
                <line x1="16" y1="13" x2="8" y2="13"></line>
                <line x1="16" y1="17" x2="8" y2="17"></line>
                <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
        </div>
    </div>

    <!-- FORM TAMBAH LAPORAN (Hanya Admin/Supervisor) -->
    {#if auth.user?.role !== 'intern'}
      <div class="card form-card animate-slide-up">
        <div class="card-header">
            <h3>Buat Laporan Baru</h3>
        </div>
        
        <div class="form-body">
            <!-- Baris 1: Intern & Tipe -->
            <div class="grid-2">
                <div class="form-group">
                    <label class="label" for="intern_id">Pilih Peserta</label>
                    <div class="select-wrapper">
                        <select class="input-field select" bind:value={form.intern_id} id="intern_id">
                            <option value="">-- Pilih Intern --</option>
                            {#each interns as i}
                                <option value={i.id}>{i.full_name}</option>
                            {/each}
                        </select>
                        <div class="select-arrow">▼</div>
                    </div>
                </div>
                <div class="form-group">
                    <label class="label" for="type">Tipe Laporan</label>
                    <div class="select-wrapper">
                        <select class="input-field select" bind:value={form.type} id="type">
                            <option value="weekly">Laporan Mingguan</option>
                            <option value="monthly">Laporan Bulanan</option>
                            <option value="final">Laporan Akhir</option>
                        </select>
                        <div class="select-arrow">▼</div>
                    </div>
                </div>
            </div>

            <!-- Baris 2: Judul -->
            <div class="form-group mt-4">
                <label class="label" for="title">Judul Kegiatan</label>
                <input class="input-field" bind:value={form.title} id="title" placeholder="Contoh: Implementasi UI Dashboard..." />
            </div>

            <!-- Baris 3: Periode -->
            <div class="grid-2 mt-4">
                <div class="form-group">
                    <label class="label" for="period_start">Tanggal Mulai</label>
                    <input class="input-field" type="date" bind:value={form.period_start} id="period_start" />
                </div>
                <div class="form-group">
                    <label class="label" for="period_end">Tanggal Selesai</label>
                    <input class="input-field" type="date" bind:value={form.period_end} id="period_end" />
                </div>
            </div>

            <!-- Baris 4: Konten -->
            <div class="form-group mt-4">
                <label class="label" for="content">Isi Laporan / Deskripsi Kegiatan</label>
                <textarea class="input-field textarea" rows="4" bind:value={form.content} id="content" placeholder="Jelaskan detail aktivitas yang dilakukan..."></textarea>
            </div>

            <div class="action-row mt-6">
                <button class="btn-primary" onclick={handleCreate} disabled={submitting}>
                    {#if submitting}
                        Menyimpan...
                    {:else}
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                        Simpan Laporan
                    {/if}
                </button>
            </div>
        </div>
      </div>
    {/if}

    <!-- TABEL DAFTAR LAPORAN -->
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
        <div class="card-header border-b">
            <h3>Daftar Laporan Masuk</h3>
            <span class="badge-count">{reports.length} File</span>
        </div>

        {#if loading}
            <div class="loading-state">Memuat data laporan...</div>
        {:else if reports.length === 0}
            <div class="empty-state">
                <div class="empty-icon">📂</div>
                <p>Belum ada laporan yang dikumpulkan.</p>
            </div>
        {:else}
            <div class="table-container">
                <table class="table">
                    <thead>
                        <tr>
                            <th>Judul & Tipe</th>
                            <th>Peserta</th>
                            <th>Periode</th>
                            <th>Status</th>
                            <th class="text-right">Opsi</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each reports as r}
                            <tr class="hover-row">
                                <td style="min-width: 240px;">
                                    <div class="report-info">
                                        <span class={`type-badge ${getTypeColor(r.type)}`}>{r.type}</span>
                                        <span class="report-title">{r.title || 'Tanpa Judul'}</span>
                                    </div>
                                </td>
                                <td>
                                    <div class="user-info">
                                        <div class="avatar">{r.intern_name?.charAt(0) || 'U'}</div>
                                        <div class="user-text">
                                            <span class="name">{r.intern_name || 'Intern'}</span>
                                        </div>
                                    </div>
                                </td>
                                <td class="text-slate-500 text-sm">
                                    {formatDate(r.period_start)} - {formatDate(r.period_end)}
                                </td>
                                <td>
                                    <span class={`status-badge ${getStatusColor(r.status || 'pending')}`}>
                                        {r.status || 'Pending'}
                                    </span>
                                </td>
                                <td class="text-right">
                                    <button class="btn-icon" title="Lihat Detail">
                                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>

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
    padding: 32px 24px;
  }

  .container { max-width: 1000px; margin: 0 auto; }

  /* --- HEADER --- */
  .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px; }
  .title { font-size: 28px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; }
  .subtitle { color: #64748b; font-size: 15px; margin: 0; }
  .header-icon {
    width: 48px; height: 48px; background: #ffffff; border-radius: 12px;
    display: flex; align-items: center; justify-content: center; color: #10b981;
    box-shadow: 0 4px 12px rgba(0,0,0,0.03); border: 1px solid #e2e8f0;
  }

  /* --- CARD --- */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02), 0 2px 4px -1px rgba(0,0,0,0.02);
    overflow: hidden; margin-bottom: 32px;
  }
  .card-header {
    padding: 20px 24px; display: flex; justify-content: space-between; align-items: center;
  }
  .card-header h3 { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; }
  .border-b { border-bottom: 1px solid #f1f5f9; }

  /* --- FORM STYLING --- */
  .form-body { padding: 24px; }
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  @media (max-width: 640px) { .grid-2 { grid-template-columns: 1fr; } }
  
  .form-group { display: flex; flex-direction: column; gap: 8px; }
  .label {
    font-size: 13px; font-weight: 600; color: #334155; text-transform: uppercase; letter-spacing: 0.02em;
  }
  
  .input-field {
    padding: 12px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; font-family: 'Inter', sans-serif; color: #0f172a;
    transition: all 0.2s; background: #fff; width: 100%; box-sizing: border-box;
  }
  .input-field:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15); }
  .textarea { resize: vertical; line-height: 1.5; }

  /* Select Custom */
  .select-wrapper { position: relative; }
  .select { appearance: none; cursor: pointer; }
  .select-arrow {
    position: absolute; right: 14px; top: 50%; transform: translateY(-50%);
    font-size: 10px; color: #64748b; pointer-events: none;
  }

  .action-row { display: flex; justify-content: flex-end; }
  
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white;
    border: none; padding: 12px 24px; border-radius: 10px; font-weight: 600; font-size: 14px;
    cursor: pointer; display: flex; align-items: center; gap: 8px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2);
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 15px rgba(16, 185, 129, 0.3); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  /* --- TABLE --- */
  .badge-count { background: #f1f5f9; color: #64748b; padding: 4px 10px; border-radius: 20px; font-size: 12px; font-weight: 600; }

  .table-container { overflow-x: auto; }
  .table { width: 100%; border-collapse: separate; border-spacing: 0; }
  
  .table th {
    text-align: left; padding: 16px 24px; font-size: 12px; font-weight: 600;
    text-transform: uppercase; color: #64748b; border-bottom: 1px solid #e2e8f0; background: #fcfcfc;
  }
  .table td {
    padding: 16px 24px; border-bottom: 1px solid #f1f5f9; vertical-align: middle; color: #334155;
  }
  .table tr:last-child td { border-bottom: none; }
  .hover-row:hover td { background-color: #f8fafc; }

  /* Table Content */
  .report-info { display: flex; flex-direction: column; gap: 4px; }
  .report-title { font-weight: 600; color: #0f172a; font-size: 14px; }
  
  .user-info { display: flex; align-items: center; gap: 12px; }
  .avatar {
    width: 32px; height: 32px; background: #0f172a; color: white; border-radius: 8px;
    display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 12px;
  }
  .name { font-weight: 500; font-size: 14px; }

  .text-right { text-align: right; }
  .text-slate-500 { color: #64748b; }
  .text-sm { font-size: 13px; }

  /* Badges */
  .type-badge {
    display: inline-block; padding: 2px 8px; border-radius: 6px;
    font-size: 10px; font-weight: 600; text-transform: uppercase;
    width: fit-content; border: 1px solid transparent;
  }
  .bg-blue-100 { background: #eff6ff; border-color: #bfdbfe; }
  .bg-purple-100 { background: #faf5ff; border-color: #e9d5ff; }
  .bg-rose-100 { background: #fff1f2; border-color: #fecdd3; }

  .status-badge {
    display: inline-flex; align-items: center; padding: 4px 12px;
    border-radius: 99px; font-size: 12px; font-weight: 600; border: 1px solid transparent; text-transform: capitalize;
  }
  .bg-emerald-100 { background: #ecfdf5; border-color: #a7f3d0; } 
  .bg-amber-100 { background: #fefce8; border-color: #fef08a; } 
  .bg-red-100 { background: #fef2f2; border-color: #fecaca; } 
  .bg-slate-100 { background: #f1f5f9; border-color: #e2e8f0; }

  .btn-icon {
    background: transparent; border: none; color: #94a3b8; cursor: pointer;
    padding: 6px; border-radius: 6px; transition: all 0.2s;
  }
  .btn-icon:hover { background: #e2e8f0; color: #0f172a; }

  /* States */
  .empty-state, .loading-state { text-align: center; padding: 60px 20px; color: #94a3b8; font-style: italic; }
  .empty-icon { font-size: 32px; margin-bottom: 12px; opacity: 0.5; }

  /* Animation */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>