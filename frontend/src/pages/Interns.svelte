<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  let interns = $state([]);
  let loading = $state(false);

  let form = $state({
    email: '',
    password: '',
    full_name: '',
    school: '',
    department: '',
    start_date: '',
    end_date: '',
  });

  async function fetchInterns() {
    loading = true;
    try {
      const res = await api.getInterns({ page: 1, limit: 50 });
      interns = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function createIntern() {
    try {
      await api.createIntern(form);
      await fetchInterns();
    } catch (err) {
      alert(err.message || 'Gagal membuat intern');
    }
  }

  onMount(fetchInterns);
</script>

<!-- <div class="card" style="margin-bottom:16px;">
  <h3>Anggota Magang</h3>
  <p class="text-muted">Kelola data peserta magang.</p>
</div> -->

<div class="card" style="margin-bottom:16px;">
  <h4>Tambah Intern</h4>
  <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
    <div class="form-group">
      <label class="form-label" for="full_name">Nama Lengkap</label>
      <input class="input" bind:value={form.full_name} id="full_name" />
    </div>
    <div class="form-group">
      <label class="form-label" for="email">Email</label>
      <input class="input" type="email" bind:value={form.email} id="email" />
    </div>
    <div class="form-group">
      <label class="form-label" for="password">Password</label>
      <input class="input" type="password" bind:value={form.password} id="password" />
    </div>
    <div class="form-group">
      <label class="form-label" for="school">Sekolah</label>
      <input class="input" bind:value={form.school} id="school" />
    </div>
    <div class="form-group">
      <label class="form-label" for="department">Jurusan</label>
      <input class="input" bind:value={form.department} id="department" />
    </div>
    <div class="form-group">
      <label class="form-label" for="start_date">Mulai</label>
      <input class="input" type="date" bind:value={form.start_date} id="start_date" />
    </div>
    <div class="form-group">
      <label class="form-label" for="end_date">Selesai</label>
      <input class="input" type="date" bind:value={form.end_date} id="end_date" />
    </div>
  </div>
  <button class="btn btn-primary" style="margin-top:12px;" onclick={createIntern}>Simpan</button>
</div>

<div class="card">
  <h4>Daftar Intern</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else if interns.length === 0}
    <div class="empty-state">Belum ada intern.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Nama</th>
          <th>Email</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {#each interns as i}
          <tr>
            <td>{i.full_name || '-'}</td>
            <td>{i.email || '-'}</td>
            <td><span class={`status status-${i.status || 'inactive'}`}>{i.status || '-'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
