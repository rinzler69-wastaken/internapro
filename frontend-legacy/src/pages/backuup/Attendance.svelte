<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api';
  import { location, getCurrentLocation, toast, todayAttendance } from '../lib/stores';
  import LoadingSpinner from '../components/LoadingSpinner.svelte';

  let loading = false;
  let submittng = false;
  let reason = '';
  
  // Live Clock
  let now = new Date();
  let timeInterval;

  // Reactive State
  $: checkedIn = $todayAttendance?.checked_in || false;
  $: attendance = $todayAttendance?.attendance;
  $: hasCheckedOut = attendance?.check_out_time != null;

  onMount(async () => {
    loading = true;
    try {
      const res = await api.getTodayAttendance();
      todayAttendance.set(res.data);
    } catch (error) {
      console.error('Failed to load status:', error);
    } finally {
      loading = false;
    }
    timeInterval = setInterval(() => { now = new Date(); }, 1000);
  });

  onDestroy(() => { clearInterval(timeInterval); });

  const handleCheckIn = async () => {
    if (!$location.latitude) {
      toast.add('📍 Acquiring location... please wait.', 'info');
      getCurrentLocation();
      return;
    }

    submittng = true;
    try {
      await api.checkIn($location.latitude, $location.longitude, reason);
      toast.add('✅ Clock In Successful!', 'success');
      const res = await api.getTodayAttendance();
      todayAttendance.set(res.data);
      reason = ''; 
    } catch (error) {
      toast.add(error.message || 'Check-in failed', 'error');
    } finally {
      submittng = false;
    }
  };

  const handleCheckOut = async () => {
    if (!$location.latitude) {
        toast.add('📍 Acquiring location...', 'info');
        getCurrentLocation();
        return;
    }
    submittng = true;
    try {
      await api.checkOut($location.latitude, $location.longitude);
      toast.add('👋 Checked out. See you tomorrow!', 'success');
      const res = await api.getTodayAttendance();
      todayAttendance.set(res.data);
    } catch (error) {
      toast.add(error.message, 'error');
    } finally {
      submittng = false;
    }
  };
</script>

<div class="max-w-xl mx-auto space-y-6 pt-6">
  
  <div class="text-center space-y-2">
    <h1 class="text-4xl font-geist font-bold text-gray-900">
      {now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' })}
    </h1>
    <p class="text-sm font-inter text-gray-500 uppercase tracking-widest">
      {now.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
    </p>
  </div>

  {#if loading}
    <div class="flex justify-center py-12"><LoadingSpinner size="lg" /></div>
  {:else}

    {#if !checkedIn}
      <div class="card p-8 border-2 border-gray-100 shadow-lg">
        <div class="text-center mb-6">
          <div class="w-16 h-16 bg-blue-50 text-blue-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
          </div>
          <h2 class="text-xl font-bold text-gray-900">Ready to work?</h2>
          <p class="text-gray-500 text-sm">Ensure you are within office range.</p>
        </div>

        <div class="bg-gray-50 rounded-lg p-3 mb-6 flex items-center justify-between text-sm">
            {#if $location.loading}
                <span class="text-gray-500">🛰️ Finding satellites...</span>
            {:else if $location.error}
                <span class="text-red-600">❌ {$location.error}</span>
                <button on:click={getCurrentLocation} class="text-blue-600 underline">Retry</button>
            {:else if $location.latitude}
                <span class="text-green-700 font-mono">📍 {$location.latitude.toFixed(5)}, {$location.longitude.toFixed(5)}</span>
                <button on:click={getCurrentLocation} class="text-gray-400 hover:text-gray-600">Refresh</button>
            {:else}
                <span class="text-gray-500">Location required</span>
                <button on:click={getCurrentLocation} class="text-blue-600 font-medium">Get Location</button>
            {/if}
        </div>

        <div class="mb-6">
            <label class="block text-xs font-bold text-gray-500 uppercase mb-1">
                Note / Late Reason 
                <span class="font-normal text-gray-400 lowercase">(Required if after 08:30)</span>
            </label>
            <textarea bind:value={reason} rows="2" class="w-full border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"></textarea>
        </div>

        <button 
          on:click={handleCheckIn}
          disabled={submittng || !$location.latitude}
          class="btn-primary w-full py-4 text-lg shadow-blue-200 shadow-lg disabled:opacity-50 disabled:shadow-none transition-all"
        >
          {submittng ? 'Checking in...' : 'CLOCK IN'}
        </button>
      </div>

    {:else}
      <div class="card p-8 border-2 border-green-100 shadow-sm">
        <div class="text-center mb-6">
          
          <div class="w-20 h-20 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-4 animate-bounce-short">
            <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>
          </div>
          
          <h2 class="text-2xl font-bold text-green-700 mb-2">Clocked In!</h2>
          
          <p class="text-gray-600">
            You clocked in 
            <span class="font-bold {attendance?.status === 'late' ? 'text-yellow-600' : 'text-green-600'}">
                {attendance?.status === 'late' ? 'Late' : 'On Time'}
            </span>
            at
            <span class="font-mono font-bold text-gray-900">
                {new Date(attendance.check_in_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })}
            </span>
          </p>

          {#if attendance?.status === 'late'}
             <p class="text-xs text-yellow-600 mt-2">Let's strive for better punctuality tomorrow!</p>
          {/if}
        </div>

        <button disabled class="w-full py-3 bg-gray-100 text-gray-400 font-bold rounded-lg cursor-not-allowed mb-4">
            CLOCKED IN
        </button>

        {#if !hasCheckedOut}
             <div class="border-t border-gray-100 pt-4">
                <button 
                    on:click={handleCheckOut}
                    disabled={submittng || !$location.latitude}
                    class="text-red-500 hover:text-red-700 text-sm font-medium hover:underline"
                >
                    {submittng ? 'Processing...' : 'End Day / Clock Out'}
                </button>
             </div>
        {:else}
             <div class="bg-blue-50 rounded-lg p-3 mt-4 text-center">
                <p class="text-blue-900 font-medium">You are clocked out.</p>
             </div>
        {/if}
      </div>
    {/if}

  {/if}
</div>