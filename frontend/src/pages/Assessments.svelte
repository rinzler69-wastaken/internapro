<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  let assessments = $state([]);
  let interns = $state([]);
  let loading = $state(false);

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
    assessment_date: '',
  });

  async function fetchData() {
    loading = true;
    try {
      const res = await api.getAssessments({ page: 1, limit: 20 });
      assessments = res.data || [];
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
      await api.createAssessment({
        ...form,
        intern_id: Number(form.intern_id),
      });
      await fetchData();
    } catch (err) {
      alert(err.message || 'Gagal membuat penilaian');
    }
  }

  onMount(fetchData);
</script>

<div class="card" style="margin-bottom:16px;">
  <h3>Penilaian</h3>
  <p class="text-muted">Pantau kualitas dan performa peserta magang.</p>
</div>

{#if auth.user?.role !== 'intern'}
  <div class="card" style="margin-bottom:16px;">
    <h4>Tambah Penilaian</h4>
    <div class="form-group">
      <label class="form-label" for="intern_id">Intern</label>
      <select class="select" bind:value={form.intern_id} id="intern_id">
        <option value="">Pilih Intern</option>
        {#each interns as i}
          <option value={i.id}>{i.full_name}</option>
        {/each}
      </select>
    </div>
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap:12px;">
      <div class="form-group">
        <label class="form-label" for="quality_score">Quality</label>
        <input class="input" type="number" min="0" max="100" bind:value={form.quality_score} id="quality_score" />
      </div>
      <div class="form-group">
        <label class="form-label" for="speed_score">Speed</label>
        <input class="input" type="number" min="0" max="100" bind:value={form.speed_score} id="speed_score" />
      </div>
      <div class="form-group">
        <label class="form-label" for="initiative_score">Initiative</label>
        <input class="input" type="number" min="0" max="100" bind:value={form.initiative_score} id="initiative_score" />
      </div>
      <div class="form-group">
        <label class="form-label" for="teamwork_score">Teamwork</label>
        <input class="input" type="number" min="0" max="100" bind:value={form.teamwork_score} id="teamwork_score" />
      </div>
      <div class="form-group">
        <label class="form-label" for="communication_score">Communication</label>
        <input class="input" type="number" min="0" max="100" bind:value={form.communication_score} id="communication_score" />
      </div>
    </div>
    <div class="form-group">
      <label class="form-label" for="strengths">Strengths</label>
      <textarea class="textarea" rows="2" bind:value={form.strengths} id="strengths"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="improvements">Improvements</label>
      <textarea class="textarea" rows="2" bind:value={form.improvements} id="improvements"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="comments">Comments</label>
      <textarea class="textarea" rows="2" bind:value={form.comments} id="comments"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="assessment_date">Tanggal</label>
      <input class="input" type="date" bind:value={form.assessment_date} id="assessment_date" />
    </div>
    <button class="btn btn-primary" onclick={handleCreate}>Simpan Penilaian</button>
  </div>
{/if}

<div class="card">
  <h4>Daftar Penilaian</h4>
  {#if loading}
    <div>Memuat...</div>
  {:else if assessments.length === 0}
    <div class="empty-state">Belum ada penilaian.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Intern</th>
          <th>Skor</th>
          <th>Kategori</th>
        </tr>
      </thead>
      <tbody>
        {#each assessments as a}
          <tr>
            <td>{a.intern_name || '-'}</td>
            <td>{a.score}</td>
            <td><span class="status">{a.category || '-'}</span></td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
