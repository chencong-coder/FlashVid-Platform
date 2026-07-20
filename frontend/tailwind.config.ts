import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#fe2c55',
        ink: '#050505',
        panel: '#171717',
      },
      boxShadow: {
        overlay: '0 10px 36px rgb(0 0 0 / 36%)',
      },
      animation: {
        'disc-spin': 'spin 4s linear infinite',
        marquee: 'marquee 8s linear infinite',
      },
      keyframes: {
        marquee: {
          '0%': { transform: 'translateX(0)' },
          '45%': { transform: 'translateX(0)' },
          '100%': { transform: 'translateX(-35%)' },
        },
      },
    },
  },
  plugins: [],
} satisfies Config
