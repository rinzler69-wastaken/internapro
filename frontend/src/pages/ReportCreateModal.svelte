<script>
    import { api } from "../lib/api.js";
    import { auth } from "../lib/auth.svelte.js";
    import { portal } from "../lib/portal.js";
    import { onMount } from "svelte";

    // Props
    let { isOpen, onClose, onSuccess } = $props();

    // State
    let title = $state("");
    let content = $state("");
    let type = $state("weekly");
    let periodStart = $state("");
    let periodEnd = $state("");
    let internId = $state("");
    let isDraft = $state(true);

    // Data
    let interns = $state([]);
    let loading = $state(false);

    $effect(() => {
        if (isOpen) {
            // Reset form
            if (!title) {
                const today = new Date();
                // Default period: start of this week to end of this week
                const day = today.getDay() || 7; // Get current day number, converting Sun. to 7
                if (day !== 1) today.setHours(-24 * (day - 1)); // set to Monday
                periodStart = today.toISOString().slice(0, 10);

                const end = new Date(today);
                end.setDate(today.getDate() + 4); // Friday
                periodEnd = end.toISOString().slice(0, 10);

                isDraft = true; // reset draft to true
            }
            if (auth.user?.role !== "intern" && interns.length === 0) {
                fetchInterns();
            }
        }
    });

    async function fetchInterns() {
        try {
            const res = await api.getInterns({
                page: 1,
                limit: 100,
                status: "active",
            });
            interns = res.data || [];
        } catch (err) {
            console.error("Gagal load interns", err);
        }
    }

    async function handleSubmit() {
        if (!title || !content || !periodStart || !periodEnd) {
            alert("Mohon lengkapi semua field wajib.");
            return;
        }

        loading = true;
        try {
            await api.createReport({
                title,
                content,
                type,
                period_start: periodStart,
                period_end: periodEnd,
                intern_id: internId ? Number(internId) : null,
                status: isDraft ? "draft" : "submitted",
            });

            // Reset
            title = "";
            content = "";

            onSuccess?.();
            onClose?.();
        } catch (err) {
            alert(err.message || "Gagal membuat laporan");
        } finally {
            loading = false;
        }
    }
</script>

{#if isOpen}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-backdrop" onclick={onClose} use:portal>
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
                    <h2 class="modal-title">Buat Laporan Baru</h2>
                </div>
                <button class="btn-close" onclick={onClose}>
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
                    <!-- Left Section: Details -->
                    <div class="section left-section">
                        <div class="form-group">
                            <label class="label" for="title"
                                >Judul Kegiatan <span class="text-red-500"
                                    >*</span
                                ></label
                            >
                            <input
                                class="input-field"
                                bind:value={title}
                                id="title"
                                placeholder="Contoh: Implementasi UI Dashboard"
                            />
                        </div>

                        <div class="form-group mt-4">
                            <label class="label" for="content"
                                >Isi Laporan <span class="text-red-500">*</span
                                ></label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="6"
                                bind:value={content}
                                id="content"
                                placeholder="Jelaskan detail aktivitas yang dilakukan..."
                            ></textarea>
                        </div>
                    </div>

                    <!-- Right Section: Meta & User -->
                    <div class="section right-section">
                        <div class="form-group">
                            <label class="label" for="type">Tipe Laporan</label>
                            <div class="select-wrapper">
                                <select
                                    class="input-field select"
                                    bind:value={type}
                                    id="type"
                                >
                                    <option value="weekly"
                                        >Laporan Mingguan</option
                                    >
                                    <option value="monthly"
                                        >Laporan Bulanan</option
                                    >
                                    <option value="final">Laporan Akhir</option>
                                </select>
                                <div class="select-arrow">▼</div>
                            </div>
                        </div>

                        <div class="grid-2 mt-4">
                            <div class="form-group">
                                <label class="label" for="periodStart"
                                    >Mulai</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={periodStart}
                                    id="periodStart"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="periodEnd"
                                    >Selesai</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={periodEnd}
                                    id="periodEnd"
                                />
                            </div>
                        </div>

                        {#if auth.user?.role !== "intern"}
                            <div class="form-group mt-4">
                                <label class="label" for="internId"
                                    >Pilih Peserta</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={internId}
                                        id="internId"
                                    >
                                        <option value=""
                                            >-- Pilih Intern --</option
                                        >
                                        {#each interns as i}
                                            <option value={i.id}
                                                >{i.full_name}</option
                                            >
                                        {/each}
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                            </div>
                        {/if}

                        <div
                            class="form-group mt-6 p-4 bg-slate-50 rounded-lg border border-slate-100"
                        >
                            <label
                                class="flex items-center gap-3 cursor-pointer"
                            >
                                <input
                                    type="checkbox"
                                    bind:checked={isDraft}
                                    class="w-5 h-5 text-emerald-600 rounded focus:ring-emerald-500 border-gray-300"
                                />
                                <span class="font-semibold text-slate-700"
                                    >Simpan sebagai draf</span
                                >
                            </label>
                            <p
                                class="text-xs text-slate-500 mt-2 ml-8 leading-relaxed"
                            >
                                {#if isDraft}
                                    Laporan akan disimpan sebagai draf dan belum
                                    dikirim ke pembimbing.
                                {:else}
                                    Laporan akan dikirim langsung. <strong
                                        class="text-amber-600"
                                        >Anda tidak dapat mengubahnya kembali
                                        menjadi draf setelah dikirim.</strong
                                    >
                                {/if}
                            </p>
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
                    disabled={loading}
                >
                    {#if loading}
                        <div class="spinner-small"></div>
                        Menyimpan...
                    {:else}
                        Simpan Laporan
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
        grid-template-columns: 1.4fr 1fr;
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
    .text-red-500 {
        color: #ef4444;
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
