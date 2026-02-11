<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { api } from "../lib/api.js";
  import { auth } from "../lib/auth.svelte.js";
  import AssessmentCreateModal from "./AssessmentCreateModal.svelte";
  import AssessmentEditModal from "./AssessmentEditModal.svelte";

  // State
  let assessments = $state([]);
  let loading = $state(true);
  let isModalOpen = $state(false);
  let isEditModalOpen = $state(false);
  let selectedAssessment = $state(null);
  let expandedAssessments = $state({});
  let activeTab = $state("supervisor"); // 'supervisor' or 'admin'
  let searchQuery = $state("");
  let searchTimeout;

  let supervisorAssessments = $derived(
    assessments.filter(
      (a) =>
        a.assessor_role === "supervisor" || a.assessor_role === "pembimbing",
    ),
  );
  let adminAssessments = $derived(
    assessments.filter((a) => a.assessor_role === "admin"),
  );

  async function fetchAssessments() {
    loading = true;
    try {
      const params = { page: 1, limit: 50 };
      if (searchQuery) params.search = searchQuery;

      const res = await api.getAssessments(params);
      assessments = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      fetchAssessments();
    }, 500);
  }

  function toggleExpand(id) {
    expandedAssessments[id] = !expandedAssessments[id];
  }

  // Helpers
  function calculateAverage(a) {
    const total =
      a.quality_score +
      a.speed_score +
      a.initiative_score +
      a.teamwork_score +
      a.communication_score;
    return (total / 5).toFixed(1);
  }

  function getScoreColor(score) {
    if (score >= 90)
      return "bg-emerald-100 text-emerald-700 border-emerald-200";
    if (score >= 80) return "bg-blue-100 text-blue-700 border-blue-200";
    if (score >= 70) return "bg-yellow-100 text-yellow-700 border-yellow-200";
    return "bg-red-100 text-red-700 border-red-200";
  }

  function formatDate(dateStr) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
    });
  }

  async function handleDelete(id, name) {
    if (
      !confirm(
        `Apakah Anda yakin ingin menghapus penilaian untuk "${name || "Intern"}"?`,
      )
    )
      return;
    try {
      await api.deleteAssessment(id);
      assessments = assessments.filter((a) => a.id !== id);
      alert("Penilaian berhasil dihapus.");
    } catch (err) {
      console.error(err);
      alert("Gagal menghapus penilaian: " + err.message);
    }
  }

  function openEditModal(assessment) {
    selectedAssessment = assessment;
    isEditModalOpen = true;
  }

  onMount(fetchAssessments);
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    <!-- Header -->
    <div class="header">
      <div class="header-content">
        <h2 class="title">Evaluasi Kinerja</h2>
        <p class="subtitle">
          Kelola penilaian kualitas dan perkembangan peserta magang.
        </p>
      </div>

      {#if auth.user?.role !== "intern"}
        <button class="btn-primary" onclick={() => (isModalOpen = true)}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          Buat Penilaian
        </button>
      {/if}
    </div>

    <!-- TABEL DAFTAR PENILAIAN -->
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">
      {#if auth.user?.role === "intern"}
        <div class="card-header border-b">
          <div class="tabs">
            <button
              class="tab-btn {activeTab === 'supervisor' ? 'active' : ''}"
              onclick={() => (activeTab = "supervisor")}
            >
              Penilaian Pembimbing
              <span class="badge-count">{supervisorAssessments.length}</span>
            </button>
            <button
              class="tab-btn {activeTab === 'admin' ? 'active' : ''}"
              onclick={() => (activeTab = "admin")}
            >
              Penilaian Admin
              <span class="badge-count">{adminAssessments.length}</span>
            </button>
          </div>
        </div>
      {:else}
        <div class="card-header border-b">
          <h3>Daftar Penilaian Anda</h3>
        </div>
      {/if}

      <div class="toolbar">
        <div class="search-wrapper">
          <span class="material-symbols-outlined search-icon">search</span>
          <input
            type="text"
            bind:value={searchQuery}
            oninput={handleSearchInput}
            onkeydown={(e) =>
              e.key === "Enter" &&
              (clearTimeout(searchTimeout), fetchAssessments())}
            placeholder="Cari Nama Intern atau Judul Tugas..."
            class="search-input"
          />
        </div>
      </div>

      {#if loading}
        <div class="loading-state">Memuat data penilaian...</div>
      {:else if (auth.user?.role === "intern" ? (activeTab === "supervisor" ? supervisorAssessments : adminAssessments) : assessments).length === 0}
        <div class="empty-state">
          <div class="empty-icon">📝</div>
          <p>
            Belum ada data penilaian {auth.user?.role === "intern"
              ? activeTab === "supervisor"
                ? "pembimbing"
                : "admin"
              : "yang Anda buat"}.
          </p>
        </div>
      {:else}
        <!-- DESKTOP TABLE -->
        <div class="table-container desktop-only">
          <table class="table">
            <thead>
              <tr>
                <th>Siswa</th>
                <th>Tugas</th>
                <th style="width: 15%">Kualitas</th>
                <th style="width: 15%">Kecepatan</th>
                <th style="width: 15%">Inisiatif</th>
                <th class="text-center">Rata-rata</th>
                <th class="text-center">Grade</th>
                <th class="text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#each auth.user?.role === "intern" ? (activeTab === "supervisor" ? supervisorAssessments : adminAssessments) : assessments as item}
                {@const avg = Number(calculateAverage(item))}
                {@const grade =
                  avg >= 85 ? "A" : avg >= 75 ? "B" : avg >= 60 ? "C" : "D"}
                {@const gradeClass =
                  avg >= 85
                    ? "grade-a"
                    : avg >= 75
                      ? "grade-b"
                      : avg >= 60
                        ? "grade-c"
                        : "grade-d"}
                <tr class="hover-row">
                  <td>
                    <div class="user-info">
                      <div class="avatar">
                        {item.intern_name?.charAt(0) || "U"}
                      </div>
                      <div class="user-text">
                        <span class="name"
                          >{item.intern_name || "Tidak Diketahui"}</span
                        >
                      </div>
                    </div>
                  </td>
                  <td class="text-cell">
                    <div class="task-info-container">
                      <span class="task-title"
                        >{item.task_title || "Penilaian Umum"}</span
                      >
                    </div>
                    <div class="date-sub">
                      {formatDate(item.assessment_date || item.created_at)}
                    </div>
                  </td>

                  <td>
                    <div class="bar-container">
                      <div class="progress-track-small">
                        <div
                          class="progress-bar-small bg-indigo-500"
                          style="width: {item.quality_score}%"
                        ></div>
                      </div>
                      <span class="score-text">{item.quality_score}</span>
                    </div>
                  </td>
                  <td>
                    <div class="bar-container">
                      <div class="progress-track-small">
                        <div
                          class="progress-bar-small bg-emerald-500"
                          style="width: {item.speed_score}%"
                        ></div>
                      </div>
                      <span class="score-text">{item.speed_score}</span>
                    </div>
                  </td>
                  <td>
                    <div class="bar-container">
                      <div class="progress-track-small">
                        <div
                          class="progress-bar-small bg-amber-500"
                          style="width: {item.initiative_score}%"
                        ></div>
                      </div>
                      <span class="score-text">{item.initiative_score}</span>
                    </div>
                  </td>

                  <td class="text-center">
                    <span class="avg-score">{avg}</span>
                  </td>
                  <td class="text-center">
                    <span class={`grade-badge ${gradeClass}`}>{grade}</span>
                  </td>

                  <td class="text-right">
                    <div class="actions">
                      <a
                        href={`/assessments/${item.id}`}
                        class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                        title="Lihat Detail"
                      >
                        <svg
                          width="18"
                          height="18"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                          ></path>
                          <circle cx="12" cy="12" r="3"></circle>
                        </svg>
                      </a>
                      {#if auth.user && ["admin", "supervisor"].includes(auth.user.role)}
                        <button
                          class="btn-icon text-amber-600 hover:text-amber-700 bg-amber-50 hover:bg-amber-100"
                          onclick={() => openEditModal(item)}
                          title="Edit Penilaian"
                        >
                          <svg
                            width="18"
                            height="18"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <path
                              d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                            ></path>
                            <path
                              d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                            ></path>
                          </svg>
                        </button>
                      {/if}
                      {#if auth.user?.role !== "intern"}
                        <button
                          class="btn-icon text-rose-600 hover:text-rose-700 bg-rose-50 hover:bg-rose-100"
                          onclick={() =>
                            handleDelete(item.id, item.intern_name)}
                          title="Hapus Penilaian"
                        >
                          <svg
                            width="18"
                            height="18"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path
                              d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                            ></path>
                          </svg>
                        </button>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- MOBILE LIST -->
        <div class="mobile-list">
          {#each auth.user?.role === "intern" ? (activeTab === "supervisor" ? supervisorAssessments : adminAssessments) : assessments as item}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-card" onclick={() => toggleExpand(item.id)}>
              <div class="entry-head">
                <div class="user-info mobile">
                  <div class="avatar">{item.intern_name?.charAt(0) || "U"}</div>
                  <div class="user-details">
                    <div class="user-name">{item.intern_name || "Intern"}</div>
                    <div class="date-badge-mobile">
                      {formatDate(item.assessment_date || item.created_at)}
                    </div>
                  </div>
                </div>
                <button class="expand-btn">
                  <span
                    class="material-symbols-outlined transition-transform {expandedAssessments[
                      item.id
                    ]
                      ? 'rotate-180'
                      : ''}">expand_more</span
                  >
                </button>
              </div>

              {#if expandedAssessments[item.id]}
                <div class="entry-details" transition:slide={{ duration: 200 }}>
                  <div class="detail-row">
                    <div class="detail-label">Tipe</div>
                    <div class="detail-value">
                      {#if item.task_title}
                        <span class="type-badge-pill task">Per Tugas</span>
                      {:else}
                        <span class="type-badge-pill general">Umum</span>
                      {/if}
                    </div>
                  </div>
                  <div class="detail-row">
                    <div class="detail-label">Tugas</div>
                    <div class="detail-value text-slate-800">
                      {item.task_title || "-"}
                    </div>
                  </div>
                  <div class="detail-row">
                    <div class="detail-label">Penilai</div>
                    <div class="detail-value text-slate-800">
                      {item.assessor_name || "-"}
                      <span class="text-xs text-slate-500"
                        >({item.assessor_role})</span
                      >
                    </div>
                  </div>

                  <!-- Bars for Mobile -->
                  <div
                    class="detail-row"
                    style="display:block; margin-bottom: 4px;"
                  >
                    <div
                      style="display:flex; justify-content:space-between; margin-bottom:4px; font-size:13px;"
                    >
                      <span class="text-slate-500">Kualitas</span>
                      <span class="font-semibold text-slate-700"
                        >{item.quality_score}</span
                      >
                    </div>
                    <div class="progress-track-small" style="height: 4px;">
                      <div
                        class="progress-bar-small bg-indigo-500"
                        style="width: {item.quality_score}%"
                      ></div>
                    </div>
                  </div>
                  <div
                    class="detail-row"
                    style="display:block; margin-bottom: 4px;"
                  >
                    <div
                      style="display:flex; justify-content:space-between; margin-bottom:4px; font-size:13px;"
                    >
                      <span class="text-slate-500">Kecepatan</span>
                      <span class="font-semibold text-slate-700"
                        >{item.speed_score}</span
                      >
                    </div>
                    <div class="progress-track-small" style="height: 4px;">
                      <div
                        class="progress-bar-small bg-emerald-500"
                        style="width: {item.speed_score}%"
                      ></div>
                    </div>
                  </div>
                  <div
                    class="detail-row"
                    style="display:block; margin-bottom: 4px;"
                  >
                    <div
                      style="display:flex; justify-content:space-between; margin-bottom:4px; font-size:13px;"
                    >
                      <span class="text-slate-500">Inisiatif</span>
                      <span class="font-semibold text-slate-700"
                        >{item.initiative_score}</span
                      >
                    </div>
                    <div class="progress-track-small" style="height: 4px;">
                      <div
                        class="progress-bar-small bg-amber-500"
                        style="width: {item.initiative_score}%"
                      ></div>
                    </div>
                  </div>

                  <div class="detail-row mt-2">
                    <div class="detail-label">Rata-rata</div>
                    <span
                      class={`score-badge ${getScoreColor(calculateAverage(item))}`}
                    >
                      {calculateAverage(item)}
                    </span>
                  </div>

                  <div
                    class="mobile-actions mt-4 pt-4 border-t border-slate-100"
                  >
                    <a
                      href={`/assessments/${item.id}`}
                      class="mini-btn mobile text-sky-600 bg-sky-50 border-sky-100"
                    >
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                        ></path>
                        <circle cx="12" cy="12" r="3"></circle>
                      </svg>
                      <span class="btn-text">Detail</span>
                    </a>

                    {#if auth.user && ["admin", "supervisor"].includes(auth.user.role)}
                      <button
                        class="mini-btn mobile"
                        onclick={(e) => {
                          e.stopPropagation();
                          openEditModal(item);
                        }}
                      >
                        <svg
                          width="18"
                          height="18"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path
                            d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                          ></path>
                          <path
                            d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                          ></path>
                        </svg>
                        <span class="btn-text">Edit</span>
                      </button>
                    {/if}
                    {#if auth.user?.role !== "intern"}
                      <button
                        class="mini-btn mobile danger"
                        onclick={(e) => {
                          e.stopPropagation();
                          handleDelete(item.id, item.intern_name);
                        }}
                      >
                        <svg
                          width="18"
                          height="18"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path
                            d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                          ></path>
                        </svg>
                        <span class="btn-text">Hapus</span>
                      </button>
                    {/if}
                  </div>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <AssessmentCreateModal
    isOpen={isModalOpen}
    onClose={() => (isModalOpen = false)}
    onSuccess={fetchAssessments}
  />

  <AssessmentEditModal
    isOpen={isEditModalOpen}
    assessment={selectedAssessment}
    onClose={() => (isEditModalOpen = false)}
    onSuccess={fetchAssessments}
  />
</div>

<style>
  :root {
    --space-xs: 8px;
    --space-sm: 12px;
    --space-md: 16px;
    --space-lg: 24px;

    --avatar-size: 32px;
  }

  :global(body) {
    font-family: "Geist", "Inter", sans-serif;
    color: #0f172a;
  }

  .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    /* background-image: radial-gradient(
        at 0% 0%,
        rgba(16, 185, 129, 0.03) 0%,
        transparent 50%
      ),
      radial-gradient(
        at 100% 100%,
        rgba(14, 165, 233, 0.03) 0%,
        transparent 50%
      ); */
    padding: 0;
    overflow-x: hidden; /* Check for overflow */
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
  }

  /* --- HEADER --- */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 16px;
  }
  .title {
    font-size: 20px;
    font-weight: 600;
    color: #0f172a;
    margin: 0 0 0 0;
    letter-spacing: -0.02em;
  }
  .subtitle {
    color: #64748b;
    font-size: 16px;
    margin: 0;
  }

  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669);
    color: white;
    padding: 10px 20px;
    border-radius: 999px;
    font-weight: 600;
    font-size: 14px;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2);
    transition: all 0.2s;
  }
  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.3);
  }

  /* --- CARD --- */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.02),
      0 2px 4px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
    margin-bottom: 32px;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }

  /* --- TOOLBAR --- */
  .toolbar {
    padding: 0;
    padding-bottom: 16px;
    background: #fcfcfc;
    border-bottom: 1px solid #f1f5f9;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
  }
  .search-wrapper {
    position: relative;
    flex: 1;
    min-width: 280px;
  }
  .search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: #94a3b8;
    font-size: 20px;
  }
  .search-input {
    width: 100%;
    padding: 10px 12px 10px 40px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    font-size: 14px;
    transition: all 0.2s;
  }
  .search-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  .filter-select {
    padding: 10px 16px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    font-size: 14px;
    color: #475569;
    background: white;
    cursor: pointer;
    min-width: 140px;
  }

  .card-header {
    padding: 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .card-header h3 {
    margin: 20px 24px;
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
  }
  .tabs {
    display: flex;
    width: 100%;
    border-bottom: 1px solid #f1f5f9;
  }
  .tab-btn {
    flex: 1;
    text-align: center;
    padding: 16px;
    font-size: 14px;
    font-weight: 600;
    color: #64748b;
    cursor: pointer;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  .tab-btn:hover {
    background: #f8fafc;
    color: #0f172a;
  }
  .tab-btn.active {
    color: #10b981;
    border-bottom-color: #10b981;
    background: #f0fdf4;
  }
  .border-b {
    border-bottom: none; /* override default border b */
  }

  /* --- TABLE --- */
  .badge-count {
    background: #f1f5f9;
    color: #64748b;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
  }

  .table-container {
    overflow-x: auto;
  }
  .table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
  }

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
    padding: var(--space-md) var(--space-lg);
    height: 64px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
    color: #334155;
  }
  .table tr:last-child td {
    border-bottom: none;
  }
  .hover-row:hover td {
    background-color: #f8fafc;
  }

  /* Table Content */
  .report-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .report-title {
    font-weight: 600;
    color: #0f172a;
    font-size: 14px;
  }

  .user-info
  /* .user-info.mobile { */ {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    min-width: 0;
  }

  .user-info.mobile {
    display: flex;
    flex-direction: row; /* critical */
    align-items: center;
    gap: 12px;
  }

  .avatar {
    width: var(--avatar-size);
    height: var(--avatar-size);
    line-height: var(--avatar-size);
    background: #0f172a;
    color: white;
    border-radius: 999px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 12px;
    flex-shrink: 0;
  }
  .name {
    font-weight: 500;
    font-size: 14px;
  }

  .name,
  .user-name,
  .report-title,
  .report-title-mobile {
    line-height: 1.2;
  }

  .text-right {
    text-align: right;
  }
  .text-slate-500 {
    color: #64748b;
  }
  .text-sm {
    font-size: 13px;
  }

  /* Badges */
  .type-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    width: fit-content;
    border: 1px solid transparent;
  }
  .bg-blue-100 {
    background: #eff6ff;
    border-color: #bfdbfe;
  }
  .bg-purple-100 {
    background: #faf5ff;
    border-color: #e9d5ff;
  }
  .bg-rose-100 {
    background: #fff1f2;
    border-color: #fecdd3;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    padding: 4px 12px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 600;
    border: 1px solid transparent;
    text-transform: capitalize;
  }
  .bg-emerald-100 {
    background: #ecfdf5;
    border-color: #a7f3d0;
  }
  .bg-amber-100 {
    background: #fefce8;
    border-color: #fef08a;
  }
  .bg-red-100 {
    background: #fef2f2;
    border-color: #fecaca;
  }
  .bg-slate-100 {
    background: #f1f5f9;
    border-color: #e2e8f0;
  }

  .btn-icon {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
  }
  .btn-icon:hover {
    background: #e2e8f0;
    color: #0f172a;
  }

  /* States */
  .empty-state,
  .loading-state {
    text-align: center;
    padding: 60px 20px;
    color: #94a3b8;
    font-style: italic;
  }
  .empty-icon {
    font-size: 32px;
    margin-bottom: 12px;
    opacity: 0.5;
  }

  /* Mobile Styles */
  .mobile-list {
    display: none;
  }
  .desktop-only {
    display: block;
  }

  @media (max-width: 900px) {
    .page-bg {
      padding: 0px;
    }
    .header {
      flex-direction: column;
      align-items: stretch;
      gap: 16px;
    }
    .btn-primary {
      width: 100%;
      justify-content: center;
      margin-top: 8px;
    }

    .desktop-only {
      display: none;
    }
    .mobile-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      padding: 0;
    }

    .card.list-card {
      background: transparent;
      border: none;
      box-shadow: none;
    }
    .card.list-card .card-header {
      display: none;
    } /* Optional: hide header on mobile if desired */

    .list-card {
      padding: 0px;
    }

    .entry-card {
      padding: var(--space-md);
      border-radius: 16px;
      border: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    }

    .entry-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      cursor: pointer;
      min-height: 64px;
    }

    .user-details {
      display: flex;
      flex-direction: column;
      justify-content: center;
    }

    *,
    *::before,
    *::after {
      box-sizing: border-box;
    }

    .user-name {
      font-weight: 600;
      font-size: 14px;
      color: #0f172a;
    }
    .report-title-mobile {
      font-size: 13px;
      color: #64748b;
    }

    .expand-btn {
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      background: #f8fafc;
      color: #64748b;
      border: none;
    }

    .entry-details {
      margin-top: 16px;
      padding-top: 16px;
      border-top: 1px solid #f1f5f9;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .detail-row {
      display: flex;
      justify-content: space-between;
      font-size: 14px;
    }
    .detail-label {
      color: #64748b;
    }
    .detail-value {
      font-weight: 500;
      color: #0f172a;
    }

    .mini-btn {
      padding: 6px 12px;
      border-radius: 8px;
      font-size: 13px;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 6px;
      border: 1px solid #e2e8f0;
      background: white;
      color: #475569;
    }
    .mini-btn.mobile {
      flex: 1;
      justify-content: center;
    }
    .mobile-actions {
      display: flex;
      gap: 12px;
    }
    .mini-btn.danger {
      color: #e11d48;
      border-color: #fecdd3;
      background: #fff1f2;
    }
  }

  /* NEW STYLES for Refined UI */
  .task-title {
    display: block;
    font-weight: 600;
    color: #334155;
    font-size: 14px;
    margin-bottom: 2px;
  }
  .date-sub {
    font-size: 12px;
    color: #94a3b8;
  }

  .bar-container {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .progress-track-small {
    flex: 1;
    height: 6px;
    background: #f1f5f9;
    border-radius: 3px;
    overflow: hidden;
    min-width: 60px;
  }
  .progress-bar-small {
    height: 100%;
    border-radius: 3px;
  }
  .score-text {
    font-size: 12px;
    font-weight: 600;
    color: #64748b;
    width: 24px;
    text-align: right;
  }

  .task-info-container {
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }

  .type-badge-pill {
    padding: 2px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.02em;
    display: inline-flex;
    align-items: center;
  }
  .type-badge-pill.task {
    background: #f0fdf4;
    color: #10b981;
    border: 1px solid #d1fae5;
  }
  .type-badge-pill.general {
    background: #f8fafc;
    color: #64748b;
    border: 1px solid #e2e8f0;
  }

  .avg-score {
    font-weight: 800;
    font-size: 16px;
    color: #0f172a;
  }

  .grade-badge {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 14px;
    margin: 0 auto;
  }
  .grade-a {
    background: #f0fdf4;
    color: #15803d;
    border: 1px solid #bbf7d0;
  }
  .grade-b {
    background: #eff6ff;
    color: #1d4ed8;
    border: 1px solid #bfdbfe;
  }
  .grade-c {
    background: #fffbeb;
    color: #b45309;
    border: 1px solid #fde68a;
  }
  .grade-d {
    background: #fff1f2;
    color: #be123c;
    border: 1px solid #fecdd3;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
</style>
