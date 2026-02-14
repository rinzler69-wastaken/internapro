<script>
    import { api } from "../lib/api.js";
    import { portal } from "../lib/portal.js";
    import { onMount } from "svelte";
    import { auth } from "../lib/auth.svelte.js";

    // Props
    let { isOpen, onClose, onSuccess, preSelected = [] } = $props();

    // State
    let title = $state("");
    let description = $state("");
    let priority = $state("medium");
    let startDate = $state("");
    let deadline = $state("");
    let deadlineTime = $state("");
    let assignTo = $state("all");
    let submissionMethod = $state("links");
    let assignerID = $state(0);
    let customAssignerName = $state("");
    let assignerType = $state("supervisor");

    // Search & Selection State
    let searchQuery = $state("");
    let results = $state([]);
    let selected = $state([]);
    let supervisors = $state([]);
    let admins = $state([]);
    let loading = $state(false);

    onMount(async () => {
        // We now fetch when modal opens
    });

    // Fetch lists when modal opens (Lazy Loading)
    $effect(() => {
        if (isOpen) {
            (async () => {
                try {
                    const [supRes, admRes] = await Promise.all([
                        api.getSupervisorsPublic(),
                        api.getAdminsPublic(),
                    ]);
                    supervisors = supRes.data || [];
                    admins = admRes.data || [];
                } catch (err) {
                    console.error(err);
                }
            })();
        }
    });

    // Initialize dates and form when modal opens
    $effect(() => {
        if (isOpen) {
            const today = new Date().toISOString().slice(0, 10);
            if (!startDate) startDate = today;
            if (!deadlineTime) deadlineTime = "23:59";

            if (auth.user?.role === "intern") {
                assignTo = "selected";
                assignerType = "supervisor";

                // Set supervisor ID if available
                if (auth.user.supervisor_id) {
                    assignerID = Number(auth.user.supervisor_id);
                }

                // Add self as selected intern
                const internProfile = {
                    id: auth.user.intern_id || 0,
                    name: auth.user.name,
                };
                selected = [internProfile];
            } else {
                if (preSelected && preSelected.length > 0) {
                    selected = [...preSelected];
                    assignTo = "selected";
                }
            }
        }
    });

    // Reset assignerID when switching between supervisor/admin tabs
    $effect(() => {
        if (auth.user?.role === "intern") {
            if (assignerType === "supervisor" && auth.user.supervisor_id) {
                assignerID = Number(auth.user.supervisor_id);
            } else if (assignerType === "admin") {
                assignerID = 0; // Reset to default when switching to admin
            }
        }
    });

    async function searchInterns() {
        if (!searchQuery) {
            results = [];
            return;
        }
        try {
            const res = await api.searchInterns(searchQuery);
            results = res.data || [];
        } catch (err) {
            console.error(err);
        }
    }

    function toggleIntern(intern) {
        if (selected.find((i) => i.id === intern.id)) {
            selected = selected.filter((i) => i.id !== intern.id);
        } else {
            selected = [...selected, intern];
        }
    }

    function removeSelected(id) {
        selected = selected.filter((i) => i.id !== id);
    }

    async function handleSubmit() {
        if (!title) {
            alert("Judul tugas wajib diisi");
            return;
        }

        if (auth.user?.role === "intern" && !assignerID) {
            alert("Silakan pilih pembimbing/admin yang menugaskan");
            return;
        }

        loading = true;
        try {
            await api.createTask({
                title,
                description,
                priority,
                start_date: startDate,
                deadline,
                deadline_time: deadlineTime,
                assign_to: assignTo,
                submission_method: submissionMethod,
                intern_ids: selected.map((i) => i.id),
                assigner_id: Number(assignerID),
                custom_assigner_name: customAssignerName,
            });

            // Reset form
            title = "";
            description = "";
            priority = "medium";
            assignTo = auth.user?.role === "intern" ? "selected" : "all";
            selected = auth.user?.role === "intern" ? selected : [];
            assignerID =
                auth.user?.role === "intern"
                    ? Number(auth.user.supervisor_id) || 0
                    : 0;
            assignerType = "supervisor";
            customAssignerName = "";

            onSuccess?.();
            onClose?.();
        } catch (err) {
            alert(err.message || "Gagal membuat tugas");
        } finally {
            loading = false;
        }
    }

    function handleClose() {
        onClose?.();
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
                    <h2 class="modal-title">Buat Tugas Baru</h2>
                </div>
                <button class="btn-close" onclick={handleClose}>
                    <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        ><line x1="18" y1="6" x2="6" y2="18"></line><line
                            x1="6"
                            y1="6"
                            x2="18"
                            y2="18"
                        ></line></svg
                    >
                </button>
            </div>

            <!-- Scrollable Body -->
            <div class="modal-body">
                <div class="grid-layout">
                    <!-- KOLOM KIRI: DETAIL TUGAS -->
                    <div class="section left-section">
                        <div class="form-group">
                            <label class="label" for="title"
                                >Judul Tugas <span class="text-red-500">*</span
                                ></label
                            >
                            <input
                                class="input-field"
                                bind:value={title}
                                id="title"
                                placeholder="Contoh: Implementasi API Login"
                            />
                        </div>

                        <div class="form-group mt-4">
                            <label class="label" for="description"
                                >Deskripsi & Instruksi</label
                            >
                            <textarea
                                class="input-field textarea"
                                rows="4"
                                bind:value={description}
                                id="description"
                                placeholder="Jelaskan detail pekerjaan..."
                            ></textarea>
                        </div>

                        <div class="grid-2 mt-4">
                            <div class="form-group">
                                <label class="label" for="priority"
                                    >Prioritas</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={priority}
                                        id="priority"
                                    >
                                        <option value="low">Low</option>
                                        <option value="medium">Medium</option>
                                        <option value="high">High</option>
                                    </select>
                                    <div class="select-arrow">▼</div>
                                </div>
                            </div>
                            <div class="form-group">
                                <label class="label" for="submissionMethod"
                                    >Metode Pengumpulan</label
                                >
                                <div class="select-wrapper">
                                    <select
                                        class="input-field select"
                                        bind:value={submissionMethod}
                                        id="submissionMethod"
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
                                <label class="label" for="startDate"
                                    >Tanggal Mulai</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={startDate}
                                    id="startDate"
                                />
                            </div>
                            <div class="form-group">
                                <label class="label" for="deadline"
                                    >Deadline</label
                                >
                                <input
                                    class="input-field"
                                    type="date"
                                    bind:value={deadline}
                                    id="deadline"
                                />
                            </div>
                        </div>

                        <div class="form-group mt-4">
                            <label class="label" for="deadlineTime"
                                >Waktu Deadline</label
                            >
                            <input
                                class="input-field"
                                type="time"
                                bind:value={deadlineTime}
                                id="deadlineTime"
                            />
                        </div>
                    </div>

                    <!-- KOLOM KANAN: PENUGASAN -->
                    <div class="section right-section">
                        {#if auth.user?.role === "intern"}
                            <!-- INTERN VIEW: SELECT ASSIGNER -->
                            <div class="form-group">
                                <label class="label">Tipe Pemberi Tugas</label>
                                <div class="radio-group mb-2">
                                    <label
                                        class={`radio-btn ${assignerType === "supervisor" ? "active" : ""}`}
                                    >
                                        <input
                                            type="radio"
                                            value="supervisor"
                                            bind:group={assignerType}
                                            hidden
                                        />
                                        <span>Pembimbing</span>
                                    </label>
                                    <label
                                        class={`radio-btn ${assignerType === "admin" ? "active" : ""}`}
                                    >
                                        <input
                                            type="radio"
                                            value="admin"
                                            bind:group={assignerType}
                                            hidden
                                        />
                                        <span>Admin</span>
                                    </label>
                                </div>
                            </div>

                            <div class="form-group mt-2">
                                <label class="label" for="assignerID">
                                    Pilih {assignerType === "admin"
                                        ? "Admin"
                                        : "Pembimbing"} Penugasan
                                    <span class="text-red-500">*</span>
                                </label>
                                {#if auth.user?.role === "intern" && assignerType === "supervisor"}
                                    <div class="read-only-card">
                                        {#if auth.user?.supervisor_id}
                                            <div class="card-avatar">
                                                {auth.user.supervisor_name?.[0]?.toUpperCase() ||
                                                    "P"}
                                            </div>
                                            <div class="card-info">
                                                <span class="card-tag"
                                                    >Pembimbing Anda</span
                                                >
                                                <span class="card-name">
                                                    {auth.user
                                                        .supervisor_name ||
                                                        "Pembimbing"}
                                                </span>
                                            </div>
                                            <div class="card-icon">
                                                <span
                                                    class="material-symbols-outlined"
                                                    >lock</span
                                                >
                                            </div>
                                        {:else}
                                            <p class="text-sm text-gray-600">
                                                Tidak ada pembimbing yang
                                                ditentukan.
                                            </p>
                                        {/if}
                                    </div>
                                {:else}
                                    <div class="select-wrapper">
                                        <select
                                            class="input-field select"
                                            bind:value={assignerID}
                                            id="assignerID"
                                        >
                                            <option value={0}
                                                >-- Pilih {assignerType ===
                                                "admin"
                                                    ? "Admin"
                                                    : "Pembimbing"} --</option
                                            >
                                            {#each assignerType === "admin" ? admins : supervisors as p}
                                                <option value={p.user_id}>
                                                    {p.name || p.full_name}
                                                </option>
                                            {/each}
                                        </select>
                                        <div class="select-arrow">▼</div>
                                    </div>
                                {/if}
                            </div>

                            <div class="form-group mt-4">
                                <label class="label" for="customAssignerName"
                                    >Nama Pemberi Tugas (Opsional)</label
                                >
                                <input
                                    class="input-field"
                                    bind:value={customAssignerName}
                                    id="customAssignerName"
                                    placeholder="Contoh: Pak Shafwan"
                                />
                                <p class="info-box">
                                    Isi bila tugas diberikan oleh staf tertentu
                                    di kantor.
                                </p>
                            </div>

                            <div class="info-box mt-6">
                                <p>
                                    Tugas ini akan dicatat sebagai laporan
                                    mandiri dan menunggu review dari pembimbing.
                                </p>
                            </div>
                        {:else}
                            <!-- ADMIN VIEW -->
                            <div class="form-group">
                                <label class="label" for="assignTo"
                                    >Penerima Tugas</label
                                >
                                <div class="radio-group">
                                    <label
                                        class={`radio-btn ${assignTo === "all" ? "active" : ""}`}
                                    >
                                        <input
                                            type="radio"
                                            value="all"
                                            bind:group={assignTo}
                                            hidden
                                        />
                                        <span>Semua</span>
                                    </label>
                                    <label
                                        class={`radio-btn ${assignTo === "selected" ? "active" : ""}`}
                                    >
                                        <input
                                            type="radio"
                                            value="selected"
                                            bind:group={assignTo}
                                            hidden
                                        />
                                        <span>Manual</span>
                                    </label>
                                </div>
                            </div>

                            {#if assignTo === "selected"}
                                <div
                                    class="search-section mt-4 animate-fade-in"
                                >
                                    <div class="form-group">
                                        <div class="search-wrapper">
                                            <svg
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                                class="search-icon"
                                                ><circle
                                                    cx="11"
                                                    cy="11"
                                                    r="8"
                                                /><line
                                                    x1="21"
                                                    y1="21"
                                                    x2="16.65"
                                                    y2="16.65"
                                                /></svg
                                            >
                                            <input
                                                class="input-field pl-9"
                                                bind:value={searchQuery}
                                                oninput={searchInterns}
                                                placeholder="Cari peserta..."
                                                id="searchQuery"
                                                style="font-size: 13px;"
                                            />
                                        </div>
                                    </div>

                                    <div class="selection-container">
                                        {#if selected.length > 0}
                                            <div class="selected-tags">
                                                {#each selected as item}
                                                    <div class="tag">
                                                        {item.label?.split(
                                                            " - ",
                                                        )[0] || item.name}
                                                        <button
                                                            onclick={() =>
                                                                removeSelected(
                                                                    item.id,
                                                                )}
                                                            class="btn-remove"
                                                            >×</button
                                                        >
                                                    </div>
                                                {/each}
                                            </div>
                                        {/if}

                                        <div class="results-list">
                                            {#if results.length > 0}
                                                {#each results as intern}
                                                    <div
                                                        class="intern-card {selected.find(
                                                            (i) =>
                                                                i.id ===
                                                                intern.id,
                                                        )
                                                            ? 'selected'
                                                            : ''}"
                                                        onclick={() =>
                                                            toggleIntern(
                                                                intern,
                                                            )}
                                                    >
                                                        <div
                                                            class="checkbox-circle"
                                                        >
                                                            {#if selected.find((i) => i.id === intern.id)}✓{/if}
                                                        </div>
                                                        <span
                                                            class="intern-name"
                                                            >{intern.label}</span
                                                        >
                                                    </div>
                                                {/each}
                                            {:else if searchQuery}
                                                <div class="empty-search">
                                                    Tidak ada hasil.
                                                </div>
                                            {/if}
                                        </div>
                                    </div>
                                </div>
                            {:else}
                                <div class="info-box mt-4">
                                    <p>
                                        Tugas akan diberikan kepada seluruh
                                        peserta magang yang statusnya <strong
                                            >Aktif</strong
                                        >.
                                    </p>
                                </div>
                            {/if}
                        {/if}
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div class="modal-footer">
                <button class="btn-secondary" onclick={handleClose}
                    >Batal</button
                >
                <button
                    class="btn-primary"
                    onclick={handleSubmit}
                    disabled={loading}
                >
                    {#if loading}
                        <div class="spinner-small"></div>
                        Menyimpan...
                    {:else}
                        Buat Tugas
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

    .section {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    /* Form Elements */
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

    /* Select */
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

    /* Right Section Specifics */
    .radio-group {
        display: flex;
        gap: 8px;
    }
    .radio-btn {
        flex: 1;
        padding: 8px;
        border: 1px solid #e2e8f0;
        border-radius: 8px;
        text-align: center;
        cursor: pointer;
        font-size: 13px;
        font-weight: 500;
        color: #64748b;
        transition: all 0.2s;
        background: #f8fafc;
    }
    .radio-btn:hover {
        background: #f1f5f9;
    }
    .radio-btn.active {
        border-color: #10b981;
        background: #ecfdf5;
        color: #059669;
        font-weight: 600;
    }

    .search-wrapper {
        position: relative;
    }
    .search-icon {
        position: absolute;
        left: 10px;
        top: 11px;
        color: #94a3b8;
        pointer-events: none;
    }
    .pl-9 {
        padding-left: 34px;
    }

    .selection-container {
        border: 1px solid #e2e8f0;
        border-radius: 8px;
        padding: 10px;
        margin-top: 8px;
        background: #fcfcfc;
        min-height: 150px;
        display: flex;
        flex-direction: column;
    }

    .selected-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        margin-bottom: 10px;
    }
    .tag {
        background: #10b981;
        color: white;
        font-size: 11px;
        padding: 3px 8px;
        border-radius: 6px;
        display: flex;
        align-items: center;
        gap: 6px;
        font-weight: 600;
    }
    .btn-remove {
        background: none;
        border: none;
        color: white;
        cursor: pointer;
        font-size: 14px;
        padding: 0;
        opacity: 0.8;
    }

    .results-list {
        overflow-y: auto;
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 200px;
    }

    .intern-card {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 8px;
        border: 1px solid #e2e8f0;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.1s;
        background: white;
    }
    .intern-card:hover {
        border-color: #cbd5e1;
        background: #f8fafc;
    }
    .intern-card.selected {
        border-color: #10b981;
        background: #ecfdf5;
    }

    .checkbox-circle {
        width: 16px;
        height: 16px;
        border-radius: 50%;
        border: 2px solid #cbd5e1;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 10px;
        color: white;
        flex-shrink: 0;
    }
    .intern-card.selected .checkbox-circle {
        background: #10b981;
        border-color: #10b981;
    }
    .intern-name {
        font-size: 13px;
        color: #334155;
        font-weight: 500;
    }

    .empty-search {
        text-align: center;
        font-size: 13px;
        color: #94a3b8;
        padding: 20px;
        font-style: italic;
    }
    .info-box {
        background: #eff6ff;
        border: 1px solid #bfdbfe;
        color: #1e40af;
        padding: 12px;
        border-radius: 8px;
        font-size: 13px;
        line-height: 1.5;
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

    .read-only-card {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 16px;
        background: #f8fafc;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        margin-top: 8px;
        transition: all 0.2s;
    }
    .read-only-card:hover {
        background: #f1f5f9;
        border-color: #cbd5e1;
    }
    .card-avatar {
        width: 36px;
        height: 36px;
        border-radius: 10px;
        background: #6366f1;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
        font-size: 14px;
        box-shadow: 0 2px 4px rgba(99, 102, 241, 0.2);
    }
    .card-info {
        display: flex;
        flex-direction: column;
    }
    .card-tag {
        font-size: 10px;
        font-weight: 700;
        color: #6366f1;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        line-height: 1;
        margin-bottom: 2px;
    }
    .card-name {
        font-size: 14px;
        font-weight: 600;
        color: #1e293b;
    }
    .card-icon {
        margin-left: auto;
        color: #94a3b8;
    }
    .card-icon span {
        font-size: 18px;
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
