<script>
    import { onMount } from "svelte";
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";
    import Chart from "chart.js/auto";
    import AssessmentEditModal from "./AssessmentEditModal.svelte";

    let { route } = $props();
    let assessmentId = $state(null);
    let assessment = $state(null);
    let loading = $state(true);
    let error = $state(null);
    let showEditModal = $state(false);
    let chartInstance = null;
    let canvasRef;

    const assessmentLabels = {
        very_good: "Sangat Baik",
        good: "Baik",
        not_good: "Kurang Baik",
        very_bad: "Tidak Baik",
    };

    // Helper: Determine Grade
    function getGrade(score) {
        if (score >= 85)
            return {
                grade: "A",
                label: assessmentLabels.very_good,
                color: "text-blue-600",
                tintBg: "#eff6ff", // blue-50
                borderColor: "#bfdbfe", // blue-200
                chartColor: "rgb(37, 99, 235)",
                chartBg: "rgba(37, 99, 235, 0.2)",
            };
        if (score >= 75)
            return {
                grade: "B",
                label: assessmentLabels.good,
                color: "text-emerald-600",
                tintBg: "#ecfdf5", // emerald-50
                borderColor: "#a7f3d0", // emerald-200
                chartColor: "rgb(5, 150, 105)",
                chartBg: "rgba(5, 150, 105, 0.2)",
            };
        if (score >= 60)
            return {
                grade: "C",
                label: assessmentLabels.not_good,
                color: "text-yellow-600",
                tintBg: "#fefce8", // yellow-50
                borderColor: "#fde047", // yellow-200
                chartColor: "rgb(202, 138, 4)",
                chartBg: "rgba(202, 138, 4, 0.2)",
            };
        return {
            grade: "D",
            label: assessmentLabels.very_bad,
            color: "text-orange-600",
            tintBg: "#fff7ed", // orange-50
            borderColor: "#fed7aa", // orange-200
            chartColor: "rgb(234, 88, 12)",
            chartBg: "rgba(234, 88, 12, 0.2)",
        };
    }

    async function loadAssessment() {
        if (!assessmentId) return;
        loading = true;
        try {
            const res = await api.getAssessment(assessmentId);
            assessment = res.data;
            initChart();
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
            // Re-init chart if it wasn't ready
            if (assessment && !chartInstance) {
                setTimeout(initChart, 100);
            }
        }
    }

    $effect(() => {
        const params = route?.result?.path?.params || {};
        if (params?.id && params.id !== assessmentId) {
            assessmentId = params.id;
            loadAssessment();
        }
    });

    function initChart() {
        if (!assessment || !canvasRef) return;
        if (chartInstance) chartInstance.destroy();

        const ctx = canvasRef.getContext("2d");
        chartInstance = new Chart(ctx, {
            type: "radar",
            data: {
                labels: [
                    "Kualitas",
                    "Kecepatan",
                    "Inisiatif",
                    "Kerjasama",
                    "Komunikasi",
                ],
                datasets: [
                    {
                        label: "Kompetensi",
                        data: [
                            assessment.quality_score,
                            assessment.speed_score,
                            assessment.initiative_score,
                            assessment.teamwork_score,
                            assessment.communication_score,
                        ],
                        fill: true,
                        backgroundColor: "rgba(99, 102, 241, 0.2)",
                        borderColor: "rgb(99, 102, 241)",
                        pointBackgroundColor: "rgb(99, 102, 241)",
                        pointBorderColor: "#fff",
                        pointHoverBackgroundColor: "#fff",
                        pointHoverBorderColor: "rgb(99, 102, 241)",
                    },
                ],
            },
            options: {
                scales: {
                    r: {
                        angleLines: { color: "#e2e8f0" },
                        grid: { color: "#e2e8f0" },
                        suggestedMin: 0,
                        suggestedMax: 100,
                        pointLabels: {
                            font: {
                                size: 12,
                                family: "'Inter', sans-serif",
                                weight: "bold",
                            },
                            color: "#64748b",
                        },
                        ticks: {
                            backdropColor: "transparent",
                            showLabelBackdrop: false,
                            stepSize: 20,
                        },
                    },
                },
                plugins: {
                    legend: { display: false },
                },
            },
        });
    }

    onMount(() => {
        return () => {
            if (chartInstance) chartInstance.destroy();
        };
    });

    function handleEditSuccess() {
        loadAssessment();
    }
</script>

