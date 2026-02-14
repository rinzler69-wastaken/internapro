
import { auth } from './auth.svelte.js';

export function getAvatarUrl(path) {
    if (!path || typeof path !== 'string') return null;
    if (path.startsWith("http")) return path;

    // Ensure path starts with /uploads/ if it's a local file
    // Some paths might already have it, some might not.
    // The backend usually stores just the filename or partial path.
    // Adjust based on observation from Profile.svelte which does:
    // const base = path.startsWith("/uploads/") ? path : `/uploads/${path}`;

    const base = path.startsWith("/uploads/") ? path : `/uploads/${path}`;

    const qs = [];
    if (auth.token) qs.push(`token=${auth.token}`);
    // Add timestamp for cache busting
    qs.push(`t=${Date.now()}`);

    return `${base}${base.includes("?") ? "&" : "?"}${qs.join("&")}`;
}
