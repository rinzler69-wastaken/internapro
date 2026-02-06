<script>
  import { auth } from '../lib/auth.svelte.js';
  import AdminDashboard from './AdminDashboard.svelte';
  import InternDashboard from './InternDashboard.svelte';
  import IncompleteProfile from './IncompleteProfile.svelte';

  const userRole = $derived(auth.user?.role);
  const hasInternProfile = $derived(!!(auth.user?.intern_id || auth.user?.internId));

  $effect(() => {
    console.log('Dashboard Auth State:', { user: auth.user, role: userRole, hasInternProfile });
  });
</script>

{#if userRole === 'admin' || userRole === 'supervisor'}
  <AdminDashboard />
{:else if userRole === 'intern'}
<InternDashboard />
  <!-- {#if hasInternProfile}
    <InternDashboard />
  {:else}
    <IncompleteProfile /> -->
  <!-- {/if} -->
{:else}
  <div class="p-8 text-center">
    <p class="text-slate-500">Loading dashboard...</p>
  </div>
{/if}