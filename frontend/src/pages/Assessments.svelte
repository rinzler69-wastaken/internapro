<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let assessments = $state([]);
  let interns = $state([]);
  let loading = $state(true);
  let submitting = $state(false);

  // Form State
  let form = $state({
    intern_id: '',
    quality_score: 80,
    speed_score: 80,
    initiative_score: 80,
    teamwork_score: 80,
    communication_score: 80,
    strengths: '',
    improvements: '',
    comments: '',
    assessment_date: new Date().toISOString().split('T')[0], // Default hari ini
  });

  async function fetchData() {
    loading = true;
    try {
      // Fetch Assessments
      const res = await api.getAssessments({ page: 1, limit: 50 });
      assessments = res.data || [];
      
      // Fetch Interns (hanya jika user bukan intern)
      if (auth.user?.role !== 'intern') {
        const internRes = await api.getInterns({ page: 1, limit: 100, status: 'active' });
        interns = internRes.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function handleCreate() {
    if (!form.intern_id) {
        alert("Silakan pilih intern terlebih dahulu.");
        return;
    }
    
    submitting = true;
    try {
      await api.createAssessment({
        ...form,
        intern_id: Number(form.intern_id),
        // Konversi string input ke number
        quality_score: Number(form.quality_score),
        speed_score: Number(form.speed_score),
        initiative_score: Number(form.initiative_score),
        teamwork_score: Number(form.teamwork_score),
        communication_score: Number(form.communication_score),
      });
      
      alert('Penilaian berhasil disimpan!');
      
      // Reset form text areas only (keep scores for convenience or reset all)
      form.strengths = '';
      form.improvements = '';
      form.comments = '';
      
      await fetchData();
    } catch (err) {
      alert(err.message || 'Gagal membuat penilaian');
    } finally {
      submitting = false;
    }
  }

  // Helper untuk menghitung rata-rata nilai
  function calculateAverage(a) {
    const total = (a.quality_score + a.speed_score + a.initiative_score + a.teamwork_score + a.communication_score);
    return (total / 5).toFixed(1);
  }

  // Helper untuk warna badge nilai
  function getScoreColor(score) {
      if (score >= 90) return 'bg-emerald-100 text-emerald-700 border-emerald-200';
      if (score >= 80) return 'bg-blue-100 text-blue-700 border-blue-200';
      if (score >= 70) return 'bg-yellow-100 text-yellow-700 border-yellow-200';
      return 'bg-red-100 text-red-700 border-red-200';
  }

  onMount(fetchData);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <div class="header">
        <div>
            <h2 class="title">Evaluasi Kinerja</h2>
            <p class="subtitle">Kelola penilaian kualitas dan perkembangan peserta magang.</p>
        </div>
        <div class="header-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                <polyline points="14 2 14 8 20 8"></polyline>
                <line x1="16" y1="13" x2="8" y2="13"></line>
                <line x1="16" y1="17" x2="8" y2="17"></line>
                <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
        </div>
    </div>

    <!-- FORM INPUT (Hanya untuk Admin/Mentor) -->
    {#if auth.user?.role !== 'intern'}
      <div class="card form-card animate-slide-up">
        <div class="card-header">
            <h3>Buat Penilaian Baru</h3>
        </div>
        
        <div class="form-grid">
            <!-- Kolom Kiri: Data & Skor -->
            <div class="form-section">
                <div class="form-group full-width">
                    <label class="label" for="intern_id">Pilih Intern</label>
                    <div class="select-wrapper">
                        <select class="input-field select" bind:value={form.intern_id} id="intern_id">
                            <option value="">-- Pilih Peserta --</option>
                            {#each interns as i}
                                <option value={i.id}>{i.full_name} - {i.school}</option>
                            {/each}
                        </select>
                        <div class="select-arrow">▼</div>
                    </div>
                </div>

                <div class="form-group full-width">
                    <label class="label" for="assessment_date">Tanggal Penilaian</label>
                    <input class="input-field" type="date" bind:value={form.assessment_date} id="assessment_date" />
                </div>

                <div class="score-grid">
                    <div class="form-group">
                        <label class="label" for="quality">Kualitas (0-100)</label>
                        <input class="input-field score-input" type="number" min="0" max="100" bind:value={form.quality_score} id="quality" />
                    </div>
                    <div class="form-group">
                        <label class="label" for="speed">Kecepatan</label>
                        <input class="input-field score-input" type="number" min="0" max="100" bind:value={form.speed_score} id="speed" />
                    </div>
                    <div class="form-group">
                        <label class="label" for="initiative">Inisiatif</label>
                        <input class="input-field score-input" type="number" min="0" max="100" bind:value={form.initiative_score} id="initiative" />
                    </div>
                    <div class="form-group">
                        <label class="label" for="teamwork">Kerjasama</label>
                        <input class="input-field score-input" type="number" min="0" max="100" bind:value={form.teamwork_score} id="teamwork" />
                    </div>
                    <div class="form-group">
                        <label class="label" for="communication">Komunikasi</label>
                        <input class="input-field score-input" type="number" min="0" max="100" bind:value={form.communication_score} id="communication" />
                    </div>
                    <!-- Indicator Average (Readonly) -->
                    <div class="form-group">
                        <p class="label muted">Rata-rata</p>
                        <div class="average-display" aria-live="polite">
                            {(
                                (Number(form.quality_score) + Number(form.speed_score) + Number(form.initiative_score) + Number(form.teamwork_score) + Number(form.communication_score)) / 5
                            ).toFixed(1)}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Kolom Kanan: Uraian -->
            <div class="form-section">
                <div class="form-group">
                    <label class="label" for="strengths">Kekuatan (Strengths)</label>
                    <textarea class="input-field text-area" rows="3" placeholder="Apa kelebihan intern ini?" bind:value={form.strengths} id="strengths"></textarea>
                </div>
                <div class="form-group">
                    <label class="label" for="improvements">Area Pengembangan (Improvements)</label>
                    <textarea class="input-field text-area" rows="3" placeholder="Apa yang perlu diperbaiki?" bind:value={form.improvements} id="improvements"></textarea>
                </div>
                <div class="form-group">
                    <label class="label" for="comments">Komentar Tambahan</label>
                    <textarea class="input-field text-area" rows="2" placeholder="Catatan lain..." bind:value={form.comments} id="comments"></textarea>
                </div>
                
                <div class="action-row">
                    <button class="btn-primary" onclick={handleCreate} disabled={submitting}>
                        {#if submitting}
                            Menyimpan...
                        {:else}
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                            Simpan Penilaian
                        {/if}
                    </button>
                </div>
            </div>
        </div>
      </div>
    {/if}

    <!-- TABEL LIST PENILAIAN -->
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
      <div class="card-header border-b">
          <h3>Riwayat Penilaian</h3>
          <span class="badge-count">{assessments.length} Data</span>
      </div>

      {#if loading}
        <div class="loading-state">Memuat data...</div>
      {:else if assessments.length === 0}
        <div class="empty-state">
            <div class="empty-icon">📝</div>
            <p>Belum ada data penilaian.</p>
        </div>
      {:else}
        <div class="table-container">
            <table class="table">
                <thead>
                    <tr>
                        <th>Intern</th>
                        <th>Tanggal</th>
                        <th>Kekuatan Utama</th>
                        <th>Area Fokus</th>
                        <th class="text-center">Skor Rata-rata</th>
                        <th class="text-right">Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#each assessments as item}
                        <tr>
                            <td>
                                <div class="user-info">
                                    <div class="avatar">{item.intern_name?.charAt(0) || 'U'}</div>
                                    <div class="user-text">
                                        <span class="name">{item.intern_name || 'Tidak Diketahui'}</span>
                                        <span class="role">Magang</span>
                                    </div>
                                </div>
                            </td>
                            <td class="date-cell">
                                <span class="date-badge">
                                    {new Date(item.assessment_date || item.created_at).toLocaleDateString('id-ID')}
                                </span>
                            </td>
                            <td class="text-cell">{item.strengths || '-'}</td>
                            <td class="text-cell">{item.improvements || '-'}</td>
                            <td class="text-center">
                                <span class={`score-badge ${getScoreColor(calculateAverage(item))}`}>
                                    {calculateAverage(item)}
                                </span>
                            </td>
                            <td class="text-right">
                                <button class="btn-icon" aria-label={`Aksi untuk ${item.intern_name || 'penilaian'}`}>
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
                                </button>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
      {/if}
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
    /* Subtle premium gradient background */
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 32px 24px;
  }

  .container {
    max-width: 1100px;
    margin: 0 auto;
  }

  /* --- HEADER --- */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 32px;
  }
  .title {
    font-size: 28px;
    font-weight: 800;
    color: #0f172a;
    margin: 0 0 6px 0;
    letter-spacing: -0.02em;
  }
  .subtitle {
    color: #64748b;
    font-size: 15px;
    margin: 0;
  }
  .header-icon {
    width: 48px; height: 48px;
    background: #ffffff;
    border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
    color: #10b981;
    box-shadow: 0 4px 12px rgba(0,0,0,0.03);
    border: 1px solid #e2e8f0;
  }

  /* --- CARDS GLOBAL --- */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
    margin-bottom: 32px;
  }
  .card-header {
    padding: 20px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .card-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
  }
  .border-b { border-bottom: 1px solid #f1f5f9; }

  /* --- FORM STYLING (PREMIUM GREEN) --- */
  .form-card {
    padding: 24px;
  }
  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 32px;
    margin-top: 20px;
  }
  @media (max-width: 768px) { .form-grid { grid-template-columns: 1fr; } }

  .form-section { display: flex; flex-direction: column; gap: 20px; }
  
  .form-group { display: flex; flex-direction: column; gap: 8px; }
  .label {
    font-size: 13px;
    font-weight: 600;
    color: #334155;
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }
  .label.muted { color: #94a3b8; }

  /* Inputs */
  .input-field {
    padding: 10px 14px;
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    font-size: 14px;
    font-family: 'Inter', sans-serif;
    color: #0f172a;
    transition: all 0.2s;
    background: #fff;
  }
  .input-field:focus {
    outline: none;
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15); /* Green Glow */
  }
  .input-field::placeholder { color: #cbd5e1; }

  /* Select Custom */
  .select-wrapper { position: relative; }
  .select { appearance: none; width: 100%; cursor: pointer; background: white; }
  .select-arrow {
    position: absolute; right: 14px; top: 50%; transform: translateY(-50%);
    font-size: 10px; color: #64748b; pointer-events: none;
  }

  /* Score Grid */
  .score-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    background: #f8fafc;
    padding: 20px;
    border-radius: 16px;
    border: 1px solid #f1f5f9;
  }
  .score-input { text-align: center; font-weight: 600; }
  
  .average-display {
    font-size: 24px;
    font-weight: 800;
    color: #10b981;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    padding: 4px;
    text-align: center;
  }

  .text-area { resize: vertical; line-height: 1.5; }

  .action-row { margin-top: 10px; display: flex; justify-content: flex-end; }
  
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); /* Premium Green Gradient */
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 10px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    display: flex; align-items: center; gap: 8px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2);
  }
  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 15px rgba(16, 185, 129, 0.3);
  }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; transform: none; }

  /* --- TABLE STYLING --- */
  .list-card { padding-bottom: 8px; }
  .badge-count {
    background: #f1f5f9; color: #64748b;
    padding: 4px 10px; border-radius: 20px;
    font-size: 12px; font-weight: 600;
  }

  .table-container { overflow-x: auto; }
  .table { width: 100%; border-collapse: separate; border-spacing: 0; }
  
  .table th {
    text-align: left;
    padding: 16px 24px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: #64748b;
    border-bottom: 1px solid #e2e8f0;
    background: #fcfcfc;
  }
  .table td {
    padding: 16px 24px;
    font-size: 14px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
    color: #334155;
  }
  .table tr:last-child td { border-bottom: none; }
  .table tr:hover td { background-color: #f8fafc; } /* Row Hover Effect */

  /* User Info in Table */
  .user-info { display: flex; align-items: center; gap: 12px; }
  .avatar {
    width: 36px; height: 36px;
    background: #0f172a; color: white;
    border-radius: 10px;
    display: flex; align-items: center; justify-content: center;
    font-weight: 600; font-size: 13px;
  }
  .user-text { display: flex; flex-direction: column; }
  .name { font-weight: 600; color: #0f172a; }
  .role { font-size: 12px; color: #94a3b8; }

  /* Badges & Text */
  .date-badge {
    background: #f1f5f9; color: #64748b;
    padding: 4px 10px; border-radius: 6px;
    font-size: 12px; font-weight: 500;
  }
  .text-cell { max-width: 250px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #64748b; }
  
  .score-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 20px;
    font-weight: 600;
    font-size: 13px;
    border: 1px solid transparent;
  }

  .text-center { text-align: center; }
  .text-right { text-align: right; }

  .btn-icon {
    background: transparent; border: none; color: #94a3b8; cursor: pointer;
    padding: 6px; border-radius: 6px; transition: all 0.2s;
  }
  .btn-icon:hover { background: #e2e8f0; color: #0f172a; }

  /* Empty State */
  .empty-state { text-align: center; padding: 60px; color: #94a3b8; }
  .empty-icon { font-size: 32px; margin-bottom: 8px; opacity: 0.5; }
  .loading-state { text-align: center; padding: 40px; color: #94a3b8; font-style: italic; }

  /* Animations */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
