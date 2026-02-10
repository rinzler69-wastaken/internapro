<script>
  import { onMount } from 'svelte';
  import { goto, route } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let tasks = $state([]);
  let pagination = $state({ page: 1, total_pages: 1 });
  let search = $state('');
  let status = $state('');
  let priority = $state('');
  let internId = $state('');
  let interns = $state([]);
  let loading = $state(true);
  let exporting = $state(false);

  const statusLabels = {
    pending: 'Pending',
    scheduled: 'Terjadwal',
    in_progress: 'Dalam Proses',
    submitted: 'Menunggu Review',
    revision: 'Revisi',
    completed: 'Selesai',
  };

  function formatDate(value) {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  function formatSubmitted(task) {
    // Prefer submitted_at; fall back to completed_at when submissions are auto-marked complete
    return formatDate(task?.submitted_at || task?.completed_at);
  }

  function getPriorityColor(p) {
    switch (p) {
      case 'high': return 'tone-rose';    // red
      case 'medium': return 'tone-amber'; // yellow
      default: return 'tone-emerald';     // green
    }
  }
  function getStatusColor(s) {
    switch (s) {
      case 'completed': return 'tone-emerald';
      case 'submitted': return 'tone-blue';
      case 'in_progress': return 'tone-amber';
      case 'revision': return 'tone-rose';
      default: return 'tone-slate';
    }
  }

  async function fetchTasks() {
    loading = true;
    try {
      const params = { page: pagination.page || 1, limit: 15 };
      if (search) params.search = search;
      if (status) params.status = status;
      if (priority) params.priority = priority;
      if (internId) params.intern_id = internId;

      const res = await api.getTasks(params);
      tasks = res.data || [];
      pagination = res.pagination || { page: 1, total_pages: 1 };
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function fetchInterns() {
    if (auth.user?.role === 'intern') return;
    try {
      const res = await api.getInterns({ status: 'active', limit: 100 });
      interns = res.data || [];
    } catch (err) {
      console.error(err);
    }
  }

  async function exportTasks() {
    exporting = true;
    try {
      const params = new URLSearchParams();
      if (status) params.append('status', status);
      if (internId) params.append('intern_id', internId);
      const res = await fetch(`/api/export/tasks${params.toString() ? `?${params.toString()}` : ''}`, {
        headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
      });
      if (!res.ok) throw new Error('Gagal mengekspor tugas');
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `Data_Tugas_${new Date().toISOString().slice(0,10)}.xlsx`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } finally {
      exporting = false;
    }
  }

  function setPage(p) {
    if (!pagination.total_pages) return;
    const target = Math.min(Math.max(1, p), pagination.total_pages);
    if (target !== pagination.page) {
      pagination = { ...pagination, page: target };
      fetchTasks();
    }
  }

  onMount(() => {
    fetchTasks();
    fetchInterns();
  });
</script>

<div class="page-shell">
  <div class="page-header">
    <div class="page-title">
      <h1>Daftar Penugasan</h1>
      <p class="muted">Pantau, review, dan kelola tugas magang dengan ringkas.</p>
    </div>
    {#if auth.user?.role !== 'intern'}
      <div class="page-actions">
        <button class="ghost" onclick={exportTasks} disabled={exporting}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2"></path><polyline points="7 11 12 16 17 11"></polyline><line x1="12" y1="4" x2="12" y2="16"></line></svg>
          {exporting ? 'Menyiapkan...' : 'Export'}
        </button>
        <a href="/tasks/create" use:route class="primary">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
          Buat Tugas
        </a>
      </div>
    {/if}
  </div>

  <section class="panel">
    <div class="filters">
      <div class="field stretch">
        <label>Cari</label>
        <div class="input-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input placeholder="Judul tugas..." bind:value={search} onkeydown={(e) => e.key === 'Enter' && fetchTasks()} />
        </div>
      </div>
      <div class="field">
        <label>Status</label>
        <select bind:value={status}>
          <option value="">Semua</option>
          <option value="pending">Pending</option>
          <option value="in_progress">Dalam Proses</option>
          <option value="submitted">Menunggu Review</option>
          <option value="revision">Revisi</option>
          <option value="completed">Selesai</option>
        </select>
      </div>
      <div class="field">
        <label>Prioritas</label>
        <select bind:value={priority}>
          <option value="">Semua</option>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
      </div>
      {#if auth.user?.role !== 'intern'}
        <div class="field">
          <label>Intern</label>
          <select bind:value={internId}>
            <option value="">Semua</option>
            {#each interns as intern}
              <option value={intern.id}>{intern.full_name || intern.name}</option>
            {/each}
          </select>
        </div>
      {/if}
      <div class="field actions">
        <button class="secondary" onclick={fetchTasks}>Terapkan</button>
        <button class="ghost" onclick={() => { search=''; status=''; priority=''; internId=''; fetchTasks(); }}>Reset</button>
      </div>
    </div>
  </section>

  <section class="panel table-panel">
    {#if loading}
      <div class="placeholder">
        <div class="spinner"></div>
        <p>Memuat daftar tugas...</p>
      </div>
    {:else if tasks.length === 0}
      <div class="placeholder">
        <div class="empty">📋</div>
        <p>Tidak ada tugas ditemukan.</p>
        {#if auth.user?.role !== 'intern'}
          <p class="muted">Coba ubah filter atau buat tugas baru.</p>
        {/if}
      </div>
    {:else}
      <div class="table-wrap desktop-only">
        <table>
          <thead>
            <tr>
              <th>Judul</th>
              {#if auth.user?.role !== 'intern'}<th>Intern</th>{/if}
              <th>Status</th>
              <th>Prioritas</th>
              <th>Deadline</th>
              <th>Diumpulkan</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each tasks as task}
              <tr class="row" onclick={() => goto(`/tasks/${task.id}`)}>
                <td>
                  <div class="title-cell">
                    <span class="title">{task.title}</span>
                    <span class="sub">{task.description || '—'}</span>
                  </div>
                </td>
                {#if auth.user?.role !== 'intern'}
                  <td>
                    <div class="user-chip">
                      <div class="avatar">{task.intern_name?.charAt(0) || 'U'}</div>
                      <div class="user-text">
                        <span class="name">{task.intern_name || '-'}</span>
                      </div>
                    </div>
                  </td>
                {/if}
                <td>
                  <div class="stack">
                    <span class={`pill status ${getStatusColor(task.status)}`}>
                      <span class="dot"></span>
                      {statusLabels[task.status] || task.status || '-'}
                    </span>
                    {#if task.is_late && task.status !== 'completed'}<span class="chip danger">Lewat</span>{/if}
                   
                  </div>
                </td>
                <td><span class={`pill priority ${getPriorityColor(task.priority)}`}><span class="dot"></span>{task.priority === 'medium' ? 'mid' : (task.priority || '-')}</span></td>
                <td class={`mono ${task.is_late && task.deadline ? 'late' : ''}`}>{formatDate(task.deadline)}</td>
                <td class="mono">{formatSubmitted(task)}</td>
                <td class="text-right">
                  <div class="flex justify-end gap-2">
                    {#if auth.user?.role !== 'intern'}
                      <button class="icon-btn rounded-full" onclick={(e) => { e.stopPropagation(); goto(`/tasks/edit/${task.id}`); }} title="Edit Tugas">
                        <span class="material-symbols-outlined text-[18px]">edit</span>
                      </button>
                    {/if}
                    <button class="icon-btn" aria-label="Lihat">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="mobile-list">
        {#each tasks as task}
          <div class="card-row" onclick={() => goto(`/tasks/${task.id}`)}>
            <div class="row-top">
              <div class="title-block">
                <div class="title">{task.title}</div>
                <div class="sub">{task.description || '—'}</div>
              </div>
              <div class={`pill priority small ${getPriorityColor(task.priority)}`}>
                <span class="dot"></span>
                {task.priority === 'medium' ? 'mid' : (task.priority || '-')}
              </div>
            </div>
            <div class="row-mid">
              <div class="stack">
                <span class={`pill status ${getStatusColor(task.status)}`}><span class="dot"></span>{statusLabels[task.status] || task.status || '-'}</span>
                {#if task.is_late && task.status !== 'completed'}<span class="chip danger">Lewat</span>{/if}
              </div>
              <div class={`mono ${task.is_late && task.deadline ? 'late' : ''}`}>{formatDate(task.deadline)}</div>
            </div>
            <div class="row-mid submitted-row">
              <span class="muted text-xs">Dikumpulkan</span>
              <span class="mono">{formatSubmitted(task)}</span>
            </div>
            {#if auth.user?.role !== 'intern'}
              <div class="row-bottom">
                <div class="user-chip">
                  <div class="avatar">{task.intern_name?.charAt(0) || 'U'}</div>
                  <div class="user-text">
                    <span class="name">{task.intern_name || '-'}</span>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </section>

  {#if pagination.total_pages > 1}
    <div class="pager">
      <button class="ghost" onclick={() => setPage((pagination.page || 1) - 1)} disabled={(pagination.page || 1) <= 1}>‹ Prev</button>
      <span class="muted">Halaman {pagination.page || 1} dari {pagination.total_pages}</span>
      <button class="ghost" onclick={() => setPage((pagination.page || 1) + 1)} disabled={(pagination.page || 1) >= pagination.total_pages}>Next ›</button>
    </div>
  {/if}
</div>

<style>
  :global(body) {
    font-family: 'Geist', 'Inter', sans-serif;
    background: #f7f8fb;
    color: #0f172a;
  }
  .page-shell {
    max-width: 1400px;
    margin: 0 auto;
    padding: 32px 20px 64px;
  }
  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 20px;
    flex-wrap: wrap;
  }
  .page-title h1 { margin: 4px 0; font-size: 20px; font-weight: 800; letter-spacing: -0.02em; }
  .eyebrow { text-transform: uppercase; font-weight: 700; font-size: 12px; color: #6366f1; letter-spacing: 0.08em; margin: 0; }
  .muted { color: #94a3b8; margin: 0; }
  .page-actions { display: flex; gap: 10px; align-items: center; }

  .panel {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 16px;
    padding: 18px;
    box-shadow: 0 12px 30px -18px rgba(15,23,42,0.4);
    margin-bottom: 16px;
  }

  .filters {
    display: grid;
    grid-template-columns: 2fr repeat(3, 1fr) auto;
    gap: 12px;
  }
  .field.stretch { flex: 1; }
  .field { display: flex; flex-direction: column; gap: 6px; }
  .field.actions { justify-content: flex-end; flex-direction: row; align-items: flex-end; }
  .field label { font-weight: 600; color: #475569; font-size: 12px; text-transform: uppercase; letter-spacing: 0.04em; }
  .input-icon { position: relative; display: flex; align-items: center; width: 100%; }
  .input-icon svg { position: absolute; left: 10px; color: #94a3b8; }
  .input-icon input { padding-left: 34px; }
  input, select {
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    padding: 10px 12px;
    font-size: 14px;
    background: #fff;
    width: 100%;
    box-sizing: border-box;
    transition: border-color .15s, box-shadow .15s;
  }
  input:focus, select:focus { outline: none; border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.12); }

  .page-actions .primary, .primary, .secondary, .ghost {
    display: inline-flex; align-items: center; gap: 8px;
    border-radius: 999px; font-weight: 700; cursor: pointer; text-decoration: none;
    border: 1px solid transparent; padding: 10px 14px; transition: all .15s;
  }
  .primary { background: linear-gradient(120deg, #22c55e, #22c55e); color: white; box-shadow: 0 12px 25px -12px rgba(79,70,229,0.35); }
  .primary:hover { transform: translateY(-1px); }
  .secondary { background: #0f172a; color: white; border-color: #0f172a; }
  .ghost { background: #f8fafc; color: #0f172a; border-color: #e2e8f0; }
  .ghost:disabled { opacity: 0.6; cursor: not-allowed; }

  .table-panel { padding: 0; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; }
  thead th {
    font-size: 12px; text-transform: uppercase; letter-spacing: 0.04em;
    color: #94a3b8; text-align: left; padding: 14px 18px; border-bottom: 1px solid #e2e8f0;
  }
  tbody td { padding: 14px 18px; border-bottom: 1px solid #f1f5f9; vertical-align: middle; }
  tr.row { transition: background .12s, transform .12s; }
  tr.row:hover { background: #f8fafc; transform: translateY(-1px); cursor: pointer; }

  .title-cell { display: flex; flex-direction: column; gap: 4px; }
  .title-block { flex: 1; min-width: 0; }
  .title { font-weight: 700; color: #0f172a; }
  .sub { color: #94a3b8; font-size: 13px; max-width: 100%; display: block; line-height: 1.4; }

  .user-chip { display: inline-flex; align-items: center; gap: 8px; }
  .avatar { width: 32px; height: 32px; border-radius: 10px; background: linear-gradient(135deg,#22c55e,#10b981); color: white; display: grid; place-items: center; font-weight: 700; }
  .name { font-weight: 600; color: #0f172a; }

  .stack { display: flex; flex-direction: column; gap: 6px; align-items: flex-start; }
  .pill {
    display: inline-flex; align-items: center; gap: 6px;
    min-width: 120px;
    justify-content: center;
    padding: 6px 10px; border-radius: 14px; font-weight: 700; font-size: 12px;
    border: 1px solid transparent;
    letter-spacing: 0.01em;
  }
  .pill.status { box-shadow: inset 0 1px 0 rgba(255,255,255,0.6); }
  .pill.priority { background: #fff; border-style: dashed; }
  .pill.subtle { background: #f8fafc; color: #0f172a; border-color: #e2e8f0; }
  .tone-rose { background: #fff1f2; color: #be123c; border-color: #fecdd3; }
  .tone-amber { background: #fffbeb; color: #b45309; border-color: #fde68a; }
  .tone-emerald { background: #ecfdf3; color: #047857; border-color: #a7f3d0; }
  .tone-blue { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; }
  .tone-slate { background: #f1f5f9; color: #475569; border-color: #e2e8f0; }
  .dot {
    width: 8px; height: 8px; border-radius: 50%; background: currentColor; opacity: 0.9;
    box-shadow: 0 0 0 3px currentColor12;
  }

  .chip { padding: 3px 8px; border-radius: 8px; font-size: 11px; font-weight: 700; text-transform: uppercase; }
  .chip.danger { background: #fef2f2; color: #b91c1c; }
  .chip.warn { background: #fff7ed; color: #b45309; }

  .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; color: #475569; }
  .mono.late { color: #b91c1c; font-weight: 700; }

  .icon-btn {
    border: 1px solid #e2e8f0; background: #fff; border-radius: 10px;
    padding: 6px; color: #94a3b8; cursor: pointer; transition: all .12s;
  }
  .icon-btn:hover { color: #0f172a; border-color: #cbd5e1; background: #f8fafc; }

  .placeholder { padding: 36px; text-align: center; color: #94a3b8; }
  .empty { font-size: 36px; margin-bottom: 8px; }
  .spinner { width: 32px; height: 32px; border: 3px solid #e2e8f0; border-top-color: #6366f1; border-radius: 50%; margin: 0 auto 14px; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .pager {
    margin-top: 12px;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 12px;
  }
  .desktop-only { display: block; }
  .mobile-list { display: none; }

  .card-row {
    border: 1px solid #e2e8f0;
    border-radius: 14px;
    padding: 14px;
    background: #fff;
    display: flex;
    flex-direction: column;
    gap: 8px;
    box-shadow: 0 10px 22px -18px rgba(15,23,42,0.4);
  }
  .row-top, .row-mid, .row-bottom { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
  .row-top { align-items: flex-start; }
  .row-mid { flex-wrap: wrap; }
  .row-bottom { justify-content: flex-start; }
  .pill.small { padding: 4px 8px; font-size: 11px; }
  .submitted-row { justify-content: space-between; gap: 8px; }

  /* Mobile / tablet responsiveness */
  @media (max-width: 1024px) {
    .filters {
      grid-template-columns: 1fr 1fr;
    }
    .field.stretch { grid-column: span 2; }
    .field.actions { grid-column: span 2; }
    .page-actions {
      width: 100%;
      justify-content: flex-start;
    }
    .page-actions .primary,
    .page-actions .ghost {
      width: 48%;
      justify-content: center;
    }
  }

  @media (max-width: 640px) {
    .page-shell { padding: 24px 16px 48px; }
    .page-title h1 { font-size: 24px; }
    .filters {
      grid-template-columns: 1fr;
    }
    .field.stretch { grid-column: span 1; }
    .field.actions { grid-column: span 1; flex-direction: row; align-items: center; gap: 8px; }
    .field.actions button { flex: 1; justify-content: center; }
    .title, .sub { padding-bottom: 4px; }
    .page-actions { flex-direction: column; gap: 8px; align-items: stretch; width: 100%; }
    .page-actions .primary,
    .page-actions .ghost { width: 100%; text-align: center; justify-content: center; }
    thead { display: none; }
    .desktop-only { display: none; }
    .mobile-list { display: flex; flex-direction: column; gap: 12px; }
  }
</style>
