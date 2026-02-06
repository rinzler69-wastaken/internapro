<script>
  import { api } from '../lib/api.js';

  const { route } = $props();
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

  $effect(() => {
    const params = route?.result?.path?.params || {};
    if (params?.id && params.id !== assignmentId) {
      assignmentId = params.id;
      fetchAssignment();
    }
  });
</script>

{#if loading}
  <div class="card">Memuat...</div>
{:else if !data}
  <div class="card">Data tidak ditemukan.</div>
{:else}
  <div class="card" style="margin-bottom:16px;">
    <h3>{data.assignment.title}</h3>
    <p class="text-muted">{data.assignment.description || 'Tidak ada deskripsi'}</p>
    <div class="stat-grid" style="margin-top:12px;">
      <div class="stat-card"><div class="badge">Total</div><h3>{data.stats.total}</h3></div>
      <div class="stat-card"><div class="badge success">Completed</div><h3>{data.stats.completed}</h3></div>
      <div class="stat-card"><div class="badge warning">In Progress</div><h3>{data.stats.in_progress}</h3></div>
    </div>
  </div>

  <div class="card">
    <h4>Tasks</h4>
    <table class="table">
      <thead>
        <tr>
          <th>Judul</th>
          <th>Intern</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {#each data.tasks as t}
          <tr>
            <td>{t.title}</td>
            <td>{t.intern_name || '-'}</td>
            <td><span class={`status status-${t.status || 'pending'}`}>{t.status || '-'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
