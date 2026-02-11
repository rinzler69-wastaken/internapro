<script>
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";
    import { onMount } from "svelte";

    // Props
    let { isOpen, onClose, onSuccess, preSelectedInternId = null } = $props();

    // State
    let form = $state({
        intern_id: "",
        task_id: "",
        quality_score: 80,
        speed_score: 80,
        initiative_score: 80,
        teamwork_score: 80,
        communication_score: 80,
        strengths: "",
        improvements: "",
        comments: "",
        assessment_date: new Date().toISOString().slice(0, 10),
    });

    let interns = $state([]);
    let internTasks = $state([]);
    let assessmentType = $state("general"); // 'general' or 'task'
    let loading = $state(false);
    let submitting = $state(false);

    $effect(() => {
        if (isOpen) {
            if (interns.length === 0) {
                fetchInterns();
            }
            if (preSelectedInternId) {
                form.intern_id = preSelectedInternId;
            }
        }
    });

    $effect(() => {
        if (isOpen && form.intern_id && assessmentType === "task") {
            fetchTasks(form.intern_id);
        } else {
            internTasks = [];
            form.task_id = "";
        }
    });

    async function fetchInterns() {
        try {
            const params = {
                page: 1,
                limit: 100,
                status: "active",
            };

            if (
                auth.user?.role === "supervisor" ||
                auth.user?.role === "pembimbing"
            ) {
                params.supervisor_id = auth.user.id;
            }

            const res = await api.getInterns(params);
            interns = res.data || [];
        } catch (err) {
            console.error("Gagal load interns", err);
        }
    }

    async function fetchTasks(internId) {
        try {
            const res = await api.getTasks({
                intern_id: internId,
                status: "submitted,completed,revision",
            });
            internTasks = res.data || [];
        } catch (err) {
            console.error("Gagal load tasks", err);
        }
    }

    async function handleSubmit() {
        if (!form.intern_id) {
            alert("Silakan pilih intern terlebih dahulu.");
            return;
        }

        if (assessmentType === "task" && !form.task_id) {
            alert("Silakan pilih tugas yang akan dinilai.");
            return;
        }

        submitting = true;
        try {
            const payload = {
                intern_id: Number(form.intern_id),
                quality_score: Number(form.quality_score),
                speed_score: Number(form.speed_score),
                initiative_score: Number(form.initiative_score),
                teamwork_score: Number(form.teamwork_score),
                communication_score: Number(form.communication_score),
                strengths: form.strengths,
                improvements: form.improvements,
                comments: form.comments,
                assessment_date: form.assessment_date,
                aspect: assessmentType === "task" ? "task" : "overall",
            };

            if (assessmentType === "task" && form.task_id) {
                payload.task_id = Number(form.task_id);
            }

            await api.createAssessment(payload);

            alert("Penilaian berhasil disimpan!");

            // Reset text fields
            form.strengths = "";
            form.improvements = "";
            form.comments = "";
            form.intern_id = "";
            form.task_id = "";
            assessmentType = "general";

            // Reset scores to default
            form.quality_score = 80;
            form.speed_score = 80;
            form.initiative_score = 80;
            form.teamwork_score = 80;
            form.communication_score = 80;

            onSuccess?.();
            onClose?.();
        } catch (err) {
            alert(err.message || "Gagal membuat penilaian");
        } finally {
            submitting = false;
        }
    }

    function calculateAverage() {
        const total =
            Number(form.quality_score) +
            Number(form.speed_score) +
            Number(form.initiative_score) +
            Number(form.teamwork_score) +
            Number(form.communication_score);
        return (total / 5).toFixed(1);
    }
</script>

