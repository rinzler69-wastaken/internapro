import { writable } from 'svelte/store';

export const isSidebarCollapsed = writable(false);
export const isMobileSidebarOpen = writable(false);
