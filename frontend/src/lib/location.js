import { readable } from 'svelte/store';

const getLocation = () => ({
  path: typeof window !== 'undefined' ? window.location.pathname : '/',
  search: typeof window !== 'undefined' ? window.location.search : '',
  hash: typeof window !== 'undefined' ? window.location.hash : '',
});

export const location = readable(getLocation(), (set) => {
  if (typeof window === 'undefined') return () => {};

  const update = () => set(getLocation());
  update();

  window.addEventListener('popstate', update);
  window.addEventListener('pushState', update);
  window.addEventListener('replaceState', update);

  return () => {
    window.removeEventListener('popstate', update);
    window.removeEventListener('pushState', update);
    window.removeEventListener('replaceState', update);
  };
});
