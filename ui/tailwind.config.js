/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx,js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        ink: {
          950: '#07090C',
          900: '#0B0E14',
          800: '#11151D',
          700: '#1A2030',
          600: '#252B3A',
          500: '#3A4256',
        },
        'signal-green': '#3DDC97',
        'signal-amber': '#FFB454',
        'signal-rose': '#FF5675',
        'thaana-text': '#E7ECF3',
        'mid': '#9AA4B2',
        'dim': '#5C6577',
        'line': 'rgba(255,255,255,0.06)',
        'violet-400': '#B968FF',
        'mint-400': '#3DDC97',
        'pink-400': '#FF6B9D',
        'sky-400': '#5EC4FF',
        'teal-400': '#4FF0C8',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        thaana: ['"MV Boli"', '"Noto Sans Thaana"', 'sans-serif'],
      },
      borderRadius: {
        lg: '0.5rem',
        md: '0.375rem',
        xl: '0.75rem',
        '2xl': '1rem',
      },
      transitionTimingFunction: {
        'out-expo': 'cubic-bezier(0.16, 1, 0.3, 1)',
      },
      transitionDuration: {
        150: '150ms',
      },
      animation: {
        shimmer: 'shimmer 1.4s linear infinite',
        'radar-sweep': 'radar-sweep 2s linear infinite',
        'pulse-flash': 'pulse-flash 0.6s ease-out',
        'aurora-loop': 'aurora-loop 30s ease-in-out infinite',
        'slide-up': 'slide-up 200ms cubic-bezier(0.16, 1, 0.3, 1)',
        'fade-in': 'fade-in 200ms ease-out',
      },
      keyframes: {
        shimmer: {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(100%)' },
        },
        'radar-sweep': {
          '0%': { transform: 'rotate(0deg)' },
          '100%': { transform: 'rotate(360deg)' },
        },
        'pulse-flash': {
          '0%': { backgroundColor: 'rgba(61, 220, 151, 0.18)' },
          '100%': { backgroundColor: 'rgba(61, 220, 151, 0)' },
        },
        'aurora-loop': {
          '0%, 100%': { transform: 'translate(0,0) scale(1)' },
          '33%': { transform: 'translate(4%, -3%) scale(1.05)' },
          '66%': { transform: 'translate(-3%, 4%) scale(0.97)' },
        },
        'slide-up': {
          '0%': { transform: 'translateY(8px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}
