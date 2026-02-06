<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let records = $state([]);
  let todayAttendance = $state(null);
  let loading = $state(false);
  let permissionStatus = $state('sick');
  let permissionNotes = $state('');
  let permissionFile = $state(null);

  const statusLabels = {
    present: 'Hadir',
    late: 'Terlambat',
    absent: 'Tidak Hadir',
    sick: 'Sakit',
    permission: 'Izin',
  };

  function formatDate(value) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  async function fetchAttendance() {
    loading = true;
    try {
      if (auth.user?.role === 'intern') {
        const todayRes = await api.getTodayAttendance();
        todayAttendance = todayRes.data?.attendance || null;

        const internRes = await api.getInterns();
        let internId = internRes.data?.id || internRes.data?.[0]?.id;
        if (internId) {
          const history = await api.getAttendanceByIntern(internId, { page: 1, limit: 20 });
          records = history.data || [];
        }
      } else {
        const res = await api.getAttendance({ page: 1, limit: 20 });
        records = res.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function getPosition() {
    return new Promise((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(resolve, reject, { enableHighAccuracy: true });
    });
  }

  async function handleCheckIn() {
    try {
      const pos = await getPosition();
      await api.checkIn(pos.coords.latitude, pos.coords.longitude);
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal check-in');
    }
  }

  async function handleCheckOut() {
    try {
      const pos = await getPosition();
      await api.checkOut(pos.coords.latitude, pos.coords.longitude);
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal check-out');
    }
  }

  async function handlePermissionSubmit() {
    try {
      await api.submitPermission({
        status: permissionStatus,
        notes: permissionNotes,
        proof_file: permissionFile,
      });
      permissionNotes = '';
      permissionFile = null;
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal mengajukan izin');
    }
  }

  onMount(fetchAttendance);
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Presensi</h3>
  {#if auth.user?.role === 'intern'}
    <div style="display:flex; gap:12px; flex-wrap:wrap; margin-top:12px;">
      <button class="btn btn-primary" onclick={handleCheckIn}>Check-in</button>
      <button class="btn btn-outline" onclick={handleCheckOut}>Check-out</button>
    </div>
    <div style="margin-top:12px;">
      <strong>Status Hari Ini:</strong>
      {#if todayAttendance}
        <span class={`status status-${todayAttendance.status || 'pending'}`}>
          {statusLabels[todayAttendance.status] || todayAttendance.status}
        </span>
      {:else}
        <span class="status status-pending">Belum Presensi</span>
      {/if}
    </div>
  {/if}
</div>

{#if auth.user?.role === 'intern'}
  <div class="card" style="margin-bottom:16px;">
    <h4>Ajukan Izin / Sakit</h4>
    <div class="form-group">
      <label class="form-label" for="permissionStatus">Status</label>
      <select class="select" bind:value={permissionStatus} id="permissionStatus">
        <option value="sick">Sakit</option>
        <option value="permission">Izin</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="permissionNotes">Catatan</label>
      <textarea class="textarea" rows="3" bind:value={permissionNotes} id="permissionNotes"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="permissionFile">Bukti (optional)</label>
      <input
        class="input"
        type="file"
        id="permissionFile"
        onchange={(e) => {
          const target = e.currentTarget;
          if (target instanceof HTMLInputElement) {
            permissionFile = target.files?.[0] || null;
          }
        }}
      />
    </div>
    <button class="btn btn-outline" onclick={handlePermissionSubmit}>Kirim Izin</button>
  </div>
{/if}

<div class="card">
  <h4>Riwayat Presensi</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else if records.length === 0}
    <div class="empty-state">Belum ada data presensi.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Tanggal</th>
          <th>Status</th>
          {#if auth.user?.role !== 'intern'}
            <th>Intern</th>
          {/if}
        </tr>
      </thead>
      <tbody>
        {#each records as r}
          <tr>
            <td>{formatDate(r.date)}</td>
            <td>
              <span class={`status status-${r.status || 'pending'}`}>
                {statusLabels[r.status] || r.status || '-'}
              </span>
            </td>
            {#if auth.user?.role !== 'intern'}
              <td>{r.intern_name}</td>
            {/if}
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
