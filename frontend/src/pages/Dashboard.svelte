<script>
  import { onMount } from 'svelte';
  import { replace } from "@mateothegreat/svelte5-router";
  import { auth } from '../lib/auth.svelte.js';
  import { api } from '../lib/api.js';
  import AdminDashboard from './AdminDashboard.svelte';
  import InternDashboard from './InternDashboard.svelte';
  import IncompleteProfile from './IncompleteProfile.svelte';
  import WaitingApproval from './WaitingApproval.svelte';

  const userRole = $derived(auth.user?.role);
  const isManager = $derived(['admin', 'supervisor', 'pembimbing'].includes(userRole));

  let loading = $state(false);
  let internProfile = $state(null);

  onMount(fetchProfileIfNeeded);

  async function fetchProfileIfNeeded() {
    if (userRole !== 'intern') return;
    loading = true;
    try {
      const res = await api.getProfile();
      const user = res.data?.user || res.data || auth.user;
      internProfile = res.data?.intern || null;
      if (auth.hydrate && user) auth.hydrate(user);
    } catch (err) {
      console.error('Failed to load profile:', err);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    console.log('Dashboard Auth State:', { user: auth.user, role: userRole, internProfile });
    if (!loading && internProfile?.status === 'pending') {
      replace('/waiting-approval');
    }
  });
</script>

{#if isManager}
  <AdminDashboard />
{:else if userRole === 'intern'}
  {#if loading}
    <div class="p-8 text-center flex justify-center items-center h-[50vh]">
      <div class="animate-pulse text-slate-400">Memuat profil...</div>
    </div>
  {:else if !internProfile || !internProfile.id}
    <!-- <InternDashboard intern={internProfile} /> -->
    <InternDashboard />
  {:else if internProfile?.status === 'pending'}
    <WaitingApproval />
  {:else}
    <InternDashboard />
  {/if}
{:else}
  <div class="p-8 text-center flex justify-center items-center h-[50vh]">
    <div class="animate-pulse text-slate-400">Loading dashboard...</div>
  </div>
{/if}
