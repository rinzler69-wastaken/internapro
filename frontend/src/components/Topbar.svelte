<script>
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';
  import { location } from '../lib/location.js';
  import { isMobileSidebarOpen, isSidebarCollapsed } from '../lib/ui.store.js';
  import { onMount } from 'svelte';
  
  const collapsed = $derived($isSidebarCollapsed);

  const titles = [
    { match: '/dashboard', label: 'Dashboard' },
    { match: '/tasks', label: 'Penugasan' },
    { match: '/task-assignments', label: 'Daftar Tugas' },
    { match: '/attendance', label: 'Presensi' },
    { match: '/assessments', label: 'Penilaian' },
    { match: '/reports', label: 'Laporan' },
    { match: '/notifications', label: 'Notifikasi' },
    { match: '/settings', label: 'Pengaturan' },
    { match: '/interns', label: 'Anggota Magang' },
    { match: '/supervisors', label: 'Pembimbing' },
    { match: '/data-tools', label: 'Export/Import' },
    { match: '/profile', label: 'Profil' },
    { match: '/profile/edit', label: 'Edit Profil' },
    { match: '/calendar', label: 'Kalender' },
  ];

  const path = $derived($location.path || '/dashboard');
  const currentTitle = $derived(
    titles.find((t) => path === t.match || path.startsWith(t.match + '/'))?.label || 'Dashboard'
  );
  const role = $derived(auth.user?.role || 'intern');
  const settingsHref = $derived(
    role === 'intern' || role === 'supervisor' || role === 'pembimbing' ? '/profile/edit' : '/settings'
  );
  const displayName = $derived(auth.user?.name || auth.user?.email || 'Anda');
  const avatarSrc = $derived(buildAvatarUrl(auth.user?.avatar));

  let notifications = $state([]);
  let unreadCount = $state(0);
  let notifDropdownOpen = $state(false);
  let userDropdownOpen = $state(false);
  let isScrolled = $state(false);
  let showLogoutModal = $state(false);

  function getSubtitle(routePath, userRole, name) {
    if (routePath === '/' || routePath === '/dashboard') {
      return `Pantau aktivitas magang anda dalam sekilas.`;
    }
    if (routePath.startsWith('/notifications')) {
      return 'Riwayat notifikasi anda.';
    }
    if (routePath.startsWith('/attendance')) {
      return userRole === 'intern'
        ? 'Tinjau kehadiran anda selama periode magang.'
        : 'Tinjau dan kelola kehadiran siswa magang.';
    }
    if (routePath.startsWith('/tasks')) {
      return userRole === 'intern' ? 'Pantau tugas dan deadline pribadi.' : 'Kelola penugasan untuk siswa magang.';
    }
    if (routePath.startsWith('/task-assignments')) {
      return 'Atur penugasan massal dan distribusikan ke siswa.';
    }
    if (routePath.startsWith('/assessments')) {
      return userRole === 'intern' ? 'Lihat penilaian kinerja anda.' : 'Evaluasi performa siswa magang.';
    }
    if (routePath.startsWith('/reports')) {
      return userRole === 'intern' ? 'Kirim dan kelola laporan kegiatan.' : 'Tinjau laporan periodik siswa.';
    }
    if (routePath.startsWith('/interns')) {
      return 'Daftar siswa magang aktif dan statusnya.';
    }
    if (routePath.startsWith('/supervisors')) {
      return 'Kelola pembimbing dan status persetujuan.';
    }
    if (routePath.startsWith('/settings')) {
      return 'Atur konfigurasi aplikasi dan keamanan akun.';
    }
    if (routePath.startsWith('/profile')) {
      return 'Perbarui informasi akun dan keamanan.';
    }
    if (routePath.startsWith('/profile/edit')) {
      return `Perbarui informasi pribadi dan preferensi, ${name}.`;
    }
    if (routePath.startsWith('/calendar')) {
      return 'Ringkasan jadwal tugas dan presensi.';
    }
    if (routePath.startsWith('/data-tools')) {
      return 'Ekspor, impor, dan sinkronisasi data.';
    }
    return 'Kelola aktivitas magang dengan mudah.';
  }

  const currentSubtitle = $derived(getSubtitle(path, role, displayName));

  function buildAvatarUrl(path) {
    if (!path) return '';
    // Pass through external URLs (e.g., Google avatar)
    if (path.startsWith('http')) return path;
    const clean = path.startsWith('/uploads/') ? path : `/uploads/${path}`;
    const token = auth.token;
    const qs = token ? `${clean.includes('?') ? '&' : '?'}token=${token}` : '';
    return `${clean}${qs}`;
  }

  async function fetchNotifications() {
    if (!auth.token || !auth.user) return;
    // Skip fetching for interns without profile to avoid 500s
    if (auth.user.role === 'intern' && !auth.user.intern_id && !auth.user.internId) return;
    try {
      const res = await api.getNotifications({ page: 1, limit: 5 });
      notifications = res.data || [];
      unreadCount = notifications.filter(n => !n.read_at).length;
    } catch (err) {
      console.error('Failed to fetch notifications:', err);
    }
  }

  async function markAllRead() {
    try {
      await api.markAllNotificationsRead();
      await fetchNotifications();
    } catch (err) {
      console.error('Failed to mark all read:', err);
    }
  }

  function handleLogout() {
    userDropdownOpen = false;
    showLogoutModal = true;
  }

  async function confirmLogout() {
    try {
      await api.logout();
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      showLogoutModal = false;
      goto('/login');
    }
  }

  function toggleMobileSidebar() {
    isMobileSidebarOpen.update(v => !v);
  }

  function toggleNotifDropdown() {
    notifDropdownOpen = !notifDropdownOpen;
    if (notifDropdownOpen) {
      userDropdownOpen = false;
    }
  }

  function toggleUserDropdown() {
    userDropdownOpen = !userDropdownOpen;
    if (userDropdownOpen) {
      notifDropdownOpen = false;
    }
  }

  function closeDropdowns() {
    notifDropdownOpen = false;
    userDropdownOpen = false;
  }

  function handleClickOutside(e) {
    const notifBtn = document.getElementById('notif-dropdown-btn');
    const notifMenu = document.getElementById('notif-dropdown-menu');
    const userBtn = document.getElementById('user-dropdown-btn');
    const userMenu = document.getElementById('user-dropdown-menu');

    if (notifBtn && notifMenu && !notifBtn.contains(e.target) && !notifMenu.contains(e.target)) {
      notifDropdownOpen = false;
    }
    if (userBtn && userMenu && !userBtn.contains(e.target) && !userMenu.contains(e.target)) {
      userDropdownOpen = false;
    }
  }

  onMount(() => {
    fetchNotifications();
    const interval = setInterval(fetchNotifications, 60000); // Refresh every minute
    document.addEventListener('click', handleClickOutside);
    
    const handleScroll = () => {
      isScrolled = window.scrollY > 0;
    };
    window.addEventListener('scroll', handleScroll);
    
    return () => {
      clearInterval(interval);
      document.removeEventListener('click', handleClickOutside);
      window.removeEventListener('scroll', handleScroll);
    };
  });

  function handleKeydown(e) {
    if (e.key === 'Escape' && showLogoutModal) {
      showLogoutModal = false;
    }
  }
