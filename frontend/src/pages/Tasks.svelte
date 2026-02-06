<script>
  import { onMount } from 'svelte';
  import { goto, route } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let tasks = $state([]);
  let pagination = $state({});
  let search = $state('');
  let status = $state('');
  let priority = $state('');
  let loading = $state(false);

  const statusLabels = {
    pending: 'Pending',
    scheduled: 'Scheduled',
    in_progress: 'In Progress',
    submitted: 'Submitted',
    revision: 'Revision',
    completed: 'Completed',
  };

  function formatDate(value) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  async function fetchTasks() {
    loading = true;
    try {
      const res = await api.getTasks({ search, status, priority, page: 1, limit: 20 });
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

<div class="card" style="margin-bottom:16px;">
  <div style="display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap;">
    <div>
      <h3>Daftar Penugasan</h3>
      <p class="text-muted">Kelola tugas peserta magang dan progres pengerjaan.</p>
    </div>
    {#if auth.user?.role !== 'intern'}
      <a href="/tasks/create" use:route class="btn btn-primary">Buat Tugas</a>
    {/if}
  </div>

  <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap:12px; margin-top:16px;">
    <input class="input" placeholder="Cari judul..." bind:value={search} />
    <select class="select" bind:value={status}>
      <option value="">Semua Status</option>
      <option value="pending">Pending</option>
      <option value="in_progress">In Progress</option>
      <option value="submitted">Submitted</option>
      <option value="revision">Revision</option>
      <option value="completed">Completed</option>
      <option value="scheduled">Scheduled</option>
    </select>
    <select class="select" bind:value={priority}>
      <option value="">Semua Prioritas</option>
      <option value="low">Low</option>
      <option value="medium">Medium</option>
      <option value="high">High</option>
    </select>
    <button class="btn btn-outline" onclick={fetchTasks}>Filter</button>
  </div>
</div>

<div class="card">
  {#if loading}
    <div>Memuat...</div>
  {:else if tasks.length === 0}
    <div class="empty-state">Tidak ada tugas ditemukan.</div>
  {:else}
  <table class="table">
      <thead>
        <tr>
          <th>Judul</th>
          <th>Intern</th>
          <th>Status</th>
          <th>Prioritas</th>
          <th>Deadline</th>
        </tr>
      </thead>
      <tbody>
        {#each tasks as task}
          <tr
            role="link"
            tabindex="0"
            onclick={() => goto(`/tasks/${task.id}`)}
            onkeydown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                goto(`/tasks/${task.id}`);
              }
            }}
            style="cursor:pointer;"
          >
            <td>{task.title}</td>
            <td>{task.intern_name || '-'}</td>
            <td>
              <span class={`status status-${task.status || 'pending'}`}>
                {statusLabels[task.status] || task.status || '-'}
              </span>
            </td>
            <td><span class="badge">{task.priority || '-'}</span></td>
            <td>{formatDate(task.deadline)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
