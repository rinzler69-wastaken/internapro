<script>
  import { onMount } from 'svelte';  import { api } from '../lib/api.js';
 
  let loading = $state(true);
  let savingAttendance = $state(false);

  let attendanceForm = $state({
    attendance_open_time: '07:00',
    check_in_time: '08:00',
    late_tolerance_minutes: 15,
    check_out_time: '17:00',
    office_latitude: '-7.052683',
    office_longitude: '110.469375',
    max_checkin_distance: 100,
    workdays: [1, 2, 3, 4, 5, 6],
    manual_off_date: '',
  });

  const weekdayOptions = [
    { value: 1, label: 'S', name: 'Senin' },
    { value: 2, label: 'S', name: 'Selasa' },
    { value: 3, label: 'R', name: 'Rabu' },
    { value: 4, label: 'K', name: 'Kamis' },
    { value: 5, label: 'J', name: 'Jumat' },
    { value: 6, label: 'S', name: 'Sabtu' },
  ];

  const todayKey = () => {
    // Normalize to WIB (UTC+7) so manual off matches server-side date
    const now = new Date();
    const utcMs = now.getTime() + now.getTimezoneOffset() * 60000;
    const jakarta = new Date(utcMs + 7 * 60 * 60 * 1000);
    const yyyy = jakarta.getFullYear();
    const mm = String(jakarta.getMonth() + 1).padStart(2, '0');
    const dd = String(jakarta.getDate()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd}`;
  };

  onMount(async () => {
   await fetchSettings();
    loading = false;
  });

  const stripSeconds = (val) => {
    if (!val) return '';
    const parts = val.split(':');
    if (parts.length >= 2) return `${parts[0].padStart(2, '0')}:${parts[1].padStart(2, '0')}`;
    return val;
  };

  const addMinutesToTime = (timeStr, minutes) => {
    if (!timeStr) return '';
    const [h, m] = timeStr.split(':').map(Number);
    if (Number.isNaN(h) || Number.isNaN(m)) return timeStr;
    const base = new Date(0, 0, 0, h, m);
    base.setMinutes(base.getMinutes() + Number(minutes || 0));
    const hh = String(base.getHours()).padStart(2, '0');
    const mm = String(base.getMinutes()).padStart(2, '0');
    return `${hh}:${mm}`;
  };

  const parseWorkdays = (val) => {
    if (!val || typeof val !== 'string') return [1, 2, 3, 4, 5, 6];
    const days = val.split(',').map((n) => Number(n)).filter((n) => !Number.isNaN(n) && n >= 0 && n <= 6);
    return days.length ? days : [1, 2, 3, 4, 5, 6];
  };

  function toggleWorkday(day) {
    const current = new Set(attendanceForm.workdays || []);
    if (current.has(day)) current.delete(day); else current.add(day);
    const next = Array.from(current).sort((a, b) => a - b);
    attendanceForm = { ...attendanceForm, workdays: next };
  }

  const isWorkdaySelected = (day) => (attendanceForm.workdays || []).includes(day);

  function toggleTodayOff(flag) {
    attendanceForm = { ...attendanceForm, manual_off_date: flag ? todayKey() : '' };
  }

  /** @param {Event & { currentTarget: HTMLInputElement }} event */
  function handleTodayOffChange(event) {
    const target = event.currentTarget;
    if (target && typeof target.checked === 'boolean') {
      toggleTodayOff(target.checked);
    }
  }

  async function fetchSettings() {
    try {
      const res = await api.getSettings();
      const list = res?.data || [];
      const map = {};
      list.forEach((item) => {
        if (item?.key) map[item.key] = item.value;
      });
 
      attendanceForm = {
        attendance_open_time: stripSeconds(map.attendance_open_time || map.office_start_time) || '07:00',
        check_in_time: stripSeconds(map.check_in_time || map.office_start_time) || '08:00',
        late_tolerance_minutes: Number(map.late_tolerance_minutes ?? 15) || 15,
        check_out_time: stripSeconds(map.check_out_time || map.office_end_time) || '17:00',
        office_latitude: map.office_latitude || '-7.052683',
        office_longitude: map.office_longitude || '110.469375',
        max_checkin_distance: Number(map.max_checkin_distance ?? map.office_radius ?? 100) || 100,
        workdays: parseWorkdays(map.workdays || map.work_days),
        manual_off_date: map.manual_off_date || '',
      };

      // derive tolerance from legacy late_tolerance_time if minutes not present
      if (!map.late_tolerance_minutes && map.late_tolerance_time) {
        const start = stripSeconds(map.check_in_time || map.office_start_time || '08:00');
        const end = stripSeconds(map.late_tolerance_time);
        const diff = (() => {
          const [sh, sm] = (start || '').split(':').map(Number);
          const [eh, em] = (end || '').split(':').map(Number);
          if ([sh, sm, eh, em].some(Number.isNaN)) return null;
          return (eh * 60 + em) - (sh * 60 + sm);
        })();
        if (diff && diff > 0) attendanceForm.late_tolerance_minutes = diff;
      }
    } catch (err) {
      console.warn('Gagal memuat pengaturan, mencoba bootstrap...', err);
      // Try to bootstrap settings table by posting defaults once
      try {
        await api.updateSettings({
          attendance_open_time: '07:00:00',
          check_in_time: '08:00:00',
          check_out_time: '17:00:00',
          late_tolerance_minutes: '15',
          office_latitude: '-7.052683',
          office_longitude: '110.469375',
          max_checkin_distance: '100',
          office_start_time: '08:00:00',
          office_end_time: '17:00:00',
          office_radius: '100',
          late_tolerance_time: '08:15:00',
          workdays: '1,2,3,4,5,6',
          manual_off_date: '',
        });
        // retry fetch once
        const retry = await api.getSettings();
        const list = retry?.data || [];
        const map = {};
        list.forEach((item) => { if (item?.key) map[item.key] = item.value; });
        attendanceForm = {
          attendance_open_time: stripSeconds(map.attendance_open_time || map.office_start_time) || '07:00',
          check_in_time: stripSeconds(map.check_in_time || map.office_start_time) || '08:00',
          late_tolerance_minutes: Number(map.late_tolerance_minutes ?? 15) || 15,
          check_out_time: stripSeconds(map.check_out_time || map.office_end_time) || '17:00',
          office_latitude: map.office_latitude || '-7.052683',
          office_longitude: map.office_longitude || '110.469375',
          max_checkin_distance: Number(map.max_checkin_distance ?? map.office_radius ?? 100) || 100,
          workdays: parseWorkdays(map.workdays || map.work_days),
          manual_off_date: map.manual_off_date || '',
        };
      } catch (err2) {
        console.error('Bootstrap settings failed', err2);
        alert('Gagal memuat pengaturan: ' + (err2.message || 'unknown error'));
      }
    }
  }

  async function saveAttendanceSettings() {
    savingAttendance = true;
    try {
      const t = (val) => (val && val.length === 5 ? `${val}:00` : val || '');
      await api.updateSettings({
        attendance_open_time: t(attendanceForm.attendance_open_time),
        check_in_time: t(attendanceForm.check_in_time),
        check_out_time: t(attendanceForm.check_out_time),
        late_tolerance_minutes: String(attendanceForm.late_tolerance_minutes),
        office_latitude: attendanceForm.office_latitude,
        office_longitude: attendanceForm.office_longitude,
        max_checkin_distance: String(attendanceForm.max_checkin_distance),
        workdays: (attendanceForm.workdays || []).join(','),
        manual_off_date: attendanceForm.manual_off_date || '',
        // legacy/compat keys
        office_start_time: t(attendanceForm.check_in_time),
        office_end_time: t(attendanceForm.check_out_time),
        office_radius: String(attendanceForm.max_checkin_distance),
        late_tolerance_time: t(addMinutesToTime(attendanceForm.check_in_time, attendanceForm.late_tolerance_minutes)),
      });
      alert('Pengaturan presensi tersimpan.');
      await fetchSettings();
    } catch (err) {
      alert('Gagal menyimpan pengaturan: ' + (err.message || 'unknown error'));
   } finally {
      savingAttendance = false;
    }
  }
</script>

<div class="page-bg">
  <div class="container animate-fade-in">
    <div class="header">
      <div>
        <h2>Pengaturan Jam & Lokasi</h2>
      </div>

                        <button class="btn-primary" onclick={saveAttendanceSettings} disabled={savingAttendance}>
        {savingAttendance ? 'Menyimpan...' : 'Simpan Pengaturan'}
      </button>    </div>

     {#if loading}
     <div class="loading">Memuat pengaturan...</div>
    {:else}
      <div class="two-col">
        <div class="content-card animate-slide-up">
          <div class="card-header">
            <h3>Jam Kerja</h3>
            <p>Jam buka presensi, jam masuk, toleransi, dan jam pulang.</p>
          </div>
          <div class="card-body space-y-4">
            <div class="form-grid">
              <div class="form-group">
                <label class="label">Presensi Dibuka</label>
                <input class="input" type="time" bind:value={attendanceForm.attendance_open_time} />
                <p class="help-text">Jam mulai sistem menerima check-in.</p>
              </div>
              <div class="form-group">
                <label class="label">Jam Masuk</label>
                <input class="input" type="time" bind:value={attendanceForm.check_in_time} />
                <p class="help-text">Waktu mulai absen masuk.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-group">
                <label class="label">Toleransi Terlambat (menit)</label>
                <input class="input" type="number" min="0" bind:value={attendanceForm.late_tolerance_minutes} />
                <p class="help-text">Menit tambahan sebelum presensi ditutup.</p>
              </div>
              <div class="form-group">
                <label class="label">Jam Pulang</label>
                <input class="input" type="time" bind:value={attendanceForm.check_out_time} />
                <p class="help-text">Waktu minimal absen pulang.</p>
              </div>
            </div>

            <div class="workdays-grid">
              <div class="form-group">
                <label class="label">Hari Kerja</label>
                <div class="day-pills">
                  {#each weekdayOptions as day}
                    <button
                      type="button"
                      class="day-pill"
                      class:active={isWorkdaySelected(day.value)}
                      title={day.name}
                      aria-label={day.name}
                      onclick={() => toggleWorkday(day.value)}>
                      {day.label}
                    </button>
                  {/each}
                </div>
                <p class="help-text">Pilih hari aktif. Minggu tidak ditampilkan.</p>
              </div>
              <div class="form-group">
                <label class="label">Liburkan Hari Ini</label>
                <div class="today-off-toggle">
                  <label class="toggle-switch">
                    <input
                      type="checkbox"
                      checked={attendanceForm.manual_off_date === todayKey()}
                      onchange={handleTodayOffChange}
                    />
                    <span class="slider"></span>
                  </label>
                  <div>
                    <div class="toggle-title">Tidak ada jadwal kantor</div>
                    <p class="help-text">Intern akan melihat kartu hijau menggantikan tombol absen untuk hari ini.</p>
                  </div>
                </div>
              </div>
            </div>

            <div class="preview-box">
              <div class="text-muted badge-heading">Pagi Hari</div>
              <div class="preview-row">
                <div>
                  <div class="text-success fw-bold">Tepat Waktu</div>
                  <small class="text-muted">Sebelum {attendanceForm.check_in_time || '08:00'}</small>
                </div>
                <div class="divider-vertical"></div>
                <div>
                  <div class="text-warning fw-bold">Terlambat</div>
                  <small class="text-muted">Ditutup pukul {addMinutesToTime(attendanceForm.check_in_time, attendanceForm.late_tolerance_minutes)}</small>
                </div>
              </div>
              <div class="text-muted badge-heading mt-4">Sore Hari</div>
              <div class="preview-row">
                <div>
                  <div class="text-success fw-bold">Boleh Pulang</div>
                  <small class="text-muted">Mulai {attendanceForm.check_out_time || '17:00'}</small>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="content-card animate-slide-up">
          <div class="card-header">
            <h3>Lokasi Absensi</h3>
            <p>Tentukan koordinat kantor dan radius maksimum.</p>
          </div>
          <div class="card-body space-y-4">
            <div class="form-grid">
              <div class="form-group">
                <label class="label">Latitude (Lintang)</label>
                <input class="input" type="text" bind:value={attendanceForm.office_latitude} placeholder="-7.052xxx" />
              </div>
              <div class="form-group">
                <label class="label">Longitude (Bujur)</label>
                <input class="input" type="text" bind:value={attendanceForm.office_longitude} placeholder="110.469xxx" />
              </div>
              <div class="form-group">
                <label class="label">Radius Maksimal (meter)</label>
                <input class="input" type="number" min="10" bind:value={attendanceForm.max_checkin_distance} />
                <p class="help-text">Jarak maksimal dari titik kantor untuk bisa absen.</p>
              </div>
            </div>

            <div class="info-box">
              <div class="icon-circle bg-emerald-light text-emerald">
                <span class="material-symbols-outlined">satellite_alt</span>
              </div>
              <div>
                <h4 class="info-title">Tips</h4>
                <p class="info-desc">Ambil koordinat dari Google Maps (klik kanan titik kantor, salin koordinat).</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(body) { font-family: 'Geist', 'Inter', sans-serif; color: #0f172a; background: #f8fafc; }
  .page-bg { min-height: 100vh; padding: 40x 24px; background: radial-gradient(at 0% 0%, rgba(16,185,129,0.03) 0%, transparent 50%); }
  .container { max-width: 1240px; margin: 0 auto; }
  .header { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 20px; }
  .header h2 { margin: 0; font-size: 20px; font-weight: 760; }
  .header p { margin: 4px 0 0; color: #64748b; }

  .two-col { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 20px; }

  .content-card { background: white; border-radius: 20px; border: 1px solid #e2e8f0; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.02); overflow: hidden; }
  .card-header { padding: 20px 24px; border-bottom: 1px solid #f1f5f9; }
  .card-header h3 { margin: 0; font-size: 18px; font-weight: 700; color: #1e293b; }
  .card-header p { margin: 4px 0 0; font-size: 14px; color: #64748b; }
  .card-body { padding: 20px 24px; }

  .form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
  .form-group { display: flex; flex-direction: column; gap: 6px; }
  .label { font-size: 13px; font-weight: 600; color: #334155; }
  .input {
    width: 100%; padding: 10px 14px; border: 1px solid #cbd5e1; border-radius: 10px;
    font-size: 14px; color: #0f172a; background: #fff; transition: all 0.2s;
  }
  .input:focus { outline: none; border-color: #10b981; box-shadow: 0 0 0 3px rgba(16,185,129,0.1); }
  .help-text { font-size: 12px; color: #94a3b8; margin: 0; }

  .workdays-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; align-items: center; }
  .day-pills { display: flex; gap: 8px; flex-wrap: wrap; }
  .day-pill {
    min-width: 36px; padding: 10px 12px; border-radius: 999px; border: 1px solid #cbd5e1; background: #fff;
    font-weight: 700; color: #0f172a; cursor: pointer; transition: all 0.2s ease; box-shadow: 0 2px 4px rgba(15,23,42,0.04);
  }
  .day-pill:hover { transform: translateY(-1px); }
  .day-pill.active { background: #10b981; border-color: #10b981; color: #fff; box-shadow: 0 6px 16px rgba(16,185,129,0.25); }

  .today-off-toggle { display: flex; align-items: center; gap: 12px; }
  .toggle-switch { position: relative; display: inline-block; width: 46px; height: 26px; }
  .toggle-switch input { opacity: 0; width: 0; height: 0; }
  .slider {
    position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
    background: #cbd5e1; transition: 0.2s; border-radius: 999px;
  }
  .slider:before {
    position: absolute; content: ""; height: 20px; width: 20px; left: 3px; bottom: 3px;
    background: white; transition: 0.2s; border-radius: 50%; box-shadow: 0 2px 4px rgba(0,0,0,0.12);
  }
  .toggle-switch input:checked + .slider { background: #10b981; }
  .toggle-switch input:checked + .slider:before { transform: translateX(20px); }
  .toggle-title { font-weight: 700; color: #0f172a; margin: 0; }

  .btn-primary {
    background: linear-gradient(135deg, #10b981, #059669); color: white; border: none;
    padding: 10px 18px; border-radius: 999px; font-weight: 600; font-size: 14px; cursor: pointer;
    box-shadow: 0 4px 6px rgba(16,185,129,0.2); transition: 0.2s;
  }
  .btn-primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 12px rgba(16,185,129,0.3); }
  .btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

  .preview-box { padding: 14px; border: 1px solid #e2e8f0; border-radius: 12px; background: #f8fafc; }
  .badge-heading { font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
  .preview-row { display: flex; align-items: center; gap: 14px; flex-wrap: wrap; }
  .divider-vertical { width: 2px; height: 32px; background: #e2e8f0; }
  .text-success { color: #059669; }
  .text-warning { color: #d97706; }

  .info-box { display: flex; gap: 12px; align-items: flex-start; padding: 14px; border: 1px solid #e2e8f0; border-radius: 12px; background: #f8fafc; }
  .icon-circle { width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
  .bg-emerald-light { background: #ecfdf5; } .text-emerald { color: #059669; }
  .info-title { margin: 0; font-size: 14px; font-weight: 700; color: #0f172a; }
  .info-desc { margin: 4px 0 0; font-size: 13px; color: #64748b; }

  .loading { padding: 20px; color: #64748b; }
 
  .info-desc { margin: 4px 0 0; font-size: 13px; color: #64748b; }

  /* PREVIEW */
  .preview-grid { display: grid; gap: 20px; }
  .badge-heading { font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
  .preview-row { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
  .divider-vertical { width: 2px; height: 40px; background: #e2e8f0; }

  /* UTILS */
  .mt-4 { margin-top: 16px; } .mt-6 { margin-top: 24px; }
  .icon-circle { width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
  .bg-emerald-light { background: #ecfdf5; } .text-emerald { color: #059669; }

  .animate-fade-in { opacity: 0; animation: fadeIn 0.6s forwards; }
  .animate-slide-up { opacity: 0; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
  @keyframes fadeIn { to { opacity: 1; } }
  @keyframes slideUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
</style>
