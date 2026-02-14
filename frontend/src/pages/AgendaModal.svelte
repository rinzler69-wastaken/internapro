<script>
    import { createEventDispatcher } from "svelte";
    import { api } from "../lib/api.js"; // Needed for delete

    // Props
    let { show = false, date = null, agenda = null } = $props();

    const dispatch = createEventDispatcher();

    let title = $state("");
    let description = $state("");
    let time = $state("");
    let loading = $state(false);
    let isEditing = $state(false); // New state to track mode

    // Reset or fill form when modal opens
    $effect(() => {
        if (show) {
            if (agenda) {
                // Existing agenda -> View mode initially, fill data
                title = agenda.title;
                description = agenda.description || "";
                time = agenda.time ? agenda.time.substring(0, 5) : "";
                isEditing = false;
            } else {
                // New agenda -> Edit mode, reset data
                title = "";
                description = "";
                time = "";
                isEditing = true;
            }
        }
    });

    function close() {
        dispatch("close");
    }

    function toggleEdit() {
        isEditing = !isEditing;
    }

    async function deleteAgenda() {
        if (!agenda || !agenda.id) return;
        if (!confirm("Apakah Anda yakin ingin menghapus agenda ini?")) return;

        loading = true;
        try {
            await api.deleteAgenda(agenda.id);
            dispatch("save", { deleted: true }); // Trigger refresh in parent
            loading = false;
        } catch (err) {
            console.error("Failed to delete agenda:", err);
            loading = false;
            // Optionally show error
        }
    }

    function save() {
        if (!title) return;

        loading = true;

        // Construct payload
        const payload = {
            title,
            description,
            date: date ? toDateKey(date) : null,
            time: time ? time + ":00" : "00:00:00",
        };

        // If updating, include ID
        if (agenda && agenda.id) {
            payload.id = agenda.id;
        }

        if (date instanceof Date) {
            payload.date = toDateKey(date);
        } else if (typeof date === "string") {
            payload.date = date;
        }

        dispatch("save", payload);
        loading = false;
    }

    function toDateKey(d) {
        if (!d) return "";
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, "0");
        const day = String(d.getDate()).padStart(2, "0");
        return `${year}-${month}-${day}`;
    }
</script>

