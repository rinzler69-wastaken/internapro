<script>
  import { onMount } from "svelte";
  import Sidebar from "./Sidebar.svelte";
  import Topbar from "./Topbar.svelte";
  import { isSidebarCollapsed } from "../lib/ui.store.js";

  const { children } = $props();
  let isMobile = $state(false);

  onMount(() => {
    try {
      const stored = localStorage.getItem("sidebar_collapsed");
      if (stored !== null) {
        isSidebarCollapsed.set(stored === "1");
      }
    } catch (err) {
      // ignore storage errors (e.g., privacy mode)
    }

    const mediaQuery = window.matchMedia("(max-width: 900px)");
    isMobile = mediaQuery.matches;
    mediaQuery.addEventListener("change", (e) => {
      isMobile = e.matches;
    });
  });

  $effect(() => {
    try {
      localStorage.setItem(
        "sidebar_collapsed",
        $isSidebarCollapsed ? "1" : "0",
      );
    } catch (err) {
      // ignore storage errors
    }
  });
</script>

<div
  class="app-container font-inter"
  class:sidebar-collapsed={!isMobile && $isSidebarCollapsed}
>
  <Sidebar />
  <main class="main-content">
    <Topbar />
    {@render children?.()}
  </main>
</div>

<div id="overlay-root"></div>

<style>
  .app-container {
    min-height: 100vh;
  }

  #overlay-root {
    position: fixed;
    inset: 0;
    z-index: 2147483647; /* max 32-bit z-index */
    pointer-events: none;
  }

  .main-content {
    /* Space for fixed topbar at the top */
    padding-top: 108px;

    /* Smooth transition when sidebar expands/collapses */
    transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);

    min-height: 100vh;
  }

  /* Desktop: offset for expanded sidebar */
  @media (min-width: 901px) {
    .main-content {
      margin-left: 256px; /* Expanded sidebar width */
    }

    /* When sidebar is collapsed */
    .app-container.sidebar-collapsed .main-content {
      margin-left: 80px; /* Collapsed sidebar width */
    }
  }

  /* Mobile: no sidebar offset */
  @media (max-width: 900px) {
    .main-content {
      margin-left: 0;
    }
  }
</style>
