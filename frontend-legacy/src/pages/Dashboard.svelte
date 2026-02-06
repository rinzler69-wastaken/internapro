<script>
  import { onMount } from 'svelte';
  import { auth, todayAttendance } from '../lib/stores';
  import { api } from '../lib/api';
  import LoadingSpinner from '../components/LoadingSpinner.svelte';

  let loading = true;
  let stats = {
    totalTasks: 0,
    completedTasks: 0,
    pendingTasks: 0,
    attendanceRate: 0,
  };

  onMount(async () => {
    try {
      // 1. Load today's attendance
      const attendance = await api.getTodayAttendance();
      todayAttendance.set(attendance.data);
      
      // 2. Load Attendance Rate (Analytics)
      // Note: Using ID 1 temporarily to match Analytics.svelte. 
      // Later this should be dynamic: $auth.user.id
      const trends = await api.getWeeklyTrends(1); 
      if (trends.data && trends.data.summary) {
        stats.attendanceRate = trends.data.summary.attendance_rate;
      }

      // TODO: Load task stats from API
      
      loading = false;
    } catch (error) {
      console.error('Failed to load dashboard:', error);
      loading = false;
    }
  });

  // Reactive State Logic
  $: checkedIn = $todayAttendance?.checked_in || false;
  $: attendance = $todayAttendance?.attendance;
  $: hasCheckedOut = attendance?.check_out_time != null;

  // Time Formatters
  $: checkInTime = attendance?.check_in_time 
    ? new Date(attendance.check_in_time).toLocaleTimeString('en-US', { 
        hour: '2-digit', minute: '2-digit' 
      })
    : null;

  $: checkOutTime = attendance?.check_out_time 
    ? new Date(attendance.check_out_time).toLocaleTimeString('en-US', { 
        hour: '2-digit', minute: '2-digit' 
      })
    : null;
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-3xl font-geist font-bold text-black">
      Welcome back{$auth.user ? `, ${$auth.user.email.split('@')[0]}` : ''}
    </h1>
    <p class="mt-1 text-sm font-inter text-vercel-gray-600">
      Here's what's happening with your internship today.
    </p>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <LoadingSpinner size="lg" />
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      
      <div class="card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-inter text-vercel-gray-600 mb-1">Today's Attendance</p>
            
            {#if hasCheckedOut}
              <p class="text-2xl font-geist font-semibold text-blue-600">
                Checked Out
              </p>
              <p class="text-xs font-inter text-vercel-gray-500 mt-1">
                at {checkOutTime}
              </p>
            {:else if checkedIn}
              <p class="text-2xl font-geist font-semibold text-green-600">
                Checked In
              </p>
              <p class="text-xs font-inter text-vercel-gray-500 mt-1">
                at {checkInTime}
              </p>
            {:else}
              <p class="text-2xl font-geist font-semibold text-vercel-gray-900">
                Not Checked In
              </p>
              <p class="text-xs font-inter text-vercel-gray-500 mt-1">
                Ready to start?
              </p>
            {/if}
          </div>

          <div class="flex-shrink-0">
            {#if hasCheckedOut}
              <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                <svg class="w-6 h-6 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
                </svg>
              </div>
            {:else if checkedIn}
              <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
                <svg class="w-6 h-6 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                </svg>
              </div>
            {:else}
              <div class="w-12 h-12 bg-vercel-gray-100 rounded-full flex items-center justify-center">
                <svg class="w-6 h-6 text-vercel-gray-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
                </svg>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-inter text-vercel-gray-600 mb-1">Pending Tasks</p>
            <p class="text-2xl font-geist font-semibold text-black">
              {stats.pendingTasks}
            </p>
          </div>
          <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
            <svg class="w-6 h-6 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
              <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
              <path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
            </svg>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-inter text-vercel-gray-600 mb-1">Attendance Rate</p>
            <p class="text-2xl font-geist font-semibold text-black">
              {stats.attendanceRate ? stats.attendanceRate.toFixed(0) : 0}%
            </p>
          </div>
          <div class="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
            <svg class="w-6 h-6 text-purple-600" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
            </svg>
          </div>
        </div>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-xl font-geist font-semibold text-black mb-4">Quick Actions</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <a href="/attendance" class="btn-secondary text-center">
          Check In/Out
        </a>
        <a href="/tasks" class="btn-secondary text-center">
          View Tasks
        </a>
        <a href="/analytics" class="btn-secondary text-center">
          Performance Analytics
        </a>
        <a href="/profile" class="btn-secondary text-center">
          My Profile
        </a>
      </div>
    </div>

    <div class="card p-6">
      <h2 class="text-xl font-geist font-semibold text-black mb-4">Recent Activity</h2>
      <div class="text-center py-8 text-vercel-gray-500 font-inter text-sm">
        <p>No recent activity</p>
        <p class="text-xs mt-1">Your recent tasks and attendance will appear here</p>
      </div>
    </div>
  {/if}
</div>