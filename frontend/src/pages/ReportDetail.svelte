<script>
    import { onMount } from "svelte";
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";
    import ReportEditModal from "./ReportEditModal.svelte";

    const { route } = $props();
    let reportId = $state("");
    let report = $state(null);
    let loading = $state(true);

    // Feedback state
    let feedback = $state("");
    let statusUpdate = $state("");
    let submitting = $state(false);

    // Edit Modal State
    let isEditModalOpen = $state(false);

    function openEditModal() {
        isEditModalOpen = true;
    }

    function handleEditSuccess() {
        fetchReport();
    }

    // Helpers
    const statusLabels = {
        pending: "Pending",
        submitted: "Dikirim",
        reviewed: "Direview",
        approved: "Disetujui",
        rejected: "Ditolak",
    };

    function getStatusColor(status) {
        switch (status) {
            case "approved":
            case "reviewed":
                return "bg-emerald-100 text-emerald-700 border-emerald-200";
            case "submitted":
                return "bg-blue-100 text-blue-700 border-blue-200";
            case "rejected":
                return "bg-rose-100 text-rose-700 border-rose-200";
            case "pending":
            default:
                return "bg-amber-100 text-amber-700 border-amber-200";
        }
    }

    function formatDate(value) {
        if (!value) return "-";
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return value;
        return date.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    async function fetchReport() {
        loading = true;
        try {
            const res = await api.getReport(reportId);
            report = res.data;
            feedback = report.feedback || "";
            statusUpdate = report.status || "pending";
        } catch (err) {
            console.error(err);
            alert("Gagal memuat laporan.");
        } finally {
            loading = false;
        }
    }

    async function submitFeedback() {
        if (!feedback.trim()) {
            alert("Feedback tidak boleh kosong.");
            return;
        }
        submitting = true;
        try {
            // Use helper if available, or raw request
            // api.js might not have specific addFeedback, check api.js content later or use updateReport
            // Based on handler, we have AddFeedback endpoint, let's assume api.js has it or we use updateReport
            // Checking api.js in memory: it has updateReport.
            // ReportHandler.AddFeedback expects { feedback: string } and sets status='reviewed'.
            // But we might want to manually set status too.
            // Let's use updateReport for flexibility if needed, or check if addFeedback exists.
            // For now, I'll use updateReport which maps to PUT /reports/:id

            const payload = {
                feedback: feedback,
                status: "reviewed", // Default to reviewed when giving feedback
            };

            await api.updateReport(reportId, payload);
            await fetchReport();
            alert("Feedback berhasil dikirim.");
        } catch (err) {
            console.error(err);
            alert("Gagal mengirim feedback: " + err.message);
        } finally {
            submitting = false;
        }
    }

    $effect(() => {
        const params = route?.result?.path?.params || {};
        if (params?.id && params.id !== reportId) {
            reportId = params.id;
            fetchReport();
        }
    });

    // Check if user can review (Admin/Supervisor)
    let canReview = $derived(
        auth.user &&
            (auth.user.role === "admin" || auth.user.role === "supervisor"),
    );
</script>

<div class="page-bg">
    <div class="container animate-fade-in">
        <!-- Header -->
        <div class="header">
            <a href="/reports" class="btn-back">
                <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><path d="M19 12H5" /><path d="M12 19l-7-7 7-7" /></svg
                >
                Kembali ke Laporan
            </a>

            <button class="btn-edit" onclick={openEditModal}>
                <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><path
                        d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                    /><path
                        d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                    /></svg
                >
                Edit
            </button>
        </div>

        {#if loading}
            <div class="loading-state">
                <div class="spinner"></div>
                <p>Memuat laporan...</p>
            </div>
        {:else if !report}
            <div class="empty-state">
                <div class="empty-icon">🚫</div>
                <p>Laporan tidak ditemukan.</p>
            </div>
        {:else}
            <div class="grid-layout animate-slide-up">
                <!-- LEFT COLUMN: Report Content -->
                <div class="card detail-card">
                    <div class="card-header border-b">
                        <div class="header-top">
                            <span
                                class={`status-badge ${getStatusColor(report.status)}`}
                            >
                                {statusLabels[report.status] || report.status}
                            </span>
                            <span class="type-badge">{report.type}</span>
                        </div>
                        <h1 class="report-title">{report.title}</h1>
                        <p class="report-meta">
                            Oleh <span class="font-semibold"
                                >{report.intern_name}</span
                            >
                            •
                            {formatDate(report.created_at)}
                        </p>
                    </div>
                    <div class="card-body">
                        <div class="meta-grid">
                            <div class="meta-item">
                                <span class="label">Periode Mulai</span>
                                <div class="value">
                                    {formatDate(report.period_start)}
                                </div>
                            </div>
                            <div class="meta-item">
                                <span class="label">Periode Selesai</span>
                                <div class="value">
                                    {formatDate(report.period_end)}
                                </div>
                            </div>
                        </div>

                        <div class="content-box">
                            <h4 class="section-title">Isi Laporan</h4>
                            <p class="report-content">
                                {report.content}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- RIGHT COLUMN: User Info & Feedback -->
                <div class="action-column">
                    <!-- INFORMATION CARD -->
                    <div class="card info-card mb-6">
                        <div class="card-header border-b">
                            <h3 class="flex items-center gap-2">
                                <svg
                                    width="20"
                                    height="20"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                    ><circle cx="12" cy="12" r="10" /><line
                                        x1="12"
                                        y1="16"
                                        x2="12"
                                        y2="12"
                                    /><line
                                        x1="12"
                                        y1="8"
                                        x2="12.01"
                                        y2="8"
                                    /></svg
                                >
                                Informasi
                            </h3>
                        </div>
                        <div class="card-body">
                            <div class="info-group mb-4">
                                <label class="label">Siswa</label>
                                <div class="user-info">
                                    <div class="avatar-small">
                                        {report.intern_name?.charAt(0) || "U"}
                                    </div>
                                    <span class="user-name"
                                        >{report.intern_name}</span
                                    >
                                </div>
                            </div>

                            <div class="info-group mb-4">
                                <label class="label">Dibuat Oleh</label>
                                <div class="value">{report.intern_name}</div>
                            </div>

                            <div class="info-group">
                                <label class="label">Tanggal Dibuat</label>
                                <div class="value">
                                    {formatDate(report.created_at)}
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- FEEDBACK CARD -->
                    <div class="card feedback-card">
                        <div class="card-header border-b bg-slate-50">
                            <h3 class="flex items-center gap-2">
                                <svg
                                    width="20"
                                    height="20"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                    ><path
                                        d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"
                                    /></svg
                                >
                                Feedback & Review
                            </h3>
                        </div>
                        <div class="card-body-feedback">
                            {#if canReview}
                                <div class="form-group mb-4">
                                    <label class="label" for="feedback"
                                        >Feedback Pembimbing</label
                                    >
                                    <textarea
                                        class="input-field textarea"
                                        id="feedback"
                                        rows="6"
                                        bind:value={feedback}
                                        placeholder="Berikan masukan atau catatan untuk laporan ini..."
                                    ></textarea>
                                </div>

                                <div class="actions">
                                    <button
                                        class="btn-primary"
                                        onclick={submitFeedback}
                                        disabled={submitting}
                                    >
                                        {#if submitting}
                                            Menyimpan...
                                        {:else}
                                            <svg
                                                width="18"
                                                height="18"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                class="mr-2"
                                                ><path
                                                    d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"
                                                /><polyline
                                                    points="16 6 12 2 8 6"
                                                /><line
                                                    x1="12"
                                                    y1="2"
                                                    x2="12"
                                                    y2="15"
                                                /></svg
                                            >
                                            Simpan Feedback
                                        {/if}
                                    </button>
                                </div>
                            {:else}
                                <!-- Intern View (Read Only) -->
                                {#if report.feedback}
                                    <div class="feedback-display">
                                        <p class="feedback-text">
                                            {report.feedback}
                                        </p>
                                        <div class="feedback-meta">
                                            Direview oleh: {report.reviewed_by_name ||
                                                "Pembimbing"}
                                        </div>
                                    </div>
                                {:else}
                                    <div class="empty-feedback">
                                        <p>
                                            Belum ada feedback dari pembimbing.
                                        </p>
                                    </div>
                                {/if}
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
        {/if}
    </div>
    <ReportEditModal
        isOpen={isEditModalOpen}
        {report}
        onClose={() => (isEditModalOpen = false)}
        onSuccess={handleEditSuccess}
    />
</div>

<style>
    .page-bg {
        min-height: 100vh;
        background-color: #f8fafc;
        /* padding: 20px; */
    }
    .container {
        max-width: 1000px;
        margin: 0 auto;
    }

    .header {
        margin-bottom: 24px;
    }
    .btn-back {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        color: #64748b;
        font-weight: 500;
        text-decoration: none;
        transition: color 0.2s;
    }
    .btn-back:hover {
        color: #0f172a;
    }

    .grid-layout {
        display: grid;
        grid-template-columns: 1fr;
        gap: 24px;
    }
    @media (min-width: 900px) {
        .grid-layout {
            grid-template-columns: 2fr 1fr;
        }
    }

    .card {
        background: white;
        border-radius: 16px;
        border: 1px solid #e2e8f0;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.02);
        overflow: hidden;
    }
    .card-header {
        padding: 12px;
    }
    .card-body {
        padding: 12px;
    }

    .card-body-feedback {
        padding: 0;
    }
    .border-b {
        border-bottom: 1px solid #f1f5f9;
    }

    .report-title {
        font-size: 24px;
        font-weight: 800;
        color: #0f172a;
        margin: 10px 0 4px 0;
    }
    .report-meta {
        color: #64748b;
        font-size: 14px;
    }

    .header-top {
        display: flex;
        gap: 8px;
        align-items: center;
    }
    .status-badge {
        padding: 4px 12px;
        border-radius: 99px;
        font-size: 12px;
        font-weight: 600;
        text-transform: capitalize;
        border: 1px solid transparent;
    }
    .type-badge {
        padding: 4px 10px;
        border-radius: 6px;
        background: #f1f5f9;
        color: #475569;
        font-size: 11px;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .meta-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
        margin-bottom: 24px;
        padding-bottom: 24px;
        border-bottom: 1px dashed #e2e8f0;
    }
    .meta-item .label {
        font-size: 11px;
        text-transform: uppercase;
        font-weight: 600;
        color: #94a3b8;
        display: block;
        margin-bottom: 4px;
    }
    .meta-item .value {
        font-size: 14px;
        font-weight: 600;
        color: #334155;
    }

    .section-title {
        font-size: 14px;
        font-weight: 700;
        color: #334155;
        text-transform: uppercase;
        margin-bottom: 12px;
        letter-spacing: 0.05em;
    }
    .report-content {
        font-size: 16px;
        line-height: 1.7;
        color: #334155;
        white-space: pre-wrap;
        background: #f8fafc;
        padding: 16px;
        border-radius: 12px;
        border: 1px solid #f1f5f9;
    }

    /* Feedback Section */
    .feedback-card h3 {
        margin: 0;
        font-size: 12px;
        font-weight: 600;
        color: #334155;
    }
    .label {
        display: block;
        font-size: 12px;
        font-weight: 600;
        color: #475569;
        margin-bottom: 8px;
        text-transform: uppercase;
    }
    .input-field {
        width: 100%;
        padding: 12px;
        border: 1px solid #cbd5e1;
        border-radius: 10px;
        font-size: 14px;
        color: #0f172a;
        transition: all 0.2s;
        background: #fff;
        box-sizing: border-box;
    }
    .input-field:focus {
        outline: none;
        border-color: #10b981;
        box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
    }

    .btn-primary {
        background: linear-gradient(135deg, #10b981, #059669);
        color: white;
        padding: 12px;
        border-radius: 999px;
        font-weight: 600;
        font-size: 14px;
        border: none;
        cursor: pointer;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.2s;
        box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    }
    .btn-primary:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
    }
    .btn-primary:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .feedback-display {
        background: #ecfdf5;
        border: 1px solid #d1fae5;
        padding: 16px;
        border-radius: 12px;
    }
    .feedback-text {
        color: #065f46;
        font-size: 15px;
        line-height: 1.6;
        margin: 0 0 12px 0;
    }
    .feedback-meta {
        font-size: 12px;
        color: #059669;
        font-weight: 600;
        text-align: right;
    }

    .empty-feedback {
        text-align: center;
        color: #94a3b8;
        padding: 20px;
        font-style: italic;
        background: #f8fafc;
        border-radius: 12px;
        border: 1px dashed #cbd5e1;
    }

    .loading-state,
    .empty-state {
        text-align: center;
        padding: 60px;
        color: #94a3b8;
    }
    .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid #e2e8f0;
        border-top-color: #10b981;
        border-radius: 50%;
        margin: 0 auto 16px;
        animation: spin 1s linear infinite;
    }
    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    .animate-fade-in {
        opacity: 0;
        animation: fadeIn 0.6s ease-out forwards;
    }
    .animate-slide-up {
        opacity: 0;
        animation: slideUp 0.6s ease-out forwards;
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

    /* INFO CARD STYLES */
    .info-card h3 {
        margin: 0;
        font-size: 16px;
        font-weight: 600;
        color: #334155;
    }
    .user-info {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    .avatar-small {
        width: 32px;
        height: 32px;
        background: #8b5cf6;
        color: white;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 600;
        font-size: 14px;
    }
    .user-name {
        font-weight: 600;
        color: #334155;
        font-size: 14px;
    }
    .info-group .label {
        margin-bottom: 4px;
        color: #94a3b8;
    }
    .info-group .value {
        font-weight: 600;
        color: #334155;
        font-size: 14px;
    }
    .mb-6 {
        margin-bottom: 24px;
    }
    .mb-4 {
        margin-bottom: 16px;
    }

    /* HEADER EDIT BUTTON */
    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .btn-edit {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        background: #f59e0b;
        color: white;
        border: none;
        padding: 8px 16px;
        border-radius: 999px;
        font-weight: 600;
        font-size: 13px;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 2px 5px rgba(245, 158, 11, 0.2);
    }
    .btn-edit:hover {
        background: #d97706;
        transform: translateY(-1px);
    }
</style>
