<script>
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";

    // Props
    let { isOpen, onClose, onSuccess, assessment } = $props();

    // State
    let form = $state({
        quality_score: 80,
        speed_score: 80,
        initiative_score: 80,
        teamwork_score: 80,
        communication_score: 80,
        strengths: "",
        improvements: "",
        comments: "",
        assessment_date: "",
    });

    let submitting = $state(false);

    $effect(() => {
        if (isOpen && assessment) {
            // Pre-fill form
            form.quality_score = assessment.quality_score || 0;
            form.speed_score = assessment.speed_score || 0;
            form.initiative_score = assessment.initiative_score || 0;
            form.teamwork_score = assessment.teamwork_score || 0;
            form.communication_score = assessment.communication_score || 0;
            form.strengths = assessment.strengths || "";
            form.improvements = assessment.improvements || "";
            form.comments = assessment.comments || "";
            form.assessment_date = assessment.assessment_date
                ? assessment.assessment_date.slice(0, 10)
                : "";
        }
    });

    async function handleSubmit() {
        submitting = true;
        try {
            await api.updateAssessment(assessment.id, {
                quality_score: Number(form.quality_score),
                speed_score: Number(form.speed_score),
                initiative_score: Number(form.initiative_score),
                teamwork_score: Number(form.teamwork_score),
                communication_score: Number(form.communication_score),
                strengths: form.strengths,
                improvements: form.improvements,
                comments: form.comments,
                assessment_date: form.assessment_date,
            });

            alert("Penilaian berhasil diperbarui!");
            onSuccess?.();
            onClose?.();
        } catch (err) {
            alert(err.message || "Gagal memperbarui penilaian");
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
                    <h2 class="modal-title">Edit Penilaian</h2>
                    {#if assessment?.intern_name}
                        <p class="modal-subtitle">
                            Intern: {assessment.intern_name}
                            {#if assessment.task_title}
                                <span class="badge-task">
                                    • {assessment.task_title}</span
                                >
                            {/if}
                        </p>
                    {/if}
                </div>
                <button
                    class="btn-close"
                    onclick={onClose}
                    aria-label="Close modal"
                >
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
                            <label class="label" for="edit-assessment_date"
                                >Tanggal Penilaian</label
                            >
                            <input
                                class="input-field"
                                type="date"
                                bind:value={form.assessment_date}
                                id="edit-assessment_date"
                            />
                        </div>

                        <div class="score-grid mt-6">
                            <div class="form-group">
                                <label class="label" for="edit-quality"
                                    >Kualitas</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.quality_score}
                                    id="edit-quality"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit-speed"
                                    >Kecepatan</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.speed_score}
                                    id="edit-speed"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit-initiative"
                                    >Inisiatif</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.initiative_score}
                                    id="edit-initiative"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit-teamwork"
                                    >Kerjasama</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.teamwork_score}
                                    id="edit-teamwork"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit-communication"
                                    >Komunikasi</label
                                >
                                <input
                                    class="input-field score-input"
                                    type="number"
                                    min="0"
                                    max="100"
                                    bind:value={form.communication_score}
                                    id="edit-communication"
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
                            <label class="label" for="edit-strengths"
                                >Kekuatan (Strengths)</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Apa kelebihan intern ini?"
                                bind:value={form.strengths}
                                id="edit-strengths"
                            ></textarea>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label" for="edit-improvements"
                                >Area Pengembangan (Improvements)</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Apa yang perlu diperbaiki?"
                                bind:value={form.improvements}
                                id="edit-improvements"
                            ></textarea>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label" for="edit-comments"
                                >Komentar Tambahan</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="3"
                                placeholder="Catatan lain..."
                                bind:value={form.comments}
                                id="edit-comments"
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
                        Simpan Perubahan
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
    .modal-subtitle {
        font-size: 13px;
        color: #64748b;
        margin: 4px 0 0 0;
        display: flex;
        align-items: center;
        gap: 6px;
    }

    .badge-task {
        background: #f1f5f9;
        color: #334155;
        padding: 2px 8px;
        border-radius: 6px;
        font-size: 11px;
        font-weight: 600;
        border: 1px solid #e2e8f0;
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
