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
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
          <path d="M9 6l6 6-6 6" />
        </svg>
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
        <span class="role-text">CONSOLE</span>
        

    </div>
    {#if isMobile || !collapsed}
      <button class="sidebar-brand-collapse" onclick={toggleCollapse} type="button" aria-label={isMobile ? "Close sidebar" : "Collapse sidebar"}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
          <path d="M15 6l-6 6 6 6" />
        </svg>
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
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M3 10.5L12 3l9 7.5" />
          <path d="M5 9.5V20a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1V9.5" />
        </svg>
      </span>
      <span>Dashboard</span>
    </a>
    {#if canManage}
      <a href="/interns" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/interns') ? 'active' : ''}`} data-tooltip="Anggota Magang">
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M16 11a4 4 0 1 0-8 0a4 4 0 0 0 8 0z" />
            <path d="M4 21v-1a6 6 0 0 1 12 0v1" />
          </svg>
        </span>
        <span>Anggota Magang</span>
      </a>
    {/if}
    <a href="/tasks" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/tasks') ? 'active' : ''}`} data-tooltip="Penugasan">
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3.5" y="3.5" width="17" height="17" rx="2" />
          <path d="M8 12l3 3 5-6" />
        </svg>
      </span>
      <span>Penugasan</span>
    </a>
    {#if canManage}
      <a href="/task-assignments" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/task-assignments') ? 'active' : ''}`} data-tooltip="Daftar Tugas">
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M12 3l8 4-8 4-8-4 8-4z" />
            <path d="M4 13l8 4 8-4" />
          </svg>
        </span>
        <span>Daftar Tugas</span>
      </a>
    {/if}
    <a href="/attendance" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/attendance') ? 'active' : ''}`} data-tooltip="Presensi">
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3.5" y="5.5" width="17" height="15" rx="2" />
          <path d="M7 3.5v4" />
          <path d="M17 3.5v4" />
          <path d="M8 13l2.5 2.5 5-5" />
        </svg>
      </span>
      <span>Presensi</span>
    </a>
    <a href="/calendar" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/calendar') ? 'active' : ''}`} data-tooltip="Kalender">
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3.5" y="5.5" width="17" height="15" rx="2" />
          <path d="M7 3.5v4" />
          <path d="M17 3.5v4" />
          <path d="M3.5 9.5h17" />
        </svg>
      </span>
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
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M6 3.5h8l4 4v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-14a2 2 0 0 1 2-2z" />
          <path d="M14 3.5v4h4" />
          <path d="M8 13h8" />
          <path d="M8 17h6" />
        </svg>
      </span>
      <span>Laporan</span>
    </a>
    <a href="/assessments" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/assessments') ? 'active' : ''}`} data-tooltip="Penilaian">
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M12 3.5l2.6 5.3 5.9.9-4.2 4.1 1 5.9-5.3-2.8-5.3 2.8 1-5.9-4.2-4.1 5.9-.9z" />
        </svg>
      </span>
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
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M18 8a6 6 0 1 0-12 0c0 7-3 8-3 8h18s-3-1-3-8" />
          <path d="M13.7 21a2 2 0 0 1-3.4 0" />
        </svg>
      </span>
      <span>Notifikasi</span>
    </a>
    <a href="/profile" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/profile') ? 'active' : ''}`} data-tooltip="Profil">
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <circle cx="12" cy="8" r="3.5" />
          <path d="M4 20c0-3.5 3.5-6 8-6s8 2.5 8 6" />
        </svg>
      </span>
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
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <circle cx="8" cy="8" r="3.5" />
            <path d="M2.5 20c0-3.5 3-6 6.5-6" />
            <circle cx="17" cy="9" r="3" />
            <path d="M13 20c0-2.8 2.4-4.8 5.5-4.8" />
          </svg>
        </span>
        <span>Pembimbing</span>
      </a>
      <a href="/settings" use:route onclick={closeMobileSidebar} class={`nav-link ${path.startsWith('/settings') ? 'active' : ''}`} data-tooltip="Pengaturan">
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <circle cx="12" cy="12" r="3" />
            <path d="M19 12h2M3 12h2M12 19v2M12 3v2M17 17l1.5 1.5M5.5 5.5L7 7M17 7l1.5-1.5M5.5 18.5L7 17" />
          </svg>
        </span>
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
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M4 5.5h16" />
            <path d="M4 12h16" />
            <path d="M4 18.5h16" />
          </svg>
        </span>
        <span>Export/Import</span>
      </a>
    </div>
  {/if}

</aside>