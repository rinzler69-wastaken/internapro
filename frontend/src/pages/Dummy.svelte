<script>
  import { auth } from '../lib/auth.svelte.js';
  import AdminDashboard from './AdminDashboard.svelte';
  import InternDashboard from './InternDashboard.svelte';
  import IncompleteProfile from './IncompleteProfile.svelte';

  const userRole = $derived(auth.user?.role);
  // Optional: Check specific intern profile fields if needed
  // const hasInternProfile = $derived(!!(auth.user?.intern_id || auth.user?.internId));

  $effect(() => {
    console.log('Dashboard Auth State:', { user: auth.user, role: userRole });
  });
</script>

{#if userRole === 'admin' || userRole === 'supervisor'}
  <AdminDashboard />
{:else if userRole === 'intern'}
  <!-- Direct access to InternDashboard as requested -->
  <InternDashboard />
{:else}
  <div class="p-8 text-center flex justify-center items-center h-[50vh]">
    <div class="animate-pulse text-slate-400">Loading dashboard...</div>
  </div>
{/if}