</script>

<svelte:head>
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0" />
</svelte:head>

<style>
  .material-symbols-outlined {
    font-weight: lighter;
  }

  .topbar {
    position: fixed;
    top: 0;
    padding-top: 16px;
    z-index: 40;
    
    /* Dynamic positioning based on sidebar state */
    left: 0;
    right: 0;
    
    background-color: rgba(255, 255, 255, 0.95);
    /* backdrop-filter: blur(8px); */
    
    /* Add border-bottom to align with sidebar separator */
    border-bottom: 1px solid rgba(229, 231, 235, 1);
    
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .topbar.scrolled {
    background-color: rgba(255, 255, 255, 0.98);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
                0 2px 4px -1px rgba(0, 0, 0, 0.06);
  }

  /* Adjust topbar position when sidebar is expanded (desktop only) */
  @media (min-width: 901px) {
    .topbar {
      left: 256px; /* Width of expanded sidebar */
    }
    
    .topbar.sidebar-collapsed {
      left: 72px; /* Width of collapsed sidebar */
    }
  }

  /* Mobile: topbar always full width */
  @media (max-width: 900px) {
    .topbar {
      left: 0;
    }
  }

  @keyframes scaleIn {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
  }
  .animate-scale-in {
    animation: scaleIn 0.2s ease-out forwards;
  }

  .user-menu-btn {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    cursor: pointer;
    box-shadow: 0 6px 20px -18px rgba(15, 23, 42, 0.25);
  }
  .user-menu-btn:hover { border-color: #cbd5e1; }

  .user-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: #0f172a;
    color: #fff;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 14px;
    overflow: hidden;
    flex-shrink: 0;
  }
  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
</style>

<svelte:window onkeydown={handleKeydown} />

<div class="topbar {collapsed ? 'sidebar-collapsed' : ''}" class:scrolled={isScrolled}>
  <div class="topbar-title">
    <button class="hamburger-btn" onclick={toggleMobileSidebar} aria-label="Toggle menu">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="3" y1="12" x2="21" y2="12"></line>
        <line x1="3" y1="6" x2="21" y2="6"></line>
        <line x1="3" y1="18" x2="21" y2="18"></line>
      </svg>
    </button>
    <h2>{currentTitle}</h2>
    <p class="topbar-subtitle">{currentSubtitle}</p>
  </div>

  <div class="topbar-actions">
    <!-- Notification Bell -->
    <div class="dropdown-container">
      <!-- <button 
        id="notif-dropdown-btn"
        class="notif-btn" 
        onclick={toggleNotifDropdown}
        aria-label="Notifications"
      >
        <span><span class="material-symbols-outlined mt-1">notifications</span></span>
        {#if unreadCount > 0}
          <span class="notif-badge">{unreadCount > 9 ? '9+' : unreadCount}</span>
        {/if}
      </button> -->
      
      <!-- {#if notifDropdownOpen}
        <div id="notif-dropdown-menu" class="dropdown-menu notif-dropdown">
          <div class="dropdown-header">
            <span class="dropdown-title">Notifikasi</span>
            {#if unreadCount > 0}
              <button class="mark-read-btn" onclick={markAllRead}>Tandai dibaca</button>
            {/if}
          </div>
          <div class="dropdown-body">
            {#if notifications.length === 0}
              <div class="empty-notif">
                <span class="material-symbols-outlined text-4xl">notifications_off</span>
                <span>Tidak ada notifikasi</span>
              </div>
            {:else}
              {#each notifications as notif}
                <a href={notif.link || '/notifications'} class="notif-item {!notif.read_at ? 'unread' : ''}">
                  <div class="notif-icon">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                      <circle cx="12" cy="12" r="10" />
                    </svg>
                  </div>
                  <div class="notif-content">
                    <div class="notif-title">{notif.title || 'Notifikasi'}</div>
                    <div class="notif-message">{notif.message}</div>
                    <div class="notif-time">{notif.created_at ? new Date(notif.created_at).toLocaleDateString('id-ID') : ''}</div>
                  </div>
                </a>
              {/each}
            {/if}
          </div>
          <a href="/notifications" class="dropdown-footer">
            Lihat Semua
          </a>
        </div>
      {/if} -->
    </div>

    <!-- User Menu -->
    <div class="dropdown-container">
      <button 
        id="user-dropdown-btn"
        class="user-menu-btn" 
        onclick={toggleUserDropdown}
      >
        <div class="user-avatar">
          {#if avatarSrc}
            <img src={avatarSrc} alt="Avatar" />
          {:else}
            {auth.user?.name ? auth.user.name.charAt(0).toUpperCase() : 'U'}
          {/if}
        </div>
        <div class="user-info">
          <span class="user-name">{auth.user?.name || auth.user?.email}</span>
          <span class="user-role-badge">{auth.user?.role || 'intern'}</span>
        </div>
        <svg class="user-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>

      {#if userDropdownOpen}
        <div id="user-dropdown-menu" class="dropdown-menu user-dropdown">
          <a href="/profile" class="dropdown-item">
            <span class="material-symbols-outlined">person</span>
            <span>Profil Saya</span>
          </a>
          <a href={settingsHref} class="dropdown-item">
            <span class="material-symbols-outlined">settings</span>
            <span>Pengaturan</span>
          </a>
          <div class="dropdown-divider"></div>
          <button class="dropdown-item logout-item" onclick={handleLogout}>
            <span class="material-symbols-outlined ml-0.5">logout</span>
            <span>Logout</span>
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if showLogoutModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[60] flex items-center justify-center p-4 sm:p-6" role="dialog" aria-modal="true">
    <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity" onclick={() => showLogoutModal = false}></div>
    <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-sm p-6 flex flex-col items-center text-center overflow-hidden">
      
      <div class="w-12 h-12 rounded-full bg-slate-50 text-slate-500 flex items-center justify-center mb-4">
           <span class="material-symbols-outlined">logout</span>
      </div>

      <h3 class="text-lg font-bold text-slate-800 mb-2">Konfirmasi Logout</h3>
      <p class="text-slate-500 text-sm mb-6">Apakah anda yakin ingin logout?</p>

      <div class="flex gap-3 w-full">
          <button onclick={() => showLogoutModal = false} class="flex-1 px-4 py-2.5 bg-white border border-slate-200 text-slate-700 rounded-xl hover:bg-slate-50 font-medium text-sm transition-all flex items-center justify-center gap-2 cursor-pointer">
              <span class="material-symbols-outlined text-[18px] cursor-pointer">close</span>
              Tidak
          </button>
          <button onclick={confirmLogout} class="flex-1 px-4 py-2.5 bg-white border border-slate-200 text-slate-700 rounded-xl hover:bg-slate-50 font-medium text-sm transition-all flex items-center justify-center gap-2 cursor-pointer">
              <span class="material-symbols-outlined text-[18px] cursor-pointer">check</span>
              Ya
          </button>
      </div>
    </div>
  </div>
{/if}
