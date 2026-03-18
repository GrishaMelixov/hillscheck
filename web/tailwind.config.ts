import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // iOS 26 system colors
        ios: {
          blue:    '#0A84FF',
          green:   '#30D158',
          red:     '#FF453A',
          orange:  '#FF9F0A',
          yellow:  '#FFD60A',
          purple:  '#BF5AF2',
          pink:    '#FF375F',
          teal:    '#5AC8FA',
          indigo:  '#5E5CE6',
        },
        // RPG accent
        gold: '#F5C518',
        // Glass surfaces
        glass: {
          '1': 'rgba(255,255,255,0.04)',
          '2': 'rgba(255,255,255,0.07)',
          '3': 'rgba(255,255,255,0.12)',
        },
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"SF Pro Display"', '"SF Pro Text"', 'Inter', 'system-ui', 'sans-serif'],
        mono: ['"SF Mono"', '"JetBrains Mono"', 'monospace'],
      },
      backdropBlur: {
        '4xl': '72px',
        '5xl': '96px',
      },
      borderRadius: {
        '4xl': '2rem',
        '5xl': '2.5rem',
      },
    },
  },
  plugins: [],
} satisfies Config
