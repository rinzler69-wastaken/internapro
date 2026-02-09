<script>
  import { onMount } from 'svelte';
  import { goto, route } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let tasks = $state([]);
  let pagination = $state({});
  let search = $state('');
  let status = $state('');
  let priority = $state('');
  let loading = $state(true);

  // Constants
  const statusLabels = {
    pending: 'Pending',
    scheduled: 'Terjadwal',
    in_progress: 'Dalam Proses',
    submitted: 'Menunggu Review',
    revision: 'Revisi',
    completed: 'Selesai',
  };

  // Helpers
  function formatDate(value) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  function getPriorityColor(p) {
    switch(p) {
        case 'high': return 'bg-rose-50 text-rose-700 border-rose-100';
        case 'medium': return 'bg-amber-50 text-amber-700 border-amber-100';
        default: return 'bg-emerald-50 text-emerald-700 border-emerald-100';
    }
  }

  function getStatusColor(s) {
    switch(s) {
        case 'completed': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
        case 'submitted': return 'bg-blue-100 text-blue-700 border-blue-200';
        case 'in_progress': return 'bg-amber-100 text-amber-700 border-amber-200';
        case 'revision': return 'bg-rose-100 text-rose-700 border-rose-200';
        default: return 'bg-slate-100 text-slate-600 border-slate-200';
    }
  }

  // API Call
  async function fetchTasks() {
    loading = true;
    try {
      const params = { page: 1, limit: 50 }; // Default limit
      if (search) params.search = search;
      if (status) params.status = status;
      if (priority) params.priority = priority;

      // getTasks di api.js terbaru sudah support object params
      const res = await api.getTasks(params);
      tasks = res.data || [];
      pagination = res.pagination || {};
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  onMount(fetchTasks);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <div>
            <h2 class="title">Daftar Penugasan</h2>
            <p class="subtitle">Kelola tugas peserta magang dan pantau progres pengerjaan.</p>
        </div>
        {#if auth.user?.role !== 'intern'}
          <a href="/tasks/create" use:route class="btn-primary">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            Buat Tugas
          </a>
        {/if}
    </div>

    <!-- Filter Card -->
    <div class="card filter-card animate-slide-up">
        <div class="filter-grid">
            <div class="search-wrapper">
                <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                <input 
                    class="input-field pl-10" 
                    placeholder="Cari judul tugas..." 
                    bind:value={search} 
                    onkeydown={(e) => e.key === 'Enter' && fetchTasks()} 
                />
            </div>
            
            <div class="select-wrapper">
                <select class="input-field select" bind:value={status}>
                    <option value="">Semua Status</option>
                    <option value="pending">Pending</option>
                    <option value="in_progress">Dalam Proses</option>
                    <option value="submitted">Menunggu Review</option>
                    <option value="revision">Revisi</option>
                    <option value="completed">Selesai</option>
                </select>
                <div class="select-arrow">▼</div>
            </div>

            <div class="select-wrapper">
                <select class="input-field select" bind:value={priority}>
                    <option value="">Semua Prioritas</option>
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                </select>
                <div class="select-arrow">▼</div>
            </div>

            <button class="btn-outline" onclick={fetchTasks}>
                Filter
            </button>
        </div>
    </div>

    <!-- Task List -->
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
        {#if loading}
            <div class="loading-state">
                <div class="spinner"></div>
                <p>Memuat daftar tugas...</p>
            </div>
        {:else if tasks.length === 0}
            <div class="empty-state">
                <div class="empty-icon">📋</div>
                <p>Tidak ada tugas ditemukan.</p>
                {#if auth.user?.role !== 'intern'}
                    <p class="text-sm mt-2 text-slate-400">Coba ubah filter atau buat tugas baru.</p>
                {/if}
            </div>
        {:else}
            <div class="table-container">
                <table class="table">
                    <thead>
                        <tr>
                            <th>Judul Tugas</th>
                            {#if auth.user?.role !== 'intern'}
                                <th>Intern</th>
                            {/if}
                            <th>Status</th>
                            <th>Prioritas</th>
                            <th>Deadline</th>
                            <th class="text-right">Aksi</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each tasks as task}
                            <tr 
                                class="hover-row cursor-pointer" 
                                onclick={() => goto(`/tasks/${task.id}`)}
                            >
                                <td>
                                    <span class="task-title">{task.title}</span>
                                </td>
                                {#if auth.user?.role !== 'intern'}
                                    <td>
                                        <div class="user-info">
                                            <div class="avatar-mini">{task.intern_name?.charAt(0) || 'U'}</div>
                                            <span class="user-name">{task.intern_name || '-'}</span>
                                        </div>
                                    </td>
                                {/if}
                                <td>
                                    <span class={`status-badge ${getStatusColor(task.status)}`}>
                                        {statusLabels[task.status] || task.status || '-'}
                                    </span>
                                </td>
                                <td>
                                    <span class={`priority-badge ${getPriorityColor(task.priority)}`}>
                                        {task.priority || '-'}
                                    </span>
                                </td>
                                <td class="text-slate-500 text-sm font-mono">
                                    {formatDate(task.deadline)}
                                </td>
                                <td class="text-right">
                                    <button class="btn-icon">
                                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
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
  :global(body) { font-family: 'Geist', 'Inter', sans-serif; color: #0f172a; }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 0px 0px;
  }
  .container { max-width: 1400px; margin: 0 auto; }

  /* HEADER */
  .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px; }
  .title { font-size: 20px; font-weight: 600; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; }
  .subtitle { color: #64748b; font-size: 16px; margin: 0; }

  /* CARDS */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
  }
  .filter-card { padding: 20px; margin-bottom: 24px; }
  .list-card { padding-bottom: 8px; }

  /* FILTER BAR */
  .filter-grid {
    display: grid; grid-template-columns: 1fr; gap: 12px;
  }
  @media (min-width: 768px) {
    .filter-grid { grid-template-columns: 2fr 1fr 1fr auto; }
  }

  .input-field {
    width: 100%; padding: 10px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; transition: all 0.2s; background: #fff; box-sizing: border-box;
    font-family: 'Inter', sans-serif;
  }
  .input-field:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }
  .pl-10 { padding-left: 40px; }

  .search-wrapper { position: relative; }
  .search-icon { position: absolute; left: 12px; top: 12px; color: #94a3b8; pointer-events: none; }

  .select-wrapper { position: relative; }
  .select { appearance: none; cursor: pointer; }
  .select-arrow {
    position: absolute; right: 14px; top: 50%; transform: translateY(-50%);
    font-size: 10px; color: #64748b; pointer-events: none;
  }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white;
    padding: 10px 20px; border-radius: 10px; font-weight: 600; font-size: 14px; border: none;
    cursor: pointer; display: flex; align-items: center; gap: 8px; text-decoration: none;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); transition: all 0.2s;
  }
  .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4); }

  .btn-outline {
    background: white; border: 1px solid #cbd5e1; color: #475569;
    padding: 10px 20px; border-radius: 10px; font-weight: 600; font-size: 14px;
    cursor: pointer; transition: all 0.2s;
  }
  .btn-outline:hover { border-color: #10b981; color: #059669; background: #ecfdf5; }

  /* TABLE */
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
  .cursor-pointer { cursor: pointer; }

  .task-title { font-weight: 600; color: #0f172a; }
  .user-info { display: flex; align-items: center; gap: 10px; }
  .avatar-mini {
    width: 28px; height: 28px; background: #0f172a; color: white;
    border-radius: 50%; display: flex; align-items: center; justify-content: center;
    font-size: 11px; font-weight: 600;
  }
  
  /* BADGES */
  .status-badge {
    display: inline-flex; align-items: center; padding: 4px 12px;
    border-radius: 99px; font-size: 12px; font-weight: 600; text-transform: capitalize; border: 1px solid transparent;
  }
  .priority-badge {
    display: inline-block; padding: 2px 8px; border-radius: 6px;
    font-size: 11px; font-weight: 600; text-transform: uppercase; border: 1px solid transparent;
  }
  
  .bg-rose-50 { background: #fff1f2; } .border-rose-100 { border-color: #fecdd3; } .text-rose-700 { color: #be123c; }
  .bg-amber-50 { background: #fffbeb; } .border-amber-100 { border-color: #fde68a; } .text-amber-700 { color: #b45309; }
  .bg-emerald-50 { background: #ecfdf5; } .border-emerald-100 { border-color: #a7f3d0; } .text-emerald-700 { color: #047857; }

  .btn-icon {
    background: transparent; border: none; color: #cbd5e1; cursor: pointer;
    padding: 6px; border-radius: 6px; transition: all 0.2s;
  }
  .hover-row:hover .btn-icon { color: #10b981; }

  .text-right { text-align: right; }
  .font-mono { font-family: monospace; }
  .text-sm { font-size: 13px; }

  /* LOADING & EMPTY */
  .loading-state, .empty-state { text-align: center; padding: 60px 20px; color: #94a3b8; font-style: italic; }
  .empty-icon { font-size: 40px; margin-bottom: 12px; opacity: 0.5; }
  .spinner { width: 32px; height: 32px; border: 3px solid #e2e8f0; border-top-color: #10b981; border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* ANIMATIONS */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s ease-out forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>