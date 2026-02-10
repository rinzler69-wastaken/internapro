<script>
  import { onMount } from 'svelte';
  import { slide } from 'svelte/transition';
  import { api } from '../lib/api.js';
  import { portal } from '../lib/portal.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let supervisors = $state([]);
  let loading = $state(false);
  let editing = $state(null);
  let showEditModal = $state(false);
  let showCreateModal = $state(false);
  let searchQuery = $state('');
  let filterStatus = $state('');
  let currentPage = $state(1);
  let totalPages = $state(1);
  let searchTimeout;
  let expandedSupervisors = $state({});

  // Keep overlay-root click-through state in sync with our modals
  $effect(() => {
    const root = typeof document !== 'undefined' ? document.querySelector('#overlay-root') : null;
    if (!root) return;
    const hasModal = showCreateModal || showEditModal;
    root.style.pointerEvents = hasModal ? 'auto' : 'none';
    if (!hasModal) {
      root.dataset.portalCount = '0';
    }
  });

  // Form State
  let form = $state({
    name: '',
    email: '',
    password: '',
    password_confirmation: '',
    nip: '',
    phone: '',
    position: '',
    address: '',
    institution: '',
    status: 'active',
  });

  // Edit Form State
  let editForm = $state({
    name: '',
    email: '',
    password: '',
    password_confirmation: '',
    nip: '',
    phone: '',
    position: '',
    address: '',
    institution: '',
    status: '',
  });

  // --- Fetch Data ---
  async function fetchSupervisors() {
    loading = true;
    try {
      const params = { page: currentPage, limit: 50 };
      if (searchQuery) params.search = searchQuery;
      if (filterStatus) params.status = filterStatus;

      const res = await api.getSupervisors(params);
      supervisors = res.data || [];
      const pagination = res.pagination || {};
      totalPages = Math.max(pagination.total_pages || 1, 1);
      currentPage = pagination.page || currentPage;
      console.log('Fetched supervisors:', supervisors);
    } catch (err) {
      console.error('Failed to fetch supervisors:', err);
      alert('Gagal memuat data pembimbing: ' + err.message);
    } finally {
      loading = false;
    }
  }

  function goToPreviousPage() {
    currentPage = Math.max(1, currentPage - 1);
    fetchSupervisors();
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;
    currentPage += 1;
    fetchSupervisors();
  }

  function handleSearchInput() {
    clearTimeout(searchTimeout);
    currentPage = 1;
    searchTimeout = setTimeout(() => {
      fetchSupervisors();
    }, 500);
  }

  // --- Create Supervisor ---
  async function createSupervisor() {
    if (!form.name || !form.email || !form.password) {
      alert('Mohon lengkapi data wajib (Nama, Email, Password)');
      return;
    }

    if (form.password !== form.password_confirmation) {
      alert('Konfirmasi password tidak cocok');
      return;
    }

    const payload = {
      name: form.name,
      email: form.email,
      password: form.password,
      nip: form.nip || '',
      phone: form.phone || '',
      position: form.position || '',
      address: form.address || '',
      institution: form.institution || '',
      status: form.status || 'active',
    };

    console.log('Creating supervisor with payload:', payload);

    try {
      const result = await api.createSupervisor(payload);
      console.log('Create result:', result);
      alert('Berhasil menambah pembimbing!');
      
      // Reset form
      form = {
        name: '',
        email: '',
        password: '',
        password_confirmation: '',
        nip: '',
        phone: '',
        position: '',
        address: '',
        institution: '',
        status: 'active',
      };
      
      showCreateModal = false;
      await fetchSupervisors();
    } catch (err) {
      console.error('Create supervisor error:', err);
      alert('Gagal membuat pembimbing: ' + err.message);
    }
  }

  // --- Update Supervisor ---
  async function updateSupervisor() {
    if (!editing) return;
    
    if (editForm.password && editForm.password !== editForm.password_confirmation) {
      alert('Konfirmasi password baru tidak cocok');
      return;
    }

    const payload = {
      name: editForm.name,
      nip: editForm.nip,
      phone: editForm.phone,
      position: editForm.position,
      address: editForm.address,
      institution: editForm.institution,
      status: editForm.status,
    };

    if (editForm.email && editForm.email !== editing.email) {
      payload.email = editForm.email;
    }

    if (editForm.password && editForm.password.trim() !== '') {
      payload.password = editForm.password;
    }

    console.log('Updating supervisor with payload:', payload);

    try {
      const result = await api.updateSupervisor(editing.id, payload);
      console.log('Update result:', result);
      alert('Data pembimbing berhasil diperbarui');
      closeEditModal();
      await fetchSupervisors();
    } catch (err) {
      console.error('Update supervisor error:', err);
      alert('Gagal memperbarui pembimbing: ' + err.message);
    }
  }

  function startEdit(supervisor) {
    editing = supervisor;
    console.log('Editing supervisor:', supervisor);
    editForm = {
      name: supervisor.full_name || supervisor.name || '',
      email: supervisor.email || '',
      password: '',
      password_confirmation: '',
      nip: supervisor.nip || '',
      phone: supervisor.phone || '',
      position: supervisor.position || '',
      address: supervisor.address || '',
      institution: supervisor.institution || '',
      status: supervisor.status || 'active',
    };
    showEditModal = true;
  }

  function closeEditModal() {
    showEditModal = false;
    editing = null;
  }

  function resetForm() {
    form = {
      name: '',
      email: '',
      password: '',
      password_confirmation: '',
      nip: '',
      phone: '',
      position: '',
      address: '',
      institution: '',
      status: 'active',
    };
  }

  // --- Approve Supervisor ---
  async function handleApprove(id, name) {
    if (!confirm(`Setujui pembimbing "${name}" menjadi Aktif?`)) return;

    try {
      await api.approveSupervisor(id);
      
      // Update tampilan tabel secara langsung
      const index = supervisors.findIndex(s => s.id === id);
      if (index !== -1) {
        supervisors[index].status = 'active';
      }
      
      alert(`Pembimbing ${name} telah disetujui`);
    } catch (err) {
      console.error('Approve error:', err);
      alert('Gagal melakukan approval: ' + err.message);
    }
  }

  // --- Deny/Delete Supervisor ---
  async function handleDeny(id, name) {
    if (!confirm(`Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pembimbing "${name}"? Data akan hilang permanen.`)) return;
    try {
      await api.rejectSupervisor(id);
      supervisors = supervisors.filter(s => s.id !== id);
      alert(`Pembimbing ${name} telah ditolak dan dihapus.`);
    } catch (err) {
      console.error('Delete error:', err);
      alert('Gagal menolak: ' + err.message);
    }
  }

  async function handleDelete(id, name) {
    if (!confirm(`Hapus pembimbing "${name}"?`)) return;
    try {
      await api.deleteSupervisor(id);
      supervisors = supervisors.filter(s => s.id !== id);
      alert(`Pembimbing ${name} telah dihapus.`);
    } catch (err) {
      console.error('Delete error:', err);
      alert('Gagal menghapus: ' + err.message);
    }
  }

  function toggleExpand(id) {
    expandedSupervisors[id] = !expandedSupervisors[id];
  }

  onMount(async () => {
    await fetchSupervisors();
  });
