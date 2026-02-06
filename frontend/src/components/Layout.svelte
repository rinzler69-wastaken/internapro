<script>
  import { onMount } from 'svelte';
  import Sidebar from './Sidebar.svelte';
  import Topbar from './Topbar.svelte';
  import { isSidebarCollapsed } from '../lib/ui.store.js';

  const { children } = $props();
  let isMobile = false;

  onMount(() => {
    try {
      const stored = localStorage.getItem('sidebar_collapsed');
      if (stored !== null) {
        isSidebarCollapsed.set(stored === '1');
      }
    } catch (err) {
      // ignore storage errors (e.g., privacy mode)
    }

    const mediaQuery = window.matchMedia('(max-width: 900px)');
    isMobile = mediaQuery.matches;
    mediaQuery.addEventListener('change', (e) => {
      isMobile = e.matches;
    });
  });

  $effect(() => {
    try {
      localStorage.setItem('sidebar_collapsed', $isSidebarCollapsed ? '1' : '0');
    } catch (err) {
      // ignore storage errors
    }
  });
</script>

<div class="app-container font-inter" class:sidebar-collapsed={!isMobile && $isSidebarCollapsed}>
  <Sidebar />
  <main class="main-content">
    <Topbar />
    {@render children?.()}
  </main>
</div>
