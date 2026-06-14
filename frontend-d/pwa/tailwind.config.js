/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: {
          base: '#0a0a0f',
          card: '#15151f',
          elevated: '#1f1f2b',
          subtle: '#26263a'
        },
        safe: '#22c55e',
        block: '#ef4444',
        warn: '#f97316',
        accent: '#8b5cf6'
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Pretendard', 'system-ui', 'sans-serif']
      }
    }
  },
  plugins: []
}
