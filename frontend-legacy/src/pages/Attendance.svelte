<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api';
  import { location, getCurrentLocation, toast, todayAttendance } from '../lib/stores';
  import LoadingSpinner from '../components/LoadingSpinner.svelte';

  let loading = false;
  let checkingIn = false;
  let checkingOut = false;
  let lateReason = '';
  let showLateReasonInput = false;

  // Clock State
  let now = new Date();
  let timeInterval;

  // Reactive Clock Formats
  $: timeStr = now.toLocaleTimeString('en-US', { 
    hour12: false, 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit' 
  });
  
  $: dateStr = now.toLocaleDateString('en-US', { 
    weekday: 'long', 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  });

  onMount(async () => {
    // Start the clock
    timeInterval = setInterval(() => {
      now = new Date();
    }, 1000);

    // Load Data
    loading = true;
    try {
      const attendance = await api.getTodayAttendance();
      todayAttendance.set(attendance.data);
    } catch (error) {
      console.error('Failed to load attendance:', error);
    } finally {
      loading = false;
    }
  });

  onDestroy(() => {
    if (timeInterval) clearInterval(timeInterval);
  });

  const handleCheckIn = async () => {
    if (!$location.latitude || !$location.longitude) {
      toast.add('Getting your location...', 'info');
      getCurrentLocation();
      return;
    }

    checkingIn = true;
    try {
      const response = await api.checkIn(
        $location.latitude, 
        $location.longitude,
        lateReason || null
      );
      
      toast.add('Checked in successfully!', 'success');
      
      // Refresh attendance data
      const attendance = await api.getTodayAttendance();
      todayAttendance.set(attendance.data);
      
      lateReason = '';
      showLateReasonInput = false;
    } catch (error) {
      const message = error.message || 'Check-in failed';
      toast.add(message, 'error');
      
      // If late, show reason input
      if (message.includes('late') || message.includes('after')) {
        showLateReasonInput = true;
      }
    } finally {
      checkingIn = false;
    }
  };

  const handleCheckOut = async () => {
    if (!$location.latitude || !$location.longitude) {
      toast.add('Getting your location...', 'info');
      getCurrentLocation();
      return;
    }

    checkingOut = true;
    try {
      await api.checkOut($location.latitude, $location.longitude);
      toast.add('Checked out successfully!', 'success');
      
      // Refresh attendance data
      const attendance = await api.getTodayAttendance();
      todayAttendance.set(attendance.data);
    } catch (error) {
      toast.add(error.message || 'Check-out failed', 'error');
    } finally {
      checkingOut = false;
    }
  };

  const refreshLocation = () => {
    getCurrentLocation();
  };

  $: checkedIn = $todayAttendance?.checked_in || false;
  $: attendanceData = $todayAttendance?.attendance;
  $: hasCheckedOut = attendanceData?.check_out_time != null;
</script>

