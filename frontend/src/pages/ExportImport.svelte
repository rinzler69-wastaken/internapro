<script>
    import { onMount } from "svelte";
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";

    // State
    let interns = $state([]);
    let supervisors = $state([]);
    let loading = $state(false);
    let isExporting = $state(false);
    let isImporting = $state(false);

    // Filter State
    let internStatus = $state("");
    let attendanceIntern = $state("");
    let attendanceStatus = $state("");
    let attendanceStart = $state("");
    let attendanceEnd = $state("");
    let taskIntern = $state("");
    let taskStatus = $state("");

    // Import State
    let importFile = $state(null);
    let importSupervisor = $state("");
    let importResult = $state(null);
    let fileInput;

    function downloadBlob(blob, filename) {
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(url);
    }

    async function fetchData() {
        loading = true;
        try {
            const internRes = await api.getInterns({ page: 1, limit: 200 });
            interns = internRes.data || [];
            if (auth.user?.role === "admin") {
                const supRes = await api.getSupervisors({
                    page: 1,
                    limit: 200,
                });
                supervisors = supRes.data || [];
            }
        } catch (err) {
            console.error(err);
        } finally {
            loading = false;
        }
    }

    async function exportInterns() {
        isExporting = true;
        try {
            const { blob, filename } = await api.exportInterns({
                status: internStatus,
            });
            downloadBlob(blob, filename || "interns.xlsx");
        } catch (e) {
            alert("Gagal export");
        } finally {
            isExporting = false;
        }
    }

    async function exportAttendance() {
        isExporting = true;
        try {
            const params = {};
            if (attendanceIntern) params.intern_id = attendanceIntern;
            if (attendanceStatus) params.status = attendanceStatus;
            if (attendanceStart) params.start_date = attendanceStart;
            if (attendanceEnd) params.end_date = attendanceEnd;
            const { blob, filename } = await api.exportAttendances(params);
            downloadBlob(blob, filename || "attendance.xlsx");
        } catch (e) {
            alert("Gagal export");
        } finally {
            isExporting = false;
        }
    }

    async function exportTasks() {
        isExporting = true;
        try {
            const params = {};
            if (taskIntern) params.intern_id = taskIntern;
            if (taskStatus) params.status = taskStatus;
            const { blob, filename } = await api.exportTasks(params);
            downloadBlob(blob, filename || "tasks.xlsx");
        } catch (e) {
            alert("Gagal export");
        } finally {
            isExporting = false;
        }
    }

    async function downloadTemplate() {
        try {
            const { blob, filename } = await api.downloadImportTemplate();
            downloadBlob(blob, filename || "template.xlsx");
        } catch (e) {
            alert("Gagal download template");
        }
    }

    async function importInterns() {
        if (!importFile) return;
        isImporting = true;
        try {
            const result = await api.importInterns(
                importFile,
                importSupervisor || null,
            );
            importResult = result?.data || result;
            await fetchData();
            // Reset file input
            importFile = null;
            if (fileInput) fileInput.value = "";
            alert(`Import berhasil: ${importResult.imported || 0} data masuk.`);
        } catch (e) {
            alert(e.message || "Gagal import");
        } finally {
            isImporting = false;
        }
    }

    function handleFileSelect(e) {
        const file = e.target.files?.[0];
        if (file) importFile = file;
    }

    onMount(fetchData);
</script>

<svelte:head>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link
        rel="preconnect"
        href="https://fonts.gstatic.com"
        crossorigin="anonymous"
    />
    <link
        href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"
        rel="stylesheet"
    />
    <link
        rel="stylesheet"
        href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@48,300,0,0"
    />
</svelte:head>

