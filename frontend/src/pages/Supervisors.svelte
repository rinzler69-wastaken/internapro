<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let supervisors = $state([]);
  let loading = $state(false);
  let search = $state('');
  let status = $state('');
  let editingId = $state(null);

  let form = $state({
    name: '',
    email: '',
    password: '',
    nip: '',
    phone: '',
    position: '',
    address: '',
    institution: '',
    status: 'active',
  });

  async function fetchSupervisors() {
    loading = true;
    try {
      const res = await api.getSupervisors({ page: 1, limit: 50, search, status });
      supervisors = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function createSupervisor() {
    if (!form.name || !form.email || (!editingId && !form.password)) return;
    try {
      if (editingId) {
        const payload = { ...form };
        if (!payload.password) delete payload.password;
        await api.updateSupervisor(editingId, payload);
      } else {
        await api.createSupervisor(form);
      }
      form = { name: '', email: '', password: '', nip: '', phone: '', position: '', address: '', institution: '', status: 'active' };
      editingId = null;
      await fetchSupervisors();
    } catch (err) {
      alert(err.message || 'Gagal membuat pembimbing');
    }
  }

  function startEdit(s) {
    editingId = s.id;
    form = {
      name: s.full_name || s.name || '',
      email: s.email || '',
      password: '',
      nip: s.nip || '',
      phone: s.phone || '',
      position: s.position || '',
      address: s.address || '',
      institution: s.institution || '',
      status: s.status || 'active',
    };
  }

  function cancelEdit() {
    editingId = null;
    form = { name: '', email: '', password: '', nip: '', phone: '', position: '', address: '', institution: '', status: 'active' };
  }

  async function approveSupervisor(id) {
    await api.approveSupervisor(id);
    await fetchSupervisors();
  }

  async function rejectSupervisor(id) {
    if (!confirm('Tolak dan hapus pembimbing ini?')) return;
    await api.rejectSupervisor(id);
    await fetchSupervisors();
  }

  async function deleteSupervisor(id) {
    if (!confirm('Hapus pembimbing ini?')) return;
    await api.deleteSupervisor(id);
    await fetchSupervisors();
  }

  onMount(fetchSupervisors);
</script>

<div class="card" style="margin-bottom:16px;">
  <div style="display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap;">
    <div>
      <h3>Pembimbing</h3>
      <p class="text-muted">Kelola data pembimbing dan status persetujuan.</p>
    </div>
  </div>

  <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap:12px; margin-top:12px;">
    <input class="input" placeholder="Cari nama/email/institusi" bind:value={search} />
    <select class="select" bind:value={status}>
      <option value="">Semua Status</option>
      <option value="pending">Pending</option>
      <option value="active">Active</option>
    </select>
    <button class="btn btn-outline" onclick={fetchSupervisors}>Filter</button>
  </div>
</div>

{#if auth.user?.role === 'admin'}
  <div class="card" style="margin-bottom:16px;">
    <h4>Tambah Pembimbing</h4>
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="sup-name">Nama</label>
        <input class="input" id="sup-name" bind:value={form.name} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-email">Email</label>
        <input class="input" id="sup-email" type="email" bind:value={form.email} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-password">Password</label>
        <input class="input" id="sup-password" type="password" bind:value={form.password} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-nip">NIP</label>
        <input class="input" id="sup-nip" bind:value={form.nip} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-phone">Telepon</label>
        <input class="input" id="sup-phone" bind:value={form.phone} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-position">Posisi</label>
        <input class="input" id="sup-position" bind:value={form.position} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-address">Alamat</label>
        <input class="input" id="sup-address" bind:value={form.address} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-institution">Institusi</label>
        <input class="input" id="sup-institution" bind:value={form.institution} />
      </div>
      <div class="form-group">
        <label class="form-label" for="sup-status">Status</label>
        <select class="select" id="sup-status" bind:value={form.status}>
          <option value="active">Active</option>
          <option value="pending">Pending</option>
        </select>
      </div>
    </div>
    <div style="display:flex; gap:8px; margin-top:12px;">
      <button class="btn btn-primary" onclick={createSupervisor}>
        {editingId ? 'Update' : 'Simpan'}
      </button>
      {#if editingId}
        <button class="btn btn-outline" onclick={cancelEdit}>Batal</button>
      {/if}
    </div>
  </div>
{/if}

<div class="card">
  <h4>Daftar Pembimbing</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else if supervisors.length === 0}
    <div class="empty-state">Belum ada pembimbing.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Nama</th>
          <th>Email</th>
          <th>Institusi</th>
          <th>Status</th>
          <th>Intern</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each supervisors as s}
          <tr>
            <td>{s.full_name || s.name || '-'}</td>
            <td>{s.email || '-'}</td>
            <td>{s.institution || '-'}</td>
            <td><span class={`status status-${s.status || 'inactive'}`}>{s.status || '-'}</span></td>
            <td>{s.interns_count || 0}</td>
            <td style="white-space:nowrap;">
              {#if auth.user?.role === 'admin'}
                {#if s.status === 'pending'}
                  <button class="btn btn-outline" onclick={() => approveSupervisor(s.id)}>Approve</button>
                  <button class="btn btn-ghost" onclick={() => rejectSupervisor(s.id)}>Reject</button>
                {:else}
                  <button class="btn btn-outline" onclick={() => startEdit(s)}>Edit</button>
                  <button class="btn btn-ghost" onclick={() => deleteSupervisor(s.id)}>Hapus</button>
                {/if}
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