<div class="max-w-2xl mx-auto space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
    <div>
      <h1 class="text-3xl font-geist font-bold text-black">Attendance</h1>
      <p class="mt-1 text-sm font-inter text-vercel-gray-600">
        Check in and out for today's work
      </p>
    </div>
    
    <div class="sm:text-right">
      <div class="text-3xl font-geist font-bold text-black font-mono tracking-tight">
        {timeStr}
      </div>
      <p class="mt-1 text-sm font-inter text-vercel-gray-600">
        {dateStr}
      </p>
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>
  {:else}
    <div class="card p-6">
      <h2 class="text-xl font-geist font-semibold text-black mb-4">Today's Status</h2>
      
      <div class="space-y-4">
        {#if checkedIn}
          <div class="flex items-center justify-between p-4 bg-green-50 rounded-vercel border border-green-200">
            <div class="flex items-center space-x-3">
              <div class="flex-shrink-0">
                <svg class="w-8 h-8 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <p class="text-sm font-inter font-medium text-green-900">Checked In</p>
                {#if attendanceData?.check_in_time}
                  <p class="text-xs font-inter text-green-700">
                    {new Date(attendanceData.check_in_time).toLocaleTimeString('en-US', { 
                      hour: '2-digit', 
                      minute: '2-digit' 
                    })}
                  </p>
                {/if}
              </div>
            </div>
            {#if attendanceData?.status === 'late'}
              <span class="badge badge-warning">Late</span>
            {:else}
              <span class="badge badge-success">On Time</span>
            {/if}
          </div>

          {#if hasCheckedOut}
            <div class="flex items-center justify-between p-4 bg-blue-50 rounded-vercel border border-blue-200">
              <div class="flex items-center space-x-3">
                <div class="flex-shrink-0">
                  <svg class="w-8 h-8 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div>
                  <p class="text-sm font-inter font-medium text-blue-900">Checked Out</p>
                  {#if attendanceData?.check_out_time}
                    <p class="text-xs font-inter text-blue-700">
                      {new Date(attendanceData.check_out_time).toLocaleTimeString('en-US', { 
                        hour: '2-digit', 
                        minute: '2-digit' 
                      })}
                    </p>
                  {/if}
                </div>
              </div>
            </div>
          {/if}
        {:else}
          <div class="text-center py-8">
            <svg class="w-16 h-16 mx-auto text-vercel-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p class="text-sm font-inter text-vercel-gray-600">You haven't checked in yet today</p>
          </div>
        {/if}
      </div>
    </div>

    <div class="card p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-geist font-semibold text-black">Your Location</h2>
        <button 
          on:click={refreshLocation}
          class="btn-secondary text-sm"
          disabled={$location.loading}
        >
          {#if $location.loading}
            <LoadingSpinner size="sm" />
          {:else}
            Refresh
          {/if}
        </button>
      </div>

      {#if $location.error}
        <div class="bg-red-50 border border-red-200 rounded-vercel p-4">
          <p class="text-sm font-inter text-red-700">{$location.error}</p>
          <button 
            on:click={refreshLocation}
            class="mt-2 text-sm font-inter font-medium text-red-800 hover:underline"
          >
            Try again
          </button>
        </div>
      {:else if $location.latitude && $location.longitude}
        <div class="space-y-2">
          <div class="flex justify-between items-center p-3 bg-vercel-gray-50 rounded-vercel">
            <span class="text-sm font-inter text-vercel-gray-600">Latitude:</span>
            <span class="text-sm font-inter font-medium text-black">{$location.latitude.toFixed(6)}</span>
          </div>
          <div class="flex justify-between items-center p-3 bg-vercel-gray-50 rounded-vercel">
            <span class="text-sm font-inter text-vercel-gray-600">Longitude:</span>
            <span class="text-sm font-inter font-medium text-black">{$location.longitude.toFixed(6)}</span>
          </div>
          <p class="text-xs font-inter text-vercel-gray-500 mt-2">
            📍 Location acquired successfully
          </p>
        </div>
      {:else}
        <button 
          on:click={refreshLocation}
          class="btn-primary w-full"
          disabled={$location.loading}
        >
          {#if $location.loading}
            <LoadingSpinner size="sm" color="white" />
            <span class="ml-2">Getting location...</span>
          {:else}
            Get My Location
          {/if}
        </button>
      {/if}
    </div>

    {#if showLateReasonInput && !checkedIn}
      <div class="card p-6">
        <label for="lateReason" class="label">Reason for being late</label>
        <textarea
          id="lateReason"
          bind:value={lateReason}
          class="input"
          rows="3"
          placeholder="Please provide a reason for arriving late..."
          required
        />
        <p class="text-xs font-inter text-vercel-gray-500 mt-1">
          You're checking in after the scheduled time. Please provide a reason.
        </p>
      </div>
    {/if}

    <div class="card p-6">
      <h2 class="text-xl font-geist font-semibold text-black mb-4">Actions</h2>
      
      <div class="space-y-3">
        {#if !checkedIn}
          <button
            on:click={handleCheckIn}
            class="btn-primary w-full"
            disabled={checkingIn || !$location.latitude}
          >
            {#if checkingIn}
              <LoadingSpinner size="sm" color="white" />
              <span class="ml-2">Checking in...</span>
            {:else}
              Check In
            {/if}
          </button>
        {:else if !hasCheckedOut}
          <button
            on:click={handleCheckOut}
            class="btn-primary w-full"
            disabled={checkingOut || !$location.latitude}
          >
            {#if checkingOut}
              <LoadingSpinner size="sm" color="white" />
              <span class="ml-2">Checking out...</span>
            {:else}
              Check Out
            {/if}
          </button>
        {:else}
          <div class="text-center py-4">
            <p class="text-sm font-inter text-vercel-gray-600">
              You've completed today's attendance
            </p>
          </div>
        {/if}

        {#if !$location.latitude}
          <p class="text-xs font-inter text-vercel-gray-500 text-center">
            Please allow location access to check in/out
          </p>
        {/if}
      </div>
    </div>
  {/if}
</div>