<div class="page">
    <div class="shell">
        <!-- EXPORT SECTION -->
        <div class="section-block">
            <h3 class="section-header pt-4">Export Data</h3>

            <div class="grid-layout">
                <!-- Card: Export Interns -->
                <div class="card">
                    <div class="card-head">
                        <div class="icon-wrap bg-emerald">
                            <span class="material-symbols-outlined">group</span>
                        </div>
                        <h3>Data Peserta</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="label" for="export-intern-status"
                                >Status</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field"
                                    id="export-intern-status"
                                    bind:value={internStatus}
                                >
                                    <option value="">Semua Status</option>
                                    <option value="active">Aktif</option>
                                    <option value="pending">Pending</option>
                                    <option value="completed">Selesai</option>
                                    <option value="cancelled">Dibatalkan</option
                                    >
                                </select>
                            </div>
                        </div>
                        <button
                            class="btn ghost w-full mt-4"
                            onclick={exportInterns}
                            disabled={isExporting}
                        >
                            <span class="material-symbols-outlined"
                                >download</span
                            >
                            Download Excel
                        </button>
                    </div>
                </div>

                <!-- Card: Export Attendance -->
                <div class="card">
                    <div class="card-head">
                        <div class="icon-wrap bg-blue">
                            <span class="material-symbols-outlined"
                                >calendar_month</span
                            >
                        </div>
                        <h3>Data Presensi</h3>
                    </div>
                    <div class="card-body">
                        <div class="grid-2">
                            <div class="form-group">
                                <label class="label" for="att-start">Dari</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    id="att-start"
                                    bind:value={attendanceStart}
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="att-end">Sampai</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    id="att-end"
                                    bind:value={attendanceEnd}
                                />
                            </div>
                        </div>
                        <div class="form-group mt-3">
                            <label class="label" for="att-status"
                                >Status Kehadiran</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field"
                                    id="att-status"
                                    bind:value={attendanceStatus}
                                >
                                    <option value="">Semua</option>
                                    <option value="present">Hadir</option>
                                    <option value="late">Terlambat</option>
                                    <option value="sick">Sakit</option>
                                    <option value="permission">Izin</option>
                                    <option value="absent"
                                        >Tanpa Keterangan</option
                                    >
                                </select>
                            </div>
                        </div>
                        <button
                            class="btn ghost w-full mt-4"
                            onclick={exportAttendance}
                            disabled={isExporting}
                        >
                            <span class="material-symbols-outlined"
                                >download</span
                            >
                            Download Excel
                        </button>
                    </div>
                </div>

                <!-- Card: Export Tasks -->
                <div class="card">
                    <div class="card-head">
                        <div class="icon-wrap bg-amber">
                            <span class="material-symbols-outlined"
                                >assignment</span
                            >
                        </div>
                        <h3>Laporan Tugas</h3>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="label" for="task-intern"
                                >Filter Intern</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field"
                                    id="task-intern"
                                    bind:value={taskIntern}
                                >
                                    <option value="">Semua Intern</option>
                                    {#each interns as i}
                                        <option value={i.id}
                                            >{i.full_name}</option
                                        >
                                    {/each}
                                </select>
                            </div>
                        </div>
                        <div class="form-group mt-3">
                            <label class="label" for="task-status"
                                >Status Tugas</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field"
                                    id="task-status"
                                    bind:value={taskStatus}
                                >
                                    <option value="">Semua</option>
                                    <option value="pending">Pending</option>
                                    <option value="in_progress"
                                        >In Progress</option
                                    >
                                    <option value="submitted">Submitted</option>
                                    <option value="completed">Completed</option>
                                </select>
                            </div>
                        </div>
                        <button
                            class="btn ghost w-full mt-4"
                            onclick={exportTasks}
                            disabled={isExporting}
                        >
                            <span class="material-symbols-outlined"
                                >download</span
                            >
                            Download Excel
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- IMPORT SECTION -->
        <div class="section-block">
            <h3 class="section-header">Import Data</h3>
            <div class="card import-card">
                <div class="import-header">
                    <div class="header-text">
                        <h3>Import Data Peserta</h3>
                        <p>
                            Tambahkan banyak peserta sekaligus menggunakan file
                            Excel/CSV.
                        </p>
                    </div>
                    <button class="btn ghost small" onclick={downloadTemplate}>
                        <span class="material-symbols-outlined">download</span>
                        Download Template
                    </button>
                </div>

                <div class="import-body">
                    <div class="form-group mb-4">
                        <label class="label" for="file-upload"
                            >Upload File (.xlsx, .csv)</label
                        >
                        <label class="file-drop" for="file-upload">
                            <div class="drop-content">
                                <span
                                    class="material-symbols-outlined text-slate-400 mb-2"
                                    style="font-size: 32px;">upload_file</span
                                >
                                {#if importFile}
                                    <span class="file-name"
                                        >{importFile.name}</span
                                    >
                                {:else}
                                    <span class="placeholder"
                                        >Klik untuk memilih file</span
                                    >
                                {/if}
                            </div>
                            <input
                                id="file-upload"
                                type="file"
                                accept=".xlsx,.xls,.csv"
                                hidden
                                bind:this={fileInput}
                                onchange={handleFileSelect}
                            />
                        </label>
                    </div>

                    {#if auth.user?.role === "admin"}
                        <div class="form-group mb-4">
                            <label class="label" for="import-supervisor"
                                >Assign Pembimbing (Opsional)</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field"
                                    id="import-supervisor"
                                    bind:value={importSupervisor}
                                >
                                    <option value=""
                                        >-- Tidak Ditentukan --</option
                                    >
                                    {#each supervisors as s}
                                        <option value={s.user_id}
                                            >{s.full_name}</option
                                        >
                                    {/each}
                                </select>
                            </div>
                        </div>
                    {/if}

                    <div class="action-footer">
                        {#if importResult}
                            <div class="result-badge">
                                <span class="dot green"></span>
                                Import Selesai: <b>{importResult.imported}</b>
                                sukses, <b>{importResult.skipped}</b> dilewati.
                            </div>
                        {/if}
                        <button
                            class="btn primary"
                            onclick={importInterns}
                            disabled={!importFile || isImporting}
                        >
                            {#if isImporting}
                                <span class="spinner-small"></span>
                                Mengimpor...
                            {:else}
                                <span class="material-symbols-outlined"
                                    >publish</span
                                >
                                Mulai Import
                            {/if}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    :global(body) {
        font-family:
            "Plus Jakarta Sans",
            "Inter",
            system-ui,
            -apple-system,
            sans-serif;
        background: #f8fafc;
        color: #0f172a;
    }

    .page {
        min-height: 100vh;
        padding: 0px;
    }

    .shell {
        max-width: 1200px;
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        gap: 32px;
        padding: 0;
    }

    /* HEADER */
    .header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 12px;
        margin-bottom: 8px;
    }

    .title {
        font-size: 24px;
        font-weight: 800;
        color: #0f172a;
        margin: 0 0 4px 0;
        letter-spacing: -0.02em;
    }
    .subtitle {
        color: #64748b;
        font-size: 15px;
        margin: 0;
    }

    .section-block {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .section-header {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        font-weight: 700;
        color: #6366f1;
        margin: 0;
    }

    /* CARD STYLES */
    .card {
        background: rgba(255, 255, 255, 0.9);
        border: 1px solid oklch(92.9% 0.013 255.508);
        border-radius: 20px;
        box-shadow: 0 10px 30px rgba(15, 23, 42, 0.04);
        padding: 20px;
        transition:
            transform 0.2s,
            box-shadow 0.2s;
        display: flex;
        flex-direction: column;
    }
    .card:hover {
        /* transform: translateY(-2px); */
        box-shadow: 0 20px 40px rgba(15, 23, 42, 0.06);
    }

    .card-head {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 20px;
        padding-bottom: 16px;
        border-bottom: 1px dashed #e2e8f0;
    }
    .card-head h3 {
        font-size: 16px;
        font-weight: 700;
        color: #0f172a;
        margin: 0;
    }
    .card-body {
        flex: 1;
        display: flex;
        flex-direction: column;
    }

    .icon-wrap {
        width: 36px;
        height: 36px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .icon-wrap span {
        font-size: 20px;
    }
    .bg-emerald {
        background: #ecfdf5;
        color: #059669;
    }
    .bg-blue {
        background: #eff6ff;
        color: #2563eb;
    }
    .bg-amber {
        background: #fffbeb;
        color: #d97706;
    }

    /* GRID LAYOUT */
    .grid-layout {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 20px;
    }
    .grid-2 {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 12px;
    }

    /* FORM ELEMENTS */
    .form-group {
        margin-bottom: 16px;
    }
    .label {
        display: block;
        font-size: 12px;
        font-weight: 700;
        color: #94a3b8;
        margin-bottom: 8px;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .input-field {
        width: 100%;
        padding: 10px 14px;
        border: 1px solid #e2e8f0;
        border-radius: 10px;
        font-size: 14px;
        color: #0f172a;
        transition: all 0.2s;
        box-sizing: border-box;
        background: #f8fafc;
        font-weight: 500;
    }
    .input-field:focus {
        outline: none;
        border-color: #6366f1;
        box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
        background: white;
    }

    .select-wrapper {
        position: relative;
    }
    /* Add custom arrow for select if needed, or rely on browser default styled nicely */
    select.input-field {
        appearance: none;
        background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
        background-position: right 0.5rem center;
        background-repeat: no-repeat;
        background-size: 1.5em 1.5em;
        padding-right: 2.5rem;
    }

    /* BUTTONS */
    .btn {
        padding: 10px 16px;
        border-radius: 999px;
        font-weight: 700;
        font-size: 14px;
        cursor: pointer;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        border: 1px solid transparent;
        transition: all 0.2s ease;
        text-decoration: none;
    }
    .btn:hover {
        transform: translateY(-1px);
    }
    .btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
        transform: none;
    }

    .btn.primary {
        color: oklch(12.9% 0.042 264.695);
        background: oklch(96.8% 0.007 247.896);
        border-color: oklch(82.3% 0.034 264.695);
        /* box-shadow: 0 10px 30px rgba(16, 185, 129, 0.25); */
    }
    .btn.primary:hover {
        box-shadow: 0 8px 16px rgba(0, 0, 0, 0.06);
    }

    .btn.ghost {
        background: #fff;
        border: 2px solid #e2e8f0;
        color: #0f172a;
        font-weight: 600;
    }
    .btn.ghost:hover {
        border-color: #cbd5e1;
        background: #f8fafc;
    }

    .btn.small {
        padding: 6px 12px;
        font-size: 12px;
    }

    .w-full {
        width: 100%;
    }
    .mt-4 {
        margin-top: 16px;
    }
    .mt-3 {
        margin-top: 12px;
    }
    .mb-4 {
        margin-bottom: 16px;
    }

    /* IMPORT SECTION */
    .import-card {
        border: 1px dashed oklch(92.9% 0.013 255.508);
        border-radius: 20px;
    }
    .import-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
        padding-bottom: 16px;
        border-bottom: 1px solid #f1f5f9;
        flex-wrap: wrap;
        gap: 12px;
    }
    .header-text h3 {
        margin: 0 0 4px 0;
        font-size: 16px;
        color: #0f172a;
        font-weight: 700;
    }
    .header-text p {
        margin: 0;
        font-size: 14px;
        color: #64748b;
    }
    .import-body {
        padding: 0;
    }

    .file-drop {
        display: block;
        width: 100%;
        border: 2px dashed #e2e8f0;
        border-radius: 16px;
        padding: 40px;
        text-align: center;
        cursor: pointer;
        transition: all 0.2s;
        background: #f8fafc;
    }
    .file-drop:hover {
        border-color: #6366f1;
        background: #eef2ff;
    }
    .drop-content {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .file-name {
        font-weight: 700;
        color: #059669;
        font-size: 14px;
    }
    .placeholder {
        color: #64748b;
        font-size: 14px;
        font-weight: 500;
    }

    .action-footer {
        display: flex;
        justify-content: flex-end;
        align-items: center;
        gap: 16px;
        margin-top: 20px;
        flex-wrap: wrap;
    }
    .result-badge {
        font-size: 13px;
        color: #475569;
        background: #f1f5f9;
        padding: 6px 12px;
        border-radius: 8px;
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        display: inline-block;
    }
    .dot.green {
        background: #10b981;
    }

    .spinner-small {
        width: 16px;
        height: 16px;
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-top-color: white;
        border-radius: 50%;
        animation: spin 1s linear infinite;
    }
    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    @media (max-width: 640px) {
        .header {
            flex-direction: column;
            align-items: flex-start;
        }
        .import-header {
            flex-direction: column;
            align-items: flex-start;
        }
    }
</style>
