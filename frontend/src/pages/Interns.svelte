<script>
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import { api } from "../lib/api.js";
  import { portal } from "../lib/portal.js";
  import { auth } from "../lib/auth.svelte.js";
  import { getAvatarUrl } from "../lib/utils.js";
  // import  attach  from 'svelte';

  // State
  let interns = $state([]);
  let supervisors = $state([]);
  let loading = $state(false);
  let editing = $state(null); // holds intern object when editing
  let showEditModal = $state(false);
  let showCreateModal = $state(false);
  let searchQuery = $state("");
  let filterStatus = $state("");
  let currentPage = $state(1);
  let totalPages = $state(1);
  let totalItems = $state(0);
  let searchTimeout;
  let expandedInterns = $state({});

  // Keep overlay-root click-through state in sync with our modals (belt-and-suspenders).
  $effect(() => {
    const root =
      typeof document !== "undefined"
        ? document.querySelector("#overlay-root")
        : null;
    if (!(root instanceof HTMLElement)) return;
    const hasModal = showCreateModal || showEditModal;
    root.style.pointerEvents = hasModal ? "auto" : "none";
    if (!hasModal) {
      root.dataset.portalCount = "0";
    }
  });

  // Form State
  let form = $state({
    email: "",
    password: "",
    password_confirmation: "",
    full_name: "",
    institution: "",
    school: "",
    department: "",
    start_date: "",
    end_date: "",
    address: "",
    supervisor_id: "",
  });

  // Edit Form State
  let editForm = $state({
    email: "",
    password: "",
    password_confirmation: "",
    full_name: "",
    institution: "",
    school: "",
    department: "",
    start_date: "",
    end_date: "",
    address: "",
    supervisor_id: "",
    status: "",
  });

  const isSupervisor = $derived(
    auth.user?.role === "supervisor" || auth.user?.role === "pembimbing",
  );
  const isAdmin = $derived(auth.user?.role === "admin");

  // --- Fetch Data ---
  async function fetchInterns() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (isSupervisor && auth.user?.id) {
        params.supervisor_id = auth.user.id;
      }
      if (searchQuery) params.search = searchQuery;
      if (filterStatus) params.status = filterStatus;

      const res = await api.getInterns(params);
      interns = res.data || [];
      const pagination = res.pagination || {};

      totalPages = Math.max(pagination.total_pages || 1, 1);
      totalItems = pagination.total_items || 0;
      currentPage = pagination.page || currentPage;
    } catch (err) {
      console.error("Failed to fetch interns:", err);
      alert("Gagal memuat data intern: " + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchInterns();
  }
  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchInterns();
  }
  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchInterns();
    }, 500);
  }

  async function fetchSupervisors() {
    if (!isAdmin) return; // avoid 403 for supervisor users
    try {
      const res = await api.getSupervisors({ status: "active", limit: 200 });
      supervisors = res.data || [];
    } catch (err) {
      console.error("Failed to load supervisors", err);
      alert("Gagal memuat data pembimbing: " + err.message);
    }
  }

  // --- Create Intern ---
  async function createIntern() {
    if (!form.email || !form.password || !form.full_name || !form.school) {
      alert("Mohon lengkapi data wajib (Email, Password, Nama, Sekolah)");
      return;
    }

    if (form.password !== form.password_confirmation) {
      alert("Konfirmasi password tidak cocok");
      return;
    }

    // Convert supervisor_id to number or null
    const supervisorId = isSupervisor
      ? auth.user?.id
      : form.supervisor_id
        ? parseInt(form.supervisor_id, 10)
        : null;

    const payload = {
      email: form.email,
      password: form.password,
      full_name: form.full_name,
      school: form.school,
      department: form.department || "",
      start_date: form.start_date || "",
      end_date: form.end_date || "",
      address: form.address || "",
      supervisor_id: supervisorId,
    };

    try {
      const result = await api.createIntern(payload);
      alert("Berhasil menambah intern!");

      // Reset form setelah berhasil
      form = {
        email: "",
        password: "",
        password_confirmation: "",
        full_name: "",
        institution: "",
        school: "",
        department: "",
        start_date: "",
        end_date: "",
        address: "",
        supervisor_id: isSupervisor ? auth.user?.id : "",
      };

      showCreateModal = false;
      // Refresh tabel
      await fetchInterns();
    } catch (err) {
      console.error("Create intern error:", err);
      alert("Gagal membuat intern: " + err.message);
    }
  }

  // --- Update Intern (Edit) ---
  async function updateIntern() {
    if (!editing) return;

    if (
      editForm.password &&
      editForm.password !== editForm.password_confirmation
    ) {
      alert("Konfirmasi password baru tidak cocok");
      return;
    }

    // Convert supervisor_id to number or null
    const supervisorId = isSupervisor
      ? auth.user?.id
      : editForm.supervisor_id
        ? parseInt(editForm.supervisor_id, 10)
        : null;

    const payload = {
      full_name: editForm.full_name,
      school: editForm.school,
      department: editForm.department,
      address: editForm.address,
      start_date: editForm.start_date,
      end_date: editForm.end_date,
      supervisor_id: supervisorId,
      status: editForm.status,
    };

    if (editForm.email && editForm.email !== editing.email) {
      payload.email = editForm.email;
    }

    if (editForm.password && editForm.password.trim() !== "") {
      payload.password = editForm.password;
    }

    try {
      const result = await api.updateIntern(editing.id, payload);
      alert("Data intern berhasil diperbarui");
      closeEditModal();
      await fetchInterns();
    } catch (err) {
      console.error("Update intern error:", err);
      alert("Gagal memperbarui intern: " + err.message);
    }
  }

  function startEdit(intern) {
    editing = intern;
    editForm = {
      email: intern.email || "",
      password: "",
      password_confirmation: "",
      full_name: intern.full_name || "",
      school: intern.school || "",
      institution: intern.institution || "",
      department: intern.department || "",
      start_date: intern.start_date ? intern.start_date.slice(0, 10) : "",
      end_date: intern.end_date ? intern.end_date.slice(0, 10) : "",
      address: intern.address || "",
      supervisor_id: intern.supervisor_id || "",
      status: intern.status || "active",
    };
    showEditModal = true;
  }

  function closeEditModal() {
    showEditModal = false;
    editing = null;
  }

  function resetForm() {
    form = {
      email: "",
      password: "",
      password_confirmation: "",
      full_name: "",
      school: "",
      institution: "",
      department: "",
      start_date: "",
      end_date: "",
      address: "",
      supervisor_id: "",
    };
  }

  // --- Approve Intern ---
  async function handleApprove(id, name) {
    if (!confirm(`Setujui siswa "${name}" menjadi peserta Aktif?`)) return;

    try {
      await api.updateInternStatus(id, "active");

      // Update tampilan tabel secara langsung (tanpa reload page)
      const index = interns.findIndex((i) => i.id === id);
      if (index !== -1) {
        interns[index].status = "active";
      }

      alert(`Siswa ${name} telah disetujui`);
    } catch (err) {
      console.error("Approve error:", err);
      alert("Gagal melakukan approval: " + err.message);
    }
  }

  // --- Deny/Delete Intern ---
  async function handleDeny(id, name) {
    if (
      !confirm(
        `Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran "${name}"? Data akan hilang permanen.`,
      )
    )
      return;
    try {
      await api.deleteIntern(id);
      interns = interns.filter((i) => i.id !== id);
      alert(`Pendaftaran ${name} telah ditolak dan dihapus.`);
    } catch (err) {
      console.error("Delete error:", err);
      alert("Gagal menolak: " + err.message);
    }
  }

  function toggleExpand(id) {
    expandedInterns[id] = !expandedInterns[id];
  }

  onMount(async () => {
    await Promise.all([fetchInterns(), fetchSupervisors()]);
  });
