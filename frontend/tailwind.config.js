/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'alfa-red': '#EF3124',
        'alfa-dark': '#1A1A1A',
        'alfa-gray': '#F5F5F5',
        'alfa-light': '#FFFFFF',
        'alfa-black': '#000',
        'alfa-placeholder': '#E9ECEF',
        'alfa-stroke': '#DFDFDF',
      },
    },
  },
  plugins: [],
}

