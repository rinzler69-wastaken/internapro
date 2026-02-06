<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let reports = $state([]);
  let interns = $state([]);
  let loading = $state(false);

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
      const res = await api.getReports({ page: 1, limit: 20 });
      reports = res.data || [];
      if (auth.user?.role !== 'intern') {
        const internRes = await api.getInterns({ page: 1, limit: 100 });
        interns = internRes.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function handleCreate() {
    try {
      await api.createReport({
        ...form,
        intern_id: Number(form.intern_id),
      });
      await fetchReports();
    } catch (err) {
      alert(err.message || 'Gagal membuat laporan');
    }
  }

  onMount(fetchReports);
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Laporan</h3>
  <p class="text-muted">Buat laporan mingguan, bulanan, dan final.</p>
</div>

{#if auth.user?.role !== 'intern'}
  <div class="card" style="margin-bottom:16px;">
    <h4>Tambah Laporan</h4>
    <div class="form-group">
      <label class="form-label" for="intern_id">Intern</label>
      <select class="select" bind:value={form.intern_id} id="intern_id">
        <option value="">Pilih Intern</option>
        {#each interns as i}
          <option value={i.id}>{i.full_name}</option>
        {/each}
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="title">Judul</label>
      <input class="input" bind:value={form.title} id="title" />
    </div>
    <div class="form-group">
      <label class="form-label" for="content">Konten</label>
      <textarea class="textarea" rows="4" bind:value={form.content} id="content"></textarea>
    </div>
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="type">Tipe</label>
        <select class="select" bind:value={form.type} id="type">
          <option value="weekly">Weekly</option>
          <option value="monthly">Monthly</option>
          <option value="final">Final</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="period_start">Mulai</label>
        <input class="input" type="date" bind:value={form.period_start} id="period_start" />
      </div>
      <div class="form-group">
        <label class="form-label" for="period_end">Selesai</label>
        <input class="input" type="date" bind:value={form.period_end} id="period_end" />
      </div>
    </div>
    <button class="btn btn-primary" onclick={handleCreate}>Simpan Laporan</button>
  </div>
{/if}

<div class="card">
  <h4>Daftar Laporan</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else if reports.length === 0}
    <div class="empty-state">Belum ada laporan.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Judul</th>
          <th>Intern</th>
          <th>Status</th>
          <th>Tipe</th>
        </tr>
      </thead>
      <tbody>
        {#each reports as r}
          <tr>
            <td>{r.title || '-'}</td>
            <td>{r.intern_name || '-'}</td>
            <td><span class={`status status-${r.status || 'pending'}`}>{r.status || '-'}</span></td>
            <td><span class="badge">{r.type || '-'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
