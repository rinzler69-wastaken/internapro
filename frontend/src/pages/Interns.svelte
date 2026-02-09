<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  // State
  let interns = $state([]);
  let loading = $state(false);

  // Form State
  let form = $state({
    email: '',
    password: '',
    full_name: '',
    school: '',
    department: '',
    start_date: '',
    end_date: '',
  });

  // --- Fetch Data ---
  async function fetchInterns() {
    loading = true;
    try {
      const res = await api.getInterns({ page: 1, limit: 100 });
      interns = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  // --- Create Intern ---
  async function createIntern() {
    if (!form.email || !form.password || !form.full_name) {
        alert('Mohon lengkapi data wajib (Email, Password, Nama)');
        return;
    }

    try {
      await api.createIntern(form);
      alert('Berhasil menambah intern!');
      
      // Reset form setelah berhasil
      form = {
        email: '',
        password: '',
        full_name: '',
        school: '',
        department: '',
        start_date: '',
        end_date: '',
      };
      
      // Refresh tabel
      await fetchInterns();
    } catch (err) {
      alert(err.message || 'Gagal membuat intern');
    }
  }

  // --- Approve Intern ---
  async function handleApprove(id, name) {
    if (!confirm(`Setujui siswa "${name}" menjadi peserta Aktif?`)) return;

    try {
      await api.updateInternStatus(id, 'active');
      
      // Update tampilan tabel secara langsung (tanpa reload page)
      const index = interns.findIndex(i => i.id === id);
      if (index !== -1) {
        interns[index].status = 'active';
      }
      
    } catch (err) {
      console.error(err);
      alert('Gagal melakukan approval: ' + (err.response?.data?.message || err.message));
    }
  }

  // --- Deny/Delete Intern ---
  async function handleDeny(id, name) {
    if (!confirm(`Apakah Anda yakin ingin MENOLAK dan MENGHAPUS pendaftaran "${name}"? Data akan hilang permanen.`)) return;
    try {
        // Menggunakan generic delete endpoint, asumsi RESTful /interns/:id
        await api.delete(`/interns/${id}`);
        interns = interns.filter(i => i.id !== id);
        alert(`Pendaftaran ${name} telah ditolak dan dihapus.`);
    } catch (err) {
        alert('Gagal menolak: ' + (err.response?.data?.message || err.message));
    }
  }

  onMount(fetchInterns);
</script>

<div class="page-container">
  
  <!-- BAGIAN FORMULIR TAMBAH -->
  <div class="card form-card">
    <div class="card-header">
        <h4>Tambah Intern Manual</h4>
        <p class="text-muted">Masukkan data siswa baru secara manual.</p>
    </div>
    
    <div class="form-grid">
      <div class="form-group">
        <label class="form-label" for="full_name">Nama Lengkap</label>
        <input class="input" bind:value={form.full_name} id="full_name" placeholder="Contoh: Budi Santoso" />
      </div>
      <div class="form-group">
        <label class="form-label" for="email">Email</label>
        <input class="input" type="email" bind:value={form.email} id="email" placeholder="email@sekolah.com" />
      </div>
      <div class="form-group">
        <label class="form-label" for="password">Password</label>
        <input class="input" type="password" bind:value={form.password} id="password" placeholder="******" />
      </div>
      <div class="form-group">
        <label class="form-label" for="school">Sekolah</label>
        <input class="input" bind:value={form.school} id="school" placeholder="Nama Sekolah" />
      </div>
      <div class="form-group">
        <label class="form-label" for="department">Jurusan</label>
        <input class="input" bind:value={form.department} id="department" placeholder="Jurusan" />
      </div>
      <div class="form-group">
        <label class="form-label" for="start_date">Mulai</label>
        <input class="input" type="date" bind:value={form.start_date} id="start_date" />
      </div>
      <div class="form-group">
        <label class="form-label" for="end_date">Selesai</label>
        <input class="input" type="date" bind:value={form.end_date} id="end_date" />
      </div>
    </div>
    
    <div class="form-actions">
        <button class="btn btn-primary" onclick={createIntern}>
            <span class="icon">+</span> Simpan Data
        </button>
    </div>
  </div>

  <!-- BAGIAN TABEL DAFTAR -->
  <div class="card table-card">
    <div class="card-header-row">
        <div>
            <h4 class="card-title">Daftar Intern</h4>
            <span class="badge-count">{interns.length} Siswa</span>
        </div>
        <button class="btn btn-outline btn-sm" onclick={fetchInterns}>
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/></svg>
            Refresh
        </button>
    </div>

    {#if loading}
      <div class="loading-state">
         <div class="spinner"></div>
         <p>Memuat data...</p>
      </div>
    {:else if interns.length === 0}
      <div class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#e5e7eb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
          <p>Belum ada data intern.</p>
      </div>
    {:else}
      <div class="table-responsive">
          <table class="table">
          <thead>
              <tr>
              <th>Nama</th>
              <th>Email</th>
              <th class="text-center">Status</th>
              <th class="text-right">Aksi</th>
              </tr>
          </thead>
          <tbody>
              {#each interns as i}
              <tr class="table-row">
                  <td>
                      <div class="user-info">
                          <div class="avatar-placeholder">{i.full_name ? i.full_name[0].toUpperCase() : '?'}</div>
                          <div class="user-details">
                              <span class="user-name">{i.full_name || '-'}</span>
                          </div>
                      </div>
                  </td>
                  <td class="text-muted">{i.email || '-'}</td>
                  
                  <td class="text-center">
                      <span class={`status-badge status-${i.status || 'inactive'}`}>
                          {i.status || 'inactive'}
                      </span>
                  </td>
                  <td class="text-right">
                      {#if i.status === 'pending'}
                        <div class="action-buttons">
                          <button 
                              class="btn-deny" 
                              onclick={() => handleDeny(i.id, i.full_name)}
                              title="Tolak & Hapus"
                          >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                              Tolak
                          </button>
                          <button 
                              class="btn-approve" 
                              onclick={() => handleApprove(i.id, i.full_name)}
                              title="Setujui Siswa Ini"
                          >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                              Approve
                          </button>
                        </div>
                      {:else}
                          <span class="text-placeholder">-</span>
                      {/if}
                  </td>
              </tr>
              {/each}
          </tbody>
          </table>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Base Layout */
  .page-container {
    animation: fadeIn 0.5s ease-out;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  /* Card Styles */
  .card {
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03);
    border: 1px solid rgba(229, 231, 235, 0.5);
    margin-bottom: 24px;
    overflow: hidden;
    transition: box-shadow 0.3s ease, transform 0.3s ease;
  }

  .card:hover {
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -2px rgba(0, 0, 0, 0.04);
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

  .btn:active {
    transform: scale(0.98);
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
    gap: 4px;
    background-color: #10b981;
    color: white;
    border: none;
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(16, 185, 129, 0.2);
  }
  
  .btn-approve:hover {
    background-color: #059669;
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.3);
    transform: translateY(-1px);
  }

  .btn-deny {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background-color: white;
    color: #ef4444;
    border: 1px solid #ef4444;
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-deny:hover {
    background-color: #ef4444;
    color: white;
    box-shadow: 0 4px 6px rgba(239, 68, 68, 0.3);
    transform: translateY(-1px);
  }

  /* Table Header Row */
  .card-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #f3f4f6;
  }

  .card-title {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #111827;
    display: inline-block;
  }

  .badge-count {
    background-color: #eff6ff;
    color: #2563eb;
    padding: 2px 8px;
    border-radius: 99px;
    font-size: 0.75rem;
    font-weight: 600;
    margin-left: 8px;
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
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .avatar-placeholder {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, #6366f1, #4f46e5);
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

  .text-muted { color: #6b7280; }
  .text-placeholder { color: #d1d5db; font-size: 1.2rem; }
  .text-center { text-align: center; }
  .text-right { text-align: right; }

  .action-buttons {
    display: inline-flex;
    gap: 8px;
    justify-content: flex-end;
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

  .status-inactive, .status-cancelled {
    background-color: #f3f4f6;
    color: #4b5563;
    border: 1px solid #e5e7eb;
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

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>