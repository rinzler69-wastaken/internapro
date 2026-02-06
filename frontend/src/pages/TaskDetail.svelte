<script>
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  const { route } = $props();
  let taskId = $state('');

  let task = $state(null);
  let loading = $state(true);
  let submissionNotes = $state('');
  let links = $state([{ label: '', url: '' }]);
  let reviewAction = $state('approve');
  let reviewScore = $state(80);
  let reviewFeedback = $state('');
  let statusUpdate = $state('pending');

  const statusLabels = {
    pending: 'Pending',
    scheduled: 'Scheduled',
    in_progress: 'In Progress',
    submitted: 'Submitted',
    revision: 'Revision',
    completed: 'Completed',
  };

  function formatDate(value) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  async function fetchTask() {
    loading = true;
    try {
      const res = await api.getTask(taskId);
      task = res.data;
      statusUpdate = task.status;
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function addLink() {
    links = [...links, { label: '', url: '' }];
  }

  async function submitTask() {
    try {
      await api.submitTask(task.id, { submission_notes: submissionNotes, links });
      await fetchTask();
    } catch (err) {
      alert(err.message || 'Gagal submit');
    }
  }

  async function reviewTask() {
    try {
      await api.reviewTask(task.id, { action: reviewAction, score: reviewAction === 'approve' ? reviewScore : null, feedback: reviewFeedback });
      await fetchTask();
    } catch (err) {
      alert(err.message || 'Gagal review');
    }
  }

  async function updateStatus() {
    try {
      await api.updateTaskStatus(task.id, { status: statusUpdate });
      await fetchTask();
    } catch (err) {
      alert(err.message || 'Gagal update status');
    }
  }

  $effect(() => {
    const params = route?.result?.path?.params || {};
    if (params?.id && params.id !== taskId) {
      taskId = params.id;
      fetchTask();
    }
  });
</script>

{#if loading}
  <div class="card">Memuat detail tugas...</div>
{:else if !task}
  <div class="card">Tugas tidak ditemukan.</div>
{:else}
  <div class="card" style="margin-bottom:16px;">
    <h3>{task.title}</h3>
    <p class="text-muted">{task.description || 'Tidak ada deskripsi'}</p>
    <div style="display:flex; gap:10px; flex-wrap:wrap; margin-top:12px;">
      <span class={`status status-${task.status || 'pending'}`}>
        {statusLabels[task.status] || task.status || '-'}
      </span>
      <span class="badge">{task.priority || '-'}</span>
      {#if task.deadline}
        <span class="badge warning">Deadline: {formatDate(task.deadline)}</span>
      {/if}
    </div>
  </div>

  {#if auth.user?.role === 'intern'}
    <div class="card" style="margin-bottom:16px;">
      <h4>Update Status</h4>
      <div style="display:flex; gap:10px; align-items:center;">
        <select class="select" bind:value={statusUpdate}>
          <option value="pending">Pending</option>
          <option value="in_progress">In Progress</option>
        </select>
        <button class="btn btn-outline" onclick={updateStatus}>Simpan</button>
      </div>
    </div>

    <div class="card">
      <h4>Submit Tugas</h4>
      <div class="form-group">
        <label class="form-label" for="submissionNotes">Catatan</label>
        <textarea class="textarea" rows="3" bind:value={submissionNotes} id="submissionNotes"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label" for="link-label-0">Link Submission</label>
        {#each links as link, index}
          <div style="display:grid; grid-template-columns: 1fr 2fr; gap:8px; margin-bottom:8px;">
            <input class="input" placeholder="Label" bind:value={link.label} id={`link-label-${index}`} />
            <input class="input" placeholder="URL" bind:value={link.url} />
          </div>
        {/each}
        <button class="btn btn-ghost" onclick={addLink}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" style="width:14px; height:14px;">
            <path d="M12 5v14" />
            <path d="M5 12h14" />
          </svg>
          Tambah Link
        </button>
      </div>
      <button class="btn btn-primary" onclick={submitTask}>Kirim Tugas</button>
    </div>
  {:else}
    <div class="card">
      <h4>Review Tugas</h4>
      <div style="display:flex; gap:8px; margin-bottom:12px;">
        <select class="select" bind:value={reviewAction}>
          <option value="approve">Approve</option>
          <option value="revision">Revision</option>
        </select>
        {#if reviewAction === 'approve'}
          <input class="input" type="number" min="0" max="100" bind:value={reviewScore} style="max-width:120px;" />
        {/if}
      </div>
      <div class="form-group">
        <label class="form-label" for="reviewFeedback">Feedback</label>
        <textarea class="textarea" rows="3" bind:value={reviewFeedback} id="reviewFeedback"></textarea>
      </div>
      <button class="btn btn-primary" onclick={reviewTask}>Simpan Review</button>
    </div>
  {/if}
{/if}