<div class="page-container">
    <div class="header">
        <div class="flex items-center gap-4">
            <a href="/assessments" class="btn-back">
                <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                >
                    <path d="M19 12H5M12 19l-7-7 7-7" />
                </svg>
            </a>
            <div>
                {#if assessment}
                    <h1 class="page-title">
                        Penilaian – {assessment.intern_name}
                    </h1>
                    <div class="meta">
                        <span class="badge"
                            >{assessmentLabels[assessment.category] ||
                                assessment.category ||
                                "Umum"}</span
                        >
                        <span class="date">{assessment.assessment_date}</span>
                    </div>
                {:else}
                    <h1 class="page-title">Detail Penilaian</h1>
                {/if}
            </div>
        </div>

        <!-- Desktop Edit Button -->
        {#if assessment && (auth.user?.role === "admin" || auth.user?.role === "supervisor" || auth.user?.role === "pembimbing")}
            <button
                class="btn-edit desktop-only"
                onclick={() => (showEditModal = true)}
            >
                <svg
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                >
                    <path
                        d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                    /><path
                        d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                    />
                </svg>
                Edit Penilaian
            </button>
        {/if}
    </div>

    <!-- Mobile Edit Button -->
    {#if assessment && (auth.user?.role === "admin" || auth.user?.role === "supervisor" || auth.user?.role === "pembimbing")}
        <button
            class="btn-edit-full mobile-only"
            onclick={() => (showEditModal = true)}
        >
            <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
            >
                <path
                    d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                /><path
                    d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                />
            </svg>
            Edit Penilaian
        </button>
    {/if}

    {#if loading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Memuat data penilaian...</p>
        </div>
    {:else if error}
        <div class="error-state">
            <p>{error}</p>
            <button class="btn-primary" onclick={loadAssessment}
                >Coba Lagi</button
            >
        </div>
    {:else if assessment}
        {@const gradeInfo = getGrade(assessment.score)}

        <div class="dashboard-grid mt-6">
            <!-- Left Column: Score & Radar -->
            <div class="left-col">
                <!-- Score Card -->
                <div
                    class="card score-card"
                    style="background-color: {gradeInfo.tintBg}; border-color: {gradeInfo.borderColor};"
                >
                    <div class="score-circle">
                        <span class="grade {gradeInfo.color}"
                            >{gradeInfo.grade}</span
                        >
                        <span class="value {gradeInfo.color}"
                            >{assessment.score}<span class="total text-gray-400"
                                >/100</span
                            ></span
                        >
                    </div>
                    <div class="assessor-info text-left">
                        <div class="label text-gray-500">Dinilai Oleh</div>
                        <div class="assessor">
                            <div
                                class="avatar {gradeInfo.bg} {gradeInfo.color}"
                            >
                                {assessment.assessor_name?.[0] || "A"}
                            </div>
                            <div class="details">
                                <span class="name text-gray-900"
                                    >{assessment.assessor_name}</span
                                >
                                <span class="role text-gray-500"
                                    >{assessment.assessor_role ||
                                        "Supervisor"}</span
                                >
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Radar Chart -->
                <div class="card radar-card">
                    <h3>Analisis Radar</h3>
                    <div class="chart-wrapper">
                        <canvas bind:this={canvasRef}></canvas>
                    </div>
                </div>
            </div>

            <!-- Right Column: Competency & Comments -->
            <div class="right-col">
                <!-- Competency Detail -->
                <div class="card competency-card">
                    <h3>
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            class="mr-2"
                            ><path d="M12 20V10" /><path d="M18 20V4" /><path
                                d="M6 20v-4"
                            /></svg
                        > Detail Kompetensi
                    </h3>

                    <div class="skill-list">
                        <div class="skill-item">
                            <div class="skill-header">
                                <span>Kualitas Kerja</span>
                                <span class="score"
                                    >{assessment.quality_score}/100</span
                                >
                            </div>
                            <div class="progress-track">
                                <div
                                    class="progress-bar bg-indigo-500"
                                    style="width: {assessment.quality_score}%"
                                ></div>
                            </div>
                        </div>
                        <div class="skill-item">
                            <div class="skill-header">
                                <span>Kecepatan</span>
                                <span class="score"
                                    >{assessment.speed_score}/100</span
                                >
                            </div>
                            <div class="progress-track">
                                <div
                                    class="progress-bar bg-emerald-500"
                                    style="width: {assessment.speed_score}%"
                                ></div>
                            </div>
                        </div>
                        <div class="skill-item">
                            <div class="skill-header">
                                <span>Inisiatif</span>
                                <span class="score"
                                    >{assessment.initiative_score}/100</span
                                >
                            </div>
                            <div class="progress-track">
                                <div
                                    class="progress-bar bg-amber-500"
                                    style="width: {assessment.initiative_score}%"
                                ></div>
                            </div>
                        </div>
                        <div class="skill-item">
                            <div class="skill-header">
                                <span>Kerjasama Tim</span>
                                <span class="score"
                                    >{assessment.teamwork_score}/100</span
                                >
                            </div>
                            <div class="progress-track">
                                <div
                                    class="progress-bar bg-sky-500"
                                    style="width: {assessment.teamwork_score}%"
                                ></div>
                            </div>
                        </div>
                        <div class="skill-item">
                            <div class="skill-header">
                                <span>Komunikasi</span>
                                <span class="score"
                                    >{assessment.communication_score}/100</span
                                >
                            </div>
                            <div class="progress-track">
                                <div
                                    class="progress-bar bg-purple-500"
                                    style="width: {assessment.communication_score}%"
                                ></div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Feedback Section -->
                <div class="feedback-grid">
                    <div class="card feedback-card strengths">
                        <div class="card-title text-emerald-700">
                            <svg
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                class="mr-2"
                                ><path
                                    d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"
                                /></svg
                            >
                            Kelebihan
                        </div>
                        <p class="content">
                            {assessment.strengths || "Tidak ada catatan."}
                        </p>
                    </div>
                    <div class="card feedback-card improvements">
                        <div class="card-title text-amber-700">
                            <svg
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                class="mr-2"
                                ><path
                                    d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
                                /><line x1="12" y1="9" x2="12" y2="13" /><line
                                    x1="12"
                                    y1="17"
                                    x2="12.01"
                                    y2="17"
                                /></svg
                            >
                            Area Perbaikan
                        </div>
                        <p class="content">
                            {assessment.improvements || "Tidak ada catatan."}
                        </p>
                    </div>
                </div>

                <!-- Additional Comments -->
                {#if assessment.comments}
                    <div class="card comment-card">
                        <h3>
                            <svg
                                width="20"
                                height="20"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                class="mr-2"
                                ><path
                                    d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                                /></svg
                            > Komentar Tambahan
                        </h3>
                        <div class="comment-content">
                            <span class="quote">“</span>
                            {assessment.comments}
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    {/if}

    <AssessmentEditModal
        isOpen={showEditModal}
        onClose={() => (showEditModal = false)}
        {assessment}
        onSuccess={handleEditSuccess}
    />
</div>

<style>
    .page-container {
        max-width: 1200px;
        margin: 0 auto;
        padding-bottom: 40px;
        animation: fadeIn 0.4s ease-out;
    }
    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
    }
    .btn-back {
        width: 40px;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: white;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        color: #64748b;
        transition: all 0.2s;
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    }
    .btn-back:hover {
        background: #f8fafc;
        color: #0f172a;
        transform: translateX(-2px);
    }
    .page-title {
        font-size: 24px;
        font-weight: 800;
        color: #0f172a;
        margin: 0;
        letter-spacing: -0.02em;
    }
    .meta {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-top: 4px;
    }
    .badge {
        background: #eff6ff;
        color: #3b82f6;
        font-size: 11px;
        font-weight: 700;
        padding: 4px 8px;
        border-radius: 6px;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }
    .date {
        color: #94a3b8;
        font-size: 13px;
        font-weight: 500;
    }

    /* Moved to bottom for specificity */

    .btn-edit {
        display: flex;
        align-items: center;
        gap: 8px;
        background: #f59e0b;
        color: white;
        font-weight: 600;
        font-size: 14px;
        padding: 10px 20px;
        border-radius: 999px;
        border: none;
        cursor: pointer;
        transition: all 0.2s;
        box-shadow: 0 4px 12px rgba(245, 158, 11, 0.2);
    }
    .btn-edit:hover {
        background: #d97706;
        transform: translateY(-1px);
        box-shadow: 0 6px 16px rgba(245, 158, 11, 0.3);
    }

    @media (max-width: 768px) {
        .btn-edit-full {
            width: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
            background: #f59e0b;
            color: white;
            font-weight: 600;
            font-size: 14px;
            padding: 12px 20px;
            border-radius: 9999px; /* Pill shaped */
            border: none;
            cursor: pointer;
            transition: all 0.2s;
            box-shadow: 0 4px 12px rgba(245, 158, 11, 0.2);
        }
    }
    .btn-edit-full:hover {
        background: #d97706;
        /* No translate for mobile button usually, but keeping inconsistent with desktop if preferred */
        transform: translateY(-1px);
        box-shadow: 0 6px 16px rgba(245, 158, 11, 0.3);
    }

    .dashboard-grid {
        display: grid;
        grid-template-columns: 350px 1fr;
        gap: 24px;
    }
    @media (max-width: 1024px) {
        .dashboard-grid {
            grid-template-columns: 1fr;
        }
    }

    .left-col,
    .right-col {
        display: flex;
        flex-direction: column;
        gap: 24px;
    }

    .card {
        background: white;
        border-radius: 16px;
        padding: 24px;
        box-shadow:
            0 4px 6px -1px rgba(0, 0, 0, 0.02),
            0 2px 4px -1px rgba(0, 0, 0, 0.02);
        border: 1px solid #e2e8f0; /* Darker border */
        transition: box-shadow 0.2s ease; /* Removed transform */
    }
    .card:hover {
        box-shadow:
            0 10px 15px -3px rgba(0, 0, 0, 0.05),
            0 4px 6px -2px rgba(0, 0, 0, 0.025);
        /* No translate */
    }

    .score-card {
        text-align: center;
        /* color: white; - Removed to use dynamic text color */
    }

    .score-circle {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-bottom: 24px;
    }
    .score-circle .grade {
        font-size: 80px;
        font-weight: 900;
        line-height: 1;
        /* color: #fbbf24; - dynamic now */
        text-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    }
    .score-circle .value {
        font-size: 24px;
        font-weight: 600;
        opacity: 0.9;
    }
    .score-circle .total {
        font-size: 16px;
        opacity: 0.6;
    }

    .assessor-info {
        background: rgba(
            255,
            255,
            255,
            0.8
        ); /* Slight transparency to blend better with tint */
        border: 1px solid rgba(226, 232, 240, 0.6);
        border-radius: 12px;
        padding: 12px 16px;
        text-align: left;
    }
    .assessor-info .label {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        margin-bottom: 8px;
    }
    .assessor {
        display: flex;
        align-items: center;
        gap: 12px;
    }
    .avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        /* background: white; - dynamic */
        /* color: #4f46e5; - dynamic */
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
        font-size: 14px;
    }
    .details {
        display: flex;
        flex-direction: column;
    }
    .name {
        font-weight: 600;
        font-size: 14px;
    }
    .role {
        font-size: 12px;
        opacity: 0.8;
    }

    h3 {
        font-size: 16px;
        font-weight: 700;
        color: #1e293b;
        margin: 0 0 20px 0;
        display: flex;
        align-items: center;
    }

    .radar-card {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    .chart-wrapper {
        width: 100%;
        max-width: 300px;
        position: relative;
    }

    .skill-list {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }
    .skill-item {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }
    .skill-header {
        display: flex;
        justify-content: space-between;
        font-size: 14px;
        font-weight: 600;
        color: #334155;
    }
    .skill-header .score {
        color: #64748b;
        font-weight: 500;
    }
    .progress-track {
        height: 8px;
        background: #f1f5f9;
        border-radius: 4px;
        overflow: hidden;
    }
    .progress-bar {
        height: 100%;
        border-radius: 4px;
        transition: width 1s cubic-bezier(0.4, 0, 0.2, 1);
    }

    .feedback-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
    }
    @media (max-width: 640px) {
        .feedback-grid {
            grid-template-columns: 1fr;
        }
    }
    .feedback-card {
        background: #fafafa; /* Slightly distinct from white to stand out as zones */
        border: 1px solid #f1f5f9;
    }
    .card-title {
        font-weight: 700;
        font-size: 14px;
        margin-bottom: 12px;
        display: flex;
        align-items: center;
    }
    .feedback-card .content {
        font-size: 14px;
        color: #475569;
        line-height: 1.6;
        margin: 0;
    }
    .strengths {
        background: #f0fdf4;
        border-color: #dcfce7;
    }
    .improvements {
        background: #fffbeb;
        border-color: #fef3c7;
    }

    .comment-card {
        background: #f8fafc;
        border: 1px dashed #cbd5e1;
    }
    .comment-content {
        position: relative;
        padding-left: 24px;
        color: #475569;
        font-style: italic;
        line-height: 1.6;
    }
    .quote {
        position: absolute;
        left: 0;
        top: -10px;
        font-size: 40px;
        color: #cbd5e1;
        font-family: serif;
    }

    .loading-state,
    .error-state {
        text-align: center;
        padding: 60px;
        color: #64748b;
    }
    .spinner {
        width: 40px;
        height: 40px;
        border: 4px solid #e2e8f0;
        border-top-color: #4f46e5;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin: 0 auto 16px;
    }
    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    /* Responsive Visibility - At the bottom to override other display properties */
    .desktop-only {
        display: flex !important;
    }
    .mobile-only {
        display: none !important;
    }
    @media (max-width: 768px) {
        .desktop-only {
            display: none !important;
        }
        .mobile-only {
            display: flex !important;
        }
    }
</style>