</script>

<div class="page-container animate-fade-in">

  <div class="flex items-center gap-3 pb-8">
    <h4 class="card-title">Daftar Pembimbing</h4>
    <span class="badge-count">{supervisors.length} Pembimbing</span>
  </div>
  
  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card animate-slide-up" style="animation-delay: 0.1s;">
    <div class="card-header-row border-b">
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        {#if auth.user?.role === 'admin'}
          <button class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-slate-900 text-white hover:bg-slate-800 transition-all shadow-sm flex items-center justify-center gap-2" onclick={() => showCreateModal = true}>
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            <span>Tambah</span>
          </button>
        {/if}
        <button class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2" onclick={fetchSupervisors}>
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/></svg>
          <span>Refresh</span>
        </button>
      </div>
      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2" onclick={goToPreviousPage} disabled={currentPage <= 1}>
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
          <span>Prev</span>
        </button>

        <div class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2 pagination-pill">
          <span>{currentPage}</span>
          <span class="text-slate-500">of</span>
          <span>{totalPages}</span>
        </div>

        <button class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold bg-white text-slate-900 border border-slate-200 hover:border-slate-300 transition-all flex items-center justify-center gap-2" onclick={goToNextPage} disabled={currentPage >= totalPages}>
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18l6-6-6-6"/></svg>
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
          onkeydown={(e) => e.key === 'Enter' && (clearTimeout(searchTimeout), fetchSupervisors())}
          placeholder="Cari Nama, Email, atau Institusi..." 
          class="search-input"
        />
      </div>
      
      <select bind:value={filterStatus} onchange={fetchSupervisors} class="filter-select">
        <option value="">Semua Status</option>
        <option value="active">Aktif</option>
        <!-- <option value="pending">Pending</option> -->
      </select>
    </div>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Memuat data...</p>
      </div>
    {:else if supervisors.length === 0}
      <div class="empty-state">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#e5e7eb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
        <p>Belum ada data pembimbing.</p>
      </div>
    {:else}
      <div class="table-responsive desktop-only">
        <table class="table">
          <thead>
            <tr>
              <th>Nama</th>
              <th>Email</th>
              <th>Institusi</th>
              <th>Posisi</th>
              <th class="text-center">Jumlah Intern</th>
              <th class="text-center">Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each supervisors as s}
              <tr class="table-row">
                <td>
                  <div class="user-info">
                    <div class="avatar-placeholder">{s.full_name ? s.full_name[0].toUpperCase() : s.name ? s.name[0].toUpperCase() : 'S'}</div>
                    <div class="user-details">
                      <span class="user-name">{s.full_name || s.name || '-'}</span>
                    </div>
                  </div>
                </td>
                <td class="text-muted">{s.email || '-'}</td>
                <td class="text-muted">{s.institution || '-'}</td>
                <td class="text-muted">{s.position || '-'}</td>
                <td class="text-center text-muted">{s.interns_count || 0}</td>
                <td class="text-center">
                  <span class={`status-badge status-${s.status || 'inactive'}`}>
                    {s.status || 'inactive'}
                  </span>
                </td>
                <td class="text-right">
                  <div class="action-buttons responsive">
                    {#if s.status === 'pending'}
                      <button 
                        class="btn-icon btn-deny flex-1" 
                        onclick={() => handleDeny(s.id, s.full_name || s.name)}
                        title="Tolak & Hapus"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                        <span class="btn-label">Tolak</span>
                      </button>
                      <button 
                        class="btn-icon btn-approve flex-1" 
                        onclick={() => handleApprove(s.id, s.full_name || s.name)}
                        title="Setujui Pembimbing Ini"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                        <span class="btn-label">Terima</span>
                      </button>
                    {:else}
                      {#if auth.user?.role === 'admin'}
                        <button class="px-2 py-2 rounded-full text-sm font-semibold border border-slate-200 hover:border-slate-300 text-slate-200 hover:text-white bg-slate-900 hover:bg-slate-800 pointer-cursor transition-all flex items-center justify-center gap-2 flex-1" onclick={() => startEdit(s)} title="Edit data">
                          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9" /><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4Z" /></svg>
                          <span>Edit</span>
                        </button>
                        <button class="px-2 py-2 rounded-full text-sm font-semibold border border-red-200 hover:border-red-300 text-slate-200 hover:text-white bg-red-600 hover:bg-red-500 pointer-cursor transition-all flex items-center justify-center gap-2 flex-1" onclick={() => handleDelete(s.id, s.full_name || s.name)} title="Hapus">
                          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                          <span>Hapus</span>
                        </button>
                      {/if}
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mobile-list">
        {#each supervisors as s}
          <div class="entry-card">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="entry-head" onclick={() => toggleExpand(s.id)}>
              <div class="user-info">
                <div class="avatar-placeholder">{s.full_name ? s.full_name[0].toUpperCase() : s.name ? s.name[0].toUpperCase() : 'S'}</div>
                <div class="user-details">
                  <div class="user-name">{s.full_name || s.name || '-'}</div>
                  <div class="text-muted small">{s.email || '-'}</div>
                </div>
              </div>
              <button class="expand-btn">
                <span class="material-symbols-outlined transition-transform duration-200 {expandedSupervisors[s.id] ? 'rotate-180' : ''}">expand_more</span>
              </button>
            </div>
            
            {#if expandedSupervisors[s.id]}
              <div class="entry-details" transition:slide={{ duration: 200 }}>
                <div class="detail-row">
                  <div class="detail-label">STATUS</div>
                  <span class={`status-badge equal-badge ${s.status === 'active' ? 'bg-emerald-100 text-emerald-700' : s.status === 'pending' ? 'bg-yellow-100 text-yellow-700' : 'bg-slate-100 text-slate-600'}`}>
                    {s.status || 'inactive'}
                  </span>
                </div>
                <div class="detail-row">
                  <div class="detail-label">INSTITUSI</div>
                  <div class="detail-value">{s.institution || '-'}</div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">POSISI</div>
                  <div class="detail-value">{s.position || '-'}</div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">NIP</div>
                  <div class="detail-value">{s.nip || '-'}</div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">TELEPON</div>
                  <div class="detail-value">{s.phone || '-'}</div>
                </div>
                <div class="detail-row">
                  <div class="detail-label">JUMLAH INTERN</div>
                  <div class="detail-value">{s.interns_count || 0}</div>
                </div>

                <div class="mobile-actions mt-4 pt-4 border-t border-slate-100">
                  {#if s.status === 'pending'}
                    <button class="mini-btn mobile danger" onclick={(e) => { e.stopPropagation(); handleDeny(s.id, s.full_name || s.name); }}>
                      <span class="material-symbols-outlined">close</span>
                      <span class="btn-text">Tolak</span>
                    </button>
                    <button class="mini-btn mobile success" onclick={(e) => { e.stopPropagation(); handleApprove(s.id, s.full_name || s.name); }}>
                      <span class="material-symbols-outlined">check</span>
                      <span class="btn-text">Terima</span>
                    </button>
                  {:else}
                    {#if auth.user?.role === 'admin'}
                      <button class="mini-btn mobile" onclick={(e) => { e.stopPropagation(); startEdit(s); }}>
                        <span class="material-symbols-outlined">edit</span>
                        <span class="btn-text">Edit</span>
                      </button>
                      <button class="mini-btn mobile danger" onclick={(e) => { e.stopPropagation(); handleDelete(s.id, s.full_name || s.name); }}>
                        <span class="material-symbols-outlined">delete</span>
                        <span class="btn-text">Hapus</span>
                      </button>
                    {/if}
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
    <div class="fixed inset-0 z-120 flex items-center justify-center p-4 sm:p-6" use:portal>
      <div class="absolute inset-0 z-110 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>
      <div class="relative bg-white z-120 rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
        <div class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <h3 class="font-bold text-lg text-slate-800">Tambah Pembimbing</h3>
          <button onclick={() => showCreateModal = false} class="text-slate-400 hover:text-slate-600">
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>
        
        <div class="p-6 overflow-y-auto">
          <div class="form-grid">
            <div class="form-group">
              <label class="form-label" for="name">Nama Lengkap <span class="text-red-500">*</span></label>
              <input class="input" bind:value={form.name} id="name" placeholder="Contoh: Dr. Budi Santoso" required />
            </div>
            <div class="form-group">
              <label class="form-label" for="email">Email <span class="text-red-500">*</span></label>
              <input class="input" type="email" bind:value={form.email} id="email" placeholder="email@institusi.com" required />
            </div>
            <div class="form-group">
              <label class="form-label" for="password">Password <span class="text-red-500">*</span></label>
              <input class="input" type="password" bind:value={form.password} id="password" placeholder="Min. 6 karakter" required />
            </div>
            <div class="form-group">
              <label class="form-label" for="password_confirmation">Ulangi Password <span class="text-red-500">*</span></label>
              <input class="input" type="password" bind:value={form.password_confirmation} id="password_confirmation" placeholder="Ulangi password" required />
            </div>
            <div class="form-group">
              <label class="form-label" for="nip">NIP</label>
              <input class="input" bind:value={form.nip} id="nip" placeholder="Nomor Induk Pegawai" />
            </div>
            <div class="form-group">
              <label class="form-label" for="phone">Telepon</label>
              <input class="input" bind:value={form.phone} id="phone" placeholder="Nomor telepon" />
            </div>
            <div class="form-group">
              <label class="form-label" for="position">Posisi</label>
              <input class="input" bind:value={form.position} id="position" placeholder="Jabatan" />
            </div>
            <div class="form-group">
              <label class="form-label" for="institution">Institusi</label>
              <input class="input" bind:value={form.institution} id="institution" placeholder="Nama institusi" />
            </div>
            <div class="form-group">
              <label class="form-label" for="address">Alamat</label>
              <input class="input" bind:value={form.address} id="address" placeholder="Alamat lengkap" />
            </div>
            <div class="form-group">
              <label class="form-label" for="status">Status</label>
              <select class="input" id="status" bind:value={form.status}>
                <option value="active">Aktif</option>
                <option value="pending">Pending</option>
              </select>
            </div>
          </div>
        </div>

        <div class="p-5 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3">
          <button class="btn btn-outline" onclick={() => showCreateModal = false}>Batal</button>
          <button class="btn btn-primary" onclick={createSupervisor}>Simpan Data</button>
        </div>
      </div>
    </div>
  {/if}

  <!-- EDIT MODAL -->
  {#if showEditModal}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="fixed inset-0 z-120 flex items-center justify-center p-4 sm:p-6" use:portal>
      <div class="absolute inset-0 z-110 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>
      <div class="relative bg-white z-120 rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
        <div class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <h3 class="font-bold text-lg text-slate-800">Edit Data Pembimbing</h3>
          <button onclick={closeEditModal} class="text-slate-400 hover:text-slate-600">
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>
        
        <div class="p-6 overflow-y-auto">
          <div class="form-grid">
            <div class="form-group">
              <label class="form-label" for="edit_name">Nama Lengkap</label>
              <input class="input" bind:value={editForm.name} id="edit_name" placeholder="Contoh: Dr. Budi Santoso" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_email">Email</label>
              <input class="input" type="email" bind:value={editForm.email} id="edit_email" placeholder="email@institusi.com" disabled />
              <small class="text-xs text-gray-500 mt-1">Email tidak dapat diubah</small>
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_password">Password Baru (Opsional)</label>
              <input class="input" type="password" bind:value={editForm.password} id="edit_password" placeholder="Kosongkan jika tidak ingin mengubah" />
            </div>
            <div class="form-group">
                <label class="form-label" for="edit_password_confirmation">Ulangi Password Baru</label>
                <input class="input" type="password" bind:value={editForm.password_confirmation} id="edit_password_confirmation" placeholder="Ulangi password baru" />
              </div>
              <div class="form-group">
              <label class="form-label" for="edit_nip">NIP</label>
              <input class="input" bind:value={editForm.nip} id="edit_nip" placeholder="Nomor Induk Pegawai" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_phone">Telepon</label>
              <input class="input" bind:value={editForm.phone} id="edit_phone" placeholder="Nomor telepon" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_position">Posisi</label>
              <input class="input" bind:value={editForm.position} id="edit_position" placeholder="Jabatan" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_institution">Institusi</label>
              <input class="input" bind:value={editForm.institution} id="edit_institution" placeholder="Nama institusi" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_address">Alamat</label>
              <input class="input" bind:value={editForm.address} id="edit_address" placeholder="Alamat lengkap" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit_status">Status</label>
              <select class="input" id="edit_status" bind:value={editForm.status}>
                <!-- <option value="pending">Pending</option> -->
                <option value="active">Aktif</option>
              </select>
            </div>
          </div>
        </div>

        <div class="p-5 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3">
          <button class="btn btn-outline" onclick={closeEditModal}>Batal</button>
          <button class="btn btn-primary" onclick={updateSupervisor}>Simpan Perubahan</button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  /* Base Layout */

  .page-container {
    animation: fadeIn 0.5s ease-out;
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
    padding: 0 16px;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  /* Card Styles */
  .card {
    background: white;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02);
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

  .card-header h4 {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 600;
    color: #111827;
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
    background-color: #10b981;
    color: white;
    border: none;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 0.9rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2);
  }
  
  .btn-approve:hover {
    background-color: #059669;
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.3);
  }

  .btn-deny {
    display: inline-flex;
    align-items: center;
    background-color: white;
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
    background-color: #ef4444;
    color: white;
    box-shadow: 0 4px 6px rgba(239, 68, 68, 0.3);
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
  @media (max-width: 640px) {
    .card-header-row {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }
    .toolbar { padding: 14px 16px; }
    .search-wrapper { flex: 1 1 100%; }
    .filter-select { width: 100%; }
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
    font-variation-settings: 'wght' 550;
  }
  .search-input {
    flex: 1;
    border: none;
    outline: none;
    font-size: 0.95rem;
    background: transparent;
    color: #111827;
  }
  .search-input::placeholder { color: #9ca3af; }

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
    border-collapse: collapse;
    font-size: 0.925rem;
  }
  .desktop-only { display:block; }
  .mobile-list { display:none; }

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

  .text-placeholder { color: #d1d5db; font-size: 1.2rem; }
  .text-center { text-align: center; }
  .text-right { text-align: right; }

  .pagination-pill {
    min-width: 128px;
  }

  /* Modal z-index helpers */
  :global(.z-120) { z-index: 120; }
  :global(.z-110) { z-index: 110; }
  :global(.z-100) { z-index: 100; }

  .action-buttons {
    display: inline-flex;
    gap: 8px;
    justify-content: flex-end;
  }
  .action-buttons span {
    display: none;
  }
  .action-buttons.responsive {
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
  .status-inactive {
    background-color: #f3f4f6;
    color: #4b5563;
    border: 1px solid #e5e7eb;
  }
  .bg-emerald-100 { background: #ecfdf5; border-color: #a7f3d0; } .text-emerald-700 { color: #047857; }
  .bg-yellow-100 { background: #fefce8; border-color: #fef08a; } .text-yellow-700 { color: #a16207; }
  .bg-slate-100 { background: #f1f5f9; border-color: #e2e8f0; } .text-slate-600 { color: #475569; }
  .equal-badge { min-width: 96px; text-align:center; justify-content:center; display:inline-flex; }

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

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .btn-icon {
    width: 42px;
    height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }
  .btn-label { display: none; font-weight: 700; font-size: 0.85rem; }

  .text-red-500 { color: #ef4444; }
  .text-xs { font-size: 0.75rem; }
  .text-gray-500 { color: #6b7280; }
  .mt-1 { margin-top: 0.25rem; }
  .flex { display: flex; }
  .gap-2 { gap: 0.5rem; }
  .gap-3 { gap: 0.75rem; }
  .items-center { align-items: center; }
  .border-b { border-bottom: 1px solid #f1f5f9; }

  @media (max-width: 640px) {
    .table-responsive { overflow: visible; }
    .table, .table thead { display: block; }
    .table tbody { display: grid; gap: 12px; }
    .table tr { display: grid; gap: 8px; padding: 14px; border: 1px solid #e5e7eb; border-radius: 12px; box-shadow: 0 2px 6px rgba(0,0,0,0.03); }
    .table td, .table th { padding: 4px 0; border: none; }
    .text-right { text-align: left; }
    .action-buttons.responsive { width: 100%; }
    .action-buttons.responsive .btn-icon { flex: 1 1 48%; width: 100%; }
    .action-buttons span,
    .btn-label { display: inline; }
    .user-info { align-items: flex-start; }
  }
  @media (max-width: 768px) {
    .desktop-only { display:none; }
    .mobile-list { display:flex; flex-direction:column; gap:12px; }
    .entry-card {
      padding: 14px;
      border-radius: 16px;
      border: 1px solid #e2e8f0;
      background: #ffffff;
      box-shadow: 0 6px 20px -18px rgba(15,23,42,0.3);
    }
    .entry-head { display:flex; align-items:center; justify-content:space-between; gap:10px; cursor: pointer; }
    .entry-head .user-details { display: flex; flex-direction: column; min-width: 0; }
    .entry-head .user-name { font-size: 0.95rem; font-weight: 600; color: #0f172a; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .entry-head .text-muted { font-size: 0.8rem; color: #64748b; margin: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .entry-head .avatar-placeholder { width: 40px; height: 40px; font-size: 1rem; flex-shrink: 0; }
    .entry-details { margin-top: 16px; padding-top: 16px; border-top: 1px solid #f1f5f9; }

    .entry-head .user-info {
      display: grid;
      grid-template-columns: 40px 1fr;
      gap: 12px;
      align-items: center;
      flex: 1;
      min-width: 0;
    }

    .detail-row { margin-bottom: 16px; }
    .detail-row:last-child { margin-bottom: 0; }
    .detail-label { font-size:11px; font-weight:700; color:#94a3b8; text-transform:uppercase; letter-spacing:0.05em; margin-bottom: 4px; }
    .detail-value { font-weight:600; color:#0f172a; font-size: 14px; }
    
    .expand-btn {
      width: 32px; height: 32px;
      display: flex; align-items: center; justify-content: center;
      border-radius: 50%;
      background: #f8fafc;
      color: #64748b;
      border: none;
    }
    .mobile-actions { display:flex; gap:10px; }
    .mini-btn {
      display:inline-flex; align-items:center; gap:6px;
      padding:8px 16px; border-radius:9999px; border:1px solid #0f172a;
      background:#0f172a; color:#fff; font-weight:700; font-size:13px;
      cursor:pointer; transition:all 0.15s ease; flex:1; justify-content:center;
    }
    .mini-btn .btn-text { display:inline; }
    .mini-btn.success { background:#10b981; border-color:#10b981; }
    .mini-btn.danger { background:#ef4444; border-color:#ef4444; }
  }

  /* Animations */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>