<script>
  import { onMount } from 'svelte';
  import { route } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // State
  let records = $state([]);
  let todayAttendance = $state(null);
  let loading = $state(false);
  let detailOpen = $state(false);
  let detailLoading = $state(false);
  let detail = $state(null);
  let docOpen = $state(false);
  let docUrl = $state('');
  
  // Permission Form State
  let permissionStatus = $state('sick');
  let permissionNotes = $state('');
  let permissionFile = $state(null);
  /** @type {HTMLInputElement} */
  let fileInput;
  
  // --- STATE LOADING TERPISAH (SOLUSI) ---
  let isCheckingIn = $state(false);       // Khusus Check In/Out
  let isSubmittingPermission = $state(false); // Khusus Kirim Izin

  // Labels & Colors
  const statusLabels = {
    present: 'Hadir',
    late: 'Terlambat',
    absent: 'Tidak Hadir',
    sick: 'Sakit',
    permission: 'Izin',
  };

  function getStatusColor(status) {
    switch (status) {
        case 'present': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
        case 'late': return 'bg-yellow-100 text-yellow-700 border-yellow-200';
        case 'sick': return 'bg-blue-100 text-blue-700 border-blue-200';
        case 'permission': return 'bg-purple-100 text-purple-700 border-purple-200';
        case 'absent': return 'bg-red-100 text-red-600 border-red-200';
        default: return 'bg-slate-100 text-slate-600 border-slate-200';
    }
  }

  function formatDate(value) {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  function formatTime(value) {
      if(!value) return '--:--';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) {
        if (typeof value === 'string' && value.includes(':')) return value.slice(0,5);
        return '--:--';
      }
      return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
  }

  function withBase(url) {
    if (!url) return '';
    const needsToken = url.startsWith('/uploads') || url.startsWith('uploads/');
    const token = auth.token;
    const tokenQS = needsToken && token ? `${url.includes('?') ? '&' : '?'}token=${token}` : '';
    if (url.startsWith('http')) return `${url}${tokenQS}`;
    const base = import.meta.env.VITE_API_URL || '';
    const normalized = url.startsWith('/') ? url : `/${url}`;
    return `${base}${normalized}${tokenQS}`;
  }

  async function openDetail(id) {
    detailOpen = true;
    detailLoading = true;
    detail = null;
    try {
      const res = await api.getAttendanceById(id);
      detail = res.data;
    } catch (err) {
      console.error(err);
    } finally {
      detailLoading = false;
    }
  }

  function closeDetail() {
    detailOpen = false;
    detail = null;
  }

  function openDoc(url) {
    const full = withBase(url);
    docUrl = full;
    docOpen = true;
  }

  function closeDoc() {
    docOpen = false;
    docUrl = '';
  }

  function isImageFile(url) {
    if (!url) return false;
    return /\.(jpg|jpeg|png|webp|gif|svg)($|\?)/i.test(url);
  }

  function rowTint(status) {
    switch (status) {
      case 'present': return 'bg-emerald-50';
      case 'late': return 'bg-amber-50';
      case 'permission': return 'bg-purple-50';
      case 'sick': return 'bg-blue-50';
      case 'absent': return 'bg-rose-50';
      default: return '';
    }
  }

  async function fetchAttendance() {
    loading = true;
    try {
      if (auth.user?.role === 'intern') {
        const todayRes = await api.getTodayAttendance();
        todayAttendance = todayRes.data?.today_attendance || todayRes.data || null; 
        
        const historyRes = await api.getAttendance({ page: 1, limit: 10 });
        records = historyRes.data || [];
      } else {
        const res = await api.getAttendance({ page: 1, limit: 20 });
        records = res.data || [];
      }
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  const statusBadge = (status) => {
    switch (status) {
      case 'present': return { text: 'Hadir', cls: 'bg-emerald-100 text-emerald-700 border-emerald-200' };
      case 'late': return { text: 'Terlambat', cls: 'bg-amber-100 text-amber-700 border-amber-200' };
      case 'permission': return { text: 'Izin', cls: 'bg-purple-100 text-purple-700 border-purple-200' };
      case 'sick': return { text: 'Sakit', cls: 'bg-blue-100 text-blue-700 border-blue-200' };
      case 'absent': return { text: 'Tidak Hadir', cls: 'bg-rose-100 text-rose-700 border-rose-200' };
      default: return { text: status || '-', cls: 'bg-slate-100 text-slate-600 border-slate-200' };
    }
  };

  function getPosition() {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
          reject(new Error("Geolocation tidak didukung browser ini."));
          return;
      }
      navigator.geolocation.getCurrentPosition(resolve, reject, { enableHighAccuracy: true });
    });
  }

  // --- LOGIC CHECK IN/OUT ---
  async function handleCheckIn() {
    isCheckingIn = true; // Pakai variable khusus
    try {
      const pos = await getPosition();
      await api.checkIn(pos.coords.latitude, pos.coords.longitude);
      alert("Berhasil Check-in!");
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal check-in. Pastikan izin lokasi aktif.');
    } finally {
      isCheckingIn = false;
    }
  }

  async function handleCheckOut() {
    if(!confirm("Apakah Anda yakin ingin Check-out sekarang?")) return;
    isCheckingIn = true; // Pakai variable khusus
    try {
      const pos = await getPosition();
      await api.checkOut(pos.coords.latitude, pos.coords.longitude);
      alert("Berhasil Check-out!");
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal check-out');
    } finally {
      isCheckingIn = false;
    }
  }

  // --- LOGIC PERMISSION ---
  async function handlePermissionSubmit() {
    if (!permissionNotes) {
        alert("Mohon isi catatan/alasan.");
        return;
    }
    isSubmittingPermission = true; // Pakai variable khusus
    try {
      await api.submitPermission({
        status: permissionStatus,
        notes: permissionNotes,
        proof_file: permissionFile,
      });
      alert('Pengajuan izin berhasil dikirim.');
      
      permissionNotes = '';
      permissionFile = null;
      if(fileInput) fileInput.value = '';
      
      await fetchAttendance();
    } catch (err) {
      alert(err.message || 'Gagal mengajukan izin');
    } finally {
      isSubmittingPermission = false;
    }
  }

  onMount(fetchAttendance);
