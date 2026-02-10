// Simple Svelte action to move a node to another DOM container (default: #overlay-root).
// Ensures overlays/modals are not clipped by parent stacking contexts.
const COUNT_KEY = 'portalCount';

export function portal(node, target = '#overlay-root') {
  let targetEl =
    typeof target === 'string' ? document.querySelector(target) : target;

  // Fallback to document.body if overlay root is missing.
  if (!targetEl && typeof document !== 'undefined') {
    targetEl = document.body;
  }

  const getCount = () => {
    const raw = targetEl?.dataset[COUNT_KEY];
    const n = parseInt(raw || '0', 10);
    return Number.isFinite(n) ? n : 0;
  };

  const setCount = (val) => {
    if (targetEl) targetEl.dataset[COUNT_KEY] = String(val);
  };

  const updatePointerEvents = () => {
    if (targetEl) {
      const n = getCount();
      targetEl.style.pointerEvents = n > 0 ? 'auto' : 'none';
    }
  };

  if (targetEl && node.parentNode !== targetEl) {
    targetEl.appendChild(node);
    setCount(getCount() + 1);
    updatePointerEvents();
  }

  return {
    destroy() {
      if (targetEl && targetEl.contains(node)) {
        targetEl.removeChild(node);
        setCount(Math.max(0, getCount() - 1));
        updatePointerEvents();
      }
    }
  };
}
