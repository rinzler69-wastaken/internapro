<script>
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";
    import { portal } from "../lib/portal.js";

    // Props
    let { isOpen, onClose, onSuccess, taskId } = $props();

    // State
    let title = $state("");
    let description = $state("");
    let priority = $state("medium");
    let startDate = $state("");
    let deadline = $state("");
    let deadlineTime = $state("");
    let submissionMethod = $state("links");
    let internId = $state(""); // Single intern for edit

    let interns = $state([]);
    let loading = $state(false);
    let saving = $state(false);

    // Fetch details when taskId changes or modal opens
    $effect(() => {
        if (isOpen && taskId) {
            fetchTaskDetails();
            if (auth.user?.role !== "intern") {
                fetchInterns();
            }
        }
    });

    async function fetchTaskDetails() {
        loading = true;
        try {
            const res = await api.getTask(taskId);
            const data = res.data;

            title = data.title || "";
            description = data.description || "";
            priority = data.priority || "medium";
            submissionMethod = data.submission_method || "links";
            internId = data.intern_id || "";

            // Handle Dates
            if (data.start_date) startDate = data.start_date.split("T")[0];
            if (data.deadline) deadline = data.deadline.split("T")[0];
            deadlineTime = data.deadline_time
                ? data.deadline_time.slice(0, 5)
                : "23:59";
        } catch (err) {
            console.error("Failed to load task", err);
            handleClose();
        } finally {
            loading = false;
        }
    }

    async function fetchInterns() {
        if (interns.length > 0) return;
        try {
            const res = await api.getInterns({ status: "active", limit: 100 });
            interns = res.data || [];
        } catch (err) {
            console.error(err);
        }
    }

    async function handleSave() {
        if (!title) {
            alert("Judul tugas wajib diisi");
            return;
        }

        saving = true;
        try {
            await api.updateTask(taskId, {
                title,
                description,
                priority,
                start_date: startDate,
                deadline,
                deadline_time: deadlineTime,
                submission_method: submissionMethod,
                intern_id: internId ? Number(internId) : null,
            });

            onSuccess?.();
            handleClose();
        } catch (err) {
            alert(err.message || "Gagal memperbarui tugas");
        } finally {
            saving = false;
        }
    }

    function handleClose() {
        onClose?.();
        // Reset state slightly? content will be overwritten on next open
    }
</script>

{#if isOpen}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-backdrop" onclick={handleClose} use:portal>
        <!-- Modal Content -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div
            class="modal-content animate-scale-up"
            onclick={(e) => e.stopPropagation()}
            role="dialog"
            aria-modal="true"
        >
            <!-- Header -->
            <div class="modal-header">
                <div>
                    <h2 class="modal-title">Edit Tugas</h2>
                    <!-- <p class="modal-subtitle">Perbarui informasi tugas</p> -->
                </div>
                <button class="btn-close" onclick={handleClose}>
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
                <!-- {#if loading}{:else} -->
                <div class="grid-layout">
                    <!-- Left Section -->
                    <div class="section left-section">
                        <div class="form-group">
                            <label class="label" for="edit_title"
                                >Judul Tugas <span class="text-red-500">*</span
                                ></label
                            >
                            <input
                                class="input-field"
                                bind:value={title}
                                id="edit_title"
                                placeholder="Judul tugas..."
                            />
                        </div>

                        <div class="form-group mt-4">
                            <label class="label" for="edit_desc"
                                >Deskripsi & Instruksi</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="4"
                                bind:value={description}
                                id="edit_desc"
                                placeholder="Detail instruksi..."
                            ></textarea>
                        </div>

                        <div class="grid-2 mt-4">
                            <div class="form-group">
                                <label class="label" for="edit_priority"
                                    >Prioritas</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={priority}
                                        id="edit_priority"
                                    >
                                        <option value="low">Low</option>
                                        <option value="medium">Medium</option>
                                        <option value="high">High</option>
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit_method"
                                    >Metode Pengumpulan</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={submissionMethod}
                                        id="edit_method"
                                    >
                                        <option value="links">Link Only</option>
                                        <option value="files">File Only</option>
                                        <option value="both">Link & File</option
                                        >
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                            </div>
                        </div>

                        <div class="grid-2 mt-4">
                            <div class="form-group">
                                <label class="label" for="edit_startDate"
                                    >Tanggal Mulai</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={startDate}
                                    id="edit_startDate"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="edit_deadline"
                                    >Deadline</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={deadline}
                                    id="edit_deadline"
                                />
                            </div>
                        </div>
                        <div class="form-group mt-4">
                            <label class="label" for="edit_deadlineTime"
                                >Waktu Deadline</label
                            >
                            <input
                                class="input-field"
                                type="time"
                                bind:value={deadlineTime}
                                id="edit_deadlineTime"
                            />
                        </div>
                    </div>

                    <!-- Right Section: Intern Assignment (Single) -->
                    {#if auth.user?.role !== "intern"}
                        <div class="section right-section">
                            <div class="form-group">
                                <label class="label" for="edit_intern"
                                    >Ditugaskan Kepada</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={internId}
                                        id="edit_intern"
                                    >
                                        <option value=""
                                            >-- Pilih Intern --</option
                                        >
                                        {#each interns as i}
                                            <option value={i.id}
                                                >{i.full_name || i.name} - {i.school}</option
                                            >
                                        {/each}
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                                <p class="text-xs text-slate-500 mt-2">
                                    Mengubah penerima tugas akan memindahkan
                                    tugas ini ke intern yang dipilih.
                                </p>
                            </div>
                        </div>
                    {/if}
                </div>
                <!-- {/if} -->
            </div>

            <!-- Footer -->
            <div class="modal-footer">
                <button class="btn-secondary" onclick={handleClose}
                    >Batal</button
                >
                <button
                    class="btn-primary"
                    onclick={handleSave}
                    disabled={saving || loading}
                >
                    {#if saving}
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
        align-items: flex-start;
    }
    .modal-title {
        font-size: 20px;
        font-weight: 700;
        color: #0f172a;
        margin: 0;
    }
    .modal-subtitle {
        color: #64748b;
        font-size: 14px;
        margin: 4px 0 0 0;
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
        grid-template-columns: 1.4fr 1fr;
        gap: 32px;
    }
    @media (max-width: 768px) {
        .grid-layout {
            grid-template-columns: 1fr;
            gap: 24px;
        }
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
        box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
    }
    .textarea {
        resize: vertical;
        line-height: 1.5;
    }

    .grid-2 {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
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

    .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid #e2e8f0;
        border-top-color: #10b981;
        border-radius: 50%;
        margin: 0 auto 10px;
        animation: spin 1s linear infinite;
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
    .loading-state {
        text-align: center;
        padding: 40px;
        color: #94a3b8;
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
