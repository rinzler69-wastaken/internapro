<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let interns = $state([]);
  let supervisors = $state([]);
  let loading = $state(false);

  let internStatus = $state('');
  let attendanceIntern = $state('');
  let attendanceStatus = $state('');
  let attendanceStart = $state('');
  let attendanceEnd = $state('');
  let taskIntern = $state('');
  let taskStatus = $state('');

  let importFile = $state(null);
  let importSupervisor = $state('');
  let importResult = $state(null);

  function downloadBlob(blob, filename) {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  }

  async function fetchData() {
    loading = true;
    try {
      const internRes = await api.getInterns({ page: 1, limit: 200 });
      interns = internRes.data || [];
      if (auth.user?.role === 'admin') {
        const supRes = await api.getSupervisors({ page: 1, limit: 200 });
        supervisors = supRes.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function exportInterns() {
    const { blob, filename } = await api.exportInterns({ status: internStatus });
    downloadBlob(blob, filename || 'interns.xlsx');
  }

  async function exportAttendance() {
    const params = {};
    if (attendanceIntern) params.intern_id = attendanceIntern;
    if (attendanceStatus) params.status = attendanceStatus;
    if (attendanceStart) params.start_date = attendanceStart;
    if (attendanceEnd) params.end_date = attendanceEnd;
    const { blob, filename } = await api.exportAttendances(params);
    downloadBlob(blob, filename || 'attendance.xlsx');
  }

  async function exportTasks() {
    const params = {};
    if (taskIntern) params.intern_id = taskIntern;
    if (taskStatus) params.status = taskStatus;
    const { blob, filename } = await api.exportTasks(params);
    downloadBlob(blob, filename || 'tasks.xlsx');
  }

  async function downloadTemplate() {
    const { blob, filename } = await api.downloadImportTemplate();
    downloadBlob(blob, filename || 'template.xlsx');
  }

  async function importInterns() {
    if (!importFile) return;
    const result = await api.importInterns(importFile, importSupervisor || null);
    importResult = result?.data || result;
    await fetchData();
  }

  onMount(fetchData);
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Export / Import</h3>
  <p class="text-muted">Unduh data atau impor peserta magang secara massal.</p>
</div>

<div class="card" style="margin-bottom:16px;">
  <h4>Export Data</h4>
  <div style="display:grid; gap:12px;">
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="export-intern-status">Status Intern</label>
        <select class="select" id="export-intern-status" bind:value={internStatus}>
          <option value="">Semua</option>
          <option value="active">Aktif</option>
          <option value="pending">Pending</option>
          <option value="completed">Selesai</option>
          <option value="cancelled">Dibatalkan</option>
        </select>
      </div>
    </div>
    <button class="btn btn-outline" onclick={exportInterns}>Export Interns</button>

    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="export-attendance-intern">Intern</label>
        <select class="select" id="export-attendance-intern" bind:value={attendanceIntern}>
          <option value="">Semua</option>
          {#each interns as i}
            <option value={i.id}>{i.full_name || i.name || i.email}</option>
          {/each}
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="export-attendance-status">Status Presensi</label>
        <select class="select" id="export-attendance-status" bind:value={attendanceStatus}>
          <option value="">Semua</option>
          <option value="present">Hadir</option>
          <option value="late">Terlambat</option>
          <option value="absent">Tidak Hadir</option>
          <option value="sick">Sakit</option>
          <option value="permission">Izin</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="export-attendance-start">Mulai</label>
        <input class="input" id="export-attendance-start" type="date" bind:value={attendanceStart} />
      </div>
      <div class="form-group">
        <label class="form-label" for="export-attendance-end">Selesai</label>
        <input class="input" id="export-attendance-end" type="date" bind:value={attendanceEnd} />
      </div>
    </div>
    <button class="btn btn-outline" onclick={exportAttendance}>Export Presensi</button>

    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="export-task-intern">Intern</label>
        <select class="select" id="export-task-intern" bind:value={taskIntern}>
          <option value="">Semua</option>
          {#each interns as i}
            <option value={i.id}>{i.full_name || i.name || i.email}</option>
          {/each}
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="export-task-status">Status Tugas</label>
        <select class="select" id="export-task-status" bind:value={taskStatus}>
          <option value="">Semua</option>
          <option value="pending">Pending</option>
          <option value="in_progress">In Progress</option>
          <option value="submitted">Submitted</option>
          <option value="revision">Revision</option>
          <option value="completed">Completed</option>
        </select>
      </div>
    </div>
    <button class="btn btn-outline" onclick={exportTasks}>Export Tugas</button>
  </div>
</div>

<div class="card">
  <h4>Import Intern</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else}
    <div class="form-group">
      <div class="form-label" id="import-template-label">Template</div>
      <button class="btn btn-outline" aria-labelledby="import-template-label" onclick={downloadTemplate}>
        Download Template
      </button>
    </div>
    <div class="form-group">
      <label class="form-label" for="import-file">File Import</label>
      <input
        class="input"
        id="import-file"
        type="file"
        accept=".xlsx,.xls,.csv"
        onchange={(e) => {
          const target = e.currentTarget;
          if (target instanceof HTMLInputElement) {
            importFile = target.files?.[0] || null;
          }
        }}
      />
    </div>
    {#if auth.user?.role === 'admin'}
      <div class="form-group">
        <label class="form-label" for="import-supervisor">Assign Pembimbing (opsional)</label>
        <select class="select" id="import-supervisor" bind:value={importSupervisor}>
          <option value="">Tidak ditentukan</option>
          {#each supervisors as s}
            <option value={s.user_id}>{s.full_name || s.name || s.email}</option>
          {/each}
        </select>
      </div>
    {/if}
    <button class="btn btn-primary" onclick={importInterns}>Import</button>
    {#if importResult}
      <div style="margin-top:12px;" class="text-muted">
        Imported: {importResult.imported || 0}, Skipped: {importResult.skipped || 0}
      </div>
    {/if}
  {/if}
</div>
