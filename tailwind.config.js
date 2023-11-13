/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./template/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: [
      "night",
    ],
  },
  plugins: [require("daisyui")],
}

