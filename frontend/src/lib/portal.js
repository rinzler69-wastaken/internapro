export function portal(node, target = '#overlay-root') {
  let targetEl =
    typeof target === 'string' ? document.querySelector(target) : target;

  if (!targetEl && typeof document !== 'undefined') {
    targetEl = document.body;
  }

  if (targetEl && node.parentNode !== targetEl) {
    targetEl.appendChild(node);
  }

  return {
    destroy() {
      if (targetEl && targetEl.contains(node)) {
        targetEl.removeChild(node);
      }
    }
  };
}
