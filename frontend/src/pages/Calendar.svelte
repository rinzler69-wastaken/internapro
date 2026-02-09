<script>
  import { onMount } from 'svelte';
  import { goto } from '@mateothegreat/svelte5-router';
  import { api } from '../lib/api.js';
  import { auth } from '../lib/auth.svelte.js';

  // --- Configuration ---
  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
  ];

  // --- State ---
  let currentMonth = $state(new Date());
  let loading = $state(false);
  let attendance = $state([]);
  let tasks = $state([]);
  let selectedDate = $state(new Date());
  let viewMode = $state('attendance'); // 'attendance' | 'tasks'
  let showModal = $state(false);
  let showTaskModal = $state(false);
  let modalData = $state(null);
  let holidays = $state([]);
  
  // Search states
  let attendanceSearchQuery = $state('');
  let taskSearchQuery = $state('');

  // --- User Role ---
  const canManage = $derived(auth.user?.role === 'admin' || auth.user?.role === 'supervisor');

  // --- Helpers ---
  const monthKey = $derived(
    `${currentMonth.getFullYear()}-${String(currentMonth.getMonth() + 1).padStart(2, '0')}`
  );

  function startOfMonth(date) {
    return new Date(date.getFullYear(), date.getMonth(), 1);
  }

  function endOfMonth(date) {
    return new Date(date.getFullYear(), date.getMonth() + 1, 0);
  }

  function toDateKey(date) {
    const yyyy = date.getFullYear();
    const mm = String(date.getMonth() + 1).padStart(2, '0');
    const dd = String(date.getDate()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd}`;
  }

  function parseDate(value) {
    if (!value) return null;
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return null;
    return d;
  }

  // FIXED: Normalize date string to YYYY-MM-DD format for consistent comparison
  function normalizeDateString(dateStr) {
    if (!dateStr) return null;
    // If it's already in YYYY-MM-DD format, return as is
    if (/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) {
      return dateStr;
    }
    // Try parsing and reformatting
    const parsed = parseDate(dateStr);
    return parsed ? toDateKey(parsed) : null;
  }

  // Generate Calendar Grid
  function buildCalendarDays(date) {
    const start = startOfMonth(date);
    const end = endOfMonth(date);
    const startDay = start.getDay(); // 0 = Sunday
    const daysInMonth = end.getDate();

    const days = [];
    const prevMonth = new Date(date.getFullYear(), date.getMonth(), 0);
    const prevDays = prevMonth.getDate();

    // Previous month padding
    for (let i = startDay - 1; i >= 0; i--) {
      days.push({ date: new Date(date.getFullYear(), date.getMonth() - 1, prevDays - i), outside: true });
    }

    // Current month days
    for (let d = 1; d <= daysInMonth; d++) {
      days.push({ date: new Date(date.getFullYear(), date.getMonth(), d), outside: false });
    }

    // Next month padding (fill up to 42 cells for 6 rows)
    while (days.length < 42) {
      const last = days[days.length - 1].date;
      days.push({ date: new Date(last.getFullYear(), last.getMonth(), last.getDate() + 1), outside: true });
    }

    return days;
  }

  const calendarDays = $derived(buildCalendarDays(currentMonth));

  // --- Data Mapping ---
  function buildAttendanceMap(records) {
    const map = {};
    for (const record of records) {
      if (record.date) {
        const normalizedDate = normalizeDateString(record.date);
        if (normalizedDate) {
          map[normalizedDate] = record;
        }
      }
    }
    return map;
  }

  function buildTasksMap(list) {
    const map = {};
    for (const task of list) {
      const target = task.deadline || task.start_date || task.target_date;
      const parsed = parseDate(target);
      if (!parsed) continue;
      const key = toDateKey(parsed);
      if (!map[key]) map[key] = [];
      map[key].push(task);
    }
    return map;
  }

  const attendanceByDate = $derived(buildAttendanceMap(attendance));
  const tasksByDate = $derived(buildTasksMap(tasks));
  
  // --- Holidays Map ---
  function buildHolidaysMap(list) {
    const map = {};
    for (const holiday of list) {
      const normalizedDate = normalizeDateString(holiday.date);
      if (normalizedDate) {
        map[normalizedDate] = holiday;
      }
    }
    return map;
  }
  
  const holidaysByDate = $derived(buildHolidaysMap(holidays));
  
  const monthHolidays = $derived(
    holidays.filter(h => {
      const normalizedDate = normalizeDateString(h.date);
      if (!normalizedDate) return false;
      const hDate = new Date(normalizedDate + 'T00:00:00');
      return hDate.getMonth() === currentMonth.getMonth() && 
             hDate.getFullYear() === currentMonth.getFullYear();
    })
  );

  // --- Filtered Lists ---
  const filteredAttendances = $derived(
    modalData?.attendances?.filter(att => 
      att.intern_name?.toLowerCase().includes(attendanceSearchQuery.toLowerCase())
    ) || []
  );

  const filteredTasks = $derived.by(() => {
    const dateKey = toDateKey(selectedDate);
    const dailyTasks = tasksByDate[dateKey] || [];
    if (!taskSearchQuery.trim()) return dailyTasks;
    
    return dailyTasks.filter(task => 
      task.title?.toLowerCase().includes(taskSearchQuery.toLowerCase()) ||
      task.description?.toLowerCase().includes(taskSearchQuery.toLowerCase())
    );
  });

  // --- Styling Helpers (Parity with Laravel) ---
  function getPriorityColor(priority) {
    // Parity with calendar.blade.php colors
    switch (priority) {
      case 'high': return 'bg-rose-100 text-rose-800 border-rose-200';
      case 'medium': return 'bg-amber-100 text-amber-800 border-amber-200';
      case 'low': 
      default: return 'bg-emerald-100 text-emerald-800 border-emerald-200';
    }
  }

  function getAttendanceColor(status) {
    // Parity with calendar.blade.php colors
    switch (status) {
      case 'present': return { bg: 'bg-emerald-100', text: 'text-emerald-700 border border-emerald-200' };
      case 'late': return { bg: 'bg-amber-100', text: 'text-amber-700 border border-amber-200' };
      case 'absent': return { bg: 'bg-rose-100', text: 'text-rose-700 border border-rose-200' };
      case 'permission': return { bg: 'bg-sky-100', text: 'text-sky-700 border border-sky-200' };
      case 'sick': return { bg: 'bg-purple-100', text: 'text-purple-700 border border-purple-200' };
      default: return { bg: 'bg-slate-100', text: 'text-slate-700 border border-slate-200' };
    }
  }

  function getAttendanceDotColor(status) {
    switch (status) {
      case 'present': return 'bg-emerald-500';
      case 'late': return 'bg-amber-500';
      case 'absent': return 'bg-rose-500';
      case 'permission': return 'bg-sky-500';
      case 'sick': return 'bg-purple-500';
      default: return 'bg-slate-400';
    }
  }

  function getAttendanceLabel(status) {
    switch (status) {
      case 'present': return 'Hadir';
      case 'late': return 'Terlambat';
      case 'absent': return 'Tidak Hadir';
      case 'permission': return 'Izin';
      case 'sick': return 'Sakit';
      default: return '-';
    }
  }

  function getPriorityDotColor(priority) {
    switch (priority) {
      case 'high': return 'bg-rose-500';
      case 'medium': return 'bg-amber-500';
      case 'low':
      default: return 'bg-emerald-500';
    }
  }

  function getHighestPriority(tasks) {
    if (!tasks || tasks.length === 0) {
      return 'low'; // default
    }
    const priorityOrder = { 'high': 3, 'medium': 2, 'low': 1 };
    let highestPriority = 'low';
    let maxPrio = 1;
    for (const task of tasks) {
      const prioValue = priorityOrder[task.priority] || 1;
      if (prioValue > maxPrio) {
        maxPrio = prioValue;
        highestPriority = task.priority;
      }
    }
    return highestPriority;
  }

  // --- API ---
  async function fetchCalendarData() {
    loading = true;
    try {
      const year = currentMonth.getFullYear();
      const [attRes, tasksRes, holidaysRes] = await Promise.all([
        api.getAttendance({ month: monthKey, limit: 999 }),
        api.getTasks({ month: monthKey, limit: 999 }),
        api.getHolidays({ year: year })
      ]);
      
      attendance = attRes.data || [];
      tasks = tasksRes.data || [];
      holidays = holidaysRes.data || [];

      // Debug: Log attendance data format
      console.log('Sample attendance record:', attendance[0]);
      console.log('Total attendance records:', attendance.length);

    } catch (err) {
      console.error(err);
      // Optionally, show a toast notification
    } finally {
      loading = false;
    }
  }

  // --- Interaction ---
  function isToday(date) {
    return date.toDateString() === new Date().toDateString();
  }

  function isSelected(date) {
    return date.toDateString() === selectedDate.toDateString();
  }
  
  function isHoliday(date) {
    return date.getDay() === 0; // Simple Sunday check for red styling
  }
  
  function getHoliday(date) {
    const key = toDateKey(date);
    return holidaysByDate[key] || null;
  }

  function goPrevMonth() {
    currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1);
  }

  function goNextMonth() {
    currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1);
  }
  
  function selectDate(date) {
    selectedDate = date;
  }
  
  async function openAttendanceModal(date) {
    const dateKey = toDateKey(date);
    
    // Fetch all students' attendance for this date
    try {
      const res = await api.getAttendance({ date: dateKey, limit: 999 });
      const dailyAttendances = res.data || [];
      
      const stats = {
        present: dailyAttendances.filter(a => a.status === 'present').length,
        late: dailyAttendances.filter(a => a.status === 'late').length,
        absent: dailyAttendances.filter(a => a.status === 'absent').length,
        permission: dailyAttendances.filter(a => a.status === 'permission').length,
        sick: dailyAttendances.filter(a => a.status === 'sick').length,
        total: dailyAttendances.length,
      };
      
      modalData = {
        date: date.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }),
        ...stats,
        attendances: dailyAttendances,
      };
      
      attendanceSearchQuery = ''; // Reset search
      showModal = true;
    } catch (error) {
      console.error("Failed to fetch attendance details:", error);
      // Optionally show an error toast
    }
  }
  
  function closeModal() {
    showModal = false;
    modalData = null;
    attendanceSearchQuery = '';
  }
  
  function closeTaskModal() {
    showTaskModal = false;
    taskSearchQuery = '';
  }
  
  // FIXED: Now uses normalized date matching and returns events array like Laravel
  function calculateDayStats(dateKey) {
    // Filter records for this specific date with normalized comparison
    const dailyRecords = attendance.filter(a => {
      const recordDate = normalizeDateString(a.date);
      return recordDate === dateKey;
    });
    
    if (dailyRecords.length === 0) {
      return null; // No data for this day
    }

    // Build event objects similar to Laravel structure
    const events = [];
    const statuses = ['present', 'permission', 'late', 'sick'];
    
    for (const status of statuses) {
      const count = dailyRecords.filter(a => a.status === status).length;
      if (count > 0) {
        events.push({ status, count });
      }
    }
    
    return events.length > 0 ? events : null;
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      if (showTaskModal) closeTaskModal();
      if (showModal) closeModal();
    }
  }

  onMount(fetchCalendarData);
  $effect(() => {
    monthKey;
    fetchCalendarData();
  });
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="w-full max-w-7xl mx-auto space-y-6">
  <div class="card p-0 overflow-hidden bg-white border-slate-200 shadow-sm rounded-xl">
    
    <div class="p-4 sm:p-6 border-b border-slate-100 flex flex-col md:flex-row justify-between items-center gap-4 bg-slate-50/50">
      <div class="flex items-center justify-between w-full md:w-auto gap-4">
        <button onclick={goPrevMonth} class="w-8 h-8 flex items-center justify-center rounded-full bg-white border border-slate-200 text-slate-500 hover:bg-slate-100 hover:text-slate-700 transition-colors shadow-sm" aria-label="Bulan sebelumnya">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
        </button>
        <h3 class="text-lg sm:text-xl font-bold text-slate-800 text-center min-w-[150px]">
          {monthNames[currentMonth.getMonth()]} {currentMonth.getFullYear()}
        </h3>
        <button onclick={goNextMonth} class="w-8 h-8 flex items-center justify-center rounded-full bg-white border border-slate-200 text-slate-500 hover:bg-slate-100 hover:text-slate-700 transition-colors shadow-sm" aria-label="Bulan berikutnya">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
        </button>
      </div>

      <div class="flex flex-wrap md:flex-nowrap w-full md:w-auto gap-2">
        <button 
          onclick={() => viewMode = 'attendance'} 
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode === 'attendance' ? 'bg-slate-900 text-white border border-transparent' : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><path d="M9 16l2 2 4-4"/></svg>
          <span class="hidden sm:inline">Kehadiran</span>
        </button>
        <button 
          onclick={() => viewMode = 'tasks'} 
          class="flex-1 md:flex-none px-5 py-2 rounded-full text-sm font-semibold transition-all flex items-center justify-center gap-2 {viewMode === 'tasks' ? 'bg-slate-900 text-white border border-transparent' : 'bg-white text-slate-900 border border-slate-200 hover:border-slate-300'}"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
          <span class="hidden sm:inline">Tugas</span>
        </button>
              <button 
                onclick={() => viewMode === 'attendance' ? openAttendanceModal(selectedDate) : (showTaskModal = true)}
                class="w-full md:w-auto justify-center cursor-pointer px-5 py-2 rounded-full text-sm font-semibold bg-slate-900 text-white hover:bg-slate-800 transition-all shadow-sm flex items-center gap-2"
              >
                {#if viewMode === 'attendance'}
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><polyline points="17 11 19 13 23 9"/></svg>
                  Tinjau Kehadiran
                {:else}
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v2a2 2 0 0 0 2 2z"/></svg>
                  Tinjau Agenda
                {/if}
              </button>
      </div>
    </div>

    {#if loading}
      <div class="p-8 text-center text-muted" style="min-height: 750px;">Memuat data kalender...</div>
    {:else}
      <div class="p-2 sm:p-6">
        <div class="grid grid-cols-7 gap-1 sm:gap-2 mb-2">
          {#each ['MIN', 'SEN', 'SEL', 'RAB', 'KAM', 'JUM', 'SAB'] as dayName}
            <div class="text-center py-2 text-xs font-bold text-slate-400 uppercase tracking-wider">
              {dayName}
            </div>
          {/each}
        </div>

        <!-- Fixed height grid to prevent jitter - 6 rows × 112px per row = 672px -->
        <div class="grid grid-cols-7 gap-1 sm:gap-2 auto-rows-fr" style="height: 672px;">
          {#each calendarDays as day}
            {@const dateKey = toDateKey(day.date)}
            {@const att = attendanceByDate[dateKey]}
            {@const dayTasks = tasksByDate[dateKey] || []}
            {@const holidayInfo = getHoliday(day.date)}
            {@const isRedDate = isHoliday(day.date) || holidayInfo}
            
            <button
              class="relative p-2 border rounded-xl transition-all flex flex-col gap-1 text-left
                     {isToday(day.date) ? 'bg-indigo-50/50 border-indigo-200 ring-1 ring-indigo-200' : (isRedDate ? 'bg-rose-50/50 border-rose-100' : 'bg-white border-slate-100')}
                     {isSelected(day.date) ? '!ring-2 !ring-indigo-600 !border-transparent z-10' : ''}
                     {viewMode === 'attendance' && canManage ? 'cursor-pointer hover:border-indigo-300 hover:shadow-md hover:-translate-y-0.5' : 'hover:border-indigo-300 hover:shadow-sm'}"
              style={day.outside ? 'opacity: 0.5;' : ''}
              onclick={() => selectDate(day.date)}
              title={holidayInfo ? holidayInfo.name : ''}
            >
              <div class="flex justify-between items-start w-full">
                <span class="text-sm font-semibold {isToday(day.date) ? 'text-indigo-600 bg-indigo-100 w-6 h-6 rounded-lg flex items-center justify-center' : (isRedDate ? 'text-rose-500' : 'text-slate-700')}">
                  {day.date.getDate()}
                </span>
                
                {#if holidayInfo}
                  <span class="hidden sm:inline-block text-[9px] bg-rose-100 text-rose-600 px-1.5 py-0.5 rounded font-medium leading-tight text-right max-w-[85px] break-words">
                    {holidayInfo.name}
                  </span>
                {/if}
              </div>

              <div class="flex-1 w-full flex flex-col gap-1 overflow-hidden mt-1">
                
                {#if viewMode === 'attendance'}
                  {#if canManage}
                    <!-- FIXED: Admin Summary View - Laravel parity with proper styling -->
                    {@const dayEvents = calculateDayStats(dateKey)}
                    {#if dayEvents}
                      <!-- Desktop & Mobile view with proper container -->
                      <div class="flex flex-wrap gap-1 content-start mt-1">
                        {#each dayEvents as event}
                          {@const statusClasses = {
                            'present': 'bg-emerald-500',
                            'late': 'bg-amber-500',
                            'absent': 'bg-rose-500',
                            'permission': 'bg-sky-500',
                            'sick': 'bg-purple-500',
                          }}
                          {@const bgClass = statusClasses[event.status] || 'bg-slate-500'}
                          
                          <div 
                            class="w-5 h-5 sm:w-6 sm:h-6 rounded-full {bgClass} flex items-center justify-center text-[9px] sm:text-[10px] font-bold text-white ring-1 ring-white shadow-sm" 
                            title="{getAttendanceLabel(event.status)}: {event.count}"
                          >
                            {event.count}
                          </div>
                        {/each}
                      </div>
                    {/if}
                  {:else}
                    <!-- Personal Attendance View -->
                    {#if att}
                      {@const attColor = getAttendanceColor(att.status)}
                      <div class="hidden sm:flex items-center gap-1.5 px-2 py-1 rounded-md text-[10px] sm:text-xs font-semibold {attColor.bg} {attColor.text}">
                        <span class="truncate">
                          {att.check_in_time ? new Date(att.check_in_time).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit'}) : getAttendanceLabel(att.status)}
                        </span>
                      </div>
                      <div class="sm:hidden w-2 h-2 rounded-full mx-auto {getAttendanceDotColor(att.status)}"></div>
                    {/if}
                  {/if}

                {:else}
                  <!-- Tasks View -->
                  {#if dayTasks.length > 0}
                    {#each dayTasks.slice(0, 3) as task}
                      <button onclick={() => goto(`/tasks/${task.id}`)} class="w-full text-left hidden sm:block cursor-pointer">
                        <div class="px-1.5 py-0.5 rounded border text-[10px] font-medium truncate w-full {getPriorityColor(task.priority)} hover:ring-1 hover:ring-offset-1 hover:ring-indigo-400">
                          {task.title}
                        </div>
                      </button>
                    {/each}
                    {@const highestPriority = getHighestPriority(dayTasks)}
                    <div class="sm:hidden w-2 h-2 rounded-full {getPriorityDotColor(highestPriority)} mx-auto"></div>
                    
                    {#if dayTasks.length > 3}
                      <span class="text-[9px] text-slate-400 text-center hidden sm:block">+{dayTasks.length - 3} lainnya</span>
                    {/if}
                  {/if}
                {/if}

              </div>
            </button>
          {/each}
        </div>
        
        <!-- Legend -->
        <div class="flex flex-wrap items-center gap-4 mt-6 pt-4 border-t border-slate-100 text-xs">
          <!-- Holiday Legend (always visible) -->
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded bg-rose-50 border border-rose-200"></span>
            <span class="text-rose-500 font-medium">Libur/Minggu</span>
          </div>
          
          {#if viewMode === 'attendance'}
            <!-- Attendance Legends -->
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-emerald-500"></span>
              <span class="text-slate-600">Hadir</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-amber-500"></span>
              <span class="text-slate-600">Terlambat</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-rose-500"></span>
              <span class="text-slate-600">Tidak Hadir</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-sky-500"></span>
              <span class="text-slate-600">Izin</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-purple-500"></span>
              <span class="text-slate-600">Sakit</span>
            </div>
            {#if canManage}
              <div class="ml-auto text-slate-400 hidden sm:flex items-center gap-1">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 9l-3 3 3 3M9 12h12"/>
                </svg>
                Klik tanggal untuk detail
              </div>
            {/if}
          {:else}
            <!-- Task Priority Legends -->
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-rose-500"></span>
              <span class="text-slate-600">Prioritas Tinggi</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-amber-500"></span>
              <span class="text-slate-600">Prioritas Sedang</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-emerald-500"></span>
              <span class="text-slate-600">Prioritas Rendah</span>
            </div>
            

          {/if}
        </div>
      </div>
    {/if}
  </div>

  <!-- Holiday List Card -->
  <div class="min-h-40">
    {#if monthHolidays.length > 0}
      <div class="card p-4 bg-gradient-to-br from-rose-50 to-rose-100/30 border-rose-200 rounded-xl">
        <h4 class="text-sm font-bold text-rose-700 uppercase tracking-wider mb-3 flex items-center gap-2">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
            <line x1="16" y1="2" x2="16" y2="6"/>
            <line x1="8" y1="2" x2="8" y2="6"/>
            <line x1="3" y1="10" x2="21" y2="10"/>
          </svg>
          Hari Libur Bulan Ini
        </h4>
        <div class="space-y-2">
          {#each monthHolidays as holiday}
            {@const normalizedDate = normalizeDateString(holiday.date)}
            {#if normalizedDate}
              <div class="flex items-center gap-2 p-2 bg-white/60 rounded-lg border border-rose-100">
                <span class="w-8 h-8 bg-rose-100 text-rose-600 rounded-lg flex items-center justify-center font-bold text-sm">
                  {new Date(normalizedDate + 'T00:00:00').getDate()}
                </span>
                <span class="text-sm text-slate-700 font-medium">{holiday.name}</span>
              </div>
            {/if}
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <!-- Admin Attendance Modal -->
  {#if viewMode === 'attendance'}
    {#if showModal}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" onclick={closeModal}>
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>

        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
          <!-- Modal Header -->
          <div class="p-5 border-b border-slate-100 flex justify-between items-start bg-slate-50/50">
            <div>
              <h3 class="text-lg cursor-pointer font-bold text-slate-800 flex items-center gap-2">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                  <line x1="16" y1="2" x2="16" y2="6"/>
                  <line x1="8" y1="2" x2="8" y2="6"/>
                  <line x1="3" y1="10" x2="21" y2="10"/>
                </svg>
                Riwayat Kehadiran
              </h3>
              <p class="text-slate-500 text-sm mt-1">{modalData?.date || ''}</p>
            </div>
            <button onclick={closeModal} class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors" aria-label="Tutup modal">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>

          <!-- Stats Grid -->
          <div class="p-6 overflow-y-auto">
            <div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
              <div class="bg-gradient-to-br from-emerald-50 to-emerald-100/50 p-4 rounded-xl text-center border border-emerald-100">
                <div class="text-3xl font-black text-emerald-600 mb-1">{modalData?.present || 0}</div>
                <div class="text-xs font-bold text-emerald-600/70 uppercase tracking-wider">Hadir</div>
              </div>
              <div class="bg-gradient-to-br from-amber-50 to-amber-100/50 p-4 rounded-xl text-center border border-amber-100">
                <div class="text-3xl font-black text-amber-600 mb-1">{modalData?.late || 0}</div>
                <div class="text-xs font-bold text-amber-600/70 uppercase tracking-wider">Terlambat</div>
              </div>
              <div class="bg-gradient-to-br from-rose-50 to-rose-100/50 p-4 rounded-xl text-center border border-rose-100">
                <div class="text-3xl font-black text-rose-600 mb-1">{modalData?.absent || 0}</div>
                <div class="text-xs font-bold text-rose-600/70 uppercase tracking-wider">Belum Absen</div>
              </div>
              <div class="bg-gradient-to-br from-sky-50 to-sky-100/50 p-4 rounded-xl text-center border border-sky-100">
                <div class="text-3xl font-black text-sky-600 mb-1">{modalData?.permission || 0}</div>
                <div class="text-xs font-bold text-sky-600/70 uppercase tracking-wider">Izin</div>
              </div>
              <div class="bg-gradient-to-br from-purple-50 to-purple-100/50 p-4 rounded-xl text-center border border-purple-100">
                <div class="text-3xl font-black text-purple-600 mb-1">{modalData?.sick || 0}</div>
                <div class="text-xs font-bold text-purple-600/70 uppercase tracking-wider">Sakit</div>
              </div>
              <div class="bg-slate-50 p-4 rounded-xl text-center border border-slate-100">
                <div class="text-3xl font-black text-slate-600 mb-1">{modalData?.total || 0}</div>
                <div class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Siswa</div>
              </div>
            </div>

            <!-- Search Bar (hidden if no data) -->
            {#if modalData?.attendances && modalData.attendances.length > 0}
              {#if modalData.attendances.length > 1}
              <div class="mb-4">
                <div class="relative">
                  <input 
                    type="text" 
                    bind:value={attendanceSearchQuery}
                    placeholder="Cari nama siswa..."
                    class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-slate-200 focus:border-indigo-300 focus:ring-2 focus:ring-indigo-100 transition-all text-sm"
                  />
                  <svg class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="11" cy="11" r="8"/>
                    <path d="m21 21-4.35-4.35"/>
                  </svg>
                </div>
              </div>
              {/if}

              <!-- Attendance List -->
              <div>
                <h4 class="text-sm font-bold text-slate-700 uppercase tracking-wider mb-3 flex items-center gap-2">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                    <circle cx="9" cy="7" r="4"/>
                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                    <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                  </svg>
                  Detail Kehadiran
                  {#if attendanceSearchQuery}
                    <span class="text-xs font-normal text-slate-400">({filteredAttendances.length} hasil)</span>
                  {/if}
                </h4>
                
                {#if filteredAttendances.length > 0}
                  <div class="border border-slate-200 rounded-xl overflow-hidden">
                    <table class="w-full text-left text-sm">
                      <thead class="bg-slate-50 text-slate-500 font-semibold uppercase text-xs">
                        <tr>
                          <th class="px-4 py-3">Nama Siswa</th>
                          <th class="px-4 py-3">Status</th>
                          <th class="px-4 py-3 text-center">Masuk</th>
                          <th class="px-4 py-3 text-center">Keluar</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-slate-100">
                        {#each filteredAttendances as att}
                          {@const attColor = getAttendanceColor(att.status)}
                          <tr class="hover:bg-slate-50/50">
                            <td class="px-4 py-3 font-medium text-slate-700">{att.intern_name}</td>
                            <td class="px-4 py-3">
                              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold {attColor.bg} {attColor.text}">
                                {getAttendanceLabel(att.status)}
                              </span>
                            </td>
                            <td class="px-4 py-3 text-center font-mono text-xs text-slate-500">
                              {att.check_in_time ? new Date(att.check_in_time).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit'}) : '-'}
                            </td>
                            <td class="px-4 py-3 text-center font-mono text-xs text-slate-500">
                              {att.check_out_time ? new Date(att.check_out_time).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit'}) : '-'}
                            </td>
                          </tr>
                        {/each}
                      </tbody>
                    </table>
                  </div>
                {:else}
                  <div class="text-center py-8 text-slate-400 bg-slate-50 rounded-xl">
                    <p>Tidak ditemukan siswa dengan nama "{attendanceSearchQuery}"</p>
                  </div>
                {/if}
              </div>
            {:else}
              <div class="text-center py-8 text-slate-400">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mx-auto mb-2 opacity-50">
                  <path d="M3 3h18v18H3z"/>
                  <path d="M9 9h6v6H9z"/>
                </svg>
                <p>Tidak ada data kehadiran untuk tanggal ini</p>
              </div>
            {/if}
          </div>

          <!-- Modal Footer -->
          <div class="p-4 border-t border-slate-100 bg-slate-50/50 text-right">
            <button onclick={closeModal} class="px-4 py-2 bg-white border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 font-medium text-sm transition-colors shadow-sm">
              Tutup
            </button>
          </div>
        </div>
      </div>
    {/if}
  {/if}

    <!-- Task Modal -->
    {#if showTaskModal}
      {@const dateKey = toDateKey(selectedDate)}
      {@const dailyTasks = tasksByDate[dateKey] || []}
      {@const holiday = getHoliday(selectedDate)}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" onclick={closeTaskModal}>
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"></div>

        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] flex flex-col overflow-hidden" onclick={(e) => e.stopPropagation()}>
          <!-- Header -->
          <div class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <h3 class="text-lg font-bold text-slate-800 flex items-center gap-2">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
              Agenda
            </h3>
            <button onclick={closeTaskModal} class="cursor-pointer w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-rose-50 hover:text-rose-500 transition-colors">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          
          <!-- Body -->
          <div class="p-6 overflow-y-auto">
             <div class="mb-4">
                <h4 class="text-sm font-bold text-slate-500 uppercase tracking-wider">
                  {selectedDate.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
                </h4>
             </div>

             {#if holiday}
               <div class="mb-4 p-3 bg-rose-50 border border-rose-100 rounded-xl flex items-center gap-3">
                 <span class="w-8 h-8 bg-rose-100 text-rose-600 rounded-lg flex items-center justify-center font-bold text-sm">
                   <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
                 </span>
                 <div>
                   <h5 class="font-bold text-rose-700 text-sm">Hari Libur</h5>
                   <p class="text-xs text-rose-600 font-medium">{holiday.name}</p>
                 </div>
               </div>
             {/if}

             <!-- Search Bar (hidden if no tasks) -->
             {#if dailyTasks.length > 0}
               {#if dailyTasks.length > 1}
               <div class="mb-4">
                 <div class="relative">
                   <input 
                     type="text" 
                     bind:value={taskSearchQuery}
                     placeholder="Cari judul atau deskripsi tugas..."
                     class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-slate-200 focus:border-indigo-300 focus:ring-2 focus:ring-indigo-100 transition-all text-sm"
                   />
                   <svg class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                     <circle cx="11" cy="11" r="8"/>
                     <path d="m21 21-4.35-4.35"/>
                   </svg>
                 </div>
               </div>
               {/if}

               {#if filteredTasks.length > 0}
                 <div class="space-y-3">
                   {#each filteredTasks as task}
                     <button 
                       onclick={() => goto(`/tasks/${task.id}`)}
                       class="cursor-pointer w-full text-left p-4 rounded-xl border border-slate-200 hover:border-indigo-300 hover:shadow-md transition-all bg-white group"
                     >
                       <div class="flex justify-between items-start mb-2">
                         <h5 class="font-bold text-slate-800 group-hover:text-indigo-600 transition-colors line-clamp-1">{task.title}</h5>
                         <span class="px-2 py-0.5 rounded text-[10px] font-bold border uppercase {getPriorityColor(task.priority)}">
                           {task.priority}
                         </span>
                       </div>
                       <p class="text-xs text-slate-500 line-clamp-2 mb-3">{task.description || 'Tidak ada deskripsi'}</p>
                       <div class="flex items-center gap-3 text-xs text-slate-400">
                          <span class="flex items-center gap-1">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                            {task.deadline ? new Date(task.deadline).toLocaleTimeString('id-ID', {hour: '2-digit', minute:'2-digit'}) : 'All Day'}
                          </span>
                       </div>
                     </button>
                   {/each}
                 </div>
               {:else}
                 <div class="text-center py-12 text-slate-400 bg-slate-50 rounded-xl border border-dashed border-slate-200">
                   <p>Tidak ditemukan agenda dengan kata kunci "{taskSearchQuery}"</p>
                 </div>
               {/if}
             {:else}
               <div class="text-center py-12 text-slate-400 bg-slate-50 rounded-xl border border-dashed border-slate-200">
                 <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mx-auto mb-2 opacity-50"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                 <p>Tidak ada agenda untuk tanggal ini</p>
               </div>
             {/if}
          </div>
        </div>
      </div>
    {/if}
</div>