{#if isOpen}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-backdrop" onclick={onClose}>
        <!-- Modal Content -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div
            class="modal-content animate-scale-up"
            onclick={(e) => e.stopPropagation()}
            role="dialog"
            aria-modal="true"
            tabindex="-1"
        >
            <!-- Header -->
            <div class="modal-header">
                <div>
                    <h2 class="modal-title">Buat Penilaian Baru</h2>
                </div>
                <button class="btn-close" onclick={onClose} aria-label="Tutup">
                    <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>
            </div>

            <!-- Body -->
            <div class="modal-body">
                <div class="grid-layout">
                    <!-- Left Section: Data & Scores -->
                    <div class="section left-section">
                        <div class="form-group">
                            <label class="label" for="intern_id"
                                >Pilih Intern</label
                            >
                            <div class="select-wrapper">
                                <select
                                    class="input-field select"
                                    bind:value={form.intern_id}
                                    id="intern_id"
                                >
                                    <option value="">-- Pilih Peserta --</option
                                    >
                                    {#each interns as i}
                                        <option value={i.id}
                                            >{i.full_name} - {i.school}</option
                                        >
                                    {/each}
                                </select>
                                <div class="select-arrow">▼</div>
                            </div>
                        </div>

                        <div class="form-group mt-4">
                            <span class="label">Tipe Penilaian</span>
                            <div class="type-toggle">
                                <button
                                    class="type-btn {assessmentType ===
                                    'general'
                                        ? 'active'
                                        : ''}"
                                    onclick={() => (assessmentType = "general")}
                                >
                                    Umum
                                </button>
                                <button
                                    class="type-btn {assessmentType === 'task'
                                        ? 'active'
                                        : ''}"
                                    onclick={() => (assessmentType = "task")}
                                >
                                    Per Tugas
                                </button>
                            </div>
                        </div>

                        {#if assessmentType === "task"}
                            <div class="form-group mt-4 animate-fade-in">
                                <label class="label" for="task_id"
                                    >Pilih Tugas</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={form.task_id}
                                        id="task_id"
                                        disabled={!form.intern_id}
                                    >
                                        <option value=""
                                            >-- Pilih Tugas --</option
                                        >
                                        {#each internTasks as t}
                                            <option value={t.id}
                                                >{t.title}</option
                                            >
                                        {/each}
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                                {#if !form.intern_id}
                                    <p class="helper-text error">
                                        Pilih intern terlebih dahulu
                                    </p>
                                {:else if internTasks.length === 0}
                                    <p class="helper-text warning">
                                        Intern ini belum memiliki tugas yang
                                        dikerjakan
                                    </p>
                                {/if}
                            </div>
                        {/if}

                        <div class="form-group mt-4">
                            <label class="label" for="assessment_date"
                                >Tanggal Penilaian</label
                            >
                            <input
                                class="input-field"
                                type="date"
                                bind:value={form.assessment_date}
                                id="assessment_date"
                            />
                        </div>

                        <div class="score-grid mt-6">
                            <div class="form-group">
                                <label class="label" for="quality"
                                    >Kualitas</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.quality_score}
                                    id="quality"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="speed"
                                    >Kecepatan</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.speed_score}
                                    id="speed"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="initiative"
                                    >Inisiatif</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.initiative_score}
                                    id="initiative"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="teamwork"
                                    >Kerjasama</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.teamwork_score}
                                    id="teamwork"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="communication"
                                    >Komunikasi</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.communication_score}
                                    id="communication"
                                />
                            </div>

                            <!-- Average Display -->
                            <div class="form-group">
                                <p class="label muted">Rata-rata</p>
                                <div class="average-display">
                                    {calculateAverage()}
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Right Section: Comments -->
                    <div class="section right-section">
                        <div class="form-group">
                            <label class="label" for="strengths"
                                >Kekuatan (Strengths)</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Apa kelebihan intern ini?"
                                bind:value={form.strengths}
                                id="strengths"
                            ></textarea>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label" for="improvements"
                                >Area Pengembangan (Improvements)</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Apa yang perlu diperbaiki?"
                                bind:value={form.improvements}
                                id="improvements"
                            ></textarea>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label" for="comments"
                                >Komentar Tambahan</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Catatan lain..."
                                bind:value={form.comments}
                                id="comments"
                            ></textarea>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div class="modal-footer">
                <button class="btn-secondary" onclick={onClose}>Batal</button>
                <button
                    class="btn-primary"
                    onclick={handleSubmit}
                    disabled={submitting}
                >
                    {#if submitting}
                        <div class="spinner-small"></div>
                        Menyimpan...
                    {:else}
                        Simpan Penilaian
                    {/if}
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(15, 23, 42, 0.6);
        backdrop-filter: blur(4px);
        z-index: 50;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 20px;
        box-sizing: border-box;
    }

    .modal-content {
        background: white;
        width: 100%;
        max-width: 900px;
        max-height: 90vh;
        display: flex;
        flex-direction: column;
        border-radius: 20px;
        box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
        overflow: hidden;
    }

    .modal-header {
        padding: 20px 24px;
        border-bottom: 1px solid #f1f5f9;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .modal-title {
        font-size: 20px;
        font-weight: 700;
        color: #0f172a;
        margin: 0;
    }

    .btn-close {
        background: none;
        border: none;
        cursor: pointer;
        color: #94a3b8;
        padding: 4px;
        border-radius: 8px;
        transition: all 0.2s;
    }
    .btn-close:hover {
        background: #f1f5f9;
        color: #0f172a;
    }

    .modal-body {
        padding: 24px;
        overflow-y: auto;
    }

    .grid-layout {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 32px;
    }
    @media (max-width: 768px) {
        .grid-layout {
            grid-template-columns: 1fr;
            gap: 24px;
        }
    }

    .section {
        display: flex;
        flex-direction: column;
        gap: 0;
    }
    .form-group {
        margin-bottom: 0;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .label {
        font-size: 12px;
        font-weight: 600;
        color: #475569;
        text-transform: uppercase;
        letter-spacing: 0.02em;
    }
    .label.muted {
        color: #94a3b8;
    }

    .input-field {
        width: 100%;
        padding: 10px 12px;
        border: 1px solid #cbd5e1;
        border-radius: 8px;
        font-size: 14px;
        color: #0f172a;
        transition: all 0.2s;
        background: #fff;
        font-family: "Inter", sans-serif;
        box-sizing: border-box;
    }
    .input-field:focus {
        outline: none;
        border-color: #10b981;
        box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15);
    }
    .textarea {
        resize: vertical;
        line-height: 1.5;
    }

    .select-wrapper {
        position: relative;
    }
    .select {
        appearance: none;
        cursor: pointer;
    }
    .select-arrow {
        position: absolute;
        right: 12px;
        top: 50%;
        transform: translateY(-50%);
        font-size: 10px;
        color: #64748b;
        pointer-events: none;
    }

    .type-toggle {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 8px;
        background: #f1f5f9;
        padding: 4px;
        border-radius: 10px;
        border: 1px solid #e2e8f0;
    }
    .type-btn {
        padding: 8px;
        border: none;
        background: transparent;
        color: #64748b;
        font-size: 13px;
        font-weight: 600;
        border-radius: 7px;
        cursor: pointer;
        transition: all 0.2s;
    }
    .type-btn.active {
        background: white;
        color: #10b981;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
    }
    .type-btn:hover:not(.active) {
        color: #0f172a;
        background: rgba(255, 255, 255, 0.5);
    }

    .helper-text {
        font-size: 11px;
        margin: 4px 0 0 0;
    }
    .helper-text.error {
        color: #ef4444;
    }
    .helper-text.warning {
        color: #f59e0b;
    }

    .animate-fade-in {
        animation: fadeIn 0.3s ease-out;
    }
    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(-5px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    /* Score Grid Customization for Modal */
    .score-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
        background: #f8fafc;
        padding: 16px;
        border-radius: 12px;
        border: 1px solid #f1f5f9;
    }
    .score-input {
        text-align: center;
        font-weight: 600;
    }

    .average-display {
        font-size: 20px;
        font-weight: 800;
        color: #10b981;
        background: white;
        border: 1px solid #e2e8f0;
        border-radius: 8px;
        padding: 6px;
        text-align: center;
    }

    /* Footer */
    .modal-footer {
        padding: 20px 24px;
        border-top: 1px solid #f1f5f9;
        background: #fcfcfc;
        display: flex;
        justify-content: flex-end;
        gap: 12px;
    }

    .btn-secondary {
        background: white;
        border: 1px solid #e2e8f0;
        color: #475569;
        padding: 10px 18px;
        border-radius: 8px;
        font-weight: 600;
        cursor: pointer;
        font-size: 14px;
    }
    .btn-secondary:hover {
        background: #f8fafc;
        color: #0f172a;
    }

    .btn-primary {
        background: linear-gradient(135deg, #10b981, #059669);
        color: white;
        padding: 10px 20px;
        border-radius: 8px;
        font-weight: 600;
        font-size: 14px;
        border: none;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2);
    }
    .btn-primary:hover:not(:disabled) {
        transform: translateY(-1px);
        box-shadow: 0 6px 16px rgba(16, 185, 129, 0.3);
    }
    .btn-primary:disabled {
        opacity: 0.7;
        cursor: not-allowed;
        transform: none;
    }

    .spinner-small {
        width: 14px;
        height: 14px;
        border: 2px solid white;
        border-top-color: transparent;
        border-radius: 50%;
        animation: spin 1s linear infinite;
    }
    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    .animate-scale-up {
        animation: scaleUp 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
        transform: scale(0.95);
        opacity: 0;
    }
    @keyframes scaleUp {
        to {
            transform: scale(1);
            opacity: 1;
        }
    }
</style>
