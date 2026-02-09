<script>
  import { onMount } from 'svelte';
  import { goto } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';

  // State
  let assignments = $state([]);
  let loading = $state(true);

  async function fetchAssignments() {
    loading = true;
    try {
      const res = await api.getTaskAssignments({ page: 1, limit: 20 });
      assignments = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  // Helpers UI
  function getPriorityColor(priority) {
    switch(priority?.toLowerCase()) {
        case 'high': return 'bg-rose-100 text-rose-700 border-rose-200';
        case 'medium': return 'bg-amber-100 text-amber-700 border-amber-200';
        case 'low': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
        default: return 'bg-slate-100 text-slate-600 border-slate-200';
    }
  }

  function getProgressColor(percent) {
    if (percent >= 100) return 'bg-emerald-500';
    if (percent >= 50) return 'bg-blue-500';
    if (percent > 0) return 'bg-amber-500';
    return 'bg-slate-300';
  }

  onMount(fetchAssignments);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <div>
            <h2 class="title">Daftar Penugasan</h2>
            <p class="subtitle">Kelola dan pantau progres tugas yang diberikan kepada peserta.</p>
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

    <!-- Main Content -->
    <div class="card animate-slide-up">
        <div class="card-header border-b">
            <h3>Semua Tugas</h3>
            <span class="badge-count">{assignments.length} Tugas</span>
        </div>

        {#if loading}
            <div class="loading-state">
                <div class="spinner"></div>
                <p>Memuat daftar tugas...</p>
            </div>
        {:else if assignments.length === 0}
            <div class="empty-state">
                <div class="empty-icon">📂</div>
                <p>Belum ada penugasan yang dibuat.</p>
            </div>
        {:else}
            <div class="table-container">
                <table class="table">
                    <thead>
                        <tr>
                            <th>Judul Tugas</th>
                            <th>Prioritas</th>
                            <th class="text-center">Jumlah Task</th>
                            <th>Progres Pengerjaan</th>
                            <th class="text-right">Aksi</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each assignments as item}
                            <tr onclick={() => goto(`/task-assignments/${item.id}`)} class="hover-row cursor-pointer">
                                <td>
                                    <div class="task-info">
                                        <span class="task-title">{item.title}</span>
                                        {#if item.description}
                                            <span class="task-desc">{item.description.substring(0, 50)}{item.description.length > 50 ? '...' : ''}</span>
                                        {/if}
                                    </div>
                                </td>
                                <td>
                                    <span class={`priority-badge ${getPriorityColor(item.priority)}`}>
                                        {item.priority || 'Normal'}
                                    </span>
                                </td>
                                <td class="text-center">
                                    <span class="count-pill">{item.tasks_count || 0}</span>
                                </td>
                                <td style="min-width: 160px;">
                                    <div class="progress-wrapper">
                                        <div class="progress-bar-bg">
                                            <div 
                                                class="progress-bar-fill {getProgressColor(item.stats?.progress_percentage || 0)}" 
                                                style="width: {item.stats?.progress_percentage || 0}%"
                                            ></div>
                                        </div>
                                        <span class="progress-text">{item.stats?.progress_percentage || 0}%</span>
                                    </div>
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
  
  .badge-count { background: #f1f5f9; color: #64748b; padding: 4px 10px; border-radius: 20px; font-size: 12px; font-weight: 600; }

  /* --- TABLE --- */
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

  /* Table Content */
  .task-info { display: flex; flex-direction: column; }
  .task-title { font-weight: 600; color: #0f172a; font-size: 15px; }
  .task-desc { font-size: 12px; color: #94a3b8; }

  .priority-badge {
    display: inline-block; padding: 2px 8px; border-radius: 6px;
    font-size: 11px; font-weight: 600; text-transform: uppercase;
    width: fit-content; border: 1px solid transparent;
  }
  .bg-rose-100 { background: #fff1f2; border-color: #fecdd3; }
  .bg-amber-100 { background: #fffbeb; border-color: #fde68a; }
  .bg-emerald-100 { background: #ecfdf5; border-color: #a7f3d0; }
  .bg-slate-100 { background: #f1f5f9; border-color: #e2e8f0; }

  .count-pill {
    background: #f8fafc; border: 1px solid #e2e8f0; color: #475569;
    padding: 2px 8px; border-radius: 6px; font-weight: 600; font-size: 13px;
  }

  /* Progress Bar */
  .progress-wrapper { display: flex; align-items: center; gap: 10px; }
  .progress-bar-bg { flex: 1; height: 6px; background: #f1f5f9; border-radius: 3px; overflow: hidden; }
  .progress-bar-fill { height: 100%; border-radius: 3px; transition: width 0.5s ease; }
  .progress-text { font-size: 12px; font-weight: 600; color: #64748b; width: 32px; text-align: right; }
  
  .bg-emerald-500 { background: #10b981; }
  .bg-blue-500 { background: #3b82f6; }
  .bg-amber-500 { background: #f59e0b; }
  .bg-slate-300 { background: #cbd5e1; }

  .text-center { text-align: center; }
  .text-right { text-align: right; }
  
  .btn-icon {
    background: transparent; border: none; color: #cbd5e1; cursor: pointer;
    padding: 6px; border-radius: 6px; transition: all 0.2s;
  }
  .hover-row:hover .btn-icon { color: #10b981; }

  /* States */
  .empty-state, .loading-state { text-align: center; padding: 60px 20px; color: #94a3b8; font-style: italic; }
  .empty-icon { font-size: 32px; margin-bottom: 12px; opacity: 0.5; }
  .spinner { width: 32px; height: 32px; border: 3px solid #e2e8f0; border-top-color: #10b981; border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Animation */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>