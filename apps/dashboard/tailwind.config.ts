import { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        ironclad: {
          50: '#f8fafc',
          100: '#f1f5f9',
          500: '#3b82f6',
          900: '#0f172a',
          950: '#020617',
        },
      },
    },
  },

  plugins: [],
};

export default config;
