<script>
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";

  const { route } = $props();
  let taskId = $state("");

  let task = $state(null);
  let loading = $state(true);
  let submissionNotes = $state("");
  let submissionFile = $state(null);
  let links = $state([{ label: "", url: "" }]);
  let reviewAction = $state("approve");
  let reviewScore = $state(80);
  let reviewFeedback = $state("");
  let statusUpdate = $state("pending");
  let submitting = $state(false);
  let showSubmissionForm = $state(true);
  let showReviewForm = $state(false);

  // --- HELPERS ---
  const statusLabels = {
    pending: "Pending",
    scheduled: "Terjadwal",
    in_progress: "Dalam Proses",
    submitted: "Menunggu Review",
    revision: "Revisi",
    completed: "Selesai",
  };

  function getStatusColor(status) {
    switch (status) {
      case "completed":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "approved":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "submitted":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "in_progress":
        return "bg-amber-100 text-amber-700 border-amber-200";
      case "revision":
        return "bg-rose-100 text-rose-700 border-rose-200";
      default:
        return "bg-slate-100 text-slate-600 border-slate-200";
    }
  }

  function getPriorityColor(priority) {
    switch (priority) {
      case "high":
        return "text-rose-600 bg-rose-50 border-rose-100";
      case "medium":
        return "text-amber-600 bg-amber-50 border-amber-100";
      default:
        return "text-emerald-600 bg-emerald-50 border-emerald-100";
    }
  }

  function formatDate(value) {
    if (!value) return "-";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString("id-ID", {
      day: "numeric",
      month: "long",
      year: "numeric",
    });
  }
  function formatTime(value) {
    if (!value) return "";
    return value.slice(0, 5);
  }

  function withBase(url) {
    if (!url) return "";

    // Check if this is an uploads path
    const isUpload = url.startsWith("/uploads") || url.startsWith("uploads/");
    const token = auth.token;

    // For uploads, add token but don't use API base (uploads are served from root, not /api)
    if (isUpload) {
      const normalized = url.startsWith("/") ? url : `/${url}`;
      const tokenQS = token ? `?token=${token}` : "";
      return `${normalized}${tokenQS}`;
    }

    // For other URLs, use normal API base logic
    const tokenQS = token
      ? `${url.includes("?") ? "&" : "?"}token=${token}`
      : "";
    if (url.startsWith("http")) return `${url}${tokenQS}`;
    const base = import.meta.env.VITE_API_URL || "";
    const normalized = url.startsWith("/") ? url : `/${url}`;
    return `${base}${normalized}${tokenQS}`;
  }

  function ensureProtocol(url) {
    if (!url) return "";
    // If already has protocol, return as-is
    if (url.startsWith("http://") || url.startsWith("https://")) {
      return url;
    }
    // Add https:// for external links
    return `https://${url}`;
  }

  // --- API CALLS ---
  async function fetchTask() {
    loading = true;
    try {
      const res = await api.getTask(taskId);
      task = res.data;
      statusUpdate = task.status;
      reviewScore = task.score ?? 80;
      reviewFeedback = task.admin_feedback ?? "";
      showSubmissionForm =
        auth.user?.role === "intern" &&
        !["submitted", "completed"].includes(task.status);
      showReviewForm =
        auth.user?.role !== "intern" && task.status === "submitted";

      // prefill links if needed
      if (
        task.submission_links &&
        Array.isArray(task.submission_links) &&
        task.submission_links.length > 0
      ) {
        links = task.submission_links;
      }

      // Jika ada submission sebelumnya, bisa di-load ke state (opsional, tergantung struktur API)
      // Disini kita asumsi reset form submission
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function addLink() {
    links = [...links, { label: "", url: "" }];
  }

  function removeLink(index) {
    links = links.filter((_, i) => i !== index);
  }

  async function submitTask() {
    submitting = true;
    try {
      const formData = new FormData();
      formData.append("submission_notes", submissionNotes);
      formData.append("links", JSON.stringify(links));
      if (submissionFile) {
        formData.append("file", submissionFile);
      }

      await api.submitTask(task.id, formData);
      await fetchTask();
      alert("Tugas berhasil dikumpulkan!");
    } catch (err) {
      alert(err.message || "Gagal submit");
    } finally {
      submitting = false;
    }
  }

  async function reviewTask() {
    submitting = true;
    try {
      await api.reviewTask(task.id, {
        action: reviewAction,
        score: reviewAction === "approve" ? reviewScore : null,
        feedback: reviewFeedback,
      });
      await fetchTask();
      alert("Review berhasil disimpan");
    } catch (err) {
      alert(err.message || "Gagal review");
    } finally {
      submitting = false;
    }
  }

  async function updateStatus() {
    try {
      await api.updateTaskStatus(task.id, { status: statusUpdate });
      await fetchTask();
      alert("Status diperbarui");
    } catch (err) {
      alert(err.message || "Gagal update status");
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

<div class="page-bg">
  <div class="container animate-fade-in">
    <!-- Header -->
    <div class="header">
      <a href="/tasks" class="btn-back">
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><path d="M19 12H5" /><path d="M12 19l-7-7 7-7" /></svg
        >
        Kembali ke Daftar
      </a>
      <div class="mt-4">
        <h2 class="title">Detail Tugas</h2>
        <p class="subtitle">Informasi lengkap dan progres pengerjaan.</p>
      </div>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat tugas...</p>
      </div>
    {:else if !task}
      <div class="empty-state">
        <div class="empty-icon">🚫</div>
        <p>Tugas tidak ditemukan.</p>
        <a href="/tasks" class="btn-primary mt-4 inline-block">Kembali</a>
      </div>
    {:else}
      <div class="grid-layout animate-slide-up">
        <!-- KOLOM KIRI: INFO TUGAS -->
        <div class="card detail-card">
          <div class="card-header border-b">
            <div class="header-top">
              <span class={`priority-badge ${getPriorityColor(task.priority)}`}>
                {task.priority || "Normal"} Priority
              </span>
              <span class={`status-badge ${getStatusColor(task.status)}`}>
                {statusLabels[task.status] || task.status}
              </span>
            </div>
            <h3 class="task-title">{task.title}</h3>
          </div>

          <div class="card-body">
            <div class="meta-grid">
              <div class="meta-item">
                <span class="label">Deadline</span>
                <div class="value flex items-center gap-2">
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    class="text-rose-500"
                    ><rect
                      x="3"
                      y="4"
                      width="18"
                      height="18"
                      rx="2"
                      ry="2"
                    /><line x1="16" y1="2" x2="16" y2="6" /><line
                      x1="8"
                      y1="2"
                      x2="8"
                      y2="6"
                    /><line x1="3" y1="10" x2="21" y2="10" /></svg
                  >
                  {formatDate(task.deadline)}
                </div>
              </div>
              {#if task.deadline_time}
                <div class="meta-item">
                  <span class="label">Pukul</span>
                  <div class="value flex items-center gap-2">
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      class="text-slate-500"
                      ><circle cx="12" cy="12" r="10" /><polyline
                        points="12 6 12 12 16 14"
                      /></svg
                    >
                    {formatTime(task.deadline_time)}
                  </div>
                </div>
              {/if}
              {#if task.start_date}
                <div class="meta-item">
                  <span class="label">Mulai</span>
                  <div class="value flex items-center gap-2">
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      ><path d="M5 3v4" /><path d="M19 3v4" /><rect
                        width="14"
                        height="14"
                        x="5"
                        y="7"
                        rx="2"
                      /><path d="M12 11h4v4" /></svg
                    >
                    {formatDate(task.start_date)}
                  </div>
                </div>
              {/if}
              <div class="meta-item">
                <span class="label">Diberikan Oleh</span>
                <div class="value flex items-center gap-2">
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    ><path
                      d="M5.5 21h13M12 17c-2.5-2-4-4-4-6a4 4 0 1 1 8 0c0 2-1.5 4-4 6Z"
                    /></svg
                  >
                  {task.assigned_by_name || "-"}
                </div>
              </div>
              {#if task.intern_name}
                <div class="meta-item">
                  <span class="label">Ditugaskan ke</span>
                  <div class="value flex items-center gap-2">
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      ><circle cx="12" cy="7" r="4" /><path
                        d="M6 21v-1a6 6 0 0 1 12 0v1"
                      /></svg
                    >
                    {task.intern_name}
                  </div>
                </div>
              {/if}
            </div>

            <div class="description-box">
              <h4 class="desc-label">Deskripsi & Instruksi</h4>
              <p class="desc-text">
                {task.description ||
                  "Tidak ada deskripsi detail untuk tugas ini."}
              </p>
            </div>

            <!-- METODE PENGUMPULAN INFO -->
            <div
              class="description-box"
              style="margin-top: 1rem; border-left: 4px solid #6366f1;"
            >
              <h4 class="desc-label" style="color: #6366f1;">
                📋 Metode Pengumpulan
              </h4>
              <div
                style="display: flex; align-items: center; gap: 0.75rem; margin-top: 0.5rem;"
              >
                {#if task.submission_method === "links"}
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#6366f1"
                    stroke-width="2"
                  >
                    <path
                      d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
                    />
                    <path
                      d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
                    />
                  </svg>
                  <span style="font-weight: 500; color: #1e293b;"
                    >Link Pengumpulan</span
                  >
                {:else if task.submission_method === "files"}
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#6366f1"
                    stroke-width="2"
                  >
                    <path
                      d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                    />
                    <polyline points="14 2 14 8 20 8" />
                    <line x1="12" y1="18" x2="12" y2="12" />
                    <line x1="9" y1="15" x2="15" y2="15" />
                  </svg>
                  <span style="font-weight: 500; color: #1e293b;"
                    >Upload File</span
                  >
                {:else}
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#6366f1"
                    stroke-width="2"
                  >
                    <path
                      d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
                    />
                    <path
                      d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
                    />
                  </svg>
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#6366f1"
                    stroke-width="2"
                  >
                    <path
                      d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                    />
                    <polyline points="14 2 14 8 20 8" />
                  </svg>
                  <span style="font-weight: 500; color: #1e293b;"
                    >Link atau File</span
                  >
                {/if}
              </div>
            </div>
          </div>
        </div>

        <!-- KOLOM KANAN: AKSI (INTERN / ADMIN) -->
        <div class="action-column">
          <!-- SUBMISSION INFO CARD -->
          {#if (task.submission_links && task.submission_links.length > 0) || task.score}
            <div class="card mb-6">
              <div class="card-header border-b">
                <h4>Detail Pengumpulan & Penilaian</h4>
              </div>
              <div class="card-body">
                {#if task.submission_links && task.submission_links.length > 0}
                  <div class="submission-box" style="margin-top: 0;">
                    <div class="desc-label">Link Pengumpulan</div>
                    <div class="links-list">
                      {#each task.submission_links as link}
                        <a
                          class="link-pill"
                          href={ensureProtocol(link.url)}
                          target="_blank"
                          rel="noreferrer"
                        >
                          <span>{link.label}</span>
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path d="M7 17 17 7" /><path d="M7 7h10v10" /></svg
                          >
                        </a>
                      {/each}
                    </div>
                    {#if task.submission_file}
                      <div class="desc-label mt-4">File Terlampir</div>
                      <a
                        class="link-pill"
                        href={withBase(`/uploads/${task.submission_file}`)}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <span>Unduh File</span>
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          ><path
                            d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                          /><polyline points="7 10 12 15 17 10" /><line
                            x1="12"
                            y1="15"
                            x2="12"
                            y2="3"
                          /></svg
                        >
                      </a>
                    {/if}
                    {#if task.submission_notes}
                      <div class="notes-box">{task.submission_notes}</div>
                    {/if}
                  </div>
                {/if}

                {#if task.score}
                  <div
                    class="grade-box"
                    style={task.submission_links?.length
                      ? "margin-top: 16px;"
                      : "margin-top: 0;"}
                  >
                    <div class="grade-label">Nilai Akhir</div>
                    <div class="grade-value">{task.score}</div>
                    {#if task.admin_feedback}
                      <div class="grade-feedback">"{task.admin_feedback}"</div>
                    {/if}
                  </div>
                {/if}
              </div>
            </div>
          {/if}

          {#if auth.user?.role === "intern"}
            <!-- INTERN ACTIONS -->

            {#if task.status === "pending" || task.status === "in_progress"}
              <!-- 1. Update Status -->
              <div class="card mb-6">
                <div class="card-header border-b">
                  <h4>Update Status</h4>
                </div>
                <div class="card-body">
                  <div class="status-control">
                    <select
                      class="input-field select"
                      bind:value={statusUpdate}
                    >
                      <option value="pending">Pending</option>
                      <option value="in_progress">Sedang Dikerjakan</option>
                    </select>
                    <button
                      class="btn-outline w-full mt-3"
                      onclick={updateStatus}>Update Status</button
                    >
                  </div>
                </div>
              </div>
            {/if}

            {#if showSubmissionForm}
              <!-- 2. Submission Form -->
              <div class="card">
                <div class="card-header border-b bg-green-50/30">
                  <h4>
                    {task.status === "revision"
                      ? "Kumpulkan Revisi"
                      : "Pengumpulan Tugas"}
                  </h4>
                </div>
                <div class="card-body">
                  <div class="form-group mb-4">
                    <label class="label" for="notes">Catatan Pengerjaan</label>
                    <textarea
                      class="input-field textarea"
                      id="notes"
                      rows="4"
                      bind:value={submissionNotes}
                      placeholder="Jelaskan apa yang sudah Anda kerjakan..."
                    ></textarea>
                  </div>

                  {#if task.submission_method === "links" || task.submission_method === "both" || !task.submission_method}
                    <div class="form-group mb-4">
                      <p class="label">Lampiran Link</p>
                      {#each links as link, index}
                        <div class="link-row mb-2">
                          <input
                            class="input-field"
                            placeholder="Judul (mis: Github)"
                            bind:value={link.label}
                          />
                          <input
                            class="input-field"
                            placeholder="https://..."
                            bind:value={link.url}
                          />
                          {#if links.length > 1}
                            <button
                              class="btn-remove"
                              onclick={() => removeLink(index)}
                              title="Hapus">×</button
                            >
                          {/if}
                        </div>
                      {/each}
                      <button class="btn-text" onclick={addLink}
                        >+ Tambah Link Lain</button
                      >
                    </div>
                  {/if}

                  {#if task.submission_method === "files" || task.submission_method === "both"}
                    <div class="form-group mb-4">
                      <p class="label">Unggah File</p>
                      <div class="file-input-wrapper">
                        <input
                          type="file"
                          class="input-field"
                          onchange={(e) => {
                            const files = e.currentTarget.files;
                            if (files && files[0]) {
                              submissionFile = files[0];
                            }
                          }}
                        />
                        {#if submissionFile}
                          <p class="text-xs text-emerald-600 mt-1">
                            File terpilih: {submissionFile.name}
                          </p>
                        {/if}
                      </div>
                    </div>
                  {/if}

                  <button
                    class="btn-primary w-full"
                    onclick={submitTask}
                    disabled={submitting}
                  >
                    {#if submitting}Mengirim...{:else}Kirim Tugas{/if}
                  </button>
                </div>
              </div>
            {/if}
          {:else}
            <!-- ADMIN/SUPERVISOR ACTIONS -->
            {#if showReviewForm}
              <div class="card">
                <div class="card-header border-b bg-blue-50/30">
                  <h4>Review Tugas</h4>
                </div>
                <div class="card-body">
                  <div class="form-group mb-4">
                    <p class="label">Keputusan</p>
                    <div class="radio-group">
                      <label
                        class={`radio-btn ${reviewAction === "approve" ? "active-green" : ""}`}
                      >
                        <input
                          type="radio"
                          value="approve"
                          bind:group={reviewAction}
                          hidden
                        />
                        ✓ Terima
                      </label>
                      <label
                        class={`radio-btn ${reviewAction === "revision" ? "active-red" : ""}`}
                      >
                        <input
                          type="radio"
                          value="revision"
                          bind:group={reviewAction}
                          hidden
                        />
                        ↺ Revisi
                      </label>
                    </div>
                  </div>

                  {#if reviewAction === "approve"}
                    <div class="form-group mb-4 animate-fade-in">
                      <label class="label" for="score">Nilai (0-100)</label>
                      <input
                        class="input-field score-input"
                        type="number"
                        min="0"
                        max="100"
                        bind:value={reviewScore}
                        id="score"
                      />
                    </div>
                  {/if}

                  <div class="form-group mb-4">
                    <label class="label" for="feedback"
                      >Feedback / Komentar</label
                    >
                    <textarea
                      class="input-field textarea"
                      rows="3"
                      bind:value={reviewFeedback}
                      id="feedback"
                      placeholder="Berikan masukan..."
                    ></textarea>
                  </div>

                  <button
                    class="btn-primary w-full"
                    onclick={reviewTask}
                    disabled={submitting}
                  >
                    {#if submitting}Menyimpan...{:else}Simpan Review{/if}
                  </button>
                </div>
              </div>
            {/if}
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
  }
  .container {
    max-width: 1000px;
    margin: 0 auto;
  }

  /* HEADER */
  .header {
    margin-bottom: 24px;
  }
  .title {
    font-size: 20px;
    font-weight: 800;
    color: #0f172a;
    margin: 0 0 6px 0;
  }
  .subtitle {
    color: #64748b;
    font-size: 14px;
    margin: 0;
  }
  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: #64748b;
    font-weight: 500;
    text-decoration: none;
    transition: color 0.2s;
  }
  .btn-back:hover {
    color: #0f172a;
  }

  /* LAYOUT */
  .grid-layout {
    display: grid;
    grid-template-columns: 1fr;
    gap: 24px;
  }
  @media (min-width: 900px) {
    .grid-layout {
      grid-template-columns: 2fr 1fr;
    }
  }

  /* CARDS */
  .card {
    background: white;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
  }
  .card-header {
    padding: 12px;
  }
  .card-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #334155;
  }
  .card-header h4 {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: #334155;
    text-transform: uppercase;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }
  .card-body {
    padding: 12px;
  }

  /* DETAIL CARD STYLES */
  .header-top {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .task-title {
    font-size: 20px;
    font-weight: 800;
    padding-top: 24px;
    color: #0f172a;
    margin: 10px 0 4px 0;
  }

  .badge {
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    border: 1px solid transparent;
  }
  .status-badge {
    padding: 4px 12px;
    border-radius: 99px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
    border: 1px solid transparent;
  }
  .priority-badge {
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    border: 1px solid transparent;
  }

  .meta-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
    margin-bottom: 24px;
    padding-bottom: 24px;
    border-bottom: 1px dashed #e2e8f0;
  }
  @media (min-width: 640px) {
    .meta-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
  .meta-item .label {
    font-size: 11px;
    text-transform: uppercase;
    font-weight: 600;
    color: #94a3b8;
    display: block;
    margin-bottom: 4px;
  }
  .meta-item .value {
    font-size: 14px;
    font-weight: 600;
    color: #334155;
  }

  .description-box {
    background: #f8fafc;
    border: 1px solid #f1f5f9;
    padding: 16px;
    border-radius: 12px;
  }
  .desc-label {
    font-size: 14px;
    font-weight: 700;
    color: #334155;
    margin: 0 0 12px 0;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .desc-text {
    font-size: 16px;
    line-height: 1.7;
    color: #334155;
    margin: 0;
    white-space: pre-wrap;
  }
  .submission-box {
    margin-top: 0;
    padding: 12px;
  }
  .links-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .link-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 10px;
    background: #eef2ff;
    color: #4338ca;
    text-decoration: none;
    font-weight: 600;
    font-size: 13px;
    transition: all 0.2s;
  }
  .link-pill:hover {
    background: #e0e7ff;
  }
  .notes-box {
    margin-top: 10px;
    background: #fff7ed;
    border: 1px solid #fed7aa;
    color: #c2410c;
    padding: 12px;
    border-radius: 10px;
    font-size: 14px;
  }

  .grade-box {
    margin-top: 24px;
    background: #ecfdf5;
    border: 1px solid #a7f3d0;
    padding: 20px;
    border-radius: 12px;
    text-align: center;
  }
  .grade-label {
    font-size: 12px;
    font-weight: 600;
    color: #065f46;
    text-transform: uppercase;
  }
  .grade-value {
    font-size: 42px;
    font-weight: 800;
    color: #059669;
    line-height: 1;
    margin: 4px 0;
  }
  .grade-feedback {
    font-size: 14px;
    color: #047857;
    font-style: italic;
  }

  /* FORM STYLES */
  .label {
    display: block;
    font-size: 12px;
    font-weight: 600;
    color: #475569;
    margin-bottom: 8px;
    text-transform: uppercase;
  }
  .input-field {
    width: 100%;
    padding: 12px;
    border: 1px solid #cbd5e1;
    border-radius: 10px;
    font-size: 14px;
    color: #0f172a;
    transition: all 0.2s;
    background: #fff;
    box-sizing: border-box;
  }
  .input-field:focus {
    outline: none;
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  }
  .textarea {
    resize: vertical;
  }
  .score-input {
    font-size: 24px;
    font-weight: 600;
    text-align: center;
    color: #10b981;
  }

  .link-row {
    display: grid;
    grid-template-columns: 1fr 2fr auto;
    gap: 8px;
    align-items: center;
  }
  .btn-remove {
    width: 32px;
    height: 32px;
    border: 1px solid #fecaca;
    background: #fef2f2;
    color: #ef4444;
    border-radius: 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
  }
  .btn-remove:hover {
    background: #fee2e2;
  }

  /* Radio Group */
  .radio-group {
    display: flex;
    gap: 10px;
  }
  .radio-btn {
    flex: 1;
    padding: 10px;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    text-align: center;
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
    color: #64748b;
    transition: all 0.2s;
    background: #f8fafc;
  }
  .radio-btn:hover {
    background: #f1f5f9;
  }
  .radio-btn.active-green {
    border-color: #10b981;
    background: #ecfdf5;
    color: #059669;
  }
  .radio-btn.active-red {
    border-color: #f43f5e;
    background: #fff1f2;
    color: #e11d48;
  }

  /* BUTTONS */
  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    padding: 12px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    border: none;
    cursor: pointer;
    width: 100%;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  .btn-primary:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
  }
  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-outline {
    background: white;
    border: 1px solid #cbd5e1;
    color: #475569;
    padding: 10px;
    border-radius: 10px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
    width: 100%;
  }
  .btn-outline:hover {
    border-color: #10b981;
    color: #059669;
    background: #ecfdf5;
  }

  .btn-text {
    background: none;
    border: none;
    color: #10b981;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    padding: 0;
  }
  .btn-text:hover {
    text-decoration: underline;
  }

  /* UTILS */
  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #e2e8f0;
    border-top-color: #10b981;
    border-radius: 50%;
    margin: 0 auto 16px;
    animation: spin 1s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .loading-state {
    text-align: center;
    padding: 60px;
    color: #94a3b8;
  }
  .empty-state {
    text-align: center;
    padding: 60px;
    color: #94a3b8;
  }
  .empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
    opacity: 0.5;
  }

  .animate-fade-in {
    opacity: 0;
    animation: fadeIn 0.6s ease-out forwards;
  }
  .animate-slide-up {
    opacity: 0;
    animation: slideUp 0.6s ease-out forwards;
  }
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .mb-6 {
    margin-bottom: 24px;
  }
  .mb-4 {
    margin-bottom: 16px;
  }
  .mt-4 {
    margin-top: 16px;
  }
</style>
