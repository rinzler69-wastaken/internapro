<script>
    let { currentPage = 1, totalPages = 1, onPageChange } = $props();

    function goToPreviousPage() {
        if (currentPage > 1) {
            onPageChange(currentPage - 1);
        }
    }

    function goToNextPage() {
        if (currentPage < totalPages) {
            onPageChange(currentPage + 1);
        }
    }
</script>

<div
    class="pagination-wrapper {totalPages <= 1
        ? 'opacity-50 pointer-events-none'
        : ''}"
>
    <!-- Previous Button -->
    <button
        class="pagination-btn prev-btn {currentPage <= 1 ? 'disabled' : ''}"
        onclick={goToPreviousPage}
        disabled={currentPage <= 1}
        aria-label="Previous Page"
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
        >
            <path d="M15 18l-6-6 6-6" />
        </svg>
        <span class="btn-label">Prev</span>
    </button>

    <!-- Page Indicator -->
    <div class="pagination-indicator">
        <span class="current-page">{currentPage}</span>
        <span class="separator">of</span>
        <span class="total-pages">{totalPages}</span>
    </div>

    <!-- Next Button -->
    <button
        class="pagination-btn next-btn {currentPage >= totalPages
            ? 'disabled'
            : ''}"
        onclick={goToNextPage}
        disabled={currentPage >= totalPages}
        aria-label="Next Page"
    >
        <span class="btn-label">Next</span>
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
        >
            <path d="M9 18l6-6-6-6" />
        </svg>
    </button>
</div>

<style>
    .pagination-wrapper {
        display: flex;
        align-items: center;
        gap: 8px;
        width: 100%;
    }

    @media (min-width: 768px) {
        .pagination-wrapper {
            width: auto;
        }
    }

    .pagination-btn {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 8px 20px;
        border-radius: 9999px;
        font-size: 0.875rem;
        font-weight: 600;
        transition: all 0.2s;
        background-color: white;
        color: #0f172a;
        border: 1px solid #e2e8f0;
        cursor: pointer;
        min-height: 38px;
    }

    .pagination-btn:hover:not(.disabled) {
        border-color: #cbd5e1;
        background-color: #f8fafc;
    }

    .pagination-btn.disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .pagination-indicator {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 8px 20px;
        border-radius: 9999px;
        font-size: 0.875rem;
        font-weight: 600;
        background-color: white;
        color: #0f172a;
        border: 1px solid #e2e8f0;
        min-height: 38px;
    }

    .separator {
        color: #64748b;
        font-weight: 400;
    }

    @media (max-width: 767px) {
        .pagination-wrapper {
            justify-content: center;
            gap: 12px;
        }

        .pagination-btn {
            flex: 1;
            min-width: 44px;
            height: 38px;
            padding: 0 12px;
            border-radius: 9999px;
        }

        .btn-label {
            display: none;
        }

        .pagination-indicator {
            flex: 1;
            padding: 0 16px;
            background-color: white;
            border: 1px solid #e2e8f0;
            border-radius: 9999px;
            min-height: 38px;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        /* Re-center the arrows on mobile if needed */
        .pagination-btn svg {
            margin: 0;
        }
    }

    @media (min-width: 768px) {
        .pagination-btn {
            flex: 0 0 auto;
        }
        .pagination-indicator {
            flex: 0 0 auto;
        }
    }
</style>