</script>

<div class="page-container animate-fade-in">
  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Daftar Intern</h4>
    <span class="badge-count">{interns.length} dari {totalItems} Siswa</span>
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button
          class="cursor-pointer flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-slate-900 text-white hover:bg-slate-800 transition-all shadow-sm flex items-center justify-center gap-2"
          onclick={() => (showCreateModal = true)}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><line x1="12" y1="5" x2="12" y2="19"></line><line
              x1="5"
              y1="12"
              x2="19"
              y2="12"
            ></line></svg
          >
          <span>Tambah</span>
        </button>
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2"
          onclick={fetchInterns}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><path
              d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"
            /></svg
          >
          <span>Refresh</span>
        </button>
      </div>
      <div
        class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2 {totalPages <=
        1
          ? 'opacity-50 pointer-events-none'
          : ''}"
      >
        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {currentPage <=
          1
            ? 'opacity-50 cursor-not-allowed'
            : 'cursor-pointer'}"
          onclick={goToPreviousPage}
          disabled={currentPage <= 1}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M15 18l-6-6 6-6" /></svg
          >
          <span>Prev</span>
        </button>

        <div
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 pagination-pill"
        >
          <span>{currentPage}</span>
          <span class="text-slate-500">of</span>
          <span>{totalPages}</span>
        </div>

        <button
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 {currentPage >=
          totalPages
            ? 'opacity-50 cursor-not-allowed'
            : 'cursor-pointer'}"
          onclick={goToNextPage}
          disabled={currentPage >= totalPages}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M9 18l6-6-6-6" /></svg
          >
          <span>Next</span>
        </button>
      </div>
    </div>

    <div class="toolbar">
      <div class="search-wrapper">
        <span class="material-symbols-outlined search-icon">search</span>
        <input
          type="text"
          bind:value={searchQuery}
          oninput={handleSearchInput}
          onkeydown={(e) =>
            e.key === "Enter" && (clearTimeout(searchTimeout), fetchInterns())}
          placeholder="Cari Nama, Email, atau Sekolah..."
          class="search-input"
        />
      </div>

      <select
        bind:value={filterStatus}
        onchange={fetchInterns}
        class="filter-select"
      >
        <option value="">Semua Status</option>
        <option value="active">Aktif</option>
        <option value="completed">Selesai</option>
        <option value="cancelled">Cancelled</option>
        {#if isAdmin}
          <option value="pending">Pending</option>
        {/if}
      </select>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if interns.length === 0}
      <div class="empty-state">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#e5e7eb"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle
            cx="9"
            cy="7"
            r="4"
          ></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path
            d="M16 3.13a4 4 0 0 1 0 7.75"
          ></path></svg
        >
        <p>Belum ada data intern.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Nama</th>
              <th>Email</th>
              <th>Sekolah/Institusi</th>
              <th>Pembimbing</th>
              <th class="text-center">Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each interns as i}
              <tr
                class="table-row {i.status === 'pending'
                  ? 'bg-yellow-50/50 hover:bg-yellow-50'
                  : ''}"
              >
                <td>
                  <div class="user-info">
                    {#if i.avatar && getAvatarUrl(i.avatar)}
                      <img
                        src={getAvatarUrl(i.avatar)}
                        alt={i.full_name}
                        class="w-10 h-10 object-cover rounded-full shadow-sm"
                      />
                    {:else}
                      <div class="avatar-placeholder bg-slate-900">
                        {i.full_name ? i.full_name[0].toUpperCase() : "?"}
                      </div>
                    {/if}
                    <div class="user-details">
                      <span class="user-name">{i.full_name || "-"}</span>
                    </div>
                  </div>
                </td>
                <td class="text-muted">{i.email || "-"}</td>
                <td class="text-muted"
                  >{i.school || i.institution_name || "-"}</td
                >
                <td class="text-muted">{i.supervisor_name || "-"}</td>
                <td class="text-center">
                  <span class={`status-badge status-${i.status || "inactive"}`}>
                    {i.status || "inactive"}
                  </span>
                </td>
                <td class="text-right">
                  {#if i.status === "pending"}
                    <div class="flex items-center justify-end gap-2">
                      <button
                        class="btn-icon btn-deny flex-1"
                        onclick={() => handleDeny(i.id, i.full_name)}
                        title="Tolak & Hapus"
                      >
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="16"
                          height="16"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          ><line x1="18" y1="6" x2="6" y2="18"></line><line
                            x1="6"
                            y1="6"
                            x2="18"
                            y2="18"
                          ></line></svg
                        >
                        <span class="btn-label">Tolak</span>
                      </button>
                      <button
                        class="btn-icon btn-approve flex-1"
                        onclick={() => handleApprove(i.id, i.full_name)}
                        title="Setujui Siswa Ini"
                      >
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="16"
                          height="16"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          ><polyline points="20 6 9 17 4 12"></polyline></svg
                        >
                        <span class="btn-label">Terima</span>
                      </button>
                    </div>
                  {:else}
                    <a
                      class="btn-icon text-sky-600 hover:text-sky-700 bg-sky-50 hover:bg-sky-100"
                      href={`/interns/${i.id}/details`}
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
                    <button
                      class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                      onclick={() => startEdit(i)}
                      title="Edit Data"
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
                    <button
                      class="btn-icon text-slate-600 hover:text-slate-700 bg-slate-50 hover:bg-slate-100"
                      onclick={() => handleDeny(i.id, i.full_name)}
                      title="Hapus"
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
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mobile-list">
        {#each interns as i}
          <div
            class="entry-card {i.status === 'pending'
              ? 'bg-yellow-50 border-yellow-200'
              : ''}"
          >
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(i.id)}>
              <div class="user-info">
                {#if i.avatar && getAvatarUrl(i.avatar)}
                  <img
                    src={getAvatarUrl(i.avatar)}
                    alt={i.full_name}
                    class="w-10 h-10 object-cover rounded-full shadow-sm"
                  />
                {:else}
                  <div class="avatar-placeholder bg-slate-900">
                    {i.full_name ? i.full_name[0].toUpperCase() : "?"}
                  </div>
                {/if}
                <div class="user-details">
                  <div class="user-name">{i.full_name || "-"}</div>
                  <div class="text-muted small">{i.email || "-"}</div>
                </div>
              </div>
              <button class="expand-btn">
                <span
                  class="material-symbols-outlined transition-transform duration-200 {expandedInterns[
                    i.id
                  ]
                    ? 'rotate-180'
                    : ''}">expand_more</span
                >
              </button>
            </div>

            {#if expandedInterns[i.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                <div class="details-grid">
                  <div class="detail-box col-span-2 inline-box">
                    <div class="label">Status</div>
                    <span
                      class={`status-badge equal-badge ${i.status === "active" ? "bg-emerald-100 text-emerald-700" : i.status === "pending" ? "bg-yellow-100 text-yellow-700" : "bg-slate-100 text-slate-600"}`}
                    >
                      {i.status || "inactive"}
                    </span>
                  </div>
                  <div class="detail-box">
                    <div class="label">Sekolah</div>
                    <div class="value">
                      {i.school || i.institution_name || "-"}
                    </div>
                  </div>
                  <div class="detail-box">
                    <div class="label">Jurusan</div>
                    <div class="value">{i.department || "-"}</div>
                  </div>
                  <div class="detail-box">
                    <div class="label">Mulai</div>
                    <div class="value">
                      {i.start_date ? i.start_date.slice(0, 10) : "-"}
                    </div>
                  </div>
                  <div class="detail-box">
                    <div class="label">Selesai</div>
                    <div class="value">
                      {i.end_date ? i.end_date.slice(0, 10) : "-"}
                    </div>
                  </div>
                  <div class="detail-box col-span-2">
                    <div class="label">Pembimbing</div>
                    <div class="value">{i.supervisor_name || "-"}</div>
                  </div>
                </div>

                <div class="mobile-actions mt-4 pt-4 border-t border-slate-100">
                  {#if i.status === "pending"}
                    <button
                      class="mini-btn mobile danger"
                      onclick={(e) => {
                        e.stopPropagation();
                        handleDeny(i.id, i.full_name);
                      }}
                    >
                      <span class="material-symbols-outlined">close</span>
                      <span class="btn-text">Tolak</span>
                    </button>
                    <button
                      class="mini-btn mobile success"
                      onclick={(e) => {
                        e.stopPropagation();
                        handleApprove(i.id, i.full_name);
                      }}
                    >
                      <span class="material-symbols-outlined">check</span>
                      <span class="btn-text">Terima</span>
                    </button>
                  {:else}
                    <a
                      class="mini-btn mobile"
                      href={`/interns/${i.id}/details`}
                      onclick={(e) => e.stopPropagation()}
                    >
                      <span class="material-symbols-outlined">visibility</span>
                      <span class="btn-text">Detail</span>
                    </a>
                    <button
                      class="mini-btn-circle mobile"
                      onclick={(e) => {
                        e.stopPropagation();
                        startEdit(i);
                      }}
                    >
                      <span class="material-symbols-outlined">edit</span>
                      <span class="btn-text"></span>
                    </button>
                    <button
                      class="mini-btn mobile danger"
                      onclick={(e) => {
                        e.stopPropagation();
                        handleDeny(i.id, i.full_name);
                      }}
                    >
                      <span class="material-symbols-outlined">delete</span>
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

  <!-- CREATE MODAL -->
  {#if showCreateModal}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="fixed inset-0 z-120 flex items-center justify-center p-4 sm:p-6"
      use:portal
    >
      <div
        class="absolute inset-0 z-110 bg-slate-900/40 backdrop-blur-sm transition-opacity"
      ></div>
      <div
        class="relative bg-white z-120 rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden"
        onclick={(e) => e.stopPropagation()}
      >
        <div
          class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
        >
          <h3 class="font-bold text-lg text-slate-800">Tambah Intern Manual</h3>
          <button
            onclick={() => (showCreateModal = false)}
            class="text-slate-400 hover:text-slate-600"
          >
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>

        <div class="p-6 overflow-y-auto">
          <div class="form-grid">
            <div class="form-group">
              <label class="form-label" for="full_name"
                >Nama Lengkap <span class="text-red-500">*</span></label
              >
              <input
                class="input"
                bind:value={form.full_name}
                id="full_name"
                placeholder="Contoh: Budi Santoso"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="email"
                >Email <span class="text-red-500">*</span></label
              >
              <input
                class="input"
                type="email"
                bind:value={form.email}
                id="email"
                placeholder="email@sekolah.com"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="password"
                >Password <span class="text-red-500">*</span></label
              >
              <input
                class="input"
                type="password"
                bind:value={form.password}
                id="password"
                placeholder="Min. 6 karakter"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="password_confirmation"
                >Ulangi Password <span class="text-red-500">*</span></label
              >
              <input
                class="input"
                type="password"
                bind:value={form.password_confirmation}
                id="password_confirmation"
                placeholder="Ulangi password"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="school"
                >Sekolah <span class="text-red-500">*</span></label
              >
              <input
                class="input"
                bind:value={form.school}
                id="school"
                placeholder="Nama Sekolah"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="department">Jurusan</label>
              <input
                class="input"
                bind:value={form.department}
                id="department"
                placeholder="Jurusan"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="address">Alamat</label>
              <input
                class="input"
                bind:value={form.address}
                id="address"
                placeholder="Alamat domisili"
              />
            </div>
            {#if isAdmin}
              <div class="form-group">
                <label class="form-label" for="supervisor">Pembimbing</label>
                <select
                  class="input"
                  id="supervisor"
                  bind:value={form.supervisor_id}
                >
                  <option value="">Tidak ada</option>
                  {#each supervisors as s}
                    <option value={s.user_id}
                      >{s.full_name || s.name}{s.institution
                        ? ` - ${s.institution}`
                        : ""}</option
                    >
                  {/each}
                </select>
              </div>
            {/if}
            <div class="form-group">
              <label class="form-label" for="start_date">Mulai</label>
              <input
                class="input"
                type="date"
                bind:value={form.start_date}
                id="start_date"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="end_date">Selesai</label>
              <input
                class="input"
                type="date"
                bind:value={form.end_date}
                id="end_date"
              />
            </div>
          </div>
        </div>

        <div
          class="p-5 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3"
        >
          <button
            class="btn btn-outline"
            onclick={() => (showCreateModal = false)}>Batal</button
          >
          <button class="btn btn-primary" onclick={createIntern}
            >Simpan Data</button
          >
        </div>
      </div>
    </div>
  {/if}

  <!-- EDIT MODAL -->
  {#if showEditModal}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="fixed inset-0 z-120 flex items-center justify-center p-4 sm:p-6"
      use:portal
    >
      <div
        class="absolute inset-0 z-110 bg-slate-900/40 backdrop-blur-sm transition-opacity"
      ></div>
      <div
        class="relative bg-white z-120 rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden"
        onclick={(e) => e.stopPropagation()}
      >
        <div
          class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
        >
          <h3 class="font-bold text-lg text-slate-800">Edit Data Intern</h3>
          <button
            onclick={closeEditModal}
            class="text-slate-400 hover:text-slate-600"
          >
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>

        <div class="p-6 overflow-y-auto">
          <div class="form-grid">
            <div class="form-group">
              <label class="form-label" for="edit_full_name">Nama Lengkap</label
              >
              <input
                class="input"
                bind:value={editForm.full_name}
                id="edit_full_name"
                placeholder="Contoh: Budi Santoso"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_email">Email</label>
              <input
                class="input"
                type="email"
                bind:value={editForm.email}
                id="edit_email"
                placeholder="email@sekolah.com"
                disabled
              />
              <small class="text-xs text-gray-500 mt-1"
                >Email tidak dapat diubah</small
              >
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_status">Status</label>
              <select
                class="input"
                id="edit_status"
                bind:value={editForm.status}
              >
                <!-- <option value="pending">Pending</option> -->
                <option value="active">Aktif</option>
                <option value="completed">Selesai</option>
                <option value="cancelled">Dibatalkan</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_password"
                >Password Baru (Opsional)</label
              >
              <input
                class="input"
                type="password"
                bind:value={editForm.password}
                id="edit_password"
                placeholder="Kosongkan jika tidak ingin mengubah"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_password_confirmation"
                >Ulangi Password Baru</label
              >
              <input
                class="input"
                type="password"
                bind:value={editForm.password_confirmation}
                id="edit_password_confirmation"
                placeholder="Ulangi password baru"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_school">Sekolah</label>
              <input
                class="input"
                bind:value={editForm.school}
                id="edit_school"
                placeholder="Nama Sekolah"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_department">Jurusan</label>
              <input
                class="input"
                bind:value={editForm.department}
                id="edit_department"
                placeholder="Jurusan"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_address">Alamat</label>
              <input
                class="input"
                bind:value={editForm.address}
                id="edit_address"
                placeholder="Alamat domisili"
              />
            </div>
            {#if isAdmin}
              <div class="form-group">
                <label class="form-label" for="edit_supervisor"
                  >Pembimbing</label
                >
                <select
                  class="input"
                  id="edit_supervisor"
                  bind:value={editForm.supervisor_id}
                >
                  <option value="">Tidak ada</option>
                  {#each supervisors as s}
                    <option value={s.user_id}
                      >{s.full_name || s.name}{s.institution
                        ? ` - ${s.institution}`
                        : ""}</option
                    >
                  {/each}
                </select>
              </div>
            {/if}
            <div class="form-group">
              <label class="form-label" for="edit_start_date">Mulai</label>
              <input
                class="input"
                type="date"
                bind:value={editForm.start_date}
                id="edit_start_date"
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_end_date">Selesai</label>
              <input
                class="input"
                type="date"
                bind:value={editForm.end_date}
                id="edit_end_date"
              />
            </div>
          </div>
        </div>

        <div
          class="p-5 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3"
        >
          <button class="btn btn-outline" onclick={closeEditModal}>Batal</button
          >
          <button class="btn btn-primary" onclick={updateIntern}
            >Simpan Perubahan</button
          >
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Base Layout */

  .page-container {
    animation: fadeIn 0.5s ease-out;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    padding: 0 16px;
  }

  /* Avoid transforms on the page container so fixed modals cover the whole viewport */
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  /* Card Styles */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
    overflow: hidden;
  }

  .form-card {
    padding: 24px;
  }

  .table-card {
    padding: 0;
  }

  .card-header {
    margin-bottom: 20px;
    border-bottom: 1px solid #f3f4f6;
    padding-bottom: 15px;
  }

  .text-muted {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 4px 0 0 0;
  }

  /* Form Styles */
  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .form-label {
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
    margin-bottom: 6px;
  }

  .input {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid #e5e7eb;
    background-color: #f9fafb;
    border-radius: 8px;
    font-size: 0.925rem;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .input:hover {
    background-color: #fff;
    border-color: #d1d5db;
  }

  .input:focus {
    outline: none;
    background-color: #fff;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
  }

  .input:disabled {
    background-color: #f3f4f6;
    color: #9ca3af;
    cursor: not-allowed;
  }

  .form-actions {
    margin-top: 24px;
    text-align: right;
    border-top: 1px solid #f3f4f6;
    padding-top: 20px;
  }

  /* Buttons */
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 9px 18px;
    border-radius: 8px;
    cursor: pointer;
    border: none;
    font-weight: 600;
    font-size: 0.9rem;
    transition: all 0.2s;
  }

  /* .btn:active {
    transform: scale(0.98);
  } */

  .btn-primary {
    background: linear-gradient(135deg, #2563eb, #1d4ed8);
    color: white;
    box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
  }

  .btn-primary:hover {
    background: linear-gradient(135deg, #1d4ed8, #1e40af);
    box-shadow: 0 4px 6px rgba(37, 99, 235, 0.3);
  }

  .btn-outline {
    background: white;
    border: 1px solid #e5e7eb;
    color: #4b5563;
  }

  .btn-outline:hover {
    background-color: #f9fafb;
    border-color: #d1d5db;
    color: #111827;
  }

  .btn-approve {
    display: inline-flex;
    align-items: center;
    background-color: transparent;
    color: white;
    border: none;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 0.9rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
    /* box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2); */
  }

  .btn-approve:hover {
    color: #059669;
    /* box-shadow: 0 4px 6px rgba(16, 185, 129, 0.3); */
    /* transform: translateY(-1px); */
  }

  .btn-deny {
    display: inline-flex;
    align-items: center;
    background-color: transparent;
    color: #ef4444;
    border: 1px solid #ef4444;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 0.9rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-deny:hover {
    /* background-color: #ef4444; */
    color: df3434;
    /* box-shadow: 0 4px 6px rgba(239, 68, 68, 0.3); */
    /* transform: translateY(-1px); */
  }

  /* Table Header Row */
  .card-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #f3f4f6;
    background: rgba(248, 250, 252, 0.5);
  }

  .card-title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    display: inline-block;
  }
  @media (max-width: 900px) {
    .card-header-row {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }
    .toolbar {
      padding: 14px 16px;
    }
    .search-wrapper {
      flex: 1 1 100%;
    }
    .filter-select {
      width: 100%;
    }
  }

  .badge-count {
    background: #f1f5f9;
    color: #64748b;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 14px;
    font-weight: 600;
  }

  .sidebar,
  .topbar {
    z-index: -20;
    position: fixed;
  }

  /* Filters */
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    padding: 16px 24px;
    border-bottom: 1px solid #f3f4f6;
    background: #fafbfd;
  }
  .search-wrapper {
    flex: 1 1 320px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
    background: #fff;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }
  .search-icon {
    color: #9ca3af;
    font-variation-settings: "wght" 550;
  }
  .search-input {
    flex: 1;
    border: none;
    outline: none;
    font-size: 0.95rem;
    background: transparent;
    color: #111827;
  }
  .search-input::placeholder {
    color: #9ca3af;
  }

  .filter-select {
    min-width: 180px;
    border-radius: 10px;
    border: 1px solid #e5e7eb;
    padding: 10px 12px;
    background: #fff;
    font-weight: 600;
    color: #334155;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }

  /* Table Styles */
  .table-responsive {
    overflow-x: auto;
  }

  .table {
    width: 100%;
    min-width: 900px;
    border-collapse: collapse;
    font-size: 0.925rem;
  }
  .desktop-only {
    display: block;
  }
  .mobile-list {
    display: none;
  }

  .table th {
    text-align: left;
    padding: 14px 24px;
    background-color: #f8fafc;
    color: #64748b;
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid #e2e8f0;
  }

  .table td {
    padding: 14px 24px;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .table-row {
    transition: background-color 0.15s ease;
  }

  .table-row:hover {
    background-color: #f8fafc;
  }

  .table-row:last-child td {
    border-bottom: none;
  }

  /* User Info in Table */
  .user-info {
    display: grid;
    grid-template-columns: 40px 1fr;
    gap: 12px;
    align-items: center;
  }

  .avatar-placeholder {
    width: 32px;
    height: 32px;
    background: rgb(15 23 42);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.85rem;
    box-shadow: 0 2px 4px rgba(99, 102, 241, 0.2);
  }

  .user-name {
    font-weight: 600;
    color: #1f2937;
  }

  .user-name,
  .user-details .text-muted {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .text-placeholder {
    color: #d1d5db;
    font-size: 1.2rem;
  }
  .text-center {
    text-align: center;
  }
  .text-right {
    text-align: right;
    min-width: 150px;
    white-space: nowrap;
  }

  .pagination-pill {
    min-width: 128px;
  }

  /* Modal z-index helpers so overlays always sit above sidebar/topbar on mobile */
  :global(.z-120) {
    z-index: 120;
  }
  :global(.z-110) {
    z-index: 110;
  }
  :global(.z-100) {
    z-index: 100;
  }

  .action-buttons {
    display: inline-flex;
    gap: 8px;
    justify-content: flex-end;
  }
  /* Hide action button labels on desktop; show in mobile for clarity */
  .action-buttons span {
    display: none;
  }
  .action-buttons.responsive {
    flex-wrap: wrap;
  }
  .action-two {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    flex-wrap: wrap;
  }

  /* Status Badges */
  .status-badge {
    padding: 5px 12px;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    display: inline-block;
  }

  .status-active {
    background-color: #dcfce7;
    color: #15803d;
    border: 1px solid #bbf7d0;
  }
  .status-pending {
    background-color: #fef9c3;
    color: #a16207;
    border: 1px solid #fde047;
  }
  .status-inactive,
  .status-cancelled,
  .status-completed {
    background-color: #f3f4f6;
    color: #4b5563;
    border: 1px solid #e5e7eb;
  }
  .bg-emerald-100 {
    background: #ecfdf5;
    border-color: #a7f3d0;
  }
  .text-emerald-700 {
    color: #047857;
  }
  .bg-yellow-100 {
    background: #fefce8;
    border-color: #fef08a;
  }
  .text-yellow-700 {
    color: #a16207;
  }
  .bg-slate-100 {
    background: #f1f5f9;
    border-color: #e2e8f0;
  }
  .text-slate-600 {
    color: #475569;
  }
  .equal-badge {
    min-width: 96px;
    text-align: center;
    justify-content: center;
    display: inline-flex;
  }

  /* States */
  .loading-state {
    padding: 40px;
    text-align: center;
    color: #6b7280;
  }

  .empty-state {
    padding: 40px;
    text-align: center;
    color: #9ca3af;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 3px solid #e5e7eb;
    border-top: 3px solid #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 12px;
  }

  /* @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  } */
  .btn-icon {
    width: 42px;
    height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
  }
  .btn-label {
    display: none;
    font-weight: 700;
    font-size: 0.85rem;
  }

  .text-red-500 {
    color: #ef4444;
  }
  .text-xs {
    font-size: 0.75rem;
  }
  .text-gray-500 {
    color: #6b7280;
  }
  .mt-1 {
    margin-top: 0.25rem;
  }
  .flex {
    display: flex;
  }
  .gap-2 {
    gap: 0.5rem;
  }
  .gap-3 {
    gap: 0.75rem;
  }
  .items-center {
    align-items: center;
  }
  .border-b {
    border-bottom: 1px solid #f1f5f9;
  }

  @media (max-width: 900px) {
    .table-responsive {
      overflow: visible;
    }
    .table,
    .table thead {
      display: block;
    }
    .table tbody {
      display: grid;
      gap: 12px;
    }
    .table tr {
      display: grid;
      gap: 8px;
      padding: 14px;
      border: 1px solid #e5e7eb;
      border-radius: 12px;
      box-shadow: 0 2px 6px rgba(0, 0, 0, 0.03);
    }
    .table td,
    .table th {
      padding: 4px 0;
      border: none;
    }
    .text-right {
      text-align: left;
    }
    .action-buttons.responsive {
      width: 100%;
    }
    .action-buttons span,
    .btn-label {
      display: inline;
    }
    .user-info {
      align-items: flex-start;
    }
  }
  @media (max-width: 900px) {
    .desktop-only {
      display: none;
    }
    .mobile-list {
      display: flex;
      flex-direction: column;
      /* gap: 12px; */
    }
    .entry-card {
      padding: 14px;
      border-radius: 0px;
      border-top: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 6px 20px -18px rgba(15, 23, 42, 0.3);
    }
    .entry-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      cursor: pointer;
    }
    .entry-head .user-details {
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .entry-head .user-name {
      font-size: 0.95rem;
      font-weight: 600;
      color: #0f172a;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .entry-head .text-muted {
      font-size: 0.8rem;
      color: #64748b;
      margin: 0;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .entry-head .avatar-placeholder {
      width: 40px;
      height: 40px;
      font-size: 1rem;
      flex-shrink: 0;
      border-radius: 50%;
      overflow: hidden;
      background-color: #f1f5f9;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #64748b;
      font-weight: 600;
      position: relative; /* For img absolute positioning if needed, or just standard flow */
    }

    .entry-head .avatar-placeholder.bg-slate-900 {
      background-color: #0f172a;
      color: #ffffff;
    }

    .entry-head .avatar-placeholder.bg-slate-900 {
      background-color: #0f172a;
      color: #ffffff;
    }

    .entry-details {
      margin-top: 16px;
      padding-top: 16px;
      border-top: 1px solid #f1f5f9;
    }

    .entry-head .user-info {
      display: grid;
      grid-template-columns: 40px 1fr;
      gap: 12px;
      align-items: center;
      flex: 1;
      min-width: 0;
    }

    .detail-row {
      margin-bottom: 16px;
    }
    .detail-row:last-child {
      margin-bottom: 0;
    }
    .detail-label {
      font-size: 11px;
      font-weight: 700;
      color: #94a3b8;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 4px;
    }
    .detail-value {
      font-weight: 600;
      color: #0f172a;
      font-size: 14px;
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
    .mobile-actions {
      display: flex;
      gap: 10px;
    }
    .mini-btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 8px 16px;
      border-radius: 9999px;
      border: 1px solid #0f172a;
      background: #0f172a;
      color: #fff;
      font-weight: 700;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.15s ease;
      flex: 1;
      justify-content: center;
    }

    .mini-btn-circle {
      display: inline-flex;
      align-items: center;
      /* gap: 6px; */
      /* padding: 8px 16px; */
      border-radius: 9999px;
      border: 1px solid #0f172a;
      background: #0f172a;
      color: #fff;
      font-weight: 700;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.15s ease;
      width: 42px;
      height: 42px;
      /* flex: 1; */
      justify-content: center;
    }
    .mini-btn .btn-text {
      display: inline;
    }
    .mini-btn.success {
      background: #10b981;
      border-color: #10b981;
    }
    .mini-btn.danger {
      background: #ef4444;
      border-color: #ef4444;
    }
  }

  /* Animations */
  .animate-fade-in {
    opacity: 0;
    animation: fadeIn 0.6s ease-out forwards;
  }
  .animate-slide-up {
    opacity: 0;
    animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
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

  /* Action Buttons - Standardized */
  .btn-icon {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .btn-icon:hover {
    background: #e2e8f0;
    color: #0f172a;
  }
  /* Mobile Grid Styles */
  @media (max-width: 900px) {
    .details-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      margin-bottom: 12px;
    }
    .detail-box {
      padding: 12px;
      border: 1px solid #e2e8f0;
      border-radius: 14px;
      background: #f8fafc;
      text-align: center;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
    }
    .detail-box.col-span-2 {
      grid-column: span 2 / span 2;
    }
    .detail-box.inline-box {
      flex-direction: row;
      gap: 8px;
    }
    .detail-box.inline-box .label {
      margin: 0;
    }
    .detail-box .label {
      margin: 0 0 6px 0;
      font-size: 11px;
      font-weight: 700;
      color: #94a3b8;
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }
    .detail-box .value {
      font-size: 14px;
      font-weight: 600;
      color: #0f172a;
    }
  }
</style>
