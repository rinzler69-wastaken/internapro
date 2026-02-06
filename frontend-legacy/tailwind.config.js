/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        'geist': ['Geist', 'system-ui', 'sans-serif'],
        'inter': ['Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        vercel: {
          'bg': '#000000',
          'fg': '#fafafa',
          'gray': {
            50: '#fafafa',
            100: '#f5f5f5',
            200: '#e5e5e5',
            300: '#d4d4d4',
            400: '#a3a3a3',
            500: '#737373',
            600: '#525252',
            700: '#404040',
            800: '#262626',
            900: '#171717',
          },
          'border': '#262626',
          'accent': '#0070f3',
        }
      },
      boxShadow: {
        'vercel-sm': '0 2px 4px 0 rgba(0,0,0,0.05)',
        'vercel': '0 4px 8px 0 rgba(0,0,0,0.08)',
        'vercel-lg': '0 8px 16px 0 rgba(0,0,0,0.12)',
        'vercel-xl': '0 12px 24px 0 rgba(0,0,0,0.15)',
      },
      borderRadius: {
        'vercel': '8px',
      }
    },
  },
  plugins: [],
}
