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
          900: '#0f172a',
        },
      },
    },
  },
  plugins: [],
};

export default config;
