<script>
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';
  import { goto } from '@mateothegreat/svelte5-router';
  import { location } from '../lib/location.js';
  import { isMobileSidebarOpen } from '../lib/ui.store.js';
  import { onMount } from 'svelte';

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
    { match: '/calendar', label: 'Kalender' },
  ];

  const path = $derived($location.path || '/dashboard');
  const currentTitle = $derived(
    titles.find((t) => path === t.match || path.startsWith(t.match + '/'))?.label || 'Dashboard'
  );
  const role = $derived(auth.user?.role || 'intern');
  const displayName = $derived(auth.user?.name || auth.user?.email || 'Anda');

  let notifications = $state([]);
  let unreadCount = $state(0);
  let notifDropdownOpen = $state(false);
  let userDropdownOpen = $state(false);

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
    if (routePath.startsWith('/calendar')) {
      return 'Ringkasan jadwal tugas dan presensi.';
    }
    if (routePath.startsWith('/data-tools')) {
      return 'Ekspor, impor, dan sinkronisasi data.';
    }
    return 'Kelola aktivitas magang dengan mudah.';
  }

  const currentSubtitle = $derived(getSubtitle(path, role, displayName));

  async function fetchNotifications() {
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

  async function handleLogout() {
    await api.logout();
    goto('/login');
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
    
    return () => {
      clearInterval(interval);
      document.removeEventListener('click', handleClickOutside);
    };
  });
</script>

<div class="topbar">
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
      <button 
        id="notif-dropdown-btn"
        class="notif-btn" 
        onclick={toggleNotifDropdown}
        aria-label="Notifications"
      >
        <svg viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
          <path d="M18 8a6 6 0 1 0-12 0c0 7-3 8-3 8h18s-3-1-3-8" />
          <path d="M13.7 21a2 2 0 0 1-3.4 0" />
        </svg>
        {#if unreadCount > 0}
          <span class="notif-badge">{unreadCount > 9 ? '9+' : unreadCount}</span>
        {/if}
      </button>
      
      {#if notifDropdownOpen}
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
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M18 8a6 6 0 1 0-12 0c0 7-3 8-3 8h18s-3-1-3-8" />
                  <path d="M13.7 21a2 2 0 0 1-3.4 0" />
                  <line x1="5" y1="5" x2="19" y2="19" />
                </svg>
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
      {/if}
    </div>

    <!-- User Menu -->
    <div class="dropdown-container">
      <button 
        id="user-dropdown-btn"
        class="user-menu-btn" 
        onclick={toggleUserDropdown}
      >
        <div class="user-avatar">
          {auth.user?.name ? auth.user.name.charAt(0).toUpperCase() : 'U'}
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
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="8" r="3.5" />
              <path d="M4 20c0-3.5 3.5-6 8-6s8 2.5 8 6" />
            </svg>
            <span>Profil Saya</span>
          </a>
          <a href="/settings" class="dropdown-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="3" />
              <path d="M19 12h2M3 12h2M12 19v2M12 3v2M17 17l1.5 1.5M5.5 5.5L7 7M17 7l1.5-1.5M5.5 18.5L7 17" />
            </svg>
            <span>Pengaturan</span>
          </a>
          <div class="dropdown-divider"></div>
          <button class="dropdown-item logout-item" onclick={handleLogout}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" y1="12" x2="9" y2="12" />
            </svg>
            <span>Logout</span>
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>