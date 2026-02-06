<script>
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';

  let title = $state('');
  let description = $state('');
  let priority = $state('medium');
  let startDate = $state('');
  let deadline = $state('');
  let deadlineTime = $state('');
  let assignTo = $state('all');
  let searchQuery = $state('');
  let results = $state([]);
  let selected = $state([]);
  let loading = $state(false);

  async function searchInterns() {
    if (!searchQuery) {
      results = [];
      return;
    }
    const res = await api.searchInterns(searchQuery);
    results = res.data || [];
  }

  function toggleIntern(intern) {
    if (selected.find((i) => i.id === intern.id)) {
      selected = selected.filter((i) => i.id !== intern.id);
    } else {
      selected = [...selected, intern];
    }
  }

  async function handleSubmit() {
    loading = true;
    try {
      await api.createTask({
        title,
        description,
        priority,
        start_date: startDate,
        deadline,
        deadline_time: deadlineTime,
        assign_to: assignTo,
        intern_ids: selected.map((i) => i.id),
      });
      goto('/tasks');
    } catch (err) {
      alert(err.message || 'Gagal membuat tugas');
    } finally {
      loading = false;
    }
  }
</script>

<div class="card">
  <h3>Buat Tugas Baru</h3>

  <div class="form-group">
    <label class="form-label" for="title">Judul</label>
    <input class="input" bind:value={title} id="title" />
  </div>
  <div class="form-group">
    <label class="form-label" for="description">Deskripsi</label>
    <textarea class="textarea" rows="4" bind:value={description} id="description"></textarea>
  </div>

  <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap:12px;">
    <div class="form-group">
      <label class="form-label" for="priority">Prioritas</label>
      <select class="select" bind:value={priority} id="priority">
        <option value="low">Low</option>
        <option value="medium">Medium</option>
        <option value="high">High</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="startDate">Start Date</label>
      <input class="input" type="date" bind:value={startDate} id="startDate" />
    </div>
    <div class="form-group">
      <label class="form-label" for="deadline">Deadline</label>
      <input class="input" type="date" bind:value={deadline} id="deadline" />
    </div>
    <div class="form-group">
      <label class="form-label" for="deadlineTime">Deadline Time</label>
      <input class="input" type="time" bind:value={deadlineTime} id="deadlineTime" />
    </div>
  </div>

  <div class="form-group">
    <label class="form-label" for="assignTo">Target Penugasan</label>
    <select class="select" bind:value={assignTo} id="assignTo">
      <option value="all">Semua Intern Aktif</option>
      <option value="selected">Pilih Intern</option>
    </select>
  </div>

  {#if assignTo === 'selected'}
    <div class="form-group">
      <label class="form-label" for="searchQuery">Cari Intern</label>
      <input class="input" bind:value={searchQuery} oninput={searchInterns} placeholder="Cari nama/sekolah/jurusan" id="searchQuery" />
    </div>
    <div style="display:grid; gap:8px;">
      {#each results as intern}
        <label style="display:flex; align-items:center; gap:8px; border:1px solid var(--border); border-radius:8px; padding:8px 10px; background:#ffffff; cursor:pointer;">
          <input type="checkbox" checked={selected.find((i) => i.id === intern.id)} onchange={() => toggleIntern(intern)} />
          <span>{intern.label}</span>
        </label>
      {/each}
    </div>
  {/if}

  <div style="margin-top:16px;">
    <button class="btn btn-primary" onclick={handleSubmit} disabled={loading}>{loading ? 'Menyimpan...' : 'Simpan Tugas'}</button>
  </div>
</div>
