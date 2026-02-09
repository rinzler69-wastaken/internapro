<script>
  import { route } from '@mateothegreat/svelte5-router';
  import { auth } from '../lib/auth.svelte.js';
  import { location } from '../lib/location.js';
  import { isSidebarCollapsed, isMobileSidebarOpen } from '../lib/ui.store.js';
  

  const path = $derived($location.path || '/');
  const role = $derived(auth.user?.role || 'intern');
  const canManage = $derived(role === 'admin' || role === 'pembimbing' || role === 'supervisor');
  const collapsed = $derived($isSidebarCollapsed);
  const mobileOpen = $derived($isMobileSidebarOpen);

    const statusLabels = {
    present: 'Hadir',
    late: 'Terlambat',
    absent: 'Tidak Hadir',
    sick: 'Sakit',
    permission: 'Izin',
  };

  let isMobile = $state(false);

  $effect(() => {
    const checkMobile = () => {
      const wasMobile = isMobile;
      isMobile = window.innerWidth <= 900;
      
      // When switching to mobile, ensure sidebar is expanded
      if (isMobile && !wasMobile) {
        isSidebarCollapsed.set(false);
      }
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  });

  // Click outside to close mobile sidebar
  $effect(() => {
    if (!isMobile || !mobileOpen) return;

    const handleClickOutside = (e) => {
      const sidebar = document.querySelector('.sidebar');
      if (sidebar && !sidebar.contains(e.target)) {
        isMobileSidebarOpen.set(false);
      }
    };

    // Add a small delay to prevent immediate closing when opening
    const timeoutId = setTimeout(() => {
      document.addEventListener('click', handleClickOutside);
    }, 100);

    return () => {
      clearTimeout(timeoutId);
      document.removeEventListener('click', handleClickOutside);
    };
  });

  function toggleCollapse() {
    if (isMobile) {
      // On mobile, close the sidebar
      isMobileSidebarOpen.set(false);
    } else {
      // On desktop, toggle collapse state
      isSidebarCollapsed.update((value) => !value);
    }
  }

  function closeMobileSidebar() {
    isMobileSidebarOpen.set(false);
  }
</script>

<svelte:head>
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0" />
</svelte:head>

<style>
  .material-symbols-outlined {

    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: lighter;
  }

  .sidebar {
    z-index: 50;
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  
  .sidebar.open {
    transform: translateX(0);
    box-shadow: 4px 0 12px rgba(0, 0, 0, 0.15);
  }

  /* Dark overlay on main content when sidebar is open */
  .sidebar.open ~ .main-content::before {
    content: "";
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 40;
    /* animation: fadeIn 0.25s ease; */
  }
</style>

<aside class={`sidebar ${!isMobile && collapsed ? 'collapsed' : ''} ${mobileOpen ? 'open' : ''}`}>
  <div class="sidebar-brand">
    <button
      class="sidebar-brand-icon"
      onclick={toggleCollapse}
      type="button"
      aria-label={isMobile ? 'Close sidebar' : (collapsed ? 'Expand sidebar' : 'Sidebar brand')}
      disabled={!collapsed && !isMobile}
    >
      {#if collapsed && !isMobile}
        <!-- Right arrow for collapsed desktop -->
        <span class="material-symbols-outlined">chevron_right</span>
      {:else}
        <!-- Logo for expanded desktop and all mobile states -->
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
          <path d="M3 7.5l9-4 9 4-9 4-9-4z" />
          <path d="M6 10.5v4.5c0 2 3 3.5 6 3.5s6-1.5 6-3.5v-4.5" />
        </svg>
      {/if}
    </button>
    <div class="sidebar-brand-text">
      <h1>InternaPro</h1>
        <span class="role-text">version 1.0-DSI</span>
        

    </div>
    {#if isMobile || !collapsed}
      <button class="sidebar-brand-collapse" onclick={toggleCollapse} type="button" aria-label={isMobile ? "Close sidebar" : "Collapse sidebar"}>
        <span class="material-symbols-outlined">chevron_left</span>
      </button>
    {/if}
  </div>

  <div class="nav-section">
    <div class="nav-section-title">
      <span class="nav-icon section-title-dots" aria-hidden="true">
        <span class="dot"></span>
        <span class="dot"></span>
        <span class="dot"></span>
      </span>
      <span class="section-title-text">Menu</span>
    </div>
    <a href="/dashboard" use:route onclick={closeMobileSidebar} class={`nav-link ${path === '/dashboard' ? 'active' : ''}`} data-tooltip="Dashboard">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">readiness_score</span>
      <span>Dashboard</span>
    </a>
    {#if canManage}
      <a href="/interns" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/interns') ? 'active' : ''}`} data-tooltip="Anggota Magang">
        <span class="nav-icon material-symbols-outlined" aria-hidden="true">groups</span>
        <span>Anggota Magang</span>
      </a>
    {/if}
    <a href="/tasks" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/tasks') ? 'active' : ''}`} data-tooltip="Penugasan">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">assignment</span>
      <span>Daftar Penugasan</span>
    </a>
    {#if canManage}
      <a href="/task-assignments" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/task-assignments') ? 'active' : ''}`} data-tooltip="Daftar Tugas">
        <span class="nav-icon material-symbols-outlined" aria-hidden="true">assignment_ind</span>
        <span>Daftar Tugas</span>
      </a>
    {/if}
    <a href="/attendance" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/attendance') ? 'active' : ''}`} data-tooltip="Presensi">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">punch_clock</span>
      <span>Riwayat Presensi</span>
    </a>
    <a href="/calendar" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/calendar') ? 'active' : ''}`} data-tooltip="Kalender">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">calendar_today</span>
      <span>Kalender</span>
    </a>
  </div>

  {#if canManage}
    <div class="nav-section">
    <div class="nav-section-title">
      <span class="nav-icon section-title-dots" aria-hidden="true">
        <span class="dot"></span>
        <span class="dot"></span>
        <span class="dot"></span>
      </span>
      <span class="section-title-text">Evaluasi</span>
    </div>
    <a href="/reports" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/reports') ? 'active' : ''}`} data-tooltip="Laporan">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">description</span>
      <span>Laporan</span>
    </a>
    <a href="/assessments" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/assessments') ? 'active' : ''}`} data-tooltip="Penilaian">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">star</span>
      <span>Penilaian</span>
    </a>
  </div>
  {/if}

  <div class="nav-section">
    <div class="nav-section-title">
      <span class="nav-icon section-title-dots" aria-hidden="true">
        <span class="dot"></span>
        <span class="dot"></span>
        <span class="dot"></span>
      </span>
      <span class="section-title-text">Aktivitas</span>
    </div>
    <a href="/notifications" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/notifications') ? 'active' : ''}`} data-tooltip="Notifikasi">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">notifications</span>
      <span>Notifikasi</span>
    </a>
    <a href="/profile" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/profile') ? 'active' : ''}`} data-tooltip="Profil">
      <span class="nav-icon material-symbols-outlined" aria-hidden="true">person</span>
      <span>Profil</span>
    </a>
  </div>

  {#if role === 'admin'}
    <div class="nav-section">
      <div class="nav-section-title">
        <span class="nav-icon section-title-dots" aria-hidden="true">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
        </span>
        <span class="section-title-text">Sistem</span>
      </div>
      <a href="/supervisors" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/supervisors') ? 'active' : ''}`} data-tooltip="Pembimbing">
        <span class="nav-icon material-symbols-outlined" aria-hidden="true">supervisor_account</span>
        <span>Pembimbing</span>
      </a>
      <a href="/settings" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/settings') ? 'active' : ''}`} data-tooltip="Pengaturan">
        <span class="nav-icon material-symbols-outlined" aria-hidden="true">settings</span>
        <span>Pengaturan</span>
      </a>
    </div>
  {/if}

  {#if canManage}
    <div class="nav-section">
      <div class="nav-section-title">
        <span class="nav-icon section-title-dots" aria-hidden="true">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
        </span>
        <span class="section-title-text">Data</span>
      </div>
      <a href="/data-tools" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/data-tools') ? 'active' : ''}`} data-tooltip="Export/Import">
        <span class="nav-icon material-symbols-outlined" aria-hidden="true">import_export</span>
        <span>Export/Import</span>
      </a>
    </div>
  {/if}

</aside>