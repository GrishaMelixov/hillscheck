import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        rpg: {
          gold:   '#f5c518',
          purple: '#7c3aed',
          red:    '#dc2626',
          blue:   '#2563eb',
          green:  '#16a34a',
        },
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', 'monospace'],
      },
    },
  },
  plugins: [],
} satisfies Config