</script>

<!-- <div class="page-bg"> -->
  <div class="container animate-fade-in">
    
    <!-- Header -->
    <!-- <div class="header">
        <div>
            <h2 class="title">Presensi & Kehadiran</h2>
            <p class="subtitle">Kelola jam kerja dan riwayat kehadiran Anda.</p>
        </div>
        <div class="header-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                <line x1="16" y1="2" x2="16" y2="6"></line>
                <line x1="8" y1="2" x2="8" y2="6"></line>
                <line x1="3" y1="10" x2="21" y2="10"></line>
                <path d="M12 14h.01"></path>
            </svg>
        </div>
    </div> -->

    <!-- {#if auth.user?.role === 'intern'}
        <div class="grid-layout animate-slide-up"> -->
            
            <!-- ABSENSI CARD -->
            <!-- <div class="card main-action-card">
                <div class="card-header">
                    <h3>Status Hari Ini</h3>
                    <div class="date-badge">
                        {new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long' })}
                    </div>
                </div>

                <div class="status-display">
                    {#if todayAttendance && todayAttendance.status}
                        <div class={`status-circle ${todayAttendance.status === 'present' ? 'status-green' : 'status-yellow'}`}>
                            <span class="status-text">{statusLabels[todayAttendance.status] || todayAttendance.status}</span>
                            <span class="time-text">
                                {todayAttendance.check_in_time ? formatTime(todayAttendance.check_in_time) : '--:--'}
                            </span>
                        </div>
                    {:else}
                        <div class="status-circle status-gray">
                            <span class="status-text">Belum Absen</span>
                            <span class="time-text">--:--</span>
                        </div>
                    {/if}
                </div>

                <div class="action-buttons">
                    <button 
                        class="btn-checkin" 
                        onclick={handleCheckIn} 
                        disabled={isCheckingIn || (todayAttendance && todayAttendance.check_in_time)}
                    >
                        {#if isCheckingIn && !todayAttendance?.check_in_time}
                            <div class="spinner-small"></div>
                        {:else}
                            <div class="btn-icon">📍</div>
                        {/if}
                        <span>Check In</span>
                    </button>
                    
                    <button 
                        class="btn-checkout" 
                        onclick={handleCheckOut} 
                        disabled={isCheckingIn || !todayAttendance || !todayAttendance.check_in_time || todayAttendance.check_out_time}
                    >
                        {#if isCheckingIn && todayAttendance?.check_in_time}
                            <div class="spinner-small dark"></div>
                        {:else}
                            <div class="btn-icon">👋</div>
                        {/if}
                        <span>Check Out</span>
                    </button>
                </div>
            </div> -->

            <!-- FORM IZIN CARD -->
            <!-- <div class="card permission-card">
                <div class="card-header">
                    <h3>Pengajuan Izin / Sakit</h3>
                </div>
                <div class="form-content">
                    <div class="form-group">
                        <label class="label" for="perm-status">Kategori</label>
                        <div class="radio-group">
                            <label class={`radio-btn ${permissionStatus === 'sick' ? 'active' : ''}`}>
                                <input type="radio" value="sick" bind:group={permissionStatus} hidden>
                                🤒 Sakit
                            </label>
                            <label class={`radio-btn ${permissionStatus === 'permission' ? 'active' : ''}`}>
                                <input type="radio" value="permission" bind:group={permissionStatus} hidden>
                                📝 Izin
                            </label>
                        </div>
                    </div>

                    <div class="form-group">
                        <label class="label" for="perm-notes">Keterangan</label>
                        <textarea class="textarea" id="perm-notes" rows="3" placeholder="Jelaskan alasan Anda..." bind:value={permissionNotes}></textarea>
                    </div>

                    <div class="form-group">
                        <label class="label" for="file-upload">Bukti Pendukung (Opsional)</label>
                        <label class="file-drop" for="file-upload">
                            {#if permissionFile}
                                <span class="file-name">{permissionFile.name}</span>
                            {:else}
                                <span class="placeholder">Klik untuk upload surat dokter/dokumen</span>
                            {/if}
                            <input 
                                id="file-upload" 
                                type="file" 
                                hidden 
                                bind:this={fileInput}
                                onchange={(e) => permissionFile = e.currentTarget.files?.[0] || null}
                            >
                        </label>
                    </div> -->

                    <!-- Gunakan isSubmittingPermission -->
                    <!-- <button class="btn-submit" onclick={handlePermissionSubmit} disabled={isSubmittingPermission}>
                        {isSubmittingPermission ? 'Mengirim...' : 'Kirim Pengajuan'}
                    </button>
                </div>
            </div> -->
        <!-- </div>
    {/if} -->

    <!-- TABEL RIWAYAT -->

            <div class="card-header border-b" style="justify-content: flex-start; gap: 12px;">
            <h3>Riwayat Presensi</h3>
            <span class="badge-count">{records.length} Data</span>
        </div>
     
    <div class="card list-card animate-slide-up" style="animation-delay: 0.1s;">


        {#if loading}
            <div class="loading-state">Memuat data...</div>
        {:else if records.length === 0}
            <div class="empty-state">Belum ada riwayat presensi.</div>
        {:else}
            <div class="table-container desktop-only">
                <table class="table desktop-table">
                    <thead>
                        <tr>
                            <th>Tanggal</th>
                            <th>Waktu Masuk</th>
                            <th>Waktu Keluar</th>
                            <th>Status</th>
                            {#if auth.user?.role !== 'intern'}
                                <th>Intern</th>
                            {/if}
                            <th>Aksi</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each records as r}
                            <tr class={`hover-row ${rowTint(r.status)}`}>
                                <td class="font-medium">
                                    {formatDate(r.date)}
                                </td>
                                <td class="text-slate-500">
                                    {formatTime(r.check_in_time)}
                                </td>
                                <td class="text-slate-500">
                                    {formatTime(r.check_out_time)}
                                </td>
                                <td>
                                    <span class={`status-badge equal-badge ${getStatusColor(r.status)}`}>
                                        {statusLabels[r.status] || r.status || '-'}
                                    </span>
                                </td>
                                {#if auth.user?.role !== 'intern'}
                                    <td>
                                        <div class="user-info">
                                            <div class="avatar-mini">{r.intern_name?.charAt(0) || 'U'}</div>
                                            <span>{r.intern_name || '-'}</span>
                                        </div>
                                    </td>
                                {/if}
                                <td class="action-cell">
                                  <button class="mini-btn icon-only" onclick={() => openDetail(r.id)} aria-label="Detail presensi">
                                    <span class="material-symbols-outlined">info</span>
                                  </button>
                                  <button
                                    class="mini-btn icon-only {r.proof_file ? '' : 'disabled'}"
                                    onclick={() => r.proof_file && openDoc(r.proof_file)}
                                    aria-label="Bukti izin"
                                    disabled={!r.proof_file}
                                  >
                                    <span class="material-symbols-outlined">attach_file</span>
                                  </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>

            <div class="mobile-list">
              {#each records as r}
                {@const badge = statusBadge(r.status)}
                <div class={`entry-card ${rowTint(r.status)}`}>
                  <div class="entry-head">
                    <div class="date-row">
                      <span class="material-symbols-outlined date-icon">calendar_month</span>
                      <span class="date-text">{formatDate(r.date)}</span>
                    </div>
                    <span class={`status-badge equal-badge ${badge.cls}`}>{badge.text}</span>
                  </div>
                  {#if auth.user?.role !== 'intern'}
                    <div class="intern-grid">
                      <div class="intern-box">
                      <div class="avatar-mini">{r.intern_name?.charAt(0) || 'U'}</div>
                      <span class="intern-box-label">{r.intern_name || '-'}</span>
                    </div>
                    </div>
                  {/if}
                  <div class="time-grid">
                    <div class="time-box">
                      <p class="label">Presensi Masuk</p>
                      <p class="time-value">{formatTime(r.check_in_time)}</p>
                    </div>
                    <div class="time-box">
                      <p class="label">Presensi Keluar</p>
                      <p class="time-value">{formatTime(r.check_out_time)}</p>
                    </div>
                  </div>
                  <div class="mobile-actions">
                    <button class="mini-btn mobile" onclick={() => openDetail(r.id)}>
                      <span class="material-symbols-outlined">info</span>
                      <span class="btn-text">Detail</span>
                    </button>
                    <button class="mini-btn mobile {r.proof_file ? '' : 'disabled'}" onclick={() => r.proof_file && openDoc(r.proof_file)} disabled={!r.proof_file}>
                      <span class="material-symbols-outlined">attach_file</span>
                      <span class="btn-text">Lampiran</span>
                    </button>
                  </div>
                </div>
              {/each}
            </div>
        {/if}
    </div>
  </div>

    {#if detailOpen}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6" onclick={closeDetail}>
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>
        
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
          <!-- Header -->
          <div class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <div>
              <h3 class="text-lg font-bold text-slate-800">Detail Presensi</h3>
              {#if detail}
                <p class="text-slate-500 text-sm mt-0.5">{detail.intern_name || auth.user?.name || 'Intern'}</p>
              {/if}
            </div>
            <button onclick={closeDetail} class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors">
              <span class="material-symbols-outlined text-[20px]">close</span>
            </button>
          </div>

          <!-- Body -->
          <div class="p-6 overflow-y-auto">
            {#if detailLoading}
              <div class="flex flex-col items-center justify-center py-8 text-slate-500">
                <div class="w-8 h-8 border-2 border-slate-200 border-t-indigo-500 rounded-full animate-spin mb-2"></div>
                <p class="text-sm">Memuat detail...</p>
              </div>
            {:else if detail}
              {@const badge = statusBadge(detail.status)}
              <div class="space-y-6">
                <!-- Status Banner -->
                <div class="flex items-center justify-between p-4 rounded-xl {badge.cls.replace('text-', 'bg-opacity-10 bg-').replace('border-', 'border-opacity-20 border-')} border">
                   <div class="flex flex-col">
                      <span class="text-xs font-bold uppercase tracking-wider opacity-70">Status</span>
                      <span class="font-bold text-lg">{badge.text}</span>
                   </div>
                   <div class="text-right">
                      <span class="text-xs font-bold uppercase tracking-wider opacity-70">Tanggal</span>
                      <div class="font-bold">{formatDate(detail.date || detail.check_in_time)}</div>
                   </div>
                </div>

                <!-- Time Grid -->
                <div class="grid grid-cols-2 gap-4">
                  <div class="p-4 bg-slate-50 rounded-xl border border-slate-100 text-center">
                    <div class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-1">Masuk</div>
                    <div class="text-xl font-mono font-bold text-slate-700">{formatTime(detail.check_in_time)}</div>
                  </div>
                  <div class="p-4 bg-slate-50 rounded-xl border border-slate-100 text-center">
                    <div class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-1">Keluar</div>
                    <div class="text-xl font-mono font-bold text-slate-700">{formatTime(detail.check_out_time)}</div>
                  </div>
                </div>

                <!-- Details -->
                <div class="space-y-4">
                  {#if detail.late_reason}
                    <div>
                      <h4 class="text-sm font-bold text-slate-700 mb-1">Alasan Terlambat</h4>
                      <p class="text-sm text-slate-600 bg-amber-50 p-3 rounded-lg border border-amber-100">{detail.late_reason}</p>
                    </div>
                  {/if}
                  
                  {#if detail.notes}
                    <div>
                      <h4 class="text-sm font-bold text-slate-700 mb-1">Catatan / Izin</h4>
                      <p class="text-sm text-slate-600 bg-slate-50 p-3 rounded-lg border border-slate-100">{detail.notes}</p>
                    </div>
                  {/if}

                  {#if detail.proof_file}
                    <div>
                      <h4 class="text-sm font-bold text-slate-700 mb-2">Bukti Lampiran</h4>
                      <a href={withBase(detail.proof_file)} target="_blank" rel="noreferrer" class="flex items-center gap-3 p-3 rounded-lg border border-slate-200 hover:border-indigo-300 hover:bg-indigo-50 transition-all group">
                        <div class="w-10 h-10 rounded-lg bg-indigo-100 text-indigo-600 flex items-center justify-center">
                          <span class="material-symbols-outlined">description</span>
                        </div>
                        <div class="flex-1">
                          <p class="text-sm font-bold text-slate-700 group-hover:text-indigo-700">Lihat Dokumen</p>
                          <p class="text-xs text-slate-500">Klik untuk membuka</p>
                        </div>
                        <span class="material-symbols-outlined text-slate-400 group-hover:text-indigo-500">open_in_new</span>
                      </a>
                    </div>
                  {/if}
                </div>
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <p>Detail tidak ditemukan.</p>
              </div>
            {/if}
          </div>
          
          <!-- Footer -->
          <div class="p-4 border-t border-slate-100 bg-slate-50/50 text-right">
            <button onclick={closeDetail} class="px-4 py-2 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 font-medium text-sm transition-colors shadow-sm">
              Tutup
            </button>
          </div>
        </div>
      </div>
    {/if}

    {#if docOpen}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6" onclick={closeDoc}>
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>
        
        <div class="relative bg-white rounded-2xl shadow-xl flex flex-col overflow-hidden transition-all duration-200"
             class:w-full={!isImageFile(docUrl)}
             class:max-w-4xl={!isImageFile(docUrl)}
             class:h-[85vh]={!isImageFile(docUrl)}
             class:w-auto={isImageFile(docUrl)}
             class:max-w-[90vw]={isImageFile(docUrl)}
             class:max-h-[90vh]={isImageFile(docUrl)}
             onclick={(e) => e.stopPropagation()}>
          <!-- Header -->
          <div class="p-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50 shrink-0 min-w-[320px]">
            <div>
              <h3 class="text-lg font-bold text-slate-800">Lampiran Dokumen</h3>
            </div>
            <div class="flex items-center gap-2">
              {#if docUrl}
                <a href={docUrl} target="_blank" rel="noreferrer" class="px-3 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 font-medium text-xs transition-colors shadow-sm flex items-center gap-1">
                  <span class="material-symbols-outlined text-[16px]">open_in_new</span>
                  Fullscreen
                </a>
              {/if}
              <button onclick={closeDoc} class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors">
                <span class="material-symbols-outlined text-[20px]">close</span>
              </button>
            </div>
          </div>

          <!-- Body -->
          <div class="flex-1 bg-slate-100 relative overflow-hidden flex items-center justify-center">
            {#if docUrl}
              {#if isImageFile(docUrl)}
                <img src={docUrl} alt="Lampiran" class="block max-w-full max-h-[calc(90vh-80px)] object-contain" />
              {:else}
                <iframe title="Lampiran" src={docUrl} class="w-full h-full border-0"></iframe>
              {/if}
            {:else}
              <div class="absolute inset-0 flex items-center justify-center text-slate-400">
                <p>Lampiran tidak tersedia.</p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}

<style>
  :global(body) {
    font-family: 'Geist', 'Inter', sans-serif;
    color: #0f172a;
  }



  /* .page-bg {
    min-height: 100vh;
    background-color: #f8fafc;
    background-image: radial-gradient(at 0% 0%, rgba(16, 185, 129, 0.03) 0%, transparent 50%),
                      radial-gradient(at 100% 100%, rgba(14, 165, 233, 0.03) 0%, transparent 50%);
    padding: 0px 0px;
  } */

  .container { max-width: 1400px; margin: 0 auto; }

  /* Header */
  /* .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px; } */
  /* .title { font-size: 28px; font-weight: 800; color: #0f172a; margin: 0 0 6px 0; letter-spacing: -0.02em; } */
  /* .subtitle { color: #64748b; font-size: 15px; margin: 0; } */
  .header-icon {
    width: 48px; height: 48px; background: #ffffff; border-radius: 12px;
    display: flex; align-items: center; justify-content: center; color: #10b981;
    box-shadow: 0 4px 12px rgba(0,0,0,0.03); border: 1px solid #e2e8f0;
  }

  /* Layout */
  .grid-layout {
    display: grid; grid-template-columns: 1fr 1fr; gap: 24px; margin-bottom: 32px;
  }
  @media (max-width: 768px) { .grid-layout { grid-template-columns: 1fr; } }

  /* Cards */
  .card {
    background: white; border-radius: 20px; border: 1px solid #e2e8f0;
    box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden;
  }
  .card-header {
    padding-bottom: 20px; display: flex; justify-content: space-between; align-items: center;
  }
  .card-header h3 { margin: 0; font-size: 20px; font-weight: 600; color: #1e293b; }
  .border-b { border-bottom: 1px solid #f1f5f9; }

  /* Main Action Card */
  .main-action-card { display: flex; flex-direction: column; justify-content: space-between; }
  .date-badge { 
    font-size: 12px; font-weight: 600; color: #64748b; background: #f1f5f9; 
    padding: 4px 10px; border-radius: 99px;
  }
  
  .status-display { padding: 24px; display: flex; justify-content: center; }
  .status-circle {
    width: 140px; height: 140px; border-radius: 50%;
    display: flex; flex-direction: column; align-items: center; justify-content: center;
    border: 4px solid;
  }
  .status-green { background: #ecfdf5; border-color: #10b981; color: #065f46; }
  .status-yellow { background: #fffbeb; border-color: #f59e0b; color: #92400e; }
  .status-gray { background: #f8fafc; border-color: #cbd5e1; color: #64748b; }
  .status-red { background: #fef2f2; border-color: #ef4444; color: #991b1b; }
  
  .status-text { font-size: 14px; font-weight: 600; text-transform: uppercase; margin-bottom: 4px; }
  .time-text { font-size: 28px; font-weight: 800; letter-spacing: -1px; }

  .action-buttons {
    padding: 24px; display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
    background: #f8fafc; border-top: 1px solid #f1f5f9;
  }
  .btn-checkin, .btn-checkout {
    display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
    padding: 16px; border-radius: 16px; border: none; cursor: pointer; font-weight: 600;
    transition: all 0.2s;
  }
  .btn-checkin { background: #10b981; color: white; box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); }
  .btn-checkin:hover:not(:disabled) { background: #059669; transform: translateY(-2px); }
  .btn-checkin:disabled { opacity: 0.5; cursor: not-allowed; background: #a7f3d0; box-shadow: none; }

  .btn-checkout { background: white; border: 2px solid #e2e8f0; color: #475569; }
  .btn-checkout:hover:not(:disabled) { border-color: #ef4444; color: #ef4444; background: #fef2f2; }
  .btn-checkout:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-icon { font-size: 20px; }

  /* Permission Form */
  .permission-card { padding: 0 0 24px 0; }
  .form-content { padding: 0 24px; }
  .form-group { margin-bottom: 16px; }
  .label { display: block; font-size: 13px; font-weight: 600; color: #334155; margin-bottom: 8px; }
  
  .radio-group { display: flex; gap: 12px; }
  .radio-btn {
    flex: 1; padding: 10px; border: 1px solid #e2e8f0; border-radius: 10px;
    text-align: center; cursor: pointer; font-size: 14px; font-weight: 500; color: #64748b;
    transition: all 0.2s;
  }
  .radio-btn.active { border-color: #10b981; background: #ecfdf5; color: #065f46; font-weight: 600; }

  .textarea {
    width: 100%; padding: 12px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-family: inherit; font-size: 14px; resize: vertical; box-sizing: border-box;
  }
  .textarea:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1); }

  .file-drop {
    display: block; width: 100%; padding: 16px; border: 2px dashed #cbd5e1; border-radius: 10px;
    text-align: center; cursor: pointer; color: #64748b; font-size: 13px; box-sizing: border-box;
    transition: all 0.2s;
  }
  .file-drop:hover { border-color: #10b981; background: #f0fdf4; }
  .file-name { color: #10b981; font-weight: 600; }

  .btn-submit {
    width: 100%; padding: 12px; background: #0f172a; color: white; border: none;
    border-radius: 10px; font-weight: 600; font-size: 14px; cursor: pointer;
    transition: all 0.2s;
  }
  .btn-submit:hover:not(:disabled) { background: #1e293b; }
  .btn-submit:disabled { opacity: 0.7; cursor: not-allowed; }

  /* Table */
  .table-container { overflow-x: auto; }
  .desktop-only { display:block; }
  .mobile-list { display:none; }
  .table { width: 100%; border-collapse: separate; border-spacing: 0; }
  @media (max-width: 768px) {
    .desktop-only { display:none; }
    .mobile-list { display:flex; flex-direction:column; gap:12px; }
  }
  
  .table th {
    text-align: left; padding: 14px 24px; font-size: 12px; font-weight: 600; 
    text-transform: uppercase; color: #64748b; background: #fcfcfc; border-bottom: 1px solid #e2e8f0;
  }
  .table td {
    padding: 16px 24px; border-bottom: 1px solid #f1f5f9; vertical-align: middle;
    font-size: 14px; color: #334155;
  }
  .hover-row:hover td { background-color: #f8fafc; }
  .action-cell { white-space: nowrap; display:flex; gap:8px; align-items:center; }
  .mini-btn {
    display:inline-flex; align-items:center; gap:6px;
    padding:6px 10px; border-radius:999px; border:1px solid #0f172a;
    background:#0f172a; color:#fff; font-weight:600; font-size:12px;
    cursor:pointer; transition:all 0.15s ease;
  }
  .mini-btn.icon-only { width:34px; height:34px; padding:0; justify-content:center; }
  .mini-btn.disabled { background:#e2e8f0; border-color:#e2e8f0; color:#94a3b8; cursor:not-allowed; }
  .mini-btn:hover { background:#111827; border-color:#111827; }
  .mini-btn.mobile { flex:1; justify-content:center; }
  .btn-text { display:none; }
  @media (max-width: 768px) {
    .mini-btn.mobile .btn-text { display:inline; font-weight:700; font-size:12px; }
    .mini-btn.mobile { height:38px; }
  }
  /* Mobile cards */
  .entry-card {
    padding: 14px;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    background: #ffffff;
    box-shadow: 0 6px 20px -18px rgba(15,23,42,0.3);
  }
  .entry-head { display:flex; align-items:center; justify-content:space-between; gap:10px; margin-bottom:12px; }
  .date-row { display:flex; align-items:center; gap:8px; color:#0f172a; font-weight:700; }
  .date-icon { color:#6366f1; }
  .date-text { font-size: 0.95rem; }
  .time-grid { display:grid; grid-template-columns: repeat(2, minmax(0,1fr)); gap:10px; margin-bottom:12px; }
  .time-box {
    padding:12px;
    border:1px solid #e2e8f0;
    border-radius:14px;
    background:#f8fafc;
    text-align:center;
  }

    .intern-grid { display:grid; gap:10px; margin-bottom:12px; }


    .intern-box {
    padding:12px;
    border:1px solid #e2e8f0;
    border-radius:14px;
    background:#f8fafc;
    text-align:center;
  }

    .intern-name-mobile {
    font-weight: 600;
    color: #334155;
  }
  
    .intern-box-label {
    margin:6px;
    font-size:18px;
    font-weight:700;
    padding-bottom: 20px;
    color:#0f172a;
    align-items: center;
    justify-content:center;
  }

  .time-box .label { margin:0; font-size:12px; font-weight:700; color:#94a3b8; text-transform:uppercase; letter-spacing:0.03em; }
  .time-value { margin:6px 0 0 0; font-size:18px; font-weight:800; color:#0f172a; letter-spacing:-0.02em; }
  .mobile-actions { display:flex; gap:8px; }
  
  .font-medium { font-weight: 500; }
  .text-slate-500 { color: #64748b; }

  .status-badge {
    display: inline-flex; align-items: center; padding: 4px 12px;
    border-radius: 99px; font-size: 12px; font-weight: 600; border: 1px solid transparent;
  }
  .equal-badge { min-width: 96px; text-align:center; justify-content:center; }
  .bg-emerald-100 { background: #ecfdf5; border-color: #a7f3d0; } .text-emerald-700 { color: #047857; }
  .bg-yellow-100 { background: #fefce8; border-color: #fef08a; } .text-yellow-700 { color: #a16207; }
  .bg-blue-100 { background: #eff6ff; border-color: #bfdbfe; } .text-blue-700 { color: #1d4ed8; }
  .bg-purple-100 { background: #faf5ff; border-color: #e9d5ff; } .text-purple-700 { color: #7e22ce; }
  .bg-slate-100 { background: #f1f5f9; border-color: #e2e8f0; } .text-slate-600 { color: #475569; }
  .bg-red-100 { background: #fef2f2; border-color: #fecaca; } .text-red-700 { color: #b91c1c; }

  .user-info { display: flex; align-items: center; gap: 10px; }
  .avatar-mini {
    width: 28px; height: 28px; background: #0f172a; color: white;
    border-radius: 50%; display: flex; align-items: center; justify-content: center;
    font-size: 11px; font-weight: 600;
  }
  .badge-count { background: #f1f5f9; color: #64748b; padding: 4px 10px; border-radius: 20px; font-size: 14px; font-weight: 600; }

  /* States */
  .empty-state, .loading-state { text-align: center; padding: 40px; color: #94a3b8; font-style: italic; }

  /* Utils */
  .spinner-small { width: 14px; height: 14px; border: 2px solid white; border-top-color: transparent; border-radius: 50%; animation: spin 1s infinite linear; display: inline-block; }
  .spinner-small.dark { border-color: #64748b; border-top-color: transparent; }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Animations */
  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s ease-out forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

</style>
