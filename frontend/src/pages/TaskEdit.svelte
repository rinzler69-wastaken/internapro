<script>
  import { onMount } from 'svelte';
  import { goto } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  const { route } = $props();
  let taskId = $state('');
  let loading = $state(true);
  let saving = $state(false);
  let interns = $state([]);

  let form = $state({
    title: '',
    description: '',
    deadline: '',
    deadline_time: '23:59',
    start_date: '',
    priority: 'medium',
    intern_id: '',
    submission_method: 'links'
  });

  async function fetchTask() {
    loading = true;
    try {
      const res = await api.getTask(taskId);
      const data = res.data;
      form = {
        title: data.title || '',
        description: data.description || '',
        deadline: data.deadline ? data.deadline.split('T')[0] : '',
        deadline_time: data.deadline_time ? data.deadline_time.slice(0, 5) : '23:59',
        start_date: data.start_date ? data.start_date.split('T')[0] : '',
        priority: data.priority || 'medium',
        intern_id: data.intern_id || '',
        submission_method: data.submission_method || 'links'
      };
    } catch (err) {
      console.error(err);
      alert('Gagal memuat data tugas');
    } finally {
      loading = false;
    }
  }

  async function fetchInterns() {
    if (auth.user?.role === 'intern') return;
    try {
      const res = await api.getInterns({ status: 'active', limit: 200 });
      interns = res.data || [];
    } catch (err) {
      console.error(err);
    }
  }

  async function handleSave() {
    saving = true;
    try {
      await api.updateTask(taskId, form);
      alert('Tugas berhasil diperbarui');
      goto(`/tasks/${taskId}`);
    } catch (err) {
      alert(err.message || 'Gagal memperbarui tugas');
    } finally {
      saving = false;
    }
  }

  $effect(() => {
    const params = route?.result?.path?.params || {};
    if (params?.id && params.id !== taskId) {
      taskId = params.id;
      fetchTask();
    }
  });

  onMount(fetchInterns);
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet" />
  <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0" />
</svelte:head>

<div class="page-bg">
  <div class="container animate-fade-in">
    <div class="header">
      <button class="btn-back" onclick={() => goto('/tasks')}>
        <span class="material-symbols-outlined">arrow_back</span>
        Kembali
      </button>
      <div class="mt-4">
        <h2 class="title">Edit Tugas</h2>
        <p class="subtitle">Perbarui informasi dan instruksi penugasan.</p>
      </div>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else}
      <div class="card animate-slide-up">
        <div class="card-body">
          <div class="form-grid">
            <div class="form-group full-width">
              <label class="label" for="title">Judul Tugas</label>
              <input id="title" class="input-field" bind:value={form.title} placeholder="Masukkan judul tugas..." />
            </div>

            <div class="form-group full-width">
              <label class="label" for="description">Deskripsi & Instruksi</label>
              <textarea id="description" class="input-field textarea" rows="5" bind:value={form.description} placeholder="Berikan instruksi detail..."></textarea>
            </div>

            {#if auth.user?.role !== 'intern'}
              <div class="form-group">
                <label class="label" for="internSelect">Intern</label>
                <select id="internSelect" class="input-field" bind:value={form.intern_id}>
                  <option value="">Pilih Intern</option>
                  {#each interns as intern}
                    <option value={intern.id}>{intern.full_name || intern.name}</option>
                  {/each}
                </select>
              </div>
            {/if}

            <div class="form-group">
              <label class="label" for="priority">Prioritas</label>
              <select id="priority" class="input-field" bind:value={form.priority}>
                <option value="low">Low</option>
                <option value="medium">Mid</option>
                <option value="high">High</option>
              </select>
            </div>

            <div class="form-group">
              <label class="label" for="submissionMethod">Metode Pengumpulan</label>
              <select id="submissionMethod" class="input-field" bind:value={form.submission_method}>
                <option value="links">Hanya Link</option>
                <option value="files">Hanya File</option>
                <option value="both">Link & File</option>
              </select>
            </div>

            <div class="form-group">
              <label class="label" for="startDate">Tanggal Mulai</label>
              <input id="startDate" type="date" class="input-field" bind:value={form.start_date} />
            </div>

            <div class="form-group">
              <label class="label" for="deadline">Deadline</label>
              <input id="deadline" type="date" class="input-field" bind:value={form.deadline} />
            </div>

            <div class="form-group">
              <label class="label" for="deadlineTime">Waktu Deadline</label>
              <input id="deadlineTime" type="time" class="input-field" bind:value={form.deadline_time} />
            </div>
          </div>

          <div class="form-actions mt-8">
            <button class="btn-primary" onclick={handleSave} disabled={saving}>
              {saving ? 'Menyimpan...' : 'Simpan Perubahan'}
            </button>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .page-bg { min-height: 100vh; background: #f8fafc; padding: 20px 16px; }
  .container { max-width: 800px; margin: 0 auto; }
  .header { margin-bottom: 24px; }
  .title { font-size: 24px; font-weight: 800; color: #0f172a; margin: 0; }
  .subtitle { color: #64748b; font-size: 14px; margin: 4px 0 0; }
  .btn-back { display: inline-flex; align-items: center; gap: 8px; background: white; border: 1px solid #e2e8f0; padding: 8px 16px; border-radius: 99px; font-weight: 600; cursor: pointer; font-size: 13px; }
  .card { background: white; border-radius: 20px; border: 1px solid #e2e8f0; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); }
  .card-body { padding: 24px; }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .full-width { grid-column: span 2; }
  .label { font-size: 12px; font-weight: 700; color: #475569; text-transform: uppercase; }
  .input-field { width: 100%; padding: 12px; border: 1px solid #cbd5e1; border-radius: 10px; font-size: 14px; box-sizing: border-box; }
  .textarea { resize: vertical; }
  .btn-primary { background: linear-gradient(135deg, #10b981, #059669); color: white; padding: 14px; border-radius: 12px; font-weight: 700; border: none; cursor: pointer; width: 100%; box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2); }
  .loading-state { text-align: center; padding: 60px; color: #94a3b8; }
  .spinner { width: 40px; height: 40px; border: 3px solid #e2e8f0; border-top-color: #10b981; border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
