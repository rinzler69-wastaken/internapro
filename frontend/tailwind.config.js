/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      fontFamily: {
        geist: ['Geist', 'sans-serif'],
        mono: ['Geist Mono', 'monospace'],
        jakarta: ['Plus Jakarta Sans', 'sans-serif'],
      },
      colors: {
        // Vercel-ish Monochrome Scale
        background: '#ffffff',
        foreground: '#09090b',
        muted: '#71717a',
        border: '#e4e4e7',
        accent: '#8b5cf6',
      }
    },
  },
  plugins: [],
}
