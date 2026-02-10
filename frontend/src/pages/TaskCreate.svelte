<script>
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';
  import { onMount } from 'svelte';

  // State
  let title = $state('');
  let description = $state('');
  let priority = $state('medium');
  let startDate = $state('');
  let deadline = $state('');
  let deadlineTime = $state('');
  let assignTo = $state('all');
  let submissionMethod = $state('links');
  
  // Search & Selection State
  let searchQuery = $state('');
  let results = $state([]);
  let selected = $state([]);
  let loading = $state(false);

  async function searchInterns() {
    if (!searchQuery) {
      results = [];
      return;
    }
    try {
      const res = await api.searchInterns(searchQuery);
      results = res.data || [];
    } catch (err) {
      console.error(err);
    }
  }

  function toggleIntern(intern) {
    if (selected.find((i) => i.id === intern.id)) {
      selected = selected.filter((i) => i.id !== intern.id);
    } else {
      selected = [...selected, intern];
    }
  }

  function removeSelected(id) {
    selected = selected.filter((i) => i.id !== id);
  }

  async function handleSubmit() {
    if (!title) {
        alert("Judul tugas wajib diisi");
        return;
    }

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
        submission_method: submissionMethod,
        intern_ids: selected.map((i) => i.id),
      });
      goto('/tasks');
    } catch (err) {
      alert(err.message || 'Gagal membuat tugas');
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    const today = new Date().toISOString().slice(0, 10);
    if (!startDate) startDate = today;
    if (!deadlineTime) deadlineTime = '23:59';
  });
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <a href="/tasks" class="btn-back">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5"/><path d="M12 19l-7-7 7-7"/></svg>
            Kembali ke Daftar
        </a>
        <div class="mt-4">
            <h2 class="title">Buat Tugas Baru</h2>
            <p class="subtitle">Berikan penugasan kepada peserta magang.</p>
        </div>
    </div>

    <div class="grid-layout animate-slide-up">
        
        <!-- KOLOM KIRI: DETAIL TUGAS -->
        <div class="card left-card">
            <div class="card-header border-b">
                <div class="icon-circle bg-emerald">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                </div>
                <h3>Detail Tugas</h3>
            </div>
            
            <div class="card-body">
                <div class="form-group">
                    <label class="label" for="title">Judul Tugas <span class="text-red-500">*</span></label>
                    <input class="input-field" bind:value={title} id="title" placeholder="Contoh: Implementasi API Login" />
                </div>
                
                <div class="form-group mt-4">
                    <label class="label" for="description">Deskripsi & Instruksi</label>
                    <textarea class="input-field textarea" rows="6" bind:value={description} id="description" placeholder="Jelaskan detail pekerjaan yang harus dilakukan..."></textarea>
                </div>

                <div class="grid-2 mt-4">
                    <div class="form-group">
                        <label class="label" for="priority">Prioritas</label>
                        <div class="select-wrapper">
                            <select class="input-field select" bind:value={priority} id="priority">
                                <option value="low">Low (Rendah)</option>
                                <option value="medium">Medium (Sedang)</option>
                                <option value="high">High (Tinggi)</option>
                            </select>
                            <div class="select-arrow">▼</div>
                        </div>
                    </div>
                    <div class="form-group">
                        <label class="label" for="submissionMethod">Metode Pengumpulan</label>
                        <div class="select-wrapper">
                            <select class="input-field select" bind:value={submissionMethod} id="submissionMethod">
                                <option value="links">Hanya Link</option>
                                <option value="files">Hanya File</option>
                                <option value="both">Link & File</option>
                            </select>
                            <div class="select-arrow">▼</div>
                        </div>
                    </div>
                    <div class="form-group">
                        <label class="label" for="startDate">Tanggal Mulai</label>
                        <input class="input-field" type="date" bind:value={startDate} id="startDate" />
                    </div>
                </div>

                <div class="grid-2 mt-4">
                    <div class="form-group">
                        <label class="label" for="deadline">Tenggat Waktu (Deadline)</label>
                        <input class="input-field" type="date" bind:value={deadline} id="deadline" />
                    </div>
                    <div class="form-group">
                        <label class="label" for="deadlineTime">Jam Tenggat</label>
                        <input class="input-field" type="time" bind:value={deadlineTime} id="deadlineTime" />
                    </div>
                </div>
            </div>
        </div>

        <!-- KOLOM KANAN: PENUGASAN -->
        <div class="card right-card">
            <div class="card-header border-b">
                <div class="icon-circle bg-blue">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                </div>
                <h3>Penerima Tugas</h3>
            </div>

            <div class="card-body">
                <div class="form-group">
                    <label class="label" for="assignTo">Tipe Penugasan</label>
                    <div class="radio-group">
                        <label class={`radio-btn ${assignTo === 'all' ? 'active' : ''}`}>
                            <input type="radio" value="all" bind:group={assignTo} hidden>
                            <span>Semua Intern</span>
                        </label>
                        <label class={`radio-btn ${assignTo === 'selected' ? 'active' : ''}`}>
                            <input type="radio" value="selected" bind:group={assignTo} hidden>
                            <span>Pilih Manual</span>
                        </label>
                    </div>
                </div>

                {#if assignTo === 'selected'}
                    <div class="search-section mt-4 animate-fade-in">
                        <div class="form-group">
                            <label class="label" for="searchQuery">Cari Peserta</label>
                            <div class="search-wrapper">
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                                <input class="input-field pl-10" bind:value={searchQuery} oninput={searchInterns} placeholder="Ketik nama peserta..." id="searchQuery" />
                            </div>
                        </div>

                        {#if selected.length > 0}
                            <div class="selected-tags mt-3">
                                {#each selected as item}
                                    <div class="tag">
                                        {item.label.split(' - ')[0]} 
                                        <button onclick={() => removeSelected(item.id)} class="btn-remove">×</button>
                                    </div>
                                {/each}
                            </div>
                        {/if}

                        <div class="results-list mt-3">
                            {#if results.length > 0}
                                {#each results as intern}
                                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                                    <div 
                                        class="intern-card {selected.find((i) => i.id === intern.id) ? 'selected' : ''}"
                                        onclick={() => toggleIntern(intern)}
                                    >
                                        <div class="checkbox-circle">
                                            {#if selected.find((i) => i.id === intern.id)}✓{/if}
                                        </div>
                                        <div class="intern-info">
                                            <span class="intern-name">{intern.label}</span>
                                        </div>
                                    </div>
                                {/each}
                            {:else if searchQuery}
                                <div class="empty-search">Tidak ada hasil ditemukan.</div>
                            {/if}
                        </div>
                    </div>
                {:else}
                    <div class="info-box mt-4">
                        <p>Tugas akan diberikan kepada seluruh peserta magang yang statusnya <strong>Aktif</strong>.</p>
                    </div>
                {/if}
            </div>

            <div class="card-footer">
                <button class="btn-primary w-full" onclick={handleSubmit} disabled={loading}>
                    {#if loading}
                        <div class="spinner-small"></div> Menyimpan...
                    {:else}
                        Buat Tugas
                    {/if}
                </button>
            </div>
        </div>

    </div>
  </div>
</div>

<style>
  :global(body) {
    font-family: 'Geist', 'Inter', sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 32px 24px;
  }

  .container { max-width: 1100px; margin: 0 auto; }

  /* HEADER */
  .header { margin-bottom: 32px; }
  .title { font-size: 28px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; }
  .subtitle { color: #64748b; font-size: 15px; margin: 0; }
  .btn-back {
      display: inline-flex; align-items: center; gap: 8px; color: #64748b; 
      font-weight: 600; font-size: 13px; text-decoration: none; transition: all 0.2s;
      background: white; padding: 8px 14px; border-radius: 99px; border: 1px solid #e2e8f0;
  }
  .btn-back:hover { color: #0f172a; border-color: #cbd5e1; transform: translateX(-2px); }

  /* LAYOUT */
  .grid-layout {
    display: grid; grid-template-columns: 1fr; gap: 24px;
  }
  @media (min-width: 900px) { .grid-layout { grid-template-columns: 1.5fr 1fr; } }

  /* CARDS */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
    display: flex; flex-direction: column; height: fit-content;
  }
  .card-header {
    padding: 20px 24px; border-bottom: 1px solid #f1f5f9; display: flex; align-items: center; gap: 12px;
    background: #ffffff;
  }
  .card-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
  .icon-circle { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
  .bg-emerald { background: #ecfdf5; color: #059669; }
  .bg-blue { background: #eff6ff; color: #2563eb; }

  .card-body { padding: 24px; flex: 1; }
  .card-footer { padding: 20px 24px; border-top: 1px solid #f8fafc; background: #fcfcfc; }

  /* FORM */
  .form-group { margin-bottom: 0; }
  .label { display: block; font-size: 12px; font-weight: 600; color: #475569; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.02em; }
  
  .input-field {
    width: 100%; padding: 12px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; transition: all 0.2s; background: #fff; box-sizing: border-box;
    font-family: 'Inter', sans-serif;
  }
  .input-field:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }
  .textarea { resize: vertical; line-height: 1.5; }
  
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  @media (max-width: 600px) { .grid-2 { grid-template-columns: 1fr; } }

  /* Select */
  .select-wrapper { position: relative; }
  .select { appearance: none; cursor: pointer; }
  .select-arrow {
    position: absolute; right: 14px; top: 50%; transform: translateY(-50%);
    font-size: 10px; color: #64748b; pointer-events: none;
  }

  /* Radio Group */
  .radio-group { display: flex; gap: 10px; }
  .radio-btn {
    flex: 1; padding: 10px; border: 1px solid #e2e8f0; border-radius: 10px;
    text-align: center; cursor: pointer; font-size: 13px; font-weight: 500; color: #64748b;
    transition: all 0.2s; background: #f8fafc;
  }
  .radio-btn:hover { background: #f1f5f9; }
  .radio-btn.active { border-color: #10b981; background: #ecfdf5; color: #059669; font-weight: 600; }

  /* Search & Selection */
  .search-wrapper { position: relative; }
  .search-icon { position: absolute; left: 12px; top: 12px; color: #94a3b8; pointer-events: none; }
  .pl-10 { padding-left: 40px; }

  .results-list { max-height: 300px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
  
  .intern-card {
    display: flex; align-items: center; gap: 12px; padding: 10px;
    border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer;
    transition: all 0.2s; background: white;
  }
  .intern-card:hover { border-color: #cbd5e1; background: #f8fafc; }
  .intern-card.selected { border-color: #10b981; background: #ecfdf5; }
  
  .checkbox-circle {
    width: 20px; height: 20px; border-radius: 50%; border: 2px solid #cbd5e1;
    display: flex; align-items: center; justify-content: center; font-size: 12px; color: white;
    flex-shrink: 0;
  }
  .intern-card.selected .checkbox-circle { background: #10b981; border-color: #10b981; }
  
  .intern-name { font-size: 13px; color: #334155; font-weight: 500; }
  .empty-search { text-align: center; font-size: 13px; color: #94a3b8; padding: 20px; font-style: italic; }

  .selected-tags { display: flex; flex-wrap: wrap; gap: 6px; }
  .tag {
    background: #10b981; color: white; font-size: 11px; padding: 4px 8px; border-radius: 6px;
    display: flex; align-items: center; gap: 6px; font-weight: 500;
  }
  .btn-remove { background: none; border: none; color: white; cursor: pointer; font-size: 14px; padding: 0; opacity: 0.8; }
  .btn-remove:hover { opacity: 1; }

  .info-box { background: #eff6ff; border: 1px solid #bfdbfe; color: #1e40af; padding: 12px; border-radius: 10px; font-size: 13px; line-height: 1.5; }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white;
    padding: 12px; border-radius: 10px; font-weight: 600; font-size: 14px; border: none;
    cursor: pointer; width: 100%; transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  .spinner-small { width: 16px; height: 16px; border: 2px solid white; border-top-color: transparent; border-radius: 50%; animation: spin 1s linear infinite; display: inline-block; vertical-align: middle; margin-right: 6px; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s ease-out forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
