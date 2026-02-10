<script>
  import { onMount } from 'svelte';
  import Button from '../components/ui/Button.svelte';
  
  // State Runes
  let activeTab = $state('presence'); 
  let gps = $state({ lat: null, lng: null, error: null });
  
  // Office Config (Eventually fetch from API)
  const OFFICE = { lat: -7.0355, lng: 110.4746 }; 

  function getDistance(lat1, lon1, lat2, lon2) {
     const R = 6371e3; // metres
     const φ1 = lat1 * Math.PI/180;
     const φ2 = lat2 * Math.PI/180;
     const Δφ = (lat2-lat1) * Math.PI/180;
     const Δλ = (lon2-lon1) * Math.PI/180;
     const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
               Math.cos(φ1) * Math.cos(φ2) *
               Math.sin(Δλ/2) * Math.sin(Δλ/2);
     const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
     return R * c;
  }

  // Derived State
  let distance = $derived(
    gps.lat ? Math.round(getDistance(gps.lat, gps.lng, OFFICE.lat, OFFICE.lng)) : null
  );
  
  let canCheckIn = $derived(distance !== null && distance <= 1000);

  onMount(() => {
    navigator.geolocation.watchPosition(
      (pos) => { gps.lat = pos.coords.latitude; gps.lng = pos.coords.longitude; },
      (err) => { gps.error = err.message; },
      { enableHighAccuracy: true }
    );
  });
</script>

<div class="main-content" style="margin-left:0;">
  <div class="card" style="margin-bottom:16px;">
    <h3 class="font-geist">Dashboard</h3>
    <p class="text-muted">Welcome back, Intern.</p>
  </div>

  <div class="tab-bar">
    <button
      class={`tab-btn ${activeTab === 'presence' ? 'active' : ''}`}
      onclick={() => (activeTab = 'presence')}
    >
      Presensi
    </button>
    <button
      class={`tab-btn ${activeTab === 'tasks' ? 'active' : ''}`}
      onclick={() => (activeTab = 'tasks')}
    >
      Tugas Saya
    </button>
  </div>

  {#if activeTab === 'presence'}
    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap:16px;">
      <div class="card">
        <h4>Location Status</h4>
        {#if gps.error}
          <div class="status status-absent">{gps.error}</div>
        {:else if distance !== null}
           <div class="font-geist" style="font-size:28px; font-weight:600; margin-bottom:6px;">{distance}m</div>
           <p class="text-muted">from Office Center</p>
           
           <div class="mt-6">
             {#if canCheckIn}
               <Button class="w-full">CHECK IN NOW</Button>
             {:else}
               <Button variant="outline" class="w-full opacity-50 cursor-not-allowed" disabled>
                 TOO FAR TO CHECK IN
               </Button>
             {/if}
           </div>
        {:else}
           <p class="text-muted">Locating...</p>
        {/if}
      </div>
      
      <div style="display:grid; gap:12px;">
        <div class="card">
          <span class="text-muted" style="font-size:11px; text-transform:uppercase; letter-spacing:0.12em;">Attendance Rate</span>
          <div class="font-geist" style="font-size:22px; font-weight:600; margin-top:8px;">0%</div>
        </div>
        <div class="card">
          <span class="text-muted" style="font-size:11px; text-transform:uppercase; letter-spacing:0.12em;">Tasks Completed</span>
          <div class="font-geist" style="font-size:22px; font-weight:600; margin-top:8px;">0</div>
        </div>
      </div>
    </div>
  {:else}
    <div class="empty-state">
      <p>Task Board Component Coming Soon...</p>
    </div>
  {/if}
</div>