{#if show}
    <div
        class="fixed inset-0 z-[60] flex items-center justify-center p-4 sm:p-6"
        role="dialog"
        aria-modal="true"
    >
        <div
            class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"
            onclick={close}
            aria-hidden="true"
        ></div>

        <div
            class="relative bg-white rounded-2xl shadow-xl w-full max-w-md flex flex-col overflow-hidden animate-scale-in"
        >
            <!-- Header -->
            <div
                class="p-5 border-b border-slate-100 flex justify-between items-center bg-slate-50/50"
            >
                <h3
                    class="text-lg font-bold text-slate-800 flex items-center gap-2"
                >
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        class="text-indigo-600"
                    >
                        {#if isEditing}
                            <path
                                d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                            ></path>
                            <path
                                d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                            ></path>
                        {:else}
                            <rect
                                x="3"
                                y="4"
                                width="18"
                                height="18"
                                rx="2"
                                ry="2"
                            ></rect>
                            <line x1="16" y1="2" x2="16" y2="6"></line>
                            <line x1="8" y1="2" x2="8" y2="6"></line>
                            <line x1="3" y1="10" x2="21" y2="10"></line>
                        {/if}
                    </svg>
                    {isEditing
                        ? agenda
                            ? "Edit Agenda"
                            : "Tambah Agenda"
                        : "Detail Agenda"}
                </h3>
                <div class="flex items-center gap-2">
                    {#if !isEditing && agenda}
                        <button
                            onclick={toggleEdit}
                            class="w-8 h-8 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center hover:bg-indigo-100 transition-colors"
                            title="Edit Agenda"
                        >
                            <svg
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <path
                                    d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                                ></path>
                                <path
                                    d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                                ></path>
                            </svg>
                        </button>
                        <button
                            onclick={deleteAgenda}
                            class="w-8 h-8 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center hover:bg-rose-100 transition-colors"
                            title="Hapus Agenda"
                        >
                            <svg
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <polyline points="3 6 5 6 21 6"></polyline>
                                <path
                                    d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                                ></path>
                            </svg>
                        </button>
                    {/if}
                    <button
                        onclick={close}
                        class="w-8 h-8 rounded-lg bg-slate-100 text-slate-500 flex items-center justify-center hover:bg-slate-200 transition-colors"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            ><line x1="18" y1="6" x2="6" y2="18" /><line
                                x1="6"
                                y1="6"
                                x2="18"
                                y2="18"
                            /></svg
                        >
                    </button>
                </div>
            </div>

            <!-- Body -->
            <div class="p-6 space-y-4">
                {#if date}
                    <div
                        class="flex items-center gap-2 text-sm text-slate-500 bg-slate-50 p-2 rounded-lg border border-slate-100"
                    >
                        <svg
                            width="16"
                            height="16"
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
                        <span class="font-medium">
                            {date instanceof Date
                                ? date.toLocaleDateString("id-ID", {
                                      weekday: "long",
                                      day: "numeric",
                                      month: "long",
                                      year: "numeric",
                                  })
                                : date}
                        </span>
                    </div>
                {/if}

                {#if isEditing}
                    <!-- EDIT FORM -->
                    <div>
                        <label
                            for="title"
                            class="block text-sm font-bold text-slate-700 mb-1"
                            >Judul Agenda <span class="text-rose-500">*</span
                            ></label
                        >
                        <input
                            id="title"
                            type="text"
                            bind:value={title}
                            placeholder="Contoh: Meeting Tim, Deadline Project"
                            class="w-full px-4 py-2 rounded-xl border border-slate-200 focus:border-indigo-300 focus:ring-2 focus:ring-indigo-100 transition-all outline-none"
                        />
                    </div>

                    <div>
                        <label
                            for="time"
                            class="block text-sm font-bold text-slate-700 mb-1"
                            >Waktu</label
                        >
                        <input
                            id="time"
                            type="time"
                            bind:value={time}
                            class="w-full px-4 py-2 rounded-xl border border-slate-200 focus:border-indigo-300 focus:ring-2 focus:ring-indigo-100 transition-all outline-none"
                        />
                    </div>

                    <div>
                        <label
                            for="description"
                            class="block text-sm font-bold text-slate-700 mb-1"
                            >Deskripsi</label
                        >
                        <textarea
                            id="description"
                            bind:value={description}
                            rows="3"
                            placeholder="Tambahkan detail agenda..."
                            class="w-full px-4 py-2 rounded-xl border border-slate-200 focus:border-indigo-300 focus:ring-2 focus:ring-indigo-100 transition-all outline-none resize-none"
                        ></textarea>
                    </div>
                {:else}
                    <!-- VIEW MODE -->
                    <div class="space-y-4">
                        <div>
                            <h4 class="text-sm font-bold text-slate-500 mb-1">
                                Judul Agenda
                            </h4>
                            <p class="text-lg font-bold text-slate-800">
                                {title}
                            </p>
                        </div>
                        <div>
                            <h4 class="text-sm font-bold text-slate-500 mb-1">
                                Waktu
                            </h4>
                            <div class="flex items-center gap-2">
                                <span
                                    class="font-mono font-bold text-blue-600 bg-blue-50 px-2 py-1 rounded text-sm border border-blue-100"
                                >
                                    {time || "--:--"}
                                </span>
                            </div>
                        </div>
                        {#if description}
                            <div>
                                <h4
                                    class="text-sm font-bold text-slate-500 mb-1"
                                >
                                    Deskripsi
                                </h4>
                                <p
                                    class="text-slate-700 whitespace-pre-line leading-relaxed"
                                >
                                    {description}
                                </p>
                            </div>
                        {/if}
                    </div>
                {/if}
            </div>

            <!-- Footer -->
            {#if isEditing}
                <div
                    class="p-4 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-2"
                >
                    <button
                        onclick={() => {
                            if (agenda) {
                                isEditing = false; // Cancel edit, back to view
                                // Reset fields to agenda values
                                title = agenda.title;
                                description = agenda.description || "";
                                time = agenda.time
                                    ? agenda.time.substring(0, 5)
                                    : "";
                            } else {
                                close(); // Cancel creation, close modal
                            }
                        }}
                        class="px-4 py-2 rounded-lg border border-slate-200 text-slate-600 font-semibold text-sm hover:bg-white hover:border-slate-300 transition-all"
                    >
                        Batal
                    </button>
                    <button
                        onclick={save}
                        disabled={!title || loading}
                        class="px-5 py-2 rounded-lg bg-indigo-600 text-white font-bold text-sm hover:bg-indigo-700 shadow-md shadow-indigo-200 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
                    >
                        {#if loading}
                            <svg
                                class="animate-spin -ml-1 mr-1 h-4 w-4 text-white"
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                            >
                                <circle
                                    class="opacity-25"
                                    cx="12"
                                    cy="12"
                                    r="10"
                                    stroke="currentColor"
                                    stroke-width="4"
                                ></circle>
                                <path
                                    class="opacity-75"
                                    fill="currentColor"
                                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                ></path>
                            </svg>
                            Menyimpan...
                        {:else}
                            Simpan Agenda
                        {/if}
                    </button>
                </div>
            {/if}
        </div>
    </div>
{/if}
