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

<div class="page-bg">
    <div class="container animate-fade-in">
        <!-- EXPORT SECTION -->
        <div class="section-title animate-slide-up">
            <h3>Export Data</h3>
            <p>Unduh laporan dalam format Excel (.xlsx)</p>
        </div>

        <div class="grid-layout animate-slide-up">
            <!-- Card: Export Interns -->
            <div class="card">
                <div class="card-header">
                    <div class="icon-circle bg-emerald">
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
                            /><circle cx="9" cy="7" r="4" /><path
                                d="M23 21v-2a4 4 0 0 0-3-3.87"
                            /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg
                        >
                    </div>
                    <h4>Data Peserta</h4>
                </div>
                <div class="card-body">
                    <div class="form-group">
                        <label class="label" for="export-intern-status"
                            >Status</label
                        >
                        <select
                            class="input-field"
                            id="export-intern-status"
                            bind:value={internStatus}
                        >
                            <option value="">Semua Status</option>
                            <option value="active">Aktif</option>
                            <option value="pending">Pending</option>
                            <option value="completed">Selesai</option>
                            <option value="cancelled">Dibatalkan</option>
                        </select>
                    </div>
                    <button
                        class="btn-outline w-full mt-4"
                        onclick={exportInterns}
                        disabled={isExporting}
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                            /><polyline points="7 10 12 15 17 10" /><line
                                x1="12"
                                y1="15"
                                x2="12"
                                y2="3"
                            /></svg
                        >
                        Download Excel
                    </button>
                </div>
            </div>

            <!-- Card: Export Attendance -->
            <div class="card">
                <div class="card-header">
                    <div class="icon-circle bg-blue">
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><rect
                                x="3"
                                y="4"
                                width="18"
                                height="18"
                                rx="2"
                                ry="2"
                            /><line x1="16" y1="2" x2="16" y2="6" /><line
                                x1="8"
                                y1="2"
                                x2="8"
                                y2="6"
                            /><line x1="3" y1="10" x2="21" y2="10" /></svg
                        >
                    </div>
                    <h4>Data Presensi</h4>
                </div>
                <div class="card-body">
                    <div class="grid-2">
                        <div class="form-group">
                            <label class="label" for="att-start">Dari</label>
                            <input
                                class="input-field"
                                type="date"
                                id="att-start"
                                bind:value={attendanceStart}
                            />
                        </div>
                        <div class="form-group">
                            <label class="label" for="att-end">Sampai</label>
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
                            <option value="absent">Tanpa Keterangan</option>
                        </select>
                    </div>
                    <button
                        class="btn-outline w-full mt-4"
                        onclick={exportAttendance}
                        disabled={isExporting}
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                            /><polyline points="7 10 12 15 17 10" /><line
                                x1="12"
                                y1="15"
                                x2="12"
                                y2="3"
                            /></svg
                        >
                        Download Excel
                    </button>
                </div>
            </div>

            <!-- Card: Export Tasks -->
            <div class="card">
                <div class="card-header">
                    <div class="icon-circle bg-amber">
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                            /><polyline points="14 2 14 8 20 8" /><line
                                x1="16"
                                y1="13"
                                x2="8"
                                y2="13"
                            /><line x1="16" y1="17" x2="8" y2="17" /><polyline
                                points="10 9 9 9 8 9"
                            /></svg
                        >
                    </div>
                    <h4>Laporan Tugas</h4>
                </div>
                <div class="card-body">
                    <div class="form-group">
                        <label class="label" for="task-intern"
                            >Filter Intern</label
                        >
                        <select
                            class="input-field"
                            id="task-intern"
                            bind:value={taskIntern}
                        >
                            <option value="">Semua Intern</option>
                            {#each interns as i}
                                <option value={i.id}>{i.full_name}</option>
                            {/each}
                        </select>
                    </div>
                    <div class="form-group mt-3">
                        <label class="label" for="task-status"
                            >Status Tugas</label
                        >
                        <select
                            class="input-field"
                            id="task-status"
                            bind:value={taskStatus}
                        >
                            <option value="">Semua</option>
                            <option value="pending">Pending</option>
                            <option value="in_progress">In Progress</option>
                            <option value="submitted">Submitted</option>
                            <option value="completed">Completed</option>
                        </select>
                    </div>
                    <button
                        class="btn-outline w-full mt-4"
                        onclick={exportTasks}
                        disabled={isExporting}
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                            /><polyline points="7 10 12 15 17 10" /><line
                                x1="12"
                                y1="15"
                                x2="12"
                                y2="3"
                            /></svg
                        >
                        Download Excel
                    </button>
                </div>
            </div>
        </div>

        <!-- IMPORT SECTION -->
        <div
            class="import-section animate-slide-up"
            style="animation-delay: 0.1s;"
        >
            <div class="card import-card">
                <div class="import-header">
                    <div class="header-text">
                        <h3>Import Data Peserta</h3>
                        <p>
                            Tambahkan banyak peserta sekaligus menggunakan file
                            Excel/CSV.
                        </p>
                    </div>
                    <button class="btn-text" onclick={downloadTemplate}>
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><path
                                d="M3 15v4c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2v-4M17 9l-5 5-5-5M12 12.8V2.5"
                            /></svg
                        >
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
                                <svg
                                    width="32"
                                    height="32"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                    class="text-slate-400 mb-2"
                                    ><path
                                        d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                                    /><polyline points="17 8 12 3 7 8" /><line
                                        x1="12"
                                        y1="3"
                                        x2="12"
                                        y2="15"
                                    /></svg
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
                            <select
                                class="input-field"
                                id="import-supervisor"
                                bind:value={importSupervisor}
                            >
                                <option value="">-- Tidak Ditentukan --</option>
                                {#each supervisors as s}
                                    <option value={s.user_id}
                                        >{s.full_name}</option
                                    >
                                {/each}
                            </select>
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
                            class="btn-primary"
                            onclick={importInterns}
                            disabled={!importFile || isImporting}
                        >
                            {isImporting ? "Mengimpor..." : "Mulai Import"}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    :global(body) {
        font-family: "Geist", "Inter", sans-serif;
        color: #0f172a;
    }

    .page-bg {
        min-height: 100vh;
        background-color: #f8fafc;
        padding: 0;
    }

    .container {
        max-width: 1000px;
        margin: 0 auto;
    }

    /* HEADER */
    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 40px;
    }
    .title {
        font-size: 28px;
        font-weight: 800;
        color: #0f172a;
        margin: 0 0 6px 0;
        letter-spacing: -0.02em;
    }
    .subtitle {
        color: #64748b;
        font-size: 15px;
        margin: 0;
    }
    .header-icon {
        width: 48px;
        height: 48px;
        background: #ffffff;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #10b981;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
        border: 1px solid #e2e8f0;
    }

    .section-title {
        margin-bottom: 24px;
    }
    .section-title h3 {
        font-size: 18px;
        font-weight: 600;
        color: #1e293b;
        margin: 0 0 4px 0;
    }
    .section-title p {
        font-size: 13px;
        color: #64748b;
        margin: 0;
    }

    /* CARD STYLES */
    .card {
        background: white;
        border-radius: 16px;
        border: 1px solid #e2e8f0;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
        overflow: hidden;
        transition:
            transform 0.2s,
            box-shadow 0.2s;
    }
    .card:hover {
        /* transform: translateY(-2px); */
        box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
    }

    .card-header {
        padding: 16px 20px;
        border-bottom: 1px solid #f1f5f9;
        display: flex;
        align-items: center;
        gap: 12px;
    }
    .card-header h4 {
        font-size: 15px;
        font-weight: 600;
        color: #1e293b;
        margin: 0;
    }
    .card-body {
        padding: 20px;
    }

    .icon-circle {
        width: 32px;
        height: 32px;
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
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
        grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        gap: 24px;
        margin-bottom: 40px;
    }
    .grid-2 {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
    }

    /* FORM ELEMENTS */
    .form-group {
        margin-bottom: 12px;
    }
    .label {
        display: block;
        font-size: 12px;
        font-weight: 600;
        color: #475569;
        margin-bottom: 6px;
        text-transform: uppercase;
        letter-spacing: 0.02em;
    }

    .input-field {
        width: 100%;
        padding: 10px 12px;
        border: 1px solid #cbd5e1;
        border-radius: 8px;
        font-size: 13px;
        color: #1e293b;
        transition: all 0.2s;
        box-sizing: border-box;
        background: #fff;
    }
    .input-field:focus {
        outline: none;
        border-color: #10b981;
        box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
    }

    /* BUTTONS */
    .btn-outline {
        width: 100%;
        padding: 10px;
        background: white;
        border: 1px solid #cbd5e1;
        color: #475569;
        font-weight: 600;
        font-size: 13px;
        border-radius: 8px;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        transition: all 0.2s;
    }
    .btn-outline:hover {
        border-color: #10b981;
        color: #059669;
        background: #ecfdf5;
    }
    .btn-outline:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .btn-primary {
        padding: 10px 24px;
        background: linear-gradient(135deg, #10b981, #059669);
        color: white;
        border: none;
        font-weight: 600;
        font-size: 14px;
        border-radius: 999px;
        cursor: pointer;
        transition: all 0.2s;
    }
    .btn-primary:hover {
        box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
        /* transform: translateY(-1px); */
    }
    .btn-primary:disabled {
        background: #94a3b8;
        cursor: not-allowed;
        box-shadow: none;
        transform: none;
    }

    .btn-text {
        background: transparent;
        border: none;
        color: #10b981;
        font-weight: 600;
        font-size: 13px;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 6px;
    }
    .btn-text:hover {
        text-decoration: underline;
    }

    /* IMPORT SECTION */
    .import-section {
        margin-top: 24px;
    }
    .import-card {
        border: 1px dashed #cbd5e1;
    }
    .import-header {
        padding: 24px;
        border-bottom: 1px solid #f1f5f9;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .header-text h3 {
        margin: 0 0 4px 0;
        font-size: 16px;
        color: #1e293b;
        font-weight: 600;
    }
    .header-text p {
        margin: 0;
        font-size: 13px;
        color: #64748b;
    }
    .import-body {
        padding: 24px;
    }

    .file-drop {
        display: block;
        width: 100%;
        border: 2px dashed #e2e8f0;
        border-radius: 12px;
        padding: 32px;
        text-align: center;
        cursor: pointer;
        transition: all 0.2s;
        background: #f8fafc;
    }
    .file-drop:hover {
        border-color: #10b981;
        background: #ecfdf5;
    }
    .drop-content {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .file-name {
        font-weight: 600;
        color: #059669;
    }
    .placeholder {
        color: #64748b;
        font-size: 13px;
    }

    .action-footer {
        display: flex;
        justify-content: flex-end;
        align-items: center;
        gap: 16px;
        margin-top: 8px;
    }
    .result-badge {
        font-size: 13px;
        color: #475569;
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .dot.green {
        width: 8px;
        height: 8px;
        background: #10b981;
        border-radius: 50%;
    }

    /* UTILS */
    .mt-4 {
        margin-top: 16px;
    }
    .mt-3 {
        margin-top: 12px;
    }
    .w-full {
        width: 100%;
    }

    /* ANIMATIONS */
    .animate-fade-in {
        opacity: 0;
        animation: fadeIn 0.6s ease-out forwards;
    }
    .animate-slide-up {
        opacity: 0;
        animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    }
    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }
    @keyframes slideUp {
        from {
            opacity: 0;
            transform: translateY(20px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
</style>
