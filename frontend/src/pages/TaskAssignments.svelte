<script>
  import { onMount } from 'svelte';
  import { goto } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';

  let assignments = $state([]);
  let loading = $state(false);

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

  onMount(fetchAssignments);
</script>

<div class="card">
  <div style="display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap;">
    <div>
      <h3>Daftar Tugas</h3>
      <p class="text-muted">Penugasan yang dikelompokkan berdasarkan assignment.</p>
    </div>
  </div>
  {#if loading}
    <div>Memuat...</div>
  {:else if assignments.length === 0}
    <div class="empty-state">Belum ada penugasan.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Judul</th>
          <th>Prioritas</th>
          <th>Jumlah Task</th>
          <th>Progress</th>
        </tr>
      </thead>
      <tbody>
        {#each assignments as item}
          <tr onclick={() => goto(`/task-assignments/${item.id}`)} style="cursor:pointer;">
            <td>{item.title}</td>
            <td><span class="badge">{item.priority || '-'}</span></td>
            <td>{item.tasks_count}</td>
            <td>{item.stats?.progress_percentage || 0}%</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
