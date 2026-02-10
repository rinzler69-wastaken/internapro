<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // Props dari router
  const { route } = $props(); 
  
  // State
  let assignmentId = $state('');
  let data = $state(null);
  let loading = $state(true);

  async function fetchAssignment() {
    loading = true;
    try {
      const res = await api.getTaskAssignment(assignmentId);
      data = res.data;
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  // Reactive effect untuk mendeteksi perubahan ID di URL
  $effect(() => {
    const params = route?.result?.path?.params || {};
    if (params?.id && params.id !== assignmentId) {
      assignmentId = params.id;
      fetchAssignment();
    }
  });

  // Helper warna status
  function getStatusColor(status) {
    switch (status) {
      case 'completed': return 'bg-green-soft text-green';
      case 'approved': return 'bg-green-soft text-green';
      case 'submitted': return 'bg-blue-soft text-blue';
      case 'in_progress': return 'bg-amber-soft text-amber';
      case 'pending': return 'bg-slate-soft text-slate';
      case 'late': return 'bg-red-soft text-red';
      default: return 'bg-slate-soft text-slate';
    }
  }

  function getStatusLabel(status) {
    switch (status) {
      case 'completed': return 'Selesai';
      case 'approved': return 'Disetujui';
      case 'submitted': return 'Menunggu Review';
      case 'in_progress': return 'Dikerjakan';
      case 'pending': return 'Belum Diambil';
      default: return status;
    }
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header with Back Button -->
    <div class="header">
        <a href="/dashboard" class="btn-back">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5"/><path d="M12 19l-7-7 7-7"/></svg>
            Kembali
        </a>
    </div>

    {#if loading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Memuat detail tugas...</p>
        </div>
    {:else if !data}
        <div class="empty-state">
            <div class="empty-icon">📂</div>
            <h3>Data Tidak Ditemukan</h3>
            <p>Tugas yang Anda cari mungkin telah dihapus atau tidak tersedia.</p>
            <a href="/dashboard" class="btn-primary">Ke Dashboard</a>
        </div>
    {:else}
    
        <!-- Info Utama Tugas -->
        <div class="content-grid animate-slide-up">
            
            <!-- Kolom Kiri: Detail & Stats -->
            <div class="left-col">
                <div class="card detail-card">
                    <div class="card-header">
                        <div class="icon-wrapper bg-green-gradient">
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                        </div>
                        <div>
                            <h1 class="title">{data.assignment.title}</h1>
                            <p class="meta">
                                Dibuat pada {new Date(data.assignment.created_at || Date.now()).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })}
                            </p>
                        </div>
                    </div>
                    
                    <div class="description">
                        <p>{data.assignment.description || 'Tidak ada deskripsi tambahan untuk tugas ini.'}</p>
                    </div>

                    <!-- Statistik Grid -->
                    <div class="stats-mini-grid">
                        <div class="stat-item">
                            <span class="stat-label">Total Partisipan</span>
                            <span class="stat-val">{data.stats.total}</span>
                        </div>
                        <div class="stat-item green">
                            <span class="stat-label">Selesai</span>
                            <span class="stat-val">{data.stats.completed}</span>
                        </div>
                        <div class="stat-item amber">
                            <span class="stat-label">Proses</span>
                            <span class="stat-val">{data.stats.in_progress}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Nilai Rata-rata</span>
                            <span class="stat-val">{data.stats.average_score || 0}</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Kolom Kanan / Bawah: Daftar Pengerjaan -->
            <div class="right-col">
                <div class="card list-card">
                    <div class="card-title-row">
                        <h3>Status Pengerjaan</h3>
                        <span class="badge-count">{data.tasks.length} Siswa</span>
                    </div>

                    <div class="table-container">
                        <table class="table">
                            <thead>
                                <tr>
                                    <th>Judul / Subtask</th>
                                    <th>Siswa</th>
                                    <th>Status</th>
                                    <th>Dikumpulkan</th>
                                    <th>Nilai</th>
                                    <th class="text-right">Aksi</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each data.tasks as t}
                                    <tr class="hover-row">
                                        <td>
                                            <span class="task-title">{t.title}</span>
                                        </td>
                                        <td>
                                            <div class="user-info">
                                                <div class="avatar-mini">{t.intern_name?.charAt(0) || 'U'}</div>
                                                <span class="user-name">{t.intern_name || 'Tidak Diketahui'}</span>
                                            </div>
                                        </td>
                                        <td>
                                            <span class={`status-badge ${getStatusColor(t.status)}`}>
                                                {getStatusLabel(t.status)}
                                            </span>
                                        </td>
                                        <td>
                                            {#if t.submitted_at}
                                                {new Date(t.submitted_at).toLocaleDateString('id-ID', { day:'2-digit', month:'short' })} {new Date(t.submitted_at).toLocaleTimeString('id-ID', { hour:'2-digit', minute:'2-digit' })}
                                                {#if t.is_late}
                                                    <span class="chip-late">Terlambat</span>
                                                {/if}
                                            {:else}
                                                -
                                            {/if}
                                        </td>
                                        <td>
                                            {#if t.score !== null && t.score !== undefined}
                                                <span class="score-pill">{t.score}</span>
                                            {:else}
                                                <span class="text-slate-400">-</span>
                                            {/if}
                                        </td>
                                        <td class="text-right">
                                            <a href={`/tasks/${t.id}`} class="btn-icon" title="Lihat Detail">
                                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                                            </a>
                                        </td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                        
                        {#if data.tasks.length === 0}
                            <div class="empty-list">
                                Belum ada siswa yang mengerjakan tugas ini.
                            </div>
                        {/if}
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
    padding: 32px 24px;
  }

  .container {
    max-width: 1000px;
    margin: 0 auto;
  }

  /* --- HEADER --- */
  .header {
    margin-bottom: 24px;
  }
  .btn-back {
    display: inline-flex; align-items: center; gap: 8px;
    color: #64748b; font-weight: 600; font-size: 14px;
    text-decoration: none; transition: all 0.2s;
    padding: 8px 12px; border-radius: 8px;
  }
  .btn-back:hover {
    background: #ffffff; color: #0f172a; box-shadow: 0 2px 4px rgba(0,0,0,0.05);
  }

  /* --- LAYOUT --- */
  .content-grid {
    display: flex; flex-direction: column; gap: 24px;
  }

  /* --- CARDS --- */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02);
    overflow: hidden;
  }

  /* --- DETAIL CARD --- */
  .detail-card {
    padding: 32px;
  }
  .card-header {
    display: flex; align-items: flex-start; gap: 20px; margin-bottom: 24px;
  }
  .icon-wrapper {
    width: 56px; height: 56px; border-radius: 16px;
    display: flex; align-items: center; justify-content: center;
    color: #059669; flex-shrink: 0;
  }
  .bg-green-gradient { background: linear-gradient(135deg, #ecfdf5, #d1fae5); }
  
  .title { font-size: 24px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; line-height: 1.2; }
  .meta { color: #94a3b8; font-size: 14px; margin: 0; }

  .description {
    font-size: 15px; line-height: 1.6; color: #475569;
    background: #f8fafc; padding: 20px; border-radius: 12px; border: 1px solid #f1f5f9;
    margin-bottom: 32px;
  }

  /* --- STATS MINI GRID --- */
  .stats-mini-grid {
    display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px;
  }
  @media (max-width: 600px) { .stats-mini-grid { grid-template-columns: 1fr; } }

  .stat-item {
    background: white; border: 1px solid #e2e8f0; border-radius: 16px;
    padding: 16px; text-align: center;
    transition: transform 0.2s;
  }
  .stat-item:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
  
  .stat-item.green { border-bottom: 4px solid #10b981; }
  .stat-item.amber { border-bottom: 4px solid #f59e0b; }
  
  .stat-label { display: block; font-size: 12px; font-weight: 600; text-transform: uppercase; color: #94a3b8; margin-bottom: 4px; }
  .stat-val { font-size: 24px; font-weight: 800; color: #0f172a; }

  /* --- LIST CARD --- */
  .list-card { padding: 0; }
  .card-title-row {
    padding: 20px 24px; border-bottom: 1px solid #f1f5f9;
    display: flex; justify-content: space-between; align-items: center;
  }
  .card-title-row h3 { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
  .badge-count {
    background: #f1f5f9; color: #64748b; font-size: 12px; font-weight: 600;
    padding: 4px 10px; border-radius: 20px;
  }

  .table-container { overflow-x: auto; }
  .table { width: 100%; border-collapse: separate; border-spacing: 0; }
  
  .table th {
    text-align: left; padding: 14px 24px;
    font-size: 12px; font-weight: 600; text-transform: uppercase; color: #64748b;
    background: #fcfcfc; border-bottom: 1px solid #e2e8f0;
  }
  .table td {
    padding: 16px 24px; border-bottom: 1px solid #f1f5f9; vertical-align: middle;
    font-size: 14px; color: #334155;
  }
  .table tr:last-child td { border-bottom: none; }
  .hover-row:hover td { background-color: #f8fafc; }

  /* Table Components */
  .task-title { font-weight: 600; color: #0f172a; }
  
  .user-info { display: flex; align-items: center; gap: 10px; }
  .avatar-mini {
    width: 32px; height: 32px; background: #0f172a; color: white;
    border-radius: 50%; display: flex; align-items: center; justify-content: center;
    font-size: 12px; font-weight: 600;
  }
  .user-name { font-weight: 500; }

  /* Status Badges */
  .status-badge {
    display: inline-flex; align-items: center; padding: 4px 12px;
    border-radius: 99px; font-size: 12px; font-weight: 600;
  }
  .bg-green-soft { background: #ecfdf5; } .text-green { color: #059669; }
  .bg-blue-soft { background: #eff6ff; } .text-blue { color: #2563eb; }
  .bg-amber-soft { background: #fffbeb; } .text-amber { color: #d97706; }
  .bg-slate-soft { background: #f1f5f9; } .text-slate { color: #64748b; }
  .bg-red-soft { background: #fef2f2; } .text-red { color: #dc2626; }
  .chip-late { background: #fef2f2; color: #b91c1c; padding: 2px 6px; border-radius: 8px; font-size: 11px; font-weight: 700; margin-left: 6px; }
  .score-pill { background: #ecfdf3; color: #047857; padding: 4px 10px; border-radius: 10px; font-weight: 700; }

  .text-right { text-align: right; }
  .btn-icon {
    display: inline-flex; padding: 8px; border-radius: 8px; color: #94a3b8; transition: all 0.2s;
  }
  .btn-icon:hover { background: #e2e8f0; color: #0f172a; }

  /* Empty & Loading */
  .empty-state, .loading-state { text-align: center; padding: 60px 20px; }
  .empty-icon { font-size: 40px; margin-bottom: 16px; opacity: 0.5; }
  .empty-list { text-align: center; padding: 40px; color: #94a3b8; font-style: italic; font-size: 14px; }
  .btn-primary { 
    background: #10b981; color: white; padding: 10px 20px; border-radius: 8px; 
    text-decoration: none; font-weight: 600; font-size: 14px; display: inline-block; margin-top: 16px;
  }
  .spinner {
    width: 40px; height: 40px; border: 3px solid #e2e8f0; border-top-color: #10b981;
    border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Animation */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.5s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.5s ease-out forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
