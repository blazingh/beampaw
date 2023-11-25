/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./template/**/*.{html,js}"],
  theme: {
    extend: {},
    fontFamily: {
      sans: ["Poppins", "sans-serif"],
    },
  },
  daisyui: {
    themes: [
      "night",
    ],
  },
  plugins: [
    require("daisyui"),
  ],